ALTER TABLE app_user ADD COLUMN IF NOT EXISTS pid BIGINT NOT NULL DEFAULT 0;
ALTER TABLE app_user ADD COLUMN IF NOT EXISTS tid BIGINT NOT NULL DEFAULT 0;
ALTER TABLE app_user ADD COLUMN IF NOT EXISTS money NUMERIC(10,2) NOT NULL DEFAULT 0;
ALTER TABLE app_user ADD COLUMN IF NOT EXISTS score NUMERIC(10,2) NOT NULL DEFAULT 0;
ALTER TABLE app_user ADD COLUMN IF NOT EXISTS code VARCHAR(128) NOT NULL DEFAULT '';
ALTER TABLE app_user ADD COLUMN IF NOT EXISTS avatar VARCHAR(512) NOT NULL DEFAULT '';
ALTER TABLE app_user ADD COLUMN IF NOT EXISTS status_text VARCHAR(128) NOT NULL DEFAULT '';
ALTER TABLE app_user ADD COLUMN IF NOT EXISTS wechat_unionid VARCHAR(128) NOT NULL DEFAULT '';
ALTER TABLE app_user ADD COLUMN IF NOT EXISTS wechat_openid VARCHAR(128) NOT NULL DEFAULT '';
ALTER TABLE app_user ADD COLUMN IF NOT EXISTS version INTEGER NOT NULL DEFAULT 1;

CREATE INDEX IF NOT EXISTS idx_app_user_pid ON app_user (pid);
CREATE INDEX IF NOT EXISTS idx_app_user_tid ON app_user (tid);
CREATE INDEX IF NOT EXISTS idx_app_user_mobile ON app_user (mobile);

CREATE TABLE IF NOT EXISTS user_money_log (
  id BIGSERIAL PRIMARY KEY,
  user_id BIGINT NOT NULL,
  type INT NOT NULL DEFAULT 1,
  money NUMERIC(10,2) NOT NULL DEFAULT 0,
  before_money NUMERIC(10,2) NOT NULL DEFAULT 0,
  after_money NUMERIC(10,2) NOT NULL DEFAULT 0,
  note VARCHAR(255) NOT NULL DEFAULT '',
  source VARCHAR(255) NOT NULL DEFAULT '',
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_user_money_log_user_id ON user_money_log (user_id);
CREATE INDEX IF NOT EXISTS idx_user_money_log_created_at ON user_money_log (created_at);

CREATE TABLE IF NOT EXISTS user_score_log (
  id BIGSERIAL PRIMARY KEY,
  user_id BIGINT NOT NULL,
  type INT NOT NULL DEFAULT 1,
  score NUMERIC(10,2) NOT NULL DEFAULT 0,
  before_score NUMERIC(10,2) NOT NULL DEFAULT 0,
  after_score NUMERIC(10,2) NOT NULL DEFAULT 0,
  note VARCHAR(255) NOT NULL DEFAULT '',
  source VARCHAR(255) NOT NULL DEFAULT '',
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_user_score_log_user_id ON user_score_log (user_id);
CREATE INDEX IF NOT EXISTS idx_user_score_log_created_at ON user_score_log (created_at);

INSERT INTO admin_menu (id, parent_id, name, enname, title, path, component, menu_type, permission_code, icon, sort, visible, status, iframe, external, fixed_tag, show_tag, ext, created_at, updated_at) VALUES
(9, 0, 'user', 'User', '用户管理', '/user/list', 'user/list', 'menu', '', 'lucide:user-round', 9, TRUE, 1, '', '', 0, 0, '{}'::jsonb, NOW(), NOW()),
(161, 9, 'app-user-list', 'User List', '用户列表', '', '', 'button', 'app_user.list', '', 161, TRUE, 1, '', '', 0, 0, '{}'::jsonb, NOW(), NOW()),
(162, 9, 'app-user-save', 'User Save', '用户保存', '', '', 'button', 'app_user.save', '', 162, TRUE, 1, '', '', 0, 0, '{}'::jsonb, NOW(), NOW()),
(163, 9, 'app-user-delete', 'User Delete', '用户删除', '', '', 'button', 'app_user.delete', '', 163, TRUE, 1, '', '', 0, 0, '{}'::jsonb, NOW(), NOW()),
(164, 9, 'app-user-money', 'User Money', '余额调整', '', '', 'button', 'app_user.money', '', 164, TRUE, 1, '', '', 0, 0, '{}'::jsonb, NOW(), NOW()),
(165, 9, 'app-user-score', 'User Score', '积分调整', '', '', 'button', 'app_user.score', '', 165, TRUE, 1, '', '', 0, 0, '{}'::jsonb, NOW(), NOW())
ON CONFLICT DO NOTHING;

INSERT INTO admin_role_menu (role_id, menu_id, ext, created_at, updated_at)
SELECT 1, id, '{}'::jsonb, NOW(), NOW()
FROM admin_menu
WHERE id IN (9, 161, 162, 163, 164, 165)
ON CONFLICT DO NOTHING;
