package admin_menu

import (
	"fmt"
	"strings"
	"time"

	"github.com/lvjiaben/goweb-core/httpx"
	"github.com/lvjiaben/goweb-scaffold/internal/bootstrap"
	"github.com/lvjiaben/goweb-scaffold/internal/modules/model"
	"gorm.io/gorm"
)

type Module struct{}

type saveRequest struct {
	ID             int64  `json:"id"`
	ParentID       int64  `json:"parent_id"`
	Name           string `json:"name" validate:"required"`
	Title          string `json:"title" validate:"required"`
	Path           string `json:"path"`
	Component      string `json:"component"`
	MenuType       string `json:"menu_type" validate:"required"`
	PermissionCode string `json:"permission_code"`
	Icon           string `json:"icon"`
	Sort           int    `json:"sort"`
	Visible        bool   `json:"visible"`
	Status         int    `json:"status"`
}

type menuTreeItem struct {
	ID             int64          `json:"id"`
	ParentID       int64          `json:"parent_id"`
	Name           string         `json:"name"`
	Title          string         `json:"title"`
	Path           string         `json:"path"`
	Component      string         `json:"component"`
	MenuType       string         `json:"menu_type"`
	PermissionCode string         `json:"permission_code"`
	Icon           string         `json:"icon"`
	Sort           int            `json:"sort"`
	Visible        bool           `json:"visible"`
	Status         int            `json:"status"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	Children       []menuTreeItem `json:"children,omitempty"`
}

type menuOption struct {
	Label    string       `json:"label"`
	Value    int64        `json:"value"`
	MenuType string       `json:"menu_type"`
	Children []menuOption `json:"children,omitempty"`
}

func (Module) Name() string { return "admin_menu" }

func (Module) Register(runtime *bootstrap.Runtime) error {
	runtime.AdminProtectedGroup.GET("/admin_menu/list", list(runtime), httpx.WithPermission("admin_menu.list"))
	runtime.AdminProtectedGroup.GET("/admin_menu/detail", detail(runtime), httpx.WithPermission("admin_menu.list"))
	runtime.AdminProtectedGroup.GET("/admin_menu/tree", tree(runtime), httpx.WithPermission("admin_role.save|admin_menu.save"))
	runtime.AdminProtectedGroup.GET("/admin_menu/options", options(runtime), httpx.WithPermission("admin_menu.save"))
	runtime.AdminProtectedGroup.POST("/admin_menu/save", save(runtime), httpx.WithPermission("admin_menu.save"))
	runtime.AdminProtectedGroup.POST("/admin_menu/delete", deleteMenus(runtime), httpx.WithPermission("admin_menu.delete"))
	return nil
}

func list(runtime *bootstrap.Runtime) httpx.HandlerFunc {
	return func(c *httpx.Context) {
		var menus []model.AdminMenu
		if err := runtime.DB.Order("sort ASC, id ASC").Find(&menus).Error; err != nil {
			c.Error(err)
			return
		}
		c.Success(map[string]any{"list": buildTree(menus)})
	}
}

func detail(runtime *bootstrap.Runtime) httpx.HandlerFunc {
	return func(c *httpx.Context) {
		id, err := c.QueryInt64("id")
		if err != nil || id <= 0 {
			c.BadRequest("invalid id")
			return
		}

		var menu model.AdminMenu
		if err := runtime.DB.First(&menu, id).Error; err != nil {
			c.Error(err)
			return
		}

		c.Success(map[string]any{
			"id":              menu.ID,
			"parent_id":       menu.ParentID,
			"name":            menu.Name,
			"title":           menu.Title,
			"path":            menu.Path,
			"component":       menu.Component,
			"menu_type":       menu.MenuType,
			"permission_code": menu.PermissionCode,
			"icon":            menu.Icon,
			"sort":            menu.Sort,
			"visible":         menu.Visible,
			"status":          menu.Status,
		})
	}
}

func tree(runtime *bootstrap.Runtime) httpx.HandlerFunc {
	return func(c *httpx.Context) {
		var menus []model.AdminMenu
		if err := runtime.DB.Order("sort ASC, id ASC").Find(&menus).Error; err != nil {
			c.Error(err)
			return
		}
		c.Success(map[string]any{"list": buildTree(menus)})
	}
}

func options(runtime *bootstrap.Runtime) httpx.HandlerFunc {
	return func(c *httpx.Context) {
		var menus []model.AdminMenu
		if err := runtime.DB.Where("menu_type = ?", model.MenuTypeMenu).Order("sort ASC, id ASC").Find(&menus).Error; err != nil {
			c.Error(err)
			return
		}
		c.Success(map[string]any{"list": buildOptions(buildTree(menus))})
	}
}

func save(runtime *bootstrap.Runtime) httpx.HandlerFunc {
	return func(c *httpx.Context) {
		var req saveRequest
		if err := c.BindJSON(&req); err != nil {
			c.Error(err)
			return
		}
		req.Name = strings.TrimSpace(req.Name)
		req.Title = strings.TrimSpace(req.Title)
		req.Path = strings.TrimSpace(req.Path)
		req.Component = strings.TrimSpace(req.Component)
		req.PermissionCode = strings.TrimSpace(req.PermissionCode)
		req.Icon = strings.TrimSpace(req.Icon)
		if req.Status == 0 {
			req.Status = 1
		}
		if err := runtime.Validator.Struct(req); err != nil {
			c.BadRequest(err.Error())
			return
		}
		if req.MenuType != model.MenuTypeMenu && req.MenuType != model.MenuTypeButton {
			c.BadRequest("invalid menu_type")
			return
		}
		if err := validateMenuSave(runtime, req); err != nil {
			c.BadRequest(err.Error())
			return
		}

		if req.ID == 0 {
			menu := model.AdminMenu{
				ParentID:       req.ParentID,
				Name:           req.Name,
				Title:          req.Title,
				Path:           req.Path,
				Component:      req.Component,
				MenuType:       req.MenuType,
				PermissionCode: req.PermissionCode,
				Icon:           req.Icon,
				Sort:           req.Sort,
				Visible:        req.Visible,
				Status:         req.Status,
			}
			if err := runtime.DB.Create(&menu).Error; err != nil {
				c.Error(err)
				return
			}
			c.Success(map[string]any{"id": menu.ID})
			return
		}

		var menu model.AdminMenu
		if err := runtime.DB.First(&menu, req.ID).Error; err != nil {
			c.Error(err)
			return
		}

		if err := runtime.DB.Model(&menu).Updates(map[string]any{
			"parent_id":       req.ParentID,
			"name":            req.Name,
			"title":           req.Title,
			"path":            req.Path,
			"component":       req.Component,
			"menu_type":       req.MenuType,
			"permission_code": req.PermissionCode,
			"icon":            req.Icon,
			"sort":            req.Sort,
			"visible":         req.Visible,
			"status":          req.Status,
			"updated_at":      time.Now(),
		}).Error; err != nil {
			c.Error(err)
			return
		}
		c.Success(map[string]any{"id": menu.ID})
	}
}

func deleteMenus(runtime *bootstrap.Runtime) httpx.HandlerFunc {
	return func(c *httpx.Context) {
		var req bootstrap.IDsPayload
		if err := c.BindJSON(&req); err != nil {
			c.Error(err)
			return
		}
		ids := req.Values()
		if len(ids) == 0 {
			c.BadRequest("ids is required")
			return
		}

		var menus []model.AdminMenu
		if err := runtime.DB.Select("id", "parent_id").Find(&menus).Error; err != nil {
			c.Error(err)
			return
		}
		deleteIDs := collectDescendantIDs(menus, ids)

		err := runtime.DB.Transaction(func(tx *gorm.DB) error {
			if err := tx.Where("menu_id IN ?", deleteIDs).Delete(&model.AdminRoleMenu{}).Error; err != nil {
				return err
			}
			return tx.Where("id IN ?", deleteIDs).Delete(&model.AdminMenu{}).Error
		})
		if err != nil {
			c.Error(err)
			return
		}
		c.Success(map[string]any{"deleted": len(deleteIDs)})
	}
}

func buildTree(menus []model.AdminMenu) []menuTreeItem {
	items := make(map[int64]menuTreeItem, len(menus))
	children := make(map[int64][]int64)
	rootIDs := make([]int64, 0)

	for _, menu := range menus {
		items[menu.ID] = menuTreeItem{
			ID:             menu.ID,
			ParentID:       menu.ParentID,
			Name:           menu.Name,
			Title:          menu.Title,
			Path:           menu.Path,
			Component:      menu.Component,
			MenuType:       menu.MenuType,
			PermissionCode: menu.PermissionCode,
			Icon:           menu.Icon,
			Sort:           menu.Sort,
			Visible:        menu.Visible,
			Status:         menu.Status,
			CreatedAt:      menu.CreatedAt,
			UpdatedAt:      menu.UpdatedAt,
			Children:       []menuTreeItem{},
		}
	}

	for _, menu := range menus {
		if menu.ParentID <= 0 {
			rootIDs = append(rootIDs, menu.ID)
			continue
		}
		if _, ok := items[menu.ParentID]; !ok {
			rootIDs = append(rootIDs, menu.ID)
			continue
		}
		children[menu.ParentID] = append(children[menu.ParentID], menu.ID)
	}

	result := make([]menuTreeItem, 0, len(rootIDs))
	for _, id := range rootIDs {
		result = append(result, buildTreeNode(id, items, children))
	}
	return result
}

func buildTreeNode(id int64, items map[int64]menuTreeItem, children map[int64][]int64) menuTreeItem {
	node := items[id]
	node.Children = []menuTreeItem{}
	for _, childID := range children[id] {
		node.Children = append(node.Children, buildTreeNode(childID, items, children))
	}
	return node
}

func collectDescendantIDs(menus []model.AdminMenu, initial []int64) []int64 {
	children := make(map[int64][]int64)
	for _, menu := range menus {
		children[menu.ParentID] = append(children[menu.ParentID], menu.ID)
	}

	queue := append([]int64{}, initial...)
	seen := make(map[int64]struct{}, len(queue))
	result := make([]int64, 0, len(queue))

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		if _, ok := seen[current]; ok {
			continue
		}
		seen[current] = struct{}{}
		result = append(result, current)
		queue = append(queue, children[current]...)
	}
	return result
}

func buildOptions(items []menuTreeItem) []menuOption {
	result := make([]menuOption, 0, len(items))
	for _, item := range items {
		option := menuOption{
			Label:    item.Title,
			Value:    item.ID,
			MenuType: item.MenuType,
			Children: buildOptions(item.Children),
		}
		result = append(result, option)
	}
	return result
}

func validateMenuSave(runtime *bootstrap.Runtime, req saveRequest) error {
	if req.ID > 0 && req.ParentID == req.ID {
		return fmt.Errorf("parent_id cannot be self")
	}

	if req.ParentID > 0 {
		var parent model.AdminMenu
		if err := runtime.DB.Select("id", "menu_type").First(&parent, req.ParentID).Error; err != nil {
			return err
		}
		if parent.MenuType != model.MenuTypeMenu {
			return fmt.Errorf("parent menu must be a menu node")
		}
	}

	if req.MenuType == model.MenuTypeButton && req.PermissionCode == "" {
		return fmt.Errorf("button menu requires permission_code")
	}

	if req.MenuType == model.MenuTypeMenu {
		if req.Name == "" || req.Title == "" || req.Path == "" {
			return fmt.Errorf("menu requires name, title and path")
		}
		if !strings.HasPrefix(req.Path, "/") {
			return fmt.Errorf("menu path must start with /")
		}
	}

	if req.ID > 0 && req.ParentID > 0 {
		var menus []model.AdminMenu
		if err := runtime.DB.Select("id", "parent_id").Find(&menus).Error; err != nil {
			return err
		}
		parentMap := make(map[int64]int64, len(menus))
		for _, item := range menus {
			parentMap[item.ID] = item.ParentID
		}

		cursor := req.ParentID
		visited := map[int64]struct{}{}
		for cursor > 0 {
			if cursor == req.ID {
				return fmt.Errorf("parent relation creates cycle")
			}
			if _, ok := visited[cursor]; ok {
				break
			}
			visited[cursor] = struct{}{}
			cursor = parentMap[cursor]
		}
	}

	return nil
}
