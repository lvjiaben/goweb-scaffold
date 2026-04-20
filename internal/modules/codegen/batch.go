package codegen

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/lvjiaben/goweb-scaffold/internal/gen/service"
)

type BatchMode string

const (
	BatchModePreview       BatchMode = "preview"
	BatchModeDiff          BatchMode = "diff"
	BatchModeGenerate      BatchMode = "generate"
	BatchModeRegenerate    BatchMode = "regenerate"
	BatchModeRemove        BatchMode = "remove"
	BatchModeExport        BatchMode = "export"
	BatchModeCheckBreaking BatchMode = "check-breaking"
)

type BatchInput struct {
	PlanPath        string
	Mode            BatchMode
	ContinueOnError bool
}

type BatchModuleResult struct {
	ModuleName string                       `json:"module_name"`
	TableName  string                       `json:"table_name,omitempty"`
	SourceKind string                       `json:"source_kind,omitempty"`
	Status     string                       `json:"status"`
	Preview    *service.Preview             `json:"preview,omitempty"`
	Diff       *service.DiffResult          `json:"diff,omitempty"`
	Generate   *service.GenerateResult      `json:"generate,omitempty"`
	Remove     *service.RemoveResult        `json:"remove,omitempty"`
	Export     *service.ExportFile          `json:"export,omitempty"`
	Breaking   *service.BreakingCheckResult `json:"breaking,omitempty"`
	Error      string                       `json:"error,omitempty"`
}

type BatchResult struct {
	Mode             BatchMode           `json:"mode"`
	PlanPath         string              `json:"plan_path"`
	Total            int                 `json:"total"`
	SuccessCount     int                 `json:"success_count"`
	FailedCount      int                 `json:"failed_count"`
	SkippedCount     int                 `json:"skipped_count"`
	SameCount        int                 `json:"same_count,omitempty"`
	NonBreakingCount int                 `json:"non_breaking_count,omitempty"`
	BreakingCount    int                 `json:"breaking_count,omitempty"`
	Results          []BatchModuleResult `json:"results"`
}

type resolvedPlanEntry struct {
	ModuleName       string
	TableName        string
	Payload          json.RawMessage
	SourceKind       string
	Overwrite        bool
	RegisterModule   bool
	UpsertMenu       bool
	RemoveFiles      bool
	UnregisterModule bool
	RemoveMenu       bool
	RemoveHistory    bool
	RemoveLock       bool
}

func (r *Runner) RunBatch(input BatchInput) (BatchResult, error) {
	plan, err := LoadPlan(input.PlanPath)
	if err != nil {
		return BatchResult{}, err
	}
	result := BatchResult{
		Mode:     input.Mode,
		PlanPath: input.PlanPath,
		Total:    len(plan.Modules),
		Results:  make([]BatchModuleResult, 0, len(plan.Modules)),
	}

	for _, entry := range plan.Modules {
		resolved, err := r.resolvePlanEntry(plan, entry)
		if err != nil {
			moduleName := strings.TrimSpace(entry.ModuleName)
			if moduleName == "" {
				moduleName = strings.TrimSpace(entry.From)
			}
			item := BatchModuleResult{
				ModuleName: moduleName,
				Status:     "failed",
				Error:      err.Error(),
			}
			result.Results = append(result.Results, item)
			result.FailedCount++
			if !input.ContinueOnError {
				return result, nil
			}
			continue
		}

		item := BatchModuleResult{
			ModuleName: resolved.ModuleName,
			TableName:  resolved.TableName,
			SourceKind: resolved.SourceKind,
			Status:     "success",
		}

		switch input.Mode {
		case BatchModePreview:
			previewPayload, _, err := r.Preview(resolved.ModuleName, resolved.TableName, resolved.Payload)
			if err != nil {
				item.Status = "failed"
				item.Error = err.Error()
				result.FailedCount++
			} else {
				item.Preview = &previewPayload
				result.SuccessCount++
			}
		case BatchModeDiff:
			diffResult, err := r.Diff(ActionInput{
				ModuleName:     resolved.ModuleName,
				TableName:      resolved.TableName,
				Payload:        resolved.Payload,
				Overwrite:      resolved.Overwrite,
				RegisterModule: resolved.RegisterModule,
				UpsertMenu:     resolved.UpsertMenu,
			})
			if err != nil {
				item.Status = "failed"
				item.Error = err.Error()
				result.FailedCount++
			} else {
				if len(diffResult.WouldCreateFiles) == 0 && len(diffResult.WouldOverwriteFiles) == 0 {
					item.Status = "skipped"
					result.SkippedCount++
				} else {
					result.SuccessCount++
				}
				item.Diff = &diffResult
			}
		case BatchModeGenerate:
			generateResult, err := r.Generate(ActionInput{
				ModuleName:     resolved.ModuleName,
				TableName:      resolved.TableName,
				Payload:        resolved.Payload,
				Overwrite:      resolved.Overwrite,
				RegisterModule: resolved.RegisterModule,
				UpsertMenu:     resolved.UpsertMenu,
			})
			if err != nil {
				item.Status = "failed"
				item.Error = err.Error()
				result.FailedCount++
			} else {
				if len(generateResult.GeneratedFiles) == 0 && len(generateResult.OverwrittenFiles) == 0 {
					item.Status = "skipped"
					result.SkippedCount++
				} else {
					result.SuccessCount++
				}
				item.Generate = &generateResult
			}
		case BatchModeRegenerate:
			generateResult, err := r.Regenerate(RegenerateInput{
				ModuleName:     resolved.ModuleName,
				Overwrite:      resolved.Overwrite,
				RegisterModule: resolved.RegisterModule,
				UpsertMenu:     resolved.UpsertMenu,
			})
			if err != nil {
				item.Status = "failed"
				item.Error = err.Error()
				result.FailedCount++
			} else {
				if len(generateResult.GeneratedFiles) == 0 && len(generateResult.OverwrittenFiles) == 0 {
					item.Status = "skipped"
					result.SkippedCount++
				} else {
					result.SuccessCount++
				}
				item.Generate = &generateResult
			}
		case BatchModeRemove:
			removeResult, err := r.Remove(service.RemoveInput{
				ModuleName:       resolved.ModuleName,
				RemoveFiles:      resolved.RemoveFiles,
				UnregisterModule: resolved.UnregisterModule,
				RemoveMenu:       resolved.RemoveMenu,
				RemoveHistory:    resolved.RemoveHistory,
				RemoveLock:       resolved.RemoveLock,
			})
			if err != nil {
				item.Status = "failed"
				item.Error = err.Error()
				result.FailedCount++
			} else {
				if len(removeResult.RemovedFiles) == 0 && len(removeResult.RegeneratedRegistryFiles) == 0 && len(removeResult.RemovedMenuRecords) == 0 {
					item.Status = "skipped"
					result.SkippedCount++
				} else {
					result.SuccessCount++
				}
				item.Remove = &removeResult
			}
		case BatchModeExport:
			exportResult, err := r.Export(ExportInput{
				ModuleName: resolved.ModuleName,
			})
			if err != nil {
				item.Status = "failed"
				item.Error = err.Error()
				result.FailedCount++
			} else {
				item.Export = &exportResult
				result.SuccessCount++
			}
		case BatchModeCheckBreaking:
			breakingResult, err := r.CheckBreaking(CheckBreakingInput{
				ModuleName:     resolved.ModuleName,
				TableName:      resolved.TableName,
				Payload:        resolved.Payload,
				RegisterModule: resolved.RegisterModule,
			})
			if err != nil {
				item.Status = "failed"
				item.Error = err.Error()
				result.FailedCount++
			} else {
				item.Breaking = &breakingResult
				item.Status = breakingResult.Level
				result.SuccessCount++
				switch breakingResult.Level {
				case service.CompatibilitySame:
					result.SameCount++
				case service.CompatibilityNonBreaking:
					result.NonBreakingCount++
				case service.CompatibilityBreaking:
					result.BreakingCount++
				}
			}
		default:
			return result, fmt.Errorf("unsupported batch mode: %s", input.Mode)
		}

		result.Results = append(result.Results, item)
		if item.Status == "failed" && !input.ContinueOnError {
			return result, nil
		}
	}

	return result, nil
}

func (r *Runner) resolvePlanEntry(plan BatchPlan, entry PlanModuleEntry) (resolvedPlanEntry, error) {
	resolvedSource, err := r.ResolveInput(SourceInput{
		ModuleName: strings.TrimSpace(entry.ModuleName),
		TableName:  strings.TrimSpace(entry.TableName),
		FromPath:   strings.TrimSpace(entry.From),
	})
	if err != nil && strings.TrimSpace(entry.From) != "" {
		return resolvedPlanEntry{}, err
	}
	if len(entry.Payload) > 0 {
		resolvedSource.Payload = entry.Payload
		if resolvedSource.SourceKind == "" {
			resolvedSource.SourceKind = "payload"
		}
	}
	if strings.TrimSpace(resolvedSource.ModuleName) == "" {
		return resolvedPlanEntry{}, fmt.Errorf("module_name is required")
	}
	if strings.TrimSpace(resolvedSource.TableName) == "" && strings.TrimSpace(entry.TableName) == "" && strings.TrimSpace(entry.From) == "" {
		return resolvedPlanEntry{}, fmt.Errorf("table_name is required")
	}

	resolved := resolvedPlanEntry{
		ModuleName:       resolvedSource.ModuleName,
		TableName:        resolvedSource.TableName,
		Payload:          resolvedSource.Payload,
		SourceKind:       firstNonEmptyString(resolvedSource.SourceKind, "direct"),
		Overwrite:        plan.EffectiveBool(entry.Overwrite, plan.Defaults.Overwrite, true),
		RegisterModule:   plan.EffectiveBool(entry.RegisterModule, plan.Defaults.RegisterModule, true),
		UpsertMenu:       plan.EffectiveBool(entry.UpsertMenu, plan.Defaults.UpsertMenu, true),
		RemoveFiles:      plan.EffectiveBool(entry.RemoveFiles, plan.Defaults.RemoveFiles, true),
		UnregisterModule: plan.EffectiveBool(entry.UnregisterModule, plan.Defaults.UnregisterModule, true),
		RemoveMenu:       plan.EffectiveBool(entry.RemoveMenu, plan.Defaults.RemoveMenu, true),
		RemoveHistory:    plan.EffectiveBool(entry.RemoveHistory, plan.Defaults.RemoveHistory, false),
		RemoveLock:       plan.EffectiveBool(entry.RemoveLock, plan.Defaults.RemoveLock, true),
	}
	if strings.TrimSpace(entry.TableName) != "" {
		resolved.TableName = strings.TrimSpace(entry.TableName)
	}
	return resolved, nil
}
