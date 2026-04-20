package service

import (
	"encoding/json"
	"testing"
)

func TestBuildExportFromLockRoundTrip(t *testing.T) {
	lock := LockFile{
		GeneratedBy:     GeneratorName,
		ModuleName:      "demo_article",
		TableName:       "demo_article",
		GeneratedAt:     "2026-04-20T05:57:50+08:00",
		TemplateVersion: TemplateVersion,
		Payload: PayloadConfig{
			ListFields:   []string{"id", "title"},
			FormFields:   []string{"title"},
			SearchFields: []string{"title"},
		},
		PreviewSummary: LockPreviewSummary{
			TableComment: "演示文章",
			Page: PageMeta{
				RoutePath: "/system/demo-article",
				PageName:  "DemoArticlePage",
				ViewFile:  "views/system/DemoArticlePage.vue",
			},
			API: APIMeta{
				ModuleCode: "demo_article",
				List:       "/admin-api/demo_article/list",
				Detail:     "/admin-api/demo_article/detail",
				Save:       "/admin-api/demo_article/save",
				Delete:     "/admin-api/demo_article/delete",
			},
		},
		PermissionCodes: []string{"demo_article.list", "demo_article.save", "demo_article.delete"},
		RoutePath:       "/system/demo-article",
		GeneratedFiles:  []string{"internal/modules/demo_article/module.go"},
	}

	exportFile := BuildExportFromLock(lock)
	if exportFile.Format != ExportFormatName {
		t.Fatalf("expected export format %s, got %s", ExportFormatName, exportFile.Format)
	}

	raw, err := json.Marshal(exportFile)
	if err != nil {
		t.Fatalf("marshal export: %v", err)
	}
	doc, err := DecodeSourceDocument(raw)
	if err != nil {
		t.Fatalf("decode export document: %v", err)
	}
	if doc.Kind != "export" {
		t.Fatalf("expected export source kind, got %s", doc.Kind)
	}
	if doc.ModuleName != lock.ModuleName || doc.TableName != lock.TableName {
		t.Fatalf("unexpected export source %+v", doc)
	}
	if len(doc.PermissionCodes) != len(lock.PermissionCodes) {
		t.Fatalf("expected permission codes preserved, got %+v", doc.PermissionCodes)
	}

	lockRaw, err := json.Marshal(lock)
	if err != nil {
		t.Fatalf("marshal lock: %v", err)
	}
	lockDoc, err := DecodeSourceDocument(lockRaw)
	if err != nil {
		t.Fatalf("decode lock document: %v", err)
	}
	if lockDoc.Kind != "lock" {
		t.Fatalf("expected lock source kind, got %s", lockDoc.Kind)
	}
}

func TestDecodeSourceDocumentRejectsUnsupportedJSON(t *testing.T) {
	if _, err := DecodeSourceDocument([]byte(`{"module_name":"demo_article"}`)); err == nil {
		t.Fatalf("expected unsupported source error")
	}
}
