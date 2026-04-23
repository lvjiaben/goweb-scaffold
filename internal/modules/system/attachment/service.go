package attachment

import (
	"strings"
	"time"

	corefiles "github.com/lvjiaben/goweb-core/files"
	"github.com/lvjiaben/goweb-scaffold/internal/bootstrap"
)

type Service struct {
	repo         *Repo
	fileStore    corefiles.LocalStore
	publicPrefix string
}

func NewService(runtime *bootstrap.Runtime) *Service {
	publicPrefix := "/uploads"
	if runtime != nil && runtime.Config != nil {
		publicPrefix = strings.TrimRight(runtime.Config.Storage.PublicPrefix, "/")
	}
	if publicPrefix == "" {
		publicPrefix = "/uploads"
	}
	return &Service{
		repo:         NewRepo(runtime),
		fileStore:    runtime.FileStore,
		publicPrefix: publicPrefix,
	}
}

func (s *Service) Upload(req UploadRequest) (UploadResponse, error) {
	parent := normalizeDirectory(req.Parent)
	saveDir := time.Now().Format("20060102")
	if parent != "" {
		saveDir = parent
	}

	saved, err := s.fileStore.Save(req.File, req.Header, saveDir)
	if err != nil {
		return UploadResponse{}, err
	}
	record := FileAttachment{
		OriginalName: saved.OriginalName,
		SavedName:    saved.SavedName,
		FilePath:     saved.RelativePath,
		FileURL:      s.publicPrefix + "/" + strings.TrimLeft(saved.RelativePath, "/"),
		FileExt:      saved.Ext,
		MimeType:     req.Header.Header.Get("Content-Type"),
		FileSize:     saved.Size,
		UploaderKind: req.UploaderKind,
		UploaderID:   req.UploaderID,
	}
	if record.UploaderKind == "" {
		record.UploaderKind = "admin"
	}
	if err := s.repo.Create(&record); err != nil {
		return UploadResponse{}, err
	}
	return UploadResponse{
		ID:           record.ID,
		OriginalName: record.OriginalName,
		FileURL:      record.FileURL,
		FilePath:     record.FilePath,
		Parent:       fileParent(record.FilePath),
		URL:          record.FileURL,
		Size:         record.FileSize,
	}, nil
}

func (s *Service) Directories() (map[string]any, error) {
	rows, err := s.repo.FilePaths()
	if err != nil {
		return nil, err
	}
	counts := map[string]int{}
	for _, row := range rows {
		counts[fileParent(row.FilePath)]++
	}

	items := make([]DirectoryItem, 0, len(counts)+1)
	items = append(items, DirectoryItem{Name: "全部", Path: "", Count: len(rows)})
	for path, count := range counts {
		name := path
		if name == "" {
			name = "根目录"
		}
		items = append(items, DirectoryItem{Name: name, Path: path, Count: count})
	}
	return map[string]any{"list": items}, nil
}

func (s *Service) List(params ListParams) (map[string]any, error) {
	filter := listFilter{
		Search: bootstrap.LikeKeyword(params.Keyword),
		Parent: normalizeDirectory(params.Parent),
	}
	total, err := s.repo.Count(filter)
	if err != nil {
		return nil, err
	}
	rows, err := s.repo.List(filter, params.Page, params.PageSize)
	if err != nil {
		return nil, err
	}
	items := make([]ListItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, ListItem{
			ID:           row.ID,
			OriginalName: row.OriginalName,
			SavedName:    row.SavedName,
			URL:          row.FileURL,
			FileURL:      row.FileURL,
			FilePath:     row.FilePath,
			Path:         row.FilePath,
			Parent:       fileParent(row.FilePath),
			FileExt:      row.FileExt,
			MimeType:     row.MimeType,
			FileSize:     row.FileSize,
			UploaderKind: row.UploaderKind,
			UploaderID:   row.UploaderID,
			CreatedAt:    row.CreatedAt,
			UpdatedAt:    row.UpdatedAt,
		})
	}
	return bootstrap.PagedResult(items, total, params.Page, params.PageSize), nil
}

func (s *Service) DeleteFiles(ids []int64) (DeleteResult, error) {
	ids = bootstrap.NormalizeIDs(ids...)
	if len(ids) == 0 {
		return DeleteResult{}, validationError("ids is required")
	}
	rows, err := s.repo.FindByIDs(ids)
	if err != nil {
		return DeleteResult{}, err
	}
	if err := s.repo.DeleteByIDs(ids); err != nil {
		return DeleteResult{}, err
	}
	for _, row := range rows {
		if err := s.fileStore.Delete(row.FilePath); err != nil {
			return DeleteResult{}, err
		}
	}
	return DeleteResult{Deleted: len(ids)}, nil
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

type validationError string

func (e validationError) Error() string {
	return string(e)
}

func isValidationError(err error) bool {
	_, ok := err.(validationError)
	return ok
}
