<template>
  <div class="home-view">
    <!-- Hero Banner -->
    <div class="hero-banner">
      <div class="hero-content">
        <h1 class="hero-title">{{ t('home.heroTitle') }}</h1>
        <p class="hero-subtitle">{{ t('home.heroSubtitle') }}</p>
        
        <div class="hero-search-box" ref="searchBoxRef">
          <el-input
            ref="searchInputRef"
            v-model="heroQuery"
            size="large"
            :placeholder="t('home.searchPlaceholder')"
            clearable
            @input="onQueryInput"
            @focus="onInputFocus"
            @keydown.down.prevent="onKeyDown"
            @keydown.up.prevent="onKeyUp"
            @keydown.enter.prevent="onKeyEnter"
            @keydown.esc="closeSuggestions"
          >
            <template #prefix>
              <el-icon><Search /></el-icon>
            </template>
            <template #append>
              <el-button type="primary" @click="handleHeroSearch">{{ t('home.searchBtn') }}</el-button>
            </template>
          </el-input>

          <!-- 渐进式搜索联想与自动补全浮层 -->
          <transition name="el-zoom-in-top">
            <div v-show="showSuggestions" class="search-suggestions-dropdown">
              <!-- 加载中状态 -->
              <div v-if="loadingSuggestions" class="sugg-loading">
                <el-icon class="is-loading"><Loading /></el-icon>
                <span>{{ t('home.suggSearching') }}</span>
              </div>

              <!-- 当有搜索词时：显示联想结果与完整检索入口 -->
              <template v-else-if="heroQuery.trim()">
                <div class="sugg-header">
                  <span class="sugg-title-label">{{ t('home.suggHeader') }}</span>
                  <span class="sugg-hint-label">Enter: {{ t('home.searchBtn') }} / ESC: 关闭</span>
                </div>

                <div v-if="suggestions.length > 0" class="sugg-list" role="listbox">
                  <div
                    v-for="(item, idx) in suggestions"
                    :key="item.slug || item.game_id || idx"
                    class="sugg-item"
                    :class="{ 'is-active': activeSuggestIndex === idx }"
                    @mouseenter="activeSuggestIndex = idx"
                    @click="selectSuggestion(item)"
                  >
                    <img :src="item.cover_url || DEFAULT_GAME_COVER" class="sugg-thumb" />
                    <div class="sugg-info">
                      <div class="sugg-name-row">
                        <span class="sugg-title">{{ item.game_name_cn || item.game_name_en }}</span>
                        <span v-if="item.game_name_cn && item.game_name_en && item.game_name_cn !== item.game_name_en" class="sugg-sub-name">
                          ({{ item.game_name_en }})
                        </span>
                      </div>
                      <div class="sugg-tags-row">
                        <el-tag v-if="item.is_downloaded" size="small" type="success" effect="dark">{{ t('home.tagDownloaded') }}</el-tag>
                        <el-tag v-else size="small" type="info">{{ t('home.tagOnline') }}</el-tag>
                        <span v-if="item.is_favorite" class="sugg-fav-icon" :title="t('search.favorite')">
                          <el-icon><StarFilled /></el-icon>
                        </span>
                        <span class="sugg-source-badge">{{ (item.source_id || 'FLiNG').toUpperCase() }}</span>
                      </div>
                    </div>
                    <div class="sugg-action">
                      <span class="sugg-action-text">{{ t('home.suggViewDetail') }}</span>
                      <el-icon><Right /></el-icon>
                    </div>
                  </div>
                </div>

                <div v-else class="sugg-empty">
                  <el-icon><InfoFilled /></el-icon>
                  <span>{{ t('home.suggEmpty') }}</span>
                </div>

                <!-- 查看全部搜索结果入口 -->
                <div
                  class="sugg-footer-search"
                  :class="{ 'is-active': activeSuggestIndex === suggestions.length }"
                  @mouseenter="activeSuggestIndex = suggestions.length"
                  @click="handleHeroSearch"
                >
                  <div class="sugg-footer-content">
                    <el-icon><Search /></el-icon>
                    <span>{{ t('home.suggViewAll', { query: heroQuery.trim() }) }}</span>
                  </div>
                  <span class="sugg-kbd">↵ Enter</span>
                </div>
              </template>

              <!-- 当无搜索词聚焦时：展示最近搜索历史与访问记录 -->
              <template v-else>
                <!-- 搜索历史 -->
                <div v-if="searchStore.searchHistory.length > 0" class="sugg-section">
                  <div class="sugg-header">
                    <span class="sugg-title-label">{{ t('home.suggRecentSearch') }}</span>
                    <el-button link size="small" type="danger" @click.stop="searchStore.clearSearchHistory()">清空</el-button>
                  </div>
                  <div class="sugg-history-tags">
                    <el-tag
                      v-for="kw in searchStore.searchHistory.slice(0, 10)"
                      :key="kw"
                      class="sugg-tag"
                      closable
                      @click="quickSearch(kw)"
                      @close="searchStore.removeSearchHistory(kw)"
                    >
                      {{ kw }}
                    </el-tag>
                  </div>
                </div>

                <!-- 最近访问游戏 -->
                <div v-if="recentGames.length > 0" class="sugg-section">
                  <div class="sugg-header">
                    <span class="sugg-title-label">{{ t('home.suggRecentVisited') }}</span>
                  </div>
                  <div class="sugg-quick-grid">
                    <div
                      v-for="g in recentGames.slice(0, 4)"
                      :key="g.id"
                      class="sugg-quick-card"
                      @click="selectGameQuick(g)"
                    >
                      <img :src="g.cover_url || DEFAULT_GAME_COVER" class="sugg-quick-thumb" />
                      <div class="sugg-quick-title">{{ g.name_cn || g.name_en }}</div>
                    </div>
                  </div>
                </div>

                <div v-if="searchStore.searchHistory.length === 0 && recentGames.length === 0" class="sugg-empty">
                  <el-icon><Search /></el-icon>
                  <span>{{ t('home.searchPlaceholder') }}</span>
                </div>
              </template>
            </div>
          </transition>
        </div>
      </div>
    </div>

    <!-- Quick Stats -->
    <div class="stats-grid">
      <div v-if="statsLoading && !indexedGamesCount" v-for="i in 4" :key="i" class="stat-card">
        <el-skeleton animated>
          <template #template>
            <el-skeleton-item variant="h1" style="width: 50%; height: 28px; margin-bottom: 8px;" />
            <el-skeleton-item variant="text" style="width: 70%; height: 14px;" />
          </template>
        </el-skeleton>
      </div>
      <template v-else>
        <div class="stat-card">
          <div class="stat-val">{{ indexedGamesCount }}</div>
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
      </template>
    </div>

    <!-- My Favorites Section (if any) -->
    <div v-if="store.favorites.length > 0" class="home-section">
      <div class="section-header">
        <div class="section-title">
          <el-icon class="text-amber"><StarFilled /></el-icon>
          <span>{{ t('home.favoritesTitle') }}</span>
          <span class="count-tag">({{ store.favorites.length }})</span>
        </div>
      </div>
      <div class="favorites-grid">
        <div
          v-for="fav in store.favorites"
          :key="fav.id"
          class="favorite-card"
          @click="goToGame(fav)"
        >
          <img :src="fav.cover_url || DEFAULT_GAME_COVER" class="favorite-cover" />
          <div class="favorite-info">
            <div class="fav-name-cn">{{ fav.name_cn || fav.name_en }}</div>
            <div class="fav-name-en">{{ fav.name_en }}</div>
          </div>
        </div>
      </div>
    </div>

    <!-- Recently Visited Games & Recently Used Trainers row -->
    <div class="dual-columns-grid">
      <!-- Recently Visited Games -->
      <div class="column-card">
        <div class="column-header">
          <div class="column-title">
            <el-icon><Compass /></el-icon>
            <span>{{ t('home.recentVisited') }}</span>
          </div>
        </div>
        <div v-if="recentGamesLoading && recentGames.length === 0" class="item-list-skeleton">
          <div v-for="n in 3" :key="n" class="skeleton-row">
            <el-skeleton animated>
              <template #template>
                <div style="display: flex; align-items: center; gap: 12px; padding: 4px 0;">
                  <el-skeleton-item variant="image" style="width: 36px; height: 36px; border-radius: 4px;" />
                  <div style="flex: 1;">
                    <el-skeleton-item variant="text" style="width: 60%; height: 14px; margin-bottom: 4px;" />
                    <el-skeleton-item variant="text" style="width: 40%; height: 12px;" />
                  </div>
                </div>
              </template>
            </el-skeleton>
          </div>
        </div>
        <div v-else-if="recentGames.length > 0" class="item-list">
          <div
            v-for="g in recentGames"
            :key="g.id"
            class="list-row"
            @click="goToGame(g)"
          >
            <img :src="g.cover_url || DEFAULT_GAME_COVER" class="row-thumb" />
            <div class="row-info">
              <div class="row-title">{{ g.name_cn || g.name_en }}</div>
              <div class="row-sub">{{ g.name_en }}</div>
            </div>
            <div class="row-time">{{ formatTime(g.updated_at) }}</div>
          </div>
        </div>
        <div v-else class="column-empty">{{ t('home.noRecentGames') }}</div>
      </div>

      <!-- Recently Used Trainers -->
      <div class="column-card">
        <div class="column-header">
          <div class="column-title">
            <el-icon><Cpu /></el-icon>
            <span>{{ t('home.recentTrainers') }}</span>
          </div>
        </div>
        <div v-if="recentTrainersLoading && recentTrainers.length === 0" class="item-list-skeleton">
          <div v-for="n in 3" :key="n" class="skeleton-row">
            <el-skeleton animated>
              <template #template>
                <div style="display: flex; align-items: center; gap: 12px; padding: 4px 0;">
                  <el-skeleton-item variant="circle" style="width: 36px; height: 36px;" />
                  <div style="flex: 1;">
                    <el-skeleton-item variant="text" style="width: 60%; height: 14px; margin-bottom: 4px;" />
                    <el-skeleton-item variant="text" style="width: 40%; height: 12px;" />
                  </div>
                </div>
              </template>
            </el-skeleton>
          </div>
        </div>
        <div v-else-if="recentTrainers.length > 0" class="item-list">
          <div
            v-for="tr in recentTrainers"
            :key="tr.local_trainer.id"
            class="list-row"
            @click="goToLibrary"
          >
            <div class="trainer-avatar">⚡</div>
            <div class="row-info">
              <div class="row-title">{{ tr.game?.name_cn || tr.game?.name_en || tr.local_trainer.file_path.split('\\').pop() }}</div>
              <div class="row-sub">
                <span v-if="tr.version?.version_name">{{ tr.version.version_name }} · </span>
                <span>运行 {{ tr.local_trainer.run_count }} 次</span>
              </div>
            </div>
            <div class="row-time">{{ formatTime(tr.local_trainer.last_run_at || tr.local_trainer.created_at) }}</div>
          </div>
        </div>
        <div v-else class="column-empty">{{ t('home.noRecentTrainers') }}</div>
      </div>
    </div>

    <!-- Latest Updated Grid -->
    <div class="home-section mt-6">
      <div class="section-header">
        <div class="section-title">
          <el-icon><Clock /></el-icon>
          <span>{{ t('home.latestTitle') }}</span>
        </div>
        <div class="section-actions">
          <el-button size="small" :loading="store.isSyncing" @click="triggerSync" plain>
            <el-icon><Refresh /></el-icon> {{ t('home.syncBtn') }}
          </el-button>
        </div>
      </div>

      <!-- 状态横幅提示（网络波动重试、离线缓存模式、请求异常兜底） -->
      <transition name="el-fade-in">
        <div v-if="store.latestStatus === 'retrying'" class="status-banner retrying">
          <el-icon class="is-loading"><Loading /></el-icon>
          <span>{{ t('home.retryingNotice', { count: store.latestRetryCount }) }}</span>
        </div>
        <div v-else-if="isOffline" class="status-banner offline">
          <el-icon><WarningFilled /></el-icon>
          <span>{{ t('home.offlineNotice') }}</span>
          <el-tag size="small" type="info" class="banner-tag">{{ t('home.cachedTag') }}</el-tag>
        </div>
        <div v-else-if="store.isLatestStale && store.latestError" class="status-banner stale">
          <el-icon><InfoFilled /></el-icon>
          <span>{{ t('home.staleNotice') }}</span>
          <el-button link type="primary" size="small" :loading="manualReloading" @click="handleManualReload">
            <el-icon><Refresh /></el-icon> {{ t('home.retryBtn') }}
          </el-button>
        </div>
      </transition>

      <!-- 1. 加载中骨架屏（无缓存首次加载或重置加载） -->
      <div v-if="store.latestStatus === 'loading' && store.latestTrainers.length === 0" class="trainers-grid">
        <div v-for="n in 8" :key="n" class="trainer-card-skeleton">
          <el-skeleton animated>
            <template #template>
              <el-skeleton-item variant="image" style="width: 100%; height: 120px;" />
              <div style="padding: 12px;">
                <el-skeleton-item variant="text" style="width: 70%; height: 16px; margin-bottom: 8px;" />
                <el-skeleton-item variant="text" style="width: 50%; height: 12px; margin-bottom: 12px;" />
                <div style="display: flex; justify-content: space-between; align-items: center;">
                  <el-skeleton-item variant="text" style="width: 30%; height: 12px;" />
                  <el-skeleton-item variant="rect" style="width: 48px; height: 20px; border-radius: 4px;" />
                </div>
              </div>
            </template>
          </el-skeleton>
        </div>
      </div>

      <!-- 2. 请求彻底失败且无任何本地缓存时的错误状态卡片 -->
      <div v-else-if="store.latestStatus === 'error' && store.latestTrainers.length === 0" class="latest-error-card">
        <el-result
          icon="error"
          :title="t('home.fetchError', { error: store.latestError || 'Network Error' })"
          sub-title="未检测到可用本地缓存。请检查网络配置或点击下方按钮重新加载。"
        >
          <template #extra>
            <el-button type="primary" :loading="manualReloading" @click="handleManualReload">
              <el-icon><Refresh /></el-icon> {{ t('home.retryBtn') }}
            </el-button>
          </template>
        </el-result>
      </div>

      <!-- 3. 数据加载成功或有本地缓存兜底时的修改器网格 -->
      <div v-else-if="store.latestTrainers.length > 0" class="trainers-grid">
        <div
          v-for="item in store.latestTrainers"
          :key="item.slug"
          class="trainer-card"
          @click="goToDetail(item)"
        >
          <div class="card-thumb">
            <img :src="item.cover_url || DEFAULT_GAME_COVER" :alt="item.game_name_en" />
            <div class="badge-source">{{ item.source_id.toUpperCase() }}</div>
            <!-- 收藏切换按钮（右上角） -->
            <button
              class="star-btn"
              :class="{ 'is-fav': item.is_favorite }"
              :title="item.is_favorite ? t('search.unfavorite') : t('search.favorite')"
              @click.stop="toggleLatestFavorite(item)"
            >
              <el-icon>
                <StarFilled v-if="item.is_favorite" />
                <Star v-else />
              </el-icon>
            </button>
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

      <!-- 4. 真实空数据状态（请求成功但数据为空） -->
      <div v-else-if="store.latestStatus === 'empty' || store.latestTrainers.length === 0" class="latest-empty-card">
        <el-empty :description="t('home.emptyLatest')">
          <el-button type="primary" size="small" :loading="store.isSyncing" @click="triggerSync">
            <el-icon><Refresh /></el-icon> {{ t('home.syncBtn') }}
          </el-button>
        </el-empty>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, onActivated } from 'vue'
import { useRouter } from 'vue-router'
import {
  Search,
  Clock,
  Refresh,
  Star,
  StarFilled,
  Compass,
  Cpu,
  Loading,
  Right,
  InfoFilled,
  WarningFilled
} from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import { useAppStore } from '../stores/appStore'
import { useSearchStore } from '../stores/searchStore'
import { api } from '../api/wails'
import { t } from '../i18n'
import { DEFAULT_GAME_COVER } from '../utils/assets'
import type { SearchResult, Game, LocalTrainerDetail } from '../types'

defineOptions({
  name: 'HomeView'
})

const router = useRouter()
const store = useAppStore()
const searchStore = useSearchStore()

const searchBoxRef = ref<HTMLElement | null>(null)
const searchInputRef = ref<any>(null)
const heroQuery = ref('')
const indexedGamesCount = ref('767+')
const recentGames = ref<Game[]>([])
const recentTrainers = ref<LocalTrainerDetail[]>([])

// 独立模块加载状态，隔离异常
const statsLoading = ref(false)
const favoritesLoading = ref(false)
const recentGamesLoading = ref(false)
const recentTrainersLoading = ref(false)
const manualReloading = ref(false)
const isOffline = ref(typeof navigator !== 'undefined' ? !navigator.onLine : false)

// 搜索建议联想状态
const showSuggestions = ref(false)
const loadingSuggestions = ref(false)
const suggestions = ref<SearchResult[]>([])
const activeSuggestIndex = ref(-1)
let debounceTimer: any = null
let currentSeq = 0

// 1. 独立统计数据加载
async function loadStats() {
  statsLoading.value = true
  try {
    const count = await api.getGameCount()
    if (count && count > 0) {
      indexedGamesCount.value = `${count}+`
    }
  } catch (e) {
    console.warn('Game count fetch failed', e)
  } finally {
    statsLoading.value = false
  }
}

// 2. 独立收藏夹数据加载
async function loadFavorites() {
  favoritesLoading.value = true
  try {
    await store.refreshFavorites()
  } catch (e) {
    console.warn('Favorites fetch failed', e)
  } finally {
    favoritesLoading.value = false
  }
}

// 3. 独立最近访问游戏加载
async function loadRecentGames() {
  recentGamesLoading.value = true
  try {
    const visited = await api.getRecentGames(6)
    recentGames.value = visited || []
  } catch (e) {
    console.warn('Recent games fetch failed', e)
  } finally {
    recentGamesLoading.value = false
  }
}

// 4. 独立最近使用修改器加载
async function loadRecentTrainers() {
  recentTrainersLoading.value = true
  try {
    const rec = await api.getRecentTrainers(6)
    recentTrainers.value = rec || []
  } catch (e) {
    console.warn('Recent trainers fetch failed', e)
  } finally {
    recentTrainersLoading.value = false
  }
}

// 综合手动刷新入口
async function handleManualReload() {
  manualReloading.value = true
  try {
    await Promise.allSettled([
      store.refreshLatest(true),
      loadStats(),
      loadFavorites(),
      loadRecentGames(),
      loadRecentTrainers()
    ])
  } finally {
    manualReloading.value = false
  }
}

// 网络状态监听回调
function handleOnline() {
  isOffline.value = false
  ElMessage.success(t('home.networkRestored'))
  store.refreshLatest(true)
  loadStats()
  loadRecentGames()
  loadRecentTrainers()
}

function handleOffline() {
  isOffline.value = true
  ElMessage.warning(t('home.offlineNotice'))
}

function onInputFocus() {
  showSuggestions.value = true
  if (heroQuery.value.trim() && suggestions.value.length === 0) {
    fetchSuggestions(heroQuery.value.trim())
  }
}

function onQueryInput() {
  showSuggestions.value = true
  activeSuggestIndex.value = -1
  if (debounceTimer) clearTimeout(debounceTimer)
  const q = heroQuery.value.trim()
  if (!q) {
    suggestions.value = []
    loadingSuggestions.value = false
    return
  }
  loadingSuggestions.value = true
  debounceTimer = setTimeout(() => {
    fetchSuggestions(q)
  }, 250)
}

async function fetchSuggestions(q: string) {
  const seq = ++currentSeq
  try {
    const res = await api.getSearchSuggestions(q)
    if (seq === currentSeq) {
      suggestions.value = (res || []).slice(0, 6)
      loadingSuggestions.value = false
    }
  } catch (e) {
    if (seq === currentSeq) {
      console.warn('Failed to get suggestions', e)
      suggestions.value = []
      loadingSuggestions.value = false
    }
  }
}

function closeSuggestions() {
  showSuggestions.value = false
  activeSuggestIndex.value = -1
}

function onKeyDown() {
  if (!showSuggestions.value) {
    showSuggestions.value = true
    return
  }
  const max = heroQuery.value.trim() ? suggestions.value.length : -1
  if (max >= 0) {
    activeSuggestIndex.value = (activeSuggestIndex.value + 1) % (max + 1)
  }
}

function onKeyUp() {
  if (!showSuggestions.value) return
  const max = heroQuery.value.trim() ? suggestions.value.length : -1
  if (max >= 0) {
    activeSuggestIndex.value = (activeSuggestIndex.value - 1 + max + 1) % (max + 1)
  }
}

function onKeyEnter() {
  if (!showSuggestions.value) {
    handleHeroSearch()
    return
  }
  if (heroQuery.value.trim() && activeSuggestIndex.value >= 0 && activeSuggestIndex.value < suggestions.value.length) {
    selectSuggestion(suggestions.value[activeSuggestIndex.value])
    return
  }
  handleHeroSearch()
}

function selectSuggestion(item: SearchResult) {
  closeSuggestions()
  searchStore.openDetail(item)
}

function selectGameQuick(game: Game) {
  closeSuggestions()
  searchStore.openDetailByGame(game)
}

function quickSearch(kw: string) {
  heroQuery.value = kw
  closeSuggestions()
  router.push({ path: '/search', query: { q: kw } })
}

function handleHeroSearch() {
  closeSuggestions()
  const q = heroQuery.value.trim()
  if (q) {
    router.push({ path: '/search', query: { q } })
  } else {
    router.push('/search')
  }
}

function handleClickOutside(e: MouseEvent) {
  if (searchBoxRef.value && !searchBoxRef.value.contains(e.target as Node)) {
    closeSuggestions()
  }
}

onMounted(() => {
  document.addEventListener('click', handleClickOutside)
  if (typeof window !== 'undefined') {
    window.addEventListener('online', handleOnline)
    window.addEventListener('offline', handleOffline)
  }
  searchStore.loadSearchHistory()
  // 各模块并发独立加载，异常隔离
  Promise.allSettled([
    loadStats(),
    loadFavorites(),
    loadRecentGames(),
    loadRecentTrainers(),
    store.refreshLocalTrainers(),
    store.refreshDownloads(),
    store.refreshLatest(false)
  ])
})

onUnmounted(() => {
  document.removeEventListener('click', handleClickOutside)
  if (typeof window !== 'undefined') {
    window.removeEventListener('online', handleOnline)
    window.removeEventListener('offline', handleOffline)
  }
  if (debounceTimer) clearTimeout(debounceTimer)
})

onActivated(() => {
  searchStore.loadSearchHistory()
  // 切换回首页时静默刷新轻量数据，不阻断界面操作
  loadStats()
  loadFavorites()
  loadRecentGames()
  loadRecentTrainers()
  // 若数据过旧或已失效则后台静默刷新
  if (store.isLatestStale || !store.latestLastUpdated || Date.now() - store.latestLastUpdated > 300000) {
    store.refreshLatest(false)
  }
})

async function toggleLatestFavorite(item: SearchResult) {
  await searchStore.toggleFavorite(item)
  await loadFavorites()
}

function goToDetail(item: SearchResult) {
  searchStore.openDetail(item)
}

function goToGame(game: Game) {
  if (game) {
    searchStore.openDetailByGame(game)
  }
}

function goToLibrary() {
  router.push('/library')
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

function formatTime(dateStr: string): string {
  if (!dateStr) return ''
  try {
    const d = new Date(dateStr)
    const month = (d.getMonth() + 1).toString().padStart(2, '0')
    const day = d.getDate().toString().padStart(2, '0')
    const hour = d.getHours().toString().padStart(2, '0')
    const min = d.getMinutes().toString().padStart(2, '0')
    return `${month}-${day} ${hour}:${min}`
  } catch {
    return dateStr
  }
}

async function triggerSync() {
  store.isSyncing = true
  try {
    const rec = await api.triggerSync()
    ElMessage.success(t('home.syncSuccess', { added: rec?.added_count || 0, updated: rec?.updated_count || 0 }))
    await Promise.allSettled([
      store.refreshLatest(true),
      loadStats(),
      loadRecentGames()
    ])
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
  background: linear-gradient(135deg, rgba(30, 41, 59, 0.95) 0%, rgba(15, 23, 42, 0.98) 100%);
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
  position: relative;
  max-width: 680px;
}
.search-suggestions-dropdown {
  position: absolute;
  top: calc(100% + 8px);
  left: 0;
  right: 0;
  z-index: 100;
  background: var(--bg-card);
  border: 1px solid var(--border-color);
  border-radius: 10px;
  box-shadow: 0 14px 35px rgba(0, 0, 0, 0.45), 0 4px 12px rgba(0, 0, 0, 0.25);
  backdrop-filter: blur(12px);
  max-height: 440px;
  overflow-y: auto;
  padding: 8px 0;
}
.sugg-loading {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 10px;
  padding: 24px;
  color: var(--accent-cyan);
  font-size: 13px;
}
.sugg-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 6px 16px 8px;
  border-bottom: 1px solid var(--border-color);
  margin-bottom: 4px;
}
.sugg-title-label {
  font-size: 12px;
  font-weight: 700;
  color: var(--text-muted);
  text-transform: uppercase;
  letter-spacing: 0.5px;
}
.sugg-hint-label {
  font-size: 11px;
  color: var(--text-gray);
}
.sugg-list {
  display: flex;
  flex-direction: column;
}
.sugg-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 8px 16px;
  cursor: pointer;
  transition: all 0.15s ease;
  border-left: 3px solid transparent;
}
.sugg-item:hover,
.sugg-item.is-active {
  background: var(--bg-card-hover);
  border-left-color: var(--accent-cyan);
}
.sugg-thumb {
  width: 44px;
  height: 44px;
  border-radius: 6px;
  object-fit: cover;
  flex-shrink: 0;
  background: var(--bg-card-subtle);
  border: 1px solid var(--border-color);
}
.sugg-info {
  flex: 1;
  min-width: 0;
}
.sugg-name-row {
  display: flex;
  align-items: center;
  gap: 6px;
  margin-bottom: 4px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.sugg-title {
  font-size: 13px;
  font-weight: 600;
  color: var(--text-main);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.sugg-sub-name {
  font-size: 12px;
  color: var(--text-muted);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.sugg-tags-row {
  display: flex;
  align-items: center;
  gap: 8px;
}
.sugg-fav-icon {
  color: #fbbf24;
  font-size: 14px;
  display: flex;
  align-items: center;
}
.sugg-source-badge {
  font-size: 10px;
  padding: 1px 5px;
  border-radius: 4px;
  background: var(--accent-tag-bg);
  color: var(--accent-cyan);
  font-weight: 600;
}
.sugg-action {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 12px;
  color: var(--text-muted);
  opacity: 0.7;
  transition: opacity 0.15s ease;
  white-space: nowrap;
}
.sugg-item:hover .sugg-action,
.sugg-item.is-active .sugg-action {
  color: var(--accent-cyan);
  opacity: 1;
}
.sugg-empty {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  padding: 24px 16px;
  font-size: 13px;
  color: var(--text-muted);
}
.sugg-footer-search {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 10px 16px;
  margin-top: 4px;
  border-top: 1px solid var(--border-color);
  background: var(--bg-card-subtle);
  cursor: pointer;
  transition: all 0.15s ease;
}
.sugg-footer-search:hover,
.sugg-footer-search.is-active {
  background: var(--bg-card-hover);
  color: var(--accent-cyan);
}
.sugg-footer-content {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 13px;
  font-weight: 500;
  color: var(--text-main);
}
.sugg-footer-search:hover .sugg-footer-content,
.sugg-footer-search.is-active .sugg-footer-content {
  color: var(--accent-cyan);
}
.sugg-kbd {
  font-size: 11px;
  background: var(--border-color);
  color: var(--text-secondary);
  padding: 2px 6px;
  border-radius: 4px;
  font-family: monospace;
}
.sugg-section {
  padding: 8px 16px;
}
.sugg-section:not(:last-child) {
  border-bottom: 1px solid var(--border-color);
}
.sugg-history-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin-top: 8px;
}
.sugg-tag {
  cursor: pointer;
  background: var(--bg-card-subtle);
  border-color: var(--border-color);
  color: var(--text-secondary);
  transition: all 0.2s ease;
}
.sugg-tag:hover {
  border-color: var(--accent-cyan);
  color: var(--accent-cyan);
}
.sugg-quick-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 8px;
  margin-top: 8px;
}
.sugg-quick-card {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 6px 8px;
  border-radius: 6px;
  background: var(--bg-card-subtle);
  border: 1px solid var(--border-color);
  cursor: pointer;
  transition: all 0.2s ease;
}
.sugg-quick-card:hover {
  background: var(--bg-card-hover);
  border-color: var(--accent-cyan);
}
.sugg-quick-thumb {
  width: 32px;
  height: 32px;
  border-radius: 4px;
  object-fit: cover;
  flex-shrink: 0;
}
.sugg-quick-title {
  font-size: 12px;
  font-weight: 500;
  color: var(--text-main);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
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
.home-section {
  margin-bottom: 24px;
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
  font-weight: 700;
  color: var(--text-main);
}
.count-tag {
  font-size: 13px;
  font-weight: 400;
  color: var(--text-muted);
}
.text-amber {
  color: #fbbf24;
}
.favorites-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
  gap: 14px;
}
.favorite-card {
  background: var(--bg-card);
  border: 1px solid var(--border-color);
  border-radius: 8px;
  padding: 10px;
  display: flex;
  align-items: center;
  gap: 12px;
  cursor: pointer;
  transition: all 0.2s ease;
}
.favorite-card:hover {
  transform: translateY(-2px);
  border-color: #fbbf24;
  background: var(--bg-card-hover);
}
.favorite-cover {
  width: 48px;
  height: 48px;
  border-radius: 6px;
  object-fit: cover;
  flex-shrink: 0;
}
.favorite-info {
  min-width: 0;
  flex: 1;
}
.fav-name-cn {
  font-size: 13px;
  font-weight: 600;
  color: var(--text-main);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.fav-name-en {
  font-size: 11px;
  color: var(--text-muted);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.dual-columns-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 16px;
  margin-bottom: 24px;
}
.column-card {
  background: var(--bg-card);
  border: 1px solid var(--border-color);
  border-radius: 8px;
  padding: 16px;
}
.column-header {
  margin-bottom: 12px;
}
.column-title {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 14px;
  font-weight: 700;
  color: var(--text-main);
}
.item-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}
.list-row {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 8px 10px;
  border-radius: 6px;
  cursor: pointer;
  background: var(--bg-card-subtle);
  transition: all 0.2s ease;
}
.list-row:hover {
  background: var(--bg-card-hover);
}
.row-thumb {
  width: 36px;
  height: 36px;
  border-radius: 4px;
  object-fit: cover;
  flex-shrink: 0;
}
.trainer-avatar {
  width: 36px;
  height: 36px;
  border-radius: 4px;
  background: var(--accent-tag-bg);
  color: var(--accent-cyan);
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: bold;
  font-size: 16px;
  flex-shrink: 0;
}
.row-info {
  flex: 1;
  min-width: 0;
}
.row-title {
  font-size: 13px;
  font-weight: 600;
  color: var(--text-main);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.row-sub {
  font-size: 11px;
  color: var(--text-muted);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.row-time {
  font-size: 11px;
  color: var(--text-gray);
  white-space: nowrap;
  flex-shrink: 0;
}
.column-empty {
  font-size: 12px;
  color: var(--text-muted);
  text-align: center;
  padding: 24px 0;
}
.trainers-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(240px, 1fr));
  gap: 16px;
}
.latest-empty-card {
  padding: 40px 20px;
  background: var(--bg-card);
  border: 1px dashed var(--border-color);
  border-radius: 8px;
  display: flex;
  justify-content: center;
  align-items: center;
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
  transform: translateY(-2px);
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
  left: 8px;
  background: rgba(0, 0, 0, 0.75);
  color: var(--accent-cyan);
  font-size: 10px;
  font-weight: 700;
  padding: 2px 6px;
  border-radius: 4px;
}
.star-btn {
  position: absolute;
  top: 8px;
  right: 8px;
  width: 28px;
  height: 28px;
  border-radius: 50%;
  background: rgba(15, 23, 42, 0.7);
  border: 1px solid rgba(255, 255, 255, 0.15);
  color: #94a3b8;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  transition: all 0.2s ease;
  z-index: 2;
}
.star-btn:hover {
  background: rgba(15, 23, 42, 0.9);
  color: #fbbf24;
  transform: scale(1.1);
  border-color: #fbbf24;
}
.star-btn.is-fav {
  color: #fbbf24;
  background: rgba(251, 191, 36, 0.2);
  border-color: #fbbf24;
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
  margin-bottom: 12px;
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
.mt-6 {
  margin-top: 24px;
}
.status-banner {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px 16px;
  border-radius: 8px;
  margin-bottom: 16px;
  font-size: 13px;
  line-height: 1.5;
  transition: all 0.3s ease;
}
.status-banner.retrying {
  background: rgba(245, 158, 11, 0.12);
  border: 1px solid rgba(245, 158, 11, 0.35);
  color: #f59e0b;
}
.status-banner.offline {
  background: rgba(239, 68, 68, 0.12);
  border: 1px solid rgba(239, 68, 68, 0.35);
  color: #ef4444;
}
.status-banner.stale {
  background: rgba(14, 165, 233, 0.12);
  border: 1px solid rgba(14, 165, 233, 0.35);
  color: #38bdf8;
}
.banner-tag {
  margin-left: auto;
}
.latest-error-card {
  padding: 32px 20px;
  background: var(--bg-card);
  border: 1px solid var(--border-color);
  border-radius: 8px;
  display: flex;
  justify-content: center;
  align-items: center;
}
.trainer-card-skeleton {
  background: var(--bg-card);
  border: 1px solid var(--border-color);
  border-radius: 8px;
  overflow: hidden;
}
.item-list-skeleton {
  display: flex;
  flex-direction: column;
  gap: 8px;
}
.skeleton-row {
  padding: 8px 10px;
  border-radius: 6px;
  background: var(--bg-card-subtle);
}
</style>
