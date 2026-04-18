package app_user

import (
	"time"

	"github.com/lvjiaben/goweb-core/httpx"
	"github.com/lvjiaben/goweb-scaffold/internal/bootstrap"
	"github.com/lvjiaben/goweb-scaffold/internal/modules/model"
	"golang.org/x/crypto/bcrypt"
)

type Module struct{}

type profileSaveRequest struct {
	Nickname string `json:"nickname" validate:"required"`
	Email    string `json:"email"`
	Mobile   string `json:"mobile"`
}

type passwordChangeRequest struct {
	OldPassword string `json:"old_password" validate:"required"`
	NewPassword string `json:"new_password" validate:"required"`
}

func (Module) Name() string { return "app_user" }

func (Module) Register(runtime *bootstrap.Runtime) error {
	runtime.AppUserGroup.GET("/profile", profile(runtime))
	runtime.AppUserGroup.POST("/profile/save", profileSave(runtime))
	runtime.AppUserGroup.POST("/password/change", passwordChange(runtime))
	return nil
}

func profile(runtime *bootstrap.Runtime) httpx.HandlerFunc {
	return func(c *httpx.Context) {
		user, ok := bootstrap.CurrentAppUser(c)
		if !ok {
			c.Unauthorized("app user missing")
			return
		}
		c.Success(map[string]any{
			"id":         user.ID,
			"username":   user.Username,
			"nickname":   user.Nickname,
			"email":      user.Email,
			"mobile":     user.Mobile,
			"status":     user.Status,
			"created_at": user.CreatedAt,
			"updated_at": user.UpdatedAt,
		})
	}
}

func profileSave(runtime *bootstrap.Runtime) httpx.HandlerFunc {
	return func(c *httpx.Context) {
		var req profileSaveRequest
		if err := c.BindJSON(&req); err != nil {
			c.Error(err)
			return
		}
		if err := runtime.Validator.Struct(req); err != nil {
			c.BadRequest(err.Error())
			return
		}

		user, ok := bootstrap.CurrentAppUser(c)
		if !ok {
			c.Unauthorized("app user missing")
			return
		}

		if err := runtime.DB.Model(&model.AppUser{}).Where("id = ?", user.ID).Updates(map[string]any{
			"nickname":   req.Nickname,
			"email":      req.Email,
			"mobile":     req.Mobile,
			"updated_at": time.Now(),
		}).Error; err != nil {
			c.Error(err)
			return
		}
		c.Success(map[string]any{"id": user.ID})
	}
}

func passwordChange(runtime *bootstrap.Runtime) httpx.HandlerFunc {
	return func(c *httpx.Context) {
		var req passwordChangeRequest
		if err := c.BindJSON(&req); err != nil {
			c.Error(err)
			return
		}
		if err := runtime.Validator.Struct(req); err != nil {
			c.BadRequest(err.Error())
			return
		}

		user, ok := bootstrap.CurrentAppUser(c)
		if !ok {
			c.Unauthorized("app user missing")
			return
		}
		if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.OldPassword)) != nil {
			c.BadRequest("old password is incorrect")
			return
		}

		passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
		if err != nil {
			c.Error(err)
			return
		}

		if err := runtime.DB.Model(&model.AppUser{}).Where("id = ?", user.ID).Updates(map[string]any{
			"password_hash": string(passwordHash),
			"updated_at":    time.Now(),
		}).Error; err != nil {
			c.Error(err)
			return
		}
		c.Success(map[string]any{"changed": true})
	}
}
