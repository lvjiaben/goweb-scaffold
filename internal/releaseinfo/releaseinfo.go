package releaseinfo

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	gentemplates "github.com/lvjiaben/goweb-scaffold/internal/gen/templates"
)

const (
	RepoName             = "goweb-scaffold"
	CompatibleCore       = "v0.9.x"
	CompatibleCoreStrict = "v0.9.0-rc.1"
)

type VersionInfo struct {
	Repo                      string   `json:"repo"`
	Version                   string   `json:"version"`
	TemplateVersion           string   `json:"template_version"`
	CompatibleCore            string   `json:"compatible_core"`
	SupportedTemplateVersions []string `json:"supported_template_versions"`
}

func DetectRepoRoot(start string) (string, error) {
	current := strings.TrimSpace(start)
	if current == "" {
		wd, err := os.Getwd()
		if err != nil {
			return "", err
		}
		current = wd
	}
	current, err := filepath.Abs(current)
	if err != nil {
		return "", err
	}

	for {
		if fileExists(filepath.Join(current, "go.mod")) && fileExists(filepath.Join(current, "VERSION")) {
			return current, nil
		}
		parent := filepath.Dir(current)
		if parent == current {
			return "", fmt.Errorf("repo root not found from %s", start)
		}
		current = parent
	}
}

func ReadVersion(repoRoot string) (string, error) {
	root, err := DetectRepoRoot(repoRoot)
	if err != nil {
		return "", err
	}
	raw, err := os.ReadFile(filepath.Join(root, "VERSION"))
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(raw)), nil
}

func LoadVersionInfo(repoRoot string) (VersionInfo, error) {
	version, err := ReadVersion(repoRoot)
	if err != nil {
		return VersionInfo{}, err
	}
	supported := make([]string, 0, len(gentemplates.SupportedVersions))
	for _, item := range gentemplates.SupportedVersions {
		supported = append(supported, item.Name)
	}
	return VersionInfo{
		Repo:                      RepoName,
		Version:                   version,
		TemplateVersion:           gentemplates.CurrentVersion,
		CompatibleCore:            CompatibleCore,
		SupportedTemplateVersions: supported,
	}, nil
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}
