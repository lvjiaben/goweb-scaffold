package admin_user

import (
	"fmt"
	"strings"
	"time"

	"github.com/lvjiaben/goweb-core/httpx"
	"github.com/lvjiaben/goweb-scaffold/internal/bootstrap"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type saveRequest struct {
	ID       int64   `json:"id"`
	Username string  `json:"username" validate:"required"`
	Password string  `json:"password"`
	Nickname string  `json:"nickname"`
	RealName string  `json:"realname"`
	Status   int     `json:"status"`
	IsSuper  bool    `json:"is_super"`
	RoleIDs  []int64 `json:"role_ids"`
}

func list(runtime *bootstrap.Runtime) httpx.HandlerFunc {
	return func(c *httpx.Context) {
		page, pageSize := bootstrap.Pagination(c)
		filters := bootstrap.ParseFilter(c)
		keyword := bootstrap.LikeKeyword(bootstrap.SearchKeyword(c))
		username := bootstrap.LikeKeyword(bootstrap.FilterString(filters, "username"))
		nickname := bootstrap.LikeKeyword(bootstrap.FilterString(filters, "nickname", "realname"))
		status, hasStatus := bootstrap.FilterInt64(filters, "status")

		query := runtime.DB.Model(&AdminUser{}).Order("id DESC")
		if keyword != "" {
			query = query.Where("username ILIKE ? OR nickname ILIKE ?", keyword, keyword)
		}
		if username != "" {
			query = query.Where("username ILIKE ?", username)
		}
		if nickname != "" {
			query = query.Where("nickname ILIKE ?", nickname)
		}
		if hasStatus {
			query = query.Where("status = ?", status)
		}

		var total int64
		if err := query.Count(&total).Error; err != nil {
			c.Error(err)
			return
		}

		var users []AdminUser
		if err := query.Offset((page - 1) * pageSize).Limit(pageSize).Find(&users).Error; err != nil {
			c.Error(err)
			return
		}

		userIDs := make([]int64, 0, len(users))
		for _, user := range users {
			userIDs = append(userIDs, user.ID)
		}

		roleMap := make(map[int64][]int64, len(userIDs))
		roleNameMap := make(map[int64][]string, len(userIDs))
		if len(userIDs) > 0 {
			var mappings []AdminUserRole
			if err := runtime.DB.Where("user_id IN ?", userIDs).Find(&mappings).Error; err != nil {
				c.Error(err)
				return
			}

			roleIDs := make([]int64, 0, len(mappings))
			for _, item := range mappings {
				roleMap[item.UserID] = append(roleMap[item.UserID], item.RoleID)
				roleIDs = append(roleIDs, item.RoleID)
			}

			if len(roleIDs) > 0 {
				var roles []AdminRole
				if err := runtime.DB.Where("id IN ?", bootstrap.NormalizeIDs(roleIDs...)).Find(&roles).Error; err != nil {
					c.Error(err)
					return
				}
				roleIndex := make(map[int64]string, len(roles))
				for _, role := range roles {
					roleIndex[role.ID] = role.Name
				}
				for userID, ids := range roleMap {
					for _, roleID := range ids {
						if name := roleIndex[roleID]; name != "" {
							roleNameMap[userID] = append(roleNameMap[userID], name)
						}
					}
				}
			}
		}

		items := make([]map[string]any, 0, len(users))
		for _, user := range users {
			items = append(items, map[string]any{
				"id":         user.ID,
				"username":   user.Username,
				"nickname":   user.Nickname,
				"realname":   user.Nickname,
				"status":     user.Status,
				"is_super":   user.IsSuper,
				"role_ids":   roleMap[user.ID],
				"role_names": roleNameMap[user.ID],
				"created_at": user.CreatedAt,
				"updated_at": user.UpdatedAt,
			})
		}

		c.Success(bootstrap.PagedResult(items, total, page, pageSize))
	}
}

func detail(runtime *bootstrap.Runtime) httpx.HandlerFunc {
	return func(c *httpx.Context) {
		id, err := c.QueryInt64("id")
		if err != nil || id <= 0 {
			c.BadRequest("invalid id")
			return
		}

		var user AdminUser
		if err := runtime.DB.First(&user, id).Error; err != nil {
			c.Error(err)
			return
		}

		var roleIDs []int64
		if err := runtime.DB.Model(&AdminUserRole{}).Where("user_id = ?", user.ID).Pluck("role_id", &roleIDs).Error; err != nil {
			c.Error(err)
			return
		}

		c.Success(map[string]any{
			"id":         user.ID,
			"username":   user.Username,
			"nickname":   user.Nickname,
			"realname":   user.Nickname,
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
		req.Username = strings.TrimSpace(req.Username)
		if strings.TrimSpace(req.Nickname) == "" {
			req.Nickname = strings.TrimSpace(req.RealName)
		}
		req.Nickname = strings.TrimSpace(req.Nickname)
		if err := runtime.Validator.Struct(req); err != nil {
			c.BadRequest(err.Error())
			return
		}
		if req.Nickname == "" {
			c.BadRequest("nickname is required")
			return
		}
		if req.ID == 0 && req.Password == "" {
			c.BadRequest("password is required")
			return
		}
		if err := ensureUsernameUnique(runtime, req.ID, req.Username); err != nil {
			c.BadRequest(err.Error())
			return
		}

		tx := runtime.DB.Begin()
		if tx.Error != nil {
			c.Error(tx.Error)
			return
		}

		var user AdminUser
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
			user = AdminUser{
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

		if err := tx.Where("user_id = ?", user.ID).Delete(&AdminUserRole{}).Error; err != nil {
			tx.Rollback()
			c.Error(err)
			return
		}
		for _, roleID := range bootstrap.NormalizeIDs(req.RoleIDs...) {
			if err := tx.Create(&AdminUserRole{UserID: user.ID, RoleID: roleID}).Error; err != nil {
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

func ensureUsernameUnique(runtime *bootstrap.Runtime, currentID int64, username string) error {
	var count int64
	query := runtime.DB.Model(&AdminUser{}).Where("username = ?", username)
	if currentID > 0 {
		query = query.Where("id <> ?", currentID)
	}
	if err := query.Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return fmt.Errorf("username already exists")
	}
	return nil
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
			if err := tx.Where("user_id IN ?", ids).Delete(&AdminUserRole{}).Error; err != nil {
				return err
			}
			if err := tx.Where("admin_user_id IN ?", ids).Delete(&AdminSession{}).Error; err != nil {
				return err
			}
			return tx.Where("id IN ?", ids).Delete(&AdminUser{}).Error
		})
		if err != nil {
			c.Error(err)
			return
		}
		c.Success(map[string]any{"deleted": len(ids)})
	}
}
