package app_user

import (
	"github.com/lvjiaben/goweb-core/httpx"
	"github.com/lvjiaben/goweb-scaffold/internal/bootstrap"
)

func registerAPI(runtime *bootstrap.Runtime) error {
	publicAuth := runtime.AppPublicGroup.Group("/auth")
	authedAuth := runtime.AppAuthedGroup.Group("/auth")
	publicAuth.POST("/login", apiLogin(runtime))
	publicAuth.POST("/register", apiRegister(runtime))
	authedAuth.POST("/logout", apiLogout(runtime))

	publicUser := runtime.AppPublicGroup.Group("/user")
	protectedUser := runtime.AppProtectedGroup.Group("/user")
	publicUser.POST("/login", apiLogin(runtime))
	protectedUser.POST("/logout", apiLogout(runtime))
	protectedUser.GET("/profile", apiProfile(runtime))
	protectedUser.POST("/profile/save", apiProfileSave(runtime))
	protectedUser.POST("/password/change", apiPasswordChange(runtime))
	return nil
}

func apiRegister(runtime *bootstrap.Runtime) httpx.HandlerFunc {
	service := NewService(runtime)
	return func(c *httpx.Context) {
		var req RegisterRequest
		if err := c.BindJSON(&req); err != nil {
			c.Error(err)
			return
		}
		result, err := service.Register(req)
		if err != nil {
			respondServiceError(c, err)
			return
		}
		c.Success(result)
	}
}

func apiLogin(runtime *bootstrap.Runtime) httpx.HandlerFunc {
	service := NewService(runtime)
	return func(c *httpx.Context) {
		var req LoginRequest
		if err := c.BindJSON(&req); err != nil {
			c.Error(err)
			return
		}
		result, err := service.Login(req, RequestMeta{
			IP:        c.ClientIP(),
			UserAgent: c.Request.UserAgent(),
		})
		if err != nil {
			respondServiceError(c, err)
			return
		}
		c.Success(result)
	}
}

func apiLogout(runtime *bootstrap.Runtime) httpx.HandlerFunc {
	service := NewService(runtime)
	return func(c *httpx.Context) {
		claims, ok := bootstrap.CurrentAppClaims(c)
		if !ok {
			c.Unauthorized("app claims missing")
			return
		}
		result, err := service.Logout(claims.SessionID)
		if err != nil {
			c.Error(err)
			return
		}
		c.Success(result)
	}
}

func apiProfile(runtime *bootstrap.Runtime) httpx.HandlerFunc {
	service := NewService(runtime)
	return func(c *httpx.Context) {
		user, ok := bootstrap.CurrentAppUser(c)
		if !ok {
			c.Unauthorized("app user missing")
			return
		}
		result, err := service.Profile(user.ID)
		if err != nil {
			respondServiceError(c, err)
			return
		}
		c.Success(result)
	}
}

func apiProfileSave(runtime *bootstrap.Runtime) httpx.HandlerFunc {
	service := NewService(runtime)
	return func(c *httpx.Context) {
		var req ProfileSaveRequest
		if err := c.BindJSON(&req); err != nil {
			c.Error(err)
			return
		}
		user, ok := bootstrap.CurrentAppUser(c)
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

func apiPasswordChange(runtime *bootstrap.Runtime) httpx.HandlerFunc {
	service := NewService(runtime)
	return func(c *httpx.Context) {
		var req PasswordChangeRequest
		if err := c.BindJSON(&req); err != nil {
			c.Error(err)
			return
		}
		user, ok := bootstrap.CurrentAppUser(c)
		if !ok {
			c.Unauthorized("app user missing")
			return
		}
		result, err := service.ChangePassword(user.ID, req)
		if err != nil {
			respondServiceError(c, err)
			return
		}
		c.Success(result)
	}
}

func respondServiceError(c *httpx.Context, err error) {
	switch {
	case isValidationError(err):
		c.BadRequest(err.Error())
	case isAuthError(err):
		c.Unauthorized(err.Error())
	default:
		c.Error(err)
	}
}
