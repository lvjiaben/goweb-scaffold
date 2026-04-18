package app_user_auth

import (
	"time"

	coreauth "github.com/lvjiaben/goweb-core/auth"
	"github.com/lvjiaben/goweb-core/httpx"
	"github.com/lvjiaben/goweb-scaffold/internal/bootstrap"
	"github.com/lvjiaben/goweb-scaffold/internal/modules/model"
	"golang.org/x/crypto/bcrypt"
)

type Module struct{}

type registerRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
	Nickname string `json:"nickname" validate:"required"`
}

type loginRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

func (Module) Name() string { return "app_user_auth" }

func (Module) Register(runtime *bootstrap.Runtime) error {
	runtime.AppPublicGroup.POST("/login", login(runtime))
	runtime.AppPublicGroup.POST("/register", register(runtime))
	runtime.AppAuthGroup.POST("/logout", logout(runtime))
	return nil
}

func register(runtime *bootstrap.Runtime) httpx.HandlerFunc {
	return func(c *httpx.Context) {
		var req registerRequest
		if err := c.BindJSON(&req); err != nil {
			c.Error(err)
			return
		}
		if err := runtime.Validator.Struct(req); err != nil {
			c.BadRequest(err.Error())
			return
		}

		passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			c.Error(err)
			return
		}

		user := model.AppUser{
			Username:     req.Username,
			PasswordHash: string(passwordHash),
			Nickname:     req.Nickname,
			Status:       1,
		}
		if err := runtime.DB.Create(&user).Error; err != nil {
			c.Error(err)
			return
		}

		c.Success(map[string]any{
			"id":       user.ID,
			"username": user.Username,
			"nickname": user.Nickname,
		})
	}
}

func login(runtime *bootstrap.Runtime) httpx.HandlerFunc {
	return func(c *httpx.Context) {
		var req loginRequest
		if err := c.BindJSON(&req); err != nil {
			c.Error(err)
			return
		}
		if err := runtime.Validator.Struct(req); err != nil {
			c.BadRequest(err.Error())
			return
		}

		var user model.AppUser
		if err := runtime.DB.Where("username = ? AND deleted_at IS NULL", req.Username).First(&user).Error; err != nil {
			c.Unauthorized("用户名或密码错误")
			return
		}
		if user.Status != 1 || bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)) != nil {
			c.Unauthorized("用户名或密码错误")
			return
		}

		expireAt := time.Now().Add(runtime.Config.JWT.App.Expire)
		session := model.AppUserSession{
			AppUserID:  user.ID,
			ExpiresAt:  expireAt,
			LastSeenAt: time.Now(),
			UserAgent:  c.Request.UserAgent(),
			IP:         c.ClientIP(),
		}
		if err := runtime.DB.Create(&session).Error; err != nil {
			c.Error(err)
			return
		}

		token, _, err := runtime.AppJWT.Issue(coreauth.IssuePayload{
			UserID:    user.ID,
			SessionID: session.ID,
			UserType:  "app_user",
			Username:  user.Username,
			Expire:    runtime.Config.JWT.App.Expire,
		})
		if err != nil {
			c.Error(err)
			return
		}

		c.Success(map[string]any{
			"token":      token,
			"expires_at": expireAt,
			"user": map[string]any{
				"id":       user.ID,
				"username": user.Username,
				"nickname": user.Nickname,
			},
		})
	}
}

func logout(runtime *bootstrap.Runtime) httpx.HandlerFunc {
	return func(c *httpx.Context) {
		claims, ok := bootstrap.CurrentAppClaims(c)
		if !ok {
			c.Unauthorized("app claims missing")
			return
		}
		if err := runtime.DB.Where("id = ?", claims.SessionID).Delete(&model.AppUserSession{}).Error; err != nil {
			c.Error(err)
			return
		}
		c.Success(map[string]any{"logout": true})
	}
}
