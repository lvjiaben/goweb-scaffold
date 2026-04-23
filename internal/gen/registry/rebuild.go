package registry

import (
	"go/format"
	"sort"
	"strconv"
	"strings"
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
	ModuleName string
	RoutePath  string
	PageName   string
	Title      string
	Component  string
}

func RenderBackendModulesFile(repoRoot string, upsertModules ...GeneratedModule) ([]byte, error) {
	return RenderBackendModulesFileWithOptions(repoRoot, upsertModules, nil)
}

func RenderBackendModulesFileWithOptions(repoRoot string, upsertModules []GeneratedModule, excludeModules []string) ([]byte, error) {
	items, err := DiscoverGeneratedModules(repoRoot)
	if err != nil {
		return nil, err
	}
	items = mergeGeneratedModules(items, upsertModules...)
	items = excludeGeneratedModules(items, excludeModules)

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

func RebuildBackendModulesFile(repoRoot string, upsertModules ...GeneratedModule) (string, string, error) {
	content, err := RenderBackendModulesFileWithOptions(repoRoot, upsertModules, nil)
	if err != nil {
		return "", "", err
	}
	return writer.New(repoRoot).Write("internal/gen/modules_gen.go", content, true)
}

func RenderFrontendRouteModule(repoRoot string, upsertModules ...GeneratedModule) ([]byte, error) {
	return RenderFrontendRouteModuleWithOptions(repoRoot, upsertModules, nil)
}

func RenderFrontendRouteModuleWithOptions(repoRoot string, upsertModules []GeneratedModule, excludeModules []string) ([]byte, error) {
	items, err := DiscoverGeneratedModules(repoRoot)
	if err != nil {
		return nil, err
	}
	items = mergeGeneratedModules(items, upsertModules...)
	items = excludeGeneratedModules(items, excludeModules)

	routes := make([]routeItem, 0, len(items))
	for _, item := range items {
		component := strings.TrimPrefix(strings.TrimSuffix(item.ViewFile, ".vue"), "views/")
		component = strings.TrimPrefix(component, "/")
		if component == "" {
			component = strings.TrimSpace(item.ModuleName) + "/list"
		}
		routes = append(routes, routeItem{
			ModuleName: item.ModuleName,
			RoutePath:  item.RoutePath,
			PageName:   item.PageName,
			Title:      item.Title,
			Component:  component,
		})
	}

	return gentemplates.Render("admin_frontend/routes.ts.tmpl", map[string]any{
		"Routes": routes,
	}, template.FuncMap{
		"quote": strconv.Quote,
	})
}

func RebuildFrontendRouteModule(repoRoot string, upsertModules ...GeneratedModule) (string, string, error) {
	content, err := RenderFrontendRouteModuleWithOptions(repoRoot, upsertModules, nil)
	if err != nil {
		return "", "", err
	}
	return writer.New(repoRoot).Write("vben-admin/apps/backend/src/router/routes/modules/generated.ts", content, true)
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

func excludeGeneratedModules(items []GeneratedModule, excludeModules []string) []GeneratedModule {
	if len(excludeModules) == 0 {
		return items
	}

	excluded := make(map[string]struct{}, len(excludeModules))
	for _, item := range excludeModules {
		moduleName := strings.TrimSpace(item)
		if moduleName == "" {
			continue
		}
		excluded[moduleName] = struct{}{}
	}
	if len(excluded) == 0 {
		return items
	}

	result := make([]GeneratedModule, 0, len(items))
	for _, item := range items {
		if _, ok := excluded[item.ModuleName]; ok {
			continue
		}
		result = append(result, item)
	}
	return result
}
