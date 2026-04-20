package templates

type VersionInfo struct {
	Name        string `json:"name"`
	Status      string `json:"status"`
	Description string `json:"description"`
}

const (
	LegacyVersion  = "v5"
	CurrentVersion = "v6"
	DefaultVersion = CurrentVersion
)

var SupportedVersions = []VersionInfo{
	{
		Name:        LegacyVersion,
		Status:      "legacy",
		Description: "历史模板版本，page meta 不包含统一的 menu_title / feature_flags。",
	},
	{
		Name:        CurrentVersion,
		Status:      "current",
		Description: "当前模板版本，统一补齐 page.menu_title 和 page.feature_flags，支持 source migration 与 batch。",
	},
}

func IsSupported(version string) bool {
	for _, item := range SupportedVersions {
		if item.Name == version {
			return true
		}
	}
	return false
}
