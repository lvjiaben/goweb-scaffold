package bootstrap

import (
	"errors"
	"net/http"
	"strings"
	"time"

	coreauth "github.com/lvjiaben/goweb-core/auth"
	"github.com/lvjiaben/goweb-core/httpx"
	corerbac "github.com/lvjiaben/goweb-core/rbac"
	"gorm.io/gorm"
)

const (
	adminUserContextKey  = "admin_user"
	adminClaimContextKey = "admin_claims"
	appUserContextKey    = "app_user"
	appClaimContextKey   = "app_claims"
)

func (r *Runtime) AdminAuthMiddleware() httpx.Middleware {
	return func(next httpx.HandlerFunc) httpx.HandlerFunc {
		return func(c *httpx.Context) {
			token := bearerToken(c)
			if token == "" {
				authExpired(c, "missing admin token")
				return
			}

			claims, err := r.AdminJWT.Parse(token)
			if err != nil || claims.UserType != "admin" {
				authExpired(c, "invalid admin token")
				return
			}

			var session AdminSession
			if err := r.DB.Where("id = ? AND admin_user_id = ? AND expires_at > ?", claims.SessionID, claims.UserID, time.Now()).
				First(&session).Error; err != nil {
				authExpired(c, "admin session expired")
				return
			}

			var user AdminUser
			if err := r.DB.Where("id = ? AND status = ?", claims.UserID, 1).First(&user).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					authExpired(c, "admin user not found")
					return
				}
				c.Error(err)
				return
			}

			var roleIDs []int64
			if err := r.DB.Model(&AdminUserRole{}).Where("user_id = ?", user.ID).Pluck("role_id", &roleIDs).Error; err != nil {
				c.Error(err)
				return
			}

			corerbac.SetIdentity(c, &corerbac.Identity{
				UserID:   user.ID,
				UserType: "admin",
				RoleIDs:  roleIDs,
				IsSuper:  user.IsSuper,
			})
			c.Set(adminUserContextKey, &user)
			c.Set(adminClaimContextKey, claims)

			_ = r.DB.Model(&AdminSession{}).Where("id = ?", session.ID).Update("last_seen_at", time.Now()).Error
			next(c)
		}
	}
}

func (r *Runtime) AppUserAuthMiddleware() httpx.Middleware {
	return func(next httpx.HandlerFunc) httpx.HandlerFunc {
		return func(c *httpx.Context) {
			token := bearerToken(c)
			if token == "" {
				authExpired(c, "missing app token")
				return
			}

			claims, err := r.AppJWT.Parse(token)
			if err != nil || claims.UserType != "app_user" {
				authExpired(c, "invalid app token")
				return
			}

			var session AppUserSession
			if err := r.DB.Where("id = ? AND app_user_id = ? AND expires_at > ?", claims.SessionID, claims.UserID, time.Now()).
				First(&session).Error; err != nil {
				authExpired(c, "app session expired")
				return
			}

			var user AppUser
			if err := r.DB.Where("id = ? AND status = ?", claims.UserID, 1).First(&user).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					authExpired(c, "app user not found")
					return
				}
				c.Error(err)
				return
			}

			c.Set(appUserContextKey, &user)
			c.Set(appClaimContextKey, claims)

			_ = r.DB.Model(&AppUserSession{}).Where("id = ?", session.ID).Update("last_seen_at", time.Now()).Error
			next(c)
		}
	}
}

func CurrentAdminUser(c *httpx.Context) (*AdminUser, bool) {
	value, ok := c.Get(adminUserContextKey)
	if !ok {
		return nil, false
	}
	user, ok := value.(*AdminUser)
	return user, ok
}

func CurrentAdminClaims(c *httpx.Context) (*coreauth.Claims, bool) {
	value, ok := c.Get(adminClaimContextKey)
	if !ok {
		return nil, false
	}
	claims, ok := value.(*coreauth.Claims)
	return claims, ok
}

func CurrentAppUser(c *httpx.Context) (*AppUser, bool) {
	value, ok := c.Get(appUserContextKey)
	if !ok {
		return nil, false
	}
	user, ok := value.(*AppUser)
	return user, ok
}

func CurrentAppClaims(c *httpx.Context) (*coreauth.Claims, bool) {
	value, ok := c.Get(appClaimContextKey)
	if !ok {
		return nil, false
	}
	claims, ok := value.(*coreauth.Claims)
	return claims, ok
}

func bearerToken(c *httpx.Context) string {
	return strings.TrimSpace(c.Request.Header.Get("Authorization"))
}

func authExpired(c *httpx.Context, message string) {
	c.JSON(http.StatusOK, http.StatusUnauthorized, message, map[string]any{})
}
