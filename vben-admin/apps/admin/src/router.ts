import { createRouter, createWebHistory } from 'vue-router';

import { adminState, bootstrapAdminSession } from './auth';
import AttachmentView from './views/AttachmentView.vue';
import CrudView from './views/CrudView.vue';
import DashboardView from './views/DashboardView.vue';
import LoginView from './views/LoginView.vue';
import AdminLayout from './views/AdminLayout.vue';

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/login',
      name: 'admin-login',
      component: LoginView,
    },
    {
      path: '/',
      component: AdminLayout,
      children: [
        {
          path: '',
          redirect: '/dashboard',
        },
        {
          path: '/dashboard',
          name: 'dashboard',
          component: DashboardView,
          meta: { title: '工作台' },
        },
        {
          path: '/system/admin-user',
          component: CrudView,
          meta: { title: '管理员', module: 'admin_user' },
        },
        {
          path: '/system/admin-role',
          component: CrudView,
          meta: { title: '角色管理', module: 'admin_role' },
        },
        {
          path: '/system/admin-menu',
          component: CrudView,
          meta: { title: '菜单管理', module: 'admin_menu' },
        },
        {
          path: '/system/system-config',
          component: CrudView,
          meta: { title: '系统配置', module: 'system_config' },
        },
        {
          path: '/system/attachment',
          component: AttachmentView,
          meta: { title: '附件管理' },
        },
        {
          path: '/system/codegen',
          component: CrudView,
          meta: { title: '代码生成', module: 'codegen' },
        },
      ],
    },
  ],
});

router.beforeEach(async (to) => {
  if (to.path === '/login') {
    if (adminState.token) {
      if (!adminState.bootstrapped) {
        await bootstrapAdminSession();
      }
      if (adminState.token) {
        return '/dashboard';
      }
    }
    return true;
  }

  if (!adminState.token) {
    return '/login';
  }
  if (!adminState.bootstrapped) {
    const ok = await bootstrapAdminSession();
    if (!ok) {
      return '/login';
    }
  }
  return true;
});

export default router;
