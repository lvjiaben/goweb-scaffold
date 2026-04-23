package admin_role

import (
	"github.com/lvjiaben/goweb-core/httpx"
	"github.com/lvjiaben/goweb-scaffold/internal/bootstrap"
)

type Module struct{}

func (Module) Name() string { return "admin_role" }

func (Module) Register(runtime *bootstrap.Runtime) error {
	group := runtime.BackendProtectedGroup.Group("/admin/role")
	group.GET("/list", list(runtime), httpx.WithPermission("admin_role.list"))
	group.GET("/detail", detail(runtime), httpx.WithPermission("admin_role.list"))
	group.GET("/options", options(runtime), httpx.WithPermission("admin_user.save"))
	group.POST("/save", save(runtime), httpx.WithPermission("admin_role.save"))
	group.POST("/delete", deleteRoles(runtime), httpx.WithPermission("admin_role.delete"))
	return nil
}
