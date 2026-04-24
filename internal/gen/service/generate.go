package service

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/lvjiaben/goweb-scaffold/internal/gen/registry"
	"github.com/lvjiaben/goweb-scaffold/internal/gen/writer"
	"gorm.io/gorm"
)

type templateField struct {
	ColumnName      string
	GoFieldName     string
	GoType          string
	GormTag         string
	RequestType     string
	RequestKind     string
	FormTSType      string
	SearchTSType    string
	TSType          string
	Component       string
	SearchComponent string
	Display         string
	SearchOperator  string
	StartQueryKey   string
	EndQueryKey     string
	Label           string
	Width           string
	Options         []FieldOption
	Placeholder     string
	IsPrimaryKey    bool
	IsNullable      bool
	IsListField     bool
	IsSaveField     bool
	IsSearchField   bool
	Sortable        bool
	Searchable      bool
	IsBoolean       bool
	IsInteger       bool
	IsBigInteger    bool
	IsTimestamp     bool
	IsJSON          bool
	Required        bool
	Readonly        bool
	Hidden          bool
	DefaultValue    any
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
	Fields           []templateField
	SaveFields       []templateField
	SearchFields     []templateField
	LikeFields       []templateField
	ExactFields      []templateField
	RangeFields      []templateField
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

type frontendDataTemplateData struct {
	ModuleName   string
	PascalName   string
	Title        string
	FormFields   []templateField
	ListFields   []templateField
	SearchFields []templateField
}

type frontendListTemplateData struct {
	ModuleName       string
	PascalName       string
	Title            string
	ListFields       []templateField
	PermissionSave   string
	PermissionDelete string
}

type frontendFormTemplateData struct {
	ModuleName string
	PascalName string
	Title      string
	FormFields []templateField
}

type generatedArtifact struct {
	Path    string
	Content []byte
}

type generationBundle struct {
	Meta        ModuleMeta
	Preview     Preview
	Artifacts   []generatedArtifact
	Lock        LockFile
	Generated   []string
	RegistryRef registry.GeneratedModule
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

	bundle, err := s.prepareBundle(input)
	if err != nil {
		return result, err
	}

	w := writer.New(s.RepoRoot)
	for _, item := range bundle.Artifacts {
		status, warning, err := w.Write(item.Path, item.Content, input.Overwrite)
		if err != nil {
			return result, err
		}
		result = applyWriteResult(result, item.Path, status, warning)
	}

	if input.UpsertMenu {
		if !input.RegisterModule {
			result.Warnings = append(result.Warnings, "register_module=false 且 upsert_menu=true，菜单已写入，但重启前端/后端前路由不会生效。")
		}
		menuResult, warnings, err := s.upsertMenus(bundle.Meta)
		if err != nil {
			return result, err
		}
		result.MenuRecords = menuResult.Records
		result.Warnings = append(result.Warnings, warnings...)
	} else {
		result.Warnings = append(result.Warnings, "upsert_menu=false，未写入 admin_menu 和 admin_role_menu。")
	}

	result.ModuleName = bundle.Meta.ModuleName
	result.RoutePath = bundle.Meta.RoutePath
	result.PermissionCodes = append([]string{}, bundle.Meta.PermissionCodes...)
	result.Warnings = uniqueStrings(result.Warnings)
	return result, nil
}

func (s GeneratorService) Diff(input GenerateInput) (DiffResult, error) {
	result := DiffResult{
		WouldCreateFiles:    []string{},
		WouldOverwriteFiles: []string{},
		WouldSkipFiles:      []string{},
		PerFileDiffSummary:  []DiffFileSummary{},
		PermissionCodes:     []string{},
		Warnings:            []string{},
	}

	bundle, err := s.prepareBundle(input)
	if err != nil {
		return result, err
	}

	w := writer.New(s.RepoRoot)
	for _, item := range bundle.Artifacts {
		status, warning, existing, err := w.Inspect(item.Path, item.Content, input.Overwrite)
		if err != nil {
			return result, err
		}

		summary := DiffFileSummary{
			Path:            item.Path,
			Status:          diffStatus(status),
			ChangedSections: summarizeFileDiff(status, warning, existing, item.Content),
			OldHash:         hashContent(existing),
			NewHash:         hashContent(item.Content),
		}
		result.PerFileDiffSummary = append(result.PerFileDiffSummary, summary)

		switch status {
		case "generated":
			result.WouldCreateFiles = append(result.WouldCreateFiles, item.Path)
		case "overwritten":
			result.WouldOverwriteFiles = append(result.WouldOverwriteFiles, item.Path)
		default:
			result.WouldSkipFiles = append(result.WouldSkipFiles, item.Path)
		}
		if warning != "" {
			result.Warnings = append(result.Warnings, warning)
		}
	}

	result.ModuleName = bundle.Meta.ModuleName
	result.RoutePath = bundle.Meta.RoutePath
	result.PermissionCodes = append([]string{}, bundle.Meta.PermissionCodes...)
	result.Warnings = uniqueStrings(result.Warnings)
	return result, nil
}

func (s GeneratorService) prepareBundle(input GenerateInput) (generationBundle, error) {
	bundle := generationBundle{}
	moduleName := ToSnake(strings.TrimSpace(input.ModuleName))
	tableName := strings.TrimSpace(input.TableName)
	if moduleName == "" || tableName == "" {
		return bundle, fmt.Errorf("module_name and table_name are required")
	}
	if s.RepoRoot == "" {
		return bundle, fmt.Errorf("repo root is required")
	}
	if len(input.Columns) == 0 {
		return bundle, fmt.Errorf("columns are required for generation")
	}

	preview := input.Preview
	if preview.ModuleName == "" {
		preview = BuildPreview(moduleName, tableName, mustMarshalPayload(input.Payload), input.Columns)
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
		Title:       firstNonEmpty(strings.TrimSpace(preview.Payload.Title), strings.TrimSpace(preview.TableComment), HumanizeModuleName(moduleName)),
		ViewFile:    preview.Page.ViewFile,
		PermissionCodes: []string{
			moduleName + ".list",
			moduleName + ".save",
			moduleName + ".delete",
		},
	}
	bundle.Meta = meta
	bundle.Preview = preview
	bundle.RegistryRef = registry.GeneratedModule{
		ModuleName: meta.ModuleName,
		RoutePath:  meta.RoutePath,
		PageName:   meta.PageName,
		Title:      meta.Title,
		ViewFile:   meta.ViewFile,
	}

	basicGeneratedFiles := []string{
		paths["module"],
		paths["handler"],
		paths["service"],
		paths["repo"],
		paths["dto"],
		paths["model"],
		paths["meta"],
		paths["lock"],
		paths["api"],
		paths["view_data"],
		paths["view_list"],
		paths["view_form"],
	}
	if input.RegisterModule {
		basicGeneratedFiles = append(basicGeneratedFiles, "internal/gen/modules_gen.go", paths["route_module"])
	}
	bundle.Generated = append([]string{}, basicGeneratedFiles...)

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
		Fields:           fields,
		SaveFields:       selectFields(fields, func(item templateField) bool { return item.IsSaveField }),
		SearchFields:     selectFields(fields, func(item templateField) bool { return item.IsSearchField }),
		LikeFields:       selectFields(fields, func(item templateField) bool { return item.IsSearchField && item.SearchOperator == "like" }),
		ExactFields:      selectFields(fields, func(item templateField) bool { return item.IsSearchField && item.SearchOperator == "eq" }),
		RangeFields:      selectFields(fields, func(item templateField) bool { return item.IsSearchField && item.SearchOperator == "between" }),
		RequiredFields: selectFields(fields, func(item templateField) bool {
			return item.IsSaveField && item.Required && (item.RequestKind == "string" || item.RequestKind == "time" || item.RequestKind == "json")
		}),
		UsesStrings: true,
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
	dataData := frontendDataTemplateData{
		ModuleName:   moduleName,
		PascalName:   meta.PascalName,
		Title:        meta.Title,
		FormFields:   selectFields(fields, func(item templateField) bool { return item.IsSaveField }),
		ListFields:   selectFields(fields, func(item templateField) bool { return item.IsListField }),
		SearchFields: selectFields(fields, func(item templateField) bool { return item.IsSearchField }),
	}
	listData := frontendListTemplateData{
		ModuleName:       moduleName,
		PascalName:       meta.PascalName,
		Title:            meta.Title,
		ListFields:       selectFields(fields, func(item templateField) bool { return item.IsListField }),
		PermissionSave:   meta.PermissionCodes[1],
		PermissionDelete: meta.PermissionCodes[2],
	}
	formData := frontendFormTemplateData{
		ModuleName: moduleName,
		PascalName: meta.PascalName,
		Title:      meta.Title,
		FormFields: selectFields(fields, func(item templateField) bool { return item.IsSaveField }),
	}

	artifacts := []generatedArtifact{}
	appendTemplateArtifact := func(path string, templatePath string, data any) error {
		content, err := renderTemplate(templatePath, data)
		if err != nil {
			return err
		}
		artifacts = append(artifacts, generatedArtifact{Path: path, Content: content})
		return nil
	}

	if err := appendTemplateArtifact(paths["module"], "backend/module.go.tmpl", moduleData); err != nil {
		return bundle, err
	}
	if err := appendTemplateArtifact(paths["backend_handler"], "backend/backend_handler.go.tmpl", moduleData); err != nil {
		return bundle, err
	}
	if err := appendTemplateArtifact(paths["api_handler"], "backend/api_handler.go.tmpl", moduleData); err != nil {
		return bundle, err
	}
	if err := appendTemplateArtifact(paths["service"], "backend/service.go.tmpl", moduleData); err != nil {
		return bundle, err
	}
	if err := appendTemplateArtifact(paths["repo"], "backend/repo.go.tmpl", moduleData); err != nil {
		return bundle, err
	}
	if err := appendTemplateArtifact(paths["backend_dto"], "backend/backend_dto.go.tmpl", typesData); err != nil {
		return bundle, err
	}
	if err := appendTemplateArtifact(paths["api_dto"], "backend/api_dto.go.tmpl", typesData); err != nil {
		return bundle, err
	}
	if err := appendTemplateArtifact(paths["model"], "backend/model.go.tmpl", modelData); err != nil {
		return bundle, err
	}
	if err := appendTemplateArtifact(paths["meta"], "backend/meta.go.tmpl", meta); err != nil {
		return bundle, err
	}
	if err := appendTemplateArtifact(paths["api"], "admin_frontend/api.ts.tmpl", apiData); err != nil {
		return bundle, err
	}
	if err := appendTemplateArtifact(paths["view_data"], "admin_frontend/data.ts.tmpl", dataData); err != nil {
		return bundle, err
	}
	if err := appendTemplateArtifact(paths["view_list"], "admin_frontend/list.vue.tmpl", listData); err != nil {
		return bundle, err
	}
	if err := appendTemplateArtifact(paths["view_form"], "admin_frontend/form-drawer.vue.tmpl", formData); err != nil {
		return bundle, err
	}

	if input.RegisterModule {
		content, err := registry.RenderBackendModulesFile(s.RepoRoot, bundle.RegistryRef)
		if err != nil {
			return bundle, err
		}
		artifacts = append(artifacts, generatedArtifact{Path: "internal/gen/modules_gen.go", Content: content})

		routeContent, err := registry.RenderFrontendRouteModule(s.RepoRoot, bundle.RegistryRef)
		if err != nil {
			return bundle, err
		}
		artifacts = append(artifacts, generatedArtifact{Path: paths["route_module"], Content: routeContent})
	}

	lockFile, lockContent, err := s.buildLockFile(paths["lock"], meta, preview, preview.Payload, basicGeneratedFiles, artifacts, input.GeneratedAt)
	if err != nil {
		return bundle, err
	}
	bundle.Lock = lockFile
	artifacts = append(artifacts, generatedArtifact{Path: paths["lock"], Content: lockContent})

	bundle.Artifacts = artifacts
	return bundle, nil
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
			DefaultValue:  guessDefaultValue(column, guessFormComponent(column)),
		}

		if inferred, ok := inferredIndex[column.ColumnName]; ok {
			item.Label = inferred.GuessedLabel
			item.Component = inferred.GuessedFormComponent
			item.Display = inferred.GuessedListDisplay
			item.Searchable = inferred.GuessedSearchable
			item.Sortable = inferred.GuessedSortable
		}
		if schema, ok := formIndex[column.ColumnName]; ok {
			item.IsSaveField = true
			if schema.Label != "" {
				item.Label = schema.Label
			}
			item.Component = schema.Component
			item.Required = schema.Required
			item.Readonly = schema.Readonly
			item.Hidden = schema.Hidden
			if strings.TrimSpace(schema.Width) != "" {
				item.Width = schema.Width
			}
			item.Options = append([]FieldOption{}, schema.Options...)
			item.Placeholder = schema.Placeholder
			item.DefaultValue = firstNonNil(schema.DefaultValue, item.DefaultValue)
		}
		if schema, ok := listIndex[column.ColumnName]; ok {
			item.IsListField = true
			if schema.Label != "" {
				item.Label = schema.Label
			}
			if schema.Display != "" {
				item.Display = schema.Display
			}
			item.Sortable = schema.Sortable
			item.Searchable = schema.Searchable
			if strings.TrimSpace(schema.Width) != "" {
				item.Width = schema.Width
			}
			if len(schema.Options) > 0 {
				item.Options = append([]FieldOption{}, schema.Options...)
			}
		}
		if schema, ok := searchIndex[column.ColumnName]; ok {
			item.IsSearchField = true
			if schema.Label != "" {
				item.Label = schema.Label
			}
			if schema.Component != "" {
				item.SearchComponent = schema.Component
			}
			item.SearchOperator = schema.Operator
			item.Searchable = schema.Searchable
			if strings.TrimSpace(schema.Width) != "" {
				item.Width = schema.Width
			}
			if len(schema.Options) > 0 {
				item.Options = append([]FieldOption{}, schema.Options...)
			}
			if schema.Placeholder != "" {
				item.Placeholder = schema.Placeholder
			}
		}
		if item.Label == "" {
			item.Label = preferredLabel(column)
		}
		if item.SearchComponent == "" {
			item.SearchComponent = guessSearchComponent(column, item.Component, item.Options)
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
		return "number"
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

func (s GeneratorService) buildLockFile(lockPath string, meta ModuleMeta, preview Preview, payload PayloadConfig, generatedFiles []string, artifacts []generatedArtifact, generatedAt time.Time) (LockFile, []byte, error) {
	lock := LockFile{
		GeneratedBy:     GeneratorName,
		ModuleName:      meta.ModuleName,
		TableName:       meta.TableName,
		TemplateVersion: TemplateVersion,
		Payload:         payload,
		PreviewSummary: LockPreviewSummary{
			TableComment:   preview.TableComment,
			Page:           preview.Page,
			API:            preview.API,
			InferredFields: preview.InferredFields,
			FormSchema:     preview.FormSchema,
			ListSchema:     preview.ListSchema,
			SearchSchema:   preview.SearchSchema,
		},
		Snapshot:        buildCurrentSnapshot(meta.ModuleName, meta.TableName, payload, preview, artifacts),
		PermissionCodes: append([]string{}, meta.PermissionCodes...),
		RoutePath:       meta.RoutePath,
		GeneratedFiles:  append([]string{}, generatedFiles...),
	}

	lock.GeneratedAt = s.resolveGeneratedAt(lockPath, lock, generatedAt)
	content, err := json.MarshalIndent(lock, "", "  ")
	if err != nil {
		return LockFile{}, nil, err
	}
	content = append(content, '\n')
	return lock, content, nil
}

func (s GeneratorService) resolveGeneratedAt(lockPath string, desired LockFile, override time.Time) string {
	if !override.IsZero() {
		return override.Format(time.RFC3339)
	}

	existing, err := s.readLockFile(lockPath)
	if err == nil && sameLockCore(existing, desired) {
		return existing.GeneratedAt
	}
	return time.Now().Format(time.RFC3339)
}

func (s GeneratorService) readLockFile(relPath string) (LockFile, error) {
	fullPath := filepath.Join(s.RepoRoot, filepath.Clean(relPath))
	raw, err := os.ReadFile(fullPath)
	if err != nil {
		return LockFile{}, err
	}
	var lock LockFile
	if err := json.Unmarshal(raw, &lock); err != nil {
		return LockFile{}, err
	}
	return lock, nil
}

func sameLockCore(left LockFile, right LockFile) bool {
	left.GeneratedAt = ""
	right.GeneratedAt = ""
	leftRaw, _ := json.Marshal(left)
	rightRaw, _ := json.Marshal(right)
	return bytes.Equal(leftRaw, rightRaw)
}

func diffStatus(status string) string {
	switch status {
	case "generated":
		return "create"
	case "overwritten":
		return "overwrite"
	default:
		return "skip"
	}
}

func summarizeFileDiff(status string, warning string, oldContent []byte, newContent []byte) []string {
	switch status {
	case "generated":
		return []string{"new file"}
	case "overwritten":
		return summarizeChangedLines(oldContent, newContent)
	default:
		if warning != "" {
			return []string{warning}
		}
		if bytes.Equal(oldContent, newContent) {
			return []string{"no content changes"}
		}
		return summarizeChangedLines(oldContent, newContent)
	}
}

func summarizeChangedLines(oldContent []byte, newContent []byte) []string {
	oldLines := strings.Split(strings.ReplaceAll(string(oldContent), "\r\n", "\n"), "\n")
	newLines := strings.Split(strings.ReplaceAll(string(newContent), "\r\n", "\n"), "\n")

	prefix := 0
	for prefix < len(oldLines) && prefix < len(newLines) && oldLines[prefix] == newLines[prefix] {
		prefix++
	}

	suffix := 0
	for suffix < len(oldLines)-prefix && suffix < len(newLines)-prefix &&
		oldLines[len(oldLines)-1-suffix] == newLines[len(newLines)-1-suffix] {
		suffix++
	}

	oldStart := prefix + 1
	oldEnd := len(oldLines) - suffix
	newStart := prefix + 1
	newEnd := len(newLines) - suffix

	if oldEnd < oldStart {
		oldEnd = oldStart
	}
	if newEnd < newStart {
		newEnd = newStart
	}

	return []string{
		fmt.Sprintf("old lines %d-%d -> new lines %d-%d", oldStart, oldEnd, newStart, newEnd),
	}
}

func hashContent(content []byte) string {
	if len(content) == 0 {
		return ""
	}
	sum := sha256.Sum256(content)
	return hex.EncodeToString(sum[:8])
}

func applyWriteResult(result GenerateResult, path string, status string, warning string) GenerateResult {
	switch status {
	case "generated":
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
	if strings.Contains(lower, "status") || strings.Contains(lower, "state") {
		return 1
	}
	return 0
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

		menuRecord, err := upsertAdminMenuRecord(tx, AdminMenu{
			ParentID:  parentID,
			Name:      ToKebab(meta.ModuleName),
			EnName:    ToPascal(meta.ModuleName),
			Title:     meta.Title,
			Path:      meta.RoutePath,
			Component: meta.ModuleName + "/list",
			MenuType:  MenuTypeMenu,
			Icon:      "lucide:file-text",
			Sort:      200,
			Visible:   true,
			Status:    1,
		}, "path", meta.RoutePath)
		if err != nil {
			return err
		}
		result.Records = append(result.Records, menuRecord)

		menuID, _ := menuRecord["id"].(int64)
		buttons := []AdminMenu{
			{
				ParentID:       menuID,
				Name:           ToKebab(meta.ModuleName) + "-list",
				EnName:         ToPascal(meta.ModuleName) + "List",
				Title:          meta.Title + "列表",
				MenuType:       MenuTypeButton,
				PermissionCode: meta.PermissionCodes[0],
				Sort:           201,
				Visible:        true,
				Status:         1,
			},
			{
				ParentID:       menuID,
				Name:           ToKebab(meta.ModuleName) + "-save",
				EnName:         ToPascal(meta.ModuleName) + "Save",
				Title:          meta.Title + "保存",
				MenuType:       MenuTypeButton,
				PermissionCode: meta.PermissionCodes[1],
				Sort:           202,
				Visible:        true,
				Status:         1,
			},
			{
				ParentID:       menuID,
				Name:           ToKebab(meta.ModuleName) + "-delete",
				EnName:         ToPascal(meta.ModuleName) + "Delete",
				Title:          meta.Title + "删除",
				MenuType:       MenuTypeButton,
				PermissionCode: meta.PermissionCodes[2],
				Sort:           203,
				Visible:        true,
				Status:         1,
			},
		}

		menuIDs := []int64{menuID}
		for _, button := range buttons {
			record, err := upsertAdminMenuRecord(tx, button, "permission_code", button.PermissionCode)
			if err != nil {
				return err
			}
			result.Records = append(result.Records, record)
			if id, ok := record["id"].(int64); ok {
				menuIDs = append(menuIDs, id)
			}
		}

		for _, currentMenuID := range menuIDs {
			link := AdminRoleMenu{RoleID: 1, MenuID: currentMenuID}
			if err := tx.Where("role_id = ? AND menu_id = ?", 1, currentMenuID).FirstOrCreate(&link).Error; err != nil {
				return err
			}
		}
		warnings = append(warnings, "已写入菜单和权限点；新生成模块需要重启后端服务后才能注册生效。")
		return nil
	})

	return result, warnings, err
}

func ensureSystemParentMenu(tx *gorm.DB) (int64, error) {
	var systemMenu AdminMenu
	err := tx.Where("path = ? AND menu_type = ?", "/system", MenuTypeMenu).First(&systemMenu).Error
	if err == nil {
		return systemMenu.ID, nil
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, err
	}

	systemMenu = AdminMenu{
		ParentID:  0,
		Name:      "system",
		EnName:    "System",
		Title:     "系统管理",
		Path:      "/system",
		Component: "layout",
		MenuType:  MenuTypeMenu,
		Icon:      "lucide:settings",
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

func upsertAdminMenuRecord(tx *gorm.DB, payload AdminMenu, lookupField string, lookupValue any) (map[string]any, error) {
	var row AdminMenu
	err := tx.Where(lookupField+" = ?", lookupValue).First(&row).Error
	switch {
	case err == nil:
		if err := tx.Model(&row).Updates(map[string]any{
			"parent_id":       payload.ParentID,
			"name":            payload.Name,
			"enname":          payload.EnName,
			"title":           payload.Title,
			"path":            payload.Path,
			"component":       payload.Component,
			"menu_type":       payload.MenuType,
			"permission_code": payload.PermissionCode,
			"iframe":          payload.Iframe,
			"external":        payload.External,
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
		"enname":          payload.EnName,
		"title":           payload.Title,
		"path":            payload.Path,
		"menu_type":       payload.MenuType,
		"permission_code": payload.PermissionCode,
	}, nil
}
