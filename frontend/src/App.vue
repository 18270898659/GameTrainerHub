<template>
  <div class="app-layout">
    <!-- Left Sidebar -->
    <aside class="sidebar">
      <div class="app-brand">
        <div class="logo-icon">⚡</div>
        <div class="brand-text">
          <div class="brand-name">Trainer Hub</div>
          <div class="brand-sub">PC 修改器管理端</div>
        </div>
      </div>

      <nav class="nav-menu">
        <router-link to="/" class="nav-item">
          <el-icon><House /></el-icon>
          <span>首页</span>
        </router-link>

        <router-link to="/search" class="nav-item">
          <el-icon><Search /></el-icon>
          <span>游戏搜索</span>
        </router-link>

        <router-link to="/downloads" class="nav-item">
          <el-icon><Download /></el-icon>
          <span>下载管理</span>
          <span v-if="store.activeDownloadsCount > 0" class="nav-badge">
            {{ store.activeDownloadsCount }}
          </span>
        </router-link>

        <router-link to="/library" class="nav-item">
          <el-icon><Collection /></el-icon>
          <span>本地修改器</span>
        </router-link>

        <router-link to="/settings" class="nav-item">
          <el-icon><Setting /></el-icon>
          <span>客户端设置</span>
        </router-link>
      </nav>

      <div class="sidebar-footer">
        <div class="source-status">
          <span class="status-dot"></span>
          <span>FLiNG 源已连接</span>
        </div>
      </div>
    </aside>

    <!-- Main Content -->
    <main class="main-content">
      <header class="top-bar">
        <div class="window-title">{{ pageTitle }}</div>
      </header>

      <div class="page-container">
        <router-view />
      </div>
    </main>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted } from 'vue'
import { useRoute } from 'vue-router'
import { House, Search, Download, Collection, Setting } from '@element-plus/icons-vue'
import { useAppStore } from './stores/appStore'
import { api } from './api/wails'

const route = useRoute()
const store = useAppStore()

const pageTitle = computed(() => {
  switch (route.path) {
    case '/': return '首页概览'
    case '/search': return '中英双向修改器检索'
    case '/downloads': return '下载与任务管理'
    case '/library': return '本地修改器库与启动器'
    case '/settings': return '客户端配置'
    default: return 'Game Trainer Hub'
  }
})

let unregisterEvents: (() => void) | undefined

onMounted(async () => {
  await store.refreshDownloads()
  await store.refreshLocalTrainers()

  // 注册 Wails 事件流
  unregisterEvents = api.onDownloadProgress((task) => {
    store.handleProgressUpdate(task)
  })
})

onUnmounted(() => {
  if (unregisterEvents) {
    unregisterEvents()
  }
})
</script>

<style>
/* 全局暗色电竞主题基础样式 */
:root {
  --bg-dark: #0f172a;
  --bg-sidebar: #090d16;
  --accent-cyan: #38bdf8;
  --accent-hover: #0284c7;
  --text-main: #f8fafc;
  --text-muted: #94a3b8;
  --border-color: rgba(255, 255, 255, 0.08);
}

body {
  margin: 0;
  padding: 0;
  font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif;
  background-color: var(--bg-dark);
  color: var(--text-main);
  user-select: none;
  overflow: hidden;
}

.app-layout {
  display: flex;
  height: 100vh;
  width: 100vw;
  background-color: var(--bg-dark);
}

.sidebar {
  width: 230px;
  background-color: var(--bg-sidebar);
  border-right: 1px solid var(--border-color);
  display: flex;
  flex-direction: column;
  flex-shrink: 0;
}

.app-brand {
  padding: 24px 20px;
  display: flex;
  align-items: center;
  gap: 12px;
  border-bottom: 1px solid var(--border-color);
}

.logo-icon {
  font-size: 24px;
  background: linear-gradient(135deg, #0284c7, #38bdf8);
  width: 40px;
  height: 40px;
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: 0 0 15px rgba(56, 189, 248, 0.4);
}

.brand-name {
  font-size: 16px;
  font-weight: 700;
  color: #fff;
  letter-spacing: 0.5px;
}

.brand-sub {
  font-size: 11px;
  color: var(--text-muted);
}

.nav-menu {
  padding: 16px 12px;
  display: flex;
  flex-direction: column;
  gap: 6px;
  flex: 1;
}

.nav-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 10px 14px;
  color: var(--text-muted);
  text-decoration: none;
  border-radius: 6px;
  font-size: 14px;
  font-weight: 500;
  transition: all 0.2s ease;
  position: relative;
}

.nav-item:hover {
  color: #fff;
  background: rgba(255, 255, 255, 0.05);
}

.nav-item.router-link-exact-active {
  color: #fff;
  background: rgba(56, 189, 248, 0.15);
  font-weight: 600;
}

.nav-item.router-link-exact-active::before {
  content: '';
  position: absolute;
  left: 0;
  top: 6px;
  bottom: 6px;
  width: 3px;
  background: var(--accent-cyan);
  border-radius: 0 2px 2px 0;
}

.nav-badge {
  margin-left: auto;
  background: #ef4444;
  color: #fff;
  font-size: 10px;
  font-weight: 700;
  padding: 2px 6px;
  border-radius: 10px;
}

.sidebar-footer {
  padding: 16px 20px;
  border-top: 1px solid var(--border-color);
}

.source-status {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 12px;
  color: #10b981;
}

.status-dot {
  width: 8px;
  height: 8px;
  background: #10b981;
  border-radius: 50%;
  box-shadow: 0 0 8px #10b981;
}

.main-content {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.top-bar {
  height: 52px;
  padding: 0 24px;
  display: flex;
  align-items: center;
  border-bottom: 1px solid var(--border-color);
  background: rgba(15, 23, 42, 0.6);
  backdrop-filter: blur(8px);
}

.window-title {
  font-size: 15px;
  font-weight: 600;
  color: #e2e8f0;
}

.page-container {
  flex: 1;
  padding: 24px;
  overflow-y: auto;
}
</style>

