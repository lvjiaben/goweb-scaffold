package admin_auth

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

func (r *Repo) FindUserByUsername(username string) (AdminUser, error) {
	var user AdminUser
	err := r.db.Where("username = ? AND deleted_at IS NULL", username).First(&user).Error
	return user, err
}

func (r *Repo) CreateSession(session *AdminSession) error {
	return r.db.Create(session).Error
}

func (r *Repo) DeleteSession(id int64) error {
	return r.db.Where("id = ?", id).Delete(&AdminSession{}).Error
}

func (r *Repo) UpdateLastLogin(userID int64, ip string, at time.Time) error {
	return r.db.Model(&AdminUser{}).Where("id = ?", userID).Updates(map[string]any{
		"last_login_at": at,
		"last_login_ip": ip,
		"updated_at":    at,
	}).Error
}

func (r *Repo) CreateLoginLog(log *AdminLoginLog) error {
	return r.db.Create(log).Error
}
