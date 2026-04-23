package model

import "strings"

var legacyMenuIconMap = map[string]string{
	"code":      "lucide:code",
	"dashboard": "lucide:layout-dashboard",
	"file":      "lucide:file-text",
	"home":      "lucide:house",
	"menu":      "lucide:menu",
	"paperclip": "lucide:paperclip",
	"setting":   "lucide:settings",
	"settings":  "lucide:settings",
	"shield":    "lucide:shield",
	"user":      "lucide:user",
	"users":     "lucide:users",
}

func NormalizeMenuIcon(icon string) string {
	value := strings.TrimSpace(icon)
	if value == "" || strings.Contains(value, ":") || strings.HasPrefix(value, "http://") || strings.HasPrefix(value, "https://") {
		return value
	}
	if mapped, ok := legacyMenuIconMap[strings.ToLower(value)]; ok {
		return mapped
	}
	return value
}
