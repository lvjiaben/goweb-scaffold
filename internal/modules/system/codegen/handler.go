package codegen

import (
	"strconv"
	"strings"

	"github.com/lvjiaben/goweb-core/httpx"
	"github.com/lvjiaben/goweb-scaffold/internal/bootstrap"
)

func list(runtime *bootstrap.Runtime) httpx.HandlerFunc {
	service := NewService(runtime)
	return func(c *httpx.Context) {
		result, err := service.HistoryList()
		if err != nil {
			c.Error(err)
			return
		}
		c.Success(result)
	}
}

func tables(runtime *bootstrap.Runtime) httpx.HandlerFunc {
	service := NewService(runtime)
	return func(c *httpx.Context) {
		result, err := service.Tables()
		if err != nil {
			c.Error(err)
			return
		}
		c.Success(result)
	}
}

func modules(runtime *bootstrap.Runtime) httpx.HandlerFunc {
	service := NewService(runtime)
	return func(c *httpx.Context) {
		result, err := service.Modules()
		if err != nil {
			c.Error(err)
			return
		}
		c.Success(result)
	}
}

func tableColumns(runtime *bootstrap.Runtime) httpx.HandlerFunc {
	service := NewService(runtime)
	return func(c *httpx.Context) {
		result, err := service.TableColumns(c.Query("table_name"))
		if err != nil {
			respondCodegenError(c, err)
			return
		}
		c.Success(result)
	}
}

func exportFile(runtime *bootstrap.Runtime) httpx.HandlerFunc {
	service := NewService(runtime)
	return func(c *httpx.Context) {
		historyID, err := parseOptionalInt64(c.Query("history_id"), "history_id")
		if err != nil {
			c.BadRequest(err.Error())
			return
		}
		result, err := service.Export(ExportRequest{
			ModuleName: strings.TrimSpace(c.Query("module_name")),
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
	service := NewService(runtime)
	return func(c *httpx.Context) {
		var req SaveRequest
		if err := bindAndValidate(c, runtime, &req); err != nil {
			return
		}
		result, err := service.Preview(req)
		if err != nil {
			respondCodegenError(c, err)
			return
		}
		c.Success(result)
	}
}

func diff(runtime *bootstrap.Runtime) httpx.HandlerFunc {
	service := NewService(runtime)
	return func(c *httpx.Context) {
		var req GenerateRequest
		if err := bindAndValidate(c, runtime, &req); err != nil {
			return
		}
		result, err := service.Diff(req)
		if err != nil {
			respondCodegenError(c, err)
			return
		}
		c.Success(result)
	}
}

func checkBreaking(runtime *bootstrap.Runtime) httpx.HandlerFunc {
	service := NewService(runtime)
	return func(c *httpx.Context) {
		var req CheckBreakingRequest
		if err := bindAndValidate(c, runtime, &req); err != nil {
			return
		}
		result, err := service.CheckBreaking(req)
		if err != nil {
			respondCodegenError(c, err)
			return
		}
		c.Success(result)
	}
}

func generate(runtime *bootstrap.Runtime) httpx.HandlerFunc {
	service := NewService(runtime)
	return func(c *httpx.Context) {
		var req GenerateRequest
		if err := bindAndValidate(c, runtime, &req); err != nil {
			return
		}
		result, err := service.Generate(req)
		if err != nil {
			respondCodegenError(c, err)
			return
		}
		c.Success(result)
	}
}

func regenerate(runtime *bootstrap.Runtime) httpx.HandlerFunc {
	service := NewService(runtime)
	return func(c *httpx.Context) {
		var req RegenerateRequest
		if err := c.BindJSON(&req); err != nil {
			c.Error(err)
			return
		}
		result, err := service.Regenerate(req)
		if err != nil {
			respondCodegenError(c, err)
			return
		}
		c.Success(result)
	}
}

func remove(runtime *bootstrap.Runtime) httpx.HandlerFunc {
	service := NewService(runtime)
	return func(c *httpx.Context) {
		var req RemoveRequest
		if err := bindAndValidate(c, runtime, &req); err != nil {
			return
		}
		result, err := service.Remove(req)
		if err != nil {
			respondCodegenError(c, err)
			return
		}
		c.Success(result)
	}
}

func save(runtime *bootstrap.Runtime) httpx.HandlerFunc {
	service := NewService(runtime)
	return func(c *httpx.Context) {
		var req SaveRequest
		if err := bindAndValidate(c, runtime, &req); err != nil {
			return
		}
		result, err := service.SaveDraft(req)
		if err != nil {
			respondCodegenError(c, err)
			return
		}
		c.Success(result)
	}
}

func deleteHistory(runtime *bootstrap.Runtime) httpx.HandlerFunc {
	service := NewService(runtime)
	return func(c *httpx.Context) {
		var req bootstrap.IDsPayload
		if err := c.BindJSON(&req); err != nil {
			c.Error(err)
			return
		}
		result, err := service.DeleteHistory(req.Values())
		if err != nil {
			respondCodegenError(c, err)
			return
		}
		c.Success(result)
	}
}

func bindAndValidate(c *httpx.Context, runtime *bootstrap.Runtime, req any) error {
	if err := c.BindJSON(req); err != nil {
		c.Error(err)
		return err
	}
	if runtime != nil && runtime.Validator != nil {
		if err := runtime.Validator.Struct(req); err != nil {
			c.BadRequest(err.Error())
			return err
		}
	}
	return nil
}

func parseOptionalInt64(raw string, name string) (int64, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 0, nil
	}
	value, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return 0, validationError(name + " must be a valid integer")
	}
	return value, nil
}
