package codegen

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/lvjiaben/goweb-scaffold/internal/bootstrap"
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

type columnInfo struct {
	ColumnName    string `json:"column_name" gorm:"column:column_name"`
	DataType      string `json:"data_type" gorm:"column:data_type"`
	IsNullable    bool   `json:"is_nullable" gorm:"column:is_nullable"`
	ColumnDefault string `json:"column_default" gorm:"column:column_default"`
	OrdinalPos    int    `json:"ordinal_position" gorm:"column:ordinal_position"`
	IsPrimaryKey  bool   `json:"is_primary_key" gorm:"column:is_primary_key"`
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
		if _, ok := excludedTables[row.TableName]; ok {
			continue
		}
		if strings.HasPrefix(row.TableName, "pg_") || strings.HasPrefix(row.TableName, "sql_") {
			continue
		}
		items = append(items, map[string]any{
			"table_name":   row.TableName,
			"display_name": row.TableName,
		})
	}
	return items, nil
}

func listTableColumns(runtime *bootstrap.Runtime, tableName string) ([]columnInfo, error) {
	if _, ok := excludedTables[tableName]; ok {
		return []columnInfo{}, nil
	}
	var rows []columnInfo
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

func buildPreviewPayload(moduleName string, tableName string, rawPayload json.RawMessage, columns []columnInfo) map[string]any {
	modulePath := strings.ReplaceAll(moduleName, "_", "-")
	apiModule := strings.ReplaceAll(moduleName, "-", "_")

	return map[string]any{
		"module_name": moduleName,
		"table_name":  tableName,
		"page": map[string]any{
			"route_path": fmt.Sprintf("/system/%s", modulePath),
			"page_name":  toPascal(moduleName) + "Page",
			"view_file":  fmt.Sprintf("views/system/%sPage.vue", toPascal(moduleName)),
		},
		"api": map[string]any{
			"list":   fmt.Sprintf("/admin-api/%s/list", apiModule),
			"detail": fmt.Sprintf("/admin-api/%s/detail", apiModule),
			"save":   fmt.Sprintf("/admin-api/%s/save", apiModule),
			"delete": fmt.Sprintf("/admin-api/%s/delete", apiModule),
		},
		"columns": columns,
		"payload": json.RawMessage(rawPayload),
		"notes": []string{
			"当前阶段只输出元数据预览，不生成文件。",
			"将来代码生成器只允许生成 admin 后台 CRUD，不生成 user 端页面。",
		},
	}
}

func toPascal(raw string) string {
	parts := strings.FieldsFunc(raw, func(r rune) bool {
		return r == '_' || r == '-' || r == ' '
	})
	builder := strings.Builder{}
	for _, part := range parts {
		if part == "" {
			continue
		}
		builder.WriteString(strings.ToUpper(part[:1]))
		if len(part) > 1 {
			builder.WriteString(part[1:])
		}
	}
	return builder.String()
}
