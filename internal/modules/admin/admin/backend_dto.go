package admin_admin

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
	ID       int64   `json:"id"`
	Username string  `json:"username"`
	Password string  `json:"password"`
	Nickname string  `json:"nickname"`
	RealName string  `json:"realname"`
	Status   int     `json:"status"`
	IsSuper  bool    `json:"is_super"`
	RoleIDs  []int64 `json:"role_ids"`
}

type ListItem struct {
	ID        int64     `json:"id"`
	Username  string    `json:"username"`
	Nickname  string    `json:"nickname"`
	RealName  string    `json:"realname"`
	Status    int       `json:"status"`
	IsSuper   bool      `json:"is_super"`
	RoleIDs   []int64   `json:"role_ids"`
	RoleNames []string  `json:"role_names"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type DetailResponse struct {
	ID        int64     `json:"id"`
	Username  string    `json:"username"`
	Nickname  string    `json:"nickname"`
	RealName  string    `json:"realname"`
	Status    int       `json:"status"`
	IsSuper   bool      `json:"is_super"`
	RoleIDs   []int64   `json:"role_ids"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type SaveResult struct {
	ID int64 `json:"id"`
}

type DeleteResult struct {
	Deleted int `json:"deleted"`
}

type userListFilter struct {
	KeywordPlain  string
	UsernamePlain string
	NicknamePlain string
	Status        *int64
	SortBy        string
	SortOrder     string
}
