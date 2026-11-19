<template>
  <el-dialog
    v-model="visible"
    :title="t('security.modalTitle')"
    width="580px"
    class="security-modal"
    :close-on-click-modal="false"
  >
    <div v-if="card" class="security-body">
      <el-alert
        :title="t('security.alertTitle')"
        type="warning"
        :closable="false"
        show-icon
        class="mb-3"
      >
        <template #default>
          <div class="alert-desc">
            {{ t('security.alertDesc') }}
          </div>
        </template>
      </el-alert>

      <div class="info-table">
        <div class="info-row">
          <span class="label">{{ t('security.targetGame') }}</span>
          <span class="val font-bold">{{ card.game_name }}</span>
        </div>
        <div class="info-row">
          <span class="label">{{ t('security.fileName') }}</span>
          <span class="val">{{ card.file_name }}</span>
        </div>
        <div class="info-row">
          <span class="label">{{ t('security.fileSize') }}</span>
          <span class="val">{{ formatBytes(card.file_size) }}</span>
        </div>
        <div class="info-row">
          <span class="label">{{ t('security.resourceSource') }}</span>
          <el-tag size="small" type="success">{{ card.source_name }}</el-tag>
        </div>
        <div class="info-row hash-row">
          <span class="label">{{ t('security.sha256') }}</span>
          <div class="val hash-text">
            <code>{{ card.sha256 }}</code>
            <el-button link type="primary" size="small" @click="copyHash">{{ t('security.copy') }}</el-button>
          </div>
        </div>
        <div class="info-row">
          <span class="label">{{ t('security.localPath') }}</span>
          <span class="val path-text" :title="card.file_path">{{ card.file_path }}</span>
        </div>
      </div>

      <div class="admin-checkbox mt-3">
        <el-checkbox v-model="runAsAdmin">
          {{ t('security.runAsAdmin') }}
          <span v-if="card.requires_admin" class="admin-hint-alert">{{ t('security.adminAlert') }}</span>
          <span v-else class="admin-hint">{{ t('security.adminHint') }}</span>
        </el-checkbox>
      </div>
    </div>

    <template #footer>
      <div class="dialog-footer">
        <el-button @click="openFolder" plain>
          {{ t('security.openFolder') }}
        </el-button>
        <el-button @click="visible = false">{{ t('library.cancel') }}</el-button>
        <el-button type="primary" :loading="launching" @click="confirmLaunch">
          {{ t('security.confirmLaunch') }}
        </el-button>
      </div>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { ElMessage } from 'element-plus'
import { api } from '../api/wails'
import { t } from '../i18n'
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
    ElMessage.success(t('security.copied'))
  }
}

async function openFolder() {
  if (card.value?.file_path) {
    try {
      await api.openContainingFolder(card.value.file_path)
    } catch (e: any) {
      ElMessage.error(e?.message || String(e))
    }
  }
}

async function confirmLaunch() {
  if (!card.value) return
  launching.value = true
  try {
    await api.launchTrainer(card.value.trainer_id, runAsAdmin.value)
    ElMessage.success(t('security.launchSuccess'))
    visible.value = false
  } catch (e: any) {
    const msg = e?.message || e || ''
    if (msg.includes('取消') || msg.includes('cancel')) {
      ElMessage.warning(msg)
    } else {
      ElMessage.error(t('security.launchFailed') + msg)
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
