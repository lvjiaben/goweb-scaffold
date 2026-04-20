package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

const (
	ExportFormatName    = "codegen-export"
	ExportFormatVersion = "v1"
)

type ExportFile struct {
	GeneratedBy     string             `json:"generated_by"`
	Format          string             `json:"format"`
	Version         string             `json:"version"`
	ModuleName      string             `json:"module_name"`
	TableName       string             `json:"table_name"`
	TemplateVersion string             `json:"template_version,omitempty"`
	Payload         PayloadConfig      `json:"payload"`
	PreviewSummary  LockPreviewSummary `json:"preview_summary"`
	PermissionCodes []string           `json:"permission_codes"`
	RoutePath       string             `json:"route_path"`
}

type SourceDocument struct {
	Kind            string
	GeneratedBy     string
	ModuleName      string
	TableName       string
	TemplateVersion string
	Payload         PayloadConfig
	PreviewSummary  LockPreviewSummary
	PermissionCodes []string
	RoutePath       string
}

func BuildExportFromLock(lock LockFile) ExportFile {
	return ExportFile{
		GeneratedBy:     GeneratorName,
		Format:          ExportFormatName,
		Version:         ExportFormatVersion,
		ModuleName:      lock.ModuleName,
		TableName:       lock.TableName,
		TemplateVersion: lock.TemplateVersion,
		Payload:         lock.Payload,
		PreviewSummary:  lock.PreviewSummary,
		PermissionCodes: append([]string{}, lock.PermissionCodes...),
		RoutePath:       lock.RoutePath,
	}
}

func DecodeSourceDocument(raw []byte) (SourceDocument, error) {
	if len(raw) == 0 {
		return SourceDocument{}, errors.New("source file is empty")
	}

	var lock LockFile
	if err := json.Unmarshal(raw, &lock); err == nil && strings.TrimSpace(lock.ModuleName) != "" && strings.TrimSpace(lock.TableName) != "" {
		if strings.TrimSpace(lock.GeneratedBy) == GeneratorName && len(lock.GeneratedFiles) > 0 {
			doc := SourceDocument{
				Kind:            "lock",
				GeneratedBy:     lock.GeneratedBy,
				ModuleName:      lock.ModuleName,
				TableName:       lock.TableName,
				TemplateVersion: lock.TemplateVersion,
				Payload:         lock.Payload,
				PreviewSummary:  lock.PreviewSummary,
				PermissionCodes: append([]string{}, lock.PermissionCodes...),
				RoutePath:       lock.RoutePath,
			}
			next, _, migrateErr := MigrateSourceDocument(doc)
			return next, migrateErr
		}
	}

	var exportFile ExportFile
	if err := json.Unmarshal(raw, &exportFile); err == nil && strings.TrimSpace(exportFile.ModuleName) != "" && strings.TrimSpace(exportFile.TableName) != "" {
		if strings.TrimSpace(exportFile.Format) == ExportFormatName {
			doc := SourceDocument{
				Kind:            "export",
				GeneratedBy:     exportFile.GeneratedBy,
				ModuleName:      exportFile.ModuleName,
				TableName:       exportFile.TableName,
				TemplateVersion: exportFile.TemplateVersion,
				Payload:         exportFile.Payload,
				PreviewSummary:  exportFile.PreviewSummary,
				PermissionCodes: append([]string{}, exportFile.PermissionCodes...),
				RoutePath:       exportFile.RoutePath,
			}
			next, _, migrateErr := MigrateSourceDocument(doc)
			return next, migrateErr
		}
	}

	return SourceDocument{}, fmt.Errorf("unsupported source document")
}
