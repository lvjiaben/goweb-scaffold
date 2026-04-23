package app_user_auth

import "github.com/lvjiaben/goweb-scaffold/internal/bootstrap"

type Module struct{}

func (Module) Name() string { return "app_user_auth" }

func (Module) Register(runtime *bootstrap.Runtime) error {
	public := runtime.AppPublicGroup.Group("/auth")
	authed := runtime.AppAuthedGroup.Group("/auth")
	public.POST("/login", login(runtime))
	public.POST("/register", register(runtime))
	authed.POST("/logout", logout(runtime))
	runtime.AppPublicGroup.Group("/user").POST("/login", login(runtime))
	runtime.AppProtectedGroup.Group("/user").POST("/logout", logout(runtime))
	return nil
}
