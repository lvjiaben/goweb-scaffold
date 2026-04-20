CREATE TABLE IF NOT EXISTS demo_notice (
  id BIGSERIAL PRIMARY KEY,
  title VARCHAR(255) NOT NULL DEFAULT '',
  content TEXT NOT NULL DEFAULT '',
  status SMALLINT NOT NULL DEFAULT 1,
  sort INTEGER NOT NULL DEFAULT 0,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  deleted_at TIMESTAMPTZ NULL
);

CREATE INDEX IF NOT EXISTS idx_demo_notice_status ON demo_notice (status);
CREATE INDEX IF NOT EXISTS idx_demo_notice_sort ON demo_notice (sort);

COMMENT ON TABLE demo_notice IS '演示公告';
COMMENT ON COLUMN demo_notice.title IS '公告标题';
COMMENT ON COLUMN demo_notice.content IS '公告内容';
COMMENT ON COLUMN demo_notice.status IS '状态';
COMMENT ON COLUMN demo_notice.sort IS '排序值';
COMMENT ON COLUMN demo_notice.created_at IS '创建时间';
COMMENT ON COLUMN demo_notice.updated_at IS '更新时间';

INSERT INTO demo_notice (title, content, status, sort, created_at, updated_at)
SELECT '系统升级通知', '今晚 23:00 将进行例行维护，请提前保存工作内容。', 1, 10, NOW(), NOW()
WHERE NOT EXISTS (
  SELECT 1 FROM demo_notice WHERE title = '系统升级通知' AND deleted_at IS NULL
);

INSERT INTO demo_notice (title, content, status, sort, created_at, updated_at)
SELECT '新模块上线', 'demo_notice 用于验证第八阶段 batch codegen 的真实生成链路。', 1, 20, NOW(), NOW()
WHERE NOT EXISTS (
  SELECT 1 FROM demo_notice WHERE title = '新模块上线' AND deleted_at IS NULL
);
