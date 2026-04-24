package admin_auth

import (
	"strconv"

	"github.com/lvjiaben/goweb-core/httpx"
	corerbac "github.com/lvjiaben/goweb-core/rbac"
	"github.com/lvjiaben/goweb-scaffold/internal/bootstrap"
	admin_menu "github.com/lvjiaben/goweb-scaffold/internal/modules/admin/menu"
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
	menuService := admin_menu.NewService(runtime)
	return func(c *httpx.Context) {
		identity, _ := corerbac.GetIdentity(c)
		routes, err := menuService.GetVbenRoutes(c.Request.Context(), identity, c.Request.Header.Get("Accept-Language"))
		if err != nil {
			c.Error(err)
			return
		}
		c.Success(map[string]any{"list": routes})
	}
}

func profile(runtime *bootstrap.Runtime) httpx.HandlerFunc {
	service := NewService(runtime)
	return func(c *httpx.Context) {
		user, ok := bootstrap.CurrentAdminUser(c)
		if !ok {
			c.Unauthorized("admin user missing")
			return
		}
		var req ProfileRequest
		if err := c.BindJSON(&req); err != nil {
			c.Error(err)
			return
		}
		result, err := service.UpdateProfile(user.ID, req)
		if err != nil {
			respondServiceError(c, err)
			return
		}
		c.Success(result)
	}
}

func password(runtime *bootstrap.Runtime) httpx.HandlerFunc {
	service := NewService(runtime)
	return func(c *httpx.Context) {
		user, ok := bootstrap.CurrentAdminUser(c)
		if !ok {
			c.Unauthorized("admin user missing")
			return
		}
		var req PasswordRequest
		if err := c.BindJSON(&req); err != nil {
			c.Error(err)
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

func logs(runtime *bootstrap.Runtime) httpx.HandlerFunc {
	service := NewService(runtime)
	return func(c *httpx.Context) {
		user, ok := bootstrap.CurrentAdminUser(c)
		if !ok {
			c.Unauthorized("admin user missing")
			return
		}
		params := LogParams{
			Page:     queryInt(c.Query("page"), 1),
			PageSize: queryInt(prefer(c.Query("page_size"), c.Query("limit")), 20),
		}
		result, err := service.Logs(user.ID, params)
		if err != nil {
			c.Error(err)
			return
		}
		c.Success(result)
	}
}

func queryInt(value string, fallback int) int {
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func prefer(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
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
