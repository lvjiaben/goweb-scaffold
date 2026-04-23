package codegen

import (
	sharedmodel "github.com/lvjiaben/goweb-scaffold/internal/shared/model"
	"gorm.io/datatypes"
)

type CodegenHistory struct {
	sharedmodel.BaseModel
	ModuleName  string         `gorm:"column:module_name"`
	SourceTable string         `gorm:"column:table_name"`
	Status      string         `gorm:"column:status"`
	Payload     datatypes.JSON `gorm:"column:payload;type:jsonb"`
	Remark      string         `gorm:"column:remark"`
}

func (CodegenHistory) TableName() string { return "codegen_history" }
