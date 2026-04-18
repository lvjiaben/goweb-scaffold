<template>
  <section class="card home">
    <p class="kicker">phase one</p>
    <h1>用户端只保留登录、资料、密码修改这条闭环</h1>
    <p>
      这里直接请求 `/api/auth/*` 和 `/api/user/*`，不包含 codegen，不引入后台菜单或后台生成页。
    </p>
  </section>
</template>

<style scoped>
.home {
  padding: 26px;
}

.kicker {
  margin: 0;
  color: #1d4ed8;
  letter-spacing: 0.16em;
  text-transform: uppercase;
}

.home h1 {
  margin: 12px 0;
  font-size: clamp(28px, 5vw, 54px);
  line-height: 1;
}
</style>
