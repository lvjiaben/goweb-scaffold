package admin_role

import sharedmodel "github.com/lvjiaben/goweb-scaffold/internal/shared/model"

type AdminRole struct {
	sharedmodel.BaseModel
	Name   string `gorm:"column:name"`
	Code   string `gorm:"column:code"`
	Status int    `gorm:"column:status"`
}

func (AdminRole) TableName() string { return "admin_role" }

type AdminRoleMenu struct {
	sharedmodel.BaseModel
	RoleID int64 `gorm:"column:role_id"`
	MenuID int64 `gorm:"column:menu_id"`
}

func (AdminRoleMenu) TableName() string { return "admin_role_menu" }

type AdminUserRole struct {
	sharedmodel.BaseModel
	UserID int64 `gorm:"column:user_id"`
	RoleID int64 `gorm:"column:role_id"`
}

func (AdminUserRole) TableName() string { return "admin_user_role" }
