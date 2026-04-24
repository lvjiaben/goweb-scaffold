package system_home

import (
	"time"

	"github.com/lvjiaben/goweb-scaffold/internal/bootstrap"
	"gorm.io/gorm"
)

type Repo struct {
	db *gorm.DB
}

func NewRepo(runtime *bootstrap.Runtime) *Repo {
	return &Repo{db: runtime.DB}
}

func (r *Repo) CountAppUsers() (int64, error) {
	var total int64
	return total, r.db.Model(&AppUser{}).Count(&total).Error
}

func (r *Repo) CountAppUsersCreatedSince(at time.Time) (int64, error) {
	var total int64
	return total, r.db.Model(&AppUser{}).Where("created_at >= ?", at).Count(&total).Error
}

func (r *Repo) CountDisabledAppUsers() (int64, error) {
	var total int64
	return total, r.db.Model(&AppUser{}).Where("status <> ?", 1).Count(&total).Error
}

func (r *Repo) CountAdmins() (int64, error) {
	var total int64
	return total, r.db.Model(&AdminUser{}).Count(&total).Error
}

func (r *Repo) CountUploads() (int64, error) {
	var total int64
	return total, r.db.Model(&FileAttachment{}).Count(&total).Error
}

func (r *Repo) CountUploadsCreatedSince(at time.Time) (int64, error) {
	var total int64
	return total, r.db.Model(&FileAttachment{}).Where("created_at >= ?", at).Count(&total).Error
}
