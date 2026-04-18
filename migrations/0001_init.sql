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
  title VARCHAR(128) NOT NULL,
  path VARCHAR(255) NOT NULL DEFAULT '',
  component VARCHAR(255) NOT NULL DEFAULT '',
  menu_type VARCHAR(32) NOT NULL,
  permission_code VARCHAR(128) NOT NULL DEFAULT '',
  icon VARCHAR(128) NOT NULL DEFAULT '',
  sort INTEGER NOT NULL DEFAULT 0,
  visible BOOLEAN NOT NULL DEFAULT TRUE,
  status SMALLINT NOT NULL DEFAULT 1,
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
  username VARCHAR(64) NOT NULL,
  password_hash VARCHAR(255) NOT NULL,
  nickname VARCHAR(128) NOT NULL DEFAULT '',
  email VARCHAR(128) NOT NULL DEFAULT '',
  mobile VARCHAR(32) NOT NULL DEFAULT '',
  status SMALLINT NOT NULL DEFAULT 1,
  ext JSONB DEFAULT '{}'::jsonb,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  deleted_at TIMESTAMPTZ NULL
);
CREATE UNIQUE INDEX IF NOT EXISTS ux_app_user_username ON app_user (username) WHERE deleted_at IS NULL;

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
