package service

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestCheckBreakingSame(t *testing.T) {
	repoRoot := newTempRepo(t)
	svc := GeneratorService{RepoRoot: repoRoot}
	input := demoArticleInput(t, repoRoot)

	if _, err := svc.Generate(input); err != nil {
		t.Fatalf("generate baseline: %v", err)
	}

	result, err := svc.CheckBreaking(input)
	if err != nil {
		t.Fatalf("check breaking: %v", err)
	}
	if result.Level != CompatibilitySame {
		t.Fatalf("expected same, got %+v", result)
	}
}

func TestCheckBreakingNonBreakingForFieldPresentationChange(t *testing.T) {
	repoRoot := newTempRepo(t)
	svc := GeneratorService{RepoRoot: repoRoot}
	input := demoArticleInput(t, repoRoot)

	if _, err := svc.Generate(input); err != nil {
		t.Fatalf("generate baseline: %v", err)
	}

	input.Payload.FieldOverrides["title"] = FieldOverride{
		Label:       "文章名称",
		Placeholder: "请输入新的文章名称",
		Width:       "260px",
	}
	input.Preview = BuildPreview(input.ModuleName, input.TableName, mustMarshalPayload(input.Payload), input.Columns)

	result, err := svc.CheckBreaking(input)
	if err != nil {
		t.Fatalf("check breaking: %v", err)
	}
	if result.Level != CompatibilityNonBreaking {
		t.Fatalf("expected non_breaking, got %+v", result)
	}
	if !containsString(result.ChangedAreas, "schema_presentation") {
		t.Fatalf("expected schema_presentation change, got %+v", result)
	}
}

func TestCheckBreakingClassifiesRouteChangeAsBreaking(t *testing.T) {
	repoRoot := newTempRepo(t)
	svc := GeneratorService{RepoRoot: repoRoot}
	input := demoArticleInput(t, repoRoot)

	if _, err := svc.Generate(input); err != nil {
		t.Fatalf("generate baseline: %v", err)
	}

	lock, err := svc.readLockFile(filepath.Join("internal/modules/app", "demo_article", "codegen.lock.json"))
	if err != nil {
		t.Fatalf("read lock: %v", err)
	}
	lock, _, err = MigrateLockFile(lock)
	if err != nil {
		t.Fatalf("migrate lock: %v", err)
	}

	bundle, err := svc.prepareBundle(input)
	if err != nil {
		t.Fatalf("prepare bundle: %v", err)
	}
	bundle.Meta.RoutePath = "/system/demo-article-v2"
	bundle.Preview.Page.RoutePath = bundle.Meta.RoutePath

	result := BreakingCheckResult{
		ModuleName:              "demo_article",
		TableName:               "demo_article",
		PreviousTemplateVersion: lock.TemplateVersion,
		CurrentTemplateVersion:  TemplateVersion,
		Level:                   CompatibilitySame,
		ChangedAreas:            []string{},
		Reasons:                 []string{},
		SnapshotDiff: SnapshotDiff{
			SchemaHashesChanged: map[string]bool{},
			FileChanges:         []SnapshotFileChange{},
		},
	}
	compareBreakingCore(&result, lock, bundle)
	if result.Level != CompatibilityBreaking {
		t.Fatalf("expected breaking, got %+v", result)
	}
	if !containsString(result.ChangedAreas, "route_path") {
		t.Fatalf("expected route_path breaking change, got %+v", result)
	}
	if !strings.Contains(strings.Join(result.Reasons, "\n"), "route_path changed") {
		t.Fatalf("expected route_path reason, got %+v", result)
	}
}
