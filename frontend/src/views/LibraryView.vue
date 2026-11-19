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
        <el-button type="primary" plain @click="promptImport">
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
      >
        <div class="local-card-thumb">
          <img :src="item.game?.cover_url || defaultThumb" />
          <el-tag
            :type="item.local_trainer.is_custom ? 'warning' : 'success'"
            size="small"
            class="source-badge"
          >
            {{ item.local_trainer.source_name }}
          </el-tag>
        </div>

        <div class="local-card-info">
          <div class="game-name">{{ item.game?.name_cn || item.game?.name_en || t('library.unassociatedGame') }}</div>
          <div class="file-name" :title="item.local_trainer.file_path">
            {{ getFileName(item.local_trainer.file_path) }}
          </div>

          <div class="stats-row">
            <span>{{ t('library.runCount') }} {{ item.local_trainer.run_count }}</span>
            <span>{{ t('library.fileSize') }} {{ formatBytes(item.local_trainer.file_size) }}</span>
          </div>

          <div class="card-btns">
            <el-button
              type="primary"
              size="small"
              @click="openLaunchModal(item.local_trainer.id)"
            >
              <el-icon><VideoPlay /></el-icon> {{ t('library.launchBtn') }}
            </el-button>
            <el-button
              size="small"
              plain
              @click="openFolder(item.local_trainer.file_path)"
            >
              {{ t('library.folderBtn') }}
            </el-button>
            <el-popconfirm
              :title="t('library.deleteConfirm')"
              :confirm-button-text="t('library.delete')"
              :cancel-button-text="t('library.cancel')"
              @confirm="deleteTrainer(item.local_trainer.id)"
            >
              <template #reference>
                <el-button size="small" type="danger" link>
                  <el-icon><Delete /></el-icon>
                </el-button>
              </template>
            </el-popconfirm>
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

    <!-- Security Modal Reference -->
    <SecurityModal ref="securityModalRef" />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { Monitor, Plus, VideoPlay, Delete } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useAppStore } from '../stores/appStore'
import { api } from '../api/wails'
import { t } from '../i18n'
import SecurityModal from '../components/SecurityModal.vue'

defineOptions({ name: 'LibraryView' })

const store = useAppStore()
const scanningSteam = ref(false)
const securityModalRef = ref<InstanceType<typeof SecurityModal> | null>(null)
const defaultThumb = 'https://flingtrainer.com/wp-content/uploads/2020/12/header-7-200x200.jpg'

onMounted(() => {
  store.refreshLocalTrainers()
})

async function openLaunchModal(id: number) {
  try {
    const card = await api.getSecurityCard(id)
    securityModalRef.value?.open(card)
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

async function promptImport() {
  try {
    const { value: filePath } = await ElMessageBox.prompt(t('library.importExe'), t('library.importExe'), {
      confirmButtonText: t('library.importExe'),
      cancelButtonText: t('library.cancel'),
      inputPattern: /^.+\.exe$/i,
      inputErrorMessage: '.exe'
    })

    if (filePath) {
      await api.importCustomTrainer(1, filePath.trim())
      ElMessage.success(t('settings.savedSuccess'))
      await store.refreshLocalTrainers()
    }
  } catch {
    // 用户取消输入
  }
}

async function deleteTrainer(id: number) {
  try {
    await api.deleteLocalTrainer(id, false)
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
  background: rgba(30, 41, 59, 0.4);
  padding: 12px 16px;
  border-radius: 8px;
  border: 1px solid rgba(255, 255, 255, 0.06);
  margin-bottom: 20px;
}
.lib-count {
  font-size: 14px;
  color: #94a3b8;
}
.lib-count strong {
  color: #38bdf8;
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
  background: rgba(30, 41, 59, 0.6);
  border: 1px solid rgba(255, 255, 255, 0.08);
  border-radius: 8px;
  overflow: hidden;
  transition: all 0.2s ease;
}
.local-card:hover {
  transform: translateY(-2px);
  border-color: rgba(56, 189, 248, 0.4);
}
.local-card-thumb {
  position: relative;
  height: 110px;
  background: #0f172a;
}
.local-card-thumb img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}
.source-badge {
  position: absolute;
  top: 8px;
  right: 8px;
}
.local-card-info {
  padding: 14px;
}
.game-name {
  font-size: 15px;
  font-weight: 600;
  color: #f8fafc;
  margin-bottom: 4px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.file-name {
  font-size: 12px;
  color: #94a3b8;
  margin-bottom: 12px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.stats-row {
  display: flex;
  justify-content: space-between;
  font-size: 11px;
  color: #64748b;
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
  color: #64748b;
  margin-top: 6px;
}
</style>

