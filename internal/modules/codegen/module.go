package codegen

import (
	"encoding/json"

	"github.com/lvjiaben/goweb-core/httpx"
	"github.com/lvjiaben/goweb-scaffold/internal/bootstrap"
	"github.com/lvjiaben/goweb-scaffold/internal/modules/model"
)

type Module struct{}

type saveRequest struct {
	ModuleName string          `json:"module_name" validate:"required"`
	TableName  string          `json:"table_name" validate:"required"`
	Payload    json.RawMessage `json:"payload"`
}

func (Module) Name() string { return "codegen" }

func (Module) Register(runtime *bootstrap.Runtime) error {
	runtime.AdminProtectedGroup.GET("/codegen/list", list(runtime), httpx.WithPermission("codegen.list"))
	runtime.AdminProtectedGroup.POST("/codegen/save", save(runtime), httpx.WithPermission("codegen.save"))
	runtime.AdminProtectedGroup.POST("/codegen/delete", deleteHistory(runtime), httpx.WithPermission("codegen.delete"))
	return nil
}

func list(runtime *bootstrap.Runtime) httpx.HandlerFunc {
	return func(c *httpx.Context) {
		var rows []model.CodegenHistory
		if err := runtime.DB.Order("id DESC").Find(&rows).Error; err != nil {
			c.Error(err)
			return
		}

		items := make([]map[string]any, 0, len(rows))
		for _, row := range rows {
			items = append(items, map[string]any{
				"id":          row.ID,
				"module_name": row.ModuleName,
				"table_name":  row.SourceTable,
				"status":      row.Status,
				"payload":     json.RawMessage(row.Payload),
				"remark":      row.Remark,
				"created_at":  row.CreatedAt,
			})
		}
		c.Success(map[string]any{"list": items})
	}
}

func save(runtime *bootstrap.Runtime) httpx.HandlerFunc {
	return func(c *httpx.Context) {
		var req saveRequest
		if err := c.BindJSON(&req); err != nil {
			c.Error(err)
			return
		}
		if err := runtime.Validator.Struct(req); err != nil {
			c.BadRequest(err.Error())
			return
		}
		if len(req.Payload) == 0 {
			req.Payload = json.RawMessage(`{}`)
		}

		record := model.CodegenHistory{
			ModuleName:  req.ModuleName,
			SourceTable: req.TableName,
			Status:      "placeholder",
			Payload:     model.JSON(req.Payload),
			Remark:      "code generator skeleton is reserved for admin only, generation logic is not implemented in v1",
		}
		if err := runtime.DB.Create(&record).Error; err != nil {
			c.Error(err)
			return
		}

		c.Success(map[string]any{
			"id":          record.ID,
			"status":      record.Status,
			"placeholder": true,
		})
	}
}

func deleteHistory(runtime *bootstrap.Runtime) httpx.HandlerFunc {
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
		if err := runtime.DB.Where("id IN ?", ids).Delete(&model.CodegenHistory{}).Error; err != nil {
			c.Error(err)
			return
		}
		c.Success(map[string]any{"deleted": len(ids)})
	}
}
