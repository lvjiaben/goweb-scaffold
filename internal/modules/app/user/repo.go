package app_user

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

func (r *Repo) UpdateProfile(userID int64, updates map[string]any) error {
	return r.db.Model(&AppUser{}).Where("id = ?", userID).Updates(updates).Error
}

func (r *Repo) UpdatePassword(userID int64, passwordHash string, updatedAt any) error {
	return r.db.Model(&AppUser{}).Where("id = ?", userID).Updates(map[string]any{
		"password_hash": passwordHash,
		"updated_at":    updatedAt,
	}).Error
}
