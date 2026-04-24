package admin_user

import (
	"github.com/lvjiaben/goweb-scaffold/internal/bootstrap"
	sharedquery "github.com/lvjiaben/goweb-scaffold/internal/shared/query"
	"gorm.io/gorm"
)

type Repo struct {
	db *gorm.DB
}

func NewRepo(runtime *bootstrap.Runtime) *Repo {
	return &Repo{db: runtime.DB}
}

func (r *Repo) WithTransaction(fn func(tx *Repo) error) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		return fn(&Repo{db: tx})
	})
}

func (r *Repo) Count(filter userListFilter) (int64, error) {
	var total int64
	err := r.applyListFilter(r.db.Model(&AdminUser{}), filter).Count.Count(&total).Error
	return total, err
}

func (r *Repo) List(filter userListFilter, page int, pageSize int) ([]AdminUser, error) {
	var users []AdminUser
	result := r.applyListFilter(r.db.Model(&AdminUser{}), filter)
	err := result.Query.Offset((page - 1) * pageSize).Limit(pageSize).Find(&users).Error
	return users, err
}

func (r *Repo) FindByID(id int64) (AdminUser, error) {
	var user AdminUser
	err := r.db.First(&user, id).Error
	return user, err
}

func (r *Repo) RoleIDsByUserID(userID int64) ([]int64, error) {
	var roleIDs []int64
	err := r.db.Model(&AdminUserRole{}).Where("user_id = ?", userID).Pluck("role_id", &roleIDs).Error
	return roleIDs, err
}

func (r *Repo) UserRoleMappings(userIDs []int64) ([]AdminUserRole, error) {
	var mappings []AdminUserRole
	if len(userIDs) == 0 {
		return mappings, nil
	}
	err := r.db.Where("user_id IN ?", userIDs).Find(&mappings).Error
	return mappings, err
}

func (r *Repo) RolesByIDs(roleIDs []int64) ([]AdminRole, error) {
	var roles []AdminRole
	ids := bootstrap.NormalizeIDs(roleIDs...)
	if len(ids) == 0 {
		return roles, nil
	}
	err := r.db.Where("id IN ?", ids).Find(&roles).Error
	return roles, err
}

func (r *Repo) CountByUsername(username string, excludeID int64) (int64, error) {
	var count int64
	query := r.db.Model(&AdminUser{}).Where("username = ?", username)
	if excludeID > 0 {
		query = query.Where("id <> ?", excludeID)
	}
	err := query.Count(&count).Error
	return count, err
}

func (r *Repo) Create(user *AdminUser) error {
	return r.db.Create(user).Error
}

func (r *Repo) Update(user *AdminUser, updates map[string]any) error {
	return r.db.Model(user).Updates(updates).Error
}

func (r *Repo) ReplaceUserRoles(userID int64, roleIDs []int64) error {
	if err := r.db.Where("user_id = ?", userID).Delete(&AdminUserRole{}).Error; err != nil {
		return err
	}
	for _, roleID := range bootstrap.NormalizeIDs(roleIDs...) {
		if err := r.db.Create(&AdminUserRole{UserID: userID, RoleID: roleID}).Error; err != nil {
			return err
		}
	}
	return nil
}

func (r *Repo) DeleteUsersAndRelations(ids []int64) error {
	return r.WithTransaction(func(tx *Repo) error {
		if err := tx.db.Where("user_id IN ?", ids).Delete(&AdminUserRole{}).Error; err != nil {
			return err
		}
		if err := tx.db.Where("admin_user_id IN ?", ids).Delete(&AdminSession{}).Error; err != nil {
			return err
		}
		return tx.db.Where("id IN ?", ids).Delete(&AdminUser{}).Error
	})
}

func (r *Repo) applyListFilter(query *gorm.DB, filter userListFilter) sharedquery.Result {
	params := sharedquery.Params{
		Search: filter.KeywordPlain,
		Filters: map[string]any{
			"username": filter.UsernamePlain,
			"nickname": filter.NicknamePlain,
		},
		SortBy:    filter.SortBy,
		SortOrder: filter.SortOrder,
	}
	if filter.Status != nil {
		params.Filters["status"] = *filter.Status
	}
	return sharedquery.Apply(query, params, sharedquery.Options{
		SearchFields: []string{"username", "nickname"},
		LikeFields:   []string{"username", "nickname"},
		ExactFields:  []string{"status"},
		AllowedSorts: []string{"id", "username", "nickname", "status", "created_at", "updated_at"},
		DefaultSorts: sharedquery.DefaultSorts("id"),
	})
}
