package system_config

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

func (r *Repo) Count(filter configListFilter) (int64, error) {
	var total int64
	err := r.applyListFilter(r.db.Model(&SystemConfig{}), filter).Count(&total).Error
	return total, err
}

func (r *Repo) List(filter configListFilter, page int, pageSize int) ([]SystemConfig, error) {
	var rows []SystemConfig
	err := r.applyListFilter(r.db.Model(&SystemConfig{}), filter).
		Order("id DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&rows).Error
	return rows, err
}

func (r *Repo) FindByID(id int64) (SystemConfig, error) {
	var item SystemConfig
	err := r.db.First(&item, id).Error
	return item, err
}

func (r *Repo) CountByKey(key string, excludeID int64) (int64, error) {
	var count int64
	query := r.db.Model(&SystemConfig{}).Where("config_key = ?", key)
	if excludeID > 0 {
		query = query.Where("id <> ?", excludeID)
	}
	err := query.Count(&count).Error
	return count, err
}

func (r *Repo) Create(item *SystemConfig) error {
	return r.db.Create(item).Error
}

func (r *Repo) Update(item *SystemConfig, updates map[string]any) error {
	return r.db.Model(item).Updates(updates).Error
}

func (r *Repo) DeleteByIDs(ids []int64) error {
	return r.db.Where("id IN ?", ids).Delete(&SystemConfig{}).Error
}

func (r *Repo) applyListFilter(query *gorm.DB, filter configListFilter) *gorm.DB {
	if filter.Keyword != "" {
		query = query.Where("config_key ILIKE ? OR config_name ILIKE ?", filter.Keyword, filter.Keyword)
	}
	if filter.Key != "" {
		query = query.Where("config_key ILIKE ?", filter.Key)
	}
	if filter.Name != "" {
		query = query.Where("config_name ILIKE ?", filter.Name)
	}
	return query
}
