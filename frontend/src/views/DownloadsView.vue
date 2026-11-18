<template>
  <div class="downloads-view">
    <el-tabs v-model="activeTab" class="custom-tabs">
      <!-- 正在下载 -->
      <el-tab-pane :label="`下载任务 (${store.downloadTasks.length})`" name="active">
        <div v-if="store.downloadTasks.length > 0" class="task-list">
          <div
            v-for="task in store.downloadTasks"
            :key="task.id"
            class="task-card"
          >
            <div class="task-header">
              <div class="task-names">
                <span class="task-game">{{ task.game_name || '修改器下载' }}</span>
                <span class="task-file">{{ task.file_name }}</span>
              </div>
              <div class="task-actions">
                <el-button
                  v-if="task.status === 'downloading'"
                  size="small"
                  plain
                  @click="pauseTask(task.id)"
                >
                  <el-icon><VideoPause /></el-icon> 暂停
                </el-button>
                <el-button
                  v-else-if="task.status === 'paused' || task.status === 'failed'"
                  size="small"
                  type="primary"
                  plain
                  @click="resumeTask(task.id)"
                >
                  <el-icon><VideoPlay /></el-icon> 继续
                </el-button>
                <el-button
                  size="small"
                  type="danger"
                  plain
                  @click="cancelTask(task.id)"
                >
                  <el-icon><Delete /></el-icon> 取消
                </el-button>
              </div>
            </div>

            <div class="task-progress-wrap">
              <el-progress
                :percentage="Number(task.progress.toFixed(1))"
                :status="task.status === 'completed' ? 'success' : task.status === 'failed' ? 'exception' : ''"
                :stroke-width="8"
              />
            </div>

            <div class="task-footer">
              <span class="task-bytes">
                {{ formatBytes(task.downloaded_bytes) }} / {{ formatBytes(task.total_bytes) }}
              </span>
              <span v-if="task.status === 'downloading'" class="task-speed">
                {{ formatBytes(task.speed_bytes_per_sec) }}/s
              </span>
              <span v-else class="task-status-text">
                状态: {{ getStatusLabel(task.status) }}
              </span>
            </div>
          </div>
        </div>

        <div v-else class="empty-wrap">
          <el-empty description="当前暂无下载任务" />
        </div>
      </el-tab-pane>

      <!-- 下载历史 -->
      <el-tab-pane label="下载历史" name="history">
        <el-table :data="historyList" stripe style="width: 100%" class="custom-table">
          <el-table-column prop="game_name" label="游戏名称" width="220" />
          <el-table-column prop="file_name" label="文件名称" />
          <el-table-column label="大小" width="110">
            <template #default="{ row }">
              {{ formatBytes(row.file_size) }}
            </template>
          </el-table-column>
          <el-table-column label="SHA-256" width="220">
            <template #default="{ row }">
              <span class="hash-short" :title="row.sha256">{{ row.sha256.substring(0, 16) }}...</span>
              <el-button link type="primary" size="small" @click="copy(row.sha256)">复制</el-button>
            </template>
          </el-table-column>
          <el-table-column label="操作" width="140">
            <template #default="{ row }">
              <el-button size="small" link type="primary" @click="openFolder(row.save_path)">
                定位文件
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
import { VideoPause, VideoPlay, Delete } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import { useAppStore } from '../stores/appStore'
import { api } from '../api/wails'
import type { DownloadHistory } from '../types'

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
    ElMessage.error('暂停失败: ' + (e.message || e))
  }
}

async function resumeTask(id: number) {
  try {
    await api.resumeDownload(id)
    await store.refreshDownloads()
  } catch (e: any) {
    ElMessage.error('恢复失败: ' + (e.message || e))
  }
}

async function cancelTask(id: number) {
  try {
    await api.cancelDownload(id)
    await store.refreshDownloads()
    ElMessage.success('已取消下载')
  } catch (e: any) {
    ElMessage.error('取消失败: ' + (e.message || e))
  }
}

async function openFolder(path: string) {
  try {
    await api.openContainingFolder(path)
  } catch (e: any) {
    ElMessage.error('定位失败: ' + (e.message || e))
  }
}

function copy(text: string) {
  navigator.clipboard.writeText(text)
  ElMessage.success('已复制')
}

function formatBytes(bytes: number): string {
  if (!bytes || bytes <= 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return (bytes / Math.pow(k, i)).toFixed(2) + ' ' + sizes[i]
}

function getStatusLabel(status: string): string {
  switch (status) {
    case 'downloading': return '正在下载'
    case 'paused': return '已暂停'
    case 'completed': return '已完成'
    case 'failed': return '下载失败'
    default: return '等待中'
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
.task-progress-wrap {
  margin-bottom: 8px;
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
.empty-wrap {
  padding: 60px 0;
}
.hash-short {
  font-family: monospace;
  font-size: 11px;
  margin-right: 6px;
}
</style>

