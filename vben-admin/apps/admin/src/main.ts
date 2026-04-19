import { createApp } from 'vue';

import App from './App.vue';
import { clearAdminSession } from './auth';
import { adminForbiddenEvent, adminSessionExpiredEvent } from './events';
import { notifyWarning } from './notify';
import router from './router';
import './styles.css';

window.addEventListener(adminSessionExpiredEvent, async () => {
  clearAdminSession();
  notifyWarning('登录状态已失效，请重新登录');
  if (router.currentRoute.value.path !== '/login') {
    await router.replace('/login');
  }
});

window.addEventListener(adminForbiddenEvent, () => {
  notifyWarning('当前账号无权执行该操作');
});

createApp(App).use(router).mount('#app');
