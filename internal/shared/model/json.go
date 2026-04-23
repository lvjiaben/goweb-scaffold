package model

import "gorm.io/datatypes"

func JSON(value []byte) datatypes.JSON {
	if len(value) == 0 {
		return datatypes.JSON([]byte("{}"))
	}
	return datatypes.JSON(value)
}
