CREATE TABLE IF NOT EXISTS demo_article (
  id BIGSERIAL PRIMARY KEY,
  title VARCHAR(255) NOT NULL DEFAULT '',
  summary TEXT NOT NULL DEFAULT '',
  status SMALLINT NOT NULL DEFAULT 1,
  sort INTEGER NOT NULL DEFAULT 0,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  deleted_at TIMESTAMPTZ NULL
);

CREATE INDEX IF NOT EXISTS idx_demo_article_status ON demo_article (status);
CREATE INDEX IF NOT EXISTS idx_demo_article_sort ON demo_article (sort);

INSERT INTO demo_article (title, summary, status, sort, created_at, updated_at)
SELECT 'Hello Codegen', '第一条演示文章，用于验证 codegen 生成的 admin CRUD。', 1, 10, NOW(), NOW()
WHERE NOT EXISTS (
  SELECT 1 FROM demo_article WHERE title = 'Hello Codegen' AND deleted_at IS NULL
);

INSERT INTO demo_article (title, summary, status, sort, created_at, updated_at)
SELECT 'Scaffold Ready', '第二条演示文章，验证列表、搜索、编辑和删除。', 1, 20, NOW(), NOW()
WHERE NOT EXISTS (
  SELECT 1 FROM demo_article WHERE title = 'Scaffold Ready' AND deleted_at IS NULL
);
