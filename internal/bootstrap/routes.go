package bootstrap

import (
	"log/slog"

	"github.com/lvjiaben/goweb-core/httpx"
)

func LogRoutes(logger *slog.Logger, routes []*httpx.Route) {
	if logger == nil {
		return
	}
	for _, route := range routes {
		logger.Info(
			"registered route",
			"method", route.Method,
			"path", route.Path,
			"permission_code", route.PermissionCode,
		)
	}
}
