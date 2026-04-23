package modules

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

var handlerDBForbiddenTokens = []string{
	"runtime.DB",
	".Model(",
	".Where(",
	".Count(",
	".Find(",
	".First(",
	".Create(",
	".Save(",
	".Delete(",
	".Transaction(",
	".Begin(",
}

func TestHandlersDoNotOwnGORM(t *testing.T) {
	patterns := []string{
		"app/*/handler.go",
		"admin/*/handler.go",
		"system/*/handler.go",
	}

	for _, pattern := range patterns {
		files, err := filepath.Glob(pattern)
		if err != nil {
			t.Fatalf("glob %s: %v", pattern, err)
		}
		for _, file := range files {
			file = filepath.ToSlash(file)
			content, err := os.ReadFile(file)
			if err != nil {
				t.Fatalf("read %s: %v", file, err)
			}
			source := string(content)
			for _, token := range handlerDBForbiddenTokens {
				if strings.Contains(source, token) {
					t.Fatalf("%s contains forbidden DB token %q; move DB logic to repo/service", file, token)
				}
			}
		}
	}
}
