import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { api } from '../api/wails'
import type { DownloadTask, LocalTrainerDetail, Game, AppConfig, SearchResult } from '../types'

export const useAppStore = defineStore('app', () => {
  const downloadTasks = ref<DownloadTask[]>([])
  const localTrainers = ref<LocalTrainerDetail[]>([])
  const favorites = ref<Game[]>([])
  const config = ref<AppConfig | null>(null)
  const latestTrainers = ref<SearchResult[]>([])
  const isSyncing = ref(false)

  // 计算属性：当前活跃下载数
  const activeDownloadsCount = computed(() => {
    return downloadTasks.value.filter(t => t.status === 'downloading' || t.status === 'pending').length
  })

  // 刷新下载任务列表
  async function refreshDownloads() {
    try {
      downloadTasks.value = await api.listDownloadTasks()
    } catch (e) {
      console.error('Failed to list download tasks', e)
    }
  }

  // 刷新本地修改器列表
  async function refreshLocalTrainers() {
    try {
      localTrainers.value = await api.getLocalTrainers()
    } catch (e) {
      console.error('Failed to get local trainers', e)
    }
  }

  // 刷新收藏列表
  async function refreshFavorites() {
    try {
      favorites.value = await api.listFavorites()
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

  // 刷新最新推荐修改器
  async function refreshLatest() {
    try {
      latestTrainers.value = await api.getLatestTrainers(1)
    } catch (e) {
      console.error('Failed to get latest trainers', e)
    }
  }

  // 接收下载进度事件更新
  function handleProgressUpdate(task: DownloadTask) {
    const idx = downloadTasks.value.findIndex(t => t.id === task.id)
    if (idx !== -1) {
      downloadTasks.value[idx] = task
    } else {
      downloadTasks.value.unshift(task)
    }

    // 若下载完成，自动刷新本地库
    if (task.status === 'completed') {
      refreshLocalTrainers()
    }
  }

  return {
    downloadTasks,
    localTrainers,
    favorites,
    config,
    latestTrainers,
    isSyncing,
    activeDownloadsCount,
    refreshDownloads,
    refreshLocalTrainers,
    refreshFavorites,
    refreshConfig,
    refreshLatest,
    handleProgressUpdate
  }
})

