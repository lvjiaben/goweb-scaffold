package admin_auth

import "github.com/lvjiaben/goweb-scaffold/internal/bootstrap"

type Module struct{}

func (Module) Name() string { return "admin_auth" }

func (Module) Register(runtime *bootstrap.Runtime) error {
	public := runtime.BackendPublicGroup.Group("/auth")
	authed := runtime.BackendAuthedGroup.Group("/auth")
	public.POST("/login", login(runtime))
	authed.POST("/logout", logout(runtime))
	authed.GET("/me", me(runtime))
	authed.GET("/menus", menus(runtime))
	return nil
}
