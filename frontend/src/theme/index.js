import { ref } from 'vue';
const STORAGE_KEY = 'gth_theme';
export const currentTheme = ref('dark');
export function initTheme(initial) {
    let theme = 'dark';
    if (initial === 'light' || initial === 'dark') {
        theme = initial;
    }
    else if (typeof localStorage !== 'undefined') {
        const cached = localStorage.getItem(STORAGE_KEY);
        if (cached === 'light' || cached === 'dark') {
            theme = cached;
        }
    }
    currentTheme.value = theme;
    applyThemeToDOM(theme);
    return theme;
}
export function setTheme(mode) {
    currentTheme.value = mode;
    applyThemeToDOM(mode);
    if (typeof localStorage !== 'undefined') {
        localStorage.setItem(STORAGE_KEY, mode);
    }
}
function applyThemeToDOM(theme) {
    if (typeof document === 'undefined')
        return;
    const el = document.documentElement;
    el.setAttribute('data-theme', theme);
    if (theme === 'dark') {
        el.classList.add('dark');
    }
    else {
        el.classList.remove('dark');
    }
}
//# sourceMappingURL=index.js.map