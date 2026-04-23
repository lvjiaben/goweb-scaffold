package app_user_auth

import (
	"github.com/lvjiaben/goweb-core/httpx"
	"github.com/lvjiaben/goweb-scaffold/internal/bootstrap"
)

func register(runtime *bootstrap.Runtime) httpx.HandlerFunc {
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

func login(runtime *bootstrap.Runtime) httpx.HandlerFunc {
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

func logout(runtime *bootstrap.Runtime) httpx.HandlerFunc {
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
