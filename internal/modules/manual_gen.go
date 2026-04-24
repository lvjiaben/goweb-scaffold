package modules

import (
	"github.com/lvjiaben/goweb-scaffold/internal/bootstrap"
	admin_auth "github.com/lvjiaben/goweb-scaffold/internal/modules/admin/auth"
	admin_menu "github.com/lvjiaben/goweb-scaffold/internal/modules/admin/menu"
	admin_role "github.com/lvjiaben/goweb-scaffold/internal/modules/admin/role"
	admin_user "github.com/lvjiaben/goweb-scaffold/internal/modules/admin/user"
	app_auth "github.com/lvjiaben/goweb-scaffold/internal/modules/app/auth"
	app_user "github.com/lvjiaben/goweb-scaffold/internal/modules/app/user"
	system_attachment "github.com/lvjiaben/goweb-scaffold/internal/modules/system/attachment"
	system_codegen "github.com/lvjiaben/goweb-scaffold/internal/modules/system/codegen"
	system_common "github.com/lvjiaben/goweb-scaffold/internal/modules/system/common"
	system_config "github.com/lvjiaben/goweb-scaffold/internal/modules/system/config"
	system_home "github.com/lvjiaben/goweb-scaffold/internal/modules/system/home"
)

func RegisterManualModules(runtime *bootstrap.Runtime) error {
	return bootstrap.RegisterAll(
		runtime,
		admin_auth.Module{},
		system_common.Module{},
		system_home.Module{},
		admin_user.Module{},
		admin_role.Module{},
		admin_menu.Module{},
		system_config.Module{},
		system_attachment.Module{},
		app_auth.Module{},
		app_user.Module{},
		system_codegen.Module{},
	)
}
