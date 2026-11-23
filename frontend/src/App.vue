<template>
  <el-config-provider :locale="currentElementLocale">
    <div class="app-layout">
      <!-- Left Sidebar -->
      <aside class="sidebar">
        <div class="app-brand">
          <div class="logo-icon">⚡</div>
          <div class="brand-text">
            <div class="brand-name">{{ t('brand.name') }}</div>
            <div class="brand-sub">{{ t('brand.sub') }}</div>
          </div>
        </div>

        <nav class="nav-menu">
          <router-link to="/" class="nav-item">
            <el-icon><House /></el-icon>
            <span>{{ t('nav.home') }}</span>
          </router-link>

          <router-link to="/search" class="nav-item">
            <el-icon><Search /></el-icon>
            <span>{{ t('nav.search') }}</span>
          </router-link>

          <router-link to="/downloads" class="nav-item">
            <el-icon><Download /></el-icon>
            <span>{{ t('nav.downloads') }}</span>
            <span v-if="store.activeDownloadsCount > 0" class="nav-badge">
              {{ store.activeDownloadsCount }}
            </span>
          </router-link>

          <router-link to="/library" class="nav-item">
            <el-icon><Collection /></el-icon>
            <span>{{ t('nav.library') }}</span>
          </router-link>

          <router-link to="/settings" class="nav-item">
            <el-icon><Setting /></el-icon>
            <span>{{ t('nav.settings') }}</span>
          </router-link>
        </nav>

        <div class="sidebar-footer">
          <div class="source-status">
            <span class="status-dot"></span>
            <span>{{ t('nav.sourceConnected') }}</span>
          </div>
        </div>
      </aside>

      <!-- Main Content -->
      <main class="main-content">
        <header class="top-bar">
          <div class="window-title">{{ pageTitle }}</div>
        </header>

        <div class="page-container">
          <router-view v-slot="{ Component }">
            <keep-alive include="HomeView,SearchView,DownloadsView,LibraryView">
              <component :is="Component" />
            </keep-alive>
          </router-view>
        </div>
      </main>
    </div>

    <!-- 全局游戏修改器详情与版本下载抽屉 -->
    <GameDetailDrawer />
  </el-config-provider>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { House, Search, Download, Collection, Setting } from '@element-plus/icons-vue'
import GameDetailDrawer from './components/GameDetailDrawer.vue'
import { useAppStore } from './stores/appStore'
import { api } from './api/wails'
import { t, currentElementLocale, setLanguage, type SupportedLanguage } from './i18n'

const route = useRoute()
const router = useRouter()
const store = useAppStore()

const pageTitle = computed(() => {
  switch (route.path) {
    case '/': return t('nav.pageTitleHome')
    case '/search': return t('nav.pageTitleSearch')
    case '/downloads': return t('nav.pageTitleDownloads')
    case '/library': return t('nav.pageTitleLibrary')
    case '/settings': return t('nav.pageTitleSettings')
    default: return 'Game Trainer Hub'
  }
})

let unregisterEvents: (() => void) | undefined
let unregisterComplete: (() => void) | undefined
let unregisterNavigate: (() => void) | undefined

onMounted(async () => {
  await store.refreshConfig()
  if (store.config?.language) {
    setLanguage(store.config.language as SupportedLanguage)
  }
  await store.refreshDownloads()
  await store.refreshHistory()
  await store.refreshLocalTrainers()

  // 注册 Wails 事件流
  unregisterEvents = api.onDownloadProgress((task) => {
    store.handleProgressUpdate(task)
  })

  unregisterComplete = api.onDownloadCompleted((task) => {
    store.handleProgressUpdate(task)
    store.refreshLocalTrainers()
    store.refreshDownloads()
    store.refreshHistory()
  })

  // 注册快捷键导航事件 (如 Ctrl+J 直达下载管理)
  unregisterNavigate = api.onNavigate((targetRoute) => {
    router.push(targetRoute)
  })
})

onUnmounted(() => {
  if (unregisterEvents) {
    unregisterEvents()
  }
  if (unregisterComplete) {
    unregisterComplete()
  }
  if (unregisterNavigate) {
    unregisterNavigate()
  }
})
</script>

<style>
/* 全局基础样式 */
body {
  margin: 0;
  padding: 0;
  font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif;
  background-color: var(--bg-app);
  color: var(--text-main);
  user-select: none;
  overflow: hidden;
  transition: background-color 0.25s ease, color 0.25s ease;
}

.app-layout {
  display: flex;
  height: 100vh;
  width: 100vw;
  background-color: var(--bg-app);
  transition: background-color 0.25s ease;
}

.sidebar {
  width: 230px;
  background-color: var(--bg-sidebar);
  border-right: 1px solid var(--border-color);
  display: flex;
  flex-direction: column;
  flex-shrink: 0;
  transition: background-color 0.25s ease, border-color 0.25s ease;
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
  color: #fff;
}

.brand-name {
  font-size: 16px;
  font-weight: 700;
  color: var(--text-main);
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
  color: var(--text-main);
  background: var(--bg-card-hover);
}

.nav-item.router-link-exact-active {
  color: var(--accent-cyan);
  background: var(--accent-tag-bg);
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
  background: var(--bg-topbar);
  backdrop-filter: blur(8px);
  transition: background-color 0.25s ease, border-color 0.25s ease;
}

.window-title {
  font-size: 15px;
  font-weight: 600;
  color: var(--text-main);
}

.page-container {
  flex: 1;
  padding: 24px;
  overflow-y: auto;
}
</style>

