package registry

import (
	"go/format"
	"sort"
	"strconv"
	"text/template"

	gentemplates "github.com/lvjiaben/goweb-scaffold/internal/gen/templates"
	"github.com/lvjiaben/goweb-scaffold/internal/gen/writer"
)

type importItem struct {
	Alias      string
	ImportPath string
}

type moduleItem struct {
	Alias string
}

type routeItem struct {
	ModuleName     string
	RoutePath      string
	Title          string
	ViewImportPath string
}

func RenderBackendModulesFile(repoRoot string, upsertModules ...GeneratedModule) ([]byte, error) {
	items, err := DiscoverGeneratedModules(repoRoot)
	if err != nil {
		return nil, err
	}
	items = mergeGeneratedModules(items, upsertModules...)

	imports := make([]importItem, 0, len(baseModules)+len(items))
	modules := make([]moduleItem, 0, len(baseModules)+len(items))

	for _, item := range BaseModules() {
		imports = append(imports, importItem{
			Alias:      ModuleAlias(item.ModuleName),
			ImportPath: item.ImportPath,
		})
		modules = append(modules, moduleItem{Alias: ModuleAlias(item.ModuleName)})
	}

	for _, item := range items {
		imports = append(imports, importItem{
			Alias:      ModuleAlias(item.ModuleName),
			ImportPath: ImportPathForModule(item.ModuleName),
		})
		modules = append(modules, moduleItem{Alias: ModuleAlias(item.ModuleName)})
	}

	content, err := gentemplates.Render("backend/modules_gen.go.tmpl", map[string]any{
		"Imports": imports,
		"Modules": modules,
	}, template.FuncMap{
		"quote": strconv.Quote,
	})
	if err != nil {
		return nil, err
	}
	return format.Source(content)
}

func RenderAdminRoutesFile(repoRoot string, upsertModules ...GeneratedModule) ([]byte, error) {
	items, err := DiscoverGeneratedModules(repoRoot)
	if err != nil {
		return nil, err
	}
	items = mergeGeneratedModules(items, upsertModules...)

	routes := make([]routeItem, 0, len(items))
	for _, item := range items {
		routes = append(routes, routeItem{
			ModuleName:     item.ModuleName,
			RoutePath:      item.RoutePath,
			Title:          item.Title,
			ViewImportPath: "@/views/system/" + item.PageName + ".vue",
		})
	}

	return gentemplates.Render("admin_frontend/routes.ts.tmpl", map[string]any{
		"Routes": routes,
	}, template.FuncMap{
		"quote": strconv.Quote,
	})
}

func RebuildBackendModulesFile(repoRoot string, upsertModules ...GeneratedModule) (string, string, error) {
	content, err := RenderBackendModulesFile(repoRoot, upsertModules...)
	if err != nil {
		return "", "", err
	}
	return writer.New(repoRoot).Write("internal/gen/modules_gen.go", content, true)
}

func RebuildAdminRoutesFile(repoRoot string, upsertModules ...GeneratedModule) (string, string, error) {
	content, err := RenderAdminRoutesFile(repoRoot, upsertModules...)
	if err != nil {
		return "", "", err
	}
	return writer.New(repoRoot).Write("vben-admin/apps/admin/src/generated/routes.ts", content, true)
}

func mergeGeneratedModules(items []GeneratedModule, upsertModules ...GeneratedModule) []GeneratedModule {
	index := make(map[string]GeneratedModule, len(items)+len(upsertModules))
	for _, item := range items {
		index[item.ModuleName] = item
	}
	for _, item := range upsertModules {
		if item.ModuleName == "" {
			continue
		}
		index[item.ModuleName] = item
	}

	result := make([]GeneratedModule, 0, len(index))
	for _, item := range index {
		result = append(result, item)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].ModuleName < result[j].ModuleName
	})
	return result
}
