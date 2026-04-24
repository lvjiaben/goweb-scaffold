package admin_menu

import sharedmodel "github.com/lvjiaben/goweb-scaffold/internal/shared/model"

const (
	MenuTypeMenu   = "menu"
	MenuTypeButton = "button"
	MenuTypeIframe = "iframe"
	MenuTypeLink   = "link"
)

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
	FixedTag       int    `gorm:"column:fixed_tag"`
	ShowTag        int    `gorm:"column:show_tag"`
}

func (AdminMenu) TableName() string { return "admin_menu" }

type AdminRoleMenu struct {
	sharedmodel.BaseModel
	RoleID int64 `gorm:"column:role_id"`
	MenuID int64 `gorm:"column:menu_id"`
}

func (AdminRoleMenu) TableName() string { return "admin_role_menu" }
