package app_user_auth

import (
	"strings"
	"time"

	coreauth "github.com/lvjiaben/goweb-core/auth"
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
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return RegisterResult{}, err
	}
	user := AppUser{
		Username:     req.Username,
		PasswordHash: string(passwordHash),
		Nickname:     req.Nickname,
		Status:       1,
	}
	if err := s.repo.CreateUser(&user); err != nil {
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

	user, err := s.repo.FindUserByUsername(req.Username)
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

type authError string

func (e authError) Error() string {
	return string(e)
}

func isAuthError(err error) bool {
	_, ok := err.(authError)
	return ok
}
