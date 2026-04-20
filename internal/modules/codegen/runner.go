package codegen

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/lvjiaben/goweb-scaffold/internal/bootstrap"
	"github.com/lvjiaben/goweb-scaffold/internal/gen/service"
	"github.com/lvjiaben/goweb-scaffold/internal/modules/model"
)

type ActionInput struct {
	ModuleName     string
	TableName      string
	Payload        json.RawMessage
	Overwrite      bool
	RegisterModule bool
	UpsertMenu     bool
}

type CheckBreakingInput struct {
	ModuleName     string
	TableName      string
	Payload        json.RawMessage
	RegisterModule bool
}

type RegenerateInput struct {
	ModuleName     string
	HistoryID      int64
	Overwrite      bool
	RegisterModule bool
	UpsertMenu     bool
}

type ExportInput struct {
	ModuleName string
	HistoryID  int64
}

type ImportMode string

const (
	ImportModePreview  ImportMode = "preview"
	ImportModeDiff     ImportMode = "diff"
	ImportModeGenerate ImportMode = "generate"
)

type ImportInput struct {
	FromPath       string
	ModuleName     string
	TableName      string
	PayloadPath    string
	Mode           ImportMode
	Overwrite      bool
	RegisterModule bool
	UpsertMenu     bool
}

type ImportResult struct {
	Mode       ImportMode              `json:"mode"`
	SourceKind string                  `json:"source_kind"`
	ModuleName string                  `json:"module_name"`
	TableName  string                  `json:"table_name"`
	Preview    *service.Preview        `json:"preview,omitempty"`
	Diff       *service.DiffResult     `json:"diff,omitempty"`
	Generate   *service.GenerateResult `json:"generate,omitempty"`
}

type SourceInput struct {
	ModuleName  string
	TableName   string
	PayloadPath string
	FromPath    string
}

type ResolvedInput struct {
	ModuleName  string
	TableName   string
	Payload     json.RawMessage
	SourceKind  string
	RoutePath   string
	Permissions []string
}

type Runner struct {
	Runtime         *bootstrap.Runtime
	listTablesFunc  func(*bootstrap.Runtime) ([]BusinessTable, error)
	listColumnsFunc func(*bootstrap.Runtime, string) ([]service.ColumnInfo, error)
}

func NewRunner(runtime *bootstrap.Runtime) *Runner {
	return &Runner{
		Runtime:         runtime,
		listTablesFunc:  listBusinessTables,
		listColumnsFunc: listTableColumns,
	}
}

func (r *Runner) generatorService() service.GeneratorService {
	return service.GeneratorService{
		RepoRoot: r.Runtime.RepoRoot,
		DB:       r.Runtime.DB,
	}
}

func (r *Runner) Tables() ([]BusinessTable, error) {
	return r.listTablesFunc(r.Runtime)
}

func (r *Runner) Modules() ([]service.ManagedModule, error) {
	return r.generatorService().ListModules()
}

func (r *Runner) Preview(moduleName string, tableName string, payload json.RawMessage) (service.Preview, []service.ColumnInfo, error) {
	moduleName = strings.TrimSpace(moduleName)
	tableName = strings.TrimSpace(tableName)
	if moduleName == "" || tableName == "" {
		return service.Preview{}, nil, fmt.Errorf("module_name and table_name are required")
	}
	if len(payload) == 0 {
		payload = json.RawMessage(`{}`)
	}

	columns, err := r.listColumnsFunc(r.Runtime, tableName)
	if err != nil {
		return service.Preview{}, nil, err
	}
	if len(columns) == 0 {
		return service.Preview{}, nil, fmt.Errorf("table %s has no available columns for generation", tableName)
	}
	return service.BuildPreview(moduleName, tableName, payload, columns), columns, nil
}

func (r *Runner) Diff(input ActionInput) (service.DiffResult, error) {
	previewPayload, columns, err := r.Preview(input.ModuleName, input.TableName, input.Payload)
	if err != nil {
		return service.DiffResult{}, err
	}
	return r.generatorService().Diff(service.GenerateInput{
		ModuleName:     input.ModuleName,
		TableName:      input.TableName,
		Payload:        previewPayload.Payload,
		Preview:        previewPayload,
		Columns:        columns,
		Overwrite:      input.Overwrite,
		RegisterModule: input.RegisterModule,
		UpsertMenu:     input.UpsertMenu,
	})
}

func (r *Runner) CheckBreaking(input CheckBreakingInput) (service.BreakingCheckResult, error) {
	moduleName := strings.TrimSpace(input.ModuleName)
	tableName := strings.TrimSpace(input.TableName)
	payload := input.Payload
	if moduleName == "" {
		return service.BreakingCheckResult{}, fmt.Errorf("module_name is required")
	}
	if tableName == "" {
		source, err := r.resolveRegenerateSource(RegenerateInput{ModuleName: moduleName})
		if err != nil {
			return service.BreakingCheckResult{}, err
		}
		tableName = source.TableName
		if len(payload) == 0 {
			payload = source.Payload
		}
	}
	if len(payload) == 0 {
		payload = json.RawMessage(`{}`)
	}
	previewPayload, columns, err := r.Preview(moduleName, tableName, payload)
	if err != nil {
		return service.BreakingCheckResult{}, err
	}
	return r.generatorService().CheckBreaking(service.GenerateInput{
		ModuleName:     moduleName,
		TableName:      tableName,
		Payload:        previewPayload.Payload,
		Preview:        previewPayload,
		Columns:        columns,
		RegisterModule: input.RegisterModule,
	})
}

func (r *Runner) Generate(input ActionInput) (service.GenerateResult, error) {
	previewPayload, columns, err := r.Preview(input.ModuleName, input.TableName, input.Payload)
	if err != nil {
		return service.GenerateResult{}, err
	}

	result, err := r.generatorService().Generate(service.GenerateInput{
		ModuleName:     input.ModuleName,
		TableName:      input.TableName,
		Payload:        previewPayload.Payload,
		Preview:        previewPayload,
		Columns:        columns,
		Overwrite:      input.Overwrite,
		RegisterModule: input.RegisterModule,
		UpsertMenu:     input.UpsertMenu,
	})
	if err != nil {
		return service.GenerateResult{}, err
	}

	rawPayload, err := json.Marshal(previewPayload.Payload)
	if err != nil {
		return service.GenerateResult{}, err
	}
	record := model.CodegenHistory{
		ModuleName:  strings.TrimSpace(input.ModuleName),
		SourceTable: strings.TrimSpace(input.TableName),
		Status:      "generated",
		Payload:     model.JSON(rawPayload),
		Remark:      "generated admin CRUD files",
	}
	if r.Runtime.DB != nil {
		if err := r.Runtime.DB.Create(&record).Error; err != nil {
			return service.GenerateResult{}, err
		}
	}
	return result, nil
}

func (r *Runner) Regenerate(input RegenerateInput) (service.GenerateResult, error) {
	source, err := r.resolveRegenerateSource(input)
	if err != nil {
		return service.GenerateResult{}, err
	}
	result, err := r.Generate(ActionInput{
		ModuleName:     source.ModuleName,
		TableName:      source.TableName,
		Payload:        source.Payload,
		Overwrite:      input.Overwrite,
		RegisterModule: input.RegisterModule,
		UpsertMenu:     input.UpsertMenu,
	})
	if err != nil {
		return service.GenerateResult{}, err
	}

	if r.Runtime.DB != nil {
		var row model.CodegenHistory
		if err := r.Runtime.DB.Where("module_name = ?", source.ModuleName).Order("id DESC").First(&row).Error; err == nil {
			row.Status = "regenerated"
			row.Remark = "regenerated admin CRUD files"
			_ = r.Runtime.DB.Save(&row).Error
		}
	}
	return result, nil
}

func (r *Runner) Remove(input service.RemoveInput) (service.RemoveResult, error) {
	return r.generatorService().Remove(input)
}

func (r *Runner) Export(input ExportInput) (service.ExportFile, error) {
	if input.HistoryID > 0 {
		source, err := r.loadRegenerateSourceFromHistory(input.HistoryID)
		if err != nil {
			return service.ExportFile{}, err
		}
		previewPayload, _, err := r.Preview(source.ModuleName, source.TableName, source.Payload)
		if err != nil {
			return service.ExportFile{}, err
		}
		return service.ExportFile{
			GeneratedBy:     service.GeneratorName,
			Format:          service.ExportFormatName,
			Version:         service.ExportFormatVersion,
			ModuleName:      source.ModuleName,
			TableName:       source.TableName,
			TemplateVersion: service.TemplateVersion,
			Payload:         previewPayload.Payload,
			PreviewSummary: service.LockPreviewSummary{
				TableComment:   previewPayload.TableComment,
				Page:           previewPayload.Page,
				API:            previewPayload.API,
				InferredFields: previewPayload.InferredFields,
				FormSchema:     previewPayload.FormSchema,
				ListSchema:     previewPayload.ListSchema,
				SearchSchema:   previewPayload.SearchSchema,
			},
			Snapshot: service.Snapshot{},
			PermissionCodes: []string{
				service.ToSnake(source.ModuleName) + ".list",
				service.ToSnake(source.ModuleName) + ".save",
				service.ToSnake(source.ModuleName) + ".delete",
			},
			RoutePath: previewPayload.Page.RoutePath,
		}, nil
	}

	lock, err := r.generatorService().LoadLock(strings.TrimSpace(input.ModuleName))
	if err == nil {
		return service.BuildExportFromLock(lock), nil
	}
	if !os.IsNotExist(err) {
		return service.ExportFile{}, err
	}

	source, err := r.resolveRegenerateSource(RegenerateInput{ModuleName: input.ModuleName})
	if err != nil {
		return service.ExportFile{}, err
	}
	previewPayload, _, err := r.Preview(source.ModuleName, source.TableName, source.Payload)
	if err != nil {
		return service.ExportFile{}, err
	}
	return service.ExportFile{
		GeneratedBy:     service.GeneratorName,
		Format:          service.ExportFormatName,
		Version:         service.ExportFormatVersion,
		ModuleName:      source.ModuleName,
		TableName:       source.TableName,
		TemplateVersion: service.TemplateVersion,
		Payload:         previewPayload.Payload,
		PreviewSummary: service.LockPreviewSummary{
			TableComment:   previewPayload.TableComment,
			Page:           previewPayload.Page,
			API:            previewPayload.API,
			InferredFields: previewPayload.InferredFields,
			FormSchema:     previewPayload.FormSchema,
			ListSchema:     previewPayload.ListSchema,
			SearchSchema:   previewPayload.SearchSchema,
		},
		Snapshot: service.Snapshot{},
		PermissionCodes: []string{
			service.ToSnake(source.ModuleName) + ".list",
			service.ToSnake(source.ModuleName) + ".save",
			service.ToSnake(source.ModuleName) + ".delete",
		},
		RoutePath: previewPayload.Page.RoutePath,
	}, nil
}

func (r *Runner) Import(input ImportInput) (ImportResult, error) {
	resolved, err := r.ResolveInput(SourceInput{
		ModuleName:  input.ModuleName,
		TableName:   input.TableName,
		PayloadPath: input.PayloadPath,
		FromPath:    input.FromPath,
	})
	if err != nil {
		return ImportResult{}, err
	}

	result := ImportResult{
		Mode:       input.Mode,
		SourceKind: resolved.SourceKind,
		ModuleName: resolved.ModuleName,
		TableName:  resolved.TableName,
	}
	switch input.Mode {
	case ImportModeGenerate:
		generateResult, err := r.Generate(ActionInput{
			ModuleName:     resolved.ModuleName,
			TableName:      resolved.TableName,
			Payload:        resolved.Payload,
			Overwrite:      input.Overwrite,
			RegisterModule: input.RegisterModule,
			UpsertMenu:     input.UpsertMenu,
		})
		if err != nil {
			return ImportResult{}, err
		}
		result.Generate = &generateResult
	case ImportModeDiff:
		diffResult, err := r.Diff(ActionInput{
			ModuleName:     resolved.ModuleName,
			TableName:      resolved.TableName,
			Payload:        resolved.Payload,
			Overwrite:      input.Overwrite,
			RegisterModule: input.RegisterModule,
			UpsertMenu:     input.UpsertMenu,
		})
		if err != nil {
			return ImportResult{}, err
		}
		result.Diff = &diffResult
	default:
		previewPayload, _, err := r.Preview(resolved.ModuleName, resolved.TableName, resolved.Payload)
		if err != nil {
			return ImportResult{}, err
		}
		result.Mode = ImportModePreview
		result.Preview = &previewPayload
	}
	return result, nil
}

func (r *Runner) ResolveInput(input SourceInput) (ResolvedInput, error) {
	resolved := ResolvedInput{}
	if fromPath := strings.TrimSpace(input.FromPath); fromPath != "" {
		raw, err := os.ReadFile(fromPath)
		if err != nil {
			return ResolvedInput{}, err
		}
		doc, err := service.DecodeSourceDocument(raw)
		if err != nil {
			return ResolvedInput{}, err
		}
		payloadRaw, err := json.Marshal(doc.Payload)
		if err != nil {
			return ResolvedInput{}, err
		}
		resolved.ModuleName = doc.ModuleName
		resolved.TableName = doc.TableName
		resolved.Payload = payloadRaw
		resolved.SourceKind = doc.Kind
		resolved.RoutePath = doc.RoutePath
		resolved.Permissions = append([]string{}, doc.PermissionCodes...)
	}

	if moduleName := strings.TrimSpace(input.ModuleName); moduleName != "" {
		resolved.ModuleName = moduleName
	}
	if tableName := strings.TrimSpace(input.TableName); tableName != "" {
		resolved.TableName = tableName
	}
	if payloadPath := strings.TrimSpace(input.PayloadPath); payloadPath != "" {
		raw, err := os.ReadFile(payloadPath)
		if err != nil {
			return ResolvedInput{}, err
		}
		resolved.Payload = raw
		if resolved.SourceKind == "" {
			resolved.SourceKind = "payload"
		}
	}
	if len(resolved.Payload) == 0 {
		resolved.Payload = json.RawMessage(`{}`)
	}
	return resolved, nil
}

func (r *Runner) resolveRegenerateSource(input RegenerateInput) (regenerateSource, error) {
	if input.HistoryID > 0 {
		return r.loadRegenerateSourceFromHistory(input.HistoryID)
	}

	lock, err := r.generatorService().LoadLock(strings.TrimSpace(input.ModuleName))
	if err == nil {
		rawPayload, marshalErr := json.Marshal(lock.Payload)
		if marshalErr != nil {
			return regenerateSource{}, marshalErr
		}
		return regenerateSource{
			ModuleName: lock.ModuleName,
			TableName:  lock.TableName,
			Payload:    rawPayload,
		}, nil
	}
	if !os.IsNotExist(err) {
		return regenerateSource{}, err
	}
	if r.Runtime.DB == nil {
		return regenerateSource{}, fmt.Errorf("codegen source for module %s not found", input.ModuleName)
	}

	var row model.CodegenHistory
	if err := r.Runtime.DB.Where("module_name = ?", strings.TrimSpace(input.ModuleName)).Order("id DESC").First(&row).Error; err != nil {
		return regenerateSource{}, fmt.Errorf("codegen source for module %s not found", input.ModuleName)
	}
	return regenerateSource{
		ModuleName: strings.TrimSpace(row.ModuleName),
		TableName:  strings.TrimSpace(row.SourceTable),
		Payload:    json.RawMessage(row.Payload),
	}, nil
}

func (r *Runner) loadRegenerateSourceFromHistory(historyID int64) (regenerateSource, error) {
	if r.Runtime.DB == nil {
		return regenerateSource{}, fmt.Errorf("codegen history %d not found", historyID)
	}
	var row model.CodegenHistory
	if err := r.Runtime.DB.First(&row, historyID).Error; err != nil {
		return regenerateSource{}, fmt.Errorf("codegen history %d not found", historyID)
	}
	return regenerateSource{
		ModuleName: strings.TrimSpace(row.ModuleName),
		TableName:  strings.TrimSpace(row.SourceTable),
		Payload:    json.RawMessage(row.Payload),
	}, nil
}
