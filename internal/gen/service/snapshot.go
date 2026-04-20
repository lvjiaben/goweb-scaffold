package service

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func emptySnapshot() Snapshot {
	return Snapshot{
		SchemaHashes: SnapshotSchemaHashes{},
		Generated:    []SnapshotFile{},
	}
}

func hasSnapshot(snapshot Snapshot) bool {
	return strings.TrimSpace(snapshot.PreviewHash) != "" ||
		strings.TrimSpace(snapshot.SchemaHashes.InferredFields) != "" ||
		strings.TrimSpace(snapshot.SchemaHashes.FormSchema) != "" ||
		strings.TrimSpace(snapshot.SchemaHashes.ListSchema) != "" ||
		strings.TrimSpace(snapshot.SchemaHashes.SearchSchema) != "" ||
		len(snapshot.Generated) > 0
}

func buildPreviewSnapshot(moduleName string, tableName string, payload PayloadConfig, preview LockPreviewSummary) Snapshot {
	return Snapshot{
		PreviewHash: hashJSON(map[string]any{
			"module_name":   moduleName,
			"table_name":    tableName,
			"payload":       payload,
			"table_comment": preview.TableComment,
			"page":          preview.Page,
			"api":           preview.API,
		}),
		SchemaHashes: SnapshotSchemaHashes{
			InferredFields: hashJSON(preview.InferredFields),
			FormSchema:     hashJSON(preview.FormSchema),
			ListSchema:     hashJSON(preview.ListSchema),
			SearchSchema:   hashJSON(preview.SearchSchema),
		},
		Generated: []SnapshotFile{},
	}
}

func buildCurrentSnapshot(moduleName string, tableName string, payload PayloadConfig, preview Preview, artifacts []generatedArtifact) Snapshot {
	summary := LockPreviewSummary{
		TableComment:   preview.TableComment,
		Page:           preview.Page,
		API:            preview.API,
		InferredFields: preview.InferredFields,
		FormSchema:     preview.FormSchema,
		ListSchema:     preview.ListSchema,
		SearchSchema:   preview.SearchSchema,
	}
	snapshot := buildPreviewSnapshot(moduleName, tableName, payload, summary)

	files := make([]SnapshotFile, 0, len(artifacts))
	for _, item := range artifacts {
		if strings.HasSuffix(item.Path, "codegen.lock.json") {
			continue
		}
		files = append(files, SnapshotFile{
			Path:   item.Path,
			SHA256: sha256Hex(item.Content),
			Bytes:  len(item.Content),
		})
	}
	sort.Slice(files, func(i, j int) bool {
		return files[i].Path < files[j].Path
	})
	snapshot.Generated = files
	return snapshot
}

func reconstructSnapshotFromLock(repoRoot string, lock LockFile) (Snapshot, []string) {
	snapshot := buildPreviewSnapshot(lock.ModuleName, lock.TableName, lock.Payload, lock.PreviewSummary)
	warnings := []string{"snapshot missing, fallback to reconstructed comparison"}
	files := make([]SnapshotFile, 0, len(lock.GeneratedFiles))
	for _, relPath := range lock.GeneratedFiles {
		if strings.TrimSpace(relPath) == "" || strings.HasSuffix(relPath, "codegen.lock.json") {
			continue
		}
		fullPath := filepath.Join(repoRoot, filepath.Clean(relPath))
		content, err := os.ReadFile(fullPath)
		if err != nil {
			warnings = append(warnings, "failed to read generated file: "+relPath)
			continue
		}
		files = append(files, SnapshotFile{
			Path:   relPath,
			SHA256: sha256Hex(content),
			Bytes:  len(content),
		})
	}
	sort.Slice(files, func(i, j int) bool {
		return files[i].Path < files[j].Path
	})
	snapshot.Generated = files
	return snapshot, uniqueStrings(warnings)
}

func hashJSON(value any) string {
	raw, err := json.Marshal(value)
	if err != nil {
		return ""
	}
	return sha256Hex(raw)
}

func sha256Hex(content []byte) string {
	if len(content) == 0 {
		return ""
	}
	sum := sha256.Sum256(content)
	return hex.EncodeToString(sum[:])
}
