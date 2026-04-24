package model

import (
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type BaseModel struct {
	ID        int64          `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `gorm:"column:created_at" json:"created_at"`
	UpdatedAt time.Time      `gorm:"column:updated_at" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;index" json:"deleted_at,omitempty"`
}

type ExtField struct {
	Ext datatypes.JSON `gorm:"column:ext;type:jsonb"`
}
