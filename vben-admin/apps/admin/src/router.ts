import { createRouter, createWebHistory } from 'vue-router';

import { adminState, bootstrapAdminSession } from '@/auth';
import { collectMenuPaths, findFirstMenuPath, findMenuTrail } from '@/helpers';
import ForbiddenView from '@/views/ForbiddenView.vue';
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
          meta: { title: '工作台', requiresMenu: true },
        },
        {
          path: '/system/admin-user',
          component: AdminUserPage,
          meta: { title: '管理员', requiresMenu: true },
        },
        {
          path: '/system',
          redirect: '/system/admin-user',
        },
        {
          path: '/system/admin-role',
          component: AdminRolePage,
          meta: { title: '角色管理', requiresMenu: true },
        },
        {
          path: '/system/admin-menu',
          component: AdminMenuPage,
          meta: { title: '菜单管理', requiresMenu: true },
        },
        {
          path: '/system/system-config',
          component: SystemConfigPage,
          meta: { title: '系统配置', requiresMenu: true },
        },
        {
          path: '/system/attachment',
          component: AttachmentPage,
          meta: { title: '附件管理', requiresMenu: true },
        },
        {
          path: '/system/codegen',
          component: CodegenPage,
          meta: { title: '代码生成', requiresMenu: true },
        },
        {
          path: '/forbidden',
          component: ForbiddenView,
          meta: { title: '无权限', requiresMenu: false },
        },
      ],
    },
    {
      path: '/:pathMatch(.*)*',
      redirect: '/forbidden',
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
        return findFirstMenuPath(adminState.menus);
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

  if (to.path === '/forbidden') {
    return true;
  }

  if (to.matched.length === 0) {
    return findFirstMenuPath(adminState.menus);
  }

  if (to.meta.requiresMenu === false) {
    return true;
  }

  const allowedPaths = collectMenuPaths(adminState.menus);
  if (!allowedPaths.includes(to.path)) {
    return {
      path: '/forbidden',
      query: {
        from: to.fullPath,
      },
    };
  }
  return true;
});

router.afterEach((to) => {
  const trail = findMenuTrail(adminState.menus, to.path);
  const title = trail.at(-1)?.title || String(to.meta.title || 'Goweb Admin');
  document.title = `${title} - Goweb Admin`;
});

export default router;
