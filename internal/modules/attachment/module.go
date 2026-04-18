package attachment

import (
	"strings"
	"time"

	"github.com/lvjiaben/goweb-core/httpx"
	"github.com/lvjiaben/goweb-scaffold/internal/bootstrap"
	"github.com/lvjiaben/goweb-scaffold/internal/modules/model"
	"gorm.io/gorm"
)

type Module struct{}

func (Module) Name() string { return "attachment" }

func (Module) Register(runtime *bootstrap.Runtime) error {
	runtime.AdminProtectedGroup.POST("/attachment/upload", upload(runtime), httpx.WithPermission("attachment.upload"))
	runtime.AdminProtectedGroup.GET("/attachment/list", list(runtime), httpx.WithPermission("attachment.list"))
	runtime.AdminProtectedGroup.POST("/attachment/delete", deleteFiles(runtime), httpx.WithPermission("attachment.delete"))
	return nil
}

func upload(runtime *bootstrap.Runtime) httpx.HandlerFunc {
	return func(c *httpx.Context) {
		file, header, err := c.MultipartFormFile("file")
		if err != nil {
			c.Error(err)
			return
		}

		saved, err := runtime.FileStore.Save(file, header, time.Now().Format("20060102"))
		if err != nil {
			c.Error(err)
			return
		}

		adminUser, _ := bootstrap.CurrentAdminUser(c)
		uploaderID := int64(0)
		if adminUser != nil {
			uploaderID = adminUser.ID
		}

		publicPrefix := strings.TrimRight(runtime.Config.Storage.PublicPrefix, "/")
		if publicPrefix == "" {
			publicPrefix = "/uploads"
		}

		record := model.FileAttachment{
			OriginalName: saved.OriginalName,
			SavedName:    saved.SavedName,
			FilePath:     saved.RelativePath,
			FileURL:      publicPrefix + "/" + strings.TrimLeft(saved.RelativePath, "/"),
			FileExt:      saved.Ext,
			MimeType:     header.Header.Get("Content-Type"),
			FileSize:     saved.Size,
			UploaderKind: "admin",
			UploaderID:   uploaderID,
		}
		if err := runtime.DB.Create(&record).Error; err != nil {
			c.Error(err)
			return
		}

		c.Success(map[string]any{
			"id":            record.ID,
			"original_name": record.OriginalName,
			"url":           record.FileURL,
			"size":          record.FileSize,
		})
	}
}

func list(runtime *bootstrap.Runtime) httpx.HandlerFunc {
	return func(c *httpx.Context) {
		page, pageSize := bootstrap.Pagination(c)
		keyword := bootstrap.LikeKeyword(c.Query("keyword"))

		query := runtime.DB.Model(&model.FileAttachment{}).Order("id DESC")
		if keyword != "" {
			query = query.Where("original_name ILIKE ?", keyword)
		}

		var total int64
		if err := query.Count(&total).Error; err != nil {
			c.Error(err)
			return
		}

		var rows []model.FileAttachment
		if err := query.Offset((page - 1) * pageSize).Limit(pageSize).Find(&rows).Error; err != nil {
			c.Error(err)
			return
		}

		items := make([]map[string]any, 0, len(rows))
		for _, row := range rows {
			items = append(items, map[string]any{
				"id":            row.ID,
				"original_name": row.OriginalName,
				"saved_name":    row.SavedName,
				"url":           row.FileURL,
				"file_path":     row.FilePath,
				"file_ext":      row.FileExt,
				"mime_type":     row.MimeType,
				"file_size":     row.FileSize,
				"uploader_kind": row.UploaderKind,
				"uploader_id":   row.UploaderID,
				"created_at":    row.CreatedAt,
			})
		}

		c.Success(map[string]any{
			"list":      items,
			"total":     total,
			"page":      page,
			"page_size": pageSize,
		})
	}
}

func deleteFiles(runtime *bootstrap.Runtime) httpx.HandlerFunc {
	return func(c *httpx.Context) {
		var req bootstrap.IDsPayload
		if err := c.BindJSON(&req); err != nil {
			c.Error(err)
			return
		}
		ids := req.Values()
		if len(ids) == 0 {
			c.BadRequest("ids is required")
			return
		}

		var rows []model.FileAttachment
		if err := runtime.DB.Where("id IN ?", ids).Find(&rows).Error; err != nil {
			c.Error(err)
			return
		}

		err := runtime.DB.Transaction(func(tx *gorm.DB) error {
			if err := tx.Where("id IN ?", ids).Delete(&model.FileAttachment{}).Error; err != nil {
				return err
			}
			for _, row := range rows {
				if err := runtime.FileStore.Delete(row.FilePath); err != nil {
					return err
				}
			}
			return nil
		})
		if err != nil {
			c.Error(err)
			return
		}
		c.Success(map[string]any{"deleted": len(ids)})
	}
}
