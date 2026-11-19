<template>
  <div class="search-view">
    <!-- Search Bar & Filters -->
    <div class="search-header">
      <div class="search-input-wrap">
        <el-input
          v-model="store.query"
          size="large"
          placeholder="中英双向搜索：黑神话：悟空 / Black Myth / 赛博朋克2077 / Cyberpunk / sbpk / 老头环..."
          clearable
          @keyup.enter="handleSearch"
        >
          <template #prefix>
            <el-icon><Search /></el-icon>
          </template>
          <template #append>
            <el-button type="primary" :loading="store.searching" @click="handleSearch">搜索</el-button>
          </template>
        </el-input>
      </div>

      <!-- 搜索历史 -->
      <div v-if="store.searchHistory.length > 0" class="history-bar">
        <span class="history-label">最近搜索:</span>
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
          <el-button link type="info" size="small" class="clear-btn" @click="store.clearSearchHistory">清空</el-button>
        </div>
      </div>

      <div class="filter-bar">
        <el-checkbox v-model="store.localOnly" @change="handleSearch">仅搜本地已收录</el-checkbox>
        <el-radio-group v-model="store.selectedSource" size="small" @change="handleSearch">
          <el-radio-button value="all">全部来源</el-radio-button>
          <el-radio-button value="fling">FLiNG 官方源</el-radio-button>
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
            <el-tag v-if="item.is_downloaded" type="success" size="small" class="dl-tag">已下载</el-tag>
            <el-tag v-else-if="item.is_online_only" type="info" size="small" class="dl-tag">在线源</el-tag>
          </div>
          <div class="result-body">
            <div class="result-name-cn">{{ item.game_name_cn || item.game_name_en }}</div>
            <div class="result-name-en">{{ item.game_name_en }}</div>
            <div class="result-meta">
              <span class="source-pill">{{ item.source_id.toUpperCase() }}</span>
              <span class="score-tag">匹配度 {{ Math.min(100, Math.round(item.score)) }}%</span>
            </div>
          </div>
        </div>
      </div>

      <div v-else-if="!store.searching && store.hasSearched" class="empty-state">
        <el-empty description="未检索到相关修改器资源">
          <template #extra>
            <p class="empty-tip">提示：支持输入中文别名（如“老头环”）、拼音首字母（如“sbpk”）或英文全称检索。</p>
          </template>
        </el-empty>
      </div>
    </div>

    <!-- Detail Drawer -->
    <el-drawer
      v-model="store.drawerVisible"
      title="修改器详情与版本列表"
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
        <div class="section-sub">可下载版本列表</div>
        <div v-if="store.detailVersions.length > 0" class="versions-list">
          <div
            v-for="v in store.detailVersions"
            :key="v.version_name"
            class="version-item"
          >
            <div class="version-name">{{ v.version_name }}</div>
            <div class="version-sub">
              <span>文件: {{ v.file_name }}</span>
            </div>
            <div class="version-action">
              <el-button
                v-if="v.is_downloaded"
                type="success"
                size="small"
                disabled
              >
                已下载 (本地就绪)
              </el-button>
              <el-button
                v-else
                type="primary"
                size="small"
                :loading="store.downloadingVerID === v.id"
                @click="store.downloadVersion(v)"
              >
                <el-icon><Download /></el-icon> 客户端下载
              </el-button>
            </div>
          </div>
        </div>
        <div v-else class="text-gray text-sm">暂无具体版本列表</div>

        <!-- Features Cheats List -->
        <div v-if="store.detailVersions.length > 0 && store.detailVersions[0]?.features_text" class="mt-4">
          <div class="section-sub">修改器快捷键与功能项</div>
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
  color: #64748b;
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
  background: rgba(30, 41, 59, 0.7);
  border-color: rgba(255, 255, 255, 0.1);
  color: #cbd5e1;
  transition: all 0.2s ease;
}
.history-tag:hover {
  border-color: #38bdf8;
  color: #38bdf8;
}
.clear-btn {
  font-size: 11px;
  color: #64748b;
}
.clear-btn:hover {
  color: #94a3b8;
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
  background: rgba(30, 41, 59, 0.6);
  border: 1px solid rgba(255, 255, 255, 0.08);
  border-radius: 8px;
  overflow: hidden;
  cursor: pointer;
  transition: all 0.2s ease;
}
.result-card:hover {
  transform: translateY(-2px);
  border-color: #38bdf8;
  box-shadow: 0 6px 16px rgba(0, 0, 0, 0.3);
}
.result-thumb {
  position: relative;
  height: 130px;
  background: #0f172a;
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
  color: #f8fafc;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  margin-bottom: 2px;
}
.result-name-en {
  font-size: 12px;
  color: #64748b;
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
  color: #38bdf8;
  background: rgba(56, 189, 248, 0.1);
  padding: 2px 6px;
  border-radius: 4px;
}
.score-tag {
  font-size: 11px;
  color: #94a3b8;
}
.empty-state {
  padding: 40px 0;
}
.empty-tip {
  font-size: 12px;
  color: #64748b;
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
  color: #f8fafc;
}
.detail-title-group p {
  margin: 0;
  font-size: 12px;
  color: #94a3b8;
}
.section-sub {
  font-size: 13px;
  font-weight: 600;
  color: #e2e8f0;
  margin-bottom: 10px;
}
.versions-list {
  display: flex;
  flex-direction: column;
  gap: 10px;
}
.version-item {
  background: rgba(255, 255, 255, 0.03);
  border: 1px solid rgba(255, 255, 255, 0.08);
  border-radius: 6px;
  padding: 10px 12px;
}
.version-name {
  font-size: 13px;
  font-weight: 600;
  color: #38bdf8;
  margin-bottom: 4px;
}
.version-sub {
  font-size: 11px;
  color: #94a3b8;
  margin-bottom: 8px;
}
.cheats-box {
  background: #090d16;
  padding: 12px;
  border-radius: 6px;
  font-family: monospace;
  font-size: 12px;
  color: #38bdf8;
  max-height: 240px;
  overflow-y: auto;
  white-space: pre-wrap;
  border: 1px solid rgba(255, 255, 255, 0.06);
}
</style>
