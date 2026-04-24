package admin_auth

import (
	"context"
	"strings"
	"time"

	coreauth "github.com/lvjiaben/goweb-core/auth"
	corerbac "github.com/lvjiaben/goweb-core/rbac"
	"github.com/lvjiaben/goweb-scaffold/internal/bootstrap"
	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	runtime *bootstrap.Runtime
	repo    *Repo
}

func NewService(runtime *bootstrap.Runtime) *Service {
	return &Service{runtime: runtime, repo: NewRepo(runtime)}
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
		s.recordLogin(0, req.Username, meta, false, err.Error())
		return LoginResult{}, validationError(err.Error())
	}

	user, err := s.repo.FindUserByUsername(req.Username)
	if err != nil {
		s.recordLogin(0, req.Username, meta, false, "用户名或密码错误")
		return LoginResult{}, authError("用户名或密码错误")
	}
	if user.Status != 1 || bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)) != nil {
		s.recordLogin(user.ID, req.Username, meta, false, "用户名或密码错误")
		return LoginResult{}, authError("用户名或密码错误")
	}

	expireAt := time.Now().Add(s.runtime.Config.JWT.Admin.Expire)
	session := AdminSession{
		AdminUserID: user.ID,
		ExpiresAt:   expireAt,
		LastSeenAt:  time.Now(),
		UserAgent:   meta.UserAgent,
		IP:          meta.IP,
	}
	if err := s.repo.CreateSession(&session); err != nil {
		return LoginResult{}, err
	}

	token, _, err := s.runtime.AdminJWT.Issue(coreauth.IssuePayload{
		UserID:    user.ID,
		SessionID: session.ID,
		UserType:  "admin",
		Username:  user.Username,
		Expire:    s.runtime.Config.JWT.Admin.Expire,
	})
	if err != nil {
		return LoginResult{}, err
	}

	now := time.Now()
	_ = s.repo.UpdateLastLogin(user.ID, meta.IP, now)
	s.recordLogin(user.ID, req.Username, meta, true, "登录成功")

	return LoginResult{
		Token:     token,
		ExpiresAt: expireAt,
		User: LoginUserBrief{
			ID:       user.ID,
			Username: user.Username,
			Nickname: user.Nickname,
			IsSuper:  user.IsSuper,
		},
	}, nil
}

func (s *Service) Logout(sessionID int64) (LogoutResult, error) {
	if err := s.repo.DeleteSession(sessionID); err != nil {
		return LogoutResult{}, err
	}
	return LogoutResult{Logout: true}, nil
}

func (s *Service) Me(ctx context.Context, user CurrentAdmin, identity *corerbac.Identity) (MeResponse, error) {
	accessCodes, err := s.runtime.PermissionService.GetAccessCodes(ctx, identity)
	if err != nil {
		return MeResponse{}, err
	}
	return MeResponse{
		ID:          user.ID,
		Username:    user.Username,
		Nickname:    user.Nickname,
		IsSuper:     user.IsSuper,
		RoleIDs:     identity.RoleIDs,
		AccessCodes: accessCodes,
	}, nil
}

func (s *Service) UpdateProfile(userID int64, req ProfileRequest) (map[string]any, error) {
	nickname := strings.TrimSpace(req.Nickname)
	if nickname == "" {
		nickname = strings.TrimSpace(req.RealName)
	}
	if err := s.repo.UpdateProfile(userID, nickname); err != nil {
		return nil, err
	}
	user, err := s.repo.FindUserByID(userID)
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"id":       user.ID,
		"username": user.Username,
		"nickname": user.Nickname,
		"avatar":   req.Avatar,
		"email":    req.Email,
	}, nil
}

func (s *Service) ChangePassword(userID int64, req PasswordRequest) (map[string]any, error) {
	req.OldPassword = strings.TrimSpace(req.OldPassword)
	req.NewPassword = strings.TrimSpace(req.NewPassword)
	if req.OldPassword == "" || req.NewPassword == "" {
		return nil, validationError("old_password and new_password are required")
	}
	user, err := s.repo.FindUserByID(userID)
	if err != nil {
		return nil, err
	}
	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.OldPassword)) != nil {
		return nil, validationError("旧密码错误")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	if err := s.repo.UpdatePasswordHash(userID, string(hash)); err != nil {
		return nil, err
	}
	return map[string]any{"changed": true}, nil
}

func (s *Service) Logs(userID int64, params LogParams) (map[string]any, error) {
	logs, total, err := s.repo.LoginLogs(userID, params.Page, params.PageSize)
	if err != nil {
		return nil, err
	}
	items := make([]LogItem, 0, len(logs))
	for _, item := range logs {
		status := 0
		if item.Success {
			status = 1
		}
		items = append(items, LogItem{
			ID:        item.ID,
			Username:  item.Username,
			IP:        item.IP,
			Status:    status,
			Remark:    item.Remark,
			CreatedAt: item.CreatedAt.Unix(),
		})
	}
	return bootstrap.PagedResult(items, total, params.Page, params.PageSize), nil
}

func (s *Service) recordLogin(adminUserID int64, username string, meta RequestMeta, success bool, remark string) {
	_ = s.repo.CreateLoginLog(&AdminLoginLog{
		AdminUserID: adminUserID,
		Username:    username,
		IP:          meta.IP,
		UserAgent:   meta.UserAgent,
		Success:     success,
		Remark:      remark,
	})
}

type authError string

func (e authError) Error() string {
	return string(e)
}

func isAuthError(err error) bool {
	_, ok := err.(authError)
	return ok
}
