package system_config

import (
	"github.com/lvjiaben/goweb-core/httpx"
	"github.com/lvjiaben/goweb-scaffold/internal/bootstrap"
)

type Module struct{}

func (Module) Name() string { return "system_config" }

func (Module) Register(runtime *bootstrap.Runtime) error {
	group := runtime.BackendProtectedGroup.Group("/system/config")
	group.GET("/list", list(runtime), httpx.WithPermission("system_config.list"))
	group.GET("/detail", detail(runtime), httpx.WithPermission("system_config.list"))
	group.POST("/save", save(runtime), httpx.WithPermission("system_config.save"))
	group.POST("/delete", deleteConfigs(runtime), httpx.WithPermission("system_config.delete"))
	return nil
}
