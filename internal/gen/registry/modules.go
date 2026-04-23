package registry

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

type GeneratedModule struct {
	ModuleName string
	RoutePath  string
	PageName   string
	ViewFile   string
	Title      string
}

type BaseModule struct {
	ModuleName string
	ImportPath string
}

var baseModules = []BaseModule{
	{ModuleName: "admin_auth", ImportPath: "github.com/lvjiaben/goweb-scaffold/internal/modules/admin_auth"},
	{ModuleName: "admin_user", ImportPath: "github.com/lvjiaben/goweb-scaffold/internal/modules/admin_user"},
	{ModuleName: "admin_role", ImportPath: "github.com/lvjiaben/goweb-scaffold/internal/modules/admin_role"},
	{ModuleName: "admin_menu", ImportPath: "github.com/lvjiaben/goweb-scaffold/internal/modules/admin_menu"},
	{ModuleName: "system_config", ImportPath: "github.com/lvjiaben/goweb-scaffold/internal/modules/system_config"},
	{ModuleName: "attachment", ImportPath: "github.com/lvjiaben/goweb-scaffold/internal/modules/attachment"},
	{ModuleName: "app_user_auth", ImportPath: "github.com/lvjiaben/goweb-scaffold/internal/modules/app_user_auth"},
	{ModuleName: "app_user", ImportPath: "github.com/lvjiaben/goweb-scaffold/internal/modules/app_user"},
	{ModuleName: "codegen", ImportPath: "github.com/lvjiaben/goweb-scaffold/internal/modules/codegen"},
}

var (
	moduleNameRegexp = regexp.MustCompile(`GeneratedModuleName\s*=\s*"([^"]+)"`)
	routePathRegexp  = regexp.MustCompile(`GeneratedRoutePath\s*=\s*"([^"]+)"`)
	pageNameRegexp   = regexp.MustCompile(`GeneratedPageName\s*=\s*"([^"]+)"`)
	viewFileRegexp   = regexp.MustCompile(`GeneratedViewFile\s*=\s*"([^"]+)"`)
	titleRegexp      = regexp.MustCompile(`GeneratedMenuTitle\s*=\s*"([^"]+)"`)
)

func BaseModules() []BaseModule {
	return append([]BaseModule{}, baseModules...)
}

func DiscoverGeneratedModules(repoRoot string) ([]GeneratedModule, error) {
	modulesDir := filepath.Join(repoRoot, "internal/modules")
	entries, err := os.ReadDir(modulesDir)
	if err != nil {
		return nil, err
	}

	items := make([]GeneratedModule, 0)
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		metaPath := filepath.Join(modulesDir, entry.Name(), "meta.go")
		content, err := os.ReadFile(metaPath)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, err
		}

		moduleName := firstMatch(moduleNameRegexp, content)
		routePath := firstMatch(routePathRegexp, content)
		pageName := firstMatch(pageNameRegexp, content)
		viewFile := firstMatch(viewFileRegexp, content)
		title := firstMatch(titleRegexp, content)
		if moduleName == "" || routePath == "" {
			continue
		}

		if isBaseModule(moduleName) {
			continue
		}

		if pageName == "" {
			pageName = generatedPageName(moduleName)
		}
		if viewFile == "" {
			viewFile = generatedViewFile(moduleName)
		}
		if title == "" {
			title = generatedTitle(moduleName)
		}

		items = append(items, GeneratedModule{
			ModuleName: moduleName,
			RoutePath:  routePath,
			PageName:   pageName,
			ViewFile:   viewFile,
			Title:      title,
		})
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].ModuleName < items[j].ModuleName
	})
	return items, nil
}

func firstMatch(pattern *regexp.Regexp, content []byte) string {
	matched := pattern.FindSubmatch(content)
	if len(matched) < 2 {
		return ""
	}
	return string(matched[1])
}

func isBaseModule(moduleName string) bool {
	for _, item := range baseModules {
		if item.ModuleName == moduleName {
			return true
		}
	}
	return false
}

func ImportPathForModule(moduleName string) string {
	return fmt.Sprintf("github.com/lvjiaben/goweb-scaffold/internal/modules/%s", moduleName)
}

func ModuleAlias(moduleName string) string {
	return strings.TrimSpace(moduleName)
}

func generatedPageName(moduleName string) string {
	return toPascal(moduleName) + "List"
}

func generatedViewFile(moduleName string) string {
	return filepath.ToSlash(filepath.Join("views", moduleName, "list.vue"))
}

func generatedTitle(moduleName string) string {
	return humanizeModuleName(moduleName)
}

func toPascal(raw string) string {
	parts := strings.FieldsFunc(strings.TrimSpace(raw), func(r rune) bool {
		return r == '_' || r == '-' || r == ' '
	})
	builder := strings.Builder{}
	for _, part := range parts {
		if part == "" {
			continue
		}
		builder.WriteString(strings.ToUpper(part[:1]))
		if len(part) > 1 {
			builder.WriteString(part[1:])
		}
	}
	return builder.String()
}

func humanizeModuleName(raw string) string {
	parts := strings.FieldsFunc(strings.TrimSpace(raw), func(r rune) bool {
		return r == '_' || r == '-'
	})
	for index, part := range parts {
		if part == "" {
			continue
		}
		parts[index] = strings.ToUpper(part[:1]) + part[1:]
	}
	return strings.Join(parts, " ")
}
