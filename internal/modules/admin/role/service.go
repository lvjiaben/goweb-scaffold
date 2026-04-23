package admin_role

import (
	"fmt"
	"strings"
	"time"

	"github.com/lvjiaben/goweb-scaffold/internal/bootstrap"
)

type Service struct {
	repo *Repo
}

func NewService(runtime *bootstrap.Runtime) *Service {
	return &Service{repo: NewRepo(runtime)}
}

func (s *Service) Options() (map[string]any, error) {
	roles, err := s.repo.ActiveRoles()
	if err != nil {
		return nil, err
	}
	items := make([]RoleOption, 0, len(roles))
	for _, role := range roles {
		items = append(items, RoleOption{
			Label: role.Name,
			Value: role.ID,
			Code:  role.Code,
		})
	}
	return map[string]any{"list": items}, nil
}

func (s *Service) List(params ListParams) (map[string]any, error) {
	filter := roleListFilter{
		Keyword: bootstrap.LikeKeyword(params.Keyword),
		Name:    bootstrap.LikeKeyword(bootstrap.FilterString(params.Filters, "name")),
		Code:    bootstrap.LikeKeyword(bootstrap.FilterString(params.Filters, "code", "description")),
	}
	if status, ok := bootstrap.FilterInt64(params.Filters, "status"); ok {
		filter.Status = &status
	}

	total, err := s.repo.Count(filter)
	if err != nil {
		return nil, err
	}
	roles, err := s.repo.List(filter, params.Page, params.PageSize)
	if err != nil {
		return nil, err
	}
	items := make([]ListItem, 0, len(roles))
	for _, role := range roles {
		items = append(items, ListItem{
			ID:          role.ID,
			Name:        role.Name,
			Code:        role.Code,
			Description: role.Code,
			Status:      role.Status,
			CreatedAt:   role.CreatedAt,
			UpdatedAt:   role.UpdatedAt,
		})
	}
	return bootstrap.PagedResult(items, total, params.Page, params.PageSize), nil
}

func (s *Service) Detail(id int64) (DetailResponse, error) {
	role, err := s.repo.FindByID(id)
	if err != nil {
		return DetailResponse{}, err
	}
	menuIDs, err := s.repo.MenuIDsByRoleID(role.ID)
	if err != nil {
		return DetailResponse{}, err
	}
	return DetailResponse{
		ID:          role.ID,
		Name:        role.Name,
		Code:        role.Code,
		Description: role.Code,
		Status:      role.Status,
		MenuIDs:     menuIDs,
		CreatedAt:   role.CreatedAt,
		UpdatedAt:   role.UpdatedAt,
	}, nil
}

func (s *Service) SaveRole(req SaveRequest) (SaveResult, error) {
	req.Name = strings.TrimSpace(req.Name)
	if strings.TrimSpace(req.Code) == "" {
		req.Code = strings.TrimSpace(req.Description)
	}
	req.Code = strings.TrimSpace(req.Code)
	if req.Name == "" {
		return SaveResult{}, validationError("name is required")
	}
	if req.Code == "" {
		return SaveResult{}, validationError("code is required")
	}
	if err := s.ensureRoleCodeUnique(req.ID, req.Code); err != nil {
		return SaveResult{}, validationError(err.Error())
	}

	var savedID int64
	err := s.repo.WithTransaction(func(tx *Repo) error {
		role := AdminRole{}
		if req.ID == 0 {
			role = AdminRole{
				Name:   req.Name,
				Code:   req.Code,
				Status: req.Status,
			}
			if err := tx.Create(&role); err != nil {
				return err
			}
		} else {
			existing, err := tx.FindByID(req.ID)
			if err != nil {
				return err
			}
			role = existing
			if err := tx.Update(&role, map[string]any{
				"name":       req.Name,
				"code":       req.Code,
				"status":     req.Status,
				"updated_at": time.Now(),
			}); err != nil {
				return err
			}
		}
		if err := tx.ReplaceRoleMenus(role.ID, req.MenuIDs); err != nil {
			return err
		}
		savedID = role.ID
		return nil
	})
	if err != nil {
		return SaveResult{}, err
	}
	return SaveResult{ID: savedID}, nil
}

func (s *Service) DeleteRoles(ids []int64) (DeleteResult, error) {
	ids = bootstrap.NormalizeIDs(ids...)
	if len(ids) == 0 {
		return DeleteResult{}, validationError("ids is required")
	}
	for _, id := range ids {
		if id == 1 {
			return DeleteResult{}, validationError("super admin role cannot be deleted")
		}
	}
	bindCount, err := s.repo.CountUserBindings(ids)
	if err != nil {
		return DeleteResult{}, err
	}
	if bindCount > 0 {
		return DeleteResult{}, validationError("role still bound to admin users")
	}
	if err := s.repo.DeleteRolesAndRelations(ids); err != nil {
		return DeleteResult{}, err
	}
	return DeleteResult{Deleted: len(ids)}, nil
}

func (s *Service) ensureRoleCodeUnique(currentID int64, code string) error {
	count, err := s.repo.CountByCode(code, currentID)
	if err != nil {
		return err
	}
	if count > 0 {
		return fmt.Errorf("role code already exists")
	}
	return nil
}

type validationError string

func (e validationError) Error() string {
	return string(e)
}

func isValidationError(err error) bool {
	_, ok := err.(validationError)
	return ok
}
