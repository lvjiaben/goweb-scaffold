package gen

import (
	"github.com/lvjiaben/goweb-scaffold/internal/bootstrap"
	"github.com/lvjiaben/goweb-scaffold/internal/modules/admin_auth"
	"github.com/lvjiaben/goweb-scaffold/internal/modules/admin_menu"
	"github.com/lvjiaben/goweb-scaffold/internal/modules/admin_role"
	"github.com/lvjiaben/goweb-scaffold/internal/modules/admin_user"
	"github.com/lvjiaben/goweb-scaffold/internal/modules/app_user"
	"github.com/lvjiaben/goweb-scaffold/internal/modules/app_user_auth"
	"github.com/lvjiaben/goweb-scaffold/internal/modules/attachment"
	"github.com/lvjiaben/goweb-scaffold/internal/modules/codegen"
	"github.com/lvjiaben/goweb-scaffold/internal/modules/system_config"
)

func RegisterModules(runtime *bootstrap.Runtime) error {
	return bootstrap.RegisterAll(
		runtime,
		admin_auth.Module{},
		admin_user.Module{},
		admin_role.Module{},
		admin_menu.Module{},
		system_config.Module{},
		attachment.Module{},
		app_user_auth.Module{},
		app_user.Module{},
		codegen.Module{},
	)
}
