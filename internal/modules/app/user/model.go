package app_user

import sharedmodel "github.com/lvjiaben/goweb-scaffold/internal/shared/model"

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
