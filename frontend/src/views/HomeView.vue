<template>
  <div class="home-view">
    <!-- Hero Banner -->
    <div class="hero-banner">
      <div class="hero-content">
        <h1 class="hero-title">Game Trainer Hub</h1>
        <p class="hero-subtitle">PC 游戏修改器聚合管理客户端 · 官方资源直连索引 · 安全受控运行</p>
        
        <div class="hero-search-box">
          <el-input
            v-model="heroQuery"
            size="large"
            placeholder="输入游戏中文名、英文名、拼音缩写 (例如: 赛博朋克2077 / sbpk / 老头环)..."
            clearable
            @keyup.enter="handleHeroSearch"
          >
            <template #prefix>
              <el-icon><Search /></el-icon>
            </template>
            <template #append>
              <el-button type="primary" @click="handleHeroSearch">立即搜索</el-button>
            </template>
          </el-input>
        </div>
      </div>
    </div>

    <!-- Quick Stats -->
    <div class="stats-grid">
      <div class="stat-card">
        <div class="stat-val">767+</div>
        <div class="stat-label">收录修改器游戏</div>
      </div>
      <div class="stat-card">
        <div class="stat-val">{{ store.localTrainers.length }}</div>
        <div class="stat-label">本地已就绪修改器</div>
      </div>
      <div class="stat-card">
        <div class="stat-val">{{ store.activeDownloadsCount }}</div>
        <div class="stat-label">当前下载任务</div>
      </div>
      <div class="stat-card">
        <div class="stat-val">FLiNG</div>
        <div class="stat-label">首期官方同步源</div>
      </div>
    </div>

    <!-- Latest Updated Grid -->
    <div class="section-header">
      <div class="section-title">
        <el-icon><Clock /></el-icon>
        <span>原网站最近更新与发布</span>
      </div>
      <el-button size="small" :loading="store.isSyncing" @click="triggerSync" plain>
        <el-icon><Refresh /></el-icon> 同步最新资源
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
            <el-tag v-if="item.is_downloaded" size="small" type="success">已下载</el-tag>
            <el-tag v-else size="small" type="info">在线获取</el-tag>
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
    await api.triggerSync()
    ElMessage.success('已从原网站同步最新修改器')
    await store.refreshLatest()
  } catch (e: any) {
    ElMessage.error('同步失败: ' + (e.message || e))
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
  background: linear-gradient(135deg, rgba(30, 41, 59, 0.8) 0%, rgba(15, 23, 42, 0.95) 100%),
              url('https://flingtrainer.com/wp-content/uploads/2020/12/header-7-200x200.jpg') center/cover;
  border-radius: 12px;
  padding: 36px 32px;
  margin-bottom: 24px;
  border: 1px solid rgba(255, 255, 255, 0.1);
  box-shadow: 0 10px 25px -5px rgba(0, 0, 0, 0.3);
}
.hero-title {
  font-size: 28px;
  font-weight: 800;
  color: #f8fafc;
  margin: 0 0 8px 0;
  letter-spacing: -0.5px;
}
.hero-subtitle {
  color: #94a3b8;
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
  background: rgba(30, 41, 59, 0.5);
  border: 1px solid rgba(255, 255, 255, 0.08);
  border-radius: 8px;
  padding: 16px 20px;
  backdrop-filter: blur(8px);
}
.stat-val {
  font-size: 24px;
  font-weight: 700;
  color: #38bdf8;
  margin-bottom: 4px;
}
.stat-label {
  font-size: 12px;
  color: #94a3b8;
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
  color: #f1f5f9;
}
.trainers-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(220px, 1fr));
  gap: 16px;
}
.trainer-card {
  background: rgba(30, 41, 59, 0.6);
  border: 1px solid rgba(255, 255, 255, 0.06);
  border-radius: 8px;
  overflow: hidden;
  cursor: pointer;
  transition: all 0.2s ease;
}
.trainer-card:hover {
  transform: translateY(-3px);
  border-color: rgba(56, 189, 248, 0.4);
  box-shadow: 0 8px 20px rgba(0, 0, 0, 0.4);
}
.card-thumb {
  position: relative;
  height: 120px;
  background: #0f172a;
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
  background: rgba(15, 23, 42, 0.85);
  color: #38bdf8;
  padding: 2px 6px;
  font-size: 10px;
  font-weight: 700;
  border-radius: 4px;
  backdrop-filter: blur(4px);
}
.card-info {
  padding: 12px;
}
.game-name-cn {
  font-size: 14px;
  font-weight: 600;
  color: #f8fafc;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  margin-bottom: 2px;
}
.game-name-en {
  font-size: 12px;
  color: #64748b;
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
  color: #94a3b8;
}
</style>

