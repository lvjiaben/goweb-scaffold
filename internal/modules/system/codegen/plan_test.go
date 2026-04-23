package codegen

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadPlanDefaultsAndOverrides(t *testing.T) {
	planPath := filepath.Join(t.TempDir(), "codegen.plan.json")
	raw := []byte(`{
  "generated_by": "goweb-scaffold",
  "format": "codegen-plan",
  "version": "v1",
  "defaults": {
    "overwrite": true,
    "register_module": true,
    "upsert_menu": false
  },
  "modules": [
    {
      "module_name": "demo_article",
      "table_name": "demo_article"
    },
    {
      "module_name": "demo_notice",
      "table_name": "demo_notice",
      "upsert_menu": true
    }
  ]
}`)
	if err := os.WriteFile(planPath, raw, 0o644); err != nil {
		t.Fatalf("write plan: %v", err)
	}

	plan, err := LoadPlan(planPath)
	if err != nil {
		t.Fatalf("load plan: %v", err)
	}
	runner := newFixtureRunner(t)

	first, err := runner.resolvePlanEntry(plan, plan.Modules[0])
	if err != nil {
		t.Fatalf("resolve first entry: %v", err)
	}
	if !first.Overwrite || !first.RegisterModule || first.UpsertMenu {
		t.Fatalf("unexpected first defaults: %+v", first)
	}

	second, err := runner.resolvePlanEntry(plan, plan.Modules[1])
	if err != nil {
		t.Fatalf("resolve second entry: %v", err)
	}
	if !second.UpsertMenu {
		t.Fatalf("expected entry override to win, got %+v", second)
	}
}

func TestResolvePlanEntryFromAndExplicitOverride(t *testing.T) {
	runner := newFixtureRunner(t)
	exportPath := writeFixtureExportFile(t)
	plan := BatchPlan{
		Format:  PlanFormatName,
		Version: PlanFormatVersion,
		Modules: []PlanModuleEntry{{
			From:      exportPath,
			TableName: "demo_notice",
			Payload:   []byte(`{"title":"公告模块"}`),
		}},
	}

	resolved, err := runner.resolvePlanEntry(plan, plan.Modules[0])
	if err != nil {
		t.Fatalf("resolve plan entry: %v", err)
	}
	if resolved.TableName != "demo_notice" {
		t.Fatalf("expected explicit table override, got %+v", resolved)
	}
	if string(resolved.Payload) != `{"title":"公告模块"}` {
		t.Fatalf("expected explicit payload override, got %s", string(resolved.Payload))
	}
	if resolved.SourceKind != "export" {
		t.Fatalf("expected export source kind, got %+v", resolved)
	}
}

func TestValidatePlanRejectsMissingFields(t *testing.T) {
	plan := BatchPlan{
		Format:  PlanFormatName,
		Version: PlanFormatVersion,
		Modules: []PlanModuleEntry{{}},
	}
	if err := ValidatePlan(plan); err == nil {
		t.Fatalf("expected missing field validation error")
	}
}

func TestValidatePlanRejectsMissingTableNameWhenFromEmpty(t *testing.T) {
	plan := BatchPlan{
		Format:  PlanFormatName,
		Version: PlanFormatVersion,
		Modules: []PlanModuleEntry{{
			ModuleName: "demo_article",
		}},
	}
	if err := ValidatePlan(plan); err == nil {
		t.Fatalf("expected missing table_name validation error")
	}
}
