package codegen

import (
	"encoding/json"
	"time"
)

type SaveRequest struct {
	ModuleName string          `json:"module_name" validate:"required"`
	TableName  string          `json:"table_name" validate:"required"`
	Payload    json.RawMessage `json:"payload"`
}

type GenerateRequest struct {
	ModuleName     string          `json:"module_name" validate:"required"`
	TableName      string          `json:"table_name" validate:"required"`
	Payload        json.RawMessage `json:"payload"`
	Overwrite      bool            `json:"overwrite"`
	RegisterModule bool            `json:"register_module"`
	UpsertMenu     bool            `json:"upsert_menu"`
}

type CheckBreakingRequest struct {
	ModuleName     string          `json:"module_name" validate:"required"`
	TableName      string          `json:"table_name"`
	Payload        json.RawMessage `json:"payload"`
	RegisterModule *bool           `json:"register_module"`
}

type RegenerateRequest struct {
	ModuleName     string `json:"module_name"`
	HistoryID      int64  `json:"history_id"`
	Overwrite      bool   `json:"overwrite"`
	RegisterModule bool   `json:"register_module"`
	UpsertMenu     bool   `json:"upsert_menu"`
}

type RemoveRequest struct {
	ModuleName       string `json:"module_name" validate:"required"`
	RemoveFiles      bool   `json:"remove_files"`
	UnregisterModule bool   `json:"unregister_module"`
	RemoveMenu       bool   `json:"remove_menu"`
	RemoveHistory    bool   `json:"remove_history"`
	RemoveLock       bool   `json:"remove_lock"`
}

type ExportRequest struct {
	ModuleName string
	HistoryID  int64
}

type HistoryItem struct {
	ID         int64           `json:"id"`
	ModuleName string          `json:"module_name"`
	TableName  string          `json:"table_name"`
	Status     string          `json:"status"`
	Payload    json.RawMessage `json:"payload"`
	Remark     string          `json:"remark"`
	CreatedAt  time.Time       `json:"created_at"`
}

type SaveDraftResult struct {
	ID          int64  `json:"id"`
	Status      string `json:"status"`
	Placeholder bool   `json:"placeholder"`
}

type DeleteHistoryResult struct {
	Deleted int `json:"deleted"`
}

type regenerateSource struct {
	ModuleName string
	TableName  string
	Payload    json.RawMessage
}
