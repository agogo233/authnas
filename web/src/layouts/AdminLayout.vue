<script setup lang="ts">
import { useRouter, useRoute } from 'vue-router'
import { NIcon } from 'naive-ui'
import {
  DashboardIcon,
  PeopleOutlineIcon,
  FolderOutlineIcon,
  ColorPaletteOutlineIcon,
  MailOutlineIcon,
  ShieldIcon,
  SettingsOutlineIcon,
  ReturnUpBackOutlineIcon,
} from './icons'

const router = useRouter()
const route = useRoute()

const navItems = [
  { label: '概览', path: '/admin', icon: DashboardIcon },
  { label: '用户', path: '/admin/users', icon: PeopleOutlineIcon },
  { label: '用户组', path: '/admin/groups', icon: FolderOutlineIcon },
  { label: '客户端', path: '/admin/clients', icon: ColorPaletteOutlineIcon },
  { label: '邀请', path: '/admin/invitations', icon: MailOutlineIcon },
  { label: '代理认证', path: '/admin/proxyauth', icon: ShieldIcon },
  { label: '系统设置', path: '/admin/settings', icon: SettingsOutlineIcon },
]

function isActive(path: string) {
  if (path === '/admin') {
    return route.path === '/admin'
  }
  return route.path.startsWith(path)
}

function navigateTo(path: string) {
  router.push(path)
}
</script>

<template>
  <div class="admin-layout">
    <aside class="admin-sidebar">
      <div class="sidebar-header">
        <div class="logo">
          <NIcon :component="DashboardIcon" :size="20" />
          <span>管理后台</span>
        </div>
      </div>
      <nav class="sidebar-nav">
        <button
          v-for="item in navItems"
          :key="item.path"
          class="nav-item"
          :class="{ active: isActive(item.path) }"
          @click="navigateTo(item.path)"
        >
          <NIcon :component="item.icon" size="18" />
          <span>{{ item.label }}</span>
          <span v-if="isActive(item.path)" class="active-indicator" />
        </button>
      </nav>
      <div class="sidebar-footer">
        <button class="nav-item" @click="navigateTo('/user/profile')">
          <NIcon :component="ReturnUpBackOutlineIcon" size="18" />
          <span>返回用户中心</span>
        </button>
      </div>
    </aside>
    <main class="admin-main">
      <div class="content-wrapper">
        <router-view v-slot="{ Component }">
          <transition name="fade-slide" mode="out-in">
            <component :is="Component" />
          </transition>
        </router-view>
      </div>
    </main>
  </div>
</template>

<style scoped>
.admin-layout {
  display: flex;
  min-height: 100vh;
  background: var(--bg);
}

.admin-sidebar {
  width: 240px;
  background: var(--bg-card);
  border-right: 1px solid var(--border);
  position: fixed;
  left: 0;
  top: 0;
  bottom: 0;
  display: flex;
  flex-direction: column;
  z-index: 100;
}

.sidebar-header {
  padding: 20px;
  border-bottom: 1px solid var(--border);
}

.logo {
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 16px;
  font-weight: 600;
  color: var(--text-h);
}

.sidebar-nav {
  flex: 1;
  padding: 16px 12px;
  display: flex;
  flex-direction: column;
  gap: 4px;
  overflow-y: auto;
}

.sidebar-footer {
  padding: 12px;
  border-top: 1px solid var(--border);
}

.nav-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 16px;
  border: none;
  background: transparent;
  border-radius: 10px;
  cursor: pointer;
  font-size: 14px;
  font-weight: 500;
  color: var(--text);
  transition: all 0.2s cubic-bezier(0.4, 0, 0.2, 1);
  position: relative;
  text-align: left;
  width: 100%;
}

.nav-item:hover {
  background: var(--accent-bg);
  color: var(--accent-solid);
}

.nav-item.active {
  background: var(--accent-bg);
  color: var(--accent-solid);
  font-weight: 600;
}

.active-indicator {
  position: absolute;
  left: 0;
  top: 50%;
  transform: translateY(-50%);
  width: 3px;
  height: 24px;
  background: var(--accent-solid);
  border-radius: 0 3px 3px 0;
}

.admin-main {
  flex: 1;
  margin-left: 240px;
  min-height: 100vh;
  display: flex;
  flex-direction: column;
}

.content-wrapper {
  flex: 1;
  padding: 32px 40px;
  max-width: 1200px;
  width: 100%;
  margin: 0 auto;
}

.fade-slide-enter-active,
.fade-slide-leave-active {
  transition: all 0.25s cubic-bezier(0.4, 0, 0.2, 1);
}

.fade-slide-enter-from {
  opacity: 0;
  transform: translateX(12px);
}

.fade-slide-leave-to {
  opacity: 0;
  transform: translateX(-12px);
}

@media (max-width: 1024px) {
  .admin-sidebar {
    width: 100%;
    height: auto;
    position: relative;
    border-right: none;
    border-bottom: 1px solid var(--border);
  }

  .admin-main {
    margin-left: 0;
  }

  .sidebar-nav {
    flex-direction: row;
    flex-wrap: wrap;
    padding: 12px;
    gap: 8px;
  }

  .sidebar-footer {
    display: none;
  }

  .content-wrapper {
    padding: 24px 20px;
  }
}
</style>
