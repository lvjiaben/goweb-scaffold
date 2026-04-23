package admin_role

import (
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

func JSON(value []byte) datatypes.JSON {
	if len(value) == 0 {
		return datatypes.JSON([]byte("{}"))
	}
	return datatypes.JSON(value)
}

const (
	MenuTypeMenu   = "menu"
	MenuTypeButton = "button"
)

type BaseModel struct {
	ID        int64          `gorm:"primaryKey"`
	CreatedAt time.Time      `gorm:"column:created_at"`
	UpdatedAt time.Time      `gorm:"column:updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;index"`
	Ext       datatypes.JSON `gorm:"column:ext;type:jsonb"`
}

type AdminUser struct {
	BaseModel
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
	BaseModel
	Name   string `gorm:"column:name"`
	Code   string `gorm:"column:code"`
	Status int    `gorm:"column:status"`
}

func (AdminRole) TableName() string { return "admin_role" }

type AdminMenu struct {
	BaseModel
	ParentID       int64  `gorm:"column:parent_id"`
	Name           string `gorm:"column:name"`
	Title          string `gorm:"column:title"`
	Path           string `gorm:"column:path"`
	Component      string `gorm:"column:component"`
	MenuType       string `gorm:"column:menu_type"`
	PermissionCode string `gorm:"column:permission_code"`
	Icon           string `gorm:"column:icon"`
	Sort           int    `gorm:"column:sort"`
	Visible        bool   `gorm:"column:visible"`
	Status         int    `gorm:"column:status"`
}

func (AdminMenu) TableName() string { return "admin_menu" }

type AdminUserRole struct {
	BaseModel
	UserID int64 `gorm:"column:user_id"`
	RoleID int64 `gorm:"column:role_id"`
}

func (AdminUserRole) TableName() string { return "admin_user_role" }

type AdminRoleMenu struct {
	BaseModel
	RoleID int64 `gorm:"column:role_id"`
	MenuID int64 `gorm:"column:menu_id"`
}

func (AdminRoleMenu) TableName() string { return "admin_role_menu" }

type AdminLoginLog struct {
	BaseModel
	AdminUserID int64  `gorm:"column:admin_user_id"`
	Username    string `gorm:"column:username"`
	IP          string `gorm:"column:ip"`
	UserAgent   string `gorm:"column:user_agent"`
	Success     bool   `gorm:"column:success"`
	Remark      string `gorm:"column:remark"`
}

func (AdminLoginLog) TableName() string { return "admin_login_log" }

type AdminSession struct {
	BaseModel
	AdminUserID int64     `gorm:"column:admin_user_id"`
	ExpiresAt   time.Time `gorm:"column:expires_at"`
	LastSeenAt  time.Time `gorm:"column:last_seen_at"`
	UserAgent   string    `gorm:"column:user_agent"`
	IP          string    `gorm:"column:ip"`
}

func (AdminSession) TableName() string { return "admin_session" }

type AppUser struct {
	BaseModel
	Username     string `gorm:"column:username"`
	PasswordHash string `gorm:"column:password_hash"`
	Nickname     string `gorm:"column:nickname"`
	Email        string `gorm:"column:email"`
	Mobile       string `gorm:"column:mobile"`
	Status       int    `gorm:"column:status"`
}

func (AppUser) TableName() string { return "app_user" }

type AppUserSession struct {
	BaseModel
	AppUserID  int64     `gorm:"column:app_user_id"`
	ExpiresAt  time.Time `gorm:"column:expires_at"`
	LastSeenAt time.Time `gorm:"column:last_seen_at"`
	UserAgent  string    `gorm:"column:user_agent"`
	IP         string    `gorm:"column:ip"`
}

func (AppUserSession) TableName() string { return "app_user_session" }

type SystemConfig struct {
	BaseModel
	ConfigKey   string         `gorm:"column:config_key"`
	ConfigName  string         `gorm:"column:config_name"`
	ConfigValue datatypes.JSON `gorm:"column:config_value;type:jsonb"`
	Remark      string         `gorm:"column:remark"`
}

func (SystemConfig) TableName() string { return "system_config" }

type FileAttachment struct {
	BaseModel
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
	BaseModel
	ModuleName  string         `gorm:"column:module_name"`
	SourceTable string         `gorm:"column:table_name"`
	Status      string         `gorm:"column:status"`
	Payload     datatypes.JSON `gorm:"column:payload;type:jsonb"`
	Remark      string         `gorm:"column:remark"`
}

func (CodegenHistory) TableName() string { return "codegen_history" }
