CREATE TABLE IF NOT EXISTS admin_user (
  id BIGSERIAL PRIMARY KEY,
  username VARCHAR(64) NOT NULL,
  password_hash VARCHAR(255) NOT NULL,
  nickname VARCHAR(128) NOT NULL DEFAULT '',
  status SMALLINT NOT NULL DEFAULT 1,
  is_super BOOLEAN NOT NULL DEFAULT FALSE,
  last_login_at TIMESTAMPTZ NULL,
  last_login_ip VARCHAR(64) NOT NULL DEFAULT '',
  ext JSONB DEFAULT '{}'::jsonb,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  deleted_at TIMESTAMPTZ NULL
);
CREATE UNIQUE INDEX IF NOT EXISTS ux_admin_user_username ON admin_user (username) WHERE deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS admin_role (
  id BIGSERIAL PRIMARY KEY,
  name VARCHAR(128) NOT NULL,
  code VARCHAR(128) NOT NULL,
  status SMALLINT NOT NULL DEFAULT 1,
  ext JSONB DEFAULT '{}'::jsonb,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  deleted_at TIMESTAMPTZ NULL
);
CREATE UNIQUE INDEX IF NOT EXISTS ux_admin_role_code ON admin_role (code) WHERE deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS admin_menu (
  id BIGSERIAL PRIMARY KEY,
  parent_id BIGINT NOT NULL DEFAULT 0,
  name VARCHAR(128) NOT NULL,
  enname VARCHAR(128) NOT NULL DEFAULT '',
  title VARCHAR(128) NOT NULL,
  path VARCHAR(255) NOT NULL DEFAULT '',
  component VARCHAR(255) NOT NULL DEFAULT '',
  menu_type VARCHAR(32) NOT NULL,
  permission_code VARCHAR(128) NOT NULL DEFAULT '',
  iframe VARCHAR(512) NOT NULL DEFAULT '',
  external VARCHAR(512) NOT NULL DEFAULT '',
  icon VARCHAR(128) NOT NULL DEFAULT '',
  sort INTEGER NOT NULL DEFAULT 0,
  visible BOOLEAN NOT NULL DEFAULT TRUE,
  status SMALLINT NOT NULL DEFAULT 1,
  fixed_tag SMALLINT NOT NULL DEFAULT 0,
  show_tag SMALLINT NOT NULL DEFAULT 0,
  ext JSONB DEFAULT '{}'::jsonb,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  deleted_at TIMESTAMPTZ NULL
);
CREATE INDEX IF NOT EXISTS idx_admin_menu_parent_id ON admin_menu (parent_id);
CREATE INDEX IF NOT EXISTS idx_admin_menu_permission_code ON admin_menu (permission_code);

CREATE TABLE IF NOT EXISTS admin_user_role (
  id BIGSERIAL PRIMARY KEY,
  user_id BIGINT NOT NULL,
  role_id BIGINT NOT NULL,
  ext JSONB DEFAULT '{}'::jsonb,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  deleted_at TIMESTAMPTZ NULL
);
CREATE UNIQUE INDEX IF NOT EXISTS ux_admin_user_role_unique ON admin_user_role (user_id, role_id) WHERE deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS admin_role_menu (
  id BIGSERIAL PRIMARY KEY,
  role_id BIGINT NOT NULL,
  menu_id BIGINT NOT NULL,
  ext JSONB DEFAULT '{}'::jsonb,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  deleted_at TIMESTAMPTZ NULL
);
CREATE UNIQUE INDEX IF NOT EXISTS ux_admin_role_menu_unique ON admin_role_menu (role_id, menu_id) WHERE deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS admin_login_log (
  id BIGSERIAL PRIMARY KEY,
  admin_user_id BIGINT NOT NULL DEFAULT 0,
  username VARCHAR(64) NOT NULL DEFAULT '',
  ip VARCHAR(64) NOT NULL DEFAULT '',
  user_agent TEXT NOT NULL DEFAULT '',
  success BOOLEAN NOT NULL DEFAULT FALSE,
  remark TEXT NOT NULL DEFAULT '',
  ext JSONB DEFAULT '{}'::jsonb,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  deleted_at TIMESTAMPTZ NULL
);

CREATE TABLE IF NOT EXISTS admin_session (
  id BIGSERIAL PRIMARY KEY,
  admin_user_id BIGINT NOT NULL,
  expires_at TIMESTAMPTZ NOT NULL,
  last_seen_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  user_agent TEXT NOT NULL DEFAULT '',
  ip VARCHAR(64) NOT NULL DEFAULT '',
  ext JSONB DEFAULT '{}'::jsonb,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  deleted_at TIMESTAMPTZ NULL
);
CREATE INDEX IF NOT EXISTS idx_admin_session_user_id ON admin_session (admin_user_id);

CREATE TABLE IF NOT EXISTS app_user (
  id BIGSERIAL PRIMARY KEY,
  pid BIGINT NOT NULL DEFAULT 0,
  tid BIGINT NOT NULL DEFAULT 0,
  username VARCHAR(64) NOT NULL,
  password_hash VARCHAR(255) NOT NULL,
  nickname VARCHAR(128) NOT NULL DEFAULT '',
  email VARCHAR(128) NOT NULL DEFAULT '',
  mobile VARCHAR(32) NOT NULL DEFAULT '',
  avatar VARCHAR(512) NOT NULL DEFAULT '',
  code VARCHAR(128) NOT NULL DEFAULT '',
  status SMALLINT NOT NULL DEFAULT 1,
  status_text VARCHAR(128) NOT NULL DEFAULT '',
  money NUMERIC(10,2) NOT NULL DEFAULT 0,
  score NUMERIC(10,2) NOT NULL DEFAULT 0,
  wechat_unionid VARCHAR(128) NOT NULL DEFAULT '',
  wechat_openid VARCHAR(128) NOT NULL DEFAULT '',
  version INTEGER NOT NULL DEFAULT 1,
  ext JSONB DEFAULT '{}'::jsonb,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  deleted_at TIMESTAMPTZ NULL
);
CREATE UNIQUE INDEX IF NOT EXISTS ux_app_user_username ON app_user (username) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_app_user_pid ON app_user (pid);
CREATE INDEX IF NOT EXISTS idx_app_user_tid ON app_user (tid);
CREATE INDEX IF NOT EXISTS idx_app_user_mobile ON app_user (mobile);

CREATE TABLE IF NOT EXISTS app_user_session (
  id BIGSERIAL PRIMARY KEY,
  app_user_id BIGINT NOT NULL,
  expires_at TIMESTAMPTZ NOT NULL,
  last_seen_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  user_agent TEXT NOT NULL DEFAULT '',
  ip VARCHAR(64) NOT NULL DEFAULT '',
  ext JSONB DEFAULT '{}'::jsonb,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  deleted_at TIMESTAMPTZ NULL
);
CREATE INDEX IF NOT EXISTS idx_app_user_session_user_id ON app_user_session (app_user_id);

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

CREATE TABLE IF NOT EXISTS system_config (
  id BIGSERIAL PRIMARY KEY,
  config_key VARCHAR(128) NOT NULL,
  config_name VARCHAR(128) NOT NULL,
  config_value JSONB DEFAULT '{}'::jsonb,
  remark TEXT NOT NULL DEFAULT '',
  ext JSONB DEFAULT '{}'::jsonb,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  deleted_at TIMESTAMPTZ NULL
);
CREATE UNIQUE INDEX IF NOT EXISTS ux_system_config_key ON system_config (config_key) WHERE deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS file_attachment (
  id BIGSERIAL PRIMARY KEY,
  original_name VARCHAR(255) NOT NULL,
  saved_name VARCHAR(255) NOT NULL,
  file_path VARCHAR(512) NOT NULL,
  file_url VARCHAR(512) NOT NULL,
  file_ext VARCHAR(32) NOT NULL DEFAULT '',
  mime_type VARCHAR(128) NOT NULL DEFAULT '',
  file_size BIGINT NOT NULL DEFAULT 0,
  uploader_kind VARCHAR(32) NOT NULL DEFAULT '',
  uploader_id BIGINT NOT NULL DEFAULT 0,
  ext JSONB DEFAULT '{}'::jsonb,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  deleted_at TIMESTAMPTZ NULL
);

CREATE TABLE IF NOT EXISTS codegen_history (
  id BIGSERIAL PRIMARY KEY,
  module_name VARCHAR(128) NOT NULL DEFAULT '',
  table_name VARCHAR(128) NOT NULL DEFAULT '',
  status VARCHAR(32) NOT NULL DEFAULT '',
  payload JSONB DEFAULT '{}'::jsonb,
  remark TEXT NOT NULL DEFAULT '',
  ext JSONB DEFAULT '{}'::jsonb,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  deleted_at TIMESTAMPTZ NULL
);

INSERT INTO admin_role (id, name, code, status, ext, created_at, updated_at)
VALUES (1, '超级管理员', 'super_admin', 1, '{}'::jsonb, NOW(), NOW())
ON CONFLICT DO NOTHING;

INSERT INTO admin_user (id, username, password_hash, nickname, status, is_super, ext, created_at, updated_at)
VALUES (1, 'admin', '$2a$10$6OBdFu0q3mUyev14uVxBtu2IPhWzcjc609DjKGCsRFuoFOm5S0BSu', '超级管理员', 1, TRUE, '{}'::jsonb, NOW(), NOW())
ON CONFLICT DO NOTHING;

INSERT INTO admin_user_role (id, user_id, role_id, ext, created_at, updated_at)
VALUES (1, 1, 1, '{}'::jsonb, NOW(), NOW())
ON CONFLICT DO NOTHING;

INSERT INTO admin_menu (id, parent_id, name, enname, title, path, component, menu_type, permission_code, icon, sort, visible, status, iframe, external, fixed_tag, show_tag, ext, created_at, updated_at) VALUES
(1, 0, 'dashboard', 'Dashboard', '工作台', '/dashboard', 'dashboard/index', 'menu', '', 'lucide:layout-dashboard', 100, TRUE, 1, '', '', 0, 0, '{}'::jsonb, NOW(), NOW()),
(2, 0, 'system', 'System', '系统管理', '/system', 'layout', 'menu', '', 'lucide:settings', 90, TRUE, 1, '', '', 0, 0, '{}'::jsonb, NOW(), NOW()),
(3, 2, 'admin-user', 'Admins', '管理员', '/system/admin-user', 'admin/admin/list', 'menu', '', 'lucide:users', 89, TRUE, 1, '', '', 0, 0, '{}'::jsonb, NOW(), NOW()),
(4, 2, 'admin-role', 'Roles', '角色管理', '/system/admin-role', 'admin/role/list', 'menu', '', 'lucide:shield', 88, TRUE, 1, '', '', 0, 0, '{}'::jsonb, NOW(), NOW()),
(5, 2, 'admin-menu', 'Menus', '菜单管理', '/system/admin-menu', 'admin/menu/list', 'menu', '', 'lucide:menu', 87, TRUE, 1, '', '', 0, 0, '{}'::jsonb, NOW(), NOW()),
(6, 2, 'system-config', 'Configs', '系统配置', '/system/system-config', 'system/config/index', 'menu', '', 'lucide:settings', 86, TRUE, 1, '', '', 0, 0, '{}'::jsonb, NOW(), NOW()),
(7, 2, 'attachment', 'Attachments', '附件管理', '/system/attachment', 'system/attachment/index', 'menu', '', 'lucide:paperclip', 85, TRUE, 1, '', '', 0, 0, '{}'::jsonb, NOW(), NOW()),
(8, 2, 'codegen', 'Codegen', '代码生成', '/system/codegen', 'system/gen/list', 'menu', '', 'lucide:code', 84, TRUE, 1, '', '', 0, 0, '{}'::jsonb, NOW(), NOW()),
(9, 0, 'user', 'User', '用户管理', '/user/list', 'user/list', 'menu', '', 'lucide:user-round', 80, TRUE, 1, '', '', 0, 0, '{}'::jsonb, NOW(), NOW()),
(101, 3, 'admin-user-list', 'Admin List', '管理员列表', '', '', 'button', 'admin_user.list', '', 101, TRUE, 1, '', '', 0, 0, '{}'::jsonb, NOW(), NOW()),
(102, 3, 'admin-user-save', 'Admin Save', '管理员保存', '', '', 'button', 'admin_user.save', '', 102, TRUE, 1, '', '', 0, 0, '{}'::jsonb, NOW(), NOW()),
(103, 3, 'admin-user-delete', 'Admin Delete', '管理员删除', '', '', 'button', 'admin_user.delete', '', 103, TRUE, 1, '', '', 0, 0, '{}'::jsonb, NOW(), NOW()),
(111, 4, 'admin-role-list', 'Role List', '角色列表', '', '', 'button', 'admin_role.list', '', 111, TRUE, 1, '', '', 0, 0, '{}'::jsonb, NOW(), NOW()),
(112, 4, 'admin-role-save', 'Role Save', '角色保存', '', '', 'button', 'admin_role.save', '', 112, TRUE, 1, '', '', 0, 0, '{}'::jsonb, NOW(), NOW()),
(113, 4, 'admin-role-delete', 'Role Delete', '角色删除', '', '', 'button', 'admin_role.delete', '', 113, TRUE, 1, '', '', 0, 0, '{}'::jsonb, NOW(), NOW()),
(121, 5, 'admin-menu-list', 'Menu List', '菜单列表', '', '', 'button', 'admin_menu.list', '', 121, TRUE, 1, '', '', 0, 0, '{}'::jsonb, NOW(), NOW()),
(122, 5, 'admin-menu-save', 'Menu Save', '菜单保存', '', '', 'button', 'admin_menu.save', '', 122, TRUE, 1, '', '', 0, 0, '{}'::jsonb, NOW(), NOW()),
(123, 5, 'admin-menu-delete', 'Menu Delete', '菜单删除', '', '', 'button', 'admin_menu.delete', '', 123, TRUE, 1, '', '', 0, 0, '{}'::jsonb, NOW(), NOW()),
(131, 6, 'system-config-list', 'Config List', '配置列表', '', '', 'button', 'system_config.list', '', 131, TRUE, 1, '', '', 0, 0, '{}'::jsonb, NOW(), NOW()),
(132, 6, 'system-config-save', 'Config Save', '配置保存', '', '', 'button', 'system_config.save', '', 132, TRUE, 1, '', '', 0, 0, '{}'::jsonb, NOW(), NOW()),
(141, 7, 'attachment-list', 'Attachment List', '附件列表', '', '', 'button', 'attachment.list', '', 141, TRUE, 1, '', '', 0, 0, '{}'::jsonb, NOW(), NOW()),
(142, 7, 'attachment-upload', 'Attachment Upload', '附件上传', '', '', 'button', 'attachment.upload', '', 142, TRUE, 1, '', '', 0, 0, '{}'::jsonb, NOW(), NOW()),
(143, 7, 'attachment-delete', 'Attachment Delete', '附件删除', '', '', 'button', 'attachment.delete', '', 143, TRUE, 1, '', '', 0, 0, '{}'::jsonb, NOW(), NOW()),
(151, 8, 'codegen-list', 'Codegen List', '代码生成列表', '', '', 'button', 'codegen.list', '', 151, TRUE, 1, '', '', 0, 0, '{}'::jsonb, NOW(), NOW()),
(152, 8, 'codegen-save', 'Codegen Save', '代码生成保存', '', '', 'button', 'codegen.save', '', 152, TRUE, 1, '', '', 0, 0, '{}'::jsonb, NOW(), NOW()),
(153, 8, 'codegen-delete', 'Codegen Delete', '代码生成删除', '', '', 'button', 'codegen.delete', '', 153, TRUE, 1, '', '', 0, 0, '{}'::jsonb, NOW(), NOW()),
(161, 9, 'app-user-list', 'User List', '用户列表', '', '', 'button', 'app_user.list', '', 161, TRUE, 1, '', '', 0, 0, '{}'::jsonb, NOW(), NOW()),
(162, 9, 'app-user-save', 'User Save', '用户保存', '', '', 'button', 'app_user.save', '', 162, TRUE, 1, '', '', 0, 0, '{}'::jsonb, NOW(), NOW()),
(163, 9, 'app-user-delete', 'User Delete', '用户删除', '', '', 'button', 'app_user.delete', '', 163, TRUE, 1, '', '', 0, 0, '{}'::jsonb, NOW(), NOW()),
(164, 9, 'app-user-money', 'User Money', '余额调整', '', '', 'button', 'app_user.money', '', 164, TRUE, 1, '', '', 0, 0, '{}'::jsonb, NOW(), NOW()),
(165, 9, 'app-user-score', 'User Score', '积分调整', '', '', 'button', 'app_user.score', '', 165, TRUE, 1, '', '', 0, 0, '{}'::jsonb, NOW(), NOW())
ON CONFLICT DO NOTHING;

INSERT INTO admin_role_menu (role_id, menu_id, ext, created_at, updated_at)
SELECT 1, id, '{}'::jsonb, NOW(), NOW()
FROM admin_menu
ON CONFLICT DO NOTHING;

INSERT INTO system_config (id, config_key, config_name, config_value, remark, ext, created_at, updated_at)
VALUES
  (1, 'site.name', '站点名称', '{"value":"goweb-scaffold"}'::jsonb, '默认站点名称', '{}'::jsonb, NOW(), NOW()),
  (2, 'site.notice', '站点公告', '{"value":"欢迎使用 goweb-scaffold"}'::jsonb, '默认站点公告', '{}'::jsonb, NOW(), NOW())
ON CONFLICT DO NOTHING;
