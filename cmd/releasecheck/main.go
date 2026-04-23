package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type checkItem struct {
	Name   string `json:"name"`
	OK     bool   `json:"ok"`
	Detail string `json:"detail,omitempty"`
}

type repoResult struct {
	Repo    string      `json:"repo"`
	Root    string      `json:"root"`
	Version string      `json:"version,omitempty"`
	Passed  bool        `json:"passed"`
	Checks  []checkItem `json:"checks"`
}

type resultPayload struct {
	Scope  string       `json:"scope"`
	Passed bool         `json:"passed"`
	Repos  []repoResult `json:"repos"`
}

type repoSpec struct {
	Name             string
	Root             string
	RequireUpgrade   bool
	RequireTemplate  bool
	RequireCoreRef   bool
	ExpectedTemplate string
}

func main() {
	var format string
	var scope string
	var scaffoldRoot string
	var coreRoot string

	flag.StringVar(&format, "format", "text", "output format: text|json")
	flag.StringVar(&scope, "scope", "all", "check scope: scaffold|core|all")
	flag.StringVar(&scaffoldRoot, "scaffold-root", ".", "goweb-scaffold repo root")
	flag.StringVar(&coreRoot, "core-root", "../goweb-core", "goweb-core repo root")
	flag.Parse()

	payload, err := run(strings.TrimSpace(scope), scaffoldRoot, coreRoot)
	if err != nil {
		writeError(strings.TrimSpace(format), err)
		os.Exit(1)
	}
	if strings.TrimSpace(format) == "json" {
		raw, _ := json.MarshalIndent(payload, "", "  ")
		_, _ = os.Stdout.Write(append(raw, '\n'))
		if !payload.Passed {
			os.Exit(1)
		}
		return
	}
	for _, repo := range payload.Repos {
		fmt.Printf("[%s] version=%s passed=%v root=%s\n", repo.Repo, repo.Version, repo.Passed, repo.Root)
		for _, item := range repo.Checks {
			status := "OK"
			if !item.OK {
				status = "FAIL"
			}
			if item.Detail != "" {
				fmt.Printf("  - [%s] %s: %s\n", status, item.Name, item.Detail)
			} else {
				fmt.Printf("  - [%s] %s\n", status, item.Name)
			}
		}
	}
	fmt.Printf("summary scope=%s passed=%v\n", payload.Scope, payload.Passed)
	if !payload.Passed {
		os.Exit(1)
	}
}

func run(scope string, scaffoldRoot string, coreRoot string) (resultPayload, error) {
	repos := []repoSpec{}
	switch scope {
	case "scaffold":
		root, err := filepath.Abs(scaffoldRoot)
		if err != nil {
			return resultPayload{}, err
		}
		repos = append(repos, repoSpec{
			Name:             "goweb-scaffold",
			Root:             root,
			RequireUpgrade:   true,
			RequireTemplate:  true,
			RequireCoreRef:   true,
			ExpectedTemplate: "v7",
		})
	case "core":
		root, err := filepath.Abs(coreRoot)
		if err != nil {
			return resultPayload{}, err
		}
		repos = append(repos, repoSpec{Name: "goweb-core", Root: root})
	case "all":
		scaffoldAbs, err := filepath.Abs(scaffoldRoot)
		if err != nil {
			return resultPayload{}, err
		}
		coreAbs, err := filepath.Abs(coreRoot)
		if err != nil {
			return resultPayload{}, err
		}
		repos = append(repos,
			repoSpec{
				Name:             "goweb-scaffold",
				Root:             scaffoldAbs,
				RequireUpgrade:   true,
				RequireTemplate:  true,
				RequireCoreRef:   true,
				ExpectedTemplate: "v7",
			},
			repoSpec{Name: "goweb-core", Root: coreAbs},
		)
	default:
		return resultPayload{}, fmt.Errorf("unsupported scope: %s", scope)
	}

	payload := resultPayload{Scope: scope, Passed: true, Repos: make([]repoResult, 0, len(repos))}
	for _, spec := range repos {
		repo := checkRepo(spec)
		if !repo.Passed {
			payload.Passed = false
		}
		payload.Repos = append(payload.Repos, repo)
	}
	return payload, nil
}

func checkRepo(spec repoSpec) repoResult {
	result := repoResult{
		Repo:   spec.Name,
		Root:   spec.Root,
		Passed: true,
		Checks: []checkItem{},
	}

	version := readTrimmed(filepath.Join(spec.Root, "VERSION"))
	result.Version = version

	requiredFiles := []string{
		"VERSION",
		"CHANGELOG.md",
		"README.md",
		"docs/versioning.md",
		"docs/compatibility.md",
		"docs/releases/RELEASE_POLICY.md",
		"docs/releases/RELEASE_CHECKLIST.md",
	}
	for _, rel := range requiredFiles {
		result.add(fileExists(filepath.Join(spec.Root, rel)), "file:"+rel, "")
	}
	if version != "" {
		result.add(fileExists(filepath.Join(spec.Root, "docs/releases", version+".md")), "release-notes", version)
	}
	if spec.RequireUpgrade && version != "" {
		result.add(fileExists(filepath.Join(spec.Root, "docs/upgrade", "UPGRADE_"+version+".md")), "upgrade-guide", version)
	}

	readme := readTrimmed(filepath.Join(spec.Root, "README.md"))
	changelog := readTrimmed(filepath.Join(spec.Root, "CHANGELOG.md"))
	versioning := readTrimmed(filepath.Join(spec.Root, "docs/versioning.md"))
	compatibility := readTrimmed(filepath.Join(spec.Root, "docs/compatibility.md"))

	if version != "" {
		result.add(strings.Contains(readme, version), "readme-version", version)
		result.add(strings.Contains(changelog, version), "changelog-version", version)
		result.add(strings.Contains(versioning, version), "versioning-version", version)
		result.add(strings.Contains(compatibility, version), "compatibility-version", version)
	}

	result.add(strings.Contains(readme, "docs/versioning.md"), "readme-doc-link:versioning", "")
	result.add(strings.Contains(readme, "docs/compatibility.md"), "readme-doc-link:compatibility", "")
	result.add(strings.Contains(readme, "docs/releases/RELEASE_POLICY.md"), "readme-doc-link:release-policy", "")
	result.add(strings.Contains(readme, "docs/releases/RELEASE_CHECKLIST.md"), "readme-doc-link:release-checklist", "")
	if version != "" {
		result.add(strings.Contains(readme, "docs/releases/"+version+".md"), "readme-doc-link:release-notes", "")
	}

	if spec.RequireTemplate {
		result.add(strings.Contains(readme, "template version"), "readme-template-version", "")
		result.add(strings.Contains(versioning, spec.ExpectedTemplate), "versioning-template-version", spec.ExpectedTemplate)
		result.add(strings.Contains(compatibility, spec.ExpectedTemplate), "compatibility-template-version", spec.ExpectedTemplate)
		result.add(strings.Contains(readme, "docs/upgrade/UPGRADE_"+version+".md"), "readme-doc-link:upgrade", "")
	}
	if spec.RequireCoreRef {
		result.add(strings.Contains(readme, "goweb-core"), "readme-core-compatibility", "")
	}
	if spec.Name == "goweb-scaffold" {
		result.add(strings.Contains(readme, "apps/backend"), "readme-frontend-structure:backend", "")
		result.add(strings.Contains(readme, "apps/user"), "readme-frontend-structure:user", "")
		result.add(strings.Contains(readme, "apps/backend/src/views"), "readme-codegen-path:views", "")
		result.add(strings.Contains(readme, "modules/form-drawer.vue"), "readme-codegen-path:form-drawer", "")
		result.add(!strings.Contains(readme, "`apps/admin` 和 `apps/user` 双应用"), "readme-no-legacy-admin-monorepo", "")
		result.add(!strings.Contains(readme, "cd vben-admin/apps/admin"), "readme-no-legacy-admin-command", "")
	}

	return result
}

func (r *repoResult) add(ok bool, name string, detail string) {
	item := checkItem{Name: name, OK: ok, Detail: detail}
	r.Checks = append(r.Checks, item)
	if !ok {
		r.Passed = false
	}
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func readTrimmed(path string) string {
	raw, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(raw))
}

func writeError(format string, err error) {
	if format == "json" {
		raw, _ := json.MarshalIndent(map[string]any{"error": err.Error()}, "", "  ")
		_, _ = os.Stdout.Write(append(raw, '\n'))
		return
	}
	fmt.Fprintf(os.Stderr, "error: %v\n", err)
}
