package codegen

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

const (
	PlanFormatName    = "codegen-plan"
	PlanFormatVersion = "v1"
)

type PlanDefaults struct {
	Overwrite        *bool `json:"overwrite,omitempty"`
	RegisterModule   *bool `json:"register_module,omitempty"`
	UpsertMenu       *bool `json:"upsert_menu,omitempty"`
	RemoveFiles      *bool `json:"remove_files,omitempty"`
	UnregisterModule *bool `json:"unregister_module,omitempty"`
	RemoveMenu       *bool `json:"remove_menu,omitempty"`
	RemoveHistory    *bool `json:"remove_history,omitempty"`
	RemoveLock       *bool `json:"remove_lock,omitempty"`
}

type PlanModuleEntry struct {
	ModuleName       string          `json:"module_name"`
	TableName        string          `json:"table_name"`
	From             string          `json:"from"`
	Payload          json.RawMessage `json:"payload"`
	Overwrite        *bool           `json:"overwrite,omitempty"`
	RegisterModule   *bool           `json:"register_module,omitempty"`
	UpsertMenu       *bool           `json:"upsert_menu,omitempty"`
	RemoveFiles      *bool           `json:"remove_files,omitempty"`
	UnregisterModule *bool           `json:"unregister_module,omitempty"`
	RemoveMenu       *bool           `json:"remove_menu,omitempty"`
	RemoveHistory    *bool           `json:"remove_history,omitempty"`
	RemoveLock       *bool           `json:"remove_lock,omitempty"`
}

type BatchPlan struct {
	GeneratedBy string            `json:"generated_by"`
	Format      string            `json:"format"`
	Version     string            `json:"version"`
	Defaults    PlanDefaults      `json:"defaults"`
	Modules     []PlanModuleEntry `json:"modules"`
}

func LoadPlan(path string) (BatchPlan, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return BatchPlan{}, err
	}
	var plan BatchPlan
	if err := json.Unmarshal(raw, &plan); err != nil {
		return BatchPlan{}, err
	}
	if strings.TrimSpace(plan.Format) == "" {
		plan.Format = PlanFormatName
	}
	if strings.TrimSpace(plan.Version) == "" {
		plan.Version = PlanFormatVersion
	}
	if err := ValidatePlan(plan); err != nil {
		return BatchPlan{}, err
	}
	return plan, nil
}

func ValidatePlan(plan BatchPlan) error {
	if strings.TrimSpace(plan.Format) != PlanFormatName {
		return fmt.Errorf("unsupported plan format: %s", plan.Format)
	}
	if strings.TrimSpace(plan.Version) != PlanFormatVersion {
		return fmt.Errorf("unsupported plan version: %s", plan.Version)
	}
	if len(plan.Modules) == 0 {
		return fmt.Errorf("plan modules cannot be empty")
	}
	for index, entry := range plan.Modules {
		if strings.TrimSpace(entry.ModuleName) == "" && strings.TrimSpace(entry.From) == "" {
			return fmt.Errorf("plan module[%d] requires module_name or from", index)
		}
		if strings.TrimSpace(entry.From) == "" && strings.TrimSpace(entry.TableName) == "" {
			return fmt.Errorf("plan module[%d] requires table_name when from is empty", index)
		}
	}
	return nil
}

func (plan BatchPlan) EffectiveBool(entryValue *bool, defaultValue *bool, fallback bool) bool {
	if entryValue != nil {
		return *entryValue
	}
	if defaultValue != nil {
		return *defaultValue
	}
	return fallback
}
