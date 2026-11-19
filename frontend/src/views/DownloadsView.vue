<template>
  <div class="downloads-view">
    <el-tabs v-model="activeTab" class="custom-tabs">
      <!-- 正在下载 / 活动任务 -->
      <el-tab-pane :label="t('downloads.tabActive', { count: store.downloadTasks.length })" name="active">
        <div v-if="store.downloadTasks.length > 0" class="task-list">
          <div
            v-for="task in store.downloadTasks"
            :key="task.id"
            class="task-card"
            :class="{ 'card-failed': task.status === 'failed', 'card-retrying': task.status === 'retrying' }"
          >
            <div class="task-header">
              <div class="task-names">
                <span class="task-game">{{ task.game_name || t('search.drawerTitle') }}</span>
                <span class="task-file">{{ task.file_name }}</span>
              </div>
              <div class="task-actions">
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
                  v-if="task.status === 'completed'"
                  size="small"
                  type="success"
                  plain
                  @click="openFolder(task.save_path)"
                >
                  <el-icon><FolderOpened /></el-icon> {{ t('downloads.locateFile') }}
                </el-button>
                <el-button
                  size="small"
                  type="danger"
                  plain
                  @click="cancelTask(task.id)"
                >
                  <el-icon><Delete /></el-icon> {{ task.status === 'completed' ? t('downloads.clearRecord') : t('downloads.cancel') }}
                </el-button>
              </div>
            </div>

            <div class="task-progress-wrap">
              <el-progress
                :percentage="task.total_bytes > 0 ? Number(task.progress.toFixed(1)) : 0"
                :indeterminate="task.status === 'downloading' && task.total_bytes <= 0"
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
                {{ formatBytes(task.downloaded_bytes) }} / {{ task.total_bytes > 0 ? formatBytes(task.total_bytes) : t('downloads.unknownSize') }}
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

        <div v-else class="empty-wrap">
          <el-empty :description="t('downloads.emptyTasks')" />
        </div>
      </el-tab-pane>

      <!-- 下载历史 -->
      <el-tab-pane :label="t('downloads.tabHistory')" name="history">
        <el-table :data="historyList" stripe style="width: 100%" class="custom-table">
          <el-table-column prop="game_name" :label="t('downloads.historyGame')" width="220" />
          <el-table-column prop="file_name" :label="t('downloads.historyFile')" />
          <el-table-column :label="t('downloads.historySize')" width="110">
            <template #default="{ row }">
              {{ formatBytes(row.file_size) }}
            </template>
          </el-table-column>
          <el-table-column label="SHA-256" width="220">
            <template #default="{ row }">
              <span class="hash-short" :title="row.sha256">{{ (row.sha256 || '').substring(0, 16) }}...</span>
              <el-button link type="primary" size="small" @click="copy(row.sha256)">{{ t('security.copy') }}</el-button>
            </template>
          </el-table-column>
          <el-table-column :label="t('downloads.historyAction')" width="140">
            <template #default="{ row }">
              <el-button size="small" link type="primary" @click="openFolder(row.save_path)">
                {{ t('downloads.locateFile') }}
              </el-button>
            </template>
          </el-table-column>
        </el-table>
      </el-tab-pane>
    </el-tabs>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { VideoPause, VideoPlay, Delete, RefreshRight, FolderOpened, WarningFilled } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import { useAppStore } from '../stores/appStore'
import { api } from '../api/wails'
import { t } from '../i18n'
import type { DownloadHistory } from '../types'

defineOptions({ name: 'DownloadsView' })

const store = useAppStore()
const activeTab = ref('active')
const historyList = ref<DownloadHistory[]>([])

onMounted(async () => {
  await store.refreshDownloads()
  loadHistory()
})

async function loadHistory() {
  try {
    historyList.value = await api.listDownloadHistory(50)
  } catch (e: any) {
    console.error('Failed to load history', e)
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
.task-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}
.task-card {
  background: rgba(30, 41, 59, 0.5);
  border: 1px solid rgba(255, 255, 255, 0.08);
  border-radius: 8px;
  padding: 16px;
  transition: all 0.2s ease;
}
.card-failed {
  border-color: rgba(239, 68, 68, 0.4);
  background: rgba(30, 41, 59, 0.7);
}
.card-retrying {
  border-color: rgba(234, 179, 8, 0.4);
}
.task-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 12px;
}
.task-names {
  display: flex;
  flex-direction: column;
  gap: 4px;
}
.task-game {
  font-size: 15px;
  font-weight: 600;
  color: #f8fafc;
}
.task-file {
  font-size: 12px;
  color: #94a3b8;
}
.task-actions {
  display: flex;
  gap: 8px;
  align-items: center;
}
.task-progress-wrap {
  margin-bottom: 8px;
}
.task-error-alert {
  margin-bottom: 8px;
  padding: 6px 12px;
  background: rgba(239, 68, 68, 0.12);
  border: 1px solid rgba(239, 68, 68, 0.25);
  border-radius: 4px;
  color: #f87171;
  font-size: 12px;
  display: flex;
  align-items: center;
  gap: 6px;
}
.card-retrying .task-error-alert {
  background: rgba(234, 179, 8, 0.12);
  border-color: rgba(234, 179, 8, 0.25);
  color: #facc15;
}
.task-footer {
  display: flex;
  justify-content: space-between;
  font-size: 12px;
  color: #94a3b8;
}
.task-speed {
  color: #38bdf8;
  font-weight: 600;
}
.task-status-text {
  font-weight: 500;
}
.status-failed {
  color: #ef4444;
}
.status-retrying {
  color: #eab308;
}
.status-completed {
  color: #22c55e;
}
.status-downloading {
  color: #38bdf8;
}
.empty-wrap {
  padding: 60px 0;
}
.hash-short {
  font-family: monospace;
  font-size: 11px;
  margin-right: 6px;
}
</style>
