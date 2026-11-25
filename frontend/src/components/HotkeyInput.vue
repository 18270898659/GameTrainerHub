<template>
  <div
    class="hotkey-input-wrapper"
    :class="{
      'is-recording': isRecording,
      'is-disabled': disabled,
      'is-empty': !modelValue && !isRecording
    }"
    tabindex="0"
    @click="startRecording"
    @keydown="handleKeyDown"
    @keyup="handleKeyUp"
    @blur="handleBlur"
  >
    <!-- 录入状态显示 -->
    <div v-if="isRecording" class="recording-content">
      <span class="recording-pulse"></span>
      <div v-if="heldModifiers.length > 0" class="keys-display">
        <template v-for="(mod, idx) in heldModifiers" :key="'mod-' + idx">
          <kbd class="key-cap mod-cap">{{ mod }}</kbd>
          <span class="key-separator">+</span>
        </template>
        <span class="waiting-prompt">...</span>
      </div>
      <span v-else class="recording-placeholder">
        {{ placeholder || t('settings.hotkeyRecordingPrompt') }}
      </span>
      <span class="recording-hint">{{ t('settings.hotkeyRecordingHint') }}</span>
    </div>

    <!-- 正常已捕获/已保存状态显示 -->
    <div v-else class="hotkey-content">
      <div v-if="keyParts.length > 0" class="keys-display">
        <template v-for="(key, idx) in keyParts" :key="'key-' + idx">
          <kbd class="key-cap" :class="{ 'mod-cap': isModifier(key) }">{{ key }}</kbd>
          <span v-if="idx < keyParts.length - 1" class="key-separator">+</span>
        </template>
      </div>
      <span v-else class="empty-placeholder">
        {{ placeholder || t('settings.hotkeyClickToRecord') }}
      </span>

      <!-- 清空按钮 -->
      <button
        v-if="clearable && modelValue && !disabled"
        type="button"
        class="clear-btn"
        :title="t('settings.hotkeyClear')"
        @click.stop="clearHotkey"
      >
        <svg viewBox="0 0 1024 1024" width="12" height="12">
          <path
            fill="currentColor"
            d="M512 64a448 448 0 1 1 0 896 448 448 0 0 1 0-896zm0 393.6L374.4 320a38.4 38.4 0 0 0-54.4 54.4L457.6 512 320 649.6a38.4 38.4 0 0 0 54.4 54.4L512 566.4 649.6 704a38.4 38.4 0 0 0 54.4-54.4L566.4 512 704 374.4a38.4 38.4 0 0 0-54.4-54.4L512 457.6z"
          />
        </svg>
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { t } from '../i18n'
import { api } from '../api/wails'

const props = withDefaults(
  defineProps<{
    modelValue?: string
    disabled?: boolean
    placeholder?: string
    clearable?: boolean
  }>(),
  {
    modelValue: '',
    disabled: false,
    placeholder: '',
    clearable: true
  }
)

const emit = defineEmits<{
  (e: 'update:modelValue', value: string): void
  (e: 'change', value: string): void
  (e: 'recording-change', isRecording: boolean): void
}>()

const isRecording = ref(false)
const heldModifiers = ref<string[]>([])
let preRecordValue = ''

// 分解快捷键按键序列用于标签展示
const keyParts = computed(() => {
  if (!props.modelValue) return []
  return props.modelValue.split('+').map((s) => s.trim()).filter(Boolean)
})

function isModifier(key: string): boolean {
  return ['Alt', 'Ctrl', 'Shift', 'Win'].includes(key)
}

// 开始录入快捷键
async function startRecording() {
  if (props.disabled || isRecording.value) return
  preRecordValue = props.modelValue || ''
  isRecording.value = true
  heldModifiers.value = []
  emit('recording-change', true)

  // 挂起后端系统全局快捷键拦截，避免物理按键被系统抢占或触发动作
  try {
    await api.suspendHotkeys(true)
  } catch (err) {
    console.warn('Failed to suspend global hotkeys during recording', err)
  }
}

// 停止录入快捷键
async function stopRecording(revert = false) {
  if (!isRecording.value) return
  isRecording.value = false
  heldModifiers.value = []
  emit('recording-change', false)

  if (revert) {
    emit('update:modelValue', preRecordValue)
  }

  // 恢复后端全局快捷键拦截
  try {
    await api.suspendHotkeys(false)
  } catch (err) {
    console.warn('Failed to resume global hotkeys', err)
  }
}

function handleBlur() {
  if (isRecording.value) {
    stopRecording(true)
  }
}

function clearHotkey() {
  if (props.disabled) return
  emit('update:modelValue', '')
  emit('change', '')
}

// 监听键盘按键
function handleKeyDown(e: KeyboardEvent) {
  if (!isRecording.value) return

  // 阻止默认事件与冒泡，防止浏览器内置快捷键介入（如 Ctrl+J, Ctrl+Q, F1 等）
  e.preventDefault()
  e.stopPropagation()

  // 按 Escape 取消录入，恢复原值
  if (e.key === 'Escape') {
    stopRecording(true)
    return
  }

  // 按 Backspace 或 Delete 清空当前输入
  if (e.key === 'Backspace' || e.key === 'Delete') {
    emit('update:modelValue', '')
    emit('change', '')
    stopRecording(false)
    return
  }

  // 采集当前按下的修饰键
  const mods: string[] = []
  if (e.altKey) mods.push('Alt')
  if (e.ctrlKey) mods.push('Ctrl')
  if (e.shiftKey) mods.push('Shift')
  if (e.metaKey) mods.push('Win')

  // 若当前按下的键本身是修饰键，更新当前按下的修饰键预览，不结束录入
  if (['Control', 'Alt', 'Shift', 'Meta'].includes(e.key)) {
    heldModifiers.value = mods
    return
  }

  // 获取主按键规范名称
  const mainKey = parseMainKey(e)
  if (!mainKey) {
    return
  }

  // 普通字符必须包含至少一个修饰键；F1-F24 允许单独设置
  const isFn = /^F\d{1,2}$/i.test(mainKey)
  if (mods.length === 0 && !isFn) {
    return
  }

  // 组合快捷键规范字符串：Alt + Ctrl + Shift + Win + 主键
  const combo = [...mods, mainKey].join(' + ')

  emit('update:modelValue', combo)
  emit('change', combo)
  stopRecording(false)
}

function handleKeyUp(e: KeyboardEvent) {
  if (!isRecording.value) return
  e.preventDefault()
  e.stopPropagation()

  // 刷新当前按下的修饰键
  const mods: string[] = []
  if (e.altKey) mods.push('Alt')
  if (e.ctrlKey) mods.push('Ctrl')
  if (e.shiftKey) mods.push('Shift')
  if (e.metaKey) mods.push('Win')
  heldModifiers.value = mods
}

function parseMainKey(e: KeyboardEvent): string | null {
  const code = e.code
  const key = e.key

  // 字母按键 KeyA - KeyZ
  if (code.startsWith('Key') && code.length === 4) {
    return code.slice(3).toUpperCase()
  }

  // 数字按键 Digit0 - Digit9
  if (code.startsWith('Digit') && code.length === 6) {
    return code.slice(5)
  }

  // 小键盘数字 Numpad0 - Numpad9
  if (code.startsWith('Numpad') && /^\d$/.test(code.slice(6))) {
    return code.slice(6)
  }

  // 功能键 F1 - F24
  if (/^F\d{1,2}$/i.test(key)) {
    return key.toUpperCase()
  }

  // 特殊控制键
  switch (code) {
    case 'Space':
      return 'Space'
    case 'Tab':
      return 'Tab'
    case 'Enter':
    case 'NumpadEnter':
      return 'Enter'
    case 'Insert':
      return 'Insert'
    case 'Home':
      return 'Home'
    case 'End':
      return 'End'
    case 'PageUp':
      return 'PageUp'
    case 'PageDown':
      return 'PageDown'
    case 'ArrowUp':
      return 'Up'
    case 'ArrowDown':
      return 'Down'
    case 'ArrowLeft':
      return 'Left'
    case 'ArrowRight':
      return 'Right'
    case 'Minus':
    case 'NumpadSubtract':
      return '-'
    case 'Equal':
      return '='
    case 'BracketLeft':
      return '['
    case 'BracketRight':
      return ']'
    case 'Backslash':
      return '\\'
    case 'Semicolon':
      return ';'
    case 'Quote':
      return "'"
    case 'Comma':
      return ','
    case 'Period':
    case 'NumpadDecimal':
      return '.'
    case 'Slash':
    case 'NumpadDivide':
      return '/'
    case 'Backquote':
      return '`'
  }

  // 单字符兜底
  if (key && key.length === 1) {
    return key.toUpperCase()
  }

  return null
}
</script>

<style scoped>
.hotkey-input-wrapper {
  position: relative;
  display: inline-flex;
  align-items: center;
  min-height: 40px;
  width: 100%;
  padding: 4px 12px;
  background-color: var(--el-fill-color-blank, #1e222d);
  border: 1px solid var(--el-border-color, #363b4a);
  border-radius: 6px;
  cursor: pointer;
  outline: none;
  transition: all 0.2s cubic-bezier(0.4, 0, 0.2, 1);
  user-select: none;
  box-sizing: border-box;
}

.hotkey-input-wrapper:hover:not(.is-disabled) {
  border-color: var(--el-color-primary, #6366f1);
  background-color: rgba(99, 102, 241, 0.04);
}

.hotkey-input-wrapper:focus-visible,
.hotkey-input-wrapper.is-recording {
  border-color: var(--el-color-primary, #6366f1);
  box-shadow: 0 0 0 3px rgba(99, 102, 241, 0.25);
  background-color: rgba(99, 102, 241, 0.06);
}

.hotkey-input-wrapper.is-disabled {
  opacity: 0.55;
  cursor: not-allowed;
  background-color: var(--el-fill-color-light, #151821);
}

.hotkey-content,
.recording-content {
  display: flex;
  align-items: center;
  width: 100%;
  gap: 8px;
}

.recording-pulse {
  display: inline-block;
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background-color: #ef4444;
  animation: pulse-dot 1.2s infinite ease-in-out;
  flex-shrink: 0;
}

@keyframes pulse-dot {
  0% {
    transform: scale(0.9);
    opacity: 0.6;
  }
  50% {
    transform: scale(1.3);
    opacity: 1;
    box-shadow: 0 0 8px rgba(239, 68, 68, 0.8);
  }
  100% {
    transform: scale(0.9);
    opacity: 0.6;
  }
}

.recording-placeholder {
  color: var(--el-color-primary, #6366f1);
  font-weight: 500;
  font-size: 13px;
}

.recording-hint {
  margin-left: auto;
  font-size: 12px;
  color: var(--el-text-color-secondary, #909399);
}

.empty-placeholder {
  color: var(--el-text-color-placeholder, #606266);
  font-size: 13px;
}

.keys-display {
  display: flex;
  align-items: center;
  gap: 4px;
  flex-wrap: wrap;
}

.key-cap {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 24px;
  height: 24px;
  padding: 0 6px;
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", "Courier New", monospace;
  font-size: 12px;
  font-weight: 600;
  color: #f1f5f9;
  background: linear-gradient(180deg, #2a3040 0%, #1e2433 100%);
  border: 1px solid #475569;
  border-bottom: 2px solid #334155;
  border-radius: 4px;
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.2);
}

.key-cap.mod-cap {
  color: #93c5fd;
  border-color: #3b82f6;
  background: linear-gradient(180deg, #1e293b 0%, #0f172a 100%);
}

.key-separator {
  color: var(--el-text-color-secondary, #94a3b8);
  font-size: 12px;
  font-weight: bold;
  user-select: none;
}

.waiting-prompt {
  color: var(--el-color-primary, #6366f1);
  font-weight: bold;
  animation: blink 1s infinite;
}

@keyframes blink {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.3; }
}

.clear-btn {
  margin-left: auto;
  background: transparent;
  border: none;
  color: var(--el-text-color-placeholder, #64748b);
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 4px;
  border-radius: 4px;
  transition: all 0.15s;
}

.clear-btn:hover {
  color: #ef4444;
  background-color: rgba(239, 68, 68, 0.1);
}
</style>
