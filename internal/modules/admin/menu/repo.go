package admin_menu

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

func (r *Repo) List(filter menuListFilter) ([]AdminMenu, error) {
	var menus []AdminMenu
	err := r.applyListFilter(r.db.Model(&AdminMenu{}), filter).
		Order("sort DESC, id DESC").
		Find(&menus).Error
	return menus, err
}

func (r *Repo) All() ([]AdminMenu, error) {
	var menus []AdminMenu
	err := r.db.Order("sort DESC, id DESC").Find(&menus).Error
	return menus, err
}

func (r *Repo) MenuNodes() ([]AdminMenu, error) {
	var menus []AdminMenu
	err := r.db.Where("menu_type = ?", MenuTypeMenu).Order("sort DESC, id DESC").Find(&menus).Error
	return menus, err
}

func (r *Repo) VbenRouteMenus(roleIDs []int64, isSuper bool) ([]AdminMenu, error) {
	var menus []AdminMenu
	query := r.db.Model(&AdminMenu{}).
		Where("menu_type <> ? AND status = ?", MenuTypeButton, 1).
		Order("sort DESC, id DESC")
	if !isSuper {
		if len(roleIDs) == 0 {
			return menus, nil
		}
		query = query.Joins("JOIN admin_role_menu arm ON arm.menu_id = admin_menu.id AND arm.deleted_at IS NULL").
			Where("arm.role_id IN ?", roleIDs).
			Distinct("admin_menu.*")
	}
	err := query.Find(&menus).Error
	return menus, err
}

func (r *Repo) FindByID(id int64) (AdminMenu, error) {
	var menu AdminMenu
	err := r.db.First(&menu, id).Error
	return menu, err
}

func (r *Repo) FindParent(id int64) (AdminMenu, error) {
	var parent AdminMenu
	err := r.db.Select("id", "menu_type").First(&parent, id).Error
	return parent, err
}

func (r *Repo) IDParentPairs() ([]AdminMenu, error) {
	var menus []AdminMenu
	err := r.db.Select("id", "parent_id").Find(&menus).Error
	return menus, err
}

func (r *Repo) Create(menu *AdminMenu) error {
	return r.db.Create(menu).Error
}

func (r *Repo) Update(menu *AdminMenu, updates map[string]any) error {
	return r.db.Model(menu).Updates(updates).Error
}

func (r *Repo) DeleteMenusAndRoleLinks(ids []int64) error {
	return r.WithTransaction(func(tx *Repo) error {
		if err := tx.db.Where("menu_id IN ?", ids).Delete(&AdminRoleMenu{}).Error; err != nil {
			return err
		}
		return tx.db.Where("id IN ?", ids).Delete(&AdminMenu{}).Error
	})
}

func (r *Repo) applyListFilter(query *gorm.DB, filter menuListFilter) *gorm.DB {
	params := sharedquery.Params{
		Search: filter.KeywordPlain,
		Filters: map[string]any{
			"title": filter.TitlePlain,
			"path":  filter.PathPlain,
		},
		SortBy:    filter.SortBy,
		SortOrder: filter.SortOrder,
	}
	result := sharedquery.Apply(query, params, sharedquery.Options{
		SearchFields: []string{"title", "name", "path"},
		LikeFields:   []string{"title", "path"},
		DefaultSorts: sharedquery.DefaultSorts("sort", "id"),
	})
	query = result.Query
	if filter.MenuType != "" {
		query = query.Where("menu_type = ?", filter.MenuType)
	}
	if filter.Status != nil {
		query = query.Where("status = ?", *filter.Status)
	}
	return query
}
