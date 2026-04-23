package codegen

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/lvjiaben/goweb-scaffold/internal/bootstrap"
	genservice "github.com/lvjiaben/goweb-scaffold/internal/gen/service"
	sharedmodel "github.com/lvjiaben/goweb-scaffold/internal/shared/model"
)

type Service struct {
	runner *Runner
	repo   *Repo
}

func NewService(runtime *bootstrap.Runtime) *Service {
	return &Service{
		runner: NewRunner(runtime),
		repo:   NewRepo(runtime),
	}
}

func (s *Service) HistoryList() (map[string]any, error) {
	rows, err := s.repo.ListHistory()
	if err != nil {
		return nil, err
	}
	items := make([]HistoryItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, HistoryItem{
			ID:         row.ID,
			ModuleName: row.ModuleName,
			TableName:  row.SourceTable,
			Status:     row.Status,
			Payload:    json.RawMessage(row.Payload),
			Remark:     row.Remark,
			CreatedAt:  row.CreatedAt,
		})
	}
	return map[string]any{"list": items}, nil
}

func (s *Service) Tables() (map[string]any, error) {
	items, err := s.runner.Tables()
	if err != nil {
		return nil, err
	}
	return map[string]any{"list": items}, nil
}

func (s *Service) Modules() (map[string]any, error) {
	items, err := s.runner.Modules()
	if err != nil {
		return nil, err
	}
	return map[string]any{"list": items}, nil
}

func (s *Service) TableColumns(tableName string) (map[string]any, error) {
	tableName = strings.TrimSpace(tableName)
	if tableName == "" {
		return nil, validationError("table_name is required")
	}
	items, err := s.runner.listColumnsFunc(s.runner.Runtime, tableName)
	if err != nil {
		return nil, err
	}
	return map[string]any{"list": items}, nil
}

func (s *Service) Export(req ExportRequest) (genservice.ExportFile, error) {
	if strings.TrimSpace(req.ModuleName) == "" && req.HistoryID <= 0 {
		return genservice.ExportFile{}, validationError("module_name or history_id is required")
	}
	return s.runner.Export(ExportInput{
		ModuleName: req.ModuleName,
		HistoryID:  req.HistoryID,
	})
}

func (s *Service) Preview(req SaveRequest) (genservice.Preview, error) {
	req.ModuleName = strings.TrimSpace(req.ModuleName)
	req.TableName = strings.TrimSpace(req.TableName)
	previewPayload, _, err := s.runner.Preview(req.ModuleName, req.TableName, req.Payload)
	return previewPayload, err
}

func (s *Service) Diff(req GenerateRequest) (genservice.DiffResult, error) {
	return s.runner.Diff(ActionInput{
		ModuleName:     strings.TrimSpace(req.ModuleName),
		TableName:      strings.TrimSpace(req.TableName),
		Payload:        req.Payload,
		Overwrite:      req.Overwrite,
		RegisterModule: req.RegisterModule,
		UpsertMenu:     req.UpsertMenu,
	})
}

func (s *Service) CheckBreaking(req CheckBreakingRequest) (genservice.BreakingCheckResult, error) {
	return s.runner.CheckBreaking(CheckBreakingInput{
		ModuleName:     strings.TrimSpace(req.ModuleName),
		TableName:      strings.TrimSpace(req.TableName),
		Payload:        req.Payload,
		RegisterModule: req.RegisterModule == nil || *req.RegisterModule,
	})
}

func (s *Service) Generate(req GenerateRequest) (genservice.GenerateResult, error) {
	return s.runner.Generate(ActionInput{
		ModuleName:     strings.TrimSpace(req.ModuleName),
		TableName:      strings.TrimSpace(req.TableName),
		Payload:        req.Payload,
		Overwrite:      req.Overwrite,
		RegisterModule: req.RegisterModule,
		UpsertMenu:     req.UpsertMenu,
	})
}

func (s *Service) Regenerate(req RegenerateRequest) (genservice.GenerateResult, error) {
	req.ModuleName = strings.TrimSpace(req.ModuleName)
	if req.ModuleName == "" && req.HistoryID <= 0 {
		return genservice.GenerateResult{}, validationError("module_name or history_id is required")
	}
	return s.runner.Regenerate(RegenerateInput{
		ModuleName:     req.ModuleName,
		HistoryID:      req.HistoryID,
		Overwrite:      req.Overwrite,
		RegisterModule: req.RegisterModule,
		UpsertMenu:     req.UpsertMenu,
	})
}

func (s *Service) Remove(req RemoveRequest) (genservice.RemoveResult, error) {
	req.ModuleName = strings.TrimSpace(req.ModuleName)
	if req.ModuleName == "" {
		return genservice.RemoveResult{}, validationError("module_name is required")
	}
	return s.runner.Remove(genservice.RemoveInput{
		ModuleName:       req.ModuleName,
		RemoveFiles:      req.RemoveFiles,
		UnregisterModule: req.UnregisterModule,
		RemoveMenu:       req.RemoveMenu,
		RemoveHistory:    req.RemoveHistory,
		RemoveLock:       req.RemoveLock,
	})
}

func (s *Service) SaveDraft(req SaveRequest) (SaveDraftResult, error) {
	req.ModuleName = strings.TrimSpace(req.ModuleName)
	req.TableName = strings.TrimSpace(req.TableName)
	previewPayload, err := s.Preview(req)
	if err != nil {
		return SaveDraftResult{}, err
	}
	rawPayload, err := json.Marshal(previewPayload.Payload)
	if err != nil {
		return SaveDraftResult{}, err
	}
	record := CodegenHistory{
		ModuleName:  req.ModuleName,
		SourceTable: req.TableName,
		Status:      "draft",
		Payload:     sharedmodel.JSON(rawPayload),
		Remark:      "preview draft for admin codegen regenerate workflow",
	}
	if err := s.repo.CreateHistory(&record); err != nil {
		return SaveDraftResult{}, err
	}
	return SaveDraftResult{
		ID:          record.ID,
		Status:      record.Status,
		Placeholder: true,
	}, nil
}

func (s *Service) DeleteHistory(ids []int64) (DeleteHistoryResult, error) {
	ids = bootstrap.NormalizeIDs(ids...)
	if len(ids) == 0 {
		return DeleteHistoryResult{}, validationError("ids is required")
	}
	if err := s.repo.DeleteHistory(ids); err != nil {
		return DeleteHistoryResult{}, err
	}
	return DeleteHistoryResult{Deleted: len(ids)}, nil
}

func respondCodegenError(c interface {
	BadRequest(string)
	Error(error)
}, err error) {
	if err == nil {
		return
	}
	switch {
	case isValidationError(err):
		c.BadRequest(err.Error())
	case os.IsNotExist(err):
		c.BadRequest("codegen lock file not found")
	case strings.Contains(strings.ToLower(err.Error()), "not found"):
		c.BadRequest(err.Error())
	case strings.Contains(strings.ToLower(err.Error()), "required"):
		c.BadRequest(err.Error())
	case strings.Contains(strings.ToLower(err.Error()), "no available columns"):
		c.BadRequest(err.Error())
	default:
		c.Error(err)
	}
}

type validationError string

func (e validationError) Error() string {
	return string(e)
}

func isValidationError(err error) bool {
	_, ok := err.(validationError)
	return ok
}

func formatSourceNotFound(kind string, id any) error {
	return fmt.Errorf("codegen %s %v not found", kind, id)
}
