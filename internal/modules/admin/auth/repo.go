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

func (r *Repo) FindUserByID(id int64) (AdminUser, error) {
	var user AdminUser
	err := r.db.First(&user, id).Error
	return user, err
}

func (r *Repo) UpdateProfile(userID int64, nickname string) error {
	updates := map[string]any{"updated_at": time.Now()}
	if nickname != "" {
		updates["nickname"] = nickname
	}
	return r.db.Model(&AdminUser{}).Where("id = ?", userID).Updates(updates).Error
}

func (r *Repo) UpdatePasswordHash(userID int64, hash string) error {
	return r.db.Model(&AdminUser{}).Where("id = ?", userID).Updates(map[string]any{
		"password_hash": hash,
		"updated_at":    time.Now(),
	}).Error
}

func (r *Repo) LoginLogs(userID int64, page int, pageSize int) ([]AdminLoginLog, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	query := r.db.Model(&AdminLoginLog{}).Where("admin_user_id = ?", userID)
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var logs []AdminLoginLog
	err := query.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&logs).Error
	return logs, total, err
}

func (r *Repo) MenuItems(identityRoleIDs []int64, isSuper bool) ([]AdminMenu, error) {
	var menus []AdminMenu
	query := r.db.Model(&AdminMenu{}).
		Where("menu_type <> ? AND status = ?", "button", 1).
		Order("sort DESC, id DESC")
	if !isSuper {
		if len(identityRoleIDs) == 0 {
			return menus, nil
		}
		query = query.Joins("JOIN admin_role_menu ON admin_role_menu.menu_id = admin_menu.id").
			Where("admin_role_menu.role_id IN ?", identityRoleIDs).
			Distinct("admin_menu.*")
	}
	err := query.Find(&menus).Error
	return menus, err
}
