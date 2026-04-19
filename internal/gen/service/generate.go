package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/lvjiaben/goweb-scaffold/internal/gen/registry"
	"github.com/lvjiaben/goweb-scaffold/internal/gen/writer"
	"github.com/lvjiaben/goweb-scaffold/internal/modules/model"
	"gorm.io/gorm"
)

type templateField struct {
	ColumnName     string
	GoFieldName    string
	GoType         string
	GormTag        string
	RequestType    string
	RequestKind    string
	FormTSType     string
	SearchTSType   string
	TSType         string
	Component      string
	Display        string
	SearchOperator string
	StartQueryKey  string
	EndQueryKey    string
	Placeholder    string
	IsPrimaryKey   bool
	IsNullable     bool
	IsListField    bool
	IsSaveField    bool
	IsSearchField  bool
	IsBoolean      bool
	IsInteger      bool
	IsBigInteger   bool
	IsTimestamp    bool
	IsJSON         bool
	Required       bool
	Readonly       bool
	Hidden         bool
	DefaultValue   any
}

type backendModelTemplateData struct {
	PackageName   string
	TableName     string
	Fields        []templateField
	UsesJSON      bool
	UsesTime      bool
	UsesDeletedAt bool
}

type backendTypesTemplateData struct {
	PackageName string
	SaveFields  []templateField
	UsesJSON    bool
}

type backendModuleTemplateData struct {
	PackageName      string
	ModuleCode       string
	PermissionList   string
	PermissionSave   string
	PermissionDelete string
	SaveFields       []templateField
	SearchFields     []templateField
	RequiredFields   []templateField
	UsesStrings      bool
	UsesStrconv      bool
	UsesTime         bool
	UsesJSON         bool
}

type frontendAPITemplateData struct {
	ModuleName   string
	ModuleCode   string
	PascalName   string
	Fields       []templateField
	SaveFields   []templateField
	SearchFields []templateField
}

type frontendPageTemplateData struct {
	ModuleName        string
	PascalName        string
	Title             string
	PermissionSave    string
	PermissionDelete  string
	ListFields        []templateField
	ListSchemaJSON    string
	FormSchemaJSON    string
	SearchSchemaJSON  string
	ListColumnsJSON   string
	DefaultFormJSON   string
	SearchDefaultJSON string
}

func (s GeneratorService) Generate(input GenerateInput) (GenerateResult, error) {
	result := GenerateResult{
		GeneratedFiles:   []string{},
		OverwrittenFiles: []string{},
		SkippedFiles:     []string{},
		PermissionCodes:  []string{},
		MenuRecords:      []map[string]any{},
		Warnings:         []string{},
	}
	moduleName := ToSnake(strings.TrimSpace(input.ModuleName))
	tableName := strings.TrimSpace(input.TableName)
	if moduleName == "" || tableName == "" {
		return result, fmt.Errorf("module_name and table_name are required")
	}
	if s.RepoRoot == "" {
		return result, fmt.Errorf("repo root is required")
	}

	preview := input.Preview
	if preview.ModuleName == "" {
		preview = BuildPreview(moduleName, tableName, mustMarshalPayload(input.Payload), input.Columns)
	}

	if len(input.Columns) == 0 {
		return result, fmt.Errorf("columns are required for generation")
	}

	fields := buildTemplateFields(input.Columns, preview)
	paths := ModulePaths(moduleName)
	meta := ModuleMeta{
		ModuleName:  moduleName,
		TableName:   tableName,
		PackageName: moduleName,
		PascalName:  ToPascal(moduleName),
		RoutePath:   preview.Page.RoutePath,
		PageName:    preview.Page.PageName,
		Title:       firstNonEmpty(preview.Payload.Title, HumanizeModuleName(moduleName)),
		ViewFile:    preview.Page.ViewFile,
		PermissionCodes: []string{
			moduleName + ".list",
			moduleName + ".save",
			moduleName + ".delete",
		},
	}

	modelData := backendModelTemplateData{
		PackageName:   moduleName,
		TableName:     tableName,
		Fields:        fields,
		UsesJSON:      anyField(fields, func(item templateField) bool { return item.IsJSON }),
		UsesTime:      anyField(fields, func(item templateField) bool { return item.IsTimestamp }),
		UsesDeletedAt: anyField(fields, func(item templateField) bool { return item.ColumnName == "deleted_at" }),
	}
	typesData := backendTypesTemplateData{
		PackageName: moduleName,
		SaveFields:  selectFields(fields, func(item templateField) bool { return item.IsSaveField }),
		UsesJSON:    anyField(fields, func(item templateField) bool { return item.IsJSON && item.IsSaveField }),
	}
	moduleData := backendModuleTemplateData{
		PackageName:      moduleName,
		ModuleCode:       moduleName,
		PermissionList:   meta.PermissionCodes[0],
		PermissionSave:   meta.PermissionCodes[1],
		PermissionDelete: meta.PermissionCodes[2],
		SaveFields:       selectFields(fields, func(item templateField) bool { return item.IsSaveField }),
		SearchFields:     selectFields(fields, func(item templateField) bool { return item.IsSearchField }),
		RequiredFields: selectFields(fields, func(item templateField) bool {
			return item.IsSaveField && item.Required && (item.RequestKind == "string" || item.RequestKind == "time" || item.RequestKind == "json")
		}),
		UsesStrings: anyField(fields, func(item templateField) bool {
			return item.IsSaveField && item.RequestKind == "string" || item.IsSaveField && item.Required && item.RequestKind == "time" || item.IsSearchField && (item.SearchOperator == "like" || item.SearchOperator == "between")
		}),
		UsesStrconv: anyField(fields, func(item templateField) bool {
			return item.IsSearchField && item.SearchOperator == "eq" && (item.IsBoolean || item.IsInteger || item.IsBigInteger)
		}),
		UsesTime: anyField(fields, func(item templateField) bool {
			return item.IsSearchField && item.SearchOperator == "between" || item.IsSaveField && item.RequestKind == "time"
		}),
		UsesJSON: anyField(fields, func(item templateField) bool { return item.IsSaveField && item.IsJSON }),
	}
	apiData := frontendAPITemplateData{
		ModuleName:   moduleName,
		ModuleCode:   moduleName,
		PascalName:   meta.PascalName,
		Fields:       fields,
		SaveFields:   selectFields(fields, func(item templateField) bool { return item.IsSaveField }),
		SearchFields: selectFields(fields, func(item templateField) bool { return item.IsSearchField }),
	}
	pageData := frontendPageTemplateData{
		ModuleName:        moduleName,
		PascalName:        meta.PascalName,
		Title:             meta.Title,
		PermissionSave:    meta.PermissionCodes[1],
		PermissionDelete:  meta.PermissionCodes[2],
		ListFields:        selectFields(fields, func(item templateField) bool { return item.IsListField }),
		ListSchemaJSON:    MarshalIndent(preview.ListSchema),
		FormSchemaJSON:    MarshalIndent(preview.FormSchema),
		SearchSchemaJSON:  MarshalIndent(preview.SearchSchema),
		ListColumnsJSON:   MarshalIndent(buildListColumns(preview.ListSchema)),
		DefaultFormJSON:   MarshalIndent(buildDefaultFormState(fields)),
		SearchDefaultJSON: MarshalIndent(buildDefaultSearchState(preview.SearchSchema)),
	}

	w := writer.New(s.RepoRoot)
	writePlan := []struct {
		Path         string
		TemplatePath string
		Data         any
	}{
		{Path: paths["model"], TemplatePath: "backend/model.go.tmpl", Data: modelData},
		{Path: paths["types"], TemplatePath: "backend/types.go.tmpl", Data: typesData},
		{Path: paths["module"], TemplatePath: "backend/module.go.tmpl", Data: moduleData},
		{Path: paths["meta"], TemplatePath: "backend/meta.go.tmpl", Data: meta},
		{Path: paths["api"], TemplatePath: "admin_frontend/api.ts.tmpl", Data: apiData},
		{Path: paths["view"], TemplatePath: "admin_frontend/page.vue.tmpl", Data: pageData},
	}

	for _, item := range writePlan {
		content, err := renderTemplate(item.TemplatePath, item.Data)
		if err != nil {
			return result, err
		}
		status, warning, err := w.Write(item.Path, content, input.Overwrite)
		if err != nil {
			return result, err
		}
		result = applyWriteResult(result, item.Path, status, warning)
	}

	if input.RegisterModule {
		status, warning, err := registry.RebuildBackendModulesFile(s.RepoRoot)
		if err != nil {
			return result, err
		}
		result = applyWriteResult(result, "internal/gen/modules_gen.go", status, warning)

		status, warning, err = registry.RebuildAdminRoutesFile(s.RepoRoot)
		if err != nil {
			return result, err
		}
		result = applyWriteResult(result, "vben-admin/apps/admin/src/generated/routes.ts", status, warning)
	} else {
		result.Warnings = append(result.Warnings, "register_module=false，已生成文件但未重建模块注册和前端路由。")
	}

	if input.UpsertMenu {
		if !input.RegisterModule {
			result.Warnings = append(result.Warnings, "register_module=false 且 upsert_menu=true，菜单已写入，但重启前端/后端前路由不会生效。")
		}
		menuResult, warnings, err := s.upsertMenus(meta)
		if err != nil {
			return result, err
		}
		result.MenuRecords = menuResult.Records
		result.Warnings = append(result.Warnings, warnings...)
	} else {
		result.Warnings = append(result.Warnings, "upsert_menu=false，未写入 admin_menu 和 admin_role_menu。")
	}

	result.ModuleName = moduleName
	result.RoutePath = preview.Page.RoutePath
	result.PermissionCodes = append([]string{}, meta.PermissionCodes...)
	result.Warnings = uniqueStrings(result.Warnings)
	return result, nil
}

func buildTemplateFields(columns []ColumnInfo, preview Preview) []templateField {
	formIndex := make(map[string]SchemaField, len(preview.FormSchema))
	listIndex := make(map[string]SchemaField, len(preview.ListSchema))
	searchIndex := make(map[string]SchemaField, len(preview.SearchSchema))
	inferredIndex := make(map[string]InferredField, len(preview.InferredFields))

	for _, item := range preview.FormSchema {
		formIndex[item.Field] = item
	}
	for _, item := range preview.ListSchema {
		listIndex[item.Field] = item
	}
	for _, item := range preview.SearchSchema {
		searchIndex[item.Field] = item
	}
	for _, item := range preview.InferredFields {
		inferredIndex[item.ColumnName] = item
	}

	fields := make([]templateField, 0, len(columns))
	for _, column := range columns {
		item := templateField{
			ColumnName:    column.ColumnName,
			GoFieldName:   GoFieldName(column.ColumnName),
			GoType:        buildGoModelType(column),
			GormTag:       buildGormTag(column),
			RequestType:   RequestTypeForColumn(column),
			RequestKind:   requestKind(column),
			FormTSType:    formTSType(column),
			SearchTSType:  searchTSType(column),
			TSType:        itemTSType(column),
			IsPrimaryKey:  column.IsPrimaryKey,
			IsNullable:    column.IsNullable,
			IsBoolean:     isBooleanType(column.DataType),
			IsInteger:     isIntegerType(column.DataType),
			IsBigInteger:  isBigIntegerType(column.DataType),
			IsTimestamp:   isTimestampType(column.DataType),
			IsJSON:        strings.EqualFold(strings.TrimSpace(column.DataType), "jsonb"),
			StartQueryKey: column.ColumnName + "_start",
			EndQueryKey:   column.ColumnName + "_end",
			DefaultValue:  defaultFormValue(column),
		}

		if inferred, ok := inferredIndex[column.ColumnName]; ok {
			item.Component = inferred.GuessedFormComponent
			item.Display = inferred.GuessedListDisplay
		}
		if schema, ok := formIndex[column.ColumnName]; ok {
			item.IsSaveField = true
			item.Component = schema.Component
			item.Required = schema.Required
			item.Readonly = schema.Readonly
			item.Hidden = schema.Hidden
			item.Placeholder = schema.Placeholder
		}
		if schema, ok := listIndex[column.ColumnName]; ok {
			item.IsListField = true
			if schema.Display != "" {
				item.Display = schema.Display
			}
		}
		if schema, ok := searchIndex[column.ColumnName]; ok {
			item.IsSearchField = true
			item.SearchOperator = schema.Operator
		}
		fields = append(fields, item)
	}
	return fields
}

func buildGoModelType(column ColumnInfo) string {
	switch {
	case column.ColumnName == "deleted_at":
		return "gorm.DeletedAt"
	default:
		return GoTypeForColumn(column)
	}
}

func buildGormTag(column ColumnInfo) string {
	parts := []string{}
	if column.IsPrimaryKey {
		parts = append(parts, "primaryKey")
	}
	parts = append(parts, "column:"+column.ColumnName)
	if column.ColumnName == "deleted_at" {
		parts = append(parts, "index")
	}
	if strings.EqualFold(strings.TrimSpace(column.DataType), "jsonb") {
		parts = append(parts, "type:jsonb")
	}
	return strings.Join(parts, ";")
}

func requestKind(column ColumnInfo) string {
	switch {
	case strings.EqualFold(strings.TrimSpace(column.DataType), "jsonb"):
		return "json"
	case isTimestampType(column.DataType):
		return "time"
	case isBigIntegerType(column.DataType):
		return "int64"
	case isIntegerType(column.DataType):
		return "int"
	case isBooleanType(column.DataType):
		return "bool"
	default:
		return "string"
	}
}

func itemTSType(column ColumnInfo) string {
	switch {
	case isBigIntegerType(column.DataType), isIntegerType(column.DataType), column.IsPrimaryKey:
		return "number"
	case isBooleanType(column.DataType):
		return "boolean"
	case isTimestampType(column.DataType):
		if column.IsNullable {
			return "string | null"
		}
		return "string"
	case strings.EqualFold(strings.TrimSpace(column.DataType), "jsonb"):
		return "Record<string, any>"
	default:
		if column.IsNullable {
			return "string | null"
		}
		return "string"
	}
}

func formTSType(column ColumnInfo) string {
	switch requestKind(column) {
	case "bool":
		return "boolean"
	case "int", "int64":
		return "number"
	case "json":
		return "Record<string, any> | string"
	default:
		return "string"
	}
}

func searchTSType(column ColumnInfo) string {
	switch {
	case isBooleanType(column.DataType):
		return "boolean"
	case isBigIntegerType(column.DataType), isIntegerType(column.DataType), column.IsPrimaryKey:
		return "number"
	default:
		return "string"
	}
}

func defaultFormValue(column ColumnInfo) any {
	switch requestKind(column) {
	case "bool":
		return parseBoolDefault(column.ColumnDefault)
	case "int":
		return parseIntDefault(column.ColumnDefault, defaultIntegerHint(column.ColumnName))
	case "int64":
		return parseIntDefault(column.ColumnDefault, defaultIntegerHint(column.ColumnName))
	case "json":
		return "{}"
	case "time":
		return ""
	default:
		return ""
	}
}

func parseBoolDefault(raw string) bool {
	return strings.Contains(strings.ToLower(raw), "true")
}

func parseIntDefault(raw string, fallback int) int {
	trimmed := strings.TrimSpace(raw)
	trimmed = strings.Trim(trimmed, "()")
	if value, err := strconv.Atoi(trimmed); err == nil {
		return value
	}
	return fallback
}

func defaultIntegerHint(columnName string) int {
	lower := strings.ToLower(strings.TrimSpace(columnName))
	if strings.Contains(lower, "status") {
		return 1
	}
	return 0
}

func buildDefaultFormState(fields []templateField) map[string]any {
	result := map[string]any{
		"id": 0,
	}
	for _, field := range fields {
		if !field.IsSaveField {
			continue
		}
		result[field.ColumnName] = field.DefaultValue
	}
	return result
}

func buildDefaultSearchState(fields []SchemaField) map[string]any {
	result := make(map[string]any, len(fields)*2)
	for _, field := range fields {
		if field.Component == "datetime-range" {
			result[field.Field+"_start"] = ""
			result[field.Field+"_end"] = ""
			continue
		}
		result[field.Field] = ""
	}
	return result
}

func buildListColumns(fields []SchemaField) []map[string]any {
	result := make([]map[string]any, 0, len(fields)+1)
	for _, field := range fields {
		column := map[string]any{
			"key":   field.Field,
			"title": field.Label,
		}
		switch field.Field {
		case "id":
			column["width"] = "80px"
		case "created_at", "updated_at":
			column["width"] = "180px"
		}
		result = append(result, column)
	}
	result = append(result, map[string]any{
		"key":   "actions",
		"title": "操作",
		"width": "220px",
		"align": "right",
	})
	return result
}

func applyWriteResult(result GenerateResult, path string, status string, warning string) GenerateResult {
	switch status {
	case "created":
		result.GeneratedFiles = append(result.GeneratedFiles, path)
	case "overwritten":
		result.OverwrittenFiles = append(result.OverwrittenFiles, path)
	case "skipped":
		result.SkippedFiles = append(result.SkippedFiles, path)
	}
	if warning != "" {
		result.Warnings = append(result.Warnings, warning)
	}
	return result
}

func anyField(fields []templateField, fn func(templateField) bool) bool {
	for _, field := range fields {
		if fn(field) {
			return true
		}
	}
	return false
}

func selectFields(fields []templateField, fn func(templateField) bool) []templateField {
	result := make([]templateField, 0, len(fields))
	for _, field := range fields {
		if fn(field) {
			result = append(result, field)
		}
	}
	return result
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func mustMarshalPayload(payload PayloadConfig) json.RawMessage {
	raw, _ := json.Marshal(payload)
	return raw
}

func (s GeneratorService) upsertMenus(meta ModuleMeta) (MenuUpsertResult, []string, error) {
	if s.DB == nil {
		return MenuUpsertResult{}, []string{"当前运行时没有数据库连接，未执行菜单写入。"}, nil
	}

	var result MenuUpsertResult
	warnings := []string{}

	err := s.DB.Transaction(func(tx *gorm.DB) error {
		if err := syncPrimarySequence(tx, "admin_menu"); err != nil {
			return err
		}
		if err := syncPrimarySequence(tx, "admin_role_menu"); err != nil {
			return err
		}

		parentID, err := ensureSystemParentMenu(tx)
		if err != nil {
			return err
		}

		menuRecord, err := upsertAdminMenuRecord(tx, model.AdminMenu{
			ParentID:  parentID,
			Name:      ToKebab(meta.ModuleName),
			Title:     meta.Title,
			Path:      meta.RoutePath,
			Component: "system/" + ToKebab(meta.ModuleName) + "/index",
			MenuType:  model.MenuTypeMenu,
			Icon:      "file",
			Sort:      200,
			Visible:   true,
			Status:    1,
		}, "path", meta.RoutePath)
		if err != nil {
			return err
		}
		result.Records = append(result.Records, menuRecord)

		buttons := []model.AdminMenu{
			{
				ParentID:       int64(menuRecord["id"].(int64)),
				Name:           ToKebab(meta.ModuleName) + "-list",
				Title:          meta.Title + "列表",
				MenuType:       model.MenuTypeButton,
				PermissionCode: meta.PermissionCodes[0],
				Sort:           201,
				Visible:        true,
				Status:         1,
			},
			{
				ParentID:       int64(menuRecord["id"].(int64)),
				Name:           ToKebab(meta.ModuleName) + "-save",
				Title:          meta.Title + "保存",
				MenuType:       model.MenuTypeButton,
				PermissionCode: meta.PermissionCodes[1],
				Sort:           202,
				Visible:        true,
				Status:         1,
			},
			{
				ParentID:       int64(menuRecord["id"].(int64)),
				Name:           ToKebab(meta.ModuleName) + "-delete",
				Title:          meta.Title + "删除",
				MenuType:       model.MenuTypeButton,
				PermissionCode: meta.PermissionCodes[2],
				Sort:           203,
				Visible:        true,
				Status:         1,
			},
		}

		menuIDs := []int64{menuRecord["id"].(int64)}
		for _, button := range buttons {
			record, err := upsertAdminMenuRecord(tx, button, "permission_code", button.PermissionCode)
			if err != nil {
				return err
			}
			result.Records = append(result.Records, record)
			menuIDs = append(menuIDs, record["id"].(int64))
		}

		for _, menuID := range menuIDs {
			link := model.AdminRoleMenu{RoleID: 1, MenuID: menuID}
			if err := tx.Where("role_id = ? AND menu_id = ?", 1, menuID).FirstOrCreate(&link).Error; err != nil {
				return err
			}
		}
		warnings = append(warnings, "已写入菜单和权限点；新生成模块需要重启后端服务后才能注册生效。")
		return nil
	})

	return result, warnings, err
}

func ensureSystemParentMenu(tx *gorm.DB) (int64, error) {
	var systemMenu model.AdminMenu
	err := tx.Where("path = ? AND menu_type = ?", "/system", model.MenuTypeMenu).First(&systemMenu).Error
	if err == nil {
		return systemMenu.ID, nil
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, err
	}

	systemMenu = model.AdminMenu{
		ParentID:  0,
		Name:      "system",
		Title:     "系统管理",
		Path:      "/system",
		Component: "layout",
		MenuType:  model.MenuTypeMenu,
		Icon:      "setting",
		Sort:      10,
		Visible:   true,
		Status:    1,
	}
	if err := tx.Create(&systemMenu).Error; err != nil {
		return 0, err
	}
	return systemMenu.ID, nil
}

func syncPrimarySequence(tx *gorm.DB, tableName string) error {
	statement := fmt.Sprintf(
		"SELECT setval(pg_get_serial_sequence('%s', 'id'), COALESCE((SELECT MAX(id) FROM %s), 1), true)",
		tableName,
		tableName,
	)
	return tx.Exec(statement).Error
}

func upsertAdminMenuRecord(tx *gorm.DB, payload model.AdminMenu, lookupField string, lookupValue any) (map[string]any, error) {
	var row model.AdminMenu
	err := tx.Where(lookupField+" = ?", lookupValue).First(&row).Error
	switch {
	case err == nil:
		if err := tx.Model(&row).Updates(map[string]any{
			"parent_id":       payload.ParentID,
			"name":            payload.Name,
			"title":           payload.Title,
			"path":            payload.Path,
			"component":       payload.Component,
			"menu_type":       payload.MenuType,
			"permission_code": payload.PermissionCode,
			"icon":            payload.Icon,
			"sort":            payload.Sort,
			"visible":         payload.Visible,
			"status":          payload.Status,
		}).Error; err != nil {
			return nil, err
		}
	case errors.Is(err, gorm.ErrRecordNotFound):
		row = payload
		if err := tx.Create(&row).Error; err != nil {
			return nil, err
		}
	default:
		return nil, err
	}

	return map[string]any{
		"id":              row.ID,
		"name":            payload.Name,
		"title":           payload.Title,
		"path":            payload.Path,
		"menu_type":       payload.MenuType,
		"permission_code": payload.PermissionCode,
	}, nil
}
