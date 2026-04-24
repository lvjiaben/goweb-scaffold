package admin_admin

import (
	"fmt"
	"strings"
	"time"

	"github.com/lvjiaben/goweb-scaffold/internal/bootstrap"
	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	repo *Repo
}

func NewService(runtime *bootstrap.Runtime) *Service {
	return &Service{repo: NewRepo(runtime)}
}

func (s *Service) List(params ListParams) (map[string]any, error) {
	filter := userListFilter{
		KeywordPlain:  strings.TrimSpace(params.Keyword),
		UsernamePlain: strings.TrimSpace(bootstrap.FilterString(params.Filters, "username")),
		NicknamePlain: strings.TrimSpace(bootstrap.FilterString(params.Filters, "nickname", "realname")),
		SortBy:        strings.TrimSpace(params.SortBy),
		SortOrder:     strings.TrimSpace(params.SortOrder),
	}
	if status, ok := bootstrap.FilterInt64(params.Filters, "status"); ok {
		filter.Status = &status
	}

	total, err := s.repo.Count(filter)
	if err != nil {
		return nil, err
	}
	users, err := s.repo.List(filter, params.Page, params.PageSize)
	if err != nil {
		return nil, err
	}

	roleMap, roleNameMap, err := s.roleMaps(users)
	if err != nil {
		return nil, err
	}

	items := make([]ListItem, 0, len(users))
	for _, user := range users {
		items = append(items, ListItem{
			ID:        user.ID,
			Username:  user.Username,
			Nickname:  user.Nickname,
			RealName:  user.Nickname,
			Status:    user.Status,
			IsSuper:   user.IsSuper,
			RoleIDs:   roleMap[user.ID],
			RoleNames: roleNameMap[user.ID],
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		})
	}

	return bootstrap.PagedResult(items, total, params.Page, params.PageSize), nil
}

func (s *Service) Detail(id int64) (DetailResponse, error) {
	user, err := s.repo.FindByID(id)
	if err != nil {
		return DetailResponse{}, err
	}
	roleIDs, err := s.repo.RoleIDsByUserID(user.ID)
	if err != nil {
		return DetailResponse{}, err
	}
	return DetailResponse{
		ID:        user.ID,
		Username:  user.Username,
		Nickname:  user.Nickname,
		RealName:  user.Nickname,
		Status:    user.Status,
		IsSuper:   user.IsSuper,
		RoleIDs:   roleIDs,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}, nil
}

func (s *Service) SaveUser(req SaveRequest) (SaveResult, error) {
	req.Username = strings.TrimSpace(req.Username)
	if strings.TrimSpace(req.Nickname) == "" {
		req.Nickname = strings.TrimSpace(req.RealName)
	}
	req.Nickname = strings.TrimSpace(req.Nickname)
	if req.Username == "" {
		return SaveResult{}, validationError("username is required")
	}
	if req.Nickname == "" {
		return SaveResult{}, validationError("nickname is required")
	}
	if req.ID == 0 && req.Password == "" {
		return SaveResult{}, validationError("password is required")
	}
	if err := s.ensureUsernameUnique(req.ID, req.Username); err != nil {
		return SaveResult{}, validationError(err.Error())
	}

	var savedID int64
	err := s.repo.WithTransaction(func(tx *Repo) error {
		user := AdminUser{}
		if req.ID > 0 {
			existing, err := tx.FindByID(req.ID)
			if err != nil {
				return err
			}
			user = existing
		}

		if req.ID == 0 {
			passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
			if err != nil {
				return err
			}
			user = AdminUser{
				Username:     req.Username,
				PasswordHash: string(passwordHash),
				Nickname:     req.Nickname,
				Status:       req.Status,
				IsSuper:      req.IsSuper,
			}
			if err := tx.Create(&user); err != nil {
				return err
			}
		} else {
			updates := map[string]any{
				"username":   req.Username,
				"nickname":   req.Nickname,
				"status":     req.Status,
				"is_super":   req.IsSuper,
				"updated_at": time.Now(),
			}
			if req.Password != "" {
				passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
				if err != nil {
					return err
				}
				updates["password_hash"] = string(passwordHash)
			}
			if err := tx.Update(&user, updates); err != nil {
				return err
			}
		}

		if err := tx.ReplaceUserRoles(user.ID, req.RoleIDs); err != nil {
			return err
		}
		savedID = user.ID
		return nil
	})
	if err != nil {
		return SaveResult{}, err
	}
	return SaveResult{ID: savedID}, nil
}

func (s *Service) DeleteUsers(ids []int64) (DeleteResult, error) {
	ids = bootstrap.NormalizeIDs(ids...)
	if len(ids) == 0 {
		return DeleteResult{}, validationError("ids is required")
	}
	for _, id := range ids {
		if id == 1 {
			return DeleteResult{}, validationError("default admin cannot be deleted")
		}
	}
	if err := s.repo.DeleteUsersAndRelations(ids); err != nil {
		return DeleteResult{}, err
	}
	return DeleteResult{Deleted: len(ids)}, nil
}

func (s *Service) ensureUsernameUnique(currentID int64, username string) error {
	count, err := s.repo.CountByUsername(username, currentID)
	if err != nil {
		return err
	}
	if count > 0 {
		return fmt.Errorf("username already exists")
	}
	return nil
}

func (s *Service) roleMaps(users []AdminUser) (map[int64][]int64, map[int64][]string, error) {
	userIDs := make([]int64, 0, len(users))
	for _, user := range users {
		userIDs = append(userIDs, user.ID)
	}

	roleMap := make(map[int64][]int64, len(userIDs))
	roleNameMap := make(map[int64][]string, len(userIDs))
	mappings, err := s.repo.UserRoleMappings(userIDs)
	if err != nil {
		return nil, nil, err
	}
	roleIDs := make([]int64, 0, len(mappings))
	for _, item := range mappings {
		roleMap[item.UserID] = append(roleMap[item.UserID], item.RoleID)
		roleIDs = append(roleIDs, item.RoleID)
	}
	roles, err := s.repo.RolesByIDs(roleIDs)
	if err != nil {
		return nil, nil, err
	}
	roleIndex := make(map[int64]string, len(roles))
	for _, role := range roles {
		roleIndex[role.ID] = role.Name
	}
	for userID, ids := range roleMap {
		for _, roleID := range ids {
			if name := roleIndex[roleID]; name != "" {
				roleNameMap[userID] = append(roleNameMap[userID], name)
			}
		}
	}
	return roleMap, roleNameMap, nil
}

type validationError string

func (e validationError) Error() string {
	return string(e)
}

func isValidationError(err error) bool {
	_, ok := err.(validationError)
	return ok
}
