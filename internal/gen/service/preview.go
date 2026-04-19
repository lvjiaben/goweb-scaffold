package service

import (
	"encoding/json"
	"fmt"
	"strings"
)

func BuildPreview(moduleName string, tableName string, rawPayload json.RawMessage, columns []ColumnInfo) Preview {
	moduleCode := ToSnake(moduleName)
	modulePath := ToKebab(moduleName)
	payload := parsePayload(rawPayload, columns)

	inferredFields := inferFields(columns)
	fieldIndex := make(map[string]InferredField, len(inferredFields))
	for _, item := range inferredFields {
		fieldIndex[item.ColumnName] = item
	}

	return Preview{
		ModuleName: moduleName,
		TableName:  tableName,
		Page: PageMeta{
			RoutePath: fmt.Sprintf("/system/%s", modulePath),
			PageName:  ToPascal(moduleName) + "Page",
			ViewFile:  fmt.Sprintf("views/system/%sPage.vue", ToPascal(moduleName)),
		},
		API: APIMeta{
			ModuleCode: moduleCode,
			List:       fmt.Sprintf("/admin-api/%s/list", moduleCode),
			Detail:     fmt.Sprintf("/admin-api/%s/detail", moduleCode),
			Save:       fmt.Sprintf("/admin-api/%s/save", moduleCode),
			Delete:     fmt.Sprintf("/admin-api/%s/delete", moduleCode),
		},
		InferredFields: inferredFields,
		FormSchema:     buildFormSchema(payload.FormFields, fieldIndex),
		ListSchema:     buildListSchema(payload.ListFields, fieldIndex),
		SearchSchema:   buildSearchSchema(payload.SearchFields, fieldIndex),
		Payload:        payload,
		Notes:          buildPreviewNotes(columns, inferredFields),
	}
}

func parsePayload(rawPayload json.RawMessage, columns []ColumnInfo) PayloadConfig {
	payload := PayloadConfig{
		ListFields:   []string{},
		FormFields:   []string{},
		SearchFields: []string{},
	}
	if len(rawPayload) > 0 {
		_ = json.Unmarshal(rawPayload, &payload)
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

	if len(payload.ListFields) == 0 {
		payload.ListFields = defaultList
	}
	if len(payload.FormFields) == 0 {
		payload.FormFields = defaultForm
	}
	if len(payload.SearchFields) == 0 {
		payload.SearchFields = defaultSearch
	}

	payload.ListFields = uniqueStrings(payload.ListFields)
	payload.FormFields = uniqueStrings(payload.FormFields)
	payload.SearchFields = uniqueStrings(payload.SearchFields)
	return payload
}

func inferFields(columns []ColumnInfo) []InferredField {
	fields := make([]InferredField, 0, len(columns))
	for _, column := range columns {
		fields = append(fields, InferredField{
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

func buildFormSchema(fields []string, fieldIndex map[string]InferredField) []SchemaField {
	result := make([]SchemaField, 0, len(fields))
	for _, field := range fields {
		item, ok := fieldIndex[field]
		if !ok {
			continue
		}
		readonly := isReadonlyField(item.ColumnName, item.IsPrimaryKey)
		result = append(result, SchemaField{
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

func buildListSchema(fields []string, fieldIndex map[string]InferredField) []SchemaField {
	result := make([]SchemaField, 0, len(fields))
	for _, field := range fields {
		item, ok := fieldIndex[field]
		if !ok {
			continue
		}
		result = append(result, SchemaField{
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

func buildSearchSchema(fields []string, fieldIndex map[string]InferredField) []SchemaField {
	result := make([]SchemaField, 0, len(fields))
	for _, field := range fields {
		item, ok := fieldIndex[field]
		if !ok {
			continue
		}
		result = append(result, SchemaField{
			Field:      item.ColumnName,
			Label:      columnLabel(item.ColumnName),
			Component:  guessSearchComponent(item),
			Operator:   guessSearchOperator(item),
			Searchable: item.GuessedSearchable,
		})
	}
	return result
}

func buildPreviewNotes(columns []ColumnInfo, inferredFields []InferredField) []string {
	notes := []string{
		"当前阶段会生成真实 admin CRUD 文件，不生成 user 端页面。",
		"生成器输出包含后端模块、admin API 文件、admin 页面和可重建注册文件。",
		"相同输入重复生成时，生成器会保持稳定输出。",
	}

	for _, column := range columns {
		switch {
		case column.IsPrimaryKey:
			notes = append(notes, fmt.Sprintf("%s 被识别为主键，默认只在列表展示，不进入可编辑表单。", column.ColumnName))
		case isSoftDeleteField(column.ColumnName):
			notes = append(notes, fmt.Sprintf("%s 被识别为软删除字段，默认不进入表单和搜索区。", column.ColumnName))
		case column.ColumnName == "created_at" || column.ColumnName == "updated_at":
			notes = append(notes, fmt.Sprintf("%s 被识别为时间审计字段，默认作为列表字段展示。", column.ColumnName))
		}
	}

	for _, item := range inferredFields {
		if item.DataType == "jsonb" {
			notes = append(notes, fmt.Sprintf("%s 是 jsonb 字段，默认建议使用 JSON 文本编辑。", item.ColumnName))
		}
	}

	return sortedUniqueStrings(notes)
}

func guessFormComponent(column ColumnInfo) string {
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
	case strings.EqualFold(strings.TrimSpace(column.DataType), "jsonb"):
		return "json-editor"
	case isLongTextField(name, column.DataType):
		return "textarea"
	default:
		return "text-input"
	}
}

func guessListDisplay(column ColumnInfo) string {
	switch {
	case isBooleanType(column.DataType):
		return "boolean-tag"
	case isTimestampType(column.DataType):
		return "datetime"
	case strings.EqualFold(strings.TrimSpace(column.DataType), "jsonb"):
		return "json-preview"
	case column.IsPrimaryKey:
		return "id"
	default:
		return "text"
	}
}

func canGuessSearchable(column ColumnInfo) bool {
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

func canGuessSortable(column ColumnInfo) bool {
	if strings.EqualFold(strings.TrimSpace(column.DataType), "jsonb") || isLongTextField(column.ColumnName, column.DataType) {
		return false
	}
	return !isSoftDeleteField(column.ColumnName)
}

func guessSearchComponent(field InferredField) string {
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

func guessSearchOperator(field InferredField) string {
	switch guessSearchComponent(field) {
	case "datetime-range":
		return "between"
	case "select", "number-input":
		return "eq"
	default:
		return "like"
	}
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
	return strings.Contains(lowerName, "remark") || strings.Contains(lowerName, "content") || strings.Contains(lowerName, "description") || strings.Contains(lowerName, "summary")
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

func buildPlaceholder(field InferredField) string {
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
	return HumanizeModuleName(raw)
}
