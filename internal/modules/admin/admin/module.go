package admin_admin

import (
	"github.com/lvjiaben/goweb-core/httpx"
	"github.com/lvjiaben/goweb-scaffold/internal/bootstrap"
)

type Module struct{}

func (Module) Name() string { return "admin_user" }

func (Module) Register(runtime *bootstrap.Runtime) error {
	group := runtime.BackendProtectedGroup.Group("/admin/admin")
	group.GET("/list", list(runtime), httpx.WithPermission("admin_user.list"))
	group.GET("/detail", detail(runtime), httpx.WithPermission("admin_user.list"))
	group.POST("/save", save(runtime), httpx.WithPermission("admin_user.save"))
	group.POST("/delete", deleteUsers(runtime), httpx.WithPermission("admin_user.delete"))
	legacy := runtime.BackendProtectedGroup.Group("/admin/user")
	legacy.GET("/list", list(runtime), httpx.WithPermission("admin_user.list"))
	legacy.GET("/detail", detail(runtime), httpx.WithPermission("admin_user.list"))
	legacy.POST("/save", save(runtime), httpx.WithPermission("admin_user.save"))
	legacy.POST("/delete", deleteUsers(runtime), httpx.WithPermission("admin_user.delete"))
	return nil
}
