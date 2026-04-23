package common

import "github.com/lvjiaben/goweb-scaffold/internal/bootstrap"

type Module struct{}

func (Module) Name() string { return "common" }

func (Module) Register(runtime *bootstrap.Runtime) error {
	runtime.BackendPublicGroup.Group("/common").POST("/captcha", captcha(runtime))
	runtime.AppPublicGroup.Group("/common").POST("/captcha", captcha(runtime))
	return nil
}
