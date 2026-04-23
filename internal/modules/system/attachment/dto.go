package attachment

import (
	"mime/multipart"
	"time"
)

type UploadRequest struct {
	File         multipart.File
	Header       *multipart.FileHeader
	Parent       string
	UploaderID   int64
	UploaderKind string
}

type UploadResponse struct {
	ID           int64  `json:"id"`
	OriginalName string `json:"original_name"`
	FileURL      string `json:"file_url"`
	FilePath     string `json:"file_path"`
	Parent       string `json:"parent"`
	URL          string `json:"url"`
	Size         int64  `json:"size"`
}

type ListParams struct {
	Page     int
	PageSize int
	Keyword  string
	Parent   string
}

type ListItem struct {
	ID           int64     `json:"id"`
	OriginalName string    `json:"original_name"`
	SavedName    string    `json:"saved_name"`
	URL          string    `json:"url"`
	FileURL      string    `json:"file_url"`
	FilePath     string    `json:"file_path"`
	Path         string    `json:"path"`
	Parent       string    `json:"parent"`
	FileExt      string    `json:"file_ext"`
	MimeType     string    `json:"mime_type"`
	FileSize     int64     `json:"file_size"`
	UploaderKind string    `json:"uploader_kind"`
	UploaderID   int64     `json:"uploader_id"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type DirectoryItem struct {
	Name  string `json:"name"`
	Path  string `json:"path"`
	Count int    `json:"count"`
}

type DeleteResult struct {
	Deleted int `json:"deleted"`
}

type listFilter struct {
	Search string
	Parent string
}
