package service

import (
	"fmt"
	"os"
	"sort"
	"strings"
)

const (
	CompatibilitySame        = "same"
	CompatibilityNonBreaking = "non_breaking"
	CompatibilityBreaking    = "breaking"
)

func (s GeneratorService) CheckBreaking(input GenerateInput) (BreakingCheckResult, error) {
	moduleName := ToSnake(strings.TrimSpace(input.ModuleName))
	tableName := strings.TrimSpace(input.TableName)
	result := BreakingCheckResult{
		ModuleName:             moduleName,
		TableName:              tableName,
		CurrentTemplateVersion: TemplateVersion,
		Level:                  CompatibilitySame,
		ChangedAreas:           []string{},
		Reasons:                []string{},
		Warnings:               []string{},
		SnapshotDiff: SnapshotDiff{
			SchemaHashesChanged: map[string]bool{},
			FileChanges:         []SnapshotFileChange{},
		},
	}

	if moduleName == "" || tableName == "" {
		return result, fmt.Errorf("module_name and table_name are required")
	}

	paths := ModulePaths(moduleName)
	rawLock, err := s.readLockFile(paths["lock"])
	if err != nil {
		if os.IsNotExist(err) {
			result.Level = CompatibilityNonBreaking
			result.PreviousTemplateVersion = "missing"
			result.ChangedAreas = []string{"lock"}
			result.Reasons = []string{"existing codegen lock file not found; comparison falls back to current render only"}
			result.Warnings = []string{"snapshot missing, fallback to reconstructed comparison"}
			return result, nil
		}
		return result, err
	}

	result.PreviousTemplateVersion = NormalizeTemplateVersion(rawLock.TemplateVersion)
	lock, migration, err := MigrateLockFile(rawLock)
	if err != nil {
		return result, err
	}
	if len(migration.Applied) > 0 {
		result.Warnings = append(result.Warnings, "source migrated in memory before compatibility check")
	}

	bundle, err := s.prepareBundle(input)
	if err != nil {
		return result, err
	}
	currentSnapshot := buildCurrentSnapshot(bundle.Meta.ModuleName, bundle.Meta.TableName, bundle.Preview.Payload, bundle.Preview, bundle.Artifacts)

	previousSnapshot := lock.Snapshot
	if !hasSnapshot(previousSnapshot) {
		reconstructed, warnings := reconstructSnapshotFromLock(s.RepoRoot, lock)
		previousSnapshot = reconstructed
		result.Warnings = append(result.Warnings, warnings...)
	}

	compareBreakingCore(&result, lock, bundle)
	compareSnapshotDiff(&result, previousSnapshot, currentSnapshot)
	compareNonBreakingChanges(&result, lock, bundle)

	result.ChangedAreas = uniqueStrings(result.ChangedAreas)
	result.Reasons = uniqueStrings(result.Reasons)
	result.Warnings = uniqueStrings(result.Warnings)
	if len(result.Reasons) == 0 {
		result.Level = CompatibilitySame
	}
	return result, nil
}

func compareBreakingCore(result *BreakingCheckResult, previous LockFile, bundle generationBundle) {
	if previous.ModuleName != bundle.Meta.ModuleName || previous.TableName != bundle.Meta.TableName {
		markBreaking(result, "module_identity", fmt.Sprintf("module identity changed from %s/%s to %s/%s", previous.ModuleName, previous.TableName, bundle.Meta.ModuleName, bundle.Meta.TableName))
	}
	if previous.RoutePath != bundle.Meta.RoutePath {
		markBreaking(result, "route_path", fmt.Sprintf("route_path changed from %s to %s", previous.RoutePath, bundle.Meta.RoutePath))
	}
	if !equalStringSlices(previous.PermissionCodes, bundle.Meta.PermissionCodes) {
		markBreaking(result, "permission_codes", fmt.Sprintf("permission_codes changed from %v to %v", previous.PermissionCodes, bundle.Meta.PermissionCodes))
	}
	if previous.PreviewSummary.API.List != bundle.Preview.API.List ||
		previous.PreviewSummary.API.Detail != bundle.Preview.API.Detail ||
		previous.PreviewSummary.API.Save != bundle.Preview.API.Save ||
		previous.PreviewSummary.API.Delete != bundle.Preview.API.Delete {
		markBreaking(result, "api_paths", fmt.Sprintf("api paths changed from %+v to %+v", previous.PreviewSummary.API, bundle.Preview.API))
	}

	previousFields := schemaFieldSet(previous.PreviewSummary)
	currentFields := previewFieldSet(bundle.Preview)
	removedFields := []string{}
	for field := range previousFields {
		if _, ok := currentFields[field]; !ok {
			removedFields = append(removedFields, field)
		}
	}
	sort.Strings(removedFields)
	if len(removedFields) > 0 {
		markBreaking(result, "schema_fields", fmt.Sprintf("existing schema fields removed: %s", strings.Join(removedFields, ", ")))
	}

	currentPaths := make(map[string]struct{}, len(bundle.Generated))
	for _, item := range bundle.Generated {
		currentPaths[item] = struct{}{}
	}
	removedPaths := []string{}
	for _, path := range previous.GeneratedFiles {
		if strings.TrimSpace(path) == "" {
			continue
		}
		if _, ok := currentPaths[path]; !ok {
			removedPaths = append(removedPaths, path)
		}
	}
	sort.Strings(removedPaths)
	if len(removedPaths) > 0 {
		markBreaking(result, "generated_file_paths", fmt.Sprintf("generated file paths removed: %s", strings.Join(removedPaths, ", ")))
	}
}

func compareSnapshotDiff(result *BreakingCheckResult, previous Snapshot, current Snapshot) {
	result.SnapshotDiff.PreviewHashChanged = previous.PreviewHash != current.PreviewHash
	result.SnapshotDiff.SchemaHashesChanged["inferred_fields"] = previous.SchemaHashes.InferredFields != current.SchemaHashes.InferredFields
	result.SnapshotDiff.SchemaHashesChanged["form_schema"] = previous.SchemaHashes.FormSchema != current.SchemaHashes.FormSchema
	result.SnapshotDiff.SchemaHashesChanged["list_schema"] = previous.SchemaHashes.ListSchema != current.SchemaHashes.ListSchema
	result.SnapshotDiff.SchemaHashesChanged["search_schema"] = previous.SchemaHashes.SearchSchema != current.SchemaHashes.SearchSchema

	previousFiles := make(map[string]SnapshotFile, len(previous.Generated))
	currentFiles := make(map[string]SnapshotFile, len(current.Generated))
	for _, item := range previous.Generated {
		previousFiles[item.Path] = item
	}
	for _, item := range current.Generated {
		currentFiles[item.Path] = item
	}

	paths := make([]string, 0, len(previousFiles)+len(currentFiles))
	seen := map[string]struct{}{}
	for path := range previousFiles {
		paths = append(paths, path)
		seen[path] = struct{}{}
	}
	for path := range currentFiles {
		if _, ok := seen[path]; ok {
			continue
		}
		paths = append(paths, path)
	}
	sort.Strings(paths)

	for _, path := range paths {
		oldItem, hasOld := previousFiles[path]
		newItem, hasNew := currentFiles[path]
		switch {
		case !hasOld && hasNew:
			result.SnapshotDiff.FileChanges = append(result.SnapshotDiff.FileChanges, SnapshotFileChange{
				Path:    path,
				Status:  "added",
				NewHash: newItem.SHA256,
			})
		case hasOld && !hasNew:
			result.SnapshotDiff.FileChanges = append(result.SnapshotDiff.FileChanges, SnapshotFileChange{
				Path:    path,
				Status:  "removed",
				OldHash: oldItem.SHA256,
			})
		case oldItem.SHA256 != newItem.SHA256 || oldItem.Bytes != newItem.Bytes:
			result.SnapshotDiff.FileChanges = append(result.SnapshotDiff.FileChanges, SnapshotFileChange{
				Path:    path,
				Status:  "changed",
				OldHash: oldItem.SHA256,
				NewHash: newItem.SHA256,
			})
		}
	}
}

func compareNonBreakingChanges(result *BreakingCheckResult, previous LockFile, bundle generationBundle) {
	if result.Level == CompatibilityBreaking {
		return
	}

	if previous.PreviewSummary.Page.MenuTitle != bundle.Preview.Page.MenuTitle {
		markNonBreaking(result, "page_meta", fmt.Sprintf("menu_title changed from %s to %s", previous.PreviewSummary.Page.MenuTitle, bundle.Preview.Page.MenuTitle))
	}
	if !equalStringSlices(previous.PreviewSummary.Page.FeatureFlags, bundle.Preview.Page.FeatureFlags) {
		markNonBreaking(result, "page_meta", fmt.Sprintf("feature_flags changed from %v to %v", previous.PreviewSummary.Page.FeatureFlags, bundle.Preview.Page.FeatureFlags))
	}

	if result.SnapshotDiff.PreviewHashChanged {
		markNonBreaking(result, "preview_hash", "preview hash changed while route, permission codes and api paths remain compatible")
	}
	if result.SnapshotDiff.SchemaHashesChanged["inferred_fields"] ||
		result.SnapshotDiff.SchemaHashesChanged["form_schema"] ||
		result.SnapshotDiff.SchemaHashesChanged["list_schema"] ||
		result.SnapshotDiff.SchemaHashesChanged["search_schema"] {
		markNonBreaking(result, "schema_presentation", "schema hashes changed without removing existing fields")
	}
	if len(result.SnapshotDiff.FileChanges) > 0 {
		markNonBreaking(result, "generated_content", fmt.Sprintf("generated file content changed for %d file(s) without deleting old paths", len(result.SnapshotDiff.FileChanges)))
	}
}

func markBreaking(result *BreakingCheckResult, area string, reason string) {
	result.Level = CompatibilityBreaking
	result.ChangedAreas = append(result.ChangedAreas, area)
	result.Reasons = append(result.Reasons, reason)
}

func markNonBreaking(result *BreakingCheckResult, area string, reason string) {
	if result.Level != CompatibilityBreaking {
		result.Level = CompatibilityNonBreaking
	}
	result.ChangedAreas = append(result.ChangedAreas, area)
	result.Reasons = append(result.Reasons, reason)
}

func schemaFieldSet(summary LockPreviewSummary) map[string]struct{} {
	result := map[string]struct{}{}
	for _, item := range summary.InferredFields {
		result[item.ColumnName] = struct{}{}
	}
	for _, item := range summary.FormSchema {
		result[item.Field] = struct{}{}
	}
	for _, item := range summary.ListSchema {
		result[item.Field] = struct{}{}
	}
	for _, item := range summary.SearchSchema {
		result[item.Field] = struct{}{}
	}
	return result
}

func previewFieldSet(preview Preview) map[string]struct{} {
	return schemaFieldSet(LockPreviewSummary{
		InferredFields: preview.InferredFields,
		FormSchema:     preview.FormSchema,
		ListSchema:     preview.ListSchema,
		SearchSchema:   preview.SearchSchema,
	})
}

func equalStringSlices(left []string, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}
