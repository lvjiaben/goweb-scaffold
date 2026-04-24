package admin_role

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

func (r *Repo) ActiveRoles() ([]AdminRole, error) {
	var roles []AdminRole
	err := r.db.Where("status = ?", 1).Order("id DESC").Find(&roles).Error
	return roles, err
}

func (r *Repo) Count(filter roleListFilter) (int64, error) {
	var total int64
	err := r.applyListFilter(r.db.Model(&AdminRole{}), filter).Count.Count(&total).Error
	return total, err
}

func (r *Repo) List(filter roleListFilter, page int, pageSize int) ([]AdminRole, error) {
	var roles []AdminRole
	result := r.applyListFilter(r.db.Model(&AdminRole{}), filter)
	err := result.Query.Offset((page - 1) * pageSize).Limit(pageSize).Find(&roles).Error
	return roles, err
}

func (r *Repo) FindByID(id int64) (AdminRole, error) {
	var role AdminRole
	err := r.db.First(&role, id).Error
	return role, err
}

func (r *Repo) MenuIDsByRoleID(roleID int64) ([]int64, error) {
	var menuIDs []int64
	err := r.db.Model(&AdminRoleMenu{}).Where("role_id = ?", roleID).Pluck("menu_id", &menuIDs).Error
	return menuIDs, err
}

func (r *Repo) CountByCode(code string, excludeID int64) (int64, error) {
	var count int64
	query := r.db.Model(&AdminRole{}).Where("code = ?", code)
	if excludeID > 0 {
		query = query.Where("id <> ?", excludeID)
	}
	err := query.Count(&count).Error
	return count, err
}

func (r *Repo) CountUserBindings(roleIDs []int64) (int64, error) {
	var count int64
	err := r.db.Model(&AdminUserRole{}).Where("role_id IN ?", roleIDs).Count(&count).Error
	return count, err
}

func (r *Repo) Create(role *AdminRole) error {
	return r.db.Create(role).Error
}

func (r *Repo) Update(role *AdminRole, updates map[string]any) error {
	return r.db.Model(role).Updates(updates).Error
}

func (r *Repo) ReplaceRoleMenus(roleID int64, menuIDs []int64) error {
	if err := r.db.Where("role_id = ?", roleID).Delete(&AdminRoleMenu{}).Error; err != nil {
		return err
	}
	for _, menuID := range bootstrap.NormalizeIDs(menuIDs...) {
		if err := r.db.Create(&AdminRoleMenu{RoleID: roleID, MenuID: menuID}).Error; err != nil {
			return err
		}
	}
	return nil
}

func (r *Repo) DeleteRolesAndRelations(ids []int64) error {
	return r.WithTransaction(func(tx *Repo) error {
		if err := tx.db.Where("role_id IN ?", ids).Delete(&AdminRoleMenu{}).Error; err != nil {
			return err
		}
		if err := tx.db.Where("role_id IN ?", ids).Delete(&AdminUserRole{}).Error; err != nil {
			return err
		}
		return tx.db.Where("id IN ?", ids).Delete(&AdminRole{}).Error
	})
}

func (r *Repo) applyListFilter(query *gorm.DB, filter roleListFilter) sharedquery.Result {
	params := sharedquery.Params{
		Search: filter.KeywordPlain,
		Filters: map[string]any{
			"name": filter.NamePlain,
			"code": filter.CodePlain,
		},
		SortBy:    filter.SortBy,
		SortOrder: filter.SortOrder,
	}
	if filter.Status != nil {
		params.Filters["status"] = *filter.Status
	}
	return sharedquery.Apply(query, params, sharedquery.Options{
		SearchFields: []string{"name", "code"},
		LikeFields:   []string{"name", "code"},
		ExactFields:  []string{"status"},
		AllowedSorts: []string{"id", "name", "code", "status", "created_at", "updated_at"},
		DefaultSorts: sharedquery.DefaultSorts("id"),
	})
}
