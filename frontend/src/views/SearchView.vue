<template>
  <div class="search-view">
    <!-- Search Bar & Filters -->
    <div class="search-header">
      <div class="search-input-wrap">
        <el-input
          v-model="store.query"
          size="large"
          :placeholder="t('search.placeholder')"
          clearable
          @keyup.enter="handleSearch"
        >
          <template #prefix>
            <el-icon><Search /></el-icon>
          </template>
          <template #append>
            <el-button type="primary" :loading="store.searching" @click="handleSearch">{{ t('search.searchBtn') }}</el-button>
          </template>
        </el-input>
      </div>

      <!-- 搜索历史 -->
      <div v-if="store.searchHistory.length > 0" class="history-bar">
        <span class="history-label">{{ t('search.recentSearch') }}</span>
        <div class="history-tags">
          <el-tag
            v-for="item in store.searchHistory"
            :key="item"
            closable
            size="small"
            class="history-tag"
            @click="clickHistory(item)"
            @close="store.removeSearchHistory(item)"
          >
            {{ item }}
          </el-tag>
          <el-button link type="info" size="small" class="clear-btn" @click="store.clearSearchHistory">{{ t('search.clear') }}</el-button>
        </div>
      </div>

      <!-- 最近访问游戏快捷标签 -->
      <div v-if="store.recentVisited.length > 0" class="history-bar visited-bar">
        <span class="history-label">{{ t('search.visitedGames') }}</span>
        <div class="history-tags">
          <el-tag
            v-for="vg in store.recentVisited"
            :key="vg.id"
            type="warning"
            effect="plain"
            size="small"
            class="visited-tag"
            @click="clickVisited(vg)"
          >
            {{ vg.name_cn || vg.name_en }}
          </el-tag>
        </div>
      </div>

      <div class="filter-bar">
        <el-checkbox v-model="store.localOnly" @change="handleSearch">{{ t('search.localOnly') }}</el-checkbox>
        <el-radio-group v-model="store.selectedSource" size="small" @change="handleSearch">
          <el-radio-button value="all">{{ t('search.allSources') }}</el-radio-button>
          <el-radio-button value="fling">{{ t('search.flingSource') }}</el-radio-button>
        </el-radio-group>
      </div>
    </div>

    <!-- Search Results -->
    <div v-loading="store.searching" class="results-container">
      <div v-if="store.results.length > 0" class="results-grid">
        <div
          v-for="item in store.results"
          :key="item.slug || item.game_id"
          class="result-card"
          @click="store.openDetail(item)"
        >
          <div class="result-thumb">
            <img :src="item.cover_url || DEFAULT_GAME_COVER" :alt="item.game_name_en" />
            <el-tag v-if="item.is_downloaded" type="success" size="small" class="dl-tag">{{ t('search.downloaded') }}</el-tag>
            <el-tag v-else-if="item.is_online_only" type="info" size="small" class="dl-tag">{{ t('home.tagOnline') }}</el-tag>
            <!-- 收藏按钮 -->
            <button
              class="star-btn"
              :class="{ 'is-active': item.is_favorite }"
              :title="item.is_favorite ? t('search.unfavorite') : t('search.favorite')"
              @click.stop="store.toggleFavorite(item)"
            >
              <el-icon :size="16">
                <StarFilled v-if="item.is_favorite" />
                <Star v-else />
              </el-icon>
            </button>
          </div>
          <div class="result-body">
            <div class="result-name-cn">{{ item.game_name_cn || item.game_name_en }}</div>
            <div class="result-name-en">{{ item.game_name_en }}</div>
            <div class="result-meta">
              <span class="source-pill">{{ (item.source_id || 'FLiNG').toUpperCase() }}</span>
              <span class="score-tag">{{ t('search.matchScore', { score: Math.min(100, Math.round(item.score)) }) }}</span>
            </div>
          </div>
        </div>
      </div>

      <div v-else-if="!store.searching && store.hasSearched" class="empty-state">
        <el-empty :description="t('search.emptyTitle')">
          <template #extra>
            <p class="empty-tip">{{ t('search.emptyTip') }}</p>
          </template>
        </el-empty>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, watch } from 'vue'
import { useRoute } from 'vue-router'
import { Search, Star, StarFilled } from '@element-plus/icons-vue'
import { useSearchStore } from '../stores/searchStore'
import { t } from '../i18n'
import { DEFAULT_GAME_COVER } from '../utils/assets'
import type { Game } from '../types'

defineOptions({ name: 'SearchView' })

const route = useRoute()
const store = useSearchStore()

onMounted(() => {
  const initialQ = route.query.q as string
  if (initialQ && initialQ !== store.query) {
    store.executeSearch(initialQ)
  }
})

watch(() => route.query.q, (newQ) => {
  if (newQ && typeof newQ === 'string' && newQ !== store.query) {
    store.executeSearch(newQ)
  }
})

function handleSearch() {
  store.executeSearch()
}

function clickHistory(kw: string) {
  store.executeSearch(kw)
}

function clickVisited(vg: Game) {
  store.executeSearch(vg.name_cn || vg.name_en)
}
</script>

<style scoped>
.search-header {
  margin-bottom: 24px;
}
.search-input-wrap {
  margin-bottom: 8px;
}
.history-bar {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 8px;
  padding: 0 4px;
  flex-wrap: wrap;
}
.visited-bar {
  margin-bottom: 12px;
}
.history-label {
  font-size: 12px;
  color: var(--text-muted);
  white-space: nowrap;
}
.history-tags {
  display: flex;
  align-items: center;
  gap: 6px;
  flex-wrap: wrap;
}
.history-tag, .visited-tag {
  cursor: pointer;
  transition: all 0.2s ease;
}
.history-tag {
  background: var(--bg-card);
  border-color: var(--border-color);
  color: var(--text-secondary);
}
.history-tag:hover {
  border-color: var(--accent-cyan);
  color: var(--accent-cyan);
}
.visited-tag:hover {
  transform: translateY(-1px);
}
.clear-btn {
  font-size: 11px;
  color: var(--text-gray);
}
.clear-btn:hover {
  color: var(--text-muted);
}
.filter-bar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0 4px;
}
.results-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(260px, 1fr));
  gap: 16px;
}
.result-card {
  background: var(--bg-card);
  border: 1px solid var(--border-color);
  border-radius: 8px;
  overflow: hidden;
  cursor: pointer;
  transition: all 0.2s ease;
  box-shadow: var(--card-shadow);
  position: relative;
}
.result-card:hover {
  transform: translateY(-2px);
  border-color: var(--border-hover);
  background: var(--bg-card-hover);
}
.result-thumb {
  position: relative;
  height: 130px;
  background: var(--bg-card-subtle);
}
.result-thumb img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}
.dl-tag {
  position: absolute;
  top: 8px;
  right: 8px;
}
.star-btn {
  position: absolute;
  top: 8px;
  left: 8px;
  background: rgba(0, 0, 0, 0.55);
  border: 1px solid rgba(255, 255, 255, 0.2);
  border-radius: 50%;
  width: 28px;
  height: 28px;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  color: #fff;
  transition: all 0.2s ease;
  padding: 0;
}
.star-btn:hover {
  background: rgba(0, 0, 0, 0.75);
  transform: scale(1.1);
  color: #fbbf24;
}
.star-btn.is-active {
  color: #fbbf24;
  background: rgba(0, 0, 0, 0.7);
  border-color: #fbbf24;
}
.result-body {
  padding: 12px;
}
.result-name-cn {
  font-size: 15px;
  font-weight: 600;
  color: var(--text-main);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  margin-bottom: 2px;
}
.result-name-en {
  font-size: 12px;
  color: var(--text-gray);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  margin-bottom: 10px;
}
.result-meta {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.source-pill {
  font-size: 10px;
  color: var(--accent-cyan);
  background: var(--accent-tag-bg);
  padding: 2px 6px;
  border-radius: 4px;
}
.score-tag {
  font-size: 11px;
  color: var(--text-muted);
}
.empty-state {
  padding: 40px 0;
}
.empty-tip {
  font-size: 12px;
  color: var(--text-muted);
  margin-top: 8px;
}
.detail-header {
  display: flex;
  gap: 16px;
  align-items: center;
}
.detail-cover {
  width: 90px;
  height: 90px;
  border-radius: 6px;
  object-fit: cover;
}
.detail-title-group {
  flex: 1;
  min-width: 0;
}
.drawer-title-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  margin-bottom: 4px;
}
.drawer-title-row h3 {
  margin: 0;
  font-size: 16px;
  color: var(--text-main);
}
.detail-title-group p {
  margin: 0;
  font-size: 12px;
  color: var(--text-muted);
}
.section-sub {
  font-size: 13px;
  font-weight: 600;
  color: var(--text-primary);
  margin-bottom: 10px;
}
.versions-list {
  display: flex;
  flex-direction: column;
  gap: 10px;
}
.version-item {
  background: var(--bg-card);
  border: 1px solid var(--border-color);
  border-radius: 6px;
  padding: 10px 12px;
}
.version-name {
  font-size: 13px;
  font-weight: 600;
  color: var(--accent-cyan);
  margin-bottom: 4px;
}
.version-sub {
  font-size: 11px;
  color: var(--text-muted);
  margin-bottom: 8px;
}
.cheats-box {
  background: var(--code-bg);
  padding: 12px;
  border-radius: 6px;
  font-family: monospace;
  font-size: 12px;
  color: var(--code-text);
  max-height: 240px;
  overflow-y: auto;
  white-space: pre-wrap;
  border: 1px solid var(--border-color);
}
</style>
