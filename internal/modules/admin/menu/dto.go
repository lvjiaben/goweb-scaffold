package admin_menu

import "time"

type ListParams struct {
	Page     int
	PageSize int
	Keyword  string
	Filters  map[string]any
}

type SaveRequest struct {
	ID             int64  `json:"id"`
	ParentID       int64  `json:"parent_id"`
	PID            int64  `json:"pid"`
	Name           string `json:"name"`
	Title          string `json:"title"`
	Path           string `json:"path"`
	Component      string `json:"component"`
	MenuType       string `json:"menu_type"`
	Type           string `json:"type"`
	PermissionCode string `json:"permission_code"`
	Permission     string `json:"permission"`
	Icon           string `json:"icon"`
	Sort           int    `json:"sort"`
	Visible        bool   `json:"visible"`
	Status         int    `json:"status"`
}

type DetailResponse struct {
	ID             int64  `json:"id"`
	ParentID       int64  `json:"parent_id"`
	PID            int64  `json:"pid"`
	Name           string `json:"name"`
	Title          string `json:"title"`
	Path           string `json:"path"`
	Component      string `json:"component"`
	MenuType       string `json:"menu_type"`
	Type           string `json:"type"`
	PermissionCode string `json:"permission_code"`
	Permission     string `json:"permission"`
	Icon           string `json:"icon"`
	Sort           int    `json:"sort"`
	Visible        bool   `json:"visible"`
	Status         int    `json:"status"`
}

type MenuTreeItem struct {
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
	Children       []MenuTreeItem `json:"children,omitempty"`
}

type MenuOption struct {
	Label    string       `json:"label"`
	Value    int64        `json:"value"`
	MenuType string       `json:"menu_type"`
	Children []MenuOption `json:"children,omitempty"`
}

type SaveResult struct {
	ID int64 `json:"id"`
}

type DeleteResult struct {
	Deleted int `json:"deleted"`
}

type menuListFilter struct {
	Keyword  string
	Title    string
	Path     string
	MenuType string
	Status   *int64
}
