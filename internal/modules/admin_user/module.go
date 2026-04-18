package admin_user

import (
	"time"

	"github.com/lvjiaben/goweb-core/httpx"
	"github.com/lvjiaben/goweb-scaffold/internal/bootstrap"
	"github.com/lvjiaben/goweb-scaffold/internal/modules/model"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type Module struct{}

type saveRequest struct {
	ID       int64   `json:"id"`
	Username string  `json:"username" validate:"required"`
	Password string  `json:"password"`
	Nickname string  `json:"nickname" validate:"required"`
	Status   int     `json:"status"`
	IsSuper  bool    `json:"is_super"`
	RoleIDs  []int64 `json:"role_ids"`
}

func (Module) Name() string { return "admin_user" }

func (Module) Register(runtime *bootstrap.Runtime) error {
	runtime.AdminProtectedGroup.GET("/admin_user/list", list(runtime), httpx.WithPermission("admin_user.list"))
	runtime.AdminProtectedGroup.GET("/admin_user/detail", detail(runtime), httpx.WithPermission("admin_user.list"))
	runtime.AdminProtectedGroup.POST("/admin_user/save", save(runtime), httpx.WithPermission("admin_user.save"))
	runtime.AdminProtectedGroup.POST("/admin_user/delete", deleteUsers(runtime), httpx.WithPermission("admin_user.delete"))
	return nil
}

func list(runtime *bootstrap.Runtime) httpx.HandlerFunc {
	return func(c *httpx.Context) {
		page, pageSize := bootstrap.Pagination(c)
		keyword := bootstrap.LikeKeyword(c.Query("keyword"))

		query := runtime.DB.Model(&model.AdminUser{}).Order("id DESC")
		if keyword != "" {
			query = query.Where("username ILIKE ? OR nickname ILIKE ?", keyword, keyword)
		}

		var total int64
		if err := query.Count(&total).Error; err != nil {
			c.Error(err)
			return
		}

		var users []model.AdminUser
		if err := query.Offset((page - 1) * pageSize).Limit(pageSize).Find(&users).Error; err != nil {
			c.Error(err)
			return
		}

		userIDs := make([]int64, 0, len(users))
		for _, user := range users {
			userIDs = append(userIDs, user.ID)
		}

		roleMap := make(map[int64][]int64, len(userIDs))
		if len(userIDs) > 0 {
			var mappings []model.AdminUserRole
			if err := runtime.DB.Where("user_id IN ?", userIDs).Find(&mappings).Error; err != nil {
				c.Error(err)
				return
			}
			for _, item := range mappings {
				roleMap[item.UserID] = append(roleMap[item.UserID], item.RoleID)
			}
		}

		items := make([]map[string]any, 0, len(users))
		for _, user := range users {
			items = append(items, map[string]any{
				"id":         user.ID,
				"username":   user.Username,
				"nickname":   user.Nickname,
				"status":     user.Status,
				"is_super":   user.IsSuper,
				"role_ids":   roleMap[user.ID],
				"created_at": user.CreatedAt,
				"updated_at": user.UpdatedAt,
			})
		}

		c.Success(map[string]any{
			"list":      items,
			"total":     total,
			"page":      page,
			"page_size": pageSize,
		})
	}
}

func detail(runtime *bootstrap.Runtime) httpx.HandlerFunc {
	return func(c *httpx.Context) {
		id, err := c.QueryInt64("id")
		if err != nil || id <= 0 {
			c.BadRequest("invalid id")
			return
		}

		var user model.AdminUser
		if err := runtime.DB.First(&user, id).Error; err != nil {
			c.Error(err)
			return
		}

		var roleIDs []int64
		if err := runtime.DB.Model(&model.AdminUserRole{}).Where("user_id = ?", user.ID).Pluck("role_id", &roleIDs).Error; err != nil {
			c.Error(err)
			return
		}

		c.Success(map[string]any{
			"id":         user.ID,
			"username":   user.Username,
			"nickname":   user.Nickname,
			"status":     user.Status,
			"is_super":   user.IsSuper,
			"role_ids":   roleIDs,
			"created_at": user.CreatedAt,
			"updated_at": user.UpdatedAt,
		})
	}
}

func save(runtime *bootstrap.Runtime) httpx.HandlerFunc {
	return func(c *httpx.Context) {
		var req saveRequest
		if err := c.BindJSON(&req); err != nil {
			c.Error(err)
			return
		}
		if req.Status == 0 {
			req.Status = 1
		}
		if err := runtime.Validator.Struct(req); err != nil {
			c.BadRequest(err.Error())
			return
		}
		if req.ID == 0 && req.Password == "" {
			c.BadRequest("password is required")
			return
		}

		tx := runtime.DB.Begin()
		if tx.Error != nil {
			c.Error(tx.Error)
			return
		}

		var user model.AdminUser
		if req.ID > 0 {
			if err := tx.First(&user, req.ID).Error; err != nil {
				tx.Rollback()
				c.Error(err)
				return
			}
		}

		if req.ID == 0 {
			passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
			if err != nil {
				tx.Rollback()
				c.Error(err)
				return
			}
			user = model.AdminUser{
				Username:     req.Username,
				PasswordHash: string(passwordHash),
				Nickname:     req.Nickname,
				Status:       req.Status,
				IsSuper:      req.IsSuper,
			}
			if err := tx.Create(&user).Error; err != nil {
				tx.Rollback()
				c.Error(err)
				return
			}
		} else {
			updates := map[string]any{
				"username":   req.Username,
				"nickname":   req.Nickname,
				"status":     req.Status,
				"is_super":   req.IsSuper,
				"updated_at": time.Now(),
			}
			if req.Password != "" {
				passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
				if err != nil {
					tx.Rollback()
					c.Error(err)
					return
				}
				updates["password_hash"] = string(passwordHash)
			}
			if err := tx.Model(&user).Updates(updates).Error; err != nil {
				tx.Rollback()
				c.Error(err)
				return
			}
		}

		if err := tx.Where("user_id = ?", user.ID).Delete(&model.AdminUserRole{}).Error; err != nil {
			tx.Rollback()
			c.Error(err)
			return
		}
		for _, roleID := range bootstrap.NormalizeIDs(req.RoleIDs...) {
			if err := tx.Create(&model.AdminUserRole{UserID: user.ID, RoleID: roleID}).Error; err != nil {
				tx.Rollback()
				c.Error(err)
				return
			}
		}

		if err := tx.Commit().Error; err != nil {
			c.Error(err)
			return
		}
		c.Success(map[string]any{"id": user.ID})
	}
}

func deleteUsers(runtime *bootstrap.Runtime) httpx.HandlerFunc {
	return func(c *httpx.Context) {
		var req bootstrap.IDsPayload
		if err := c.BindJSON(&req); err != nil {
			c.Error(err)
			return
		}
		ids := req.Values()
		if len(ids) == 0 {
			c.BadRequest("ids is required")
			return
		}
		for _, id := range ids {
			if id == 1 {
				c.BadRequest("default admin cannot be deleted")
				return
			}
		}

		err := runtime.DB.Transaction(func(tx *gorm.DB) error {
			if err := tx.Where("user_id IN ?", ids).Delete(&model.AdminUserRole{}).Error; err != nil {
				return err
			}
			if err := tx.Where("admin_user_id IN ?", ids).Delete(&model.AdminSession{}).Error; err != nil {
				return err
			}
			return tx.Where("id IN ?", ids).Delete(&model.AdminUser{}).Error
		})
		if err != nil {
			c.Error(err)
			return
		}
		c.Success(map[string]any{"deleted": len(ids)})
	}
}
