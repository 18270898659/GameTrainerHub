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
          :key="item.slug"
          class="result-card"
          @click="store.openDetail(item)"
        >
          <div class="result-thumb">
            <img :src="item.cover_url || defaultThumb" :alt="item.game_name_en" />
            <el-tag v-if="item.is_downloaded" type="success" size="small" class="dl-tag">{{ t('search.downloaded') }}</el-tag>
            <el-tag v-else-if="item.is_online_only" type="info" size="small" class="dl-tag">{{ t('home.tagOnline') }}</el-tag>
          </div>
          <div class="result-body">
            <div class="result-name-cn">{{ item.game_name_cn || item.game_name_en }}</div>
            <div class="result-name-en">{{ item.game_name_en }}</div>
            <div class="result-meta">
              <span class="source-pill">{{ item.source_id.toUpperCase() }}</span>
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

    <!-- Detail Drawer -->
    <el-drawer
      v-model="store.drawerVisible"
      :title="t('search.drawerTitle')"
      size="480px"
      class="detail-drawer"
    >
      <div v-if="store.selectedGame" v-loading="store.loadingDetail" class="detail-drawer-content">
        <div class="detail-header">
          <img :src="store.selectedGame.cover_url || defaultThumb" class="detail-cover" />
          <div class="detail-title-group">
            <h3>{{ store.selectedGame.game_name_cn || store.selectedGame.game_name_en }}</h3>
            <p>{{ store.selectedGame.game_name_en }}</p>
          </div>
        </div>

        <el-divider />

        <!-- Versions List -->
        <div class="section-sub">{{ t('search.versionList') }}</div>
        <div v-if="store.detailVersions.length > 0" class="versions-list">
          <div
            v-for="v in store.detailVersions"
            :key="v.version_name"
            class="version-item"
          >
            <div class="version-name">{{ v.version_name }}</div>
            <div class="version-sub">
              <span>{{ v.file_name }}</span>
            </div>
            <div class="version-action">
              <el-button
                v-if="v.is_downloaded"
                type="success"
                size="small"
                disabled
              >
                {{ t('search.downloaded') }}
              </el-button>
              <el-button
                v-else
                type="primary"
                size="small"
                :loading="store.downloadingVerID === v.id"
                @click="store.downloadVersion(v)"
              >
                <el-icon><Download /></el-icon> {{ t('search.downloadBtn') }}
              </el-button>
            </div>
          </div>
        </div>
        <div v-else class="text-gray text-sm">{{ t('search.noFeatures') }}</div>

        <!-- Features Cheats List -->
        <div v-if="store.detailVersions.length > 0 && store.detailVersions[0]?.features_text" class="mt-4">
          <div class="section-sub">{{ t('search.featuresList') }}</div>
          <pre class="cheats-box">{{ store.detailVersions[0]?.features_text }}</pre>
        </div>
      </div>
    </el-drawer>
  </div>
</template>

<script setup lang="ts">
import { onMounted, watch } from 'vue'
import { useRoute } from 'vue-router'
import { Search, Download } from '@element-plus/icons-vue'
import { useSearchStore } from '../stores/searchStore'
import { t } from '../i18n'

defineOptions({ name: 'SearchView' })

const route = useRoute()
const store = useSearchStore()

const defaultThumb = 'https://flingtrainer.com/wp-content/uploads/2020/12/header-7-200x200.jpg'

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
  margin-bottom: 12px;
  padding: 0 4px;
  flex-wrap: wrap;
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
.history-tag {
  cursor: pointer;
  background: var(--bg-card);
  border-color: var(--border-color);
  color: var(--text-secondary);
  transition: all 0.2s ease;
}
.history-tag:hover {
  border-color: var(--accent-cyan);
  color: var(--accent-cyan);
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
.detail-title-group h3 {
  margin: 0 0 4px 0;
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
