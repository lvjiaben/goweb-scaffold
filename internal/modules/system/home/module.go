package system_home

import "github.com/lvjiaben/goweb-scaffold/internal/bootstrap"

type Module struct{}

func (Module) Name() string { return "system_home" }

func (Module) Register(runtime *bootstrap.Runtime) error {
	group := runtime.BackendAuthedGroup.Group("/home")
	group.GET("/index", index(runtime))
	return nil
}
