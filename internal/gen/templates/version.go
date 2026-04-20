package templates

type VersionInfo struct {
	Name        string `json:"name"`
	Status      string `json:"status"`
	Description string `json:"description"`
}

const (
	LegacyVersion  = "v5"
	V6Version      = "v6"
	CurrentVersion = "v7"
	DefaultVersion = CurrentVersion
)

var SupportedVersions = []VersionInfo{
	{
		Name:        LegacyVersion,
		Status:      "legacy",
		Description: "历史模板版本，page meta 不包含统一的 menu_title / feature_flags。",
	},
	{
		Name:        V6Version,
		Status:      "legacy",
		Description: "中间模板版本，已补齐 page.menu_title / page.feature_flags，但尚未写入稳定性 snapshot。",
	},
	{
		Name:        CurrentVersion,
		Status:      "current",
		Description: "当前模板版本，写入稳定性 snapshot，并支持 check-breaking / batch check-breaking。",
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
