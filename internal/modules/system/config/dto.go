package system_config

import (
	"encoding/json"
	"time"
)

type ListParams struct {
	Page     int
	PageSize int
	Keyword  string
	Filters  map[string]any
}

type SaveRequest struct {
	ID          int64           `json:"id"`
	ConfigKey   string          `json:"config_key"`
	ConfigName  string          `json:"config_name"`
	ConfigValue json.RawMessage `json:"config_value"`
	Remark      string          `json:"remark"`
}

type ItemResponse struct {
	ID          int64           `json:"id"`
	Dir         string          `json:"dir"`
	ConfigKey   string          `json:"config_key"`
	ConfigName  string          `json:"config_name"`
	ConfigValue json.RawMessage `json:"config_value"`
	Remark      string          `json:"remark"`
	Type        string          `json:"type"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

type SaveResult struct {
	ID int64 `json:"id"`
}

type DeleteResult struct {
	Deleted int `json:"deleted"`
}

type configListFilter struct {
	Keyword string
	Key     string
	Name    string
}
