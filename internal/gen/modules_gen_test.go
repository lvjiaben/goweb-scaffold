package gen

import (
	"io"
	"log/slog"
	"net/http"
	"testing"

	coreauth "github.com/lvjiaben/goweb-core/auth"
	"github.com/lvjiaben/goweb-core/httpx"
	"github.com/lvjiaben/goweb-core/validate"
	"github.com/lvjiaben/goweb-scaffold/internal/bootstrap"
)

func TestRegisteredRouteMethods(t *testing.T) {
	engine := httpx.NewEngine(slog.New(slog.NewTextHandler(io.Discard, nil)))
	runtime := &bootstrap.Runtime{
		Engine:         engine,
		Logger:         slog.New(slog.NewTextHandler(io.Discard, nil)),
		Validator:      validate.New(),
		AdminJWT:       coreauth.NewManager(coreauth.JWTConfig{Secret: "test-admin-secret"}),
		AppJWT:         coreauth.NewManager(coreauth.JWTConfig{Secret: "test-app-secret"}),
		CaptchaService: bootstrap.NewCaptchaService(0),
	}
	runtime.BackendPublicGroup = engine.Group("/backend")
	runtime.BackendAuthedGroup = engine.Group("/backend")
	runtime.BackendProtectedGroup = engine.Group("/backend")
	runtime.AppPublicGroup = engine.Group("/api")
	runtime.AppAuthedGroup = engine.Group("/api")
	runtime.AppProtectedGroup = engine.Group("/api")

	if err := RegisterGeneratedModules(runtime); err != nil {
		t.Fatalf("register modules: %v", err)
	}

	assertRouteMethod(t, engine.Routes(), http.MethodGet, "/backend/app/demo_article/list")
	assertRouteMethod(t, engine.Routes(), http.MethodPost, "/backend/app/demo_article/save")
	assertRouteMethod(t, engine.Routes(), http.MethodGet, "/backend/app/demo_notice/list")
}

func assertRouteMethod(t *testing.T, routes []*httpx.Route, method string, path string) {
	t.Helper()
	for _, route := range routes {
		if route.Method == method && route.Path == path {
			return
		}
	}
	t.Fatalf("route %s %s not registered", method, path)
}
