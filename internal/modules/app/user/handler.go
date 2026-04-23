package app_user

import (
	"github.com/lvjiaben/goweb-core/httpx"
	"github.com/lvjiaben/goweb-scaffold/internal/bootstrap"
)

func profile(runtime *bootstrap.Runtime) httpx.HandlerFunc {
	service := NewService(runtime)
	return func(c *httpx.Context) {
		user, ok := currentUser(c)
		if !ok {
			c.Unauthorized("app user missing")
			return
		}
		c.Success(service.Profile(user))
	}
}

func profileSave(runtime *bootstrap.Runtime) httpx.HandlerFunc {
	service := NewService(runtime)
	return func(c *httpx.Context) {
		var req ProfileSaveRequest
		if err := c.BindJSON(&req); err != nil {
			c.Error(err)
			return
		}
		user, ok := currentUser(c)
		if !ok {
			c.Unauthorized("app user missing")
			return
		}
		result, err := service.SaveProfile(user.ID, req)
		if err != nil {
			respondServiceError(c, err)
			return
		}
		c.Success(result)
	}
}

func passwordChange(runtime *bootstrap.Runtime) httpx.HandlerFunc {
	service := NewService(runtime)
	return func(c *httpx.Context) {
		var req PasswordChangeRequest
		if err := c.BindJSON(&req); err != nil {
			c.Error(err)
			return
		}
		user, ok := currentUser(c)
		if !ok {
			c.Unauthorized("app user missing")
			return
		}
		result, err := service.ChangePassword(user, req)
		if err != nil {
			respondServiceError(c, err)
			return
		}
		c.Success(result)
	}
}

func currentUser(c *httpx.Context) (CurrentUser, bool) {
	user, ok := bootstrap.CurrentAppUser(c)
	if !ok {
		return CurrentUser{}, false
	}
	return CurrentUser{
		ID:           user.ID,
		Username:     user.Username,
		PasswordHash: user.PasswordHash,
		Nickname:     user.Nickname,
		Email:        user.Email,
		Mobile:       user.Mobile,
		Status:       user.Status,
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
	}, true
}

func respondServiceError(c *httpx.Context, err error) {
	if isValidationError(err) {
		c.BadRequest(err.Error())
		return
	}
	c.Error(err)
}
