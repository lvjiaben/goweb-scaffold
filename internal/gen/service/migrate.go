package service

import (
	"fmt"
	"strings"

	gentemplates "github.com/lvjiaben/goweb-scaffold/internal/gen/templates"
)

type MigrationResult struct {
	FromVersion string   `json:"from_version"`
	ToVersion   string   `json:"to_version"`
	Applied     []string `json:"applied"`
}

type sourceMigrator struct {
	From  string
	To    string
	Apply func(doc SourceDocument) (SourceDocument, []string, error)
}

var sourceMigrators = map[string]sourceMigrator{
	gentemplates.LegacyVersion: {
		From: gentemplates.LegacyVersion,
		To:   gentemplates.V6Version,
		Apply: func(doc SourceDocument) (SourceDocument, []string, error) {
			next := doc
			title := strings.TrimSpace(next.PreviewSummary.Page.MenuTitle)
			if title == "" {
				title = firstNonEmpty(
					strings.TrimSpace(next.Payload.Title),
					strings.TrimSpace(next.PreviewSummary.TableComment),
					HumanizeModuleName(next.ModuleName),
				)
			}
			next.PreviewSummary.Page.MenuTitle = title
			if len(next.PreviewSummary.Page.FeatureFlags) == 0 {
				next.PreviewSummary.Page.FeatureFlags = []string{"admin-crud", "codegen"}
			}
			next.TemplateVersion = gentemplates.V6Version
			return next, []string{"page.menu_title", "page.feature_flags"}, nil
		},
	},
	gentemplates.V6Version: {
		From: gentemplates.V6Version,
		To:   gentemplates.CurrentVersion,
		Apply: func(doc SourceDocument) (SourceDocument, []string, error) {
			next := doc
			next.Snapshot = emptySnapshot()
			next.TemplateVersion = gentemplates.CurrentVersion
			return next, []string{"snapshot.preview_hash", "snapshot.schema_hashes", "snapshot.generated_files"}, nil
		},
	},
}

func NormalizeTemplateVersion(version string) string {
	value := strings.TrimSpace(version)
	if value == "" {
		return gentemplates.LegacyVersion
	}
	return value
}

func MigrateSourceDocument(doc SourceDocument) (SourceDocument, MigrationResult, error) {
	current := NormalizeTemplateVersion(doc.TemplateVersion)
	result := MigrationResult{
		FromVersion: current,
		ToVersion:   current,
		Applied:     []string{},
	}
	if !gentemplates.IsSupported(current) {
		return SourceDocument{}, result, fmt.Errorf("unsupported template version: %s", current)
	}

	next := doc
	next.TemplateVersion = current
	for current != gentemplates.CurrentVersion {
		migrator, ok := sourceMigrators[current]
		if !ok {
			return SourceDocument{}, result, fmt.Errorf("no migrator registered for template version %s", current)
		}
		var applied []string
		var err error
		next, applied, err = migrator.Apply(next)
		if err != nil {
			return SourceDocument{}, result, err
		}
		result.Applied = append(result.Applied, applied...)
		current = migrator.To
	}
	result.ToVersion = current
	next.TemplateVersion = current
	return next, result, nil
}

func MigrateLockFile(lock LockFile) (LockFile, MigrationResult, error) {
	doc := SourceDocument{
		Kind:            "lock",
		GeneratedBy:     lock.GeneratedBy,
		ModuleName:      lock.ModuleName,
		TableName:       lock.TableName,
		TemplateVersion: lock.TemplateVersion,
		Payload:         lock.Payload,
		PreviewSummary:  lock.PreviewSummary,
		Snapshot:        lock.Snapshot,
		PermissionCodes: append([]string{}, lock.PermissionCodes...),
		RoutePath:       lock.RoutePath,
	}
	next, migration, err := MigrateSourceDocument(doc)
	if err != nil {
		return LockFile{}, migration, err
	}

	lock.TemplateVersion = next.TemplateVersion
	lock.Payload = next.Payload
	lock.PreviewSummary = next.PreviewSummary
	lock.Snapshot = next.Snapshot
	lock.PermissionCodes = append([]string{}, next.PermissionCodes...)
	lock.RoutePath = next.RoutePath
	return lock, migration, nil
}

func MigrateExportFile(file ExportFile) (ExportFile, MigrationResult, error) {
	doc := SourceDocument{
		Kind:            "export",
		GeneratedBy:     file.GeneratedBy,
		ModuleName:      file.ModuleName,
		TableName:       file.TableName,
		TemplateVersion: file.TemplateVersion,
		Payload:         file.Payload,
		PreviewSummary:  file.PreviewSummary,
		Snapshot:        file.Snapshot,
		PermissionCodes: append([]string{}, file.PermissionCodes...),
		RoutePath:       file.RoutePath,
	}
	next, migration, err := MigrateSourceDocument(doc)
	if err != nil {
		return ExportFile{}, migration, err
	}

	file.TemplateVersion = next.TemplateVersion
	file.Payload = next.Payload
	file.PreviewSummary = next.PreviewSummary
	file.Snapshot = next.Snapshot
	file.PermissionCodes = append([]string{}, next.PermissionCodes...)
	file.RoutePath = next.RoutePath
	if strings.TrimSpace(file.Format) == "" {
		file.Format = ExportFormatName
	}
	if strings.TrimSpace(file.Version) == "" {
		file.Version = ExportFormatVersion
	}
	return file, migration, nil
}
