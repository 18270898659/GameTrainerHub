<template>
  <div class="downloads-view">
    <!-- Top View Mode Controller -->
    <div class="view-mode-bar">
      <el-radio-group v-model="viewMode" size="default" class="view-radio-group">
        <el-radio-button value="all">
          {{ t('downloads.tabAll', { count: (store.downloadTasks || []).length + (store.downloadHistory || []).length }) }}
        </el-radio-button>
        <el-radio-button value="active">
          {{ t('downloads.tabActive', { count: (store.downloadTasks || []).length }) }}
        </el-radio-button>
        <el-radio-button value="history">
          {{ t('downloads.tabHistory', { count: (store.downloadHistory || []).length }) }}
        </el-radio-button>
      </el-radio-group>
    </div>

    <!-- 区域一：当前任务 (Active Tasks Section) -->
    <section v-if="viewMode === 'all' || viewMode === 'active'" class="download-section current-tasks-section">
      <div class="section-top-bar">
        <div class="section-heading">
          <el-icon class="heading-icon text-cyan"><Download /></el-icon>
          <span class="heading-title">{{ t('downloads.sectionCurrent') }}</span>
          <el-tag size="small" type="primary" round>{{ (store.downloadTasks || []).length }}</el-tag>
        </div>

        <div class="section-controls">
          <div v-if="(store.downloadTasks || []).length > 0" class="batch-btns">
            <el-button size="small" plain @click="pauseAll">
              <el-icon><VideoPause /></el-icon> {{ t('downloads.pauseAll') }}
            </el-button>
            <el-button size="small" type="primary" plain @click="resumeAll">
              <el-icon><VideoPlay /></el-icon> {{ t('downloads.resumeAll') }}
            </el-button>
            <el-button v-if="hasCompletedTasks" size="small" plain @click="clearCompletedTasks">
              <el-icon><Delete /></el-icon> 清理已完成
            </el-button>
          </div>

          <div class="filter-radios">
            <el-radio-group v-model="statusFilter" size="small">
              <el-radio-button value="all">全部 ({{ (store.downloadTasks || []).length }})</el-radio-button>
              <el-radio-button value="downloading">下载中 ({{ downloadingCount }})</el-radio-button>
              <el-radio-button value="completed">已完成 ({{ completedCount }})</el-radio-button>
              <el-radio-button value="paused">已暂停 ({{ pausedCount }})</el-radio-button>
              <el-radio-button value="failed">失败 ({{ failedCount }})</el-radio-button>
            </el-radio-group>
          </div>
        </div>
      </div>

      <!-- Active Tasks List -->
      <div v-if="filteredTasks.length > 0" class="task-list">
        <div
          v-for="task in filteredTasks"
          :key="task.id"
          class="task-card"
          :class="{
            'card-completed': task.status === 'completed',
            'card-failed': task.status === 'failed',
            'card-retrying': task.status === 'retrying'
          }"
        >
          <div class="task-header">
            <div class="task-names">
              <span class="task-game">{{ task.game_name || formatFallbackGame(task.file_name) }}</span>
              <span class="task-file">
                {{ (task.trainer_name ? task.trainer_name + ' · ' : '') + (task.version_name || task.file_name) }}
              </span>
            </div>
            <div class="task-actions">
              <template v-if="task.status === 'completed'">
                <el-button
                  size="small"
                  type="primary"
                  plain
                  @click="openFolder(task.save_path)"
                >
                  <el-icon><FolderOpened /></el-icon> {{ t('downloads.locateFile') }}
                </el-button>
                <el-button
                  size="small"
                  plain
                  @click="removeDownloadTask(task.id)"
                >
                  <el-icon><Delete /></el-icon> {{ t('downloads.clearRecord') }}
                </el-button>
              </template>
              <template v-else-if="task.status === 'downloading' || task.status === 'retrying'">
                <el-button
                  size="small"
                  plain
                  @click="pauseTask(task.id)"
                >
                  <el-icon><VideoPause /></el-icon> {{ t('downloads.pause') }}
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
              <template v-else-if="task.status === 'paused'">
                <el-button
                  size="small"
                  type="primary"
                  plain
                  @click="resumeTask(task.id)"
                >
                  <el-icon><VideoPlay /></el-icon> {{ t('downloads.resume') }}
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
              <template v-else-if="task.status === 'failed'">
                <el-button
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

          <!-- Error Alert -->
          <div v-if="task.error_msg && (task.status === 'failed' || task.status === 'retrying')" class="task-error-alert">
            <el-icon><WarningFilled /></el-icon>
            <span>{{ task.error_msg }}</span>
          </div>

          <div class="task-footer">
            <span class="task-bytes">
              {{ formatBytes(task.downloaded_bytes || 0) }} / {{ (task.total_bytes && task.total_bytes > 0) ? formatBytes(task.total_bytes) : t('downloads.unknownSize') }}
            </span>
            <span v-if="task.status === 'downloading'" class="task-speed">
              {{ formatBytes(task.speed_bytes_per_sec) }}/s
            </span>
            <span v-else class="task-status-text" :class="`status-${task.status}`">
              {{ t('downloads.statusLabel') }}{{ getStatusLabel(task.status, task.retries) }}
            </span>
          </div>
        </div>
      </div>

      <!-- Empty Active Tasks State -->
      <div v-else class="active-empty-state">
        <div v-if="(store.downloadHistory || []).length > 0" class="active-idle-banner">
          <el-icon class="text-emerald text-lg"><Check /></el-icon>
          <div class="idle-text-wrap">
            <span class="idle-main">{{ t('downloads.noActiveTasksWithHistory') }}</span>
          </div>
        </div>
        <div v-else class="empty-wrap">
          <el-empty :description="t('downloads.emptyTasks')" />
        </div>
      </div>
    </section>

    <!-- 区域二：下载历史 (Download History Section) -->
    <section v-if="viewMode === 'all' || viewMode === 'history'" class="download-section history-section">
      <div class="section-top-bar">
        <div class="section-heading">
          <el-icon class="heading-icon text-amber"><Clock /></el-icon>
          <span class="heading-title">{{ t('downloads.sectionHistory') }}</span>
          <el-tag size="small" type="success" round>{{ (store.downloadHistory || []).length }}</el-tag>
        </div>

        <div class="section-controls">
          <el-input
            v-model="historyFilterKeyword"
            size="small"
            :placeholder="t('downloads.filterHistoryPlaceholder')"
            clearable
            class="history-filter-input"
          >
            <template #prefix>
              <el-icon><Search /></el-icon>
            </template>
          </el-input>

          <el-button
            v-if="(store.downloadHistory || []).length > 0"
            size="small"
            type="danger"
            plain
            @click="promptClearHistory"
          >
            <el-icon><Delete /></el-icon> {{ t('downloads.clearAllHistory') }}
          </el-button>
        </div>
      </div>

      <!-- History Table -->
      <div v-if="filteredHistory.length > 0" class="history-table-container">
        <el-table :data="filteredHistory" stripe style="width: 100%" class="custom-table">
          <!-- Game & Trainer Info -->
          <el-table-column :label="t('downloads.historyGame')" min-width="180">
            <template #default="{ row }">
              <div class="history-game-cell">
                <img :src="row.cover_url || DEFAULT_GAME_COVER" class="history-thumb" />
                <div class="history-game-names">
                  <span class="game-title" :title="row.game_name">{{ row.game_name }}</span>
                  <span class="trainer-sub" :title="row.trainer_name">{{ row.trainer_name || 'FLiNG 修改器' }}</span>
                </div>
              </div>
            </template>
          </el-table-column>

          <!-- Version Name -->
          <el-table-column :label="t('downloads.historyTrainer')" min-width="140">
            <template #default="{ row }">
              <el-tag size="small" type="info" effect="plain">{{ row.version_name || '-' }}</el-tag>
            </template>
          </el-table-column>

          <!-- File Name -->
          <el-table-column prop="file_name" :label="t('downloads.historyFile')" min-width="170" show-overflow-tooltip />

          <!-- File Size -->
          <el-table-column :label="t('downloads.historySize')" width="100">
            <template #default="{ row }">
              {{ formatBytes(row.file_size) }}
            </template>
          </el-table-column>

          <!-- File Exists Status -->
          <el-table-column :label="t('downloads.fileStatus')" width="100">
            <template #default="{ row }">
              <el-tag v-if="row.file_exists !== false" type="success" size="small">{{ t('downloads.fileExists') }}</el-tag>
              <el-tag v-else type="danger" size="small">{{ t('downloads.fileMissing') }}</el-tag>
            </template>
          </el-table-column>

          <!-- SHA-256 -->
          <el-table-column label="SHA-256" width="160">
            <template #default="{ row }">
              <span class="hash-short" :title="row.sha256">{{ (row.sha256 || '').substring(0, 10) }}...</span>
              <el-button link type="primary" size="small" @click="copy(row.sha256)">{{ t('security.copy') }}</el-button>
            </template>
          </el-table-column>

          <!-- Completed Time -->
          <el-table-column :label="t('downloads.historyTime')" width="150">
            <template #default="{ row }">
              {{ formatDate(row.completed_at) }}
            </template>
          </el-table-column>

          <!-- Actions -->
          <el-table-column :label="t('downloads.historyAction')" width="270" fixed="right">
            <template #default="{ row }">
              <el-button size="small" link type="primary" @click="viewHistoryDetail(row)">
                <el-icon><InfoFilled /></el-icon> {{ t('downloads.viewDetail') }}
              </el-button>
              <el-button size="small" link type="primary" @click="openFolder(row.save_path)">
                <el-icon><FolderOpened /></el-icon> {{ t('downloads.locateFile') }}
              </el-button>
              <el-button size="small" link type="warning" @click="redownload(row)">
                <el-icon><RefreshRight /></el-icon> {{ t('downloads.redownload') }}
              </el-button>
              <el-button size="small" link type="danger" @click="promptDeleteHistory(row)">
                <el-icon><Delete /></el-icon> {{ t('downloads.deleteHistory') }}
              </el-button>
            </template>
          </el-table-column>
        </el-table>
      </div>

      <!-- Empty History State -->
      <div v-else class="empty-wrap">
        <el-empty :description="historyFilterKeyword ? '未检索到匹配的下载历史记录' : t('downloads.emptyHistory')" />
      </div>
    </section>

    <!-- 删除下载历史记录对话框 (支持区分物理文件删除) -->
    <el-dialog
      v-model="deleteDialogVisible"
      title="删除下载记录"
      width="440px"
      append-to-body
    >
      <div class="delete-dialog-content">
        <p class="delete-warning-text">
          确定从下载历史中删除「<strong>{{ selectedHistoryItem?.game_name }}</strong>」的记录吗？
        </p>
        <div class="delete-tip-box">
          <el-icon class="text-amber"><InfoFilled /></el-icon>
          <span>{{ t('downloads.deleteHistoryTip') }}</span>
        </div>
        <div class="delete-checkbox-row">
          <el-checkbox v-model="deletePhysicalFile">
            {{ t('downloads.deletePhysicalFileCheck') }}
          </el-checkbox>
        </div>
      </div>
      <template #footer>
        <el-button @click="deleteDialogVisible = false">{{ t('library.cancel') }}</el-button>
        <el-button type="danger" :loading="deletingHistory" @click="confirmDeleteHistory">
          {{ t('library.delete') }}
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onActivated } from 'vue'
import {
  VideoPause,
  VideoPlay,
  Delete,
  RefreshRight,
  FolderOpened,
  WarningFilled,
  Download,
  Clock,
  Search,
  Check,
  InfoFilled
} from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useAppStore } from '../stores/appStore'
import { useSearchStore } from '../stores/searchStore'
import { api } from '../api/wails'
import { t } from '../i18n'
import { DEFAULT_GAME_COVER } from '../utils/assets'
import type { DownloadHistory, DownloadTask, SearchResult } from '../types'

defineOptions({ name: 'DownloadsView' })

const store = useAppStore()
const searchStore = useSearchStore()

const viewMode = ref<'all' | 'active' | 'history'>('all')
const statusFilter = ref('all')
const historyFilterKeyword = ref('')

// 删除对话框状态
const deleteDialogVisible = ref(false)
const selectedHistoryItem = ref<DownloadHistory | null>(null)
const deletePhysicalFile = ref(false)
const deletingHistory = ref(false)

const allTasks = computed(() => Array.isArray(store.downloadTasks) ? store.downloadTasks : [])
const allHistory = computed(() => Array.isArray(store.downloadHistory) ? store.downloadHistory : [])

const downloadingCount = computed(() => allTasks.value.filter(t => t && (t.status === 'downloading' || t.status === 'retrying')).length)
const completedCount = computed(() => allTasks.value.filter(t => t && t.status === 'completed').length)
const pausedCount = computed(() => allTasks.value.filter(t => t && t.status === 'paused').length)
const failedCount = computed(() => allTasks.value.filter(t => t && t.status === 'failed').length)
const hasCompletedTasks = computed(() => completedCount.value > 0)

const filteredTasks = computed(() => {
  const tasks = allTasks.value
  if (statusFilter.value === 'all') return tasks
  if (statusFilter.value === 'downloading') {
    return tasks.filter(t => t && (t.status === 'downloading' || t.status === 'retrying'))
  }
  return tasks.filter(t => t && t.status === statusFilter.value)
})

const filteredHistory = computed(() => {
  const history = allHistory.value
  const kw = historyFilterKeyword.value.trim().toLowerCase()
  if (!kw) return history
  return history.filter(h =>
    h && (
      (h.game_name && h.game_name.toLowerCase().includes(kw)) ||
      (h.trainer_name && h.trainer_name.toLowerCase().includes(kw)) ||
      (h.file_name && h.file_name.toLowerCase().includes(kw)) ||
      (h.version_name && h.version_name.toLowerCase().includes(kw))
    )
  )
})

onMounted(async () => {
  await Promise.all([
    store.refreshDownloads(),
    store.refreshHistory()
  ])
})

onActivated(async () => {
  await Promise.all([
    store.refreshDownloads(),
    store.refreshHistory()
  ])
})

function getTaskProgress(task?: DownloadTask): number {
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

async function removeDownloadTask(id: number) {
  try {
    await api.deleteDownloadTask(id)
    await store.refreshDownloads()
    ElMessage.success(t('downloads.clearRecord') + '成功')
  } catch (e: any) {
    ElMessage.error('清除任务记录失败: ' + (e?.message || e))
  }
}

async function clearCompletedTasks() {
  const completed = allTasks.value.filter(t => t && t.status === 'completed')
  if (completed.length === 0) return
  try {
    for (const t of completed) {
      await api.deleteDownloadTask(t.id)
    }
    await store.refreshDownloads()
    ElMessage.success('已清理所有已完成任务记录')
  } catch (e: any) {
    ElMessage.error('清理失败: ' + (e?.message || e))
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
    viewMode.value = 'all'
    await store.refreshDownloads()
  } catch (e: any) {
    ElMessage.error('创建重新下载任务失败: ' + (e?.message || e))
  }
}

function viewHistoryDetail(row: DownloadHistory) {
  if (!row) return
  const converted: SearchResult = {
    game_id: row.game_id || 0,
    slug: '',
    game_name_en: row.game_name || '',
    game_name_cn: row.game_name || '',
    cover_url: row.cover_url || '',
    source_id: 'fling',
    trainer_title: (row.trainer_name || row.game_name || '') + ' 修改器',
    detail_url: row.detail_url || '',
    version_name: row.version_name || '',
    updated_at: row.completed_at ? String(row.completed_at) : '',
    is_downloaded: row.file_exists !== false,
    is_online_only: false,
    is_favorite: false,
    score: 100
  }
  searchStore.openDetail(converted)
}

function promptDeleteHistory(item: DownloadHistory) {
  selectedHistoryItem.value = item
  deletePhysicalFile.value = false
  deleteDialogVisible.value = true
}

async function confirmDeleteHistory() {
  if (!selectedHistoryItem.value) return
  deletingHistory.value = true
  try {
    await api.deleteDownloadHistory(selectedHistoryItem.value.id, deletePhysicalFile.value)
    if (deletePhysicalFile.value) {
      ElMessage.success('已删除下载记录及本地物理文件')
      await store.refreshLocalTrainers()
    } else {
      ElMessage.success('已从下载历史中移除记录（本地文件已保留）')
    }
    await store.refreshHistory()
    deleteDialogVisible.value = false
  } catch (e: any) {
    ElMessage.error('删除下载记录失败: ' + (e?.message || e))
  } finally {
    deletingHistory.value = false
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
    await store.refreshHistory()
  } catch (e: any) {
    if (e !== 'cancel') {
      ElMessage.error('清空历史记录失败: ' + (e?.message || e))
    }
  }
}

function formatFallbackGame(fileName?: string): string {
  if (!fileName) return '未知游戏'
  return fileName.replace(/\.(exe|zip|rar)$/i, '').replace(/plus\s*\d+.*$/i, '').trim() || fileName
}

function formatDate(dateStr?: string | Date): string {
  if (!dateStr) return ''
  try {
    const d = new Date(dateStr)
    if (isNaN(d.getTime())) return String(dateStr)
    const y = d.getFullYear()
    const m = (d.getMonth() + 1).toString().padStart(2, '0')
    const day = d.getDate().toString().padStart(2, '0')
    const h = d.getHours().toString().padStart(2, '0')
    const min = d.getMinutes().toString().padStart(2, '0')
    return `${y}-${m}-${day} ${h}:${min}`
  } catch {
    return String(dateStr)
  }
}

function copy(text: string) {
  if (!text) return
  navigator.clipboard.writeText(text)
  ElMessage.success(t('security.copied'))
}

function formatBytes(bytes?: number): string {
  if (!bytes || bytes <= 0 || isNaN(bytes)) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return (bytes / Math.pow(k, i)).toFixed(2) + ' ' + (sizes[i] || 'B')
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
.downloads-view {
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.view-mode-bar {
  display: flex;
  justify-content: flex-start;
  margin-bottom: 4px;
}

.download-section {
  background: var(--bg-card);
  border: 1px solid var(--border-color);
  border-radius: 10px;
  padding: 18px 20px;
  box-shadow: var(--card-shadow);
}

.section-top-bar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
  flex-wrap: wrap;
  gap: 12px;
}

.section-heading {
  display: flex;
  align-items: center;
  gap: 8px;
}

.heading-icon {
  font-size: 18px;
}

.heading-title {
  font-size: 16px;
  font-weight: 700;
  color: var(--text-main);
}

.section-controls {
  display: flex;
  align-items: center;
  gap: 12px;
}

.batch-btns {
  display: flex;
  gap: 8px;
}

.history-filter-input {
  width: 260px;
}

.task-list {
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.task-card {
  background: var(--bg-card-subtle);
  border: 1px solid var(--border-color);
  border-radius: 8px;
  padding: 16px;
  transition: all 0.2s ease;
}

.task-card.card-retrying {
  border-color: #f59e0b;
}

.task-card.card-failed {
  border-color: #ef4444;
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

.task-card.card-completed {
  border-color: rgba(16, 185, 129, 0.4);
}

.status-completed {
  color: #10b981;
}

.status-retrying {
  color: #f59e0b;
}

.active-idle-banner {
  display: flex;
  align-items: center;
  gap: 12px;
  background: rgba(16, 185, 129, 0.08);
  border: 1px dashed rgba(16, 185, 129, 0.35);
  padding: 14px 18px;
  border-radius: 8px;
  color: var(--text-secondary);
}

.idle-main {
  font-size: 13px;
  color: var(--text-main);
}

.history-game-cell {
  display: flex;
  align-items: center;
  gap: 10px;
}

.history-thumb {
  width: 36px;
  height: 36px;
  border-radius: 4px;
  object-fit: cover;
  flex-shrink: 0;
}

.history-game-names {
  display: flex;
  flex-direction: column;
  gap: 2px;
  overflow: hidden;
}

.game-title {
  font-weight: 600;
  font-size: 13px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.trainer-sub {
  font-size: 12px;
  color: var(--text-muted);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.hash-short {
  font-family: monospace;
  font-size: 12px;
  color: var(--text-muted);
  margin-right: 6px;
}

.empty-wrap {
  padding: 30px 0;
}

.delete-dialog-content {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.delete-warning-text {
  margin: 0;
  font-size: 14px;
  color: var(--text-main);
}

.delete-tip-box {
  display: flex;
  align-items: center;
  gap: 8px;
  background: rgba(245, 158, 11, 0.1);
  border: 1px solid rgba(245, 158, 11, 0.25);
  color: #d97706;
  padding: 8px 12px;
  border-radius: 6px;
  font-size: 12px;
}

.delete-checkbox-row {
  margin-top: 4px;
}
</style>
