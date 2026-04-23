package system_config

import (
	sharedmodel "github.com/lvjiaben/goweb-scaffold/internal/shared/model"
	"gorm.io/datatypes"
)

type SystemConfig struct {
	sharedmodel.BaseModel
	ConfigKey   string         `gorm:"column:config_key"`
	ConfigName  string         `gorm:"column:config_name"`
	ConfigValue datatypes.JSON `gorm:"column:config_value;type:jsonb"`
	Remark      string         `gorm:"column:remark"`
}

func (SystemConfig) TableName() string { return "system_config" }
