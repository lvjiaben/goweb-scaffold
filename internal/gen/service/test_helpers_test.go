package service

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func loadTestJSON[T any](t *testing.T, name string) T {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatalf("read %s: %v", name, err)
	}
	var value T
	if err := json.Unmarshal(raw, &value); err != nil {
		t.Fatalf("unmarshal %s: %v", name, err)
	}
	return value
}

func mustReadFile(t *testing.T, path string) []byte {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return raw
}

func writeGoldenIfNeeded(t *testing.T, path string, value any) {
	t.Helper()
	if os.Getenv("UPDATE_GOLDEN") == "" {
		return
	}
	raw, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		t.Fatalf("marshal golden: %v", err)
	}
	raw = append(raw, '\n')
	if err := os.WriteFile(path, raw, 0o644); err != nil {
		t.Fatalf("write golden: %v", err)
	}
}

func writeSnapshotIfNeeded(t *testing.T, path string, content []byte) {
	t.Helper()
	if os.Getenv("UPDATE_SNAPSHOT") == "" {
		return
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir snapshot dir: %v", err)
	}
	if err := os.WriteFile(path, content, 0o644); err != nil {
		t.Fatalf("write snapshot: %v", err)
	}
}

func demoArticleColumns(t *testing.T) []ColumnInfo {
	t.Helper()
	return loadTestJSON[[]ColumnInfo](t, "demo_article_columns.json")
}

func demoArticlePayloadRaw(t *testing.T) json.RawMessage {
	t.Helper()
	raw := mustReadFile(t, filepath.Join("testdata", "demo_article_payload.json"))
	return json.RawMessage(bytes.TrimSpace(raw))
}

func demoArticlePreviewGolden(t *testing.T) Preview {
	t.Helper()
	return loadTestJSON[Preview](t, "demo_article_preview.golden.json")
}

func demoArticleInput(t *testing.T, repoRoot string) GenerateInput {
	t.Helper()
	columns := demoArticleColumns(t)
	rawPayload := demoArticlePayloadRaw(t)
	preview := BuildPreview("demo_article", "demo_article", rawPayload, columns)
	return GenerateInput{
		ModuleName:     "demo_article",
		TableName:      "demo_article",
		Payload:        preview.Payload,
		Preview:        preview,
		Columns:        columns,
		Overwrite:      true,
		RegisterModule: true,
		UpsertMenu:     false,
	}
}

func newTempRepo(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	dirs := []string{
		"internal/modules",
		"internal/gen",
		"vben-admin/apps/backend/src/api",
		"vben-admin/apps/backend/src/views",
		"vben-admin/apps/backend/src/router/routes/modules",
	}
	for _, dir := range dirs {
		if err := os.MkdirAll(filepath.Join(root, dir), 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", dir, err)
		}
	}
	return root
}

func containsString(items []string, target string) bool {
	for _, item := range items {
		if item == target {
			return true
		}
	}
	return false
}
