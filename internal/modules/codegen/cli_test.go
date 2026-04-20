package codegen

import (
	"bytes"
	"encoding/json"
	"errors"
	"testing"

	"github.com/lvjiaben/goweb-scaffold/internal/gen/service"
)

type fakeCLIBackend struct {
	modulesResult []service.ManagedModule
	importResult  ImportResult
	lastImport    ImportInput
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
