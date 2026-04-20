package codegen

import (
	"bytes"
	"encoding/json"
	"errors"
	"path/filepath"
	"testing"

	"github.com/lvjiaben/goweb-scaffold/internal/gen/service"
)

type fakeCLIBackend struct {
	modulesResult []service.ManagedModule
	importResult  ImportResult
	batchResult   BatchResult
	lastImport    ImportInput
	lastBatch     BatchInput
}

func (f *fakeCLIBackend) Tables() ([]BusinessTable, error) { return nil, errors.New("not implemented") }
func (f *fakeCLIBackend) Modules() ([]service.ManagedModule, error) {
	return f.modulesResult, nil
}
func (f *fakeCLIBackend) Preview(string, string, json.RawMessage) (service.Preview, []service.ColumnInfo, error) {
	return service.Preview{}, nil, errors.New("not implemented")
}
func (f *fakeCLIBackend) Diff(ActionInput) (service.DiffResult, error) {
	return service.DiffResult{}, errors.New("not implemented")
}
func (f *fakeCLIBackend) Generate(ActionInput) (service.GenerateResult, error) {
	return service.GenerateResult{}, errors.New("not implemented")
}
func (f *fakeCLIBackend) Regenerate(RegenerateInput) (service.GenerateResult, error) {
	return service.GenerateResult{}, errors.New("not implemented")
}
func (f *fakeCLIBackend) Remove(service.RemoveInput) (service.RemoveResult, error) {
	return service.RemoveResult{}, errors.New("not implemented")
}
func (f *fakeCLIBackend) Export(ExportInput) (service.ExportFile, error) {
	return service.ExportFile{}, errors.New("not implemented")
}
func (f *fakeCLIBackend) Import(input ImportInput) (ImportResult, error) {
	f.lastImport = input
	return f.importResult, nil
}
func (f *fakeCLIBackend) ResolveInput(input SourceInput) (ResolvedInput, error) {
	return ResolvedInput{
		ModuleName: input.ModuleName,
		TableName:  input.TableName,
		Payload:    json.RawMessage(`{}`),
		SourceKind: "direct",
	}, nil
}
func (f *fakeCLIBackend) RunBatch(input BatchInput) (BatchResult, error) {
	f.lastBatch = input
	return f.batchResult, nil
}

func TestParsePreviewCommand(t *testing.T) {
	cmd, err := parsePreviewCommand([]string{
		"-module", "demo_article",
		"-table", "demo_article",
		"-payload", "/tmp/demo.json",
		"-from", "/tmp/demo.codegen.json",
		"-format", "json",
		"-output", "/tmp/preview.json",
		"-config", "configs/config.example.yaml",
	})
	if err != nil {
		t.Fatalf("parse preview command: %v", err)
	}
	if cmd.moduleName != "demo_article" || cmd.tableName != "demo_article" {
		t.Fatalf("unexpected preview args: %+v", cmd)
	}
	if cmd.payload == "" || cmd.from == "" || cmd.format != "json" || cmd.outputPath == "" {
		t.Fatalf("unexpected preview flags: %+v", cmd)
	}
}

func TestCLIOutputsModulesAsJSON(t *testing.T) {
	backend := &fakeCLIBackend{
		modulesResult: []service.ManagedModule{{
			ModuleName:      "demo_article",
			TableName:       "demo_article",
			GeneratedAt:     "2026-04-20T05:57:50+08:00",
			TemplateVersion: "v5",
			RoutePath:       "/system/demo-article",
		}},
	}
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := NewCLI(backend, &stdout, &stderr).Run([]string{"modules", "-format", "json"})
	if code != 0 {
		t.Fatalf("expected success, got code=%d stderr=%s", code, stderr.String())
	}

	var payload map[string][]service.ManagedModule
	if err := json.Unmarshal(stdout.Bytes(), &payload); err != nil {
		t.Fatalf("decode modules json: %v, body=%s", err, stdout.String())
	}
	if len(payload["list"]) != 1 || payload["list"][0].ModuleName != "demo_article" {
		t.Fatalf("unexpected modules payload: %+v", payload)
	}
}

func TestCLIImportDispatchesDiffMode(t *testing.T) {
	backend := &fakeCLIBackend{
		importResult: ImportResult{
			Mode:       ImportModeDiff,
			SourceKind: "export",
			ModuleName: "demo_article",
			TableName:  "demo_article",
			Diff: &service.DiffResult{
				ModuleName: "demo_article",
			},
		},
	}
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := NewCLI(backend, &stdout, &stderr).Run([]string{"import", "-from", "/tmp/demo.codegen.json", "-diff", "-format", "json"})
	if code != 0 {
		t.Fatalf("expected success, got code=%d stderr=%s", code, stderr.String())
	}
	if backend.lastImport.Mode != ImportModeDiff {
		t.Fatalf("expected diff mode, got %+v", backend.lastImport)
	}
	if backend.lastImport.FromPath != "/tmp/demo.codegen.json" {
		t.Fatalf("expected import path forwarded, got %+v", backend.lastImport)
	}
	var payload map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &payload); err != nil {
		t.Fatalf("decode import json: %v", err)
	}
	if payload["source_kind"] != "export" {
		t.Fatalf("unexpected import payload: %+v", payload)
	}
}

func TestCLITemplatesOutputsJSON(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := NewCLI(&fakeCLIBackend{}, &stdout, &stderr).Run([]string{"templates", "-format", "json"})
	if code != 0 {
		t.Fatalf("expected success, got code=%d stderr=%s", code, stderr.String())
	}

	var payload map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &payload); err != nil {
		t.Fatalf("decode templates payload: %v", err)
	}
	if payload["current"] != "v6" {
		t.Fatalf("unexpected templates payload: %+v", payload)
	}
}

func TestCLIMigrateSourceOutputsJSON(t *testing.T) {
	path := filepath.Join("..", "..", "gen", "service", "testdata", "legacy", "demo_article_export_v5.json")
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := NewCLI(&fakeCLIBackend{}, &stdout, &stderr).Run([]string{"migrate-source", "-from", path, "-format", "json"})
	if code != 0 {
		t.Fatalf("expected success, got code=%d stderr=%s", code, stderr.String())
	}

	var payload map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &payload); err != nil {
		t.Fatalf("decode migrate-source payload: %v", err)
	}
	migration, ok := payload["migration"].(map[string]any)
	if !ok {
		t.Fatalf("unexpected migrate payload: %+v", payload)
	}
	if migration["to_version"] != "v6" {
		t.Fatalf("unexpected migration result: %+v", payload)
	}
}

func TestCLIBatchOutputsJSON(t *testing.T) {
	backend := &fakeCLIBackend{
		batchResult: BatchResult{
			Mode:         BatchModeDiff,
			PlanPath:     "examples/codegen/demo.plan.json",
			Total:        1,
			SuccessCount: 1,
			Results: []BatchModuleResult{{
				ModuleName: "demo_article",
				TableName:  "demo_article",
				SourceKind: "lock",
				Status:     "success",
				Diff: &service.DiffResult{
					ModuleName: "demo_article",
				},
			}},
		},
	}
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := NewCLI(backend, &stdout, &stderr).Run([]string{"batch", "-plan", "examples/codegen/demo.plan.json", "-mode", "diff", "-format", "json"})
	if code != 0 {
		t.Fatalf("expected success, got code=%d stderr=%s", code, stderr.String())
	}
	if backend.lastBatch.Mode != BatchModeDiff {
		t.Fatalf("expected diff mode, got %+v", backend.lastBatch)
	}

	var payload map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &payload); err != nil {
		t.Fatalf("decode batch payload: %v", err)
	}
	if payload["success_count"] != float64(1) {
		t.Fatalf("unexpected batch payload: %+v", payload)
	}
}
