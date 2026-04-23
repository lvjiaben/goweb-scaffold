package system_config

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/lvjiaben/goweb-core/httpx"
	"github.com/lvjiaben/goweb-scaffold/internal/bootstrap"
	"github.com/lvjiaben/goweb-scaffold/internal/modules/model"
)

type Module struct{}

type saveRequest struct {
	ID          int64           `json:"id"`
	ConfigKey   string          `json:"config_key" validate:"required"`
	ConfigName  string          `json:"config_name" validate:"required"`
	ConfigValue json.RawMessage `json:"config_value"`
	Remark      string          `json:"remark"`
}

func (Module) Name() string { return "system_config" }

func (Module) Register(runtime *bootstrap.Runtime) error {
	group := runtime.BackendProtectedGroup.Group("/system/config")
	group.GET("/list", list(runtime), httpx.WithPermission("system_config.list"))
	group.GET("/detail", detail(runtime), httpx.WithPermission("system_config.list"))
	group.POST("/save", save(runtime), httpx.WithPermission("system_config.save"))
	group.POST("/delete", deleteConfigs(runtime), httpx.WithPermission("system_config.delete"))
	return nil
}

func list(runtime *bootstrap.Runtime) httpx.HandlerFunc {
	return func(c *httpx.Context) {
		page, pageSize := bootstrap.Pagination(c)
		filters := bootstrap.ParseFilter(c)
		keyword := bootstrap.LikeKeyword(bootstrap.SearchKeyword(c))
		keyFilter := bootstrap.LikeKeyword(bootstrap.FilterString(filters, "config_key", "key"))
		nameFilter := bootstrap.LikeKeyword(bootstrap.FilterString(filters, "config_name", "name"))

		query := runtime.DB.Model(&model.SystemConfig{}).Order("id DESC")
		if keyword != "" {
			query = query.Where("config_key ILIKE ? OR config_name ILIKE ?", keyword, keyword)
		}
		if keyFilter != "" {
			query = query.Where("config_key ILIKE ?", keyFilter)
		}
		if nameFilter != "" {
			query = query.Where("config_name ILIKE ?", nameFilter)
		}

		var total int64
		if err := query.Count(&total).Error; err != nil {
			c.Error(err)
			return
		}

		var rows []model.SystemConfig
		if err := query.Offset((page - 1) * pageSize).Limit(pageSize).Find(&rows).Error; err != nil {
			c.Error(err)
			return
		}

		items := make([]map[string]any, 0, len(rows))
		for _, item := range rows {
			items = append(items, map[string]any{
				"id":           item.ID,
				"dir":          "system",
				"config_key":   item.ConfigKey,
				"config_name":  item.ConfigName,
				"config_value": json.RawMessage(item.ConfigValue),
				"remark":       item.Remark,
				"type":         "input",
				"created_at":   item.CreatedAt,
				"updated_at":   item.UpdatedAt,
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

		var item model.SystemConfig
		if err := runtime.DB.First(&item, id).Error; err != nil {
			c.Error(err)
			return
		}

		c.Success(map[string]any{
			"id":           item.ID,
			"dir":          "system",
			"config_key":   item.ConfigKey,
			"config_name":  item.ConfigName,
			"config_value": json.RawMessage(item.ConfigValue),
			"remark":       item.Remark,
			"type":         "input",
			"created_at":   item.CreatedAt,
			"updated_at":   item.UpdatedAt,
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
		req.ConfigKey = strings.TrimSpace(req.ConfigKey)
		req.ConfigName = strings.TrimSpace(req.ConfigName)
		req.Remark = strings.TrimSpace(req.Remark)
		if err := runtime.Validator.Struct(req); err != nil {
			c.BadRequest(err.Error())
			return
		}
		if len(req.ConfigValue) == 0 {
			req.ConfigValue = json.RawMessage(`{}`)
		}
		if err := ensureConfigKeyUnique(runtime, req.ID, req.ConfigKey); err != nil {
			c.BadRequest(err.Error())
			return
		}

		if req.ID == 0 {
			item := model.SystemConfig{
				ConfigKey:   req.ConfigKey,
				ConfigName:  req.ConfigName,
				ConfigValue: model.JSON(req.ConfigValue),
				Remark:      req.Remark,
			}
			if err := runtime.DB.Create(&item).Error; err != nil {
				c.Error(err)
				return
			}
			c.Success(map[string]any{"id": item.ID})
			return
		}

		var item model.SystemConfig
		if err := runtime.DB.First(&item, req.ID).Error; err != nil {
			c.Error(err)
			return
		}

		if err := runtime.DB.Model(&item).Updates(map[string]any{
			"config_key":   req.ConfigKey,
			"config_name":  req.ConfigName,
			"config_value": model.JSON(req.ConfigValue),
			"remark":       req.Remark,
			"updated_at":   time.Now(),
		}).Error; err != nil {
			c.Error(err)
			return
		}
		c.Success(map[string]any{"id": item.ID})
	}
}

func deleteConfigs(runtime *bootstrap.Runtime) httpx.HandlerFunc {
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
		if err := runtime.DB.Where("id IN ?", ids).Delete(&model.SystemConfig{}).Error; err != nil {
			c.Error(err)
			return
		}
		c.Success(map[string]any{"deleted": len(ids)})
	}
}

func ensureConfigKeyUnique(runtime *bootstrap.Runtime, currentID int64, key string) error {
	var count int64
	query := runtime.DB.Model(&model.SystemConfig{}).Where("config_key = ?", key)
	if currentID > 0 {
		query = query.Where("id <> ?", currentID)
	}
	if err := query.Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return fmt.Errorf("config_key already exists")
	}
	return nil
}
