package codegen

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/lvjiaben/goweb-core/httpx"
	"github.com/lvjiaben/goweb-core/validate"
	"github.com/lvjiaben/goweb-scaffold/internal/bootstrap"
	"github.com/lvjiaben/goweb-scaffold/internal/gen/service"
	"log/slog"
)

func newRunnerTestRuntime(t *testing.T) *bootstrap.Runtime {
	t.Helper()
	engine := httpx.NewEngine(slog.New(slog.NewTextHandler(ioDiscard{}, nil)))
	repoRoot := t.TempDir()
	for _, dir := range []string{
		"internal/modules",
		"internal/gen",
		"vben-admin/apps/admin/src/generated",
	} {
		if err := os.MkdirAll(filepath.Join(repoRoot, dir), 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", dir, err)
		}
	}
	return &bootstrap.Runtime{
		RepoRoot:  repoRoot,
		Engine:    engine,
		Validator: validate.New(),
	}
}

type ioDiscard struct{}

func (ioDiscard) Write(p []byte) (int, error) { return len(p), nil }

func loadRunnerColumns(t *testing.T) []service.ColumnInfo {
	t.Helper()
	path := filepath.Join("..", "..", "gen", "service", "testdata", "demo_article_columns.json")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read columns fixture: %v", err)
	}
	var columns []service.ColumnInfo
	if err := json.Unmarshal(raw, &columns); err != nil {
		t.Fatalf("decode columns fixture: %v", err)
	}
	return columns
}

func loadRunnerPayload(t *testing.T) json.RawMessage {
	t.Helper()
	path := filepath.Join("..", "..", "gen", "service", "testdata", "demo_article_payload.json")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read payload fixture: %v", err)
	}
	return raw
}

func newFixtureRunner(t *testing.T) *Runner {
	t.Helper()
	runtime := newRunnerTestRuntime(t)
	runner := NewRunner(runtime)
	columns := loadRunnerColumns(t)
	runner.listColumnsFunc = func(_ *bootstrap.Runtime, tableName string) ([]service.ColumnInfo, error) {
		if tableName != "demo_article" {
			return nil, os.ErrNotExist
		}
		return columns, nil
	}
	runner.listTablesFunc = func(_ *bootstrap.Runtime) ([]BusinessTable, error) {
		return []BusinessTable{{
			TableName:    "demo_article",
			DisplayName:  "演示文章",
			TableComment: "演示文章",
		}}, nil
	}
	return runner
}

func TestRunnerExportImportRoundTrip(t *testing.T) {
	runner := newFixtureRunner(t)
	payload := loadRunnerPayload(t)

	if _, err := runner.Generate(ActionInput{
		ModuleName:     "demo_article",
		TableName:      "demo_article",
		Payload:        payload,
		Overwrite:      true,
		RegisterModule: true,
		UpsertMenu:     false,
	}); err != nil {
		t.Fatalf("generate fixture module: %v", err)
	}

	exportFile, err := runner.Export(ExportInput{ModuleName: "demo_article"})
	if err != nil {
		t.Fatalf("export module: %v", err)
	}
	exportPath := filepath.Join(t.TempDir(), "demo_article.codegen.json")
	raw, err := json.Marshal(exportFile)
	if err != nil {
		t.Fatalf("marshal export: %v", err)
	}
	if err := os.WriteFile(exportPath, raw, 0o644); err != nil {
		t.Fatalf("write export file: %v", err)
	}

	result, err := runner.Import(ImportInput{
		FromPath:       exportPath,
		Mode:           ImportModeDiff,
		Overwrite:      true,
		RegisterModule: true,
		UpsertMenu:     false,
	})
	if err != nil {
		t.Fatalf("import diff from export: %v", err)
	}
	if result.SourceKind != "export" {
		t.Fatalf("expected export source, got %s", result.SourceKind)
	}
	if result.Diff == nil || result.Diff.ModuleName != "demo_article" {
		t.Fatalf("unexpected import diff result: %+v", result)
	}
}

func TestRunnerRemoveAndRegenerateWithLock(t *testing.T) {
	runner := newFixtureRunner(t)
	payload := loadRunnerPayload(t)

	if _, err := runner.Generate(ActionInput{
		ModuleName:     "demo_article",
		TableName:      "demo_article",
		Payload:        payload,
		Overwrite:      true,
		RegisterModule: true,
		UpsertMenu:     false,
	}); err != nil {
		t.Fatalf("generate fixture module: %v", err)
	}

	removeResult, err := runner.Remove(service.RemoveInput{
		ModuleName:       "demo_article",
		RemoveFiles:      true,
		UnregisterModule: true,
		RemoveMenu:       false,
		RemoveHistory:    false,
		RemoveLock:       false,
	})
	if err != nil {
		t.Fatalf("remove module: %v", err)
	}
	if !containsPath(removeResult.RemovedFiles, "internal/modules/demo_article/model.go") {
		t.Fatalf("expected model.go removed, got %v", removeResult.RemovedFiles)
	}

	modelPath := filepath.Join(runner.Runtime.RepoRoot, "internal/modules/demo_article/model.go")
	if _, err := os.Stat(modelPath); !os.IsNotExist(err) {
		t.Fatalf("expected model.go removed, stat err=%v", err)
	}
	lockPath := filepath.Join(runner.Runtime.RepoRoot, "internal/modules/demo_article/codegen.lock.json")
	if _, err := os.Stat(lockPath); err != nil {
		t.Fatalf("expected lock file retained, stat err=%v", err)
	}

	regenerateResult, err := runner.Regenerate(RegenerateInput{
		ModuleName:     "demo_article",
		Overwrite:      true,
		RegisterModule: true,
		UpsertMenu:     false,
	})
	if err != nil {
		t.Fatalf("regenerate module: %v", err)
	}
	if !containsPath(regenerateResult.GeneratedFiles, "internal/modules/demo_article/model.go") {
		t.Fatalf("expected model.go regenerated, got %v", regenerateResult.GeneratedFiles)
	}

	modulesGenPath := filepath.Join(runner.Runtime.RepoRoot, "internal/gen/modules_gen.go")
	raw, err := os.ReadFile(modulesGenPath)
	if err != nil {
		t.Fatalf("read modules_gen.go: %v", err)
	}
	if !strings.Contains(string(raw), "demo_article") {
		t.Fatalf("expected modules_gen.go to restore demo_article")
	}
}

func containsPath(items []string, target string) bool {
	for _, item := range items {
		if item == target {
			return true
		}
	}
	return false
}
