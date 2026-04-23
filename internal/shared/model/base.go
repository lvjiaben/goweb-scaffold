package model

import (
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type BaseModel struct {
	ID        int64          `gorm:"primaryKey"`
	CreatedAt time.Time      `gorm:"column:created_at"`
	UpdatedAt time.Time      `gorm:"column:updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;index"`
}

type ExtField struct {
	Ext datatypes.JSON `gorm:"column:ext;type:jsonb"`
}
