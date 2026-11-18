<template>
  <div class="library-view">
    <!-- Action Bar -->
    <div class="lib-actions-bar">
      <div class="left-stats">
        <span class="lib-count">已管理修改器: <strong>{{ store.localTrainers.length }}</strong> 款</span>
      </div>
      <div class="right-buttons">
        <el-button :loading="scanningSteam" plain @click="scanSteam">
          <el-icon><Monitor /></el-icon> 扫描 Steam 游戏库
        </el-button>
        <el-button type="primary" plain @click="promptImport">
          <el-icon><Plus /></el-icon> 导入外部修改器 EXE
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
          <div class="game-name">{{ item.game?.name_cn || item.game?.name_en || '未关联游戏' }}</div>
          <div class="file-name" :title="item.local_trainer.file_path">
            {{ getFileName(item.local_trainer.file_path) }}
          </div>

          <div class="stats-row">
            <span>运行次数: {{ item.local_trainer.run_count }}</span>
            <span>大小: {{ formatBytes(item.local_trainer.file_size) }}</span>
          </div>

          <div class="card-btns">
            <el-button
              type="primary"
              size="small"
              @click="openLaunchModal(item.local_trainer.id)"
            >
              <el-icon><VideoPlay /></el-icon> 启动
            </el-button>
            <el-button
              size="small"
              plain
              @click="openFolder(item.local_trainer.file_path)"
            >
              目录
            </el-button>
            <el-popconfirm
              title="确定要从本地库中删除此修改器记录吗？"
              confirm-button-text="删除"
              cancel-button-text="取消"
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
      <el-empty description="本地库尚无修改器">
        <template #extra>
          <p class="empty-tip">在搜索页面下载修改器，或点击上方“导入外部修改器 EXE”进行本地管理。</p>
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
import SecurityModal from '../components/SecurityModal.vue'

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
    ElMessage.error('读取启动信息失败: ' + (e.message || e))
  }
}

async function openFolder(path: string) {
  try {
    await api.openContainingFolder(path)
  } catch (e: any) {
    ElMessage.error('打开目录失败: ' + (e.message || e))
  }
}

async function scanSteam() {
  scanningSteam.value = true
  try {
    const count = await api.scanSteamGames()
    ElMessage.success(`Steam 库扫描完成，成功匹配 ${count} 款本地安装游戏`)
    await store.refreshLocalTrainers()
  } catch (e: any) {
    ElMessage.error('扫描失败: ' + (e.message || e))
  } finally {
    scanningSteam.value = false
  }
}

async function promptImport() {
  try {
    const { value: filePath } = await ElMessageBox.prompt('请输入本地修改器可执行程序 (.exe) 的完整绝对路径：', '导入外部修改器', {
      confirmButtonText: '导入',
      cancelButtonText: '取消',
      inputPattern: /^.+\.exe$/i,
      inputErrorMessage: '文件路径必须以 .exe 结尾'
    })

    if (filePath) {
      await api.importCustomTrainer(1, filePath.trim())
      ElMessage.success('导入成功')
      await store.refreshLocalTrainers()
    }
  } catch {
    // 用户取消输入
  }
}

async function deleteTrainer(id: number) {
  try {
    await api.deleteLocalTrainer(id, false)
    ElMessage.success('已删除记录')
    await store.refreshLocalTrainers()
  } catch (e: any) {
    ElMessage.error('删除失败: ' + (e.message || e))
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

