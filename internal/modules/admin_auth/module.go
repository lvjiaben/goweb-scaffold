package admin_auth

import (
	"time"

	coreauth "github.com/lvjiaben/goweb-core/auth"
	"github.com/lvjiaben/goweb-core/httpx"
	corerbac "github.com/lvjiaben/goweb-core/rbac"
	"github.com/lvjiaben/goweb-scaffold/internal/bootstrap"
	"github.com/lvjiaben/goweb-scaffold/internal/modules/model"
	"golang.org/x/crypto/bcrypt"
)

type Module struct{}

type captchaRequest struct {
	ID   string `json:"id"`
	Code string `json:"code"`
}

type loginRequest struct {
	Username string         `json:"username" validate:"required"`
	Password string         `json:"password" validate:"required"`
	Captcha  captchaRequest `json:"captcha"`
}

func (Module) Name() string { return "admin_auth" }

func (Module) Register(runtime *bootstrap.Runtime) error {
	runtime.AdminPublicGroup.POST("/login", login(runtime))
	runtime.AdminAuthGroup.POST("/logout", logout(runtime))
	runtime.AdminAuthGroup.GET("/me", me(runtime))
	runtime.AdminAuthGroup.GET("/menus", menus(runtime))
	return nil
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
		if err := runtime.CaptchaService.Verify(req.Captcha.ID, req.Captcha.Code); err != nil {
			recordLogin(runtime, 0, req.Username, c, false, err.Error())
			c.BadRequest(err.Error())
			return
		}

		var user model.AdminUser
		if err := runtime.DB.Where("username = ? AND deleted_at IS NULL", req.Username).First(&user).Error; err != nil {
			recordLogin(runtime, 0, req.Username, c, false, "用户名或密码错误")
			c.Unauthorized("用户名或密码错误")
			return
		}
		if user.Status != 1 || bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)) != nil {
			recordLogin(runtime, user.ID, req.Username, c, false, "用户名或密码错误")
			c.Unauthorized("用户名或密码错误")
			return
		}

		expireAt := time.Now().Add(runtime.Config.JWT.Admin.Expire)
		session := model.AdminSession{
			AdminUserID: user.ID,
			ExpiresAt:   expireAt,
			LastSeenAt:  time.Now(),
			UserAgent:   c.Request.UserAgent(),
			IP:          c.ClientIP(),
		}
		if err := runtime.DB.Create(&session).Error; err != nil {
			c.Error(err)
			return
		}

		token, _, err := runtime.AdminJWT.Issue(coreauth.IssuePayload{
			UserID:    user.ID,
			SessionID: session.ID,
			UserType:  "admin",
			Username:  user.Username,
			Expire:    runtime.Config.JWT.Admin.Expire,
		})
		if err != nil {
			c.Error(err)
			return
		}

		now := time.Now()
		_ = runtime.DB.Model(&model.AdminUser{}).Where("id = ?", user.ID).Updates(map[string]any{
			"last_login_at": now,
			"last_login_ip": c.ClientIP(),
			"updated_at":    now,
		}).Error

		recordLogin(runtime, user.ID, req.Username, c, true, "登录成功")
		c.Success(map[string]any{
			"token":      token,
			"expires_at": expireAt,
			"user": map[string]any{
				"id":       user.ID,
				"username": user.Username,
				"nickname": user.Nickname,
				"is_super": user.IsSuper,
			},
		})
	}
}

func logout(runtime *bootstrap.Runtime) httpx.HandlerFunc {
	return func(c *httpx.Context) {
		claims, ok := bootstrap.CurrentAdminClaims(c)
		if !ok {
			c.Unauthorized("admin claims missing")
			return
		}
		if err := runtime.DB.Where("id = ?", claims.SessionID).Delete(&model.AdminSession{}).Error; err != nil {
			c.Error(err)
			return
		}
		c.Success(map[string]any{"logout": true})
	}
}

func me(runtime *bootstrap.Runtime) httpx.HandlerFunc {
	return func(c *httpx.Context) {
		user, ok := bootstrap.CurrentAdminUser(c)
		if !ok {
			c.Unauthorized("admin user missing")
			return
		}
		identity, _ := corerbac.GetIdentity(c)
		accessCodes, err := runtime.PermissionService.GetAccessCodes(c.Request.Context(), identity)
		if err != nil {
			c.Error(err)
			return
		}
		c.Success(map[string]any{
			"id":           user.ID,
			"username":     user.Username,
			"nickname":     user.Nickname,
			"is_super":     user.IsSuper,
			"role_ids":     identity.RoleIDs,
			"access_codes": accessCodes,
		})
	}
}

func menus(runtime *bootstrap.Runtime) httpx.HandlerFunc {
	return func(c *httpx.Context) {
		identity, _ := corerbac.GetIdentity(c)
		menuItems, err := runtime.PermissionService.GetMenus(c.Request.Context(), identity)
		if err != nil {
			c.Error(err)
			return
		}
		c.Success(map[string]any{"list": menuItems})
	}
}

func recordLogin(runtime *bootstrap.Runtime, adminUserID int64, username string, c *httpx.Context, success bool, remark string) {
	_ = runtime.DB.Create(&model.AdminLoginLog{
		AdminUserID: adminUserID,
		Username:    username,
		IP:          c.ClientIP(),
		UserAgent:   c.Request.UserAgent(),
		Success:     success,
		Remark:      remark,
	}).Error
}
