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
	tableComment := detectTableComment(columns)

	inferredFields := inferFields(columns)
	fieldIndex := make(map[string]InferredField, len(inferredFields))
	columnIndex := make(map[string]ColumnInfo, len(columns))
	for _, item := range inferredFields {
		fieldIndex[item.ColumnName] = item
	}
	for _, item := range columns {
		columnIndex[item.ColumnName] = item
	}

	title := firstNonEmpty(strings.TrimSpace(payload.Title), strings.TrimSpace(tableComment), HumanizeModuleName(moduleName))

	return Preview{
		ModuleName:   moduleName,
		TableName:    tableName,
		TableComment: tableComment,
		Page: PageMeta{
			RoutePath:    fmt.Sprintf("/system/%s", modulePath),
			PageName:     ToPascal(moduleName) + "Page",
			ViewFile:     fmt.Sprintf("views/system/%sPage.vue", ToPascal(moduleName)),
			MenuTitle:    title,
			FeatureFlags: []string{"admin-crud", "codegen"},
		},
		API: APIMeta{
			ModuleCode: moduleCode,
			List:       fmt.Sprintf("/admin-api/%s/list", moduleCode),
			Detail:     fmt.Sprintf("/admin-api/%s/detail", moduleCode),
			Save:       fmt.Sprintf("/admin-api/%s/save", moduleCode),
			Delete:     fmt.Sprintf("/admin-api/%s/delete", moduleCode),
		},
		InferredFields: inferredFields,
		FormSchema:     buildFormSchema(payload.FormFields, fieldIndex, columnIndex, payload.FieldOverrides),
		ListSchema:     buildListSchema(payload.ListFields, fieldIndex, columnIndex, payload.FieldOverrides),
		SearchSchema:   buildSearchSchema(payload.SearchFields, fieldIndex, columnIndex, payload.FieldOverrides),
		Payload:        payload,
		Notes:          buildPreviewNotes(columns, inferredFields, title),
	}
}

func parsePayload(rawPayload json.RawMessage, columns []ColumnInfo) PayloadConfig {
	payload := PayloadConfig{
		ListFields:     []string{},
		FormFields:     []string{},
		SearchFields:   []string{},
		FieldOverrides: map[string]FieldOverride{},
	}
	if len(rawPayload) > 0 {
		_ = json.Unmarshal(rawPayload, &payload)
	}
	if payload.FieldOverrides == nil {
		payload.FieldOverrides = map[string]FieldOverride{}
	}

	defaultList := make([]string, 0, len(columns))
	defaultForm := make([]string, 0, len(columns))
	defaultSearch := make([]string, 0, 4)

	for _, column := range columns {
		name := column.ColumnName
		if name == "deleted_at" {
			continue
		}
		defaultList = append(defaultList, name)

		if !isReadonlyField(name, column.IsPrimaryKey) && !isSoftDeleteField(name) {
			defaultForm = append(defaultForm, name)
		}

		if canGuessSearchable(column) && len(defaultSearch) < 4 {
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
			ColumnComment:        strings.TrimSpace(column.ColumnComment),
			GuessedLabel:         preferredLabel(column),
			GuessedFormComponent: guessFormComponent(column),
			GuessedListDisplay:   guessListDisplay(column),
			GuessedSearchable:    canGuessSearchable(column),
			GuessedSortable:      canGuessSortable(column),
		})
	}
	return fields
}

func buildFormSchema(fields []string, inferred map[string]InferredField, columns map[string]ColumnInfo, overrides map[string]FieldOverride) []SchemaField {
	result := make([]SchemaField, 0, len(fields))
	for _, field := range fields {
		inferredField, ok := inferred[field]
		if !ok {
			continue
		}
		column := columns[field]
		override := overrides[field]
		component := firstNonEmpty(override.Component, inferredField.GuessedFormComponent)
		options := buildOptions(column, component, override)
		required := pickBool(override.Required, !column.IsNullable && !isReadonlyField(field, column.IsPrimaryKey))
		readonly := pickBool(override.Readonly, isReadonlyField(field, column.IsPrimaryKey))
		hidden := pickBool(override.Hidden, isSoftDeleteField(field))
		result = append(result, SchemaField{
			Field:        field,
			Label:        firstNonEmpty(override.Label, inferredField.GuessedLabel),
			Component:    component,
			Required:     required,
			Readonly:     readonly,
			Hidden:       hidden,
			Placeholder:  firstNonEmpty(override.Placeholder, buildPlaceholder(firstNonEmpty(override.Label, inferredField.GuessedLabel), component)),
			Width:        strings.TrimSpace(override.Width),
			Options:      options,
			DefaultValue: firstNonNil(override.DefaultValue, guessDefaultValue(column, component)),
		})
	}
	return result
}

func buildListSchema(fields []string, inferred map[string]InferredField, columns map[string]ColumnInfo, overrides map[string]FieldOverride) []SchemaField {
	result := make([]SchemaField, 0, len(fields))
	for _, field := range fields {
		inferredField, ok := inferred[field]
		if !ok {
			continue
		}
		column := columns[field]
		override := overrides[field]
		component := firstNonEmpty(override.Component, inferredField.GuessedFormComponent)
		options := buildOptions(column, component, override)
		display := guessedListDisplay(column, component, options)
		result = append(result, SchemaField{
			Field:        field,
			Label:        firstNonEmpty(override.Label, inferredField.GuessedLabel),
			Component:    component,
			Display:      display,
			Hidden:       pickBool(override.Hidden, isSoftDeleteField(field)),
			Searchable:   pickBool(override.Searchable, inferredField.GuessedSearchable),
			Sortable:     pickBool(override.Sortable, inferredField.GuessedSortable),
			Width:        strings.TrimSpace(override.Width),
			Options:      options,
			DefaultValue: firstNonNil(override.DefaultValue, guessDefaultValue(column, component)),
		})
	}
	return result
}

func buildSearchSchema(fields []string, inferred map[string]InferredField, columns map[string]ColumnInfo, overrides map[string]FieldOverride) []SchemaField {
	result := make([]SchemaField, 0, len(fields))
	for _, field := range fields {
		inferredField, ok := inferred[field]
		if !ok {
			continue
		}
		column := columns[field]
		override := overrides[field]
		formComponent := firstNonEmpty(override.Component, inferredField.GuessedFormComponent)
		options := buildOptions(column, formComponent, override)
		component := guessSearchComponent(column, formComponent, options)
		searchable := pickBool(override.Searchable, inferredField.GuessedSearchable)
		result = append(result, SchemaField{
			Field:        field,
			Label:        firstNonEmpty(override.Label, inferredField.GuessedLabel),
			Component:    component,
			Operator:     guessSearchOperator(component),
			Searchable:   searchable,
			Width:        strings.TrimSpace(override.Width),
			Options:      options,
			Placeholder:  firstNonEmpty(override.Placeholder, buildSearchPlaceholder(firstNonEmpty(override.Label, inferredField.GuessedLabel), component)),
			DefaultValue: firstNonNil(override.DefaultValue, guessDefaultValue(column, formComponent)),
		})
	}
	return result
}

func buildPreviewNotes(columns []ColumnInfo, inferredFields []InferredField, title string) []string {
	notes := []string{
		"当前阶段会生成真实 admin CRUD 文件，不生成 user 端页面。",
		"生成器输出包含后端模块、admin API 文件、admin 页面、lock 文件和可重建注册文件。",
		"preview、diff、generate、regenerate 共用同一套字段推断和字段级 overrides。",
		fmt.Sprintf("当前模块标题候选为 %s。", title),
	}

	for _, column := range columns {
		switch {
		case column.IsPrimaryKey:
			notes = append(notes, fmt.Sprintf("%s 被识别为主键，默认只在列表展示，不进入可编辑表单。", column.ColumnName))
		case isSoftDeleteField(column.ColumnName):
			notes = append(notes, fmt.Sprintf("%s 被识别为软删除字段，默认不进入表单和搜索区。", column.ColumnName))
		case column.ColumnName == "created_at" || column.ColumnName == "updated_at":
			notes = append(notes, fmt.Sprintf("%s 被识别为时间审计字段，默认作为列表字段展示。", column.ColumnName))
		case strings.TrimSpace(column.ColumnComment) != "":
			notes = append(notes, fmt.Sprintf("%s 读取到了列注释，preview 会优先使用注释作为字段标签候选。", column.ColumnName))
		}
	}

	for _, item := range inferredFields {
		if strings.EqualFold(strings.TrimSpace(item.DataType), "jsonb") {
			notes = append(notes, fmt.Sprintf("%s 是 jsonb 字段，默认建议使用 JSON 文本编辑。", item.ColumnName))
		}
	}

	return sortedUniqueStrings(notes)
}

func guessFormComponent(column ColumnInfo) string {
	name := strings.ToLower(strings.TrimSpace(column.ColumnName))
	switch {
	case column.IsPrimaryKey:
		return "readonly-text"
	case isSoftDeleteField(name):
		return "hidden"
	case name == "created_at" || name == "updated_at":
		return "readonly-datetime"
	case isBooleanType(column.DataType):
		return "switch"
	case strings.HasPrefix(name, "is_") || strings.HasPrefix(name, "has_"):
		return "switch"
	case name == "status" || name == "state":
		return "select"
	case name == "sort":
		return "number-input"
	case strings.HasSuffix(name, "_at") || isTimestampType(column.DataType):
		return "datetime-picker"
	case strings.EqualFold(strings.TrimSpace(column.DataType), "jsonb"):
		return "json-editor"
	case isLongTextField(name, column.DataType):
		return "textarea"
	case isBigIntegerType(column.DataType) || isIntegerType(column.DataType):
		return "number-input"
	default:
		return "text-input"
	}
}

func guessListDisplay(column ColumnInfo) string {
	return guessedListDisplay(column, guessFormComponent(column), buildOptions(column, guessFormComponent(column), FieldOverride{}))
}

func guessedListDisplay(column ColumnInfo, component string, options []FieldOption) string {
	switch {
	case component == "switch" || isBooleanType(column.DataType):
		return "boolean-tag"
	case len(options) > 0 && (component == "select" || component == "radio"):
		return "option-tag"
	case strings.HasSuffix(strings.ToLower(strings.TrimSpace(column.ColumnName)), "_at") || isTimestampType(column.DataType):
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
	name := strings.ToLower(strings.TrimSpace(column.ColumnName))
	if column.IsPrimaryKey || isSoftDeleteField(name) || name == "created_at" || name == "updated_at" {
		return false
	}
	switch {
	case isTextType(column.DataType):
		return true
	case isBigIntegerType(column.DataType), isIntegerType(column.DataType):
		return strings.HasSuffix(name, "_id") || strings.Contains(name, "status") || strings.Contains(name, "state") || strings.Contains(name, "type")
	case isBooleanType(column.DataType):
		return true
	case strings.HasPrefix(name, "is_") || strings.HasPrefix(name, "has_"):
		return true
	case strings.HasSuffix(name, "_at") || isTimestampType(column.DataType):
		return true
	default:
		return false
	}
}

func canGuessSortable(column ColumnInfo) bool {
	name := strings.ToLower(strings.TrimSpace(column.ColumnName))
	if isSoftDeleteField(name) {
		return false
	}
	if strings.EqualFold(strings.TrimSpace(column.DataType), "jsonb") || isLongTextField(name, column.DataType) {
		return false
	}
	return true
}

func guessSearchComponent(column ColumnInfo, formComponent string, options []FieldOption) string {
	switch {
	case len(options) > 0 && (formComponent == "select" || formComponent == "radio" || formComponent == "switch"):
		return "select"
	case formComponent == "switch":
		return "select"
	case formComponent == "datetime-picker" || formComponent == "readonly-datetime" || strings.HasSuffix(strings.ToLower(strings.TrimSpace(column.ColumnName)), "_at"):
		return "datetime-range"
	case formComponent == "number-input":
		return "number-input"
	default:
		return "text-input"
	}
}

func guessSearchOperator(component string) string {
	switch component {
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

func buildPlaceholder(label string, component string) string {
	switch component {
	case "number-input":
		return "请输入数字"
	case "switch":
		return "请选择开关状态"
	case "select", "radio":
		return "请选择" + label
	case "datetime-picker":
		return "请选择时间"
	case "json-editor":
		return "请输入 JSON 内容"
	case "textarea":
		return "请输入详细内容"
	default:
		return "请输入" + label
	}
}

func buildSearchPlaceholder(label string, component string) string {
	switch component {
	case "select":
		return "请选择" + label
	case "datetime-range":
		return "请选择" + label + "范围"
	case "number-input":
		return "请输入" + label
	default:
		return "搜索" + label
	}
}

func preferredLabel(column ColumnInfo) string {
	if comment := strings.TrimSpace(column.ColumnComment); comment != "" {
		return comment
	}
	return columnLabel(column.ColumnName)
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

func buildOptions(column ColumnInfo, component string, override FieldOverride) []FieldOption {
	if len(override.Options) > 0 {
		return normalizeOptions(override.Options)
	}

	name := strings.ToLower(strings.TrimSpace(column.ColumnName))
	switch {
	case component == "switch" || isBooleanType(column.DataType) || strings.HasPrefix(name, "is_") || strings.HasPrefix(name, "has_"):
		return []FieldOption{
			{Label: "是", Value: true},
			{Label: "否", Value: false},
		}
	case component == "select" || component == "radio":
		if name == "status" || name == "state" {
			return []FieldOption{
				{Label: "启用", Value: 1},
				{Label: "禁用", Value: 2},
			}
		}
	}
	return []FieldOption{}
}

func normalizeOptions(items []FieldOption) []FieldOption {
	result := make([]FieldOption, 0, len(items))
	for _, item := range items {
		label := strings.TrimSpace(item.Label)
		if label == "" {
			continue
		}
		result = append(result, FieldOption{Label: label, Value: item.Value})
	}
	return result
}

func guessDefaultValue(column ColumnInfo, component string) any {
	name := strings.ToLower(strings.TrimSpace(column.ColumnName))
	switch {
	case component == "switch" || isBooleanType(column.DataType):
		return parseBoolDefault(column.ColumnDefault)
	case component == "select" || component == "radio":
		if name == "status" || name == "state" {
			return parseIntDefault(column.ColumnDefault, 1)
		}
	case component == "number-input":
		return parseIntDefault(column.ColumnDefault, defaultIntegerHint(name))
	case component == "json-editor":
		return "{}"
	case component == "datetime-picker", component == "readonly-datetime":
		return ""
	}
	return ""
}

func detectTableComment(columns []ColumnInfo) string {
	for _, item := range columns {
		if strings.TrimSpace(item.TableComment) != "" {
			return strings.TrimSpace(item.TableComment)
		}
	}
	return ""
}

func pickBool(value *bool, fallback bool) bool {
	if value == nil {
		return fallback
	}
	return *value
}

func firstNonNil(value any, fallback any) any {
	if value == nil {
		return fallback
	}
	return value
}
