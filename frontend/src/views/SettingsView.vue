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

    <!-- 保存按钮栏 -->
    <div class="actions-card">
      <el-button type="primary" size="large" :loading="saving" @click="saveSettings">
        {{ t('settings.saveBtn') }}
      </el-button>
    </div>

    <!-- 关于与声明 -->
    <div class="settings-card about-card">
      <h3 class="settings-title">{{ t('settings.sectionAbout') }}</h3>
      <p class="about-p">{{ t('settings.aboutVersion') }}</p>
      <p class="about-p">{{ t('settings.aboutSource') }}</p>
      <p class="about-p text-gray">
        {{ t('settings.aboutDisclaimer') }}
      </p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { QuestionFilled, FolderOpened, FolderAdd, Moon, Sunny } from '@element-plus/icons-vue'
import { useAppStore } from '../stores/appStore'
import { api } from '../api/wails'
import { t, currentLanguage, setLanguage, type SupportedLanguage } from '../i18n'
import { currentTheme, setTheme, type ThemeMode } from '../theme'
import type { AppConfig } from '../types'

const store = useAppStore()
const form = ref<AppConfig | null>(null)
const saving = ref(false)
const selectingDir = ref(false)
const selectedLanguage = ref<SupportedLanguage>(currentLanguage.value)
const selectedTheme = ref<ThemeMode>(currentTheme.value)

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

  // 前端优先快速检查活跃下载任务
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

// 并发滑块发生改变时即时保存
async function onSliderChange() {
  await saveSettingsSilently()
}

// 语言切换处理
async function onLanguageChange(val: any) {
  const lang = val as SupportedLanguage
  setLanguage(lang)
  selectedLanguage.value = lang
  if (form.value) {
    form.value.language = lang
    await saveSettingsSilently()
  }
}

// 主题切换处理 (即时全局响应并持久化)
async function onThemeChange(val: any) {
  const mode = val as ThemeMode
  setTheme(mode)
  selectedTheme.value = mode
  if (form.value) {
    form.value.theme = mode
    await saveSettingsSilently()
  }
}

// 静默自动保存配置
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

// 手动点击保存配置
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
  font-weight: 600;
  color: var(--text-main);
}

/* 提示图标与 Tooltip */
.field-label-with-tip {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 14px;
  color: var(--text-secondary);
  font-weight: 500;
}
.tip-icon {
  font-size: 15px;
  color: var(--accent-cyan);
  cursor: pointer;
  transition: color 0.2s;
}
.tip-icon:hover {
  color: var(--accent-hover);
}

/* 存储路径卡片 */
.path-display-card {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  background: var(--bg-input);
  border: 1px solid var(--border-color);
  border-radius: 6px;
  padding: 10px 14px;
  width: 100%;
  box-sizing: border-box;
}
.path-display-text {
  display: flex;
  align-items: center;
  gap: 10px;
  cursor: pointer;
  overflow: hidden;
  flex: 1;
  transition: all 0.2s ease;
}
.path-display-text:hover .path-text {
  color: var(--accent-cyan);
  text-decoration: underline;
}
.folder-icon {
  font-size: 18px;
  color: var(--accent-cyan);
  flex-shrink: 0;
}
.path-text {
  font-family: Consolas, "Courier New", monospace;
  font-size: 13px;
  color: var(--text-primary);
  word-break: break-all;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.change-path-btn {
  flex-shrink: 0;
}

/* 并发滑块水平行 */
.slider-horizontal-row {
  display: flex;
  align-items: center;
  gap: 16px;
  width: 100%;
}
.slider-container {
  width: 280px;
  flex-shrink: 0;
}
.slider-val-tag {
  font-size: 13px;
  font-weight: 600;
  color: var(--accent-cyan);
  background: var(--accent-tag-bg);
  border: 1px solid var(--accent-tag-border);
  padding: 2px 10px;
  border-radius: 4px;
  white-space: nowrap;
  flex-shrink: 0;
}
.slider-hint-text {
  font-size: 12px;
  color: var(--text-muted);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  flex-shrink: 1;
}

/* 窗口设置模块 */
.window-settings-grid {
  display: flex;
  flex-direction: column;
}
.setting-item-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 8px 0;
}
.item-text {
  display: flex;
  flex-direction: column;
  gap: 4px;
}
.item-label {
  font-size: 14px;
  font-weight: 500;
  color: var(--text-primary);
}
.item-desc {
  font-size: 12px;
  color: var(--text-muted);
}
.inner-divider {
  margin: 12px 0;
  border-color: var(--border-subtle);
}

/* 语言与主题选择 */
.language-selector-wrap,
.theme-selector-wrap {
  display: flex;
  align-items: center;
}
.theme-btn-icon {
  margin-right: 6px;
  vertical-align: middle;
}

/* 开关行 */
.switch-row {
  display: flex;
  align-items: center;
  gap: 16px;
}
.switch-extra-hint {
  font-size: 13px;
  color: var(--text-muted);
}

.actions-card {
  display: flex;
  justify-content: flex-start;
}

.about-card {
  background: var(--bg-card-subtle);
}
.about-p {
  font-size: 13px;
  color: var(--text-muted);
  line-height: 1.6;
  margin: 6px 0;
}
.text-gray {
  color: var(--text-gray);
  font-size: 12px;
}
</style>

<style>
/* 全局 Tooltip 气泡自定义样式，确保下载路径说明完整展示不截断 */
.download-dir-tooltip {
  max-width: 420px !important;
  white-space: pre-line !important;
  line-height: 1.6 !important;
  font-size: 12px !important;
  padding: 10px 14px !important;
  background-color: var(--bg-modal) !important;
  color: var(--text-main) !important;
  border: 1px solid var(--border-color) !important;
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.4) !important;
}
</style>
