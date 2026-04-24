package app_user

import "github.com/lvjiaben/goweb-scaffold/internal/bootstrap"

type Module struct{}

func (Module) Name() string { return "app_user" }

func (Module) Register(runtime *bootstrap.Runtime) error {
	if err := registerBackend(runtime); err != nil {
		return err
	}
	if err := registerAPI(runtime); err != nil {
		return err
	}
	return nil
}
