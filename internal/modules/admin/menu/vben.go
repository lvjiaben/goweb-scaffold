package admin_menu

import (
	"context"
	"strings"

	corerbac "github.com/lvjiaben/goweb-core/rbac"
	"github.com/lvjiaben/goweb-scaffold/internal/bootstrap"
)

type VbenRoute struct {
	Name      string      `json:"name"`
	Path      string      `json:"path"`
	Component string      `json:"component,omitempty"`
	Redirect  string      `json:"redirect,omitempty"`
	Meta      VbenMeta    `json:"meta"`
	Children  []VbenRoute `json:"children,omitempty"`
}

type VbenMeta struct {
	Title      string   `json:"title"`
	Icon       string   `json:"icon,omitempty"`
	Order      int      `json:"order,omitempty"`
	AffixTab   bool     `json:"affixTab,omitempty"`
	HideInTab  bool     `json:"hideInTab,omitempty"`
	HideInMenu bool     `json:"hideInMenu,omitempty"`
	KeepAlive  bool     `json:"keepAlive,omitempty"`
	IframeSrc  string   `json:"iframeSrc,omitempty"`
	Link       string   `json:"link,omitempty"`
	Authority  []string `json:"authority,omitempty"`
}

func (s *Service) GetVbenRoutes(_ context.Context, identity *corerbac.Identity, acceptLanguage string) ([]VbenRoute, error) {
	if identity == nil {
		return []VbenRoute{}, nil
	}
	menus, err := s.repo.VbenRouteMenus(identity.RoleIDs, identity.IsSuper)
	if err != nil {
		return nil, err
	}
	menuMap := make(map[int64][]AdminMenu, len(menus))
	for _, menu := range menus {
		menuMap[menu.ParentID] = append(menuMap[menu.ParentID], menu)
	}

	useEnglish := isEnglishRequest(acceptLanguage)
	routes := make([]VbenRoute, 0, len(menuMap[0]))
	for _, root := range menuMap[0] {
		routes = append(routes, convertMenuToVbenRoute(root, menuMap, useEnglish))
	}
	return routes, nil
}

func isEnglishRequest(acceptLanguage string) bool {
	return strings.Contains(strings.ToLower(acceptLanguage), "en")
}

func convertMenuToVbenRoute(menu AdminMenu, menuMap map[int64][]AdminMenu, useEnglish bool) VbenRoute {
	route := VbenRoute{
		Name: resolveRouteName(menu),
		Path: menu.Path,
		Meta: VbenMeta{
			Title:     resolveMenuTitle(menu, useEnglish),
			Icon:      bootstrap.NormalizeMenuIcon(menu.Icon),
			KeepAlive: true,
		},
	}

	if !menu.Visible {
		route.Meta.HideInMenu = true
	}
	if menu.ShowTag == 1 {
		route.Meta.HideInTab = true
	}
	if menu.FixedTag == 1 {
		route.Meta.AffixTab = true
	}
	if strings.TrimSpace(menu.PermissionCode) != "" {
		route.Meta.Authority = []string{strings.TrimSpace(menu.PermissionCode)}
	}

	switch menu.MenuType {
	case MenuTypeMenu:
		if menu.ParentID == 0 {
			route.Component = "BasicLayout"
		} else {
			route.Component = menu.Component
		}
	case MenuTypeIframe:
		route.Component = "IFrameView"
		route.Meta.IframeSrc = menu.Iframe
	case MenuTypeLink:
		route.Meta.Link = menu.External
		if strings.TrimSpace(menu.Component) != "" {
			route.Component = menu.Component
		}
	}

	children := menuMap[menu.ID]
	if len(children) > 0 {
		route.Children = make([]VbenRoute, 0, len(children))
		for _, child := range children {
			route.Children = append(route.Children, convertMenuToVbenRoute(child, menuMap, useEnglish))
		}
		if route.Redirect == "" && len(route.Children) > 0 {
			route.Redirect = route.Children[0].Path
		}
	}
	return route
}

func resolveRouteName(menu AdminMenu) string {
	if value := strings.TrimSpace(menu.EnName); value != "" {
		return value
	}
	return strings.TrimSpace(menu.Name)
}

func resolveMenuTitle(menu AdminMenu, useEnglish bool) string {
	if useEnglish {
		if value := strings.TrimSpace(menu.EnName); value != "" {
			return value
		}
	}
	if value := strings.TrimSpace(menu.Title); value != "" {
		return value
	}
	return strings.TrimSpace(menu.Name)
}
