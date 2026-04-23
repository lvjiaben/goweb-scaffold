package admin_auth

import (
	"github.com/lvjiaben/goweb-core/httpx"
	corerbac "github.com/lvjiaben/goweb-core/rbac"
	"github.com/lvjiaben/goweb-scaffold/internal/bootstrap"
)

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
		claims, ok := bootstrap.CurrentAdminClaims(c)
		if !ok {
			c.Unauthorized("admin claims missing")
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

func me(runtime *bootstrap.Runtime) httpx.HandlerFunc {
	service := NewService(runtime)
	return func(c *httpx.Context) {
		user, ok := bootstrap.CurrentAdminUser(c)
		if !ok {
			c.Unauthorized("admin user missing")
			return
		}
		identity, _ := corerbac.GetIdentity(c)
		result, err := service.Me(c.Request.Context(), CurrentAdmin{
			ID:       user.ID,
			Username: user.Username,
			Nickname: user.Nickname,
			IsSuper:  user.IsSuper,
		}, identity)
		if err != nil {
			c.Error(err)
			return
		}
		c.Success(result)
	}
}

func menus(runtime *bootstrap.Runtime) httpx.HandlerFunc {
	service := NewService(runtime)
	return func(c *httpx.Context) {
		identity, _ := corerbac.GetIdentity(c)
		result, err := service.Menus(c.Request.Context(), identity)
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
