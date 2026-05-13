<script setup lang="ts">
import { useRouter, useRoute } from 'vue-router'
import { NIcon } from 'naive-ui'
import { useAuthStore } from '@/stores/auth'
import { UserIcon, ShieldIcon, KeyIcon, DashboardIcon, LogOutIcon } from './icons'

const router = useRouter()
const route = useRoute()
const authStore = useAuthStore()

const navItems = [
  { label: '编辑资料', path: '/user/profile', icon: UserIcon },
  { label: '安全设置', path: '/user/security', icon: ShieldIcon },
  { label: '通行密钥', path: '/user/passkeys', icon: KeyIcon },
]

function isActive(path: string) {
  return route.path === path
}

function navigateTo(path: string) {
  router.push(path)
}

function handleLogout() {
  authStore.logout()
  router.push('/login')
}
</script>

<template>
  <div class="user-layout">
    <aside class="user-sidebar">
      <div class="sidebar-header">
        <span class="sidebar-title">导航菜单</span>
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
        <button
          v-if="authStore.isAdmin"
          class="nav-item"
          :class="{ active: route.path.startsWith('/admin') }"
          @click="navigateTo('/admin')"
        >
          <NIcon :component="DashboardIcon" size="18" />
          <span>管理面板</span>
          <span v-if="route.path.startsWith('/admin')" class="active-indicator" />
        </button>
      </nav>
      <div class="sidebar-footer">
        <button class="nav-item logout-btn" @click="handleLogout">
          <NIcon :component="LogOutIcon" size="18" />
          <span>退出登录</span>
        </button>
      </div>
    </aside>
    <main class="user-main">
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
.user-layout {
  display: flex;
  min-height: 100vh;
  background: var(--bg);
}

.user-sidebar {
  width: 220px;
  background: var(--bg-card);
  padding: 0;
  position: fixed;
  left: 0;
  top: 0;
  bottom: 0;
  border-right: 1px solid var(--border);
  display: flex;
  flex-direction: column;
  z-index: 100;
}

.sidebar-header {
  padding: 24px 20px 16px;
  border-bottom: 1px solid var(--border);
}

.sidebar-title {
  font-size: 13px;
  font-weight: 600;
  color: var(--text);
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.sidebar-nav {
  display: flex;
  flex-direction: column;
  padding: 16px 12px;
  gap: 4px;
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

.sidebar-footer {
  margin-top: auto;
  padding: 16px 12px;
  border-top: 1px solid var(--border);
}

.logout-btn {
  color: #9ca3af;
}

.logout-btn:hover {
  background: rgba(239, 68, 68, 0.1);
  color: #ef4444;
}

.user-main {
  flex: 1;
  margin-left: 220px;
  min-height: 100vh;
  display: flex;
  flex-direction: column;
}

.content-wrapper {
  flex: 1;
  padding: 32px 40px;
  max-width: 1000px;
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

@media (max-width: 768px) {
  .user-sidebar {
    width: 100%;
    height: auto;
    position: relative;
    border-right: none;
    border-bottom: 1px solid var(--border);
  }

  .user-main {
    margin-left: 0;
  }

  .content-wrapper {
    padding: 24px 20px;
  }

  .sidebar-nav {
    flex-direction: row;
    flex-wrap: wrap;
    padding: 12px;
    gap: 8px;
  }

  .nav-item {
    flex: 1 1 auto;
    min-width: 100px;
    justify-content: center;
    padding: 10px 12px;
  }

  .active-indicator {
    display: none;
  }
}
</style>
