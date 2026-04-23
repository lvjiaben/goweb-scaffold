package app_user_auth

import (
	"time"

	sharedmodel "github.com/lvjiaben/goweb-scaffold/internal/shared/model"
)

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
