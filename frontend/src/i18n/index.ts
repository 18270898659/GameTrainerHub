import { ref, computed } from 'vue'
import zhCNLocale from 'element-plus/es/locale/lang/zh-cn'
import zhTWLocale from 'element-plus/es/locale/lang/zh-tw'
import enLocale from 'element-plus/es/locale/lang/en'

import zhCN from './zh-CN'
import zhTW from './zh-TW'
import enUS from './en-US'

export type SupportedLanguage = 'zh-CN' | 'zh-TW' | 'en-US'

const messages: Record<SupportedLanguage, any> = {
  'zh-CN': zhCN,
  'zh-TW': zhTW,
  'en-US': enUS
}

// 优先从 localStorage 读取，无缓存时默认简体中文
const savedLang = (typeof localStorage !== 'undefined' ? localStorage.getItem('gth_lang') : null) as SupportedLanguage | null
export const currentLanguage = ref<SupportedLanguage>(
  savedLang && messages[savedLang] ? savedLang : 'zh-CN'
)

// Element Plus 对应的国际化语言包
export const currentElementLocale = computed(() => {
  switch (currentLanguage.value) {
    case 'zh-TW':
      return zhTWLocale
    case 'en-US':
      return enLocale
    default:
      return zhCNLocale
  }
})

// 响应式文本翻译函数
export function t(path: string, params?: Record<string, string | number>): string {
  const lang = currentLanguage.value
  const dict = messages[lang] || messages['zh-CN']

  const keys = path.split('.')
  let val: any = dict
  for (const k of keys) {
    if (val && typeof val === 'object' && k in val) {
      val = val[k]
    } else {
      val = null
      break
    }
  }

  // 若当前语言找不到，回退到简体中文
  if (typeof val !== 'string') {
    let fallback: any = messages['zh-CN']
    for (const k of keys) {
      if (fallback && typeof fallback === 'object' && k in fallback) {
        fallback = fallback[k]
      } else {
        fallback = null
        break
      }
    }
    val = typeof fallback === 'string' ? fallback : path
  }

  // 变量插值替换，例如 {count} 或 {added}
  if (params && typeof val === 'string') {
    for (const [pk, pv] of Object.entries(params)) {
      val = val.replace(new RegExp(`\\{${pk}\\}`, 'g'), String(pv))
    }
  }

  return val
}

// 切换当前语言并持久化到本地存储
export function setLanguage(lang: SupportedLanguage) {
  if (messages[lang]) {
    currentLanguage.value = lang
    if (typeof localStorage !== 'undefined') {
      localStorage.setItem('gth_lang', lang)
    }
  }
}
