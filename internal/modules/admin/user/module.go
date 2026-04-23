package admin_user

import (
	"github.com/lvjiaben/goweb-core/httpx"
	"github.com/lvjiaben/goweb-scaffold/internal/bootstrap"
)

type Module struct{}

func (Module) Name() string { return "admin_user" }

func (Module) Register(runtime *bootstrap.Runtime) error {
	group := runtime.BackendProtectedGroup.Group("/admin/user")
	group.GET("/list", list(runtime), httpx.WithPermission("admin_user.list"))
	group.GET("/detail", detail(runtime), httpx.WithPermission("admin_user.list"))
	group.POST("/save", save(runtime), httpx.WithPermission("admin_user.save"))
	group.POST("/delete", deleteUsers(runtime), httpx.WithPermission("admin_user.delete"))
	return nil
}
