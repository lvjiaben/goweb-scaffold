package app_user

import (
	"github.com/lvjiaben/goweb-core/httpx"
	"github.com/lvjiaben/goweb-scaffold/internal/bootstrap"
)

func registerBackend(runtime *bootstrap.Runtime) error {
	group := runtime.BackendProtectedGroup.Group("/app/user")
	group.GET("/list", backendList(runtime), httpx.WithPermission("app_user.list"))
	group.GET("/detail", backendDetail(runtime), httpx.WithPermission("app_user.list"))
	group.POST("/save", backendSave(runtime), httpx.WithPermission("app_user.save"))
	group.POST("/create", backendSave(runtime), httpx.WithPermission("app_user.save"))
	group.POST("/update", backendSave(runtime), httpx.WithPermission("app_user.save"))
	group.POST("/delete", backendDelete(runtime), httpx.WithPermission("app_user.delete"))
	group.POST("/operate", backendOperate(runtime), httpx.WithPermission("app_user.save"))
	group.POST("/money", backendMoney(runtime), httpx.WithPermission("app_user.money"))
	group.POST("/score", backendScore(runtime), httpx.WithPermission("app_user.score"))
	group.GET("/money-logs", backendMoneyLogs(runtime), httpx.WithPermission("app_user.list"))
	group.GET("/score-logs", backendScoreLogs(runtime), httpx.WithPermission("app_user.list"))

	legacy := runtime.BackendProtectedGroup.Group("/user")
	legacy.GET("/list", backendList(runtime), httpx.WithPermission("app_user.list"))
	legacy.GET("/detail", backendDetail(runtime), httpx.WithPermission("app_user.list"))
	legacy.POST("/save", backendSave(runtime), httpx.WithPermission("app_user.save"))
	legacy.POST("/create", backendSave(runtime), httpx.WithPermission("app_user.save"))
	legacy.POST("/update", backendSave(runtime), httpx.WithPermission("app_user.save"))
	legacy.POST("/delete", backendDelete(runtime), httpx.WithPermission("app_user.delete"))
	legacy.POST("/operate", backendOperate(runtime), httpx.WithPermission("app_user.save"))
	legacy.POST("/money", backendMoney(runtime), httpx.WithPermission("app_user.money"))
	legacy.POST("/score", backendScore(runtime), httpx.WithPermission("app_user.score"))
	return nil
}

func backendList(runtime *bootstrap.Runtime) httpx.HandlerFunc {
	service := NewService(runtime)
	return func(c *httpx.Context) {
		page, pageSize := bootstrap.Pagination(c)
		result, err := service.BackendList(ListParams{
			Page:      page,
			PageSize:  pageSize,
			Keyword:   bootstrap.SearchKeyword(c),
			Filters:   bootstrap.ParseFilter(c),
			SortBy:    c.Query("sort_by"),
			SortOrder: c.Query("sort_order"),
		})
		if err != nil {
			respondServiceError(c, err)
			return
		}
		c.Success(result)
	}
}

func backendDetail(runtime *bootstrap.Runtime) httpx.HandlerFunc {
	service := NewService(runtime)
	return func(c *httpx.Context) {
		id, err := c.QueryInt64("id")
		if err != nil || id <= 0 {
			c.BadRequest("invalid id")
			return
		}
		result, err := service.BackendDetail(id)
		if err != nil {
			respondServiceError(c, err)
			return
		}
		c.Success(result)
	}
}

func backendSave(runtime *bootstrap.Runtime) httpx.HandlerFunc {
	service := NewService(runtime)
	return func(c *httpx.Context) {
		var req BackendSaveRequest
		if err := c.BindJSON(&req); err != nil {
			c.Error(err)
			return
		}
		result, err := service.BackendSave(req)
		if err != nil {
			respondServiceError(c, err)
			return
		}
		c.Success(result)
	}
}

func backendDelete(runtime *bootstrap.Runtime) httpx.HandlerFunc {
	service := NewService(runtime)
	return func(c *httpx.Context) {
		var req bootstrap.IDsPayload
		if err := c.BindJSON(&req); err != nil {
			c.Error(err)
			return
		}
		result, err := service.BackendDelete(req.Values())
		if err != nil {
			respondServiceError(c, err)
			return
		}
		c.Success(result)
	}
}

func backendOperate(runtime *bootstrap.Runtime) httpx.HandlerFunc {
	service := NewService(runtime)
	return func(c *httpx.Context) {
		var req OperateRequest
		if err := c.BindJSON(&req); err != nil {
			c.Error(err)
			return
		}
		result, err := service.BackendOperate(req)
		if err != nil {
			respondServiceError(c, err)
			return
		}
		c.Success(result)
	}
}

func backendMoney(runtime *bootstrap.Runtime) httpx.HandlerFunc {
	service := NewService(runtime)
	return func(c *httpx.Context) {
		var req MoneyRequest
		if err := c.BindJSON(&req); err != nil {
			c.Error(err)
			return
		}
		result, err := service.UpdateMoney(req)
		if err != nil {
			respondServiceError(c, err)
			return
		}
		c.Success(result)
	}
}

func backendScore(runtime *bootstrap.Runtime) httpx.HandlerFunc {
	service := NewService(runtime)
	return func(c *httpx.Context) {
		var req ScoreRequest
		if err := c.BindJSON(&req); err != nil {
			c.Error(err)
			return
		}
		result, err := service.UpdateScore(req)
		if err != nil {
			respondServiceError(c, err)
			return
		}
		c.Success(result)
	}
}

func backendMoneyLogs(runtime *bootstrap.Runtime) httpx.HandlerFunc {
	service := NewService(runtime)
	return func(c *httpx.Context) {
		page, pageSize := bootstrap.Pagination(c)
		userID, _ := c.QueryInt64("user_id")
		result, err := service.MoneyLogs(LogListParams{UserID: userID, Page: page, PageSize: pageSize})
		if err != nil {
			respondServiceError(c, err)
			return
		}
		c.Success(result)
	}
}

func backendScoreLogs(runtime *bootstrap.Runtime) httpx.HandlerFunc {
	service := NewService(runtime)
	return func(c *httpx.Context) {
		page, pageSize := bootstrap.Pagination(c)
		userID, _ := c.QueryInt64("user_id")
		result, err := service.ScoreLogs(LogListParams{UserID: userID, Page: page, PageSize: pageSize})
		if err != nil {
			respondServiceError(c, err)
			return
		}
		c.Success(result)
	}
}
