<script setup lang="ts">
import { computed } from 'vue';

import { adminState } from '@/auth';

const quickLinks = [
  { title: '管理员', path: '/system/admin-user', desc: '维护后台账号、角色绑定与状态。' },
  { title: '角色管理', path: '/system/admin-role', desc: '按菜单按钮和 permission_code 授权。' },
  { title: '菜单管理', path: '/system/admin-menu', desc: '维护菜单树与按钮权限节点。' },
  { title: '系统配置', path: '/system/system-config', desc: '统一管理 system_config 配置项。' },
  { title: '附件管理', path: '/system/attachment', desc: '上传文件、查看 URL 与删除附件。' },
  { title: '代码生成', path: '/system/codegen', desc: '读取数据表结构并保存生成配置历史。' },
];

const accessCount = computed(() => adminState.user?.access_codes.length || 0);
const roleCount = computed(() => adminState.user?.role_ids.length || 0);
</script>

<template>
  <section class="page-stack">
    <article class="card hero-card">
      <div class="hero-card__copy">
        <p class="eyebrow">Stage 2</p>
        <h2>admin 端已切到正式 CRUD 页面</h2>
        <p class="hero-card__desc">
          后台继续基于 Go + net/http + PostgreSQL + GORM，前端菜单来自
          <code>/admin-api/auth/menus</code>，按钮显隐按 <code>permission_code</code> 控制。
        </p>
      </div>
      <div class="hero-card__meta">
        <div class="hero-card__user">
          <strong>{{ adminState.user?.nickname || adminState.user?.username }}</strong>
          <span>{{ adminState.user?.is_super ? '超级管理员' : '后台管理员' }}</span>
        </div>
        <div class="hero-card__stats">
          <article class="card inset-card">
            <span>角色数</span>
            <strong>{{ roleCount }}</strong>
          </article>
          <article class="card inset-card">
            <span>权限码</span>
            <strong>{{ accessCount }}</strong>
          </article>
        </div>
      </div>
    </article>

    <section class="summary-grid">
      <article class="card stat-card">
        <span class="stat-card__label">后台认证</span>
        <strong>POST /admin-api/auth/login</strong>
        <p>管理员登录、会话写入、JWT 签发。</p>
      </article>
      <article class="card stat-card">
        <span class="stat-card__label">RBAC</span>
        <strong>permission_code</strong>
        <p>按钮和接口统一按权限码校验，不按 URL 模糊匹配。</p>
      </article>
      <article class="card stat-card">
        <span class="stat-card__label">代码生成</span>
        <strong>元数据阶段</strong>
        <p>读取 PostgreSQL 表结构，保存生成历史，预览路由和接口元数据。</p>
      </article>
    </section>

    <section class="card page-card">
      <div class="section-heading">
        <div>
          <h3>功能入口</h3>
          <p>本阶段重点是把 admin 后台从 JSON 面板升级为正式页面。</p>
        </div>
      </div>
      <div class="quick-grid">
        <RouterLink v-for="item in quickLinks" :key="item.path" :to="item.path" class="quick-link card">
          <strong>{{ item.title }}</strong>
          <p>{{ item.desc }}</p>
          <span>{{ item.path }}</span>
        </RouterLink>
      </div>
    </section>
  </section>
</template>
