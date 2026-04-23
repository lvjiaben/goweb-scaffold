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
			actualPath:   filepath.Join(repoRoot, "vben-admin/apps/backend/src/views/demo_article/list.vue"),
			snapshotPath: filepath.Join("testdata", "snapshots", "demo_article.list.vue.golden"),
		},
		{
			actualPath:   filepath.Join(repoRoot, "vben-admin/apps/backend/src/views/demo_article/data.ts"),
			snapshotPath: filepath.Join("testdata", "snapshots", "demo_article.data.ts.golden"),
		},
		{
			actualPath:   filepath.Join(repoRoot, "vben-admin/apps/backend/src/views/demo_article/modules/form-drawer.vue"),
			snapshotPath: filepath.Join("testdata", "snapshots", "demo_article.form-drawer.vue.golden"),
		},
		{
			actualPath:   filepath.Join(repoRoot, "vben-admin/apps/backend/src/views/demo_notice/list.vue"),
			snapshotPath: filepath.Join("testdata", "snapshots", "demo_notice.list.vue.golden"),
		},
		{
			actualPath:   filepath.Join(repoRoot, "vben-admin/apps/backend/src/views/demo_notice/data.ts"),
			snapshotPath: filepath.Join("testdata", "snapshots", "demo_notice.data.ts.golden"),
		},
		{
			actualPath:   filepath.Join(repoRoot, "vben-admin/apps/backend/src/views/demo_notice/modules/form-drawer.vue"),
			snapshotPath: filepath.Join("testdata", "snapshots", "demo_notice.form-drawer.vue.golden"),
		},
		{
			actualPath:   filepath.Join(repoRoot, "vben-admin/apps/backend/src/api/demo_article.ts"),
			snapshotPath: filepath.Join("testdata", "snapshots", "demo_article.ts.golden"),
		},
		{
			actualPath:   filepath.Join(repoRoot, "vben-admin/apps/backend/src/api/demo_notice.ts"),
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
