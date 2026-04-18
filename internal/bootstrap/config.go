package bootstrap

import (
	coreauth "github.com/lvjiaben/goweb-core/auth"
	coreconfig "github.com/lvjiaben/goweb-core/config"
	coredb "github.com/lvjiaben/goweb-core/db"
	"github.com/lvjiaben/goweb-core/logx"
)

type Config struct {
	App struct {
		Name string `yaml:"name"`
		Addr string `yaml:"addr"`
	} `yaml:"app"`
	Log      logx.Config           `yaml:"log"`
	Database coredb.PostgresConfig `yaml:"database"`
	JWT      JWTConfig             `yaml:"jwt"`
	CORS     CORSConfig            `yaml:"cors"`
	Storage  StorageConfig         `yaml:"storage"`
}

type JWTConfig struct {
	Admin coreauth.JWTConfig `yaml:"admin"`
	App   coreauth.JWTConfig `yaml:"app"`
}

type CORSConfig struct {
	AllowOrigins     []string `yaml:"allow_origins"`
	AllowCredentials bool     `yaml:"allow_credentials"`
	MaxAgeSeconds    int      `yaml:"max_age_seconds"`
}

type StorageConfig struct {
	UploadDir    string `yaml:"upload_dir"`
	PublicPrefix string `yaml:"public_prefix"`
}

func LoadConfig(path string) (*Config, error) {
	var cfg Config
	if err := coreconfig.LoadYAML(path, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
