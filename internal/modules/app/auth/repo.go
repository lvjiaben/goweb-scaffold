package app_user_auth

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

func (r *Repo) CreateUser(user *AppUser) error {
	return r.db.Create(user).Error
}

func (r *Repo) FindUserByUsername(username string) (AppUser, error) {
	var user AppUser
	err := r.db.Where("username = ? AND deleted_at IS NULL", username).First(&user).Error
	return user, err
}

func (r *Repo) CreateSession(session *AppUserSession) error {
	return r.db.Create(session).Error
}

func (r *Repo) DeleteSession(id int64) error {
	return r.db.Where("id = ?", id).Delete(&AppUserSession{}).Error
}
