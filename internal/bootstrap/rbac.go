package bootstrap

import (
	"context"

	corerbac "github.com/lvjiaben/goweb-core/rbac"
	"github.com/lvjiaben/goweb-scaffold/internal/modules/model"
	"gorm.io/gorm"
)

type PermissionService struct {
	db *gorm.DB
}

func NewPermissionService(db *gorm.DB) *PermissionService {
	return &PermissionService{db: db}
}

func (s *PermissionService) HasPermission(ctx context.Context, identity *corerbac.Identity, permissionCode string) (bool, error) {
	if identity == nil {
		return false, nil
	}
	if identity.IsSuper {
		return true, nil
	}
	if permissionCode == "" || len(identity.RoleIDs) == 0 {
		return false, nil
	}

	var count int64
	err := s.db.WithContext(ctx).
		Model(&model.AdminMenu{}).
		Joins("JOIN admin_role_menu arm ON arm.menu_id = admin_menu.id AND arm.deleted_at IS NULL").
		Where("arm.role_id IN ?", identity.RoleIDs).
		Where("admin_menu.permission_code = ? AND admin_menu.status = ?", permissionCode, 1).
		Count(&count).Error
	return count > 0, err
}

func (s *PermissionService) GetAccessCodes(ctx context.Context, identity *corerbac.Identity) ([]string, error) {
	if identity == nil {
		return []string{}, nil
	}

	query := s.db.WithContext(ctx).Model(&model.AdminMenu{}).Where("permission_code <> '' AND status = ?", 1)
	if !identity.IsSuper {
		if len(identity.RoleIDs) == 0 {
			return []string{}, nil
		}
		query = query.Joins("JOIN admin_role_menu arm ON arm.menu_id = admin_menu.id AND arm.deleted_at IS NULL").
			Where("arm.role_id IN ?", identity.RoleIDs)
	}

	var codes []string
	if err := query.Distinct("permission_code").Order("permission_code ASC").Pluck("permission_code", &codes).Error; err != nil {
		return nil, err
	}
	return codes, nil
}

func (s *PermissionService) GetMenus(ctx context.Context, identity *corerbac.Identity) ([]corerbac.MenuItem, error) {
	if identity == nil {
		return []corerbac.MenuItem{}, nil
	}

	query := s.db.WithContext(ctx).
		Model(&model.AdminMenu{}).
		Where("menu_type = ? AND status = ?", model.MenuTypeMenu, 1).
		Order("sort ASC, id ASC")

	if !identity.IsSuper {
		if len(identity.RoleIDs) == 0 {
			return []corerbac.MenuItem{}, nil
		}
		query = query.Joins("JOIN admin_role_menu arm ON arm.menu_id = admin_menu.id AND arm.deleted_at IS NULL").
			Where("arm.role_id IN ?", identity.RoleIDs).
			Distinct("admin_menu.id", "admin_menu.parent_id", "admin_menu.name", "admin_menu.title", "admin_menu.path", "admin_menu.component", "admin_menu.menu_type", "admin_menu.icon", "admin_menu.sort", "admin_menu.permission_code")
	}

	var rows []model.AdminMenu
	if err := query.Find(&rows).Error; err != nil {
		return nil, err
	}

	items := make([]corerbac.MenuItem, 0, len(rows))
	for _, item := range rows {
		items = append(items, corerbac.MenuItem{
			ID:             item.ID,
			ParentID:       item.ParentID,
			Name:           item.Name,
			Title:          item.Title,
			Path:           item.Path,
			Component:      item.Component,
			MenuType:       item.MenuType,
			Icon:           item.Icon,
			Sort:           item.Sort,
			PermissionCode: item.PermissionCode,
		})
	}
	return buildMenuTree(items), nil
}

func buildMenuTree(items []corerbac.MenuItem) []corerbac.MenuItem {
	index := make(map[int64]corerbac.MenuItem, len(items))
	children := make(map[int64][]int64)
	rootIDs := make([]int64, 0)

	for _, item := range items {
		item.Children = []corerbac.MenuItem{}
		index[item.ID] = item
	}

	for _, item := range items {
		if item.ParentID <= 0 {
			rootIDs = append(rootIDs, item.ID)
			continue
		}
		if _, ok := index[item.ParentID]; !ok {
			rootIDs = append(rootIDs, item.ID)
			continue
		}
		children[item.ParentID] = append(children[item.ParentID], item.ID)
	}

	result := make([]corerbac.MenuItem, 0, len(rootIDs))
	for _, id := range rootIDs {
		result = append(result, buildRBACNode(id, index, children))
	}
	return result
}

func buildRBACNode(id int64, index map[int64]corerbac.MenuItem, children map[int64][]int64) corerbac.MenuItem {
	node := index[id]
	node.Children = []corerbac.MenuItem{}
	for _, childID := range children[id] {
		node.Children = append(node.Children, buildRBACNode(childID, index, children))
	}
	return node
}
