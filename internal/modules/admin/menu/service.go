package admin_menu

import (
	"fmt"
	"strings"
	"time"

	"github.com/lvjiaben/goweb-scaffold/internal/bootstrap"
)

type Service struct {
	repo *Repo
}

func NewService(runtime *bootstrap.Runtime) *Service {
	return &Service{repo: NewRepo(runtime)}
}

func (s *Service) List(params ListParams) (map[string]any, error) {
	filter := menuListFilter{
		KeywordPlain: strings.TrimSpace(params.Keyword),
		TitlePlain:   strings.TrimSpace(bootstrap.FilterString(params.Filters, "title", "name")),
		PathPlain:    strings.TrimSpace(bootstrap.FilterString(params.Filters, "path")),
		MenuType:     strings.TrimSpace(bootstrap.FilterString(params.Filters, "menu_type", "type")),
		SortBy:       strings.TrimSpace(params.SortBy),
		SortOrder:    strings.TrimSpace(params.SortOrder),
	}
	if status, ok := bootstrap.FilterInt64(params.Filters, "status"); ok {
		filter.Status = &status
	}
	menus, err := s.repo.List(filter)
	if err != nil {
		return nil, err
	}
	treeItems := buildTree(menus)
	return bootstrap.PagedResult(treeItems, int64(len(treeItems)), params.Page, params.PageSize), nil
}

func (s *Service) Detail(id int64) (DetailResponse, error) {
	menu, err := s.repo.FindByID(id)
	if err != nil {
		return DetailResponse{}, err
	}
	return DetailResponse{
		ID:             menu.ID,
		ParentID:       menu.ParentID,
		PID:            menu.ParentID,
		Name:           menu.Name,
		EnName:         menu.EnName,
		Title:          menu.Title,
		Path:           menu.Path,
		Component:      menu.Component,
		MenuType:       menu.MenuType,
		Type:           menu.MenuType,
		PermissionCode: menu.PermissionCode,
		Permission:     menu.PermissionCode,
		Iframe:         menu.Iframe,
		External:       menu.External,
		Icon:           bootstrap.NormalizeMenuIcon(menu.Icon),
		Sort:           menu.Sort,
		Visible:        menu.Visible,
		Status:         menu.Status,
	}, nil
}

func (s *Service) Tree(page int, pageSize int) (map[string]any, error) {
	menus, err := s.repo.All()
	if err != nil {
		return nil, err
	}
	treeItems := buildTree(menus)
	return bootstrap.PagedResult(treeItems, int64(len(treeItems)), page, pageSize), nil
}

func (s *Service) Options() (map[string]any, error) {
	menus, err := s.repo.MenuNodes()
	if err != nil {
		return nil, err
	}
	return map[string]any{"list": buildOptions(buildTree(menus))}, nil
}

func (s *Service) SaveMenu(req SaveRequest) (SaveResult, error) {
	req = normalizeSaveRequest(req)
	if err := validateRequired(req); err != nil {
		return SaveResult{}, err
	}
	if err := s.validateMenuSave(req); err != nil {
		return SaveResult{}, validationError(err.Error())
	}

	if req.ID == 0 {
		menu := AdminMenu{
			ParentID:       req.ParentID,
			Name:           req.Name,
			EnName:         req.EnName,
			Title:          req.Title,
			Path:           req.Path,
			Component:      req.Component,
			MenuType:       req.MenuType,
			PermissionCode: req.PermissionCode,
			Iframe:         req.Iframe,
			External:       req.External,
			Icon:           req.Icon,
			Sort:           req.Sort,
			Visible:        req.Visible,
			Status:         req.Status,
		}
		if err := s.repo.Create(&menu); err != nil {
			return SaveResult{}, err
		}
		return SaveResult{ID: menu.ID}, nil
	}

	menu, err := s.repo.FindByID(req.ID)
	if err != nil {
		return SaveResult{}, err
	}
	if err := s.repo.Update(&menu, map[string]any{
		"parent_id":       req.ParentID,
		"name":            req.Name,
		"enname":          req.EnName,
		"title":           req.Title,
		"path":            req.Path,
		"component":       req.Component,
		"menu_type":       req.MenuType,
		"permission_code": req.PermissionCode,
		"iframe":          req.Iframe,
		"external":        req.External,
		"icon":            req.Icon,
		"sort":            req.Sort,
		"visible":         req.Visible,
		"status":          req.Status,
		"updated_at":      time.Now(),
	}); err != nil {
		return SaveResult{}, err
	}
	return SaveResult{ID: menu.ID}, nil
}

func (s *Service) DeleteMenus(ids []int64) (DeleteResult, error) {
	ids = bootstrap.NormalizeIDs(ids...)
	if len(ids) == 0 {
		return DeleteResult{}, validationError("ids is required")
	}
	menus, err := s.repo.IDParentPairs()
	if err != nil {
		return DeleteResult{}, err
	}
	deleteIDs := collectDescendantIDs(menus, ids)
	if err := s.repo.DeleteMenusAndRoleLinks(deleteIDs); err != nil {
		return DeleteResult{}, err
	}
	return DeleteResult{Deleted: len(deleteIDs)}, nil
}

func normalizeSaveRequest(req SaveRequest) SaveRequest {
	req.Name = strings.TrimSpace(req.Name)
	req.EnName = strings.TrimSpace(req.EnName)
	req.Title = strings.TrimSpace(req.Title)
	req.Path = strings.TrimSpace(req.Path)
	req.Component = strings.TrimSpace(req.Component)
	req.Iframe = strings.TrimSpace(req.Iframe)
	req.External = strings.TrimSpace(req.External)
	if req.ParentID == 0 && req.PID > 0 {
		req.ParentID = req.PID
	}
	if req.MenuType == "" {
		req.MenuType = strings.TrimSpace(req.Type)
	}
	if req.PermissionCode == "" {
		req.PermissionCode = strings.TrimSpace(req.Permission)
	}
	req.PermissionCode = strings.TrimSpace(req.PermissionCode)
	req.Icon = strings.TrimSpace(req.Icon)
	return req
}

func validateRequired(req SaveRequest) error {
	if req.Name == "" {
		return validationError("name is required")
	}
	if req.Title == "" {
		return validationError("title is required")
	}
	return nil
}

func (s *Service) validateMenuSave(req SaveRequest) error {
	if req.ID > 0 && req.ParentID == req.ID {
		return fmt.Errorf("parent_id cannot be self")
	}

	if req.ParentID > 0 {
		parent, err := s.repo.FindParent(req.ParentID)
		if err != nil {
			return err
		}
		if parent.MenuType != MenuTypeMenu {
			return fmt.Errorf("parent menu must be a menu node")
		}
	}

	if req.MenuType != MenuTypeMenu && req.MenuType != MenuTypeButton && req.MenuType != MenuTypeIframe && req.MenuType != MenuTypeLink {
		return fmt.Errorf("invalid menu_type")
	}
	if req.MenuType == MenuTypeButton && req.PermissionCode == "" {
		return fmt.Errorf("button menu requires permission_code")
	}
	if req.MenuType == MenuTypeIframe && req.Iframe == "" {
		return fmt.Errorf("iframe menu requires iframe url")
	}
	if req.MenuType == MenuTypeLink && req.External == "" {
		return fmt.Errorf("link menu requires external url")
	}
	if req.MenuType == MenuTypeMenu {
		if req.Name == "" || req.Title == "" || req.Path == "" {
			return fmt.Errorf("menu requires name, title and path")
		}
		if !strings.HasPrefix(req.Path, "/") {
			return fmt.Errorf("menu path must start with /")
		}
	}
	if req.ID > 0 && req.ParentID > 0 {
		menus, err := s.repo.IDParentPairs()
		if err != nil {
			return err
		}
		if createsCycle(menus, req.ID, req.ParentID) {
			return fmt.Errorf("parent relation creates cycle")
		}
	}
	return nil
}

func buildTree(menus []AdminMenu) []MenuTreeItem {
	items := make(map[int64]MenuTreeItem, len(menus))
	children := make(map[int64][]int64)
	rootIDs := make([]int64, 0)

	for _, menu := range menus {
		items[menu.ID] = MenuTreeItem{
			ID:             menu.ID,
			ParentID:       menu.ParentID,
			Name:           menu.Name,
			EnName:         menu.EnName,
			Title:          menu.Title,
			Path:           menu.Path,
			Component:      menu.Component,
			MenuType:       menu.MenuType,
			PermissionCode: menu.PermissionCode,
			Iframe:         menu.Iframe,
			External:       menu.External,
			Icon:           bootstrap.NormalizeMenuIcon(menu.Icon),
			Sort:           menu.Sort,
			Visible:        menu.Visible,
			Status:         menu.Status,
			CreatedAt:      menu.CreatedAt,
			UpdatedAt:      menu.UpdatedAt,
			Children:       []MenuTreeItem{},
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

	result := make([]MenuTreeItem, 0, len(rootIDs))
	for _, id := range rootIDs {
		result = append(result, buildTreeNode(id, items, children))
	}
	return result
}

func buildTreeNode(id int64, items map[int64]MenuTreeItem, children map[int64][]int64) MenuTreeItem {
	node := items[id]
	node.Children = []MenuTreeItem{}
	for _, childID := range children[id] {
		node.Children = append(node.Children, buildTreeNode(childID, items, children))
	}
	return node
}

func collectDescendantIDs(menus []AdminMenu, initial []int64) []int64 {
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

func buildOptions(items []MenuTreeItem) []MenuOption {
	result := make([]MenuOption, 0, len(items))
	for _, item := range items {
		option := MenuOption{
			Label:    item.Title,
			Value:    item.ID,
			MenuType: item.MenuType,
			Children: buildOptions(item.Children),
		}
		result = append(result, option)
	}
	return result
}

func createsCycle(menus []AdminMenu, id int64, parentID int64) bool {
	parentMap := make(map[int64]int64, len(menus))
	for _, item := range menus {
		parentMap[item.ID] = item.ParentID
	}

	cursor := parentID
	visited := map[int64]struct{}{}
	for cursor > 0 {
		if cursor == id {
			return true
		}
		if _, ok := visited[cursor]; ok {
			break
		}
		visited[cursor] = struct{}{}
		cursor = parentMap[cursor]
	}
	return false
}

type validationError string

func (e validationError) Error() string {
	return string(e)
}

func isValidationError(err error) bool {
	_, ok := err.(validationError)
	return ok
}
