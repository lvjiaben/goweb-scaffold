package app_user

import "time"

type ListParams struct {
	Page      int
	PageSize  int
	Keyword   string
	Filters   map[string]any
	SortBy    string
	SortOrder string
}

type BackendSaveRequest struct {
	ID            int64   `json:"id"`
	PID           int64   `json:"pid"`
	TID           int64   `json:"tid"`
	Username      string  `json:"username"`
	Password      string  `json:"password"`
	Nickname      string  `json:"nickname"`
	Email         string  `json:"email"`
	Mobile        string  `json:"mobile"`
	Avatar        string  `json:"avatar"`
	Code          string  `json:"code"`
	Status        int     `json:"status"`
	StatusText    string  `json:"status_text"`
	WechatUnionID string  `json:"wechat_unionid"`
	WechatOpenID  string  `json:"wechat_openid"`
	Version       int     `json:"version"`
	Money         float64 `json:"money"`
	Score         float64 `json:"score"`
}

type OperateRequest struct {
	IDs   []int64 `json:"ids"`
	Field string  `json:"field"`
	Value any     `json:"value"`
}

type MoneyRequest struct {
	ID     int64   `json:"id"`
	Type   string  `json:"type"`
	Money  float64 `json:"money"`
	Note   string  `json:"note"`
	Source string  `json:"source"`
}

type ScoreRequest struct {
	ID     int64   `json:"id"`
	Type   string  `json:"type"`
	Score  float64 `json:"score"`
	Note   string  `json:"note"`
	Source string  `json:"source"`
}

type LogListParams struct {
	UserID   int64
	Page     int
	PageSize int
}

type UserListItem struct {
	ID            int64     `json:"id"`
	PID           int64     `json:"pid"`
	TID           int64     `json:"tid"`
	Username      string    `json:"username"`
	Nickname      string    `json:"nickname"`
	Email         string    `json:"email"`
	Mobile        string    `json:"mobile"`
	Avatar        string    `json:"avatar"`
	Code          string    `json:"code"`
	Status        int       `json:"status"`
	StatusText    string    `json:"status_text"`
	WechatUnionID string    `json:"wechat_unionid"`
	WechatOpenID  string    `json:"wechat_openid"`
	Version       int       `json:"version"`
	Money         float64   `json:"money"`
	Score         float64   `json:"score"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type SaveResult struct {
	ID int64 `json:"id"`
}

type DeleteResult struct {
	Deleted int `json:"deleted"`
}

type OperateResult struct {
	Updated int `json:"updated"`
}

type AssetUpdateResult struct {
	UserID int64   `json:"user_id"`
	Before float64 `json:"before"`
	After  float64 `json:"after"`
}

type userListFilter struct {
	KeywordPlain string
	ID           *int64
	PID          *int64
	TID          *int64
	Status       *int64
	Username     string
	Mobile       string
	SortBy       string
	SortOrder    string
}
