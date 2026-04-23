package codegen

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/lvjiaben/goweb-core/httpx"
	"github.com/lvjiaben/goweb-scaffold/internal/bootstrap"
	"github.com/lvjiaben/goweb-scaffold/internal/gen/service"
)

type saveRequest struct {
	ModuleName string          `json:"module_name" validate:"required"`
	TableName  string          `json:"table_name" validate:"required"`
	Payload    json.RawMessage `json:"payload"`
}

type generateRequest struct {
	ModuleName     string          `json:"module_name" validate:"required"`
	TableName      string          `json:"table_name" validate:"required"`
	Payload        json.RawMessage `json:"payload"`
	Overwrite      bool            `json:"overwrite"`
	RegisterModule bool            `json:"register_module"`
	UpsertMenu     bool            `json:"upsert_menu"`
}

type checkBreakingRequest struct {
	ModuleName     string          `json:"module_name" validate:"required"`
	TableName      string          `json:"table_name"`
	Payload        json.RawMessage `json:"payload"`
	RegisterModule *bool           `json:"register_module"`
}

type regenerateRequest struct {
	ModuleName     string `json:"module_name"`
	HistoryID      int64  `json:"history_id"`
	Overwrite      bool   `json:"overwrite"`
	RegisterModule bool   `json:"register_module"`
	UpsertMenu     bool   `json:"upsert_menu"`
}

type removeRequest struct {
	ModuleName       string `json:"module_name" validate:"required"`
	RemoveFiles      bool   `json:"remove_files"`
	UnregisterModule bool   `json:"unregister_module"`
	RemoveMenu       bool   `json:"remove_menu"`
	RemoveHistory    bool   `json:"remove_history"`
	RemoveLock       bool   `json:"remove_lock"`
}

type regenerateSource struct {
	ModuleName string
	TableName  string
	Payload    json.RawMessage
}

func list(runtime *bootstrap.Runtime) httpx.HandlerFunc {
	return func(c *httpx.Context) {
		var rows []CodegenHistory
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

func tables(runtime *bootstrap.Runtime) httpx.HandlerFunc {
	return func(c *httpx.Context) {
		items, err := listBusinessTables(runtime)
		if err != nil {
			c.Error(err)
			return
		}
		c.Success(map[string]any{"list": items})
	}
}

func modules(runtime *bootstrap.Runtime) httpx.HandlerFunc {
	return func(c *httpx.Context) {
		items, err := (service.GeneratorService{
			RepoRoot: runtime.RepoRoot,
			DB:       runtime.DB,
		}).ListModules()
		if err != nil {
			c.Error(err)
			return
		}
		c.Success(map[string]any{"list": items})
	}
}

func tableColumns(runtime *bootstrap.Runtime) httpx.HandlerFunc {
	return func(c *httpx.Context) {
		tableName := strings.TrimSpace(c.Query("table_name"))
		if tableName == "" {
			c.BadRequest("table_name is required")
			return
		}

		items, err := listTableColumns(runtime, tableName)
		if err != nil {
			c.Error(err)
			return
		}
		c.Success(map[string]any{"list": items})
	}
}

func exportFile(runtime *bootstrap.Runtime) httpx.HandlerFunc {
	return func(c *httpx.Context) {
		moduleName := strings.TrimSpace(c.Query("module_name"))
		historyRaw := strings.TrimSpace(c.Query("history_id"))
		var historyID int64
		if historyRaw != "" {
			value, err := strconv.ParseInt(historyRaw, 10, 64)
			if err != nil {
				c.BadRequest("history_id must be a valid integer")
				return
			}
			historyID = value
		}
		if moduleName == "" && historyID <= 0 {
			c.BadRequest("module_name or history_id is required")
			return
		}

		result, err := NewRunner(runtime).Export(ExportInput{
			ModuleName: moduleName,
			HistoryID:  historyID,
		})
		if err != nil {
			respondCodegenError(c, err)
			return
		}
		c.Success(result)
	}
}

func preview(runtime *bootstrap.Runtime) httpx.HandlerFunc {
	return func(c *httpx.Context) {
		var req saveRequest
		if err := c.BindJSON(&req); err != nil {
			c.Error(err)
			return
		}
		req.ModuleName = strings.TrimSpace(req.ModuleName)
		req.TableName = strings.TrimSpace(req.TableName)
		if err := runtime.Validator.Struct(req); err != nil {
			c.BadRequest(err.Error())
			return
		}
		previewPayload, _, err := buildPreview(runtime, req.ModuleName, req.TableName, req.Payload)
		if err != nil {
			respondCodegenError(c, err)
			return
		}
		c.Success(previewPayload)
	}
}

func diff(runtime *bootstrap.Runtime) httpx.HandlerFunc {
	return func(c *httpx.Context) {
		var req generateRequest
		if err := c.BindJSON(&req); err != nil {
			c.Error(err)
			return
		}
		req.ModuleName = strings.TrimSpace(req.ModuleName)
		req.TableName = strings.TrimSpace(req.TableName)
		if err := runtime.Validator.Struct(req); err != nil {
			c.BadRequest(err.Error())
			return
		}

		previewPayload, columns, err := buildPreview(runtime, req.ModuleName, req.TableName, req.Payload)
		if err != nil {
			respondCodegenError(c, err)
			return
		}

		result, err := (service.GeneratorService{
			RepoRoot: runtime.RepoRoot,
			DB:       runtime.DB,
		}).Diff(service.GenerateInput{
			ModuleName:     req.ModuleName,
			TableName:      req.TableName,
			Payload:        previewPayload.Payload,
			Preview:        previewPayload,
			Columns:        columns,
			Overwrite:      req.Overwrite,
			RegisterModule: req.RegisterModule,
			UpsertMenu:     req.UpsertMenu,
		})
		if err != nil {
			respondCodegenError(c, err)
			return
		}

		c.Success(result)
	}
}

func checkBreaking(runtime *bootstrap.Runtime) httpx.HandlerFunc {
	return func(c *httpx.Context) {
		var req checkBreakingRequest
		if err := c.BindJSON(&req); err != nil {
			c.Error(err)
			return
		}
		req.ModuleName = strings.TrimSpace(req.ModuleName)
		req.TableName = strings.TrimSpace(req.TableName)
		if err := runtime.Validator.Struct(req); err != nil {
			c.BadRequest(err.Error())
			return
		}
		result, err := NewRunner(runtime).CheckBreaking(CheckBreakingInput{
			ModuleName:     req.ModuleName,
			TableName:      req.TableName,
			Payload:        req.Payload,
			RegisterModule: req.RegisterModule == nil || *req.RegisterModule,
		})
		if err != nil {
			respondCodegenError(c, err)
			return
		}
		c.Success(result)
	}
}

func generate(runtime *bootstrap.Runtime) httpx.HandlerFunc {
	return func(c *httpx.Context) {
		var req generateRequest
		if err := c.BindJSON(&req); err != nil {
			c.Error(err)
			return
		}
		req.ModuleName = strings.TrimSpace(req.ModuleName)
		req.TableName = strings.TrimSpace(req.TableName)
		if err := runtime.Validator.Struct(req); err != nil {
			c.BadRequest(err.Error())
			return
		}
		previewPayload, columns, err := buildPreview(runtime, req.ModuleName, req.TableName, req.Payload)
		if err != nil {
			respondCodegenError(c, err)
			return
		}
		result, err := (service.GeneratorService{
			RepoRoot: runtime.RepoRoot,
			DB:       runtime.DB,
		}).Generate(service.GenerateInput{
			ModuleName:     req.ModuleName,
			TableName:      req.TableName,
			Payload:        previewPayload.Payload,
			Preview:        previewPayload,
			Columns:        columns,
			Overwrite:      req.Overwrite,
			RegisterModule: req.RegisterModule,
			UpsertMenu:     req.UpsertMenu,
		})
		if err != nil {
			respondCodegenError(c, err)
			return
		}

		rawPayload, err := json.Marshal(previewPayload.Payload)
		if err != nil {
			c.Error(err)
			return
		}
		record := CodegenHistory{
			ModuleName:  req.ModuleName,
			SourceTable: req.TableName,
			Status:      "generated",
			Payload:     JSON(rawPayload),
			Remark:      "generated admin CRUD files",
		}
		if err := runtime.DB.Create(&record).Error; err != nil {
			c.Error(err)
			return
		}
		c.Success(result)
	}
}

func regenerate(runtime *bootstrap.Runtime) httpx.HandlerFunc {
	return func(c *httpx.Context) {
		var req regenerateRequest
		if err := c.BindJSON(&req); err != nil {
			c.Error(err)
			return
		}
		req.ModuleName = strings.TrimSpace(req.ModuleName)
		if req.ModuleName == "" && req.HistoryID <= 0 {
			c.BadRequest("module_name or history_id is required")
			return
		}

		source, err := loadRegenerateSource(runtime, req)
		if err != nil {
			respondCodegenError(c, err)
			return
		}

		previewPayload, columns, err := buildPreview(runtime, source.ModuleName, source.TableName, source.Payload)
		if err != nil {
			respondCodegenError(c, err)
			return
		}

		result, err := (service.GeneratorService{
			RepoRoot: runtime.RepoRoot,
			DB:       runtime.DB,
		}).Generate(service.GenerateInput{
			ModuleName:     source.ModuleName,
			TableName:      source.TableName,
			Payload:        previewPayload.Payload,
			Preview:        previewPayload,
			Columns:        columns,
			Overwrite:      req.Overwrite,
			RegisterModule: req.RegisterModule,
			UpsertMenu:     req.UpsertMenu,
		})
		if err != nil {
			respondCodegenError(c, err)
			return
		}

		rawPayload, err := json.Marshal(previewPayload.Payload)
		if err != nil {
			c.Error(err)
			return
		}
		record := CodegenHistory{
			ModuleName:  source.ModuleName,
			SourceTable: source.TableName,
			Status:      "regenerated",
			Payload:     JSON(rawPayload),
			Remark:      "regenerated admin CRUD files from lock/history",
		}
		if err := runtime.DB.Create(&record).Error; err != nil {
			c.Error(err)
			return
		}

		c.Success(result)
	}
}

func remove(runtime *bootstrap.Runtime) httpx.HandlerFunc {
	return func(c *httpx.Context) {
		var req removeRequest
		if err := c.BindJSON(&req); err != nil {
			c.Error(err)
			return
		}
		req.ModuleName = strings.TrimSpace(req.ModuleName)
		if err := runtime.Validator.Struct(req); err != nil {
			c.BadRequest(err.Error())
			return
		}

		result, err := (service.GeneratorService{
			RepoRoot: runtime.RepoRoot,
			DB:       runtime.DB,
		}).Remove(service.RemoveInput{
			ModuleName:       req.ModuleName,
			RemoveFiles:      req.RemoveFiles,
			UnregisterModule: req.UnregisterModule,
			RemoveMenu:       req.RemoveMenu,
			RemoveHistory:    req.RemoveHistory,
			RemoveLock:       req.RemoveLock,
		})
		if err != nil {
			respondCodegenError(c, err)
			return
		}
		c.Success(result)
	}
}

func save(runtime *bootstrap.Runtime) httpx.HandlerFunc {
	return func(c *httpx.Context) {
		var req saveRequest
		if err := c.BindJSON(&req); err != nil {
			c.Error(err)
			return
		}
		req.ModuleName = strings.TrimSpace(req.ModuleName)
		req.TableName = strings.TrimSpace(req.TableName)
		if err := runtime.Validator.Struct(req); err != nil {
			c.BadRequest(err.Error())
			return
		}

		previewPayload, _, err := buildPreview(runtime, req.ModuleName, req.TableName, req.Payload)
		if err != nil {
			respondCodegenError(c, err)
			return
		}

		rawPayload, err := json.Marshal(previewPayload.Payload)
		if err != nil {
			c.Error(err)
			return
		}

		record := CodegenHistory{
			ModuleName:  req.ModuleName,
			SourceTable: req.TableName,
			Status:      "draft",
			Payload:     JSON(rawPayload),
			Remark:      "preview draft for admin codegen regenerate workflow",
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
		if err := runtime.DB.Where("id IN ?", ids).Delete(&CodegenHistory{}).Error; err != nil {
			c.Error(err)
			return
		}
		c.Success(map[string]any{"deleted": len(ids)})
	}
}

func buildPreview(runtime *bootstrap.Runtime, moduleName string, tableName string, payload json.RawMessage) (service.Preview, []service.ColumnInfo, error) {
	if len(payload) == 0 {
		payload = json.RawMessage(`{}`)
	}

	columns, err := listTableColumns(runtime, tableName)
	if err != nil {
		return service.Preview{}, nil, err
	}
	if len(columns) == 0 {
		return service.Preview{}, nil, fmt.Errorf("table %s has no available columns for generation", tableName)
	}

	previewPayload := service.BuildPreview(moduleName, tableName, payload, columns)
	return previewPayload, columns, nil
}

func respondCodegenError(c *httpx.Context, err error) {
	if err == nil {
		return
	}
	switch {
	case os.IsNotExist(err):
		c.BadRequest("codegen lock file not found")
	case strings.Contains(strings.ToLower(err.Error()), "not found"):
		c.BadRequest(err.Error())
	case strings.Contains(strings.ToLower(err.Error()), "required"):
		c.BadRequest(err.Error())
	case strings.Contains(strings.ToLower(err.Error()), "no available columns"):
		c.BadRequest(err.Error())
	default:
		c.Error(err)
	}
}

func loadRegenerateSource(runtime *bootstrap.Runtime, req regenerateRequest) (regenerateSource, error) {
	codegenService := service.GeneratorService{
		RepoRoot: runtime.RepoRoot,
		DB:       runtime.DB,
	}
	if req.HistoryID > 0 {
		return loadRegenerateSourceFromHistory(runtime, req.HistoryID)
	}

	lock, err := codegenService.LoadLock(req.ModuleName)
	if err == nil {
		rawPayload, marshalErr := json.Marshal(lock.Payload)
		if marshalErr != nil {
			return regenerateSource{}, marshalErr
		}
		return regenerateSource{
			ModuleName: lock.ModuleName,
			TableName:  lock.TableName,
			Payload:    rawPayload,
		}, nil
	}
	if !os.IsNotExist(err) {
		return regenerateSource{}, err
	}

	var row CodegenHistory
	if err := runtime.DB.Where("module_name = ?", req.ModuleName).Order("id DESC").First(&row).Error; err != nil {
		return regenerateSource{}, fmt.Errorf("codegen source for module %s not found", req.ModuleName)
	}
	return regenerateSource{
		ModuleName: row.ModuleName,
		TableName:  row.SourceTable,
		Payload:    json.RawMessage(row.Payload),
	}, nil
}

func loadRegenerateSourceFromHistory(runtime *bootstrap.Runtime, historyID int64) (regenerateSource, error) {
	var row CodegenHistory
	if err := runtime.DB.First(&row, historyID).Error; err != nil {
		return regenerateSource{}, fmt.Errorf("codegen history %d not found", historyID)
	}
	return regenerateSource{
		ModuleName: strings.TrimSpace(row.ModuleName),
		TableName:  strings.TrimSpace(row.SourceTable),
		Payload:    json.RawMessage(row.Payload),
	}, nil
}
