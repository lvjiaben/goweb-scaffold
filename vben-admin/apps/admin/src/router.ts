import { createRouter, createWebHistory } from 'vue-router';

import { adminState, bootstrapAdminSession } from '@/auth';
import DashboardHomePage from '@/views/dashboard/DashboardHomePage.vue';
import LoginView from '@/views/LoginView.vue';
import AdminLayout from '@/views/AdminLayout.vue';
import AdminMenuPage from '@/views/system/AdminMenuPage.vue';
import AdminRolePage from '@/views/system/AdminRolePage.vue';
import AdminUserPage from '@/views/system/AdminUserPage.vue';
import AttachmentPage from '@/views/system/AttachmentPage.vue';
import CodegenPage from '@/views/system/CodegenPage.vue';
import SystemConfigPage from '@/views/system/SystemConfigPage.vue';

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
        { path: '', redirect: '/dashboard' },
        {
          path: '/dashboard',
          name: 'dashboard',
          component: DashboardHomePage,
          meta: { title: '工作台' },
        },
        {
          path: '/system/admin-user',
          component: AdminUserPage,
          meta: { title: '管理员' },
        },
        {
          path: '/system',
          redirect: '/system/admin-user',
        },
        {
          path: '/system/admin-role',
          component: AdminRolePage,
          meta: { title: '角色管理' },
        },
        {
          path: '/system/admin-menu',
          component: AdminMenuPage,
          meta: { title: '菜单管理' },
        },
        {
          path: '/system/system-config',
          component: SystemConfigPage,
          meta: { title: '系统配置' },
        },
        {
          path: '/system/attachment',
          component: AttachmentPage,
          meta: { title: '附件管理' },
        },
        {
          path: '/system/codegen',
          component: CodegenPage,
          meta: { title: '代码生成' },
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
