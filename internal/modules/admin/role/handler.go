package admin_role

import (
	"fmt"
	"strings"
	"time"

	"github.com/lvjiaben/goweb-core/httpx"
	"github.com/lvjiaben/goweb-scaffold/internal/bootstrap"
	"gorm.io/gorm"
)

type saveRequest struct {
	ID          int64   `json:"id"`
	Name        string  `json:"name" validate:"required"`
	Code        string  `json:"code"`
	Description string  `json:"description"`
	Status      int     `json:"status"`
	MenuIDs     []int64 `json:"menu_ids"`
}

func options(runtime *bootstrap.Runtime) httpx.HandlerFunc {
	return func(c *httpx.Context) {
		var roles []AdminRole
		if err := runtime.DB.Where("status = ?", 1).Order("id ASC").Find(&roles).Error; err != nil {
			c.Error(err)
			return
		}

		items := make([]map[string]any, 0, len(roles))
		for _, role := range roles {
			items = append(items, map[string]any{
				"label": role.Name,
				"value": role.ID,
				"code":  role.Code,
			})
		}
		c.Success(map[string]any{"list": items})
	}
}

func list(runtime *bootstrap.Runtime) httpx.HandlerFunc {
	return func(c *httpx.Context) {
		page, pageSize := bootstrap.Pagination(c)
		filters := bootstrap.ParseFilter(c)
		keyword := bootstrap.LikeKeyword(bootstrap.SearchKeyword(c))
		name := bootstrap.LikeKeyword(bootstrap.FilterString(filters, "name"))
		code := bootstrap.LikeKeyword(bootstrap.FilterString(filters, "code", "description"))
		status, hasStatus := bootstrap.FilterInt64(filters, "status")

		query := runtime.DB.Model(&AdminRole{}).Order("id DESC")
		if keyword != "" {
			query = query.Where("name ILIKE ? OR code ILIKE ?", keyword, keyword)
		}
		if name != "" {
			query = query.Where("name ILIKE ?", name)
		}
		if code != "" {
			query = query.Where("code ILIKE ?", code)
		}
		if hasStatus {
			query = query.Where("status = ?", status)
		}

		var total int64
		if err := query.Count(&total).Error; err != nil {
			c.Error(err)
			return
		}

		var roles []AdminRole
		if err := query.Offset((page - 1) * pageSize).Limit(pageSize).Find(&roles).Error; err != nil {
			c.Error(err)
			return
		}

		items := make([]map[string]any, 0, len(roles))
		for _, role := range roles {
			items = append(items, map[string]any{
				"id":          role.ID,
				"name":        role.Name,
				"code":        role.Code,
				"description": role.Code,
				"status":      role.Status,
				"created_at":  role.CreatedAt,
				"updated_at":  role.UpdatedAt,
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

		var role AdminRole
		if err := runtime.DB.First(&role, id).Error; err != nil {
			c.Error(err)
			return
		}

		var menuIDs []int64
		if err := runtime.DB.Model(&AdminRoleMenu{}).Where("role_id = ?", role.ID).Pluck("menu_id", &menuIDs).Error; err != nil {
			c.Error(err)
			return
		}

		c.Success(map[string]any{
			"id":          role.ID,
			"name":        role.Name,
			"code":        role.Code,
			"description": role.Code,
			"status":      role.Status,
			"menu_ids":    menuIDs,
			"created_at":  role.CreatedAt,
			"updated_at":  role.UpdatedAt,
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
		req.Name = strings.TrimSpace(req.Name)
		if strings.TrimSpace(req.Code) == "" {
			req.Code = strings.TrimSpace(req.Description)
		}
		req.Code = strings.TrimSpace(req.Code)
		if err := runtime.Validator.Struct(req); err != nil {
			c.BadRequest(err.Error())
			return
		}
		if req.Code == "" {
			c.BadRequest("code is required")
			return
		}
		if err := ensureRoleCodeUnique(runtime, req.ID, req.Code); err != nil {
			c.BadRequest(err.Error())
			return
		}

		tx := runtime.DB.Begin()
		if tx.Error != nil {
			c.Error(tx.Error)
			return
		}

		var role AdminRole
		if req.ID == 0 {
			role = AdminRole{
				Name:   req.Name,
				Code:   req.Code,
				Status: req.Status,
			}
			if err := tx.Create(&role).Error; err != nil {
				tx.Rollback()
				c.Error(err)
				return
			}
		} else {
			if err := tx.First(&role, req.ID).Error; err != nil {
				tx.Rollback()
				c.Error(err)
				return
			}
			if err := tx.Model(&role).Updates(map[string]any{
				"name":       req.Name,
				"code":       req.Code,
				"status":     req.Status,
				"updated_at": time.Now(),
			}).Error; err != nil {
				tx.Rollback()
				c.Error(err)
				return
			}
		}

		if err := tx.Where("role_id = ?", role.ID).Delete(&AdminRoleMenu{}).Error; err != nil {
			tx.Rollback()
			c.Error(err)
			return
		}
		for _, menuID := range bootstrap.NormalizeIDs(req.MenuIDs...) {
			if err := tx.Create(&AdminRoleMenu{RoleID: role.ID, MenuID: menuID}).Error; err != nil {
				tx.Rollback()
				c.Error(err)
				return
			}
		}

		if err := tx.Commit().Error; err != nil {
			c.Error(err)
			return
		}
		c.Success(map[string]any{"id": role.ID})
	}
}

func deleteRoles(runtime *bootstrap.Runtime) httpx.HandlerFunc {
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
				c.BadRequest("super admin role cannot be deleted")
				return
			}
		}

		var bindCount int64
		if err := runtime.DB.Model(&AdminUserRole{}).Where("role_id IN ?", ids).Count(&bindCount).Error; err != nil {
			c.Error(err)
			return
		}
		if bindCount > 0 {
			c.BadRequest("role still bound to admin users")
			return
		}

		err := runtime.DB.Transaction(func(tx *gorm.DB) error {
			if err := tx.Where("role_id IN ?", ids).Delete(&AdminRoleMenu{}).Error; err != nil {
				return err
			}
			if err := tx.Where("role_id IN ?", ids).Delete(&AdminUserRole{}).Error; err != nil {
				return err
			}
			return tx.Where("id IN ?", ids).Delete(&AdminRole{}).Error
		})
		if err != nil {
			c.Error(err)
			return
		}
		c.Success(map[string]any{"deleted": len(ids)})
	}
}

func ensureRoleCodeUnique(runtime *bootstrap.Runtime, currentID int64, code string) error {
	var count int64
	query := runtime.DB.Model(&AdminRole{}).Where("code = ?", code)
	if currentID > 0 {
		query = query.Where("id <> ?", currentID)
	}
	if err := query.Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return fmt.Errorf("role code already exists")
	}
	return nil
}
