INSERT INTO admin_role (id, name, code, status, ext, created_at, updated_at)
VALUES (1, '超级管理员', 'super_admin', 1, '{}'::jsonb, NOW(), NOW())
ON CONFLICT DO NOTHING;

INSERT INTO admin_user (id, username, password_hash, nickname, status, is_super, ext, created_at, updated_at)
VALUES (1, 'admin', '$2a$10$6OBdFu0q3mUyev14uVxBtu2IPhWzcjc609DjKGCsRFuoFOm5S0BSu', '超级管理员', 1, TRUE, '{}'::jsonb, NOW(), NOW())
ON CONFLICT DO NOTHING;

INSERT INTO admin_user_role (id, user_id, role_id, ext, created_at, updated_at)
VALUES (1, 1, 1, '{}'::jsonb, NOW(), NOW())
ON CONFLICT DO NOTHING;

INSERT INTO admin_menu (id, parent_id, name, title, path, component, menu_type, permission_code, icon, sort, visible, status, ext, created_at, updated_at) VALUES
(1, 0, 'dashboard', '工作台', '/dashboard', 'dashboard/index', 'menu', '', 'lucide:layout-dashboard', 1, TRUE, 1, '{}'::jsonb, NOW(), NOW()),
(2, 0, 'system', '系统管理', '/system', 'layout', 'menu', '', 'lucide:settings', 10, TRUE, 1, '{}'::jsonb, NOW(), NOW()),
(3, 2, 'admin-user', '管理员', '/system/admin-user', 'admin/admin/list', 'menu', '', 'lucide:users', 11, TRUE, 1, '{}'::jsonb, NOW(), NOW()),
(4, 2, 'admin-role', '角色管理', '/system/admin-role', 'admin/role/list', 'menu', '', 'lucide:shield', 12, TRUE, 1, '{}'::jsonb, NOW(), NOW()),
(5, 2, 'admin-menu', '菜单管理', '/system/admin-menu', 'admin/menu/list', 'menu', '', 'lucide:menu', 13, TRUE, 1, '{}'::jsonb, NOW(), NOW()),
(6, 2, 'system-config', '系统配置', '/system/system-config', 'system/config/index', 'menu', '', 'lucide:settings', 14, TRUE, 1, '{}'::jsonb, NOW(), NOW()),
(7, 2, 'attachment', '附件管理', '/system/attachment', 'system/attachment/index', 'menu', '', 'lucide:paperclip', 15, TRUE, 1, '{}'::jsonb, NOW(), NOW()),
(8, 2, 'codegen', '代码生成', '/system/codegen', 'system/gen/list', 'menu', '', 'lucide:code', 16, TRUE, 1, '{}'::jsonb, NOW(), NOW()),
(101, 3, 'admin-user-list', '管理员列表', '', '', 'button', 'admin_user.list', '', 101, TRUE, 1, '{}'::jsonb, NOW(), NOW()),
(102, 3, 'admin-user-save', '管理员保存', '', '', 'button', 'admin_user.save', '', 102, TRUE, 1, '{}'::jsonb, NOW(), NOW()),
(103, 3, 'admin-user-delete', '管理员删除', '', '', 'button', 'admin_user.delete', '', 103, TRUE, 1, '{}'::jsonb, NOW(), NOW()),
(111, 4, 'admin-role-list', '角色列表', '', '', 'button', 'admin_role.list', '', 111, TRUE, 1, '{}'::jsonb, NOW(), NOW()),
(112, 4, 'admin-role-save', '角色保存', '', '', 'button', 'admin_role.save', '', 112, TRUE, 1, '{}'::jsonb, NOW(), NOW()),
(113, 4, 'admin-role-delete', '角色删除', '', '', 'button', 'admin_role.delete', '', 113, TRUE, 1, '{}'::jsonb, NOW(), NOW()),
(121, 5, 'admin-menu-list', '菜单列表', '', '', 'button', 'admin_menu.list', '', 121, TRUE, 1, '{}'::jsonb, NOW(), NOW()),
(122, 5, 'admin-menu-save', '菜单保存', '', '', 'button', 'admin_menu.save', '', 122, TRUE, 1, '{}'::jsonb, NOW(), NOW()),
(123, 5, 'admin-menu-delete', '菜单删除', '', '', 'button', 'admin_menu.delete', '', 123, TRUE, 1, '{}'::jsonb, NOW(), NOW()),
(131, 6, 'system-config-list', '配置列表', '', '', 'button', 'system_config.list', '', 131, TRUE, 1, '{}'::jsonb, NOW(), NOW()),
(132, 6, 'system-config-save', '配置保存', '', '', 'button', 'system_config.save', '', 132, TRUE, 1, '{}'::jsonb, NOW(), NOW()),
(141, 7, 'attachment-list', '附件列表', '', '', 'button', 'attachment.list', '', 141, TRUE, 1, '{}'::jsonb, NOW(), NOW()),
(142, 7, 'attachment-upload', '附件上传', '', '', 'button', 'attachment.upload', '', 142, TRUE, 1, '{}'::jsonb, NOW(), NOW()),
(143, 7, 'attachment-delete', '附件删除', '', '', 'button', 'attachment.delete', '', 143, TRUE, 1, '{}'::jsonb, NOW(), NOW()),
(151, 8, 'codegen-list', '代码生成列表', '', '', 'button', 'codegen.list', '', 151, TRUE, 1, '{}'::jsonb, NOW(), NOW()),
(152, 8, 'codegen-save', '代码生成保存', '', '', 'button', 'codegen.save', '', 152, TRUE, 1, '{}'::jsonb, NOW(), NOW()),
(153, 8, 'codegen-delete', '代码生成删除', '', '', 'button', 'codegen.delete', '', 153, TRUE, 1, '{}'::jsonb, NOW(), NOW())
ON CONFLICT DO NOTHING;

INSERT INTO admin_role_menu (id, role_id, menu_id, ext, created_at, updated_at)
SELECT ROW_NUMBER() OVER (), 1, id, '{}'::jsonb, NOW(), NOW()
FROM admin_menu
ON CONFLICT DO NOTHING;

INSERT INTO system_config (id, config_key, config_name, config_value, remark, ext, created_at, updated_at)
VALUES
  (1, 'site.name', '站点名称', '{"value":"goweb-scaffold"}'::jsonb, '默认站点名称', '{}'::jsonb, NOW(), NOW()),
  (2, 'site.notice', '站点公告', '{"value":"欢迎使用 goweb-scaffold"}'::jsonb, '默认站点公告', '{}'::jsonb, NOW(), NOW())
ON CONFLICT DO NOTHING;
