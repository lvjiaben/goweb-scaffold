package attachment

import (
	"strings"
	"time"

	"github.com/lvjiaben/goweb-core/httpx"
	"github.com/lvjiaben/goweb-scaffold/internal/bootstrap"
	"gorm.io/gorm"
)

func upload(runtime *bootstrap.Runtime) httpx.HandlerFunc {
	return func(c *httpx.Context) {
		file, header, err := c.MultipartFormFile("file")
		if err != nil {
			c.Error(err)
			return
		}

		parent := normalizeDirectory(c.Request.FormValue("parent"))
		saveDir := time.Now().Format("20060102")
		if parent != "" {
			saveDir = parent
		}

		saved, err := runtime.FileStore.Save(file, header, saveDir)
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

		record := FileAttachment{
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
			"file_url":      record.FileURL,
			"file_path":     record.FilePath,
			"parent":        fileParent(record.FilePath),
			"url":           record.FileURL,
			"size":          record.FileSize,
		})
	}
}

func directories(runtime *bootstrap.Runtime) httpx.HandlerFunc {
	return func(c *httpx.Context) {
		var rows []FileAttachment
		if err := runtime.DB.Select("id", "file_path").Order("file_path ASC").Find(&rows).Error; err != nil {
			c.Error(err)
			return
		}

		counts := map[string]int{}
		for _, row := range rows {
			counts[fileParent(row.FilePath)]++
		}

		items := make([]map[string]any, 0, len(counts)+1)
		items = append(items, map[string]any{
			"name":  "全部",
			"path":  "",
			"count": len(rows),
		})
		for path, count := range counts {
			name := path
			if name == "" {
				name = "根目录"
			}
			items = append(items, map[string]any{
				"name":  name,
				"path":  path,
				"count": count,
			})
		}

		c.Success(map[string]any{"list": items})
	}
}

func list(runtime *bootstrap.Runtime) httpx.HandlerFunc {
	return func(c *httpx.Context) {
		page, pageSize := bootstrap.Pagination(c)
		search := bootstrap.LikeKeyword(bootstrap.SearchKeyword(c))
		parent := normalizeDirectory(bootstrap.QueryFirst(c, "parent", "directory"))

		query := runtime.DB.Model(&FileAttachment{}).Order("id DESC")
		if search != "" {
			query = query.Where("original_name ILIKE ? OR file_path ILIKE ?", search, search)
		}
		if parent != "" {
			query = query.Where("file_path ILIKE ?", parent+"/%")
		}

		var total int64
		if err := query.Count(&total).Error; err != nil {
			c.Error(err)
			return
		}

		var rows []FileAttachment
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
				"file_url":      row.FileURL,
				"file_path":     row.FilePath,
				"path":          row.FilePath,
				"parent":        fileParent(row.FilePath),
				"file_ext":      row.FileExt,
				"mime_type":     row.MimeType,
				"file_size":     row.FileSize,
				"uploader_kind": row.UploaderKind,
				"uploader_id":   row.UploaderID,
				"created_at":    row.CreatedAt,
				"updated_at":    row.UpdatedAt,
			})
		}

		c.Success(bootstrap.PagedResult(items, total, page, pageSize))
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

		var rows []FileAttachment
		if err := runtime.DB.Where("id IN ?", ids).Find(&rows).Error; err != nil {
			c.Error(err)
			return
		}

		err := runtime.DB.Transaction(func(tx *gorm.DB) error {
			if err := tx.Where("id IN ?", ids).Delete(&FileAttachment{}).Error; err != nil {
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

func normalizeDirectory(value string) string {
	return strings.Trim(strings.TrimSpace(value), "/")
}

func fileParent(path string) string {
	path = normalizeDirectory(path)
	parts := strings.Split(path, "/")
	if len(parts) <= 1 {
		return ""
	}
	return strings.Join(parts[:len(parts)-1], "/")
}
