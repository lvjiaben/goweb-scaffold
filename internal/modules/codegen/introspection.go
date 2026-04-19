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
	TableName string `json:"table_name" gorm:"column:table_name"`
}

func listBusinessTables(runtime *bootstrap.Runtime) ([]map[string]any, error) {
	var rows []tableInfo
	if err := runtime.DB.Raw(`
SELECT table_name
FROM information_schema.tables
WHERE table_schema = 'public'
  AND table_type = 'BASE TABLE'
ORDER BY table_name ASC`).Scan(&rows).Error; err != nil {
		return nil, err
	}

	items := make([]map[string]any, 0, len(rows))
	for _, row := range rows {
		if isExcludedTable(row.TableName) {
			continue
		}
		items = append(items, map[string]any{
			"table_name":   row.TableName,
			"display_name": row.TableName,
		})
	}
	return items, nil
}

func listTableColumns(runtime *bootstrap.Runtime, tableName string) ([]service.ColumnInfo, error) {
	if isExcludedTable(tableName) {
		return []service.ColumnInfo{}, nil
	}

	var rows []service.ColumnInfo
	err := runtime.DB.Raw(`
SELECT
  c.column_name,
  c.data_type,
  (c.is_nullable = 'YES') AS is_nullable,
  COALESCE(c.column_default, '') AS column_default,
  c.ordinal_position,
  COALESCE(pk.is_primary_key, FALSE) AS is_primary_key
FROM information_schema.columns c
LEFT JOIN (
  SELECT
    kcu.table_name,
    kcu.column_name,
    TRUE AS is_primary_key
  FROM information_schema.table_constraints tc
  JOIN information_schema.key_column_usage kcu
    ON tc.constraint_name = kcu.constraint_name
   AND tc.table_schema = kcu.table_schema
  WHERE tc.constraint_type = 'PRIMARY KEY'
    AND tc.table_schema = 'public'
) pk
  ON pk.table_name = c.table_name
 AND pk.column_name = c.column_name
WHERE c.table_schema = 'public'
  AND c.table_name = ?
ORDER BY c.ordinal_position ASC`, tableName).Scan(&rows).Error
	return rows, err
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
