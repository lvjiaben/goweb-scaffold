package system_home

import (
	"strconv"

	"github.com/lvjiaben/goweb-core/httpx"
	"github.com/lvjiaben/goweb-scaffold/internal/bootstrap"
)

func index(runtime *bootstrap.Runtime) httpx.HandlerFunc {
	service := NewService(runtime)
	return func(c *httpx.Context) {
		days, _ := strconv.Atoi(c.Query("time"))
		result, err := service.Dashboard(days)
		if err != nil {
			c.Error(err)
			return
		}
		c.Success(result)
	}
}
