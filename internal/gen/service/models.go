package service

import (
	"time"

	sharedmodel "github.com/lvjiaben/goweb-scaffold/internal/shared/model"
	"gorm.io/datatypes"
)

var JSON = sharedmodel.JSON

const (
	MenuTypeMenu   = "menu"
	MenuTypeButton = "button"
	MenuTypeIframe = "iframe"
	MenuTypeLink   = "link"
)

type AdminUser struct {
	sharedmodel.BaseModel
	Username     string     `gorm:"column:username"`
	PasswordHash string     `gorm:"column:password_hash"`
	Nickname     string     `gorm:"column:nickname"`
	Status       int        `gorm:"column:status"`
	IsSuper      bool       `gorm:"column:is_super"`
	LastLoginAt  *time.Time `gorm:"column:last_login_at"`
	LastLoginIP  string     `gorm:"column:last_login_ip"`
}

func (AdminUser) TableName() string { return "admin_user" }

type AdminRole struct {
	sharedmodel.BaseModel
	Name   string `gorm:"column:name"`
	Code   string `gorm:"column:code"`
	Status int    `gorm:"column:status"`
}

func (AdminRole) TableName() string { return "admin_role" }

type AdminMenu struct {
	sharedmodel.BaseModel
	ParentID       int64  `gorm:"column:parent_id"`
	Name           string `gorm:"column:name"`
	EnName         string `gorm:"column:enname"`
	Title          string `gorm:"column:title"`
	Path           string `gorm:"column:path"`
	Component      string `gorm:"column:component"`
	MenuType       string `gorm:"column:menu_type"`
	PermissionCode string `gorm:"column:permission_code"`
	Iframe         string `gorm:"column:iframe"`
	External       string `gorm:"column:external"`
	Icon           string `gorm:"column:icon"`
	Sort           int    `gorm:"column:sort"`
	Visible        bool   `gorm:"column:visible"`
	Status         int    `gorm:"column:status"`
}

func (AdminMenu) TableName() string { return "admin_menu" }

type AdminUserRole struct {
	sharedmodel.BaseModel
	UserID int64 `gorm:"column:user_id"`
	RoleID int64 `gorm:"column:role_id"`
}

func (AdminUserRole) TableName() string { return "admin_user_role" }

type AdminRoleMenu struct {
	sharedmodel.BaseModel
	RoleID int64 `gorm:"column:role_id"`
	MenuID int64 `gorm:"column:menu_id"`
}

func (AdminRoleMenu) TableName() string { return "admin_role_menu" }

type AdminLoginLog struct {
	sharedmodel.BaseModel
	AdminUserID int64  `gorm:"column:admin_user_id"`
	Username    string `gorm:"column:username"`
	IP          string `gorm:"column:ip"`
	UserAgent   string `gorm:"column:user_agent"`
	Success     bool   `gorm:"column:success"`
	Remark      string `gorm:"column:remark"`
}

func (AdminLoginLog) TableName() string { return "admin_login_log" }

type AdminSession struct {
	sharedmodel.BaseModel
	AdminUserID int64     `gorm:"column:admin_user_id"`
	ExpiresAt   time.Time `gorm:"column:expires_at"`
	LastSeenAt  time.Time `gorm:"column:last_seen_at"`
	UserAgent   string    `gorm:"column:user_agent"`
	IP          string    `gorm:"column:ip"`
}

func (AdminSession) TableName() string { return "admin_session" }

type AppUser struct {
	sharedmodel.BaseModel
	Username     string `gorm:"column:username"`
	PasswordHash string `gorm:"column:password_hash"`
	Nickname     string `gorm:"column:nickname"`
	Email        string `gorm:"column:email"`
	Mobile       string `gorm:"column:mobile"`
	Status       int    `gorm:"column:status"`
}

func (AppUser) TableName() string { return "app_user" }

type AppUserSession struct {
	sharedmodel.BaseModel
	AppUserID  int64     `gorm:"column:app_user_id"`
	ExpiresAt  time.Time `gorm:"column:expires_at"`
	LastSeenAt time.Time `gorm:"column:last_seen_at"`
	UserAgent  string    `gorm:"column:user_agent"`
	IP         string    `gorm:"column:ip"`
}

func (AppUserSession) TableName() string { return "app_user_session" }

type SystemConfig struct {
	sharedmodel.BaseModel
	ConfigKey   string         `gorm:"column:config_key"`
	ConfigName  string         `gorm:"column:config_name"`
	ConfigValue datatypes.JSON `gorm:"column:config_value;type:jsonb"`
	Remark      string         `gorm:"column:remark"`
}

func (SystemConfig) TableName() string { return "system_config" }

type FileAttachment struct {
	sharedmodel.BaseModel
	OriginalName string `gorm:"column:original_name"`
	SavedName    string `gorm:"column:saved_name"`
	FilePath     string `gorm:"column:file_path"`
	FileURL      string `gorm:"column:file_url"`
	FileExt      string `gorm:"column:file_ext"`
	MimeType     string `gorm:"column:mime_type"`
	FileSize     int64  `gorm:"column:file_size"`
	UploaderKind string `gorm:"column:uploader_kind"`
	UploaderID   int64  `gorm:"column:uploader_id"`
}

func (FileAttachment) TableName() string { return "file_attachment" }

type CodegenHistory struct {
	sharedmodel.BaseModel
	ModuleName  string         `gorm:"column:module_name"`
	SourceTable string         `gorm:"column:table_name"`
	Status      string         `gorm:"column:status"`
	Payload     datatypes.JSON `gorm:"column:payload;type:jsonb"`
	Remark      string         `gorm:"column:remark"`
}

func (CodegenHistory) TableName() string { return "codegen_history" }
