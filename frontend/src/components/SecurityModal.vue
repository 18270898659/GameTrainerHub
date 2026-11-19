<template>
  <el-dialog
    v-model="visible"
    title="修改器安全启动确认"
    width="580px"
    class="security-modal"
    :close-on-click-modal="false"
  >
    <div v-if="card" class="security-body">
      <el-alert
        title="安全合规与透明运行提示"
        type="warning"
        :closable="false"
        show-icon
        class="mb-3"
      >
        <template #default>
          <div class="alert-desc">
            游戏修改器通过读取与写入游戏内存（Memory Hooking）改变数值，极易被 Windows Defender 及第三方杀软误报为木马。本客户端绝不关闭杀毒软件、不私自加入白名单，启动前请仔细核对 SHA-256 哈希值。
          </div>
        </template>
      </el-alert>

      <div class="info-table">
        <div class="info-row">
          <span class="label">目标游戏：</span>
          <span class="val font-bold">{{ card.game_name }}</span>
        </div>
        <div class="info-row">
          <span class="label">文件名称：</span>
          <span class="val">{{ card.file_name }}</span>
        </div>
        <div class="info-row">
          <span class="label">文件大小：</span>
          <span class="val">{{ formatBytes(card.file_size) }}</span>
        </div>
        <div class="info-row">
          <span class="label">资源来源：</span>
          <el-tag size="small" type="success">{{ card.source_name }}</el-tag>
        </div>
        <div class="info-row hash-row">
          <span class="label">SHA-256：</span>
          <div class="val hash-text">
            <code>{{ card.sha256 }}</code>
            <el-button link type="primary" size="small" @click="copyHash">复制</el-button>
          </div>
        </div>
        <div class="info-row">
          <span class="label">本地路径：</span>
          <span class="val path-text" :title="card.file_path">{{ card.file_path }}</span>
        </div>
      </div>

      <div class="admin-checkbox mt-3">
        <el-checkbox v-model="runAsAdmin">
          以管理员身份运行
          <span v-if="card.requires_admin" class="admin-hint-alert"> (检测到修改器包含特权权限清单，启动时将触发 Windows UAC 授权)</span>
          <span v-else class="admin-hint"> (若游戏本身有反作弊保护或需特权权限)</span>
        </el-checkbox>
      </div>
    </div>

    <template #footer>
      <div class="dialog-footer">
        <el-button @click="openFolder" plain>
          打开所在文件夹
        </el-button>
        <el-button @click="visible = false">取消</el-button>
        <el-button type="primary" :loading="launching" @click="confirmLaunch">
          确认安全启动
        </el-button>
      </div>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { ElMessage } from 'element-plus'
import { api } from '../api/wails'
import type { SecurityCardInfo } from '../types'

const visible = ref(false)
const card = ref<SecurityCardInfo | null>(null)
const runAsAdmin = ref(false)
const launching = ref(false)

function open(info: SecurityCardInfo) {
  card.value = info
  runAsAdmin.value = info.requires_admin || false
  visible.value = true
}

function formatBytes(bytes: number): string {
  if (bytes <= 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return (bytes / Math.pow(k, i)).toFixed(2) + ' ' + sizes[i]
}

function copyHash() {
  if (card.value?.sha256) {
    navigator.clipboard.writeText(card.value.sha256)
    ElMessage.success('SHA-256 哈希值已复制到剪贴板')
  }
}

async function openFolder() {
  if (card.value?.file_path) {
    try {
      await api.openContainingFolder(card.value.file_path)
    } catch (e: any) {
      ElMessage.error('打开目录失败: ' + (e?.message || e))
    }
  }
}

async function confirmLaunch() {
  if (!card.value) return
  launching.value = true
  try {
    await api.launchTrainer(card.value.trainer_id, runAsAdmin.value)
    ElMessage.success('修改器已成功拉起运行')
    visible.value = false
  } catch (e: any) {
    const msg = e?.message || e || ''
    if (msg.includes('取消') || msg.includes('cancel')) {
      ElMessage.warning('已取消启动: ' + msg)
    } else {
      ElMessage.error('启动失败: ' + msg)
    }
  } finally {
    launching.value = false
  }
}

defineExpose({ open })
</script>

<style scoped>
.security-body {
  padding: 4px;
}
.alert-desc {
  font-size: 12px;
  line-height: 1.6;
  margin-top: 4px;
  color: #dcdfe6;
}
.info-table {
  background: rgba(255, 255, 255, 0.04);
  border-radius: 8px;
  padding: 12px 16px;
  margin-top: 12px;
  border: 1px solid rgba(255, 255, 255, 0.08);
}
.info-row {
  display: flex;
  margin-bottom: 8px;
  font-size: 13px;
  align-items: center;
}
.info-row:last-child {
  margin-bottom: 0;
}
.label {
  width: 85px;
  color: #909399;
  flex-shrink: 0;
}
.val {
  color: #f2f3f5;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.hash-text {
  display: flex;
  align-items: center;
  gap: 8px;
}
.hash-text code {
  background: rgba(0, 0, 0, 0.4);
  padding: 2px 6px;
  border-radius: 4px;
  font-family: monospace;
  font-size: 11px;
  color: #409eff;
}
.path-text {
  font-size: 12px;
  color: #a8abb2;
}
.admin-hint-alert {
  font-size: 12px;
  color: #fbbf24;
}
.admin-hint {
  font-size: 12px;
  color: #94a3b8;
}
.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
}
</style>

