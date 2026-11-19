import { defineStore } from 'pinia'
import { ref } from 'vue'
import { ElMessage } from 'element-plus'
import { api } from '../api/wails'
import type { SearchResult, TrainerVersion } from '../types'

const HISTORY_STORAGE_KEY = 'gth_search_history'

function loadHistory(): string[] {
  try {
    const raw = localStorage.getItem(HISTORY_STORAGE_KEY)
    if (raw) {
      const parsed = JSON.parse(raw)
      if (Array.isArray(parsed)) {
        return parsed
      }
    }
  } catch (e) {
    console.warn('Failed to parse search history', e)
  }
  return []
}

function saveHistory(list: string[]) {
  try {
    localStorage.setItem(HISTORY_STORAGE_KEY, JSON.stringify(list))
  } catch (e) {
    console.warn('Failed to save search history', e)
  }
}

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

  // 搜索历史记录
  const searchHistory = ref<string[]>(loadHistory())

  function addSearchHistory(keyword: string) {
    const kw = keyword.trim()
    if (!kw) return
    const filtered = searchHistory.value.filter(k => k.toLowerCase() !== kw.toLowerCase())
    filtered.unshift(kw)
    if (filtered.length > 12) {
      filtered.length = 12
    }
    searchHistory.value = filtered
    saveHistory(filtered)
  }

  function removeSearchHistory(keyword: string) {
    searchHistory.value = searchHistory.value.filter(k => k !== keyword)
    saveHistory(searchHistory.value)
  }

  function clearSearchHistory() {
    searchHistory.value = []
    saveHistory([])
  }

  async function executeSearch(targetQuery?: string) {
    if (targetQuery !== undefined) {
      query.value = targetQuery
    }
    const q = query.value.trim()
    if (!q) return

    searching.value = true
    hasSearched.value = true
    addSearchHistory(q)

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

  async function openDetail(item: SearchResult) {
    if (!item) return
    selectedGame.value = item
    drawerVisible.value = true
    loadingDetail.value = true
    detailVersions.value = []

    if (!item.detail_url) {
      loadingDetail.value = false
      ElMessage.info('该资源暂无在线详情链接')
      return
    }

    try {
      const res = await api.getTrainerDetail(item.detail_url)
      detailVersions.value = res?.versions || []
    } catch (e: any) {
      ElMessage.warning('获取版本详情失败: ' + (e?.message || e || '未知错误'))
    } finally {
      loadingDetail.value = false
    }
  }

  function closeDetail() {
    drawerVisible.value = false
  }

  async function downloadVersion(v: TrainerVersion) {
    downloadingVerID.value = v.id
    try {
      await api.startDownload(v.id)
      ElMessage.success('已添加到下载中心')
    } catch (e: any) {
      ElMessage.error('创建下载失败: ' + (e?.message || e))
    } finally {
      downloadingVerID.value = null
    }
  }

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
    addSearchHistory,
    removeSearchHistory,
    clearSearchHistory,
    executeSearch,
    openDetail,
    closeDetail,
    downloadVersion
  }
})
