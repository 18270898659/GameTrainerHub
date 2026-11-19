<template>
  <div class="downloads-view">
    <el-tabs v-model="activeTab" class="custom-tabs" @tab-change="handleTabChange">
      <!-- 正在下载 / 活动任务 -->
      <el-tab-pane :label="t('downloads.tabActive', { count: store.downloadTasks.length })" name="active">
        <div v-if="store.downloadTasks.length > 0">
          <div class="batch-bar">
            <div class="batch-actions">
              <el-button size="small" plain @click="pauseAll">
                <el-icon><VideoPause /></el-icon> {{ t('downloads.pauseAll') }}
              </el-button>
              <el-button size="small" type="primary" plain @click="resumeAll">
                <el-icon><VideoPlay /></el-icon> {{ t('downloads.resumeAll') }}
              </el-button>
            </div>
            <div class="filter-radios">
              <el-radio-group v-model="statusFilter" size="small">
                <el-radio-button label="all">全部 ({{ store.downloadTasks.length }})</el-radio-button>
                <el-radio-button label="downloading">下载中</el-radio-button>
                <el-radio-button label="paused">已暂停</el-radio-button>
                <el-radio-button label="failed">失败</el-radio-button>
                <el-radio-button v-if="store.recentlyCompletedTasks.length > 0" label="completed">
                  已完成 ({{ store.recentlyCompletedTasks.length }})
                </el-radio-button>
              </el-radio-group>
            </div>
          </div>

          <div v-if="filteredTasks.length > 0" class="task-list">
            <div
              v-for="task in filteredTasks"
              :key="task.id"
              class="task-card"
              :class="{ 'card-failed': task.status === 'failed', 'card-retrying': task.status === 'retrying', 'card-completed': task.status === 'completed' }"
            >
              <div class="task-header">
                <div class="task-names">
                  <!-- 主标题：游戏真实名称 -->
                  <span class="task-game">{{ task.game_name || formatFallbackGame(task.file_name) }}</span>
                  <!-- 副标题：修改器名称与版本号 -->
                  <span class="task-file">
                    {{ (task.trainer_name ? task.trainer_name + ' · ' : '') + (task.version_name || task.file_name) }}
                  </span>
                </div>
                <div class="task-actions">
                  <template v-if="task.status === 'completed'">
                    <el-button size="small" type="primary" @click="goToLibrary">
                      <el-icon><Collection /></el-icon> 前往本地修改器
                    </el-button>
                    <el-button size="small" plain @click="openFolder(task.save_path)">
                      <el-icon><FolderOpened /></el-icon> {{ t('downloads.locateFile') }}
                    </el-button>
                  </template>
                  <template v-else>
                    <el-button
                      v-if="task.status === 'downloading' || task.status === 'retrying'"
                      size="small"
                      plain
                      @click="pauseTask(task.id)"
                    >
                      <el-icon><VideoPause /></el-icon> {{ t('downloads.pause') }}
                    </el-button>
                    <el-button
                      v-else-if="task.status === 'paused'"
                      size="small"
                      type="primary"
                      plain
                      @click="resumeTask(task.id)"
                    >
                      <el-icon><VideoPlay /></el-icon> {{ t('downloads.resume') }}
                    </el-button>
                    <el-button
                      v-else-if="task.status === 'failed'"
                      size="small"
                      type="warning"
                      plain
                      @click="resumeTask(task.id)"
                    >
                      <el-icon><RefreshRight /></el-icon> {{ t('downloads.retry') }}
                    </el-button>
                    <el-button
                      size="small"
                      type="danger"
                      plain
                      @click="cancelTask(task.id)"
                    >
                      <el-icon><Delete /></el-icon> {{ t('downloads.cancel') }}
                    </el-button>
                  </template>
                </div>
              </div>

              <div class="task-progress-wrap">
                <el-progress
                  :percentage="getTaskProgress(task)"
                  :indeterminate="task.status === 'downloading' && (!task.total_bytes || task.total_bytes <= 0)"
                  :status="task.status === 'completed' ? 'success' : task.status === 'failed' ? 'exception' : task.status === 'retrying' ? 'warning' : ''"
                  :stroke-width="8"
                />
              </div>

              <!-- 异常与重试提示条 -->
              <div v-if="task.error_msg && (task.status === 'failed' || task.status === 'retrying')" class="task-error-alert">
                <el-icon><WarningFilled /></el-icon>
                <span>{{ task.error_msg }}</span>
              </div>

              <div class="task-footer">
                <span class="task-bytes">
                  {{ formatBytes(task.downloaded_bytes || task.total_bytes) }} / {{ (task.total_bytes && task.total_bytes > 0) ? formatBytes(task.total_bytes) : t('downloads.unknownSize') }}
                </span>
                <span v-if="task.status === 'downloading'" class="task-speed">
                  {{ formatBytes(task.speed_bytes_per_sec) }}/s
                </span>
                <span v-else-if="task.status === 'completed'" class="task-status-text status-completed">
                  <el-icon><Check /></el-icon> 下载完成，已导入本地修改器库
                </span>
                <span v-else class="task-status-text" :class="`status-${task.status}`">
                  {{ t('downloads.statusLabel') }}{{ getStatusLabel(task.status, task.retries) }}
                </span>
              </div>
            </div>
          </div>
          <div v-else class="empty-wrap">
            <el-empty :description="store.downloadTasks.length > 0 ? '当前筛选下暂无任务' : t('downloads.emptyTasks')" />
          </div>
        </div>

        <div v-else-if="store.recentlyCompletedTasks.length > 0" class="completed-summary-wrap">
          <div class="completion-banner">
            <el-alert
              title="修改器下载完成"
              type="success"
              description="任务已成功下载并自动安全导入本地修改器库，可随时安全启动！"
              show-icon
              :closable="false"
            />
          </div>

          <div class="task-list">
            <div
              v-for="task in store.recentlyCompletedTasks"
              :key="task.id"
              class="task-card card-completed"
            >
              <div class="task-header">
                <div class="task-names">
                  <span class="task-game">{{ task.game_name || formatFallbackGame(task.file_name) }}</span>
                  <span class="task-file">
                    {{ (task.trainer_name ? task.trainer_name + ' · ' : '') + (task.version_name || task.file_name) }}
                  </span>
                </div>
                <div class="task-actions">
                  <el-button size="small" type="primary" @click="goToLibrary">
                    <el-icon><Collection /></el-icon> 前往本地修改器
                  </el-button>
                  <el-button size="small" plain @click="openFolder(task.save_path)">
                    <el-icon><FolderOpened /></el-icon> {{ t('downloads.locateFile') }}
                  </el-button>
                  <el-button size="small" plain @click="activeTab = 'history'">
                    {{ t('downloads.tabHistory') }}
                  </el-button>
                </div>
              </div>

              <div class="task-progress-wrap">
                <el-progress :percentage="100" status="success" :stroke-width="8" />
              </div>

              <div class="task-footer">
                <span class="task-bytes">{{ formatBytes(task.downloaded_bytes || task.total_bytes) }}</span>
                <span class="task-status-text status-completed">
                  <el-icon><Check /></el-icon> 下载完成，已导入本地修改器库
                </span>
              </div>
            </div>
          </div>
        </div>

        <div v-else class="empty-wrap">
          <el-empty :description="t('downloads.emptyTasks')" />
        </div>
      </el-tab-pane>

      <!-- 下载历史 -->
      <el-tab-pane :label="t('downloads.tabHistory')" name="history">
        <div v-if="historyList.length > 0">
          <div class="history-toolbar">
            <span class="history-count-label">共 {{ historyList.length }} 条下载历史记录</span>
            <el-button size="small" type="danger" plain @click="promptClearHistory">
              <el-icon><Delete /></el-icon> {{ t('downloads.clearAllHistory') }}
            </el-button>
          </div>

          <el-table :data="historyList" stripe style="width: 100%" class="custom-table">
            <el-table-column prop="game_name" :label="t('downloads.historyGame')" min-width="160" show-overflow-tooltip />
            <el-table-column :label="t('downloads.historyTrainer')" min-width="170" show-overflow-tooltip>
              <template #default="{ row }">
                <span>{{ (row.trainer_name ? row.trainer_name + ' · ' : '') + (row.version_name || '-') }}</span>
              </template>
            </el-table-column>
            <el-table-column prop="file_name" :label="t('downloads.historyFile')" min-width="180" show-overflow-tooltip />
            <el-table-column :label="t('downloads.historySize')" width="100">
              <template #default="{ row }">
                {{ formatBytes(row.file_size) }}
              </template>
            </el-table-column>
            <el-table-column :label="t('downloads.fileStatus')" width="100">
              <template #default="{ row }">
                <el-tag v-if="row.file_exists !== false" type="success" size="small">{{ t('downloads.fileExists') }}</el-tag>
                <el-tag v-else type="danger" size="small">{{ t('downloads.fileMissing') }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column label="SHA-256" width="180">
              <template #default="{ row }">
                <span class="hash-short" :title="row.sha256">{{ (row.sha256 || '').substring(0, 12) }}...</span>
                <el-button link type="primary" size="small" @click="copy(row.sha256)">{{ t('security.copy') }}</el-button>
              </template>
            </el-table-column>
            <el-table-column :label="t('downloads.historyTime')" width="150">
              <template #default="{ row }">
                {{ formatDate(row.completed_at) }}
              </template>
            </el-table-column>
            <el-table-column :label="t('downloads.historyAction')" width="210" fixed="right">
              <template #default="{ row }">
                <el-button size="small" link type="primary" @click="openFolder(row.save_path)">
                  <el-icon><FolderOpened /></el-icon> {{ t('downloads.locateFile') }}
                </el-button>
                <el-button size="small" link type="warning" @click="redownload(row)">
                  <el-icon><RefreshRight /></el-icon> {{ t('downloads.redownload') }}
                </el-button>
                <el-button size="small" link type="danger" @click="deleteHistoryItem(row.id)">
                  <el-icon><Delete /></el-icon> {{ t('downloads.deleteHistory') }}
                </el-button>
              </template>
            </el-table-column>
          </el-table>
        </div>

        <div v-else class="empty-wrap">
          <el-empty :description="t('downloads.emptyHistory')" />
        </div>
      </el-tab-pane>
    </el-tabs>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onActivated, watch } from 'vue'
import { useRouter } from 'vue-router'
import { VideoPause, VideoPlay, Delete, RefreshRight, FolderOpened, WarningFilled, Collection, Check } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useAppStore } from '../stores/appStore'
import { api } from '../api/wails'
import { t } from '../i18n'
import type { DownloadHistory, DownloadTask } from '../types'

defineOptions({ name: 'DownloadsView' })

const router = useRouter()
const store = useAppStore()
const activeTab = ref('active')
const statusFilter = ref('all')
const historyList = ref<DownloadHistory[]>([])

const filteredTasks = computed(() => {
  if (statusFilter.value === 'completed') {
    return store.recentlyCompletedTasks
  }
  if (statusFilter.value === 'all') return store.downloadTasks
  if (statusFilter.value === 'downloading') {
    return store.downloadTasks.filter(t => t.status === 'downloading' || t.status === 'retrying')
  }
  return store.downloadTasks.filter(t => t.status === statusFilter.value)
})

watch(() => store.recentlyCompletedTasks.length, () => {
  loadHistory()
})

function goToLibrary() {
  router.push('/library')
}

function getTaskProgress(task?: any): number {
  if (!task) return 0
  if (task.status === 'completed') return 100
  if (task.total_bytes && task.total_bytes > 0) {
    const val = (task.progress !== undefined && task.progress !== null && !isNaN(task.progress))
      ? task.progress
      : ((task.downloaded_bytes || 0) / task.total_bytes * 100)
    return Math.min(100, Math.max(0, Number(Number(val).toFixed(1))))
  }
  return 0
}

onMounted(async () => {
  await store.refreshDownloads()
  loadHistory()
})

onActivated(async () => {
  await store.refreshDownloads()
  await loadHistory()
})

async function handleTabChange(tabName: any) {
  if (tabName === 'active') {
    await store.refreshDownloads()
  } else if (tabName === 'history') {
    await loadHistory()
  }
}

async function loadHistory() {
  try {
    historyList.value = await api.listDownloadHistory(100)
  } catch (e: any) {
    console.error('Failed to load history', e)
  }
}

async function pauseAll() {
  try {
    await api.pauseAllDownloads()
    await store.refreshDownloads()
    ElMessage.success(t('downloads.pauseAllSuccess'))
  } catch (e: any) {
    ElMessage.error(e?.message || String(e))
  }
}

async function resumeAll() {
  try {
    await api.resumeAllDownloads()
    await store.refreshDownloads()
    ElMessage.success(t('downloads.resumeAllSuccess'))
  } catch (e: any) {
    ElMessage.error(e?.message || String(e))
  }
}

async function pauseTask(id: number) {
  try {
    await api.pauseDownload(id)
    await store.refreshDownloads()
  } catch (e: any) {
    ElMessage.error(e?.message || String(e))
  }
}

async function resumeTask(id: number) {
  try {
    await api.resumeDownload(id)
    await store.refreshDownloads()
  } catch (e: any) {
    ElMessage.error(e?.message || String(e))
  }
}

async function cancelTask(id: number) {
  try {
    await api.cancelDownload(id)
    await store.refreshDownloads()
  } catch (e: any) {
    ElMessage.error(e?.message || String(e))
  }
}

async function openFolder(path: string) {
  try {
    await api.openContainingFolder(path)
  } catch (e: any) {
    ElMessage.error(e?.message || String(e))
  }
}

async function redownload(row: DownloadHistory) {
  if (!row.version_id) {
    ElMessage.warning('该记录无版本关联信息，无法直接重新下载')
    return
  }
  try {
    await api.startDownload(row.version_id)
    ElMessage.success(t('downloads.redownloadStarted'))
    activeTab.value = 'active'
    await store.refreshDownloads()
  } catch (e: any) {
    ElMessage.error('创建重新下载任务失败: ' + (e?.message || e))
  }
}

async function deleteHistoryItem(id: number) {
  try {
    await api.deleteDownloadHistory(id)
    ElMessage.success('已删除下载记录')
    await loadHistory()
  } catch (e: any) {
    ElMessage.error('删除下载记录失败: ' + (e?.message || e))
  }
}

async function promptClearHistory() {
  try {
    await ElMessageBox.confirm(
      t('downloads.clearHistoryConfirm'),
      t('downloads.clearAllHistory'),
      {
        confirmButtonText: t('downloads.clearAllHistory'),
        cancelButtonText: t('library.cancel'),
        type: 'warning'
      }
    )
    await api.clearDownloadHistory()
    ElMessage.success(t('downloads.clearHistorySuccess'))
    await loadHistory()
  } catch (e: any) {
    if (e !== 'cancel') {
      ElMessage.error('清空历史记录失败: ' + (e?.message || e))
    }
  }
}

function formatFallbackGame(fileName: string): string {
  if (!fileName) return '未知游戏'
  return fileName.replace(/\.(exe|zip|rar)$/i, '').replace(/plus\s*\d+.*$/i, '').trim() || fileName
}

function formatDate(dateStr: string): string {
  if (!dateStr) return ''
  try {
    const d = new Date(dateStr)
    const y = d.getFullYear()
    const m = (d.getMonth() + 1).toString().padStart(2, '0')
    const day = d.getDate().toString().padStart(2, '0')
    const h = d.getHours().toString().padStart(2, '0')
    const min = d.getMinutes().toString().padStart(2, '0')
    return `${y}-${m}-${day} ${h}:${min}`
  } catch {
    return dateStr
  }
}

function copy(text: string) {
  if (!text) return
  navigator.clipboard.writeText(text)
  ElMessage.success(t('security.copied'))
}

function formatBytes(bytes: number): string {
  if (!bytes || bytes <= 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return (bytes / Math.pow(k, i)).toFixed(2) + ' ' + sizes[i]
}

function getStatusLabel(status: string, retries?: number): string {
  switch (status) {
    case 'downloading': return t('downloads.statusDownloading')
    case 'retrying': return t('downloads.statusRetrying', { retries: retries || 1 })
    case 'paused': return t('downloads.statusPaused')
    case 'completed': return t('downloads.statusCompleted')
    case 'failed': return t('downloads.statusFailed')
    default: return t('downloads.statusPending')
  }
}
</script>

<style scoped>
.batch-bar {
  display: flex;
  gap: 10px;
  margin-bottom: 16px;
}
.history-toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
}
.history-count-label {
  font-size: 13px;
  color: var(--text-muted);
}
.task-list {
  display: flex;
  flex-direction: column;
  gap: 16px;
}
.task-card {
  background: var(--bg-card);
  border: 1px solid var(--border-color);
  border-radius: 8px;
  padding: 16px;
  box-shadow: var(--card-shadow);
  transition: all 0.2s ease;
}
.task-card.card-retrying {
  border-color: #f59e0b;
  background: var(--bg-card-subtle);
}
.task-card.card-failed {
  border-color: #ef4444;
}
.task-card.card-completed {
  border-left: 4px solid #10b981;
}
.completed-summary-wrap {
  display: flex;
  flex-direction: column;
  gap: 16px;
}
.completion-banner {
  margin-bottom: 4px;
}
.status-completed {
  color: #10b981;
  display: inline-flex;
  align-items: center;
  gap: 4px;
  font-weight: 600;
}
.task-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
}
.task-names {
  display: flex;
  flex-direction: column;
  gap: 4px;
}
.task-game {
  font-size: 15px;
  font-weight: 700;
  color: var(--text-main);
}
.task-file {
  font-size: 13px;
  color: var(--text-muted);
}
.task-actions {
  display: flex;
  gap: 8px;
}
.task-progress-wrap {
  margin-bottom: 12px;
}
.task-error-alert {
  display: flex;
  align-items: center;
  gap: 8px;
  background: rgba(239, 68, 68, 0.12);
  border: 1px solid rgba(239, 68, 68, 0.3);
  color: #ef4444;
  padding: 8px 12px;
  border-radius: 6px;
  font-size: 12px;
  margin-bottom: 12px;
}
.task-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 12px;
}
.task-bytes {
  color: var(--text-secondary);
}
.task-speed {
  color: var(--accent-cyan);
  font-family: monospace;
  font-weight: 600;
}
.task-status-text {
  font-weight: 500;
}
.status-downloading {
  color: var(--accent-cyan);
}
.status-paused {
  color: #f59e0b;
}
.status-failed {
  color: #ef4444;
}
.status-retrying {
  color: #f59e0b;
}
.status-completed {
  color: #10b981;
}
.empty-wrap {
  padding: 40px 0;
}
.hash-short {
  font-family: monospace;
  font-size: 12px;
  color: var(--text-muted);
  margin-right: 6px;
}
</style>
