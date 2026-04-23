package codegen

import (
	"strings"

	"github.com/lvjiaben/goweb-scaffold/internal/bootstrap"
	"github.com/lvjiaben/goweb-scaffold/internal/gen/service"
)

var excludedTables = map[string]struct{}{
	"schema_migrations": {},
	"admin_user":        {},
	"admin_role":        {},
	"admin_menu":        {},
	"admin_user_role":   {},
	"admin_role_menu":   {},
	"admin_login_log":   {},
	"admin_session":     {},
	"app_user":          {},
	"app_user_session":  {},
	"system_config":     {},
	"file_attachment":   {},
	"codegen_history":   {},
}

type tableInfo struct {
	TableName    string `json:"table_name" gorm:"column:table_name"`
	TableComment string `json:"table_comment" gorm:"column:table_comment"`
}

type BusinessTable struct {
	TableName    string `json:"table_name"`
	DisplayName  string `json:"display_name"`
	TableComment string `json:"table_comment,omitempty"`
}

func listBusinessTables(runtime *bootstrap.Runtime) ([]BusinessTable, error) {
	return NewRepo(runtime).BusinessTables()
}

func listTableColumns(runtime *bootstrap.Runtime, tableName string) ([]service.ColumnInfo, error) {
	return NewRepo(runtime).TableColumns(tableName)
}

func isExcludedTable(tableName string) bool {
	name := strings.TrimSpace(tableName)
	if name == "" {
		return true
	}
	if _, ok := excludedTables[name]; ok {
		return true
	}
	return strings.HasPrefix(name, "pg_") || strings.HasPrefix(name, "sql_")
}
