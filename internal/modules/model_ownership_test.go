package modules

import (
	"os"
	"strings"
	"testing"
)

func TestModuleModelOwnership(t *testing.T) {
	cases := []struct {
		path      string
		forbidden []string
	}{
		{
			path: "admin/user/model.go",
			forbidden: []string{
				"type AppUser",
				"type AppUserSession",
				"type SystemConfig",
				"type FileAttachment",
				"type CodegenHistory",
				"type AdminMenu",
				"type AdminRoleMenu",
			},
		},
		{
			path: "app/user/model.go",
			forbidden: []string{
				"type AdminUser",
				"type AdminRole",
				"type AdminMenu",
				"type SystemConfig",
				"type FileAttachment",
				"type CodegenHistory",
			},
		},
		{
			path: "system/config/model.go",
			forbidden: []string{
				"type AdminUser",
				"type AdminRole",
				"type AdminMenu",
				"type AppUser",
				"type FileAttachment",
				"type CodegenHistory",
			},
		},
		{
			path: "system/attachment/model.go",
			forbidden: []string{
				"type AdminUser",
				"type AdminRole",
				"type AdminMenu",
				"type AppUser",
				"type SystemConfig",
				"type CodegenHistory",
			},
		},
		{
			path: "app/demo_article/model.go",
			forbidden: []string{
				"type AdminUser",
				"type AdminRole",
				"type AdminMenu",
				"type AppUser",
				"type SystemConfig",
				"type FileAttachment",
				"type CodegenHistory",
				"type BaseModel",
			},
		},
		{
			path: "app/demo_notice/model.go",
			forbidden: []string{
				"type AdminUser",
				"type AdminRole",
				"type AdminMenu",
				"type AppUser",
				"type SystemConfig",
				"type FileAttachment",
				"type CodegenHistory",
				"type BaseModel",
			},
		},
	}

	for _, tc := range cases {
		raw, err := os.ReadFile(tc.path)
		if err != nil {
			t.Fatalf("read %s: %v", tc.path, err)
		}
		source := string(raw)
		for _, token := range tc.forbidden {
			if strings.Contains(source, token) {
				t.Fatalf("%s contains forbidden model token %q", tc.path, token)
			}
		}
	}
}
