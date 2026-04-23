package admin_menu

import (
	"github.com/lvjiaben/goweb-core/httpx"
	"github.com/lvjiaben/goweb-scaffold/internal/bootstrap"
)

func list(runtime *bootstrap.Runtime) httpx.HandlerFunc {
	service := NewService(runtime)
	return func(c *httpx.Context) {
		page, pageSize := bootstrap.Pagination(c)
		result, err := service.List(ListParams{
			Page:     page,
			PageSize: pageSize,
			Keyword:  bootstrap.SearchKeyword(c),
			Filters:  bootstrap.ParseFilter(c),
		})
		if err != nil {
			c.Error(err)
			return
		}
		c.Success(result)
	}
}

func detail(runtime *bootstrap.Runtime) httpx.HandlerFunc {
	service := NewService(runtime)
	return func(c *httpx.Context) {
		id, err := c.QueryInt64("id")
		if err != nil || id <= 0 {
			c.BadRequest("invalid id")
			return
		}
		result, err := service.Detail(id)
		if err != nil {
			c.Error(err)
			return
		}
		c.Success(result)
	}
}

func tree(runtime *bootstrap.Runtime) httpx.HandlerFunc {
	service := NewService(runtime)
	return func(c *httpx.Context) {
		page, pageSize := bootstrap.Pagination(c)
		result, err := service.Tree(page, pageSize)
		if err != nil {
			c.Error(err)
			return
		}
		c.Success(result)
	}
}

func options(runtime *bootstrap.Runtime) httpx.HandlerFunc {
	service := NewService(runtime)
	return func(c *httpx.Context) {
		result, err := service.Options()
		if err != nil {
			c.Error(err)
			return
		}
		c.Success(result)
	}
}

func save(runtime *bootstrap.Runtime) httpx.HandlerFunc {
	service := NewService(runtime)
	return func(c *httpx.Context) {
		var req SaveRequest
		if err := c.BindJSON(&req); err != nil {
			c.Error(err)
			return
		}
		result, err := service.SaveMenu(req)
		if err != nil {
			respondServiceError(c, err)
			return
		}
		c.Success(result)
	}
}

func deleteMenus(runtime *bootstrap.Runtime) httpx.HandlerFunc {
	service := NewService(runtime)
	return func(c *httpx.Context) {
		var req bootstrap.IDsPayload
		if err := c.BindJSON(&req); err != nil {
			c.Error(err)
			return
		}
		result, err := service.DeleteMenus(req.Values())
		if err != nil {
			respondServiceError(c, err)
			return
		}
		c.Success(result)
	}
}

func respondServiceError(c *httpx.Context, err error) {
	if isValidationError(err) {
		c.BadRequest(err.Error())
		return
	}
	c.Error(err)
}
