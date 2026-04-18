<template>
  <section class="dashboard">
    <article class="card hero">
      <p class="eyebrow">阶段一</p>
      <h2>后台运行时、RBAC 菜单、基础 CRUD 已接通</h2>
      <p>
        这个首页不再依赖旧仓库后端数据结构，当前只展示新脚手架的核心约束：`GET/POST`
        接口、`permission_code` 权限校验、自定义 RBAC。
      </p>
    </article>

    <article class="grid">
      <div class="card stat">
        <strong>认证</strong>
        <span>/admin-api/auth/login</span>
      </div>
      <div class="card stat">
        <strong>菜单</strong>
        <span>/admin-api/auth/menus</span>
      </div>
      <div class="card stat">
        <strong>权限</strong>
        <span>permission_code</span>
      </div>
    </article>
  </section>
</template>

<style scoped>
.dashboard {
  display: grid;
  gap: 20px;
}

.hero {
  padding: 26px;
}

.hero h2 {
  margin: 10px 0;
  font-size: 30px;
}

.eyebrow {
  margin: 0;
  color: #b15c2e;
  letter-spacing: 0.14em;
  text-transform: uppercase;
}

.grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 18px;
}

.stat {
  padding: 20px;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.stat span {
  color: rgba(20, 33, 61, 0.66);
}

@media (max-width: 900px) {
  .grid {
    grid-template-columns: 1fr;
  }
}
</style>
