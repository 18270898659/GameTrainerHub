import { ref } from 'vue'

export type ThemeMode = 'dark' | 'light'

const STORAGE_KEY = 'gth_theme'

export const currentTheme = ref<ThemeMode>('dark')

export function initTheme(initial?: string): ThemeMode {
  let theme: ThemeMode = 'dark'

  if (initial === 'light' || initial === 'dark') {
    theme = initial
  } else if (typeof localStorage !== 'undefined') {
    const cached = localStorage.getItem(STORAGE_KEY)
    if (cached === 'light' || cached === 'dark') {
      theme = cached
    }
  }

  currentTheme.value = theme
  applyThemeToDOM(theme)
  return theme
}

export function setTheme(mode: ThemeMode) {
  currentTheme.value = mode
  applyThemeToDOM(mode)
  if (typeof localStorage !== 'undefined') {
    localStorage.setItem(STORAGE_KEY, mode)
  }
}

function applyThemeToDOM(theme: ThemeMode) {
  if (typeof document === 'undefined') return
  const el = document.documentElement
  el.setAttribute('data-theme', theme)
  if (theme === 'dark') {
    el.classList.add('dark')
  } else {
    el.classList.remove('dark')
  }
}
