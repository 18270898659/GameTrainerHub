<template>
  <div class="settings-view">
    <!-- 客户端与下载设置 -->
    <div class="settings-card">
      <h3 class="settings-title">{{ t('settings.sectionClientDownload') }}</h3>

      <el-form v-if="form" label-position="top" class="settings-form">
        <!-- 下载目录展示与更改 -->
        <el-form-item>
          <template #label>
            <div class="field-label-with-tip">
              <span>{{ t('settings.downloadDirLabel') }}</span>
              <el-tooltip
                effect="dark"
                placement="top-start"
                popper-class="download-dir-tooltip"
                :content="t('settings.downloadDirTooltip')"
              >
                <el-icon class="tip-icon"><QuestionFilled /></el-icon>
              </el-tooltip>
            </div>
          </template>

          <div class="path-display-card">
            <div
              class="path-display-text"
              :title="form.download_dir"
              @click="openDownloadFolder"
            >
              <el-icon class="folder-icon"><FolderOpened /></el-icon>
              <span class="path-text">{{ form.download_dir || 'C:\\Downloads' }}</span>
            </div>
            <el-button
              type="primary"
              plain
              :loading="selectingDir"
              class="change-path-btn"
              @click="changeDownloadDirectory"
            >
              <el-icon><FolderAdd /></el-icon>
              {{ t('settings.changeDirBtn') }}
            </el-button>
          </div>
        </el-form-item>

        <!-- 最大并发下载数水平滑块 -->
        <el-form-item :label="t('settings.maxConcurrent')">
          <div class="slider-horizontal-row">
            <div class="slider-container">
              <el-slider
                v-model="form.max_concurrent_downloads"
                :min="1"
                :max="5"
                :step="1"
                show-stops
                @change="onSliderChange"
              />
            </div>
            <span class="slider-val-tag">
              {{ form.max_concurrent_downloads }} {{ t('settings.taskCountUnit') }}
            </span>
            <span class="slider-hint-text">
              ({{ t('settings.sliderHint') }})
            </span>
          </div>
        </el-form-item>

        <!-- 网络代理 -->
        <el-form-item :label="t('settings.proxy')">
          <el-input
            v-model="form.proxy_url"
            :placeholder="t('settings.proxyPlaceholder')"
            clearable
          />
        </el-form-item>

        <!-- 原网站自动后台同步 -->
        <el-divider />
        <el-form-item :label="t('settings.autoSync')">
          <div class="switch-row">
            <el-switch
              v-model="form.auto_sync_enabled"
              :active-text="t('settings.autoSyncSwitch')"
              @change="saveSettingsSilently"
            />
            <span v-if="form.auto_sync_enabled" class="switch-extra-hint">
              {{ t('settings.autoSyncInterval', { hours: form.auto_sync_interval_hours }) }}
            </span>
          </div>
        </el-form-item>

        <!-- 安全启动风险确认卡片 -->
        <el-form-item :label="t('settings.safeLaunchPrompt')">
          <el-switch
            v-model="form.safe_launch_prompt"
            :active-text="t('settings.safeLaunchSwitch')"
            @change="saveSettingsSilently"
          />
        </el-form-item>
      </el-form>
    </div>

    <!-- 窗口设置模块 -->
    <div class="settings-card">
      <h3 class="settings-title">{{ t('settings.sectionWindow') }}</h3>

      <div v-if="form" class="window-settings-grid">
        <!-- 记住窗口大小 -->
        <div class="setting-item-row">
          <div class="item-text">
            <div class="item-label">{{ t('settings.rememberWindowSize') }}</div>
            <div class="item-desc">{{ t('settings.rememberWindowSizeDesc') }}</div>
          </div>
          <el-switch
            v-model="form.remember_window_size"
            @change="saveSettingsSilently"
          />
        </div>

        <el-divider class="inner-divider" />

        <!-- 记住窗口位置 -->
        <div class="setting-item-row">
          <div class="item-text">
            <div class="item-label">{{ t('settings.rememberWindowPos') }}</div>
            <div class="item-desc">{{ t('settings.rememberWindowPosDesc') }}</div>
          </div>
          <el-switch
            v-model="form.remember_window_position"
            @change="saveSettingsSilently"
          />
        </div>

        <el-divider class="inner-divider" />

        <!-- 最小化到系统托盘 -->
        <div class="setting-item-row">
          <div class="item-text">
            <div class="item-label">{{ t('settings.closeToTray') }}</div>
            <div class="item-desc">{{ t('settings.closeToTrayDesc') }}</div>
          </div>
          <el-switch
            v-model="form.close_to_tray"
            @change="saveSettingsSilently"
          />
        </div>
      </div>
    </div>

    <!-- 全局快捷键模块 -->
    <div class="settings-card">
      <h3 class="settings-title">{{ t('settings.sectionHotkeys') }}</h3>

      <div v-if="form" class="hotkey-settings-grid">
        <div class="setting-item-row">
          <div class="item-text">
            <div class="item-label">{{ t('settings.hotkeysEnabled') }}</div>
            <div class="item-desc">{{ t('settings.hotkeysEnabledDesc') }}</div>
          </div>
          <el-switch
            v-model="form.global_hotkeys_enabled"
            @change="saveSettingsSilently"
          />
        </div>

        <template v-if="form.global_hotkeys_enabled">
          <el-divider class="inner-divider" />
          <div class="hotkey-inputs-row">
            <div class="hotkey-field">
              <label class="field-label">{{ t('settings.hotkeyToggleWindow') }}</label>
              <el-input
                v-model="form.hotkey_toggle_window"
                placeholder="Alt+Q"
                size="default"
                @blur="validateHotkey(form.hotkey_toggle_window || '', 'toggle')"
              >
                <template #append>
                  <span class="hotkey-badge">Global</span>
                </template>
              </el-input>
            </div>

            <div class="hotkey-field">
              <label class="field-label">{{ t('settings.hotkeyOpenDownloads') }}</label>
              <el-input
                v-model="form.hotkey_open_downloads"
                placeholder="Ctrl+J"
                size="default"
                @blur="validateHotkey(form.hotkey_open_downloads || '', 'downloads')"
              >
                <template #append>
                  <span class="hotkey-badge">Global</span>
                </template>
              </el-input>
            </div>
          </div>
        </template>
      </div>
    </div>

    <!-- 系统通知偏好设置模块 -->
    <div class="settings-card">
      <h3 class="settings-title">{{ t('settings.sectionNotifications') }}</h3>

      <div v-if="form" class="window-settings-grid">
        <!-- 下载完成托盘气泡通知 -->
        <div class="setting-item-row">
          <div class="item-text">
            <div class="item-label">{{ t('settings.notifyDownloadComplete') }}</div>
            <div class="item-desc">{{ t('settings.notifyDownloadCompleteDesc') }}</div>
          </div>
          <el-switch
            v-model="form.notify_on_download_complete"
            @change="saveSettingsSilently"
          />
        </div>

        <el-divider class="inner-divider" />

        <!-- 下载失败托盘气泡通知 -->
        <div class="setting-item-row">
          <div class="item-text">
            <div class="item-label">{{ t('settings.notifyDownloadFailed') }}</div>
            <div class="item-desc">{{ t('settings.notifyDownloadFailedDesc') }}</div>
          </div>
          <el-switch
            v-model="form.notify_on_download_failed"
            @change="saveSettingsSilently"
          />
        </div>

        <el-divider class="inner-divider" />

        <!-- 修改器更新通知 -->
        <div class="setting-item-row">
          <div class="item-text">
            <div class="item-label">{{ t('settings.notifyTrainerUpdate') }}</div>
            <div class="item-desc">{{ t('settings.notifyTrainerUpdateDesc') }}</div>
          </div>
          <el-switch
            v-model="form.notify_on_trainer_update"
            @change="saveSettingsSilently"
          />
        </div>
      </div>
    </div>

    <!-- 语言设置模块 -->
    <div class="settings-card">
      <h3 class="settings-title">{{ t('settings.sectionLanguage') }}</h3>

      <div class="language-selector-wrap">
        <el-radio-group
          v-model="selectedLanguage"
          size="default"
          class="custom-radio-group"
          @change="onLanguageChange"
        >
          <el-radio-button value="zh-CN">{{ t('settings.langZhCN') }}</el-radio-button>
          <el-radio-button value="zh-TW">{{ t('settings.langZhTW') }}</el-radio-button>
          <el-radio-button value="en-US">{{ t('settings.langEnUS') }}</el-radio-button>
        </el-radio-group>
      </div>
    </div>

    <!-- 外观与主题设置模块 -->
    <div class="settings-card">
      <h3 class="settings-title">{{ t('settings.sectionTheme') }}</h3>

      <div class="theme-selector-wrap">
        <el-radio-group
          v-model="selectedTheme"
          size="default"
          class="custom-radio-group"
          @change="onThemeChange"
        >
          <el-radio-button value="dark">
            <el-icon class="theme-btn-icon"><Moon /></el-icon>
            {{ t('settings.themeDark') }}
          </el-radio-button>
          <el-radio-button value="light">
            <el-icon class="theme-btn-icon"><Sunny /></el-icon>
            {{ t('settings.themeLight') }}
          </el-radio-button>
        </el-radio-group>
      </div>
    </div>

    <!-- 配置备份与恢复 -->
    <div class="settings-card">
      <h3 class="settings-title">{{ t('settings.sectionBackup') }}</h3>

      <div v-if="form" class="backup-section">
        <div class="setting-item-row mb-4">
          <div class="item-text">
            <div class="item-label">{{ t('settings.autoBackup') }}</div>
            <div class="item-desc">{{ t('settings.autoBackupDesc') }}</div>
          </div>
          <el-switch
            v-model="form.auto_backup_config"
            @change="saveSettingsSilently"
          />
        </div>

        <div class="backup-actions">
          <el-button plain @click="exportConfig">
            <el-icon><Download /></el-icon> {{ t('settings.exportConfig') }}
          </el-button>
          <el-button plain @click="importConfig">
            <el-icon><Upload /></el-icon> {{ t('settings.importConfig') }}
          </el-button>
          <el-button type="danger" plain @click="confirmResetConfig">
            <el-icon><RefreshLeft /></el-icon> {{ t('settings.resetConfig') }}
          </el-button>
        </div>
      </div>
    </div>

    <!-- 日志与故障诊断 -->
    <div class="settings-card">
      <h3 class="settings-title">{{ t('settings.sectionDiagnostics') }}</h3>

      <div class="diagnostics-actions">
        <el-button type="primary" plain @click="openDiagnosticsModal">
          <el-icon><Connection /></el-icon> {{ t('settings.runDiagnostics') }}
        </el-button>
        <el-button plain @click="openLogsModal">
          <el-icon><Document /></el-icon> {{ t('settings.viewLogs') }}
        </el-button>
        <el-button plain @click="exportDiagnosticsReport">
          <el-icon><Share /></el-icon> {{ t('settings.exportDiagnostics') }}
        </el-button>
      </div>
    </div>

    <!-- 保存按钮栏 -->
    <div class="actions-card">
      <el-button type="primary" size="large" :loading="saving" @click="saveSettings">
        {{ t('settings.saveBtn') }}
      </el-button>
    </div>

    <!-- 关于与声明 -->
    <div class="settings-card about-card">
      <div class="about-header">
        <h3 class="settings-title">{{ t('settings.sectionAbout') }}</h3>
        <el-button size="small" plain @click="checkUpdates" :loading="checkingUpdate">
          <el-icon><Refresh /></el-icon> {{ t('settings.checkUpdate') }}
        </el-button>
      </div>
      <p class="about-p">{{ t('settings.aboutVersion') }}</p>
      <p class="about-p">{{ t('settings.aboutSource') }}</p>
      <p class="about-p text-gray">
        {{ t('settings.aboutDisclaimer') }}
      </p>
    </div>

    <!-- 诊断结果弹窗 -->
    <el-dialog
      v-model="diagDialogVisible"
      :title="t('settings.diagnosticsTitle')"
      width="640px"
      append-to-body
    >
      <div v-loading="runningDiag" class="diag-modal-body">
        <div v-if="diagReport" class="diag-report-grid">
          <div class="diag-item">
            <span class="diag-label">操作系统 / 架构:</span>
            <span class="diag-val">{{ diagReport.os }} / {{ diagReport.arch }}</span>
          </div>
          <div class="diag-item">
            <span class="diag-label">存储可写状态:</span>
            <el-tag size="small" :type="diagReport.download_dir_writable ? 'success' : 'danger'">
              {{ diagReport.download_dir_writable ? '正常' : '不可写' }}
            </el-tag>
          </div>
          <div class="diag-item">
            <span class="diag-label">剩余磁盘空间:</span>
            <span class="diag-val">{{ (diagReport.available_disk_bytes / 1024 / 1024 / 1024).toFixed(2) }} GB</span>
          </div>
          <div class="diag-item">
            <span class="diag-label">数据库状态:</span>
            <el-tag size="small" :type="diagReport.db_integrity_ok ? 'success' : 'danger'">
              {{ diagReport.db_integrity_ok ? '完整正常 (OK)' : '数据库损坏' }}
            </el-tag>
          </div>
          <div v-if="diagReport.network" class="diag-item full-width">
            <span class="diag-label">源站连通性 (flingtrainer.com):</span>
            <span class="diag-val">
              DNS: {{ diagReport.network.dns_resolved ? '已解析' : '失败' }} ·
              HTTP: {{ diagReport.network.is_reachable ? '可达 (' + diagReport.network.http_status + ')' : '异常' }} ·
              延迟: {{ diagReport.network.latency_ms }} ms
            </span>
          </div>
        </div>
      </div>
      <template #footer>
        <el-button @click="diagDialogVisible = false">{{ t('library.cancel') }}</el-button>
        <el-button type="primary" @click="exportDiagnosticsReport">{{ t('settings.exportDiagnostics') }}</el-button>
      </template>
    </el-dialog>

    <!-- 脱敏日志查看弹窗 -->
    <el-dialog
      v-model="logsDialogVisible"
      :title="t('settings.logsTitle')"
      width="780px"
      append-to-body
    >
      <div class="logs-filter-bar">
        <span class="filter-label">{{ t('settings.logFilterCategory') }}</span>
        <el-radio-group v-model="logCategory" size="small" @change="refreshLogs">
          <el-radio-button value="">{{ t('settings.logAll') }}</el-radio-button>
          <el-radio-button value="download">download</el-radio-button>
          <el-radio-button value="launch">launch</el-radio-button>
          <el-radio-button value="search">search</el-radio-button>
          <el-radio-button value="system">system</el-radio-button>
        </el-radio-group>
      </div>

      <div class="log-entries-box">
        <div v-for="(entry, idx) in logEntries" :key="idx" class="log-row">
          <span class="log-time">{{ entry.timestamp }}</span>
          <span class="log-level" :class="`lvl-${entry.level.toLowerCase()}`">[{{ entry.level }}]</span>
          <span class="log-cat">[{{ entry.category }}]</span>
          <span class="log-msg">{{ entry.message }}</span>
        </div>
        <div v-if="logEntries.length === 0" class="no-logs">暂无此分类日志</div>
      </div>
      <template #footer>
        <el-button @click="logsDialogVisible = false">{{ t('library.cancel') }}</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  QuestionFilled, FolderOpened, FolderAdd, Moon, Sunny,
  Download, Upload, RefreshLeft, Connection, Document, Share, Refresh
} from '@element-plus/icons-vue'
import { useAppStore } from '../stores/appStore'
import { api } from '../api/wails'
import { t, currentLanguage, setLanguage, type SupportedLanguage } from '../i18n'
import { currentTheme, setTheme, type ThemeMode } from '../theme'
import type { AppConfig, DiagnosticReport, LogEntry } from '../types'

const store = useAppStore()
const form = ref<AppConfig | null>(null)
const saving = ref(false)
const selectingDir = ref(false)
const selectedLanguage = ref<SupportedLanguage>(currentLanguage.value)
const selectedTheme = ref<ThemeMode>(currentTheme.value)

// 诊断与更新
const diagDialogVisible = ref(false)
const runningDiag = ref(false)
const diagReport = ref<DiagnosticReport | null>(null)

const logsDialogVisible = ref(false)
const logCategory = ref('')
const logEntries = ref<LogEntry[]>([])

const checkingUpdate = ref(false)

onMounted(async () => {
  await store.refreshConfig()
  if (store.config) {
    form.value = { ...store.config }
    if (store.config.language) {
      selectedLanguage.value = store.config.language as SupportedLanguage
      setLanguage(store.config.language as SupportedLanguage)
    }
    if (store.config.theme) {
      selectedTheme.value = store.config.theme as ThemeMode
      setTheme(store.config.theme as ThemeMode)
    }
  }
})

// 点击路径打开所在文件夹
async function openDownloadFolder() {
  if (!form.value?.download_dir) return
  try {
    await api.openContainingFolder(form.value.download_dir)
  } catch (e: any) {
    ElMessage.error(e?.message || String(e))
  }
}

// 调用系统文件夹选择器更换下载目录
async function changeDownloadDirectory() {
  if (!form.value) return

  if (store.activeDownloadsCount > 0) {
    ElMessage.warning(t('settings.activeDownloadConflict'))
    return
  }

  selectingDir.value = true
  try {
    const newDir = await api.selectDownloadDirectory()
    if (newDir && newDir.trim() !== '') {
      form.value.download_dir = newDir
      await store.refreshConfig()
      ElMessage.success(t('settings.savedSuccess'))
    }
  } catch (e: any) {
    ElMessage.error(e?.message || String(e))
  } finally {
    selectingDir.value = false
  }
}

async function onSliderChange() {
  await saveSettingsSilently()
}

async function onLanguageChange(val: any) {
  const lang = val as SupportedLanguage
  setLanguage(lang)
  selectedLanguage.value = lang
  if (form.value) {
    form.value.language = lang
    await saveSettingsSilently()
  }
}

async function onThemeChange(val: any) {
  const mode = val as ThemeMode
  setTheme(mode)
  selectedTheme.value = mode
  if (form.value) {
    form.value.theme = mode
    await saveSettingsSilently()
  }
}

async function validateHotkey(keyStr: string, type: string) {
  if (!keyStr) return
  try {
    const conflict = await api.checkHotkeyConflict(keyStr)
    if (conflict) {
      ElMessage.warning(`${keyStr} ${t('settings.hotkeyConflict')}`)
    } else {
      ElMessage.success(`${keyStr} ${t('settings.hotkeyNormal')}`)
    }
  } catch (e) {
    console.warn('Hotkey conflict check failed', e)
  }
}

async function exportConfig() {
  try {
    const jsonStr = await api.exportConfig()
    await navigator.clipboard.writeText(jsonStr)
    ElMessage.success(t('settings.exportSuccess'))
  } catch (e: any) {
    ElMessage.error('导出失败: ' + (e?.message || e))
  }
}

async function importConfig() {
  try {
    const { value: jsonStr } = await ElMessageBox.prompt(
      t('settings.importPrompt'),
      t('settings.importConfig'),
      {
        confirmButtonText: t('settings.importConfig'),
        cancelButtonText: t('library.cancel'),
        inputType: 'textarea'
      }
    )
    if (jsonStr) {
      await api.importConfig(jsonStr.trim())
      ElMessage.success(t('settings.importSuccess'))
      await store.refreshConfig()
      if (store.config) {
        form.value = { ...store.config }
      }
    }
  } catch (e: any) {
    if (e !== 'cancel') {
      ElMessage.error('导入失败: ' + (e?.message || e))
    }
  }
}

async function confirmResetConfig() {
  try {
    await ElMessageBox.confirm(
      t('settings.resetConfirm'),
      t('settings.resetConfig'),
      {
        confirmButtonText: t('settings.resetConfig'),
        cancelButtonText: t('library.cancel'),
        type: 'warning'
      }
    )
    await api.resetConfig()
    ElMessage.success(t('settings.resetSuccess'))
    await store.refreshConfig()
    if (store.config) {
      form.value = { ...store.config }
    }
  } catch (e: any) {
    if (e !== 'cancel') {
      ElMessage.error('重置失败: ' + (e?.message || e))
    }
  }
}

async function openDiagnosticsModal() {
  diagDialogVisible.value = true
  runningDiag.value = true
  try {
    diagReport.value = await api.diagnoseEnvironment()
  } catch (e: any) {
    ElMessage.error('诊断异常: ' + (e?.message || e))
  } finally {
    runningDiag.value = false
  }
}

async function exportDiagnosticsReport() {
  try {
    const filePath = await api.exportDiagnostics()
    ElMessage.success(t('settings.exportDiagnosticsSuccess') + ': ' + filePath)
  } catch (e: any) {
    ElMessage.error('导出报告失败: ' + (e?.message || e))
  }
}

async function openLogsModal() {
  logsDialogVisible.value = true
  await refreshLogs()
}

async function refreshLogs() {
  try {
    logEntries.value = await api.getRecentLogs(100, logCategory.value)
  } catch (e: any) {
    console.error('Failed to get logs', e)
  }
}

async function checkUpdates() {
  checkingUpdate.value = true
  try {
    const res = await api.checkForUpdates()
    if (res.has_update) {
      ElMessage.info(t('settings.updateAvailable', { version: res.latest_version, current: res.current_version }))
    } else {
      ElMessage.success(t('settings.updateUpToDate', { version: res.current_version }))
    }
  } catch (e: any) {
    ElMessage.error('检查更新失败: ' + (e?.message || e))
  } finally {
    checkingUpdate.value = false
  }
}

async function saveSettingsSilently() {
  if (!form.value) return
  try {
    await api.saveConfig(form.value)
    if (store.config) {
      Object.assign(store.config, form.value)
    }
  } catch (e: any) {
    console.error('Failed to auto save config', e)
  }
}

async function saveSettings() {
  if (!form.value) return
  saving.value = true
  try {
    await api.saveConfig(form.value)
    ElMessage.success(t('settings.savedSuccess'))
    await store.refreshConfig()
  } catch (e: any) {
    ElMessage.error(t('settings.saveFailed') + (e.message || e))
  } finally {
    saving.value = false
  }
}
</script>

<style scoped>
.settings-view {
  max-width: 820px;
  display: flex;
  flex-direction: column;
  gap: 20px;
  padding-bottom: 40px;
}
.settings-card {
  background: var(--bg-card);
  border: 1px solid var(--border-color);
  border-radius: 8px;
  padding: 24px;
  box-shadow: var(--card-shadow);
  transition: all 0.2s ease;
}
.settings-title {
  margin: 0 0 20px 0;
  font-size: 16px;
  font-weight: 700;
  color: var(--text-main);
}
.field-label-with-tip {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 14px;
  font-weight: 500;
  color: var(--text-main);
}
.tip-icon {
  font-size: 14px;
  color: var(--text-muted);
  cursor: pointer;
  outline: none;
}
.path-display-card {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  background: var(--bg-card-subtle);
  border: 1px solid var(--border-color);
  border-radius: 6px;
  padding: 8px 12px;
  width: 100%;
  box-sizing: border-box;
}
.path-display-text {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
  flex: 1;
  min-width: 0;
}
.path-display-text:hover .path-text {
  color: var(--accent-cyan);
  text-decoration: underline;
}
.folder-icon {
  font-size: 16px;
  color: var(--accent-cyan);
  flex-shrink: 0;
}
.path-text {
  font-family: monospace;
  font-size: 13px;
  color: var(--text-main);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  transition: color 0.2s ease;
}
.slider-horizontal-row {
  display: flex;
  align-items: center;
  gap: 16px;
  width: 100%;
}
.slider-container {
  flex: 0 0 260px;
}
.slider-val-tag {
  font-size: 13px;
  font-weight: 600;
  color: var(--accent-cyan);
  min-width: 60px;
}
.slider-hint-text {
  font-size: 12px;
  color: var(--text-muted);
}
.switch-row {
  display: flex;
  align-items: center;
  gap: 16px;
}
.switch-extra-hint {
  font-size: 12px;
  color: var(--text-muted);
}
.window-settings-grid, .hotkey-settings-grid {
  display: flex;
  flex-direction: column;
}
.setting-item-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 4px 0;
}
.item-text {
  display: flex;
  flex-direction: column;
  gap: 4px;
}
.item-label {
  font-size: 14px;
  font-weight: 500;
  color: var(--text-main);
}
.item-desc {
  font-size: 12px;
  color: var(--text-muted);
}
.inner-divider {
  margin: 12px 0;
  border-color: var(--border-color);
}
.hotkey-inputs-row {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 20px;
  margin-top: 12px;
}
.hotkey-field {
  display: flex;
  flex-direction: column;
  gap: 6px;
}
.field-label {
  font-size: 13px;
  color: var(--text-muted);
}
.hotkey-badge {
  font-size: 11px;
  color: var(--text-muted);
}
.backup-actions, .diagnostics-actions {
  display: flex;
  gap: 12px;
  flex-wrap: wrap;
}
.about-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.about-p {
  margin: 6px 0;
  font-size: 13px;
  color: var(--text-muted);
}
.diag-report-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 16px;
}
.diag-item {
  display: flex;
  flex-direction: column;
  gap: 4px;
}
.diag-item.full-width {
  grid-column: span 2;
}
.diag-label {
  font-size: 12px;
  color: var(--text-muted);
}
.diag-val {
  font-size: 13px;
  font-weight: 500;
  color: var(--text-main);
}
.logs-filter-bar {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 12px;
}
.log-entries-box {
  background: var(--code-bg);
  border: 1px solid var(--border-color);
  border-radius: 6px;
  padding: 12px;
  height: 340px;
  overflow-y: auto;
  font-family: monospace;
  font-size: 12px;
  display: flex;
  flex-direction: column;
  gap: 6px;
}
.log-row {
  display: flex;
  gap: 8px;
  line-height: 1.4;
}
.log-time {
  color: var(--text-muted);
  white-space: nowrap;
}
.log-level {
  font-weight: 600;
  white-space: nowrap;
}
.lvl-info { color: #3b82f6; }
.lvl-warn { color: #f59e0b; }
.lvl-error { color: #ef4444; }
.log-cat {
  color: var(--accent-cyan);
  white-space: nowrap;
}
.log-msg {
  color: var(--code-text);
  word-break: break-all;
}
.no-logs {
  color: var(--text-muted);
  text-align: center;
  margin-top: 80px;
}
.mb-4 {
  margin-bottom: 16px;
}
</style>
