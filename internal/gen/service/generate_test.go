package service

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDiffCreatesFilesOnFirstRun(t *testing.T) {
	repoRoot := newTempRepo(t)
	input := demoArticleInput(t, repoRoot)

	result, err := (GeneratorService{RepoRoot: repoRoot}).Diff(input)
	if err != nil {
		t.Fatalf("diff first run: %v", err)
	}

	if len(result.WouldCreateFiles) == 0 {
		t.Fatalf("expected create files on first diff")
	}
	if !containsString(result.WouldCreateFiles, "internal/modules/demo_article/model.go") {
		t.Fatalf("expected model.go in create files, got %v", result.WouldCreateFiles)
	}
	if !containsString(result.WouldCreateFiles, "internal/modules/demo_article/codegen.lock.json") {
		t.Fatalf("expected codegen.lock.json in create files, got %v", result.WouldCreateFiles)
	}
}

func TestDiffSkipsExistingGeneratedFilesWhenOverwriteDisabled(t *testing.T) {
	repoRoot := newTempRepo(t)
	svc := GeneratorService{RepoRoot: repoRoot}
	input := demoArticleInput(t, repoRoot)

	if _, err := svc.Generate(input); err != nil {
		t.Fatalf("generate initial files: %v", err)
	}

	input.Overwrite = false
	result, err := svc.Diff(input)
	if err != nil {
		t.Fatalf("diff existing files: %v", err)
	}
	if !containsString(result.WouldSkipFiles, "internal/modules/demo_article/codegen.lock.json") {
		t.Fatalf("expected lock file in skipped files, got %v", result.WouldSkipFiles)
	}
}

func TestDiffMarksGeneratedFilesAsOverwriteWhenContentChanges(t *testing.T) {
	repoRoot := newTempRepo(t)
	svc := GeneratorService{RepoRoot: repoRoot}
	input := demoArticleInput(t, repoRoot)
	if _, err := svc.Generate(input); err != nil {
		t.Fatalf("generate initial files: %v", err)
	}

	input.Payload.FieldOverrides["summary"] = FieldOverride{
		Label:       "文章摘要",
		Placeholder: "新的摘要提示",
	}
	input.Preview = BuildPreview(input.ModuleName, input.TableName, mustMarshalPayload(input.Payload), input.Columns)

	result, err := svc.Diff(input)
	if err != nil {
		t.Fatalf("diff changed generated files: %v", err)
	}
	if !containsString(result.WouldOverwriteFiles, "internal/modules/demo_article/codegen.lock.json") {
		t.Fatalf("expected generated lock in overwrite files, got %v", result.WouldOverwriteFiles)
	}
}

func TestDiffSkipsHandwrittenFiles(t *testing.T) {
	repoRoot := newTempRepo(t)
	viewPath := filepath.Join(repoRoot, "vben-admin/apps/backend/src/views/demo_article/list.vue")
	if err := os.MkdirAll(filepath.Dir(viewPath), 0o755); err != nil {
		t.Fatalf("mkdir handwritten view dir: %v", err)
	}
	if err := os.WriteFile(viewPath, []byte("<template>handwritten</template>\n"), 0o644); err != nil {
		t.Fatalf("write handwritten view: %v", err)
	}

	input := demoArticleInput(t, repoRoot)
	result, err := (GeneratorService{RepoRoot: repoRoot}).Diff(input)
	if err != nil {
		t.Fatalf("diff handwritten file: %v", err)
	}
	if !containsString(result.WouldSkipFiles, "vben-admin/apps/backend/src/views/demo_article/list.vue") {
		t.Fatalf("expected handwritten view in skipped files, got %v", result.WouldSkipFiles)
	}
	if len(result.Warnings) == 0 || !strings.Contains(strings.Join(result.Warnings, "\n"), "not generator-managed") {
		t.Fatalf("expected handwritten warning, got %v", result.Warnings)
	}
}

func TestGenerateRegenerateKeepsLockStable(t *testing.T) {
	repoRoot := newTempRepo(t)
	svc := GeneratorService{RepoRoot: repoRoot}
	input := demoArticleInput(t, repoRoot)

	first, err := svc.Generate(input)
	if err != nil {
		t.Fatalf("generate first: %v", err)
	}
	if !containsString(first.GeneratedFiles, "internal/modules/demo_article/codegen.lock.json") {
		t.Fatalf("expected lock file on first generate, got %v", first.GeneratedFiles)
	}

	lockOne, err := svc.LoadLock("demo_article")
	if err != nil {
		t.Fatalf("load first lock: %v", err)
	}

	second, err := svc.Generate(input)
	if err != nil {
		t.Fatalf("generate second: %v", err)
	}
	if !containsString(second.SkippedFiles, "internal/modules/demo_article/codegen.lock.json") {
		t.Fatalf("expected lock file skipped on second generate, got %v", second.SkippedFiles)
	}

	lockTwo, err := svc.LoadLock("demo_article")
	if err != nil {
		t.Fatalf("load second lock: %v", err)
	}
	if !sameLockCore(lockOne, lockTwo) {
		t.Fatalf("expected lock core to remain stable")
	}
	if lockOne.GeneratedAt != lockTwo.GeneratedAt {
		t.Fatalf("expected generated_at to remain stable, got %s vs %s", lockOne.GeneratedAt, lockTwo.GeneratedAt)
	}
}

func TestRemoveGeneratedModuleFilesAndRegistry(t *testing.T) {
	repoRoot := newTempRepo(t)
	svc := GeneratorService{RepoRoot: repoRoot}
	input := demoArticleInput(t, repoRoot)

	if _, err := svc.Generate(input); err != nil {
		t.Fatalf("generate before remove: %v", err)
	}

	result, err := svc.Remove(RemoveInput{
		ModuleName:       "demo_article",
		RemoveFiles:      true,
		UnregisterModule: true,
		RemoveLock:       true,
	})
	if err != nil {
		t.Fatalf("remove generated module: %v", err)
	}
	if !containsString(result.RemovedFiles, "internal/modules/demo_article/model.go") {
		t.Fatalf("expected model.go removed, got %v", result.RemovedFiles)
	}
	if _, err := os.Stat(filepath.Join(repoRoot, "internal/modules/demo_article/model.go")); !os.IsNotExist(err) {
		t.Fatalf("expected model.go removed from disk, stat err=%v", err)
	}

	modulesGen := string(mustReadFile(t, filepath.Join(repoRoot, "internal/gen/modules_gen.go")))
	if strings.Contains(modulesGen, "demo_article") {
		t.Fatalf("expected modules_gen.go to exclude demo_article, got:\n%s", modulesGen)
	}
	routes := string(mustReadFile(t, filepath.Join(repoRoot, "vben-admin/apps/backend/src/router/routes/modules/generated.ts")))
	if strings.Contains(routes, "demo_article") {
		t.Fatalf("expected generated routes to exclude demo_article, got:\n%s", routes)
	}
}
