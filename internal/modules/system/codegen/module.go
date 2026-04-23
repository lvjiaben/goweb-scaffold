package codegen

import (
	"github.com/lvjiaben/goweb-core/httpx"
	"github.com/lvjiaben/goweb-scaffold/internal/bootstrap"
)

type Module struct{}

func (Module) Name() string { return "codegen" }

func (Module) Register(runtime *bootstrap.Runtime) error {
	group := runtime.BackendProtectedGroup.Group("/system/codegen")
	group.GET("/list", list(runtime), httpx.WithPermission("codegen.list"))
	group.GET("/modules", modules(runtime), httpx.WithPermission("codegen.list"))
	group.GET("/tables", tables(runtime), httpx.WithPermission("codegen.list"))
	group.GET("/table-columns", tableColumns(runtime), httpx.WithPermission("codegen.list"))
	group.GET("/export", exportFile(runtime), httpx.WithPermission("codegen.list"))
	group.POST("/preview", preview(runtime), httpx.WithPermission("codegen.save"))
	group.POST("/diff", diff(runtime), httpx.WithPermission("codegen.save"))
	group.POST("/check-breaking", checkBreaking(runtime), httpx.WithPermission("codegen.save"))
	group.POST("/generate", generate(runtime), httpx.WithPermission("codegen.save"))
	group.POST("/regenerate", regenerate(runtime), httpx.WithPermission("codegen.save"))
	group.POST("/remove", remove(runtime), httpx.WithPermission("codegen.delete"))
	group.POST("/save", save(runtime), httpx.WithPermission("codegen.save"))
	group.POST("/delete", deleteHistory(runtime), httpx.WithPermission("codegen.delete"))
	return nil
}
