package bootstrap

import (
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	coreauth "github.com/lvjiaben/goweb-core/auth"
	coredb "github.com/lvjiaben/goweb-core/db"
	corefiles "github.com/lvjiaben/goweb-core/files"
	"github.com/lvjiaben/goweb-core/httpx"
	"github.com/lvjiaben/goweb-core/logx"
	corerbac "github.com/lvjiaben/goweb-core/rbac"
	"github.com/lvjiaben/goweb-core/validate"
	"gorm.io/gorm"
)

type Runtime struct {
	Config   *Config
	Logger   *slog.Logger
	RepoRoot string

	DB        *gorm.DB
	Engine    *httpx.Engine
	Validator *validate.Validator

	AdminJWT *coreauth.Manager
	AppJWT   *coreauth.Manager

	FileStore         corefiles.LocalStore
	CaptchaService    *CaptchaService
	PermissionService *PermissionService

	AdminPublicGroup    *httpx.Group
	AdminCommonGroup    *httpx.Group
	AdminAuthGroup      *httpx.Group
	AdminProtectedGroup *httpx.Group
	AppPublicGroup      *httpx.Group
	AppCommonGroup      *httpx.Group
	AppAuthGroup        *httpx.Group
	AppUserGroup        *httpx.Group
}

func NewRuntime(configPath string) (*Runtime, error) {
	cfg, err := LoadConfig(configPath)
	if err != nil {
		return nil, err
	}

	logger := logx.New(cfg.Log)
	engine := httpx.NewEngine(logger)
	engine.Use(
		httpx.RequestID(),
		httpx.Logger(logger),
		httpx.Recover(logger),
		httpx.CORS(httpx.CORSConfig{
			AllowOrigins:     cfg.CORS.AllowOrigins,
			AllowCredentials: cfg.CORS.AllowCredentials,
			MaxAgeSeconds:    cfg.CORS.MaxAgeSeconds,
		}),
	)

	database, err := coredb.OpenPostgres(cfg.Database, logger)
	if err != nil {
		return nil, err
	}

	if err := os.MkdirAll(filepath.Clean(cfg.Storage.UploadDir), 0o755); err != nil {
		return nil, err
	}

	repoRoot, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	runtime := &Runtime{
		Config:         cfg,
		Logger:         logger,
		RepoRoot:       repoRoot,
		DB:             database,
		Engine:         engine,
		Validator:      validate.New(),
		AdminJWT:       coreauth.NewManager(cfg.JWT.Admin),
		AppJWT:         coreauth.NewManager(cfg.JWT.App),
		FileStore:      corefiles.LocalStore{BaseDir: cfg.Storage.UploadDir},
		CaptchaService: NewCaptchaService(5 * time.Minute),
	}
	runtime.PermissionService = NewPermissionService(runtime.DB)
	runtime.initGroups()
	return runtime, nil
}

func (r *Runtime) initGroups() {
	r.AdminPublicGroup = r.Engine.Group("/admin-api/auth")
	r.AdminCommonGroup = r.Engine.Group("/admin-api/common")
	r.AdminAuthGroup = r.Engine.Group("/admin-api/auth", r.AdminAuthMiddleware())
	r.AdminProtectedGroup = r.Engine.Group("/admin-api", r.AdminAuthMiddleware(), corerbac.RequirePermission(r.PermissionService))

	r.AppPublicGroup = r.Engine.Group("/api/auth")
	r.AppCommonGroup = r.Engine.Group("/api/common")
	r.AppAuthGroup = r.Engine.Group("/api/auth", r.AppUserAuthMiddleware())
	r.AppUserGroup = r.Engine.Group("/api/user", r.AppUserAuthMiddleware())
}

func (r *Runtime) Handler() http.Handler {
	fileServer := http.StripPrefix(
		r.Config.Storage.PublicPrefix,
		http.FileServer(http.Dir(r.Config.Storage.UploadDir)),
	)

	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if prefix := strings.TrimSpace(r.Config.Storage.PublicPrefix); prefix != "" && strings.HasPrefix(req.URL.Path, prefix) {
			fileServer.ServeHTTP(w, req)
			return
		}
		r.Engine.ServeHTTP(w, req)
	})
}
