package attachment

import sharedmodel "github.com/lvjiaben/goweb-scaffold/internal/shared/model"

type FileAttachment struct {
	sharedmodel.BaseModel
	OriginalName string `gorm:"column:original_name"`
	SavedName    string `gorm:"column:saved_name"`
	FilePath     string `gorm:"column:file_path"`
	FileURL      string `gorm:"column:file_url"`
	FileExt      string `gorm:"column:file_ext"`
	MimeType     string `gorm:"column:mime_type"`
	FileSize     int64  `gorm:"column:file_size"`
	UploaderKind string `gorm:"column:uploader_kind"`
	UploaderID   int64  `gorm:"column:uploader_id"`
}

func (FileAttachment) TableName() string { return "file_attachment" }
