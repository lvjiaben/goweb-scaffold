package system_home

import "github.com/lvjiaben/goweb-scaffold/internal/shared/model"

type AppUser struct {
	model.BaseModel
	Status int `gorm:"column:status"`
}

func (AppUser) TableName() string { return "app_user" }

type AdminUser struct {
	model.BaseModel
	Status int `gorm:"column:status"`
}

func (AdminUser) TableName() string { return "admin_user" }

type FileAttachment struct {
	model.BaseModel
}

func (FileAttachment) TableName() string { return "file_attachment" }
