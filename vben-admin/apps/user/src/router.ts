import { createRouter, createWebHistory } from 'vue-router';

import { bootstrapUserSession, userState } from './auth';
import HomeView from './views/HomeView.vue';
import LoginView from './views/LoginView.vue';
import ProfileView from './views/ProfileView.vue';
import UserLayout from './views/UserLayout.vue';

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/login',
      component: LoginView,
    },
    {
      path: '/',
      component: UserLayout,
      children: [
        {
          path: '',
          redirect: '/home',
        },
        {
          path: '/home',
          component: HomeView,
          meta: { title: '首页' },
        },
        {
          path: '/profile',
          component: ProfileView,
          meta: { title: '个人资料' },
        },
      ],
    },
  ],
});

router.beforeEach(async (to) => {
  if (to.path === '/login') {
    if (userState.token) {
      if (!userState.bootstrapped) {
        await bootstrapUserSession();
      }
      if (userState.token) {
        return '/home';
      }
    }
    return true;
  }

  if (!userState.token) {
    return '/login';
  }
  if (!userState.bootstrapped) {
    const ok = await bootstrapUserSession();
    if (!ok) {
      return '/login';
    }
  }
  return true;
});

export default router;
