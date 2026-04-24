package service

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/lvjiaben/goweb-scaffold/internal/gen/registry"
	"github.com/lvjiaben/goweb-scaffold/internal/gen/writer"
)

func (s GeneratorService) ListModules() ([]ManagedModule, error) {
	pattern := filepath.Join(s.RepoRoot, "internal/modules/app/*/codegen.lock.json")
	files, err := filepath.Glob(pattern)
	if err != nil {
		return nil, err
	}
	sort.Strings(files)

	items := make([]ManagedModule, 0, len(files))
	for _, file := range files {
		raw, err := os.ReadFile(file)
		if err != nil {
			return nil, err
		}

		var lock LockFile
		if err := json.Unmarshal(raw, &lock); err != nil {
			return nil, err
		}
		if migrated, _, err := MigrateLockFile(lock); err == nil {
			lock = migrated
		}
		if strings.TrimSpace(lock.ModuleName) == "" {
			continue
		}

		items = append(items, ManagedModule{
			ModuleName:      lock.ModuleName,
			TableName:       lock.TableName,
			GeneratedAt:     lock.GeneratedAt,
			TemplateVersion: lock.TemplateVersion,
			RoutePath:       lock.RoutePath,
			PermissionCodes: append([]string{}, lock.PermissionCodes...),
			Files:           append([]string{}, lock.GeneratedFiles...),
			Payload:         lock.Payload,
			PreviewSummary:  lock.PreviewSummary,
			Snapshot:        lock.Snapshot,
		})
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].ModuleName < items[j].ModuleName
	})
	return items, nil
}

func (s GeneratorService) LoadLock(moduleName string) (LockFile, error) {
	paths := ModulePaths(ToSnake(moduleName))
	lock, err := s.readLockFile(paths["lock"])
	if err != nil {
		return LockFile{}, err
	}
	lock, _, err = MigrateLockFile(lock)
	if err != nil {
		return LockFile{}, err
	}
	if strings.TrimSpace(lock.ModuleName) == "" || strings.TrimSpace(lock.TableName) == "" {
		return LockFile{}, fmt.Errorf("invalid codegen lock file for module %s", moduleName)
	}
	return lock, nil
}

func (s GeneratorService) Remove(input RemoveInput) (RemoveResult, error) {
	moduleName := ToSnake(strings.TrimSpace(input.ModuleName))
	result := RemoveResult{
		ModuleName:               moduleName,
		RemovedFiles:             []string{},
		SkippedFiles:             []string{},
		RemovedMenuRecords:       []map[string]any{},
		RemovedHistoryIDs:        []int64{},
		RegeneratedRegistryFiles: []string{},
		Warnings:                 []string{},
	}
	if moduleName == "" {
		return result, fmt.Errorf("module_name is required")
	}

	lock, lockErr := s.LoadLock(moduleName)
	if lockErr != nil && !os.IsNotExist(lockErr) {
		return result, lockErr
	}
	if os.IsNotExist(lockErr) {
		result.Warnings = append(result.Warnings, fmt.Sprintf("module %s lock file not found", moduleName))
	}

	if input.RemoveFiles || input.RemoveLock {
		fileResult, warnings, err := s.removeGeneratedFiles(moduleName, lock, input)
		if err != nil {
			return result, err
		}
		result.RemovedFiles = append(result.RemovedFiles, fileResult.removed...)
		result.SkippedFiles = append(result.SkippedFiles, fileResult.skipped...)
		result.Warnings = append(result.Warnings, warnings...)
	}

	if input.UnregisterModule {
		files, warnings, err := s.rebuildRegistryExcluding(moduleName)
		if err != nil {
			return result, err
		}
		result.RegeneratedRegistryFiles = append(result.RegeneratedRegistryFiles, files...)
		result.Warnings = append(result.Warnings, warnings...)
	}

	if input.RemoveMenu {
		menuResult, warnings, err := s.removeMenus(moduleName, lock)
		if err != nil {
			return result, err
		}
		result.RemovedMenuRecords = append(result.RemovedMenuRecords, menuResult.records...)
		result.RemovedRoleMenuLinks = menuResult.roleMenuLinks
		result.Warnings = append(result.Warnings, warnings...)
	}

	if input.RemoveHistory {
		historyIDs, warnings, err := s.removeHistory(moduleName)
		if err != nil {
			return result, err
		}
		result.RemovedHistoryIDs = append(result.RemovedHistoryIDs, historyIDs...)
		result.Warnings = append(result.Warnings, warnings...)
	}

	result.Warnings = uniqueStrings(result.Warnings)
	result.SkippedFiles = uniqueStrings(result.SkippedFiles)
	return result, nil
}

type removeFileResult struct {
	removed []string
	skipped []string
}

func (s GeneratorService) removeGeneratedFiles(moduleName string, lock LockFile, input RemoveInput) (removeFileResult, []string, error) {
	result := removeFileResult{
		removed: []string{},
		skipped: []string{},
	}
	warnings := []string{}
	paths := ModulePaths(moduleName)

	candidates := []string{}
	if input.RemoveFiles {
		candidates = append(candidates,
			paths["module"],
			paths["backend_handler"],
			paths["api_handler"],
			paths["service"],
			paths["repo"],
			paths["backend_dto"],
			paths["api_dto"],
			paths["model"],
			paths["meta"],
			paths["api"],
			paths["view_list"],
			paths["view_data"],
			paths["view_form"],
		)
	}
	if input.RemoveLock {
		candidates = append(candidates, paths["lock"])
	}
	candidates = uniqueStrings(candidates)

	lockManaged := make(map[string]struct{}, len(lock.GeneratedFiles))
	for _, item := range lock.GeneratedFiles {
		lockManaged[item] = struct{}{}
	}

	fileWriter := writer.New(s.RepoRoot)
	for _, relPath := range candidates {
		if relPath == "" {
			continue
		}
		if len(lockManaged) > 0 {
			if _, ok := lockManaged[relPath]; !ok && relPath != paths["lock"] {
				result.skipped = append(result.skipped, relPath)
				warnings = append(warnings, fmt.Sprintf("%s is not tracked in codegen lock", relPath))
				continue
			}
		}
		status, warning, err := fileWriter.Delete(relPath)
		if err != nil {
			return result, warnings, err
		}
		switch status {
		case "removed":
			result.removed = append(result.removed, relPath)
		default:
			result.skipped = append(result.skipped, relPath)
		}
		if warning != "" {
			warnings = append(warnings, warning)
		}
	}
	return result, warnings, nil
}

func (s GeneratorService) rebuildRegistryExcluding(moduleName string) ([]string, []string, error) {
	files := []string{}
	warnings := []string{}
	fileWriter := writer.New(s.RepoRoot)

	backendContent, err := registry.RenderBackendModulesFileWithOptions(s.RepoRoot, nil, []string{moduleName})
	if err != nil {
		return nil, nil, err
	}
	if status, warning, err := fileWriter.Write("internal/gen/modules_gen.go", backendContent, true); err != nil {
		return nil, nil, err
	} else {
		files = append(files, "internal/gen/modules_gen.go")
		if status == "skipped" && warning == "" {
			warnings = append(warnings, "internal/gen/modules_gen.go already up to date")
		}
		if warning != "" {
			warnings = append(warnings, warning)
		}
	}

	frontendContent, err := registry.RenderFrontendRouteModuleWithOptions(s.RepoRoot, nil, []string{moduleName})
	if err != nil {
		return nil, nil, err
	}
	if status, warning, err := fileWriter.Write("vben-admin/apps/backend/src/router/routes/modules/generated.ts", frontendContent, true); err != nil {
		return nil, nil, err
	} else {
		files = append(files, "vben-admin/apps/backend/src/router/routes/modules/generated.ts")
		if status == "skipped" && warning == "" {
			warnings = append(warnings, "vben-admin/apps/backend/src/router/routes/modules/generated.ts already up to date")
		}
		if warning != "" {
			warnings = append(warnings, warning)
		}
	}

	return files, warnings, nil
}

type removeMenuResult struct {
	records       []map[string]any
	roleMenuLinks int64
}

func (s GeneratorService) removeMenus(moduleName string, lock LockFile) (removeMenuResult, []string, error) {
	result := removeMenuResult{records: []map[string]any{}}
	if s.DB == nil {
		return result, []string{"remove_menu requested but database is not configured"}, nil
	}

	routePath := strings.TrimSpace(lock.RoutePath)
	if routePath == "" {
		routePath = "/system/" + ToKebab(moduleName)
	}
	permissionCodes := append([]string{}, lock.PermissionCodes...)
	if len(permissionCodes) == 0 {
		permissionCodes = []string{
			moduleName + ".list",
			moduleName + ".save",
			moduleName + ".delete",
		}
	}

	query := s.DB.Model(&AdminMenu{})
	if routePath != "" {
		query = query.Where("path = ?", routePath)
	}
	if len(permissionCodes) > 0 {
		query = query.Or("permission_code IN ?", permissionCodes)
	}

	var rows []AdminMenu
	if err := query.Order("id ASC").Find(&rows).Error; err != nil {
		return result, nil, err
	}
	if len(rows) == 0 {
		return result, []string{fmt.Sprintf("module %s has no menu records to remove", moduleName)}, nil
	}

	menuIDs := make([]int64, 0, len(rows))
	for _, row := range rows {
		menuIDs = append(menuIDs, row.ID)
		result.records = append(result.records, map[string]any{
			"id":              row.ID,
			"name":            row.Name,
			"title":           row.Title,
			"menu_type":       row.MenuType,
			"path":            row.Path,
			"permission_code": row.PermissionCode,
		})
	}

	tx := s.DB.Begin()
	if tx.Error != nil {
		return result, nil, tx.Error
	}
	roleMenuDelete := tx.Where("menu_id IN ?", menuIDs).Delete(&AdminRoleMenu{})
	if roleMenuDelete.Error != nil {
		tx.Rollback()
		return result, nil, roleMenuDelete.Error
	}
	result.roleMenuLinks = roleMenuDelete.RowsAffected

	if err := tx.Where("id IN ?", menuIDs).Delete(&AdminMenu{}).Error; err != nil {
		tx.Rollback()
		return result, nil, err
	}
	if err := tx.Commit().Error; err != nil {
		return result, nil, err
	}
	return result, nil, nil
}

func (s GeneratorService) removeHistory(moduleName string) ([]int64, []string, error) {
	if s.DB == nil {
		return nil, []string{"remove_history requested but database is not configured"}, nil
	}
	var rows []CodegenHistory
	if err := s.DB.Where("module_name = ?", moduleName).Order("id ASC").Find(&rows).Error; err != nil {
		return nil, nil, err
	}
	if len(rows) == 0 {
		return nil, []string{fmt.Sprintf("module %s has no codegen history to remove", moduleName)}, nil
	}

	ids := make([]int64, 0, len(rows))
	for _, row := range rows {
		ids = append(ids, row.ID)
	}
	if err := s.DB.Where("id IN ?", ids).Delete(&CodegenHistory{}).Error; err != nil {
		return nil, nil, err
	}
	return ids, nil, nil
}
