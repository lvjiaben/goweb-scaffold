package registry

import (
	"go/format"
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

func RebuildBackendModulesFile(repoRoot string) (string, string, error) {
	items, err := DiscoverGeneratedModules(repoRoot)
	if err != nil {
		return "", "", err
	}

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
		return "", "", err
	}
	content, err = format.Source(content)
	if err != nil {
		return "", "", err
	}

	return writer.New(repoRoot).Write("internal/gen/modules_gen.go", content, true)
}

func RebuildAdminRoutesFile(repoRoot string) (string, string, error) {
	items, err := DiscoverGeneratedModules(repoRoot)
	if err != nil {
		return "", "", err
	}

	routes := make([]routeItem, 0, len(items))
	for _, item := range items {
		routes = append(routes, routeItem{
			ModuleName:     item.ModuleName,
			RoutePath:      item.RoutePath,
			Title:          item.Title,
			ViewImportPath: "@/views/system/" + item.PageName + ".vue",
		})
	}

	content, err := gentemplates.Render("admin_frontend/routes.ts.tmpl", map[string]any{
		"Routes": routes,
	}, template.FuncMap{
		"quote": strconv.Quote,
	})
	if err != nil {
		return "", "", err
	}

	return writer.New(repoRoot).Write("vben-admin/apps/admin/src/generated/routes.ts", content, true)
}
