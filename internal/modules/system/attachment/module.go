package attachment

import (
	"github.com/lvjiaben/goweb-core/httpx"
	"github.com/lvjiaben/goweb-scaffold/internal/bootstrap"
)

type Module struct{}

func (Module) Name() string { return "attachment" }

func (Module) Register(runtime *bootstrap.Runtime) error {
	group := runtime.BackendProtectedGroup.Group("/system/attachment")
	group.POST("/upload", upload(runtime), httpx.WithPermission("attachment.upload"))
	group.GET("/directories", directories(runtime), httpx.WithPermission("attachment.list"))
	group.GET("/list", list(runtime), httpx.WithPermission("attachment.list"))
	group.POST("/delete", deleteFiles(runtime), httpx.WithPermission("attachment.delete"))
	return nil
}
