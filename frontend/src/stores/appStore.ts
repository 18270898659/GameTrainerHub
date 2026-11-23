import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { api } from '../api/wails'
import { fetchWithRetry, type LoadingState } from '../utils/retry'
import { DEFAULT_SEED_LATEST } from '../utils/seedData'
import type { DownloadTask, DownloadHistory, LocalTrainerDetail, Game, AppConfig, SearchResult } from '../types'

const CACHE_KEY_LATEST = 'gth_latest_trainers_cache'
const CACHE_KEY_TIME = 'gth_latest_trainers_time'

export const useAppStore = defineStore('app', () => {
  const downloadTasks = ref<DownloadTask[]>([])
  const downloadHistory = ref<DownloadHistory[]>([])
  const recentlyCompletedTasks = ref<DownloadTask[]>([])
  const localTrainers = ref<LocalTrainerDetail[]>([])
  const favorites = ref<Game[]>([])
  const config = ref<AppConfig | null>(null)
  const isSyncing = ref(false)

  // 首页最新修改器状态机与本地缓存（默认装载种子数据，确保 0ms 秒开且绝对不为空）
  const latestTrainers = ref<SearchResult[]>([...DEFAULT_SEED_LATEST])
  const latestStatus = ref<LoadingState>('stale')
  const latestError = ref<string | null>(null)
  const latestRetryCount = ref<number>(0)
  const isLatestStale = ref<boolean>(true)
  const latestLastUpdated = ref<number | null>(null)

  // 1. 初始化时尝试恢复本地持久化缓存（若存在则覆盖种子数据）
  try {
    const cachedStr = localStorage.getItem(CACHE_KEY_LATEST)
    const cachedTime = localStorage.getItem(CACHE_KEY_TIME)
    if (cachedStr) {
      const parsed = JSON.parse(cachedStr)
      if (Array.isArray(parsed) && parsed.length > 0) {
        latestTrainers.value = parsed
        latestStatus.value = 'stale'
        isLatestStale.value = true
        latestLastUpdated.value = cachedTime ? Number(cachedTime) : Date.now()
      }
    }
  } catch (e) {
    console.warn('Failed to parse cached latest trainers from localStorage', e)
  }

  // 计算属性：当前活跃下载数
  const activeDownloadsCount = computed(() => {
    if (!Array.isArray(downloadTasks.value)) return 0
    return downloadTasks.value.filter(
      (t) => t && (t.status === 'downloading' || t.status === 'pending' || t.status === 'retrying')
    ).length
  })

  // 刷新下载任务列表
  async function refreshDownloads() {
    try {
      const res = await api.listDownloadTasks()
      downloadTasks.value = Array.isArray(res) ? res : []
    } catch (e) {
      console.error('Failed to list download tasks', e)
      if (!Array.isArray(downloadTasks.value)) {
        downloadTasks.value = []
      }
    }
  }

  // 刷新下载历史记录
  async function refreshHistory(limit: number = 100) {
    try {
      const res = await api.listDownloadHistory(limit)
      downloadHistory.value = Array.isArray(res) ? res : []
    } catch (e) {
      console.error('Failed to list download history', e)
      if (!Array.isArray(downloadHistory.value)) {
        downloadHistory.value = []
      }
    }
  }

  // 刷新本地修改器列表
  async function refreshLocalTrainers() {
    try {
      const res = await api.getLocalTrainers()
      if (Array.isArray(res)) {
        localTrainers.value = res
      }
    } catch (e) {
      console.error('Failed to get local trainers', e)
    }
  }

  // 刷新收藏列表
  async function refreshFavorites() {
    try {
      const res = await api.listFavorites()
      if (Array.isArray(res)) {
        favorites.value = res
      }
    } catch (e) {
      console.error('Failed to list favorites', e)
    }
  }

  // 刷新配置
  async function refreshConfig() {
    try {
      config.value = await api.getConfig()
    } catch (e) {
      console.error('Failed to get config', e)
    }
  }

  // 请求去重与竞态序号
  let activeRefreshPromise: Promise<SearchResult[]> | null = null
  let currentLatestSeq = 0

  /**
   * 刷新最新推荐修改器（支持指数退避重试、状态机流转、本地缓存兜底与竞态防护）
   */
  async function refreshLatest(
    opts?: boolean | { forceRefresh?: boolean; silent?: boolean }
  ): Promise<SearchResult[]> {
    const options = typeof opts === 'boolean' ? { forceRefresh: opts } : (opts || {})
    const seq = ++currentLatestSeq

    // 请求去重：如果已有正在进行的请求且非强制刷新，复用现有 Promise
    if (activeRefreshPromise && !options.forceRefresh) {
      return activeRefreshPromise
    }

    const hasExistingData = latestTrainers.value.length > 0
    if (!options.silent) {
      if (hasExistingData) {
        isLatestStale.value = true
        latestStatus.value = 'stale'
      } else {
        latestStatus.value = 'loading'
      }
    }
    latestError.value = null
    latestRetryCount.value = 0

    activeRefreshPromise = (async () => {
      try {
        const res = await fetchWithRetry(
          async () => {
            return await api.getLatestTrainers(1)
          },
          {
            maxRetries: 3,
            baseDelayMs: 600,
            maxDelayMs: 3000,
            timeoutMs: 6000,
            onRetry: (attempt, delayMs, err) => {
              if (seq === currentLatestSeq) {
                latestStatus.value = 'retrying'
                latestRetryCount.value = attempt
                latestError.value = `正在连接数据源重试 (${attempt}/3): ${err?.message || err}`
              }
            }
          }
        )

        // 竞态校验：仅当自身仍是最新一次请求时采纳结果
        if (seq === currentLatestSeq) {
          const list = Array.isArray(res) && res.length > 0 ? res : (latestTrainers.value.length > 0 ? latestTrainers.value : DEFAULT_SEED_LATEST)
          latestTrainers.value = list
          latestLastUpdated.value = Date.now()
          isLatestStale.value = false
          latestStatus.value = 'success'
          latestError.value = null

          // 写入本地持久化缓存
          try {
            localStorage.setItem(CACHE_KEY_LATEST, JSON.stringify(list))
            localStorage.setItem(CACHE_KEY_TIME, String(Date.now()))
          } catch (e) {
            console.warn('Failed to cache latest trainers to localStorage', e)
          }
          return list
        }
        return latestTrainers.value
      } catch (e: any) {
        if (seq === currentLatestSeq) {
          const errorMsg = e?.message || String(e)
          latestError.value = errorMsg
          console.error('Failed to refresh latest trainers after retries:', errorMsg)

          // 核心兜底：确保界面始终具备数据呈现，绝不清空
          if (latestTrainers.value.length === 0) {
            latestTrainers.value = [...DEFAULT_SEED_LATEST]
          }
          latestStatus.value = 'stale'
          isLatestStale.value = true
        }
        return latestTrainers.value
      } finally {
        if (seq === currentLatestSeq) {
          activeRefreshPromise = null
        }
      }
    })()

    return activeRefreshPromise
  }

  // 接收下载进度事件更新
  function handleProgressUpdate(task: DownloadTask) {
    if (!task) return
    if (!Array.isArray(downloadTasks.value)) {
      downloadTasks.value = []
    }
    const idx = downloadTasks.value.findIndex((t) => t.id === task.id)
    if (idx !== -1) {
      downloadTasks.value[idx] = { ...downloadTasks.value[idx], ...task }
    } else {
      downloadTasks.value.unshift(task)
    }

    if (task.status === 'completed') {
      const cIdx = recentlyCompletedTasks.value.findIndex((t) => t.id === task.id)
      if (cIdx !== -1) {
        recentlyCompletedTasks.value[cIdx] = task
      } else {
        recentlyCompletedTasks.value.unshift(task)
        if (recentlyCompletedTasks.value.length > 20) {
          recentlyCompletedTasks.value.pop()
        }
      }
      refreshLocalTrainers()
      refreshHistory()
      refreshDownloads()
    }
  }

  return {
    downloadTasks,
    downloadHistory,
    recentlyCompletedTasks,
    localTrainers,
    favorites,
    config,
    latestTrainers,
    latestStatus,
    latestError,
    latestRetryCount,
    isLatestStale,
    latestLastUpdated,
    isSyncing,
    activeDownloadsCount,
    refreshDownloads,
    refreshHistory,
    refreshLocalTrainers,
    refreshFavorites,
    refreshConfig,
    refreshLatest,
    handleProgressUpdate
  }
})
