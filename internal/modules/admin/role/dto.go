package admin_role

import "time"

type ListParams struct {
	Page      int
	PageSize  int
	Keyword   string
	Filters   map[string]any
	SortBy    string
	SortOrder string
}

type SaveRequest struct {
	ID          int64   `json:"id"`
	Name        string  `json:"name"`
	Code        string  `json:"code"`
	Description string  `json:"description"`
	Status      int     `json:"status"`
	MenuIDs     []int64 `json:"menu_ids"`
}

type RoleOption struct {
	Label string `json:"label"`
	Value int64  `json:"value"`
	Code  string `json:"code"`
}

type ListItem struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Code        string    `json:"code"`
	Description string    `json:"description"`
	Status      int       `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type DetailResponse struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Code        string    `json:"code"`
	Description string    `json:"description"`
	Status      int       `json:"status"`
	MenuIDs     []int64   `json:"menu_ids"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type SaveResult struct {
	ID int64 `json:"id"`
}

type DeleteResult struct {
	Deleted int `json:"deleted"`
}

type roleListFilter struct {
	KeywordPlain string
	NamePlain    string
	CodePlain    string
	Status       *int64
	SortBy       string
	SortOrder    string
}
