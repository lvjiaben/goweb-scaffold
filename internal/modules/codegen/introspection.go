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

type previewField struct {
	ColumnName           string `json:"column_name"`
	DataType             string `json:"data_type"`
	IsNullable           bool   `json:"is_nullable"`
	IsPrimaryKey         bool   `json:"is_primary_key"`
	GuessedFormComponent string `json:"guessed_form_component"`
	GuessedListDisplay   string `json:"guessed_list_display"`
	GuessedSearchable    bool   `json:"guessed_searchable"`
	GuessedSortable      bool   `json:"guessed_sortable"`
}

type schemaField struct {
	Field       string `json:"field"`
	Label       string `json:"label"`
	Component   string `json:"component"`
	Display     string `json:"display,omitempty"`
	Operator    string `json:"operator,omitempty"`
	Required    bool   `json:"required,omitempty"`
	Readonly    bool   `json:"readonly,omitempty"`
	Hidden      bool   `json:"hidden,omitempty"`
	Searchable  bool   `json:"searchable,omitempty"`
	Sortable    bool   `json:"sortable,omitempty"`
	Placeholder string `json:"placeholder,omitempty"`
}

type previewPayloadConfig struct {
	ListFields   []string `json:"list_fields"`
	FormFields   []string `json:"form_fields"`
	SearchFields []string `json:"search_fields"`
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
	config := parsePreviewPayload(rawPayload, columns)
	inferredFields := inferPreviewFields(columns)
	fieldIndex := make(map[string]previewField, len(inferredFields))
	for _, item := range inferredFields {
		fieldIndex[item.ColumnName] = item
	}

	return map[string]any{
		"module_name": moduleName,
		"table_name":  tableName,
		"page": map[string]any{
			"route_path": fmt.Sprintf("/system/%s", modulePath),
			"page_name":  toPascal(moduleName) + "Page",
			"view_file":  fmt.Sprintf("views/system/%sPage.vue", toPascal(moduleName)),
		},
		"api": map[string]any{
			"module_code": apiModule,
			"list":        fmt.Sprintf("/admin-api/%s/list", apiModule),
			"detail":      fmt.Sprintf("/admin-api/%s/detail", apiModule),
			"save":        fmt.Sprintf("/admin-api/%s/save", apiModule),
			"delete":      fmt.Sprintf("/admin-api/%s/delete", apiModule),
		},
		"inferred_fields": inferredFields,
		"form_schema":     buildFormSchema(config.FormFields, fieldIndex),
		"list_schema":     buildListSchema(config.ListFields, fieldIndex),
		"search_schema":   buildSearchSchema(config.SearchFields, fieldIndex),
		"payload":         config,
		"notes":           buildPreviewNotes(columns, inferredFields),
	}
}

func parsePreviewPayload(rawPayload json.RawMessage, columns []columnInfo) previewPayloadConfig {
	config := previewPayloadConfig{
		ListFields:   []string{},
		FormFields:   []string{},
		SearchFields: []string{},
	}
	if len(rawPayload) > 0 {
		_ = json.Unmarshal(rawPayload, &config)
	}

	defaultList := make([]string, 0, len(columns))
	defaultForm := make([]string, 0, len(columns))
	defaultSearch := make([]string, 0, 3)

	for _, column := range columns {
		name := column.ColumnName
		if name == "deleted_at" {
			continue
		}
		defaultList = append(defaultList, name)

		if !isReadonlyField(name, column.IsPrimaryKey) && !isSoftDeleteField(name) {
			defaultForm = append(defaultForm, name)
		}

		if canGuessSearchable(column) && len(defaultSearch) < 3 {
			defaultSearch = append(defaultSearch, name)
		}
	}

	if len(config.ListFields) == 0 {
		config.ListFields = defaultList
	}
	if len(config.FormFields) == 0 {
		config.FormFields = defaultForm
	}
	if len(config.SearchFields) == 0 {
		config.SearchFields = defaultSearch
	}

	config.ListFields = uniqueStrings(config.ListFields)
	config.FormFields = uniqueStrings(config.FormFields)
	config.SearchFields = uniqueStrings(config.SearchFields)
	return config
}

func inferPreviewFields(columns []columnInfo) []previewField {
	fields := make([]previewField, 0, len(columns))
	for _, column := range columns {
		fields = append(fields, previewField{
			ColumnName:           column.ColumnName,
			DataType:             column.DataType,
			IsNullable:           column.IsNullable,
			IsPrimaryKey:         column.IsPrimaryKey,
			GuessedFormComponent: guessFormComponent(column),
			GuessedListDisplay:   guessListDisplay(column),
			GuessedSearchable:    canGuessSearchable(column),
			GuessedSortable:      canGuessSortable(column),
		})
	}
	return fields
}

func buildFormSchema(fields []string, fieldIndex map[string]previewField) []schemaField {
	result := make([]schemaField, 0, len(fields))
	for _, field := range fields {
		item, ok := fieldIndex[field]
		if !ok {
			continue
		}
		readonly := isReadonlyField(item.ColumnName, item.IsPrimaryKey)
		result = append(result, schemaField{
			Field:       item.ColumnName,
			Label:       columnLabel(item.ColumnName),
			Component:   item.GuessedFormComponent,
			Required:    !item.IsNullable && !readonly,
			Readonly:    readonly,
			Hidden:      isSoftDeleteField(item.ColumnName),
			Placeholder: buildPlaceholder(item),
		})
	}
	return result
}

func buildListSchema(fields []string, fieldIndex map[string]previewField) []schemaField {
	result := make([]schemaField, 0, len(fields))
	for _, field := range fields {
		item, ok := fieldIndex[field]
		if !ok {
			continue
		}
		result = append(result, schemaField{
			Field:      item.ColumnName,
			Label:      columnLabel(item.ColumnName),
			Display:    item.GuessedListDisplay,
			Sortable:   item.GuessedSortable,
			Searchable: item.GuessedSearchable,
			Hidden:     isSoftDeleteField(item.ColumnName),
		})
	}
	return result
}

func buildSearchSchema(fields []string, fieldIndex map[string]previewField) []schemaField {
	result := make([]schemaField, 0, len(fields))
	for _, field := range fields {
		item, ok := fieldIndex[field]
		if !ok {
			continue
		}
		result = append(result, schemaField{
			Field:      item.ColumnName,
			Label:      columnLabel(item.ColumnName),
			Component:  guessSearchComponent(item),
			Operator:   guessSearchOperator(item),
			Searchable: item.GuessedSearchable,
		})
	}
	return result
}

func buildPreviewNotes(columns []columnInfo, inferredFields []previewField) []string {
	notes := []string{
		"当前阶段只输出方案生成稿，不生成真实文件。",
		"下一阶段将基于 form_schema、list_schema、search_schema 生成 admin 后台 CRUD。",
		"代码生成器仍然只允许生成 admin 页面，不生成 user 端页面。",
	}

	for _, column := range columns {
		switch {
		case column.IsPrimaryKey:
			notes = append(notes, fmt.Sprintf("%s 被识别为主键，默认只在列表展示，不建议作为可编辑字段。", column.ColumnName))
		case isSoftDeleteField(column.ColumnName):
			notes = append(notes, fmt.Sprintf("%s 被识别为软删除字段，默认不进入表单和搜索区。", column.ColumnName))
		case column.ColumnName == "created_at" || column.ColumnName == "updated_at":
			notes = append(notes, fmt.Sprintf("%s 被识别为时间审计字段，默认作为列表字段展示。", column.ColumnName))
		}
	}

	for _, item := range inferredFields {
		if item.DataType == "jsonb" {
			notes = append(notes, fmt.Sprintf("%s 是 jsonb 字段，默认建议使用 JSON 编辑器或 textarea。", item.ColumnName))
		}
	}

	return uniqueStrings(notes)
}

func guessFormComponent(column columnInfo) string {
	name := column.ColumnName
	switch {
	case column.IsPrimaryKey:
		return "readonly-text"
	case isSoftDeleteField(name):
		return "hidden"
	case name == "created_at" || name == "updated_at":
		return "readonly-datetime"
	case isBooleanType(column.DataType):
		return "switch"
	case isIntegerType(column.DataType):
		return "number-input"
	case isTimestampType(column.DataType):
		return "datetime-picker"
	case column.DataType == "jsonb":
		return "json-editor"
	case isLongTextField(name, column.DataType):
		return "textarea"
	default:
		return "text-input"
	}
}

func guessListDisplay(column columnInfo) string {
	switch {
	case isBooleanType(column.DataType):
		return "boolean-tag"
	case isTimestampType(column.DataType):
		return "datetime"
	case column.DataType == "jsonb":
		return "json-preview"
	case column.IsPrimaryKey:
		return "id"
	default:
		return "text"
	}
}

func canGuessSearchable(column columnInfo) bool {
	name := column.ColumnName
	if column.IsPrimaryKey || isSoftDeleteField(name) || name == "created_at" || name == "updated_at" {
		return false
	}
	switch {
	case isTextType(column.DataType):
		return true
	case isIntegerType(column.DataType):
		return strings.HasSuffix(name, "_id") || strings.Contains(name, "status") || strings.Contains(name, "type")
	case isBooleanType(column.DataType):
		return true
	case isTimestampType(column.DataType):
		return true
	default:
		return false
	}
}

func canGuessSortable(column columnInfo) bool {
	if column.DataType == "jsonb" || isLongTextField(column.ColumnName, column.DataType) {
		return false
	}
	return !isSoftDeleteField(column.ColumnName)
}

func guessSearchComponent(field previewField) string {
	switch {
	case field.GuessedFormComponent == "switch":
		return "select"
	case field.GuessedFormComponent == "datetime-picker" || strings.Contains(field.ColumnName, "_at"):
		return "datetime-range"
	case field.GuessedFormComponent == "number-input":
		return "number-input"
	default:
		return "text-input"
	}
}

func guessSearchOperator(field previewField) string {
	switch guessSearchComponent(field) {
	case "datetime-range":
		return "between"
	case "select":
		return "eq"
	default:
		return "like"
	}
}

func isIntegerType(dataType string) bool {
	switch strings.ToLower(strings.TrimSpace(dataType)) {
	case "bigint", "integer", "smallint":
		return true
	default:
		return false
	}
}

func isBooleanType(dataType string) bool {
	return strings.EqualFold(strings.TrimSpace(dataType), "boolean")
}

func isTimestampType(dataType string) bool {
	normalized := strings.ToLower(strings.TrimSpace(dataType))
	return normalized == "timestamp with time zone" || normalized == "timestamp without time zone" || normalized == "timestamp" || normalized == "timestamptz"
}

func isTextType(dataType string) bool {
	switch strings.ToLower(strings.TrimSpace(dataType)) {
	case "character varying", "varchar", "text":
		return true
	default:
		return false
	}
}

func isLongTextField(name string, dataType string) bool {
	if strings.EqualFold(strings.TrimSpace(dataType), "text") {
		return true
	}
	lowerName := strings.ToLower(strings.TrimSpace(name))
	return strings.Contains(lowerName, "remark") || strings.Contains(lowerName, "content") || strings.Contains(lowerName, "description")
}

func isSoftDeleteField(name string) bool {
	return strings.EqualFold(strings.TrimSpace(name), "deleted_at")
}

func isReadonlyField(name string, isPrimaryKey bool) bool {
	if isPrimaryKey {
		return true
	}
	switch strings.TrimSpace(name) {
	case "created_at", "updated_at":
		return true
	default:
		return false
	}
}

func buildPlaceholder(field previewField) string {
	switch field.GuessedFormComponent {
	case "number-input":
		return "请输入数字"
	case "switch":
		return "请选择开关状态"
	case "datetime-picker":
		return "请选择时间"
	case "json-editor":
		return "请输入 JSON 内容"
	case "textarea":
		return "请输入详细内容"
	default:
		return "请输入" + columnLabel(field.ColumnName)
	}
}

func columnLabel(raw string) string {
	switch strings.TrimSpace(raw) {
	case "id":
		return "ID"
	case "created_at":
		return "创建时间"
	case "updated_at":
		return "更新时间"
	case "deleted_at":
		return "删除时间"
	}

	parts := strings.Split(strings.TrimSpace(raw), "_")
	for index, part := range parts {
		if part == "" {
			continue
		}
		parts[index] = strings.ToUpper(part[:1]) + part[1:]
	}
	return strings.Join(parts, " ")
}

func uniqueStrings(values []string) []string {
	result := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		item := strings.TrimSpace(value)
		if item == "" {
			continue
		}
		if _, ok := seen[item]; ok {
			continue
		}
		seen[item] = struct{}{}
		result = append(result, item)
	}
	return result
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
