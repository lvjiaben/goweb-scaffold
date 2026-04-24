package admin_admin

import (
	"time"

	sharedmodel "github.com/lvjiaben/goweb-scaffold/internal/shared/model"
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

type AdminUserRole struct {
	sharedmodel.BaseModel
	UserID int64 `gorm:"column:user_id"`
	RoleID int64 `gorm:"column:role_id"`
}

func (AdminUserRole) TableName() string { return "admin_user_role" }

type AdminSession struct {
	sharedmodel.BaseModel
	AdminUserID int64     `gorm:"column:admin_user_id"`
	ExpiresAt   time.Time `gorm:"column:expires_at"`
	LastSeenAt  time.Time `gorm:"column:last_seen_at"`
	UserAgent   string    `gorm:"column:user_agent"`
	IP          string    `gorm:"column:ip"`
}

func (AdminSession) TableName() string { return "admin_session" }
