package service

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	gentemplates "github.com/lvjiaben/goweb-scaffold/internal/gen/templates"
)

func TestMigrateLegacyExportV5ToV6(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("testdata", "legacy", "demo_article_export_v5.json"))
	if err != nil {
		t.Fatalf("read legacy export: %v", err)
	}
	var file ExportFile
	if err := json.Unmarshal(raw, &file); err != nil {
		t.Fatalf("decode legacy export: %v", err)
	}

	next, migration, err := MigrateExportFile(file)
	if err != nil {
		t.Fatalf("migrate export: %v", err)
	}
	if migration.FromVersion != gentemplates.LegacyVersion || migration.ToVersion != gentemplates.CurrentVersion {
		t.Fatalf("unexpected migration result: %+v", migration)
	}
	if next.TemplateVersion != gentemplates.CurrentVersion {
		t.Fatalf("expected export template version upgraded, got %s", next.TemplateVersion)
	}
	if next.PreviewSummary.Page.MenuTitle == "" {
		t.Fatalf("expected menu_title filled after migration")
	}
	if len(next.PreviewSummary.Page.FeatureFlags) == 0 {
		t.Fatalf("expected feature_flags filled after migration")
	}
	if next.Snapshot.Generated == nil {
		t.Fatalf("expected snapshot structure initialized")
	}
}

func TestMigrateLegacyLockV5ToV6AndPreviewStillWorks(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("testdata", "legacy", "demo_article_lock_v5.json"))
	if err != nil {
		t.Fatalf("read legacy lock: %v", err)
	}
	var lock LockFile
	if err := json.Unmarshal(raw, &lock); err != nil {
		t.Fatalf("decode legacy lock: %v", err)
	}

	next, migration, err := MigrateLockFile(lock)
	if err != nil {
		t.Fatalf("migrate lock: %v", err)
	}
	if migration.ToVersion != gentemplates.CurrentVersion {
		t.Fatalf("expected current version, got %+v", migration)
	}

	columns := readDemoArticleColumnsFixture(t)
	preview := BuildPreview(next.ModuleName, next.TableName, mustMarshalPayload(next.Payload), columns)
	if preview.Page.MenuTitle == "" {
		t.Fatalf("expected preview menu_title after migration")
	}
	if len(preview.Page.FeatureFlags) == 0 {
		t.Fatalf("expected preview feature_flags after migration")
	}
}

func TestMigrateLegacyExportV6ToV7(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("testdata", "legacy", "demo_article_export_v6.json"))
	if err != nil {
		t.Fatalf("read legacy export v6: %v", err)
	}
	var file ExportFile
	if err := json.Unmarshal(raw, &file); err != nil {
		t.Fatalf("decode legacy export v6: %v", err)
	}

	next, migration, err := MigrateExportFile(file)
	if err != nil {
		t.Fatalf("migrate export v6: %v", err)
	}
	if migration.FromVersion != gentemplates.V6Version || migration.ToVersion != gentemplates.CurrentVersion {
		t.Fatalf("unexpected migration result: %+v", migration)
	}
	if next.TemplateVersion != gentemplates.CurrentVersion {
		t.Fatalf("expected export template version upgraded, got %s", next.TemplateVersion)
	}
	rawNext, _ := json.Marshal(next)
	if !bytes.Contains(rawNext, []byte(`"snapshot"`)) {
		t.Fatalf("expected snapshot in migrated export json")
	}
}

func TestMigrateLegacyLockV6ToV7(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("testdata", "legacy", "demo_article_lock_v6.json"))
	if err != nil {
		t.Fatalf("read legacy lock v6: %v", err)
	}
	var lock LockFile
	if err := json.Unmarshal(raw, &lock); err != nil {
		t.Fatalf("decode legacy lock v6: %v", err)
	}

	next, migration, err := MigrateLockFile(lock)
	if err != nil {
		t.Fatalf("migrate lock v6: %v", err)
	}
	if migration.FromVersion != gentemplates.V6Version || migration.ToVersion != gentemplates.CurrentVersion {
		t.Fatalf("unexpected migration result: %+v", migration)
	}
	rawNext, _ := json.Marshal(next)
	if !bytes.Contains(rawNext, []byte(`"snapshot"`)) {
		t.Fatalf("expected snapshot in migrated lock json")
	}
}

func TestMigrateCurrentSourceNoop(t *testing.T) {
	doc := SourceDocument{
		Kind:            "export",
		GeneratedBy:     GeneratorName,
		ModuleName:      "demo_article",
		TableName:       "demo_article",
		TemplateVersion: gentemplates.CurrentVersion,
		Payload: PayloadConfig{
			ListFields: []string{"id"},
		},
		PreviewSummary: LockPreviewSummary{
			Page: PageMeta{
				RoutePath:    "/system/demo-article",
				PageName:     "DemoArticlePage",
				ViewFile:     "views/system/DemoArticlePage.vue",
				MenuTitle:    "演示文章",
				FeatureFlags: []string{"admin-crud", "codegen"},
			},
		},
	}

	next, migration, err := MigrateSourceDocument(doc)
	if err != nil {
		t.Fatalf("migrate current source: %v", err)
	}
	if migration.FromVersion != gentemplates.CurrentVersion || migration.ToVersion != gentemplates.CurrentVersion {
		t.Fatalf("unexpected noop migration: %+v", migration)
	}
	if len(migration.Applied) != 0 {
		t.Fatalf("expected no applied steps, got %+v", migration.Applied)
	}
	if next.PreviewSummary.Page.MenuTitle != "演示文章" {
		t.Fatalf("unexpected noop source change: %+v", next.PreviewSummary.Page)
	}
}

func readDemoArticleColumnsFixture(t *testing.T) []ColumnInfo {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("testdata", "demo_article_columns.json"))
	if err != nil {
		t.Fatalf("read columns fixture: %v", err)
	}
	var columns []ColumnInfo
	if err := json.Unmarshal(raw, &columns); err != nil {
		t.Fatalf("decode columns fixture: %v", err)
	}
	return columns
}
