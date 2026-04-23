package admin_menu

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

func (r *Repo) WithTransaction(fn func(tx *Repo) error) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		return fn(&Repo{db: tx})
	})
}

func (r *Repo) List(filter menuListFilter) ([]AdminMenu, error) {
	var menus []AdminMenu
	err := r.applyListFilter(r.db.Model(&AdminMenu{}), filter).
		Order("sort ASC, id ASC").
		Find(&menus).Error
	return menus, err
}

func (r *Repo) All() ([]AdminMenu, error) {
	var menus []AdminMenu
	err := r.db.Order("sort ASC, id ASC").Find(&menus).Error
	return menus, err
}

func (r *Repo) MenuNodes() ([]AdminMenu, error) {
	var menus []AdminMenu
	err := r.db.Where("menu_type = ?", MenuTypeMenu).Order("sort ASC, id ASC").Find(&menus).Error
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
	if filter.Keyword != "" {
		query = query.Where("title ILIKE ? OR name ILIKE ? OR path ILIKE ?", filter.Keyword, filter.Keyword, filter.Keyword)
	}
	if filter.Title != "" {
		query = query.Where("title ILIKE ? OR name ILIKE ?", filter.Title, filter.Title)
	}
	if filter.Path != "" {
		query = query.Where("path ILIKE ?", filter.Path)
	}
	if filter.MenuType != "" {
		query = query.Where("menu_type = ?", filter.MenuType)
	}
	if filter.Status != nil {
		query = query.Where("status = ?", *filter.Status)
	}
	return query
}
