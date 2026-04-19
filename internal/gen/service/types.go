package service

import "gorm.io/gorm"

type ColumnInfo struct {
	ColumnName    string `json:"column_name"`
	DataType      string `json:"data_type"`
	IsNullable    bool   `json:"is_nullable"`
	ColumnDefault string `json:"column_default"`
	OrdinalPos    int    `json:"ordinal_position"`
	IsPrimaryKey  bool   `json:"is_primary_key"`
}

type InferredField struct {
	ColumnName           string `json:"column_name"`
	DataType             string `json:"data_type"`
	IsNullable           bool   `json:"is_nullable"`
	IsPrimaryKey         bool   `json:"is_primary_key"`
	GuessedFormComponent string `json:"guessed_form_component"`
	GuessedListDisplay   string `json:"guessed_list_display"`
	GuessedSearchable    bool   `json:"guessed_searchable"`
	GuessedSortable      bool   `json:"guessed_sortable"`
}

type SchemaField struct {
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

type PayloadConfig struct {
	ListFields   []string `json:"list_fields"`
	FormFields   []string `json:"form_fields"`
	SearchFields []string `json:"search_fields"`
	Title        string   `json:"title,omitempty"`
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

type GeneratorService struct {
	RepoRoot string
	DB       *gorm.DB
}
