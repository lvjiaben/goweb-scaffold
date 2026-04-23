package admin_menu

import (
	"github.com/lvjiaben/goweb-core/httpx"
	"github.com/lvjiaben/goweb-scaffold/internal/bootstrap"
)

type Module struct{}

func (Module) Name() string { return "admin_menu" }

func (Module) Register(runtime *bootstrap.Runtime) error {
	group := runtime.BackendProtectedGroup.Group("/admin/menu")
	group.GET("/list", list(runtime), httpx.WithPermission("admin_menu.list"))
	group.GET("/detail", detail(runtime), httpx.WithPermission("admin_menu.list"))
	group.GET("/tree", tree(runtime), httpx.WithPermission("admin_role.save|admin_menu.save"))
	group.GET("/options", options(runtime), httpx.WithPermission("admin_menu.save"))
	group.POST("/save", save(runtime), httpx.WithPermission("admin_menu.save"))
	group.POST("/delete", deleteMenus(runtime), httpx.WithPermission("admin_menu.delete"))
	return nil
}
