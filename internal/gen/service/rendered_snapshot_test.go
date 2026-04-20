package service

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRenderedSnapshots(t *testing.T) {
	repoRoot := filepath.Clean(filepath.Join("..", "..", ".."))
	cases := []struct {
		actualPath   string
		snapshotPath string
	}{
		{
			actualPath:   filepath.Join(repoRoot, "vben-admin/apps/admin/src/views/system/DemoArticlePage.vue"),
			snapshotPath: filepath.Join("testdata", "snapshots", "DemoArticlePage.vue.golden"),
		},
		{
			actualPath:   filepath.Join(repoRoot, "vben-admin/apps/admin/src/views/system/DemoNoticePage.vue"),
			snapshotPath: filepath.Join("testdata", "snapshots", "DemoNoticePage.vue.golden"),
		},
		{
			actualPath:   filepath.Join(repoRoot, "vben-admin/apps/admin/src/api/demo_article.ts"),
			snapshotPath: filepath.Join("testdata", "snapshots", "demo_article.ts.golden"),
		},
		{
			actualPath:   filepath.Join(repoRoot, "vben-admin/apps/admin/src/api/demo_notice.ts"),
			snapshotPath: filepath.Join("testdata", "snapshots", "demo_notice.ts.golden"),
		},
		{
			actualPath:   filepath.Join(repoRoot, "internal/modules/demo_article/meta.go"),
			snapshotPath: filepath.Join("testdata", "snapshots", "demo_article.meta.go.golden"),
		},
		{
			actualPath:   filepath.Join(repoRoot, "internal/modules/demo_notice/meta.go"),
			snapshotPath: filepath.Join("testdata", "snapshots", "demo_notice.meta.go.golden"),
		},
	}

	for _, item := range cases {
		t.Run(filepath.Base(item.actualPath), func(t *testing.T) {
			actual, err := os.ReadFile(item.actualPath)
			if err != nil {
				t.Fatalf("read actual snapshot file: %v", err)
			}
			writeSnapshotIfNeeded(t, item.snapshotPath, actual)

			expected, err := os.ReadFile(item.snapshotPath)
			if err != nil {
				t.Fatalf("read snapshot golden: %v", err)
			}
			if string(actual) != string(expected) {
				t.Fatalf("snapshot mismatch for %s", item.actualPath)
			}
		})
	}
}
