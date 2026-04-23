package attachment

import (
	"github.com/lvjiaben/goweb-scaffold/internal/bootstrap"
	"gorm.io/gorm"
)

type Repo struct {
	db *gorm.DB
}

func NewRepo(runtime *bootstrap.Runtime) *Repo {
	return &Repo{db: runtime.DB}
}

func (r *Repo) Create(record *FileAttachment) error {
	return r.db.Create(record).Error
}

func (r *Repo) FilePaths() ([]FileAttachment, error) {
	var rows []FileAttachment
	err := r.db.Select("id", "file_path").Order("file_path ASC").Find(&rows).Error
	return rows, err
}

func (r *Repo) Count(filter listFilter) (int64, error) {
	var total int64
	err := r.applyListFilter(r.db.Model(&FileAttachment{}), filter).Count(&total).Error
	return total, err
}

func (r *Repo) List(filter listFilter, page int, pageSize int) ([]FileAttachment, error) {
	var rows []FileAttachment
	err := r.applyListFilter(r.db.Model(&FileAttachment{}), filter).
		Order("id DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&rows).Error
	return rows, err
}

func (r *Repo) FindByIDs(ids []int64) ([]FileAttachment, error) {
	var rows []FileAttachment
	err := r.db.Where("id IN ?", ids).Find(&rows).Error
	return rows, err
}

func (r *Repo) DeleteByIDs(ids []int64) error {
	return r.db.Where("id IN ?", ids).Delete(&FileAttachment{}).Error
}

func (r *Repo) applyListFilter(query *gorm.DB, filter listFilter) *gorm.DB {
	if filter.Search != "" {
		query = query.Where("original_name ILIKE ? OR file_path ILIKE ?", filter.Search, filter.Search)
	}
	if filter.Parent != "" {
		query = query.Where("file_path ILIKE ?", filter.Parent+"/%")
	}
	return query
}
