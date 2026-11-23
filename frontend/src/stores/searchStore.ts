import { defineStore } from 'pinia'
import { ref } from 'vue'
import { ElMessage } from 'element-plus'
import router from '../router'
import { api } from '../api/wails'
import { t } from '../i18n'
import type { SearchResult, TrainerVersion, Game } from '../types'

export const useSearchStore = defineStore('search', () => {
  const query = ref('')
  const localOnly = ref(false)
  const selectedSource = ref('all')
  const searching = ref(false)
  const hasSearched = ref(false)
  const results = ref<SearchResult[]>([])

  // 详情抽屉状态
  const drawerVisible = ref(false)
  const selectedGame = ref<SearchResult | null>(null)
  const loadingDetail = ref(false)
  const detailVersions = ref<TrainerVersion[]>([])
  const downloadingVerID = ref<number | null>(null)

  // 搜索历史与最近访问
  const searchHistory = ref<string[]>([])
  const recentVisited = ref<Game[]>([])

  async function loadSearchHistory() {
    try {
      const history = await api.getSearchHistory(15)
      searchHistory.value = history || []
    } catch (e) {
      console.warn('Failed to load search history from backend', e)
    }
  }

  async function loadRecentVisited() {
    try {
      const visited = await api.getRecentGames(10)
      recentVisited.value = visited || []
    } catch (e) {
      console.warn('Failed to load recent games', e)
    }
  }

  async function addSearchHistory(keyword: string) {
    const kw = keyword.trim()
    if (!kw) return
    const filtered = searchHistory.value.filter(k => k.toLowerCase() !== kw.toLowerCase())
    filtered.unshift(kw)
    if (filtered.length > 15) {
      filtered.length = 15
    }
    searchHistory.value = filtered
    try {
      await api.addSearchHistory(kw)
    } catch (e) {
      console.warn('Failed to persist search history', e)
    }
  }

  async function removeSearchHistory(keyword: string) {
    searchHistory.value = searchHistory.value.filter(k => k !== keyword)
    try {
      await api.removeSearchHistory(keyword)
    } catch (e) {
      console.warn('Failed to remove search history item', e)
    }
  }

  async function clearSearchHistory() {
    searchHistory.value = []
    try {
      await api.clearSearchHistory()
    } catch (e) {
      console.warn('Failed to clear search history', e)
    }
  }

  async function executeSearch(targetQuery?: string) {
    if (targetQuery !== undefined) {
      query.value = targetQuery
    }
    const q = query.value.trim()
    if (!q) return

    searching.value = true
    hasSearched.value = true
    await addSearchHistory(q)

    try {
      results.value = await api.search({
        query: q,
        local_only: localOnly.value,
        source: selectedSource.value,
        page: 1,
        page_size: 50
      })
    } catch (e: any) {
      ElMessage.error('搜索异常: ' + (e?.message || e || '网络请求失败'))
    } finally {
      searching.value = false
    }
  }

  async function toggleFavorite(item: SearchResult | any) {
    if (!item) return
    try {
      let nextStatus = false
      const gameID = item.game_id || item.id || 0
      const slug = item.slug || ''
      const nameEN = item.game_name_en || item.name_en || ''
      const nameCN = item.game_name_cn || item.name_cn || ''
      const coverURL = item.cover_url || ''

      if (gameID && gameID > 0) {
        nextStatus = await api.toggleFavorite(gameID)
      } else if (slug) {
        nextStatus = await api.toggleFavoriteBySlug(slug, nameEN, nameCN, coverURL)
      } else {
        return
      }

      item.is_favorite = nextStatus
      if (selectedGame.value) {
        if ((gameID > 0 && selectedGame.value.game_id === gameID) || (slug && selectedGame.value.slug === slug)) {
          selectedGame.value.is_favorite = nextStatus
        }
      }
      for (const res of results.value) {
        if ((gameID > 0 && res.game_id === gameID) || (slug && res.slug === slug)) {
          res.is_favorite = nextStatus
        }
      }
      ElMessage.success(t(nextStatus ? 'search.favoriteAdded' : 'search.favoriteRemoved'))
      // 同步刷新全局收藏列表
      import('./appStore').then(({ useAppStore }) => {
        useAppStore().refreshFavorites()
      }).catch(() => {})
    } catch (e: any) {
      ElMessage.error('收藏操作失败: ' + (e?.message || e))
    }
  }

  async function openDetail(item: SearchResult) {
    if (!item) return
    selectedGame.value = item
    drawerVisible.value = true
    loadingDetail.value = true
    detailVersions.value = []

    // 记录最近访问
    if (item.game_id && item.game_id > 0) {
      api.recordGameVisit(item.game_id).then(() => {
        loadRecentVisited()
      }).catch(err => console.warn('Record visit failed', err))
    }

    try {
      let res: any = null
      if (item.detail_url) {
        try {
          res = await api.getTrainerDetail(item.detail_url)
        } catch (detailErr) {
          if (item.game_id && item.game_id > 0) {
            res = await api.getGameDetail(item.game_id)
          } else {
            throw detailErr
          }
        }
      } else if (item.game_id && item.game_id > 0) {
        res = await api.getGameDetail(item.game_id)
      } else {
        loadingDetail.value = false
        ElMessage.info('该资源暂无在线详情链接')
        return
      }

      detailVersions.value = res?.versions || []
      if (res?.trainer?.detail_url && !selectedGame.value.detail_url) {
        selectedGame.value.detail_url = res.trainer.detail_url
      }
      if (res?.trainer?.cover_url && !selectedGame.value.cover_url) {
        selectedGame.value.cover_url = res.trainer.cover_url
      }
    } catch (e: any) {
      ElMessage.warning('获取版本详情失败: ' + (e?.message || e || '未知错误'))
    } finally {
      loadingDetail.value = false
    }
  }

  async function openDetailByGame(game: Game) {
    if (!game) return
    const converted: SearchResult = {
      game_id: game.id,
      slug: game.slug,
      game_name_en: game.name_en,
      game_name_cn: game.name_cn,
      cover_url: game.cover_url,
      source_id: 'fling',
      trainer_title: game.name_cn || game.name_en,
      detail_url: game.detail_url || '',
      version_name: '',
      updated_at: game.updated_at,
      is_downloaded: false,
      is_online_only: false,
      is_favorite: game.is_favorite,
      score: 100
    }
    await openDetail(converted)
  }

  function closeDetail() {
    drawerVisible.value = false
  }

  async function downloadVersion(v: TrainerVersion) {
    downloadingVerID.value = v.id
    try {
      await api.startDownload(v.id)
      ElMessage.success(t('search.downloadStarted'))
      // 1. 关闭版本详情抽屉
      drawerVisible.value = false
      // 2. 自动跳转至下载与任务管理界面
      if (router.currentRoute.value.path !== '/downloads') {
        router.push('/downloads')
      }
      // 3. 立即刷新下载任务列表
      import('./appStore').then(({ useAppStore }) => {
        useAppStore().refreshDownloads()
      }).catch(() => {})
    } catch (e: any) {
      ElMessage.error(e?.message || String(e))
    } finally {
      downloadingVerID.value = null
    }
  }

  // 初始化加载历史与访问记录
  loadSearchHistory()
  loadRecentVisited()

  return {
    query,
    localOnly,
    selectedSource,
    searching,
    hasSearched,
    results,
    drawerVisible,
    selectedGame,
    loadingDetail,
    detailVersions,
    downloadingVerID,
    searchHistory,
    recentVisited,
    loadSearchHistory,
    loadRecentVisited,
    addSearchHistory,
    removeSearchHistory,
    clearSearchHistory,
    executeSearch,
    toggleFavorite,
    openDetail,
    openDetailByGame,
    closeDetail,
    downloadVersion
  }
})
