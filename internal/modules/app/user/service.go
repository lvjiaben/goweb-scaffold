package app_user

import (
	"errors"
	"fmt"
	"strings"
	"time"

	coreauth "github.com/lvjiaben/goweb-core/auth"
	"github.com/lvjiaben/goweb-scaffold/internal/bootstrap"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type Service struct {
	runtime *bootstrap.Runtime
	repo    *Repo
}

func NewService(runtime *bootstrap.Runtime) *Service {
	return &Service{runtime: runtime, repo: NewRepo(runtime)}
}

func (s *Service) Register(req RegisterRequest) (RegisterResult, error) {
	req.Username = strings.TrimSpace(req.Username)
	req.Password = strings.TrimSpace(req.Password)
	req.Nickname = strings.TrimSpace(req.Nickname)
	if req.Username == "" {
		return RegisterResult{}, validationError("username is required")
	}
	if req.Password == "" {
		return RegisterResult{}, validationError("password is required")
	}
	if req.Nickname == "" {
		return RegisterResult{}, validationError("nickname is required")
	}
	exists, err := s.repo.UsernameExists(req.Username, 0)
	if err != nil {
		return RegisterResult{}, err
	}
	if exists {
		return RegisterResult{}, validationError("username already exists")
	}
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return RegisterResult{}, err
	}
	user := AppUser{
		Username:     req.Username,
		PasswordHash: string(passwordHash),
		Nickname:     req.Nickname,
		Status:       1,
		StatusText:   "正常",
		Version:      1,
	}
	if err := s.repo.Create(&user); err != nil {
		return RegisterResult{}, err
	}
	return RegisterResult{ID: user.ID, Username: user.Username, Nickname: user.Nickname}, nil
}

func (s *Service) Login(req LoginRequest, meta RequestMeta) (LoginResult, error) {
	req.Username = strings.TrimSpace(req.Username)
	if req.Username == "" {
		return LoginResult{}, validationError("username is required")
	}
	if strings.TrimSpace(req.Password) == "" {
		return LoginResult{}, validationError("password is required")
	}
	if err := s.runtime.CaptchaService.Verify(req.Captcha.ID, req.Captcha.Code); err != nil {
		return LoginResult{}, validationError(err.Error())
	}

	user, err := s.repo.FindByUsername(req.Username)
	if err != nil {
		return LoginResult{}, authError("用户名或密码错误")
	}
	if user.Status != 1 || bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)) != nil {
		return LoginResult{}, authError("用户名或密码错误")
	}

	expireAt := time.Now().Add(s.runtime.Config.JWT.App.Expire)
	session := AppUserSession{
		AppUserID:  user.ID,
		ExpiresAt:  expireAt,
		LastSeenAt: time.Now(),
		UserAgent:  meta.UserAgent,
		IP:         meta.IP,
	}
	if err := s.repo.CreateSession(&session); err != nil {
		return LoginResult{}, err
	}

	token, _, err := s.runtime.AppJWT.Issue(coreauth.IssuePayload{
		UserID:    user.ID,
		SessionID: session.ID,
		UserType:  "app_user",
		Username:  user.Username,
		Expire:    s.runtime.Config.JWT.App.Expire,
	})
	if err != nil {
		return LoginResult{}, err
	}
	return LoginResult{
		Token:       token,
		AccessToken: token,
		ExpiresAt:   expireAt,
		User: LoginUserBrief{
			ID:       user.ID,
			Username: user.Username,
			Nickname: user.Nickname,
		},
	}, nil
}

func (s *Service) Logout(sessionID int64) (LogoutResult, error) {
	if err := s.repo.DeleteSession(sessionID); err != nil {
		return LogoutResult{}, err
	}
	return LogoutResult{Logout: true}, nil
}

func (s *Service) Profile(userID int64) (ProfileResponse, error) {
	user, err := s.repo.FindByID(userID)
	if err != nil {
		return ProfileResponse{}, err
	}
	return ProfileResponse{
		ID:        user.ID,
		Username:  user.Username,
		Nickname:  user.Nickname,
		Email:     user.Email,
		Mobile:    user.Mobile,
		Avatar:    user.Avatar,
		Money:     user.Money,
		Score:     user.Score,
		Status:    user.Status,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}, nil
}

func (s *Service) SaveProfile(userID int64, req ProfileSaveRequest) (SaveResult, error) {
	req.Nickname = strings.TrimSpace(req.Nickname)
	req.Email = strings.TrimSpace(req.Email)
	req.Mobile = strings.TrimSpace(req.Mobile)
	req.Avatar = strings.TrimSpace(req.Avatar)
	if req.Nickname == "" {
		return SaveResult{}, validationError("nickname is required")
	}
	if err := s.repo.UpdateProfile(userID, map[string]any{
		"nickname":   req.Nickname,
		"email":      req.Email,
		"mobile":     req.Mobile,
		"avatar":     req.Avatar,
		"updated_at": time.Now(),
	}); err != nil {
		return SaveResult{}, err
	}
	return SaveResult{ID: userID}, nil
}

func (s *Service) ChangePassword(userID int64, req PasswordChangeRequest) (PasswordChangeResult, error) {
	req.OldPassword = strings.TrimSpace(req.OldPassword)
	req.NewPassword = strings.TrimSpace(req.NewPassword)
	if req.OldPassword == "" {
		return PasswordChangeResult{}, validationError("old_password is required")
	}
	if req.NewPassword == "" {
		return PasswordChangeResult{}, validationError("new_password is required")
	}
	user, err := s.repo.FindByID(userID)
	if err != nil {
		return PasswordChangeResult{}, err
	}
	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.OldPassword)) != nil {
		return PasswordChangeResult{}, validationError("old password is incorrect")
	}
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return PasswordChangeResult{}, err
	}
	if err := s.repo.UpdatePassword(user.ID, string(passwordHash), time.Now()); err != nil {
		return PasswordChangeResult{}, err
	}
	return PasswordChangeResult{Changed: true}, nil
}

func (s *Service) BackendList(params ListParams) (map[string]any, error) {
	if params.Page <= 0 {
		params.Page = 1
	}
	if params.PageSize <= 0 {
		params.PageSize = 20
	}
	list, total, err := s.repo.List(params)
	if err != nil {
		return nil, err
	}
	items := make([]UserListItem, 0, len(list))
	for _, user := range list {
		items = append(items, toUserListItem(user))
	}
	return bootstrap.PagedResult(items, total, params.Page, params.PageSize), nil
}

func (s *Service) BackendDetail(id int64) (UserListItem, error) {
	user, err := s.repo.FindByID(id)
	if err != nil {
		return UserListItem{}, err
	}
	return toUserListItem(user), nil
}

func (s *Service) BackendSave(req BackendSaveRequest) (SaveResult, error) {
	req.Username = strings.TrimSpace(req.Username)
	req.Password = strings.TrimSpace(req.Password)
	req.Nickname = strings.TrimSpace(req.Nickname)
	req.Email = strings.TrimSpace(req.Email)
	req.Mobile = strings.TrimSpace(req.Mobile)
	req.StatusText = strings.TrimSpace(req.StatusText)
	if req.Username == "" {
		return SaveResult{}, validationError("username is required")
	}
	if req.ID == 0 && req.Password == "" {
		return SaveResult{}, validationError("password is required")
	}
	exists, err := s.repo.UsernameExists(req.Username, req.ID)
	if err != nil {
		return SaveResult{}, err
	}
	if exists {
		return SaveResult{}, validationError("username already exists")
	}
	if req.Status == 0 {
		req.Status = 1
	}
	if req.Version == 0 {
		req.Version = 1
	}
	if req.StatusText == "" {
		req.StatusText = defaultStatusText(req.Status)
	}
	if req.ID == 0 {
		passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			return SaveResult{}, err
		}
		user := AppUser{
			PID:           req.PID,
			TID:           req.TID,
			Username:      req.Username,
			PasswordHash:  string(passwordHash),
			Nickname:      req.Nickname,
			Email:         req.Email,
			Mobile:        req.Mobile,
			Avatar:        req.Avatar,
			Code:          req.Code,
			Status:        req.Status,
			StatusText:    req.StatusText,
			Money:         req.Money,
			Score:         req.Score,
			WechatUnionID: req.WechatUnionID,
			WechatOpenID:  req.WechatOpenID,
			Version:       req.Version,
		}
		if err := s.repo.Create(&user); err != nil {
			return SaveResult{}, err
		}
		return SaveResult{ID: user.ID}, nil
	}
	updates := map[string]any{
		"pid":            req.PID,
		"tid":            req.TID,
		"username":       req.Username,
		"nickname":       req.Nickname,
		"email":          req.Email,
		"mobile":         req.Mobile,
		"avatar":         req.Avatar,
		"code":           req.Code,
		"status":         req.Status,
		"status_text":    req.StatusText,
		"wechat_unionid": req.WechatUnionID,
		"wechat_openid":  req.WechatOpenID,
		"version":        req.Version,
		"updated_at":     time.Now(),
	}
	if req.Password != "" {
		passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			return SaveResult{}, err
		}
		updates["password_hash"] = string(passwordHash)
	}
	if err := s.repo.Update(req.ID, updates); err != nil {
		return SaveResult{}, err
	}
	return SaveResult{ID: req.ID}, nil
}

func (s *Service) BackendDelete(ids []int64) (DeleteResult, error) {
	if len(ids) == 0 {
		return DeleteResult{}, validationError("ids is required")
	}
	deleted, err := s.repo.Delete(ids)
	if err != nil {
		return DeleteResult{}, err
	}
	return DeleteResult{Deleted: deleted}, nil
}

func (s *Service) BackendOperate(req OperateRequest) (OperateResult, error) {
	ids := bootstrap.NormalizeIDs(req.IDs...)
	if len(ids) == 0 {
		return OperateResult{}, validationError("ids is required")
	}
	field := strings.TrimSpace(req.Field)
	updates := map[string]any{"updated_at": time.Now()}
	switch field {
	case "status":
		status, err := intValue(req.Value)
		if err != nil {
			return OperateResult{}, err
		}
		updates["status"] = status
		updates["status_text"] = defaultStatusText(status)
	case "pid", "tid":
		value, err := int64Value(req.Value)
		if err != nil {
			return OperateResult{}, err
		}
		updates[field] = value
	default:
		return OperateResult{}, validationError("unsupported operate field")
	}
	updated, err := s.repo.UpdateMany(ids, updates)
	if err != nil {
		return OperateResult{}, err
	}
	return OperateResult{Updated: updated}, nil
}

func (s *Service) UpdateMoney(req MoneyRequest) (AssetUpdateResult, error) {
	kind, err := assetOperationType(req.Type)
	if err != nil {
		return AssetUpdateResult{}, err
	}
	if req.ID <= 0 {
		return AssetUpdateResult{}, validationError("id is required")
	}
	if req.Money <= 0 {
		return AssetUpdateResult{}, validationError("money must be greater than 0")
	}
	before, after, err := s.repo.UpdateMoneyWithLog(req.ID, kind, req.Money, req.Note, req.Source)
	if err != nil {
		return AssetUpdateResult{}, err
	}
	return AssetUpdateResult{UserID: req.ID, Before: before, After: after}, nil
}

func (s *Service) UpdateScore(req ScoreRequest) (AssetUpdateResult, error) {
	kind, err := assetOperationType(req.Type)
	if err != nil {
		return AssetUpdateResult{}, err
	}
	if req.ID <= 0 {
		return AssetUpdateResult{}, validationError("id is required")
	}
	if req.Score <= 0 {
		return AssetUpdateResult{}, validationError("score must be greater than 0")
	}
	before, after, err := s.repo.UpdateScoreWithLog(req.ID, kind, req.Score, req.Note, req.Source)
	if err != nil {
		return AssetUpdateResult{}, err
	}
	return AssetUpdateResult{UserID: req.ID, Before: before, After: after}, nil
}

func (s *Service) MoneyLogs(params LogListParams) (map[string]any, error) {
	normalizeLogPage(&params)
	list, total, err := s.repo.MoneyLogs(params)
	if err != nil {
		return nil, err
	}
	return bootstrap.PagedResult(list, total, params.Page, params.PageSize), nil
}

func (s *Service) ScoreLogs(params LogListParams) (map[string]any, error) {
	normalizeLogPage(&params)
	list, total, err := s.repo.ScoreLogs(params)
	if err != nil {
		return nil, err
	}
	return bootstrap.PagedResult(list, total, params.Page, params.PageSize), nil
}

type validationError string

func (e validationError) Error() string { return string(e) }

func isValidationError(err error) bool {
	_, ok := err.(validationError)
	return ok
}

type authError string

func (e authError) Error() string { return string(e) }

func isAuthError(err error) bool {
	_, ok := err.(authError)
	return ok
}

func toUserListItem(user AppUser) UserListItem {
	return UserListItem{
		ID:            user.ID,
		PID:           user.PID,
		TID:           user.TID,
		Username:      user.Username,
		Nickname:      user.Nickname,
		Email:         user.Email,
		Mobile:        user.Mobile,
		Avatar:        user.Avatar,
		Code:          user.Code,
		Status:        user.Status,
		StatusText:    user.StatusText,
		Money:         user.Money,
		Score:         user.Score,
		WechatUnionID: user.WechatUnionID,
		WechatOpenID:  user.WechatOpenID,
		Version:       user.Version,
		CreatedAt:     user.CreatedAt,
		UpdatedAt:     user.UpdatedAt,
	}
}

func defaultStatusText(status int) string {
	if status == 1 {
		return "正常"
	}
	return "禁用"
}

func assetOperationType(raw string) (int, error) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "", "add", "increase", "inc", "plus", "1":
		return 1, nil
	case "sub", "decrease", "dec", "minus", "2":
		return 2, nil
	default:
		return 0, validationError("unsupported operation type")
	}
}

func intValue(value any) (int, error) {
	v, err := int64Value(value)
	return int(v), err
}

func int64Value(value any) (int64, error) {
	switch typed := value.(type) {
	case int64:
		return typed, nil
	case int:
		return int64(typed), nil
	case float64:
		return int64(typed), nil
	case string:
		var result int64
		if _, err := fmt.Sscan(strings.TrimSpace(typed), &result); err != nil {
			return 0, validationError("invalid integer value")
		}
		return result, nil
	default:
		return 0, validationError("invalid integer value")
	}
}

func normalizeLogPage(params *LogListParams) {
	if params.Page <= 0 {
		params.Page = 1
	}
	if params.PageSize <= 0 {
		params.PageSize = 20
	}
}

func notFound(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}
