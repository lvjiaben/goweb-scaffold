import type { RouteRecordRaw } from 'vue-router';

const routes: RouteRecordRaw[] = [
  {
    path: '/userinfo',
    name: 'UserInfo',
    component: () => import('#/views/home/userinfo.vue'),
    meta: {
      hideInMenu: true,
      title: '个人中心',
    },
  },
];

export default routes;
