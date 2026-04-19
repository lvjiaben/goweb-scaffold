package service

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestBuildPreviewGolden(t *testing.T) {
	columns := demoArticleColumns(t)
	rawPayload := demoArticlePayloadRaw(t)

	actual := BuildPreview("demo_article", "demo_article", rawPayload, columns)
	goldenPath := filepath.Join("testdata", "demo_article_preview.golden.json")
	writeGoldenIfNeeded(t, goldenPath, actual)

	expected := demoArticlePreviewGolden(t)

	actualRaw, err := json.MarshalIndent(actual, "", "  ")
	if err != nil {
		t.Fatalf("marshal actual preview: %v", err)
	}
	expectedRaw, err := json.MarshalIndent(expected, "", "  ")
	if err != nil {
		t.Fatalf("marshal expected preview: %v", err)
	}
	if string(actualRaw) != string(expectedRaw) {
		t.Fatalf("preview golden mismatch\nexpected:\n%s\nactual:\n%s", string(expectedRaw), string(actualRaw))
	}
	if os.Getenv("UPDATE_GOLDEN") != "" {
		t.Log("preview golden updated")
	}
}
