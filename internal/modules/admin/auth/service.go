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

func (s *Service) Menus(ctx context.Context, identity *corerbac.Identity) (map[string]any, error) {
	menuItems, err := s.runtime.PermissionService.GetMenus(ctx, identity)
	if err != nil {
		return nil, err
	}
	return map[string]any{"list": menuItems}, nil
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
