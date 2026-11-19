<template>
  <div class="library-view">
    <!-- Action Bar -->
    <div class="lib-actions-bar">
      <div class="left-stats">
        <span class="lib-count">{{ t('library.managedCount') }} <strong>{{ store.localTrainers.length }}</strong> {{ t('library.managedUnit') }}</span>
      </div>
      <div class="right-buttons">
        <el-button :loading="scanningSteam" plain @click="scanSteam">
          <el-icon><Monitor /></el-icon> {{ t('library.scanSteam') }}
        </el-button>
        <el-button type="primary" plain @click="selectAndImportExe">
          <el-icon><Plus /></el-icon> {{ t('library.importExe') }}
        </el-button>
      </div>
    </div>

    <!-- Trainers Grid -->
    <div v-if="store.localTrainers.length > 0" class="local-grid">
      <div
        v-for="item in store.localTrainers"
        :key="item.local_trainer.id"
        class="local-card"
        :class="{ 'card-missing': item.file_exists === false, 'card-running': item.is_running }"
      >
        <div class="local-card-thumb">
          <img :src="item.game?.cover_url || DEFAULT_GAME_COVER" />
          <div class="badge-group-top">
            <el-tag
              :type="item.local_trainer.is_custom ? 'warning' : 'success'"
              size="small"
              class="source-badge"
            >
              {{ item.local_trainer.source_name }}
            </el-tag>
            <!-- 运行状态徽章 -->
            <el-tag
              v-if="item.is_running"
              type="danger"
              effect="dark"
              size="small"
              class="status-badge"
            >
              {{ t('library.statusRunning') }}
            </el-tag>
            <!-- 丢失状态徽章 -->
            <el-tag
              v-else-if="item.file_exists === false"
              type="danger"
              size="small"
              class="status-badge"
            >
              {{ t('library.fileMissing') }}
            </el-tag>
          </div>
        </div>

        <div class="local-card-info">
          <div class="game-name">{{ item.game?.name_cn || item.game?.name_en || t('library.unassociatedGame') }}</div>
          <div class="file-name" :title="item.local_trainer.file_path">
            {{ getFileName(item.local_trainer.file_path) }}
          </div>

          <div v-if="item.target_game_version" class="target-version-tag">
            <el-tag size="small" type="info" effect="plain">
              {{ t('library.targetGameVersion', { version: item.target_game_version }) }}
            </el-tag>
          </div>

          <div class="stats-row">
            <span>{{ t('library.runCount') }} {{ item.local_trainer.run_count }}</span>
            <span>{{ t('library.fileSize') }} {{ formatBytes(item.local_trainer.file_size) }}</span>
          </div>

          <div class="card-btns">
            <!-- 运行状态下提供停止按钮，否则提供安全启动 -->
            <el-button
              v-if="item.is_running"
              type="danger"
              size="small"
              @click="confirmStop(item.local_trainer.id)"
            >
              <el-icon><VideoPause /></el-icon> {{ t('library.stopTrainer') }}
            </el-button>
            <el-button
              v-else-if="item.file_exists !== false"
              type="primary"
              size="small"
              @click="openLaunchModal(item.local_trainer.id, item)"
            >
              <el-icon><VideoPlay /></el-icon> {{ t('library.launchBtn') }}
            </el-button>

            <!-- 文件丢失时提供重新下载与重新定位按钮 -->
            <template v-else>
              <el-button
                type="warning"
                size="small"
                @click="promptMissingFileRecovery(item)"
              >
                <el-icon><RefreshRight /></el-icon> 重新下载
              </el-button>
              <el-button
                size="small"
                plain
                @click="relocateTrainerFile(item.local_trainer.id)"
              >
                <el-icon><Connection /></el-icon> {{ t('library.relocate') }}
              </el-button>
            </template>

            <el-button
              v-if="item.file_exists !== false"
              size="small"
              plain
              @click="openFolder(item.local_trainer.file_path)"
            >
              {{ t('library.folderBtn') }}
            </el-button>

            <!-- 校验完整性按钮 -->
            <el-button
              size="small"
              plain
              :title="t('library.verifyIntegrity')"
              @click="checkIntegrity(item.local_trainer.id)"
            >
              <el-icon><Check /></el-icon>
            </el-button>

            <!-- 删除按钮 -->
            <el-button
              size="small"
              type="danger"
              link
              @click="promptDelete(item.local_trainer.id)"
            >
              <el-icon><Delete /></el-icon>
            </el-button>
          </div>
        </div>
      </div>
    </div>

    <div v-else class="empty-wrap">
      <el-empty :description="t('library.emptyLibrary')">
        <template #extra>
          <p class="empty-tip">{{ t('library.emptyLibraryTip') }}</p>
        </template>
      </el-empty>
    </div>

    <!-- 完整性校验结果弹窗 -->
    <el-dialog
      v-model="integrityDialogVisible"
      :title="t('library.integrityTitle')"
      width="560px"
      append-to-body
    >
      <div v-loading="checkingIntegrity" class="integrity-dialog-body">
        <div v-if="integrityResult" class="integrity-content">
          <el-alert
            :type="integrityResult.is_valid ? 'success' : 'error'"
            :title="integrityResult.is_valid ? t('library.integrityMatch') : t('library.integrityMismatch')"
            :closable="false"
            show-icon
            class="mb-4"
          />
          <div class="hash-field">
            <span class="hash-label">{{ t('library.expectedHash') }}</span>
            <code class="hash-code">{{ integrityResult.expected_sha256 }}</code>
          </div>
          <div class="hash-field mt-2">
            <span class="hash-label">{{ t('library.actualHash') }}</span>
            <code class="hash-code">{{ integrityResult.actual_sha256 }}</code>
          </div>
        </div>
      </div>
      <template #footer>
        <el-button @click="integrityDialogVisible = false">{{ t('library.cancel') }}</el-button>
      </template>
    </el-dialog>

    <!-- 删除确认弹窗 (包含物理删除勾选) -->
    <el-dialog
      v-model="deleteDialogVisible"
      :title="t('library.deleteDialogTitle')"
      width="440px"
      append-to-body
    >
      <div class="delete-dialog-content">
        <p class="delete-warning-text">{{ t('library.deleteConfirm') }}</p>
        <el-checkbox v-model="deletePhysicalFile">
          {{ t('library.deleteFileCheck') }}
        </el-checkbox>
      </div>
      <template #footer>
        <el-button @click="deleteDialogVisible = false">{{ t('library.cancel') }}</el-button>
        <el-button type="danger" @click="confirmDelete">{{ t('library.delete') }}</el-button>
      </template>
    </el-dialog>

    <!-- Security Modal Reference -->
    <SecurityModal ref="securityModalRef" @missing-file="onMissingFileModal" />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onActivated } from 'vue'
import { Monitor, Plus, VideoPlay, VideoPause, Delete, Check, Connection, RefreshRight } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useAppStore } from '../stores/appStore'
import { api } from '../api/wails'
import { t } from '../i18n'
import { DEFAULT_GAME_COVER } from '../utils/assets'
import SecurityModal from '../components/SecurityModal.vue'
import type { IntegrityResult } from '../types'

defineOptions({ name: 'LibraryView' })

const store = useAppStore()
const scanningSteam = ref(false)
const securityModalRef = ref<InstanceType<typeof SecurityModal> | null>(null)

// 完整性校验状态
const integrityDialogVisible = ref(false)
const checkingIntegrity = ref(false)
const integrityResult = ref<IntegrityResult | null>(null)

// 删除对话框状态
const deleteDialogVisible = ref(false)
const deletingTrainerID = ref<number | null>(null)
const deletePhysicalFile = ref(false)

onMounted(() => {
  store.refreshLocalTrainers()
})

onActivated(() => {
  store.refreshLocalTrainers()
})

async function openLaunchModal(id: number, item?: any) {
  try {
    const card = await api.getSecurityCard(id)
    securityModalRef.value?.open(card)
  } catch (e: any) {
    const msg = e?.message || String(e)
    if (msg.includes('不存在') || msg.includes('not exist') || msg.includes('cannot find')) {
      promptMissingFileRecovery(item || { local_trainer: { id } })
    } else {
      ElMessage.error(msg)
    }
  }
}

async function promptMissingFileRecovery(item: any) {
  try {
    await ElMessageBox.confirm(
      '修改器文件不存在，可能已被删除或移动，是否重新下载？',
      '文件缺失提示',
      {
        confirmButtonText: '重新下载',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )
    await api.redownloadLocalTrainer(item.local_trainer.id)
    ElMessage.success('已创建重新下载任务，下载完成后将自动更新本地修改器')
    await store.refreshDownloads()
  } catch (e: any) {
    if (e !== 'cancel') {
      ElMessage.error('创建重新下载任务失败: ' + (e?.message || e))
    }
  }
}

function onMissingFileModal(card: any) {
  promptMissingFileRecovery({
    local_trainer: { id: card.trainer_id }
  })
}

async function confirmStop(id: number) {
  try {
    await ElMessageBox.confirm(t('library.stopConfirm'), t('library.stopTrainer'), {
      confirmButtonText: t('library.stopTrainer'),
      cancelButtonText: t('library.cancel'),
      type: 'warning'
    })
    await api.stopTrainer(id)
    ElMessage.success(t('library.stopSuccess'))
    await store.refreshLocalTrainers()
  } catch (e: any) {
    if (e !== 'cancel') {
      ElMessage.error(t('library.stopFailed') + (e?.message || e))
    }
  }
}

async function openFolder(path: string) {
  try {
    await api.openContainingFolder(path)
  } catch (e: any) {
    ElMessage.error(e?.message || String(e))
  }
}

async function scanSteam() {
  scanningSteam.value = true
  try {
    const count = await api.scanSteamGames()
    ElMessage.success(t('library.scanSuccess', { count }))
    await store.refreshLocalTrainers()
  } catch (e: any) {
    ElMessage.error(t('library.scanFailed') + (e.message || e))
  } finally {
    scanningSteam.value = false
  }
}

async function selectAndImportExe() {
  try {
    const selected = await api.selectTrainerExecutable()
    if (selected) {
      await api.importCustomTrainer(0, selected.trim())
      ElMessage.success(t('settings.savedSuccess'))
      await store.refreshLocalTrainers()
    }
  } catch (e: any) {
    ElMessage.error('导入修改器失败: ' + (e?.message || e))
  }
}

async function relocateTrainerFile(id: number) {
  try {
    const newPath = await api.selectTrainerExecutable()
    if (!newPath) return
    await api.relocateTrainer(id, newPath)
    ElMessage.success(t('library.relocateSuccess'))
    await store.refreshLocalTrainers()
  } catch (e: any) {
    ElMessage.error(t('library.relocateFailed') + (e?.message || e))
  }
}

async function checkIntegrity(id: number) {
  integrityDialogVisible.value = true
  checkingIntegrity.value = true
  integrityResult.value = null
  try {
    integrityResult.value = await api.verifyTrainerIntegrity(id)
  } catch (e: any) {
    ElMessage.error('校验失败: ' + (e?.message || e))
    integrityDialogVisible.value = false
  } finally {
    checkingIntegrity.value = false
  }
}

function promptDelete(id: number) {
  deletingTrainerID.value = id
  deletePhysicalFile.value = false
  deleteDialogVisible.value = true
}

async function confirmDelete() {
  if (!deletingTrainerID.value) return
  try {
    await api.deleteLocalTrainer(deletingTrainerID.value, deletePhysicalFile.value)
    ElMessage.success(t('settings.savedSuccess'))
    deleteDialogVisible.value = false
    await store.refreshLocalTrainers()
  } catch (e: any) {
    ElMessage.error(e?.message || String(e))
  }
}

function getFileName(path: string): string {
  if (!path) return ''
  const parts = path.split(/[/\\]/)
  return parts[parts.length - 1]
}

function formatBytes(bytes: number): string {
  if (!bytes || bytes <= 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return (bytes / Math.pow(k, i)).toFixed(1) + ' ' + sizes[i]
}
</script>

<style scoped>
.lib-actions-bar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  background: var(--bg-card);
  padding: 12px 16px;
  border-radius: 8px;
  border: 1px solid var(--border-color);
  margin-bottom: 20px;
}
.lib-count {
  font-size: 14px;
  color: var(--text-muted);
}
.lib-count strong {
  color: var(--accent-cyan);
}
.right-buttons {
  display: flex;
  gap: 10px;
}
.local-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  gap: 16px;
}
.local-card {
  background: var(--bg-card);
  border: 1px solid var(--border-color);
  border-radius: 8px;
  overflow: hidden;
  transition: all 0.2s ease;
  box-shadow: var(--card-shadow);
}
.local-card:hover {
  transform: translateY(-2px);
  border-color: var(--border-hover);
  background: var(--bg-card-hover);
}
.local-card.card-missing {
  border-color: rgba(239, 68, 68, 0.4);
}
.local-card.card-running {
  border-color: rgba(239, 68, 68, 0.6);
  box-shadow: 0 0 10px rgba(239, 68, 68, 0.2);
}
.local-card-thumb {
  position: relative;
  height: 110px;
  background: var(--bg-card-subtle);
}
.local-card-thumb img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}
.badge-group-top {
  position: absolute;
  top: 8px;
  right: 8px;
  display: flex;
  gap: 6px;
}
.local-card-info {
  padding: 14px;
}
.game-name {
  font-size: 15px;
  font-weight: 600;
  color: var(--text-main);
  margin-bottom: 4px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.file-name {
  font-size: 12px;
  color: var(--text-muted);
  margin-bottom: 8px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.target-version-tag {
  margin-bottom: 10px;
}
.stats-row {
  display: flex;
  justify-content: space-between;
  font-size: 11px;
  color: var(--text-gray);
  margin-bottom: 14px;
}
.card-btns {
  display: flex;
  align-items: center;
  gap: 8px;
}
.empty-wrap {
  padding: 60px 0;
}
.empty-tip {
  font-size: 12px;
  color: var(--text-muted);
  margin-top: 6px;
}
.integrity-dialog-body {
  min-height: 100px;
}
.hash-field {
  display: flex;
  flex-direction: column;
  gap: 4px;
}
.hash-label {
  font-size: 12px;
  color: var(--text-muted);
}
.hash-code {
  font-family: monospace;
  font-size: 11px;
  background: var(--code-bg);
  color: var(--code-text);
  padding: 6px 8px;
  border-radius: 4px;
  word-break: break-all;
  border: 1px solid var(--border-color);
}
.delete-dialog-content {
  display: flex;
  flex-direction: column;
  gap: 16px;
}
.delete-warning-text {
  margin: 0;
  font-size: 14px;
  color: var(--text-main);
}
.mb-4 {
  margin-bottom: 16px;
}
.mt-2 {
  margin-top: 8px;
}
</style>
