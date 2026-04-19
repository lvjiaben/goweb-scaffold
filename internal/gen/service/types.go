package service

import (
	"time"

	"gorm.io/gorm"
)

const (
	TemplateVersion = "v5"
	GeneratorName   = "goweb-scaffold"
)

type ColumnInfo struct {
	ColumnName    string `json:"column_name"`
	DataType      string `json:"data_type"`
	IsNullable    bool   `json:"is_nullable"`
	ColumnDefault string `json:"column_default"`
	OrdinalPos    int    `json:"ordinal_position"`
	IsPrimaryKey  bool   `json:"is_primary_key"`
	ColumnComment string `json:"column_comment,omitempty"`
	TableComment  string `json:"table_comment,omitempty"`
}

type FieldOption struct {
	Label string `json:"label"`
	Value any    `json:"value"`
}

type FieldOverride struct {
	Label        string        `json:"label,omitempty"`
	Component    string        `json:"component,omitempty"`
	Placeholder  string        `json:"placeholder,omitempty"`
	Required     *bool         `json:"required,omitempty"`
	Readonly     *bool         `json:"readonly,omitempty"`
	Hidden       *bool         `json:"hidden,omitempty"`
	Sortable     *bool         `json:"sortable,omitempty"`
	Searchable   *bool         `json:"searchable,omitempty"`
	Width        string        `json:"width,omitempty"`
	Options      []FieldOption `json:"options,omitempty"`
	DefaultValue any           `json:"default_value,omitempty"`
}

type InferredField struct {
	ColumnName           string `json:"column_name"`
	DataType             string `json:"data_type"`
	IsNullable           bool   `json:"is_nullable"`
	IsPrimaryKey         bool   `json:"is_primary_key"`
	ColumnComment        string `json:"column_comment,omitempty"`
	GuessedLabel         string `json:"guessed_label"`
	GuessedFormComponent string `json:"guessed_form_component"`
	GuessedListDisplay   string `json:"guessed_list_display"`
	GuessedSearchable    bool   `json:"guessed_searchable"`
	GuessedSortable      bool   `json:"guessed_sortable"`
}

type SchemaField struct {
	Field        string        `json:"field"`
	Label        string        `json:"label"`
	Component    string        `json:"component"`
	Display      string        `json:"display,omitempty"`
	Operator     string        `json:"operator,omitempty"`
	Required     bool          `json:"required,omitempty"`
	Readonly     bool          `json:"readonly,omitempty"`
	Hidden       bool          `json:"hidden,omitempty"`
	Searchable   bool          `json:"searchable,omitempty"`
	Sortable     bool          `json:"sortable,omitempty"`
	Placeholder  string        `json:"placeholder,omitempty"`
	Width        string        `json:"width,omitempty"`
	Options      []FieldOption `json:"options,omitempty"`
	DefaultValue any           `json:"default_value,omitempty"`
}

type PayloadConfig struct {
	ListFields     []string                 `json:"list_fields"`
	FormFields     []string                 `json:"form_fields"`
	SearchFields   []string                 `json:"search_fields"`
	Title          string                   `json:"title,omitempty"`
	FieldOverrides map[string]FieldOverride `json:"field_overrides,omitempty"`
}

type PageMeta struct {
	RoutePath string `json:"route_path"`
	PageName  string `json:"page_name"`
	ViewFile  string `json:"view_file"`
}

type APIMeta struct {
	ModuleCode string `json:"module_code"`
	List       string `json:"list"`
	Detail     string `json:"detail"`
	Save       string `json:"save"`
	Delete     string `json:"delete"`
}

type Preview struct {
	ModuleName     string          `json:"module_name"`
	TableName      string          `json:"table_name"`
	TableComment   string          `json:"table_comment,omitempty"`
	Page           PageMeta        `json:"page"`
	API            APIMeta         `json:"api"`
	InferredFields []InferredField `json:"inferred_fields"`
	FormSchema     []SchemaField   `json:"form_schema"`
	ListSchema     []SchemaField   `json:"list_schema"`
	SearchSchema   []SchemaField   `json:"search_schema"`
	Payload        PayloadConfig   `json:"payload"`
	Notes          []string        `json:"notes"`
}

type GenerateInput struct {
	ModuleName     string
	TableName      string
	Payload        PayloadConfig
	Preview        Preview
	Columns        []ColumnInfo
	Overwrite      bool
	RegisterModule bool
	UpsertMenu     bool
	GeneratedAt    time.Time
}

type GenerateResult struct {
	GeneratedFiles   []string         `json:"generated_files"`
	OverwrittenFiles []string         `json:"overwritten_files"`
	SkippedFiles     []string         `json:"skipped_files"`
	ModuleName       string           `json:"module_name"`
	RoutePath        string           `json:"route_path"`
	PermissionCodes  []string         `json:"permission_codes"`
	MenuRecords      []map[string]any `json:"menu_records"`
	Warnings         []string         `json:"warnings"`
}

type DiffFileSummary struct {
	Path            string   `json:"path"`
	Status          string   `json:"status"`
	ChangedSections []string `json:"changed_sections"`
	OldHash         string   `json:"old_hash,omitempty"`
	NewHash         string   `json:"new_hash,omitempty"`
}

type DiffResult struct {
	WouldCreateFiles    []string          `json:"would_create_files"`
	WouldOverwriteFiles []string          `json:"would_overwrite_files"`
	WouldSkipFiles      []string          `json:"would_skip_files"`
	PerFileDiffSummary  []DiffFileSummary `json:"per_file_diff_summary"`
	ModuleName          string            `json:"module_name"`
	RoutePath           string            `json:"route_path"`
	PermissionCodes     []string          `json:"permission_codes"`
	Warnings            []string          `json:"warnings"`
}

type ManagedModule struct {
	ModuleName      string             `json:"module_name"`
	TableName       string             `json:"table_name"`
	GeneratedAt     string             `json:"generated_at"`
	TemplateVersion string             `json:"template_version"`
	RoutePath       string             `json:"route_path"`
	PermissionCodes []string           `json:"permission_codes"`
	Files           []string           `json:"files"`
	Payload         PayloadConfig      `json:"payload"`
	PreviewSummary  LockPreviewSummary `json:"preview_summary"`
}

type RemoveInput struct {
	ModuleName       string
	RemoveFiles      bool
	UnregisterModule bool
	RemoveMenu       bool
	RemoveHistory    bool
	RemoveLock       bool
}

type RemoveResult struct {
	ModuleName               string           `json:"module_name"`
	RemovedFiles             []string         `json:"removed_files"`
	SkippedFiles             []string         `json:"skipped_files"`
	RemovedMenuRecords       []map[string]any `json:"removed_menu_records"`
	RemovedRoleMenuLinks     int64            `json:"removed_role_menu_links"`
	RemovedHistoryIDs        []int64          `json:"removed_history_ids"`
	RegeneratedRegistryFiles []string         `json:"regenerated_registry_files"`
	Warnings                 []string         `json:"warnings"`
}

type MenuUpsertResult struct {
	Records []map[string]any
}

type ModuleMeta struct {
	ModuleName      string
	TableName       string
	PackageName     string
	PascalName      string
	RoutePath       string
	PageName        string
	Title           string
	ViewFile        string
	PermissionCodes []string
}

type LockPreviewSummary struct {
	TableComment   string          `json:"table_comment,omitempty"`
	Page           PageMeta        `json:"page"`
	API            APIMeta         `json:"api"`
	InferredFields []InferredField `json:"inferred_fields"`
	FormSchema     []SchemaField   `json:"form_schema"`
	ListSchema     []SchemaField   `json:"list_schema"`
	SearchSchema   []SchemaField   `json:"search_schema"`
}

type LockFile struct {
	GeneratedBy     string             `json:"generated_by"`
	ModuleName      string             `json:"module_name"`
	TableName       string             `json:"table_name"`
	GeneratedAt     string             `json:"generated_at"`
	TemplateVersion string             `json:"template_version"`
	Payload         PayloadConfig      `json:"payload"`
	PreviewSummary  LockPreviewSummary `json:"preview_summary"`
	PermissionCodes []string           `json:"permission_codes"`
	RoutePath       string             `json:"route_path"`
	GeneratedFiles  []string           `json:"generated_files"`
}

type GeneratorService struct {
	RepoRoot string
	DB       *gorm.DB
}
