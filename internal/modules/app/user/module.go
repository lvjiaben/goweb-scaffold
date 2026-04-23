package app_user

import "github.com/lvjiaben/goweb-scaffold/internal/bootstrap"

type Module struct{}

func (Module) Name() string { return "app_user" }

func (Module) Register(runtime *bootstrap.Runtime) error {
	group := runtime.AppProtectedGroup.Group("/user")
	group.GET("/profile", profile(runtime))
	group.POST("/profile/save", profileSave(runtime))
	group.POST("/password/change", passwordChange(runtime))
	return nil
}
