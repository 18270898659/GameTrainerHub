<template>
  <div class="home-view">
    <!-- Hero Banner -->
    <div class="hero-banner">
      <div class="hero-content">
        <h1 class="hero-title">{{ t('home.heroTitle') }}</h1>
        <p class="hero-subtitle">{{ t('home.heroSubtitle') }}</p>
        
        <div class="hero-search-box">
          <el-input
            v-model="heroQuery"
            size="large"
            :placeholder="t('home.searchPlaceholder')"
            clearable
            @keyup.enter="handleHeroSearch"
          >
            <template #prefix>
              <el-icon><Search /></el-icon>
            </template>
            <template #append>
              <el-button type="primary" @click="handleHeroSearch">{{ t('home.searchBtn') }}</el-button>
            </template>
          </el-input>
        </div>
      </div>
    </div>

    <!-- Quick Stats -->
    <div class="stats-grid">
      <div class="stat-card">
        <div class="stat-val">767+</div>
        <div class="stat-label">{{ t('home.statsIndexedGames') }}</div>
      </div>
      <div class="stat-card">
        <div class="stat-val">{{ store.localTrainers.length }}</div>
        <div class="stat-label">{{ t('home.statsReadyTrainers') }}</div>
      </div>
      <div class="stat-card">
        <div class="stat-val">{{ store.activeDownloadsCount }}</div>
        <div class="stat-label">{{ t('home.statsActiveDownloads') }}</div>
      </div>
      <div class="stat-card">
        <div class="stat-val">FLiNG</div>
        <div class="stat-label">{{ t('home.statsOfficialSource') }}</div>
      </div>
    </div>

    <!-- Latest Updated Grid -->
    <div class="section-header">
      <div class="section-title">
        <el-icon><Clock /></el-icon>
        <span>{{ t('home.latestTitle') }}</span>
      </div>
      <el-button size="small" :loading="store.isSyncing" @click="triggerSync" plain>
        <el-icon><Refresh /></el-icon> {{ t('home.syncBtn') }}
      </el-button>
    </div>

    <div v-loading="loading" class="trainers-grid">
      <div
        v-for="item in store.latestTrainers"
        :key="item.slug"
        class="trainer-card"
        @click="goToDetail(item)"
      >
        <div class="card-thumb">
          <img :src="item.cover_url || defaultThumb" :alt="item.game_name_en" />
          <div class="badge-source">{{ item.source_id.toUpperCase() }}</div>
        </div>
        <div class="card-info">
          <div class="game-name-cn">{{ item.game_name_cn || item.game_name_en }}</div>
          <div class="game-name-en">{{ item.game_name_en }}</div>
          <div class="card-footer">
            <span class="update-time">{{ formatDate(item.updated_at) }}</span>
            <el-tag v-if="item.is_downloaded" size="small" type="success">{{ t('home.tagDownloaded') }}</el-tag>
            <el-tag v-else size="small" type="info">{{ t('home.tagOnline') }}</el-tag>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { Search, Clock, Refresh } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import { useAppStore } from '../stores/appStore'
import { api } from '../api/wails'
import { t } from '../i18n'
import type { SearchResult } from '../types'

const router = useRouter()
const store = useAppStore()
const heroQuery = ref('')
const loading = ref(false)
const defaultThumb = 'https://flingtrainer.com/wp-content/uploads/2020/12/header-7-200x200.jpg'

onMounted(async () => {
  loading.value = true
  await Promise.all([
    store.refreshLatest(),
    store.refreshLocalTrainers(),
    store.refreshDownloads()
  ])
  loading.value = false
})

function handleHeroSearch() {
  if (heroQuery.value.trim()) {
    router.push({ path: '/search', query: { q: heroQuery.value.trim() } })
  } else {
    router.push('/search')
  }
}

function goToDetail(item: SearchResult) {
  router.push({ path: '/search', query: { q: item.game_name_cn || item.game_name_en } })
}

function formatDate(dateStr: string): string {
  if (!dateStr) return ''
  try {
    const d = new Date(dateStr)
    return `${d.getFullYear()}-${(d.getMonth() + 1).toString().padStart(2, '0')}-${d.getDate().toString().padStart(2, '0')}`
  } catch {
    return dateStr
  }
}

async function triggerSync() {
  store.isSyncing = true
  try {
    const rec = await api.triggerSync()
    ElMessage.success(t('home.syncSuccess', { added: rec?.added_count || 0, updated: rec?.updated_count || 0 }))
    await store.refreshLatest()
  } catch (e: any) {
    ElMessage.error(t('home.syncFailed') + (e.message || e))
  } finally {
    store.isSyncing = false
  }
}
</script>

<style scoped>
.home-view {
  padding-bottom: 40px;
}
.hero-banner {
  background: linear-gradient(135deg, rgba(30, 41, 59, 0.85) 0%, rgba(15, 23, 42, 0.95) 100%),
              url('https://flingtrainer.com/wp-content/uploads/2020/12/header-7-200x200.jpg') center/cover;
  border-radius: 12px;
  padding: 36px 32px;
  margin-bottom: 24px;
  border: 1px solid var(--border-color);
  box-shadow: var(--card-shadow);
  transition: all 0.25s ease;
}
:root[data-theme="light"] .hero-banner {
  background: linear-gradient(135deg, #0284c7 0%, #0369a1 100%);
}
.hero-title {
  font-size: 28px;
  font-weight: 800;
  color: #f8fafc;
  margin: 0 0 8px 0;
  letter-spacing: -0.5px;
}
.hero-subtitle {
  color: rgba(255, 255, 255, 0.85);
  font-size: 14px;
  margin: 0 0 24px 0;
}
.hero-search-box {
  max-width: 680px;
}
.stats-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 16px;
  margin-bottom: 28px;
}
.stat-card {
  background: var(--bg-card);
  border: 1px solid var(--border-color);
  border-radius: 8px;
  padding: 16px 20px;
  backdrop-filter: blur(8px);
  box-shadow: var(--card-shadow);
  transition: all 0.25s ease;
}
.stat-val {
  font-size: 24px;
  font-weight: 700;
  color: var(--accent-cyan);
  margin-bottom: 4px;
}
.stat-label {
  font-size: 12px;
  color: var(--text-muted);
}
.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}
.section-title {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 16px;
  font-weight: 600;
  color: var(--text-main);
}
.trainers-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(220px, 1fr));
  gap: 16px;
}
.trainer-card {
  background: var(--bg-card);
  border: 1px solid var(--border-color);
  border-radius: 8px;
  overflow: hidden;
  cursor: pointer;
  transition: all 0.2s ease;
  box-shadow: var(--card-shadow);
}
.trainer-card:hover {
  transform: translateY(-3px);
  border-color: var(--border-hover);
  background: var(--bg-card-hover);
}
.card-thumb {
  position: relative;
  height: 120px;
  background: var(--bg-card-subtle);
}
.card-thumb img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}
.badge-source {
  position: absolute;
  top: 8px;
  right: 8px;
  background: var(--bg-topbar);
  color: var(--accent-cyan);
  padding: 2px 6px;
  font-size: 10px;
  font-weight: 700;
  border-radius: 4px;
  backdrop-filter: blur(4px);
  border: 1px solid var(--border-subtle);
}
.card-info {
  padding: 12px;
}
.game-name-cn {
  font-size: 14px;
  font-weight: 600;
  color: var(--text-main);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  margin-bottom: 2px;
}
.game-name-en {
  font-size: 12px;
  color: var(--text-gray);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  margin-bottom: 10px;
}
.card-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.update-time {
  font-size: 11px;
  color: var(--text-muted);
}
</style>

