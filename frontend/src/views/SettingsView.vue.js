/// <reference types="../../node_modules/.vue-global-types/vue_3.5_0_0_0.d.ts" />
import { ref, onMounted } from 'vue';
import { ElMessage, ElMessageBox } from 'element-plus';
import { QuestionFilled, FolderOpened, FolderAdd, Moon, Sunny, Download, Upload, RefreshLeft, Connection, Document, Share, Refresh } from '@element-plus/icons-vue';
import { useAppStore } from '../stores/appStore';
import { api } from '../api/wails';
import { t, currentLanguage, setLanguage } from '../i18n';
import { currentTheme, setTheme } from '../theme';
const store = useAppStore();
const form = ref(null);
const saving = ref(false);
const selectingDir = ref(false);
const selectedLanguage = ref(currentLanguage.value);
const selectedTheme = ref(currentTheme.value);
// 诊断与更新
const diagDialogVisible = ref(false);
const runningDiag = ref(false);
const diagReport = ref(null);
const logsDialogVisible = ref(false);
const logCategory = ref('');
const logEntries = ref([]);
const checkingUpdate = ref(false);
onMounted(async () => {
    await store.refreshConfig();
    if (store.config) {
        form.value = { ...store.config };
        if (store.config.language) {
            selectedLanguage.value = store.config.language;
            setLanguage(store.config.language);
        }
        if (store.config.theme) {
            selectedTheme.value = store.config.theme;
            setTheme(store.config.theme);
        }
    }
});
// 点击路径打开所在文件夹
async function openDownloadFolder() {
    if (!form.value?.download_dir)
        return;
    try {
        await api.openContainingFolder(form.value.download_dir);
    }
    catch (e) {
        ElMessage.error(e?.message || String(e));
    }
}
// 调用系统文件夹选择器更换下载目录
async function changeDownloadDirectory() {
    if (!form.value)
        return;
    if (store.activeDownloadsCount > 0) {
        ElMessage.warning(t('settings.activeDownloadConflict'));
        return;
    }
    selectingDir.value = true;
    try {
        const newDir = await api.selectDownloadDirectory();
        if (newDir && newDir.trim() !== '') {
            form.value.download_dir = newDir;
            await store.refreshConfig();
            ElMessage.success(t('settings.savedSuccess'));
        }
    }
    catch (e) {
        ElMessage.error(e?.message || String(e));
    }
    finally {
        selectingDir.value = false;
    }
}
async function onSliderChange() {
    await saveSettingsSilently();
}
async function onLanguageChange(val) {
    const lang = val;
    setLanguage(lang);
    selectedLanguage.value = lang;
    if (form.value) {
        form.value.language = lang;
        await saveSettingsSilently();
    }
}
async function onThemeChange(val) {
    const mode = val;
    setTheme(mode);
    selectedTheme.value = mode;
    if (form.value) {
        form.value.theme = mode;
        await saveSettingsSilently();
    }
}
async function validateHotkey(keyStr, type) {
    if (!keyStr)
        return;
    try {
        const conflict = await api.checkHotkeyConflict(keyStr);
        if (conflict) {
            ElMessage.warning(`${keyStr} ${t('settings.hotkeyConflict')}`);
        }
        else {
            ElMessage.success(`${keyStr} ${t('settings.hotkeyNormal')}`);
        }
    }
    catch (e) {
        console.warn('Hotkey conflict check failed', e);
    }
}
async function exportConfig() {
    try {
        const jsonStr = await api.exportConfig();
        await navigator.clipboard.writeText(jsonStr);
        ElMessage.success(t('settings.exportSuccess'));
    }
    catch (e) {
        ElMessage.error('导出失败: ' + (e?.message || e));
    }
}
async function importConfig() {
    try {
        const { value: jsonStr } = await ElMessageBox.prompt(t('settings.importPrompt'), t('settings.importConfig'), {
            confirmButtonText: t('settings.importConfig'),
            cancelButtonText: t('library.cancel'),
            inputType: 'textarea'
        });
        if (jsonStr) {
            await api.importConfig(jsonStr.trim());
            ElMessage.success(t('settings.importSuccess'));
            await store.refreshConfig();
            if (store.config) {
                form.value = { ...store.config };
            }
        }
    }
    catch (e) {
        if (e !== 'cancel') {
            ElMessage.error('导入失败: ' + (e?.message || e));
        }
    }
}
async function confirmResetConfig() {
    try {
        await ElMessageBox.confirm(t('settings.resetConfirm'), t('settings.resetConfig'), {
            confirmButtonText: t('settings.resetConfig'),
            cancelButtonText: t('library.cancel'),
            type: 'warning'
        });
        await api.resetConfig();
        ElMessage.success(t('settings.resetSuccess'));
        await store.refreshConfig();
        if (store.config) {
            form.value = { ...store.config };
        }
    }
    catch (e) {
        if (e !== 'cancel') {
            ElMessage.error('重置失败: ' + (e?.message || e));
        }
    }
}
async function openDiagnosticsModal() {
    diagDialogVisible.value = true;
    runningDiag.value = true;
    try {
        diagReport.value = await api.diagnoseEnvironment();
    }
    catch (e) {
        ElMessage.error('诊断异常: ' + (e?.message || e));
    }
    finally {
        runningDiag.value = false;
    }
}
async function exportDiagnosticsReport() {
    try {
        const filePath = await api.exportDiagnostics();
        ElMessage.success(t('settings.exportDiagnosticsSuccess') + ': ' + filePath);
    }
    catch (e) {
        ElMessage.error('导出报告失败: ' + (e?.message || e));
    }
}
async function openLogsModal() {
    logsDialogVisible.value = true;
    await refreshLogs();
}
async function refreshLogs() {
    try {
        logEntries.value = await api.getRecentLogs(100, logCategory.value);
    }
    catch (e) {
        console.error('Failed to get logs', e);
    }
}
async function checkUpdates() {
    checkingUpdate.value = true;
    try {
        const res = await api.checkForUpdates();
        if (res.has_update) {
            ElMessage.info(t('settings.updateAvailable', { version: res.latest_version, current: res.current_version }));
        }
        else {
            ElMessage.success(t('settings.updateUpToDate', { version: res.current_version }));
        }
    }
    catch (e) {
        ElMessage.error('检查更新失败: ' + (e?.message || e));
    }
    finally {
        checkingUpdate.value = false;
    }
}
async function saveSettingsSilently() {
    if (!form.value)
        return;
    try {
        await api.saveConfig(form.value);
        if (store.config) {
            Object.assign(store.config, form.value);
        }
    }
    catch (e) {
        console.error('Failed to auto save config', e);
    }
}
async function saveSettings() {
    if (!form.value)
        return;
    saving.value = true;
    try {
        await api.saveConfig(form.value);
        ElMessage.success(t('settings.savedSuccess'));
        await store.refreshConfig();
    }
    catch (e) {
        ElMessage.error(t('settings.saveFailed') + (e.message || e));
    }
    finally {
        saving.value = false;
    }
}
debugger; /* PartiallyEnd: #3632/scriptSetup.vue */
const __VLS_ctx = {};
let __VLS_components;
let __VLS_directives;
/** @type {__VLS_StyleScopedClasses['path-display-text']} */ ;
/** @type {__VLS_StyleScopedClasses['path-text']} */ ;
/** @type {__VLS_StyleScopedClasses['diag-item']} */ ;
// CSS variable injection 
// CSS variable injection end 
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "settings-view" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "settings-card" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.h3, __VLS_intrinsicElements.h3)({
    ...{ class: "settings-title" },
});
(__VLS_ctx.t('settings.sectionClientDownload'));
if (__VLS_ctx.form) {
    const __VLS_0 = {}.ElForm;
    /** @type {[typeof __VLS_components.ElForm, typeof __VLS_components.elForm, typeof __VLS_components.ElForm, typeof __VLS_components.elForm, ]} */ ;
    // @ts-ignore
    const __VLS_1 = __VLS_asFunctionalComponent(__VLS_0, new __VLS_0({
        labelPosition: "top",
        ...{ class: "settings-form" },
    }));
    const __VLS_2 = __VLS_1({
        labelPosition: "top",
        ...{ class: "settings-form" },
    }, ...__VLS_functionalComponentArgsRest(__VLS_1));
    __VLS_3.slots.default;
    const __VLS_4 = {}.ElFormItem;
    /** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
    // @ts-ignore
    const __VLS_5 = __VLS_asFunctionalComponent(__VLS_4, new __VLS_4({}));
    const __VLS_6 = __VLS_5({}, ...__VLS_functionalComponentArgsRest(__VLS_5));
    __VLS_7.slots.default;
    {
        const { label: __VLS_thisSlot } = __VLS_7.slots;
        __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
            ...{ class: "field-label-with-tip" },
        });
        __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({});
        (__VLS_ctx.t('settings.downloadDirLabel'));
        const __VLS_8 = {}.ElTooltip;
        /** @type {[typeof __VLS_components.ElTooltip, typeof __VLS_components.elTooltip, typeof __VLS_components.ElTooltip, typeof __VLS_components.elTooltip, ]} */ ;
        // @ts-ignore
        const __VLS_9 = __VLS_asFunctionalComponent(__VLS_8, new __VLS_8({
            effect: "dark",
            placement: "top-start",
            popperClass: "download-dir-tooltip",
            content: (__VLS_ctx.t('settings.downloadDirTooltip')),
        }));
        const __VLS_10 = __VLS_9({
            effect: "dark",
            placement: "top-start",
            popperClass: "download-dir-tooltip",
            content: (__VLS_ctx.t('settings.downloadDirTooltip')),
        }, ...__VLS_functionalComponentArgsRest(__VLS_9));
        __VLS_11.slots.default;
        const __VLS_12 = {}.ElIcon;
        /** @type {[typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, ]} */ ;
        // @ts-ignore
        const __VLS_13 = __VLS_asFunctionalComponent(__VLS_12, new __VLS_12({
            ...{ class: "tip-icon" },
        }));
        const __VLS_14 = __VLS_13({
            ...{ class: "tip-icon" },
        }, ...__VLS_functionalComponentArgsRest(__VLS_13));
        __VLS_15.slots.default;
        const __VLS_16 = {}.QuestionFilled;
        /** @type {[typeof __VLS_components.QuestionFilled, ]} */ ;
        // @ts-ignore
        const __VLS_17 = __VLS_asFunctionalComponent(__VLS_16, new __VLS_16({}));
        const __VLS_18 = __VLS_17({}, ...__VLS_functionalComponentArgsRest(__VLS_17));
        var __VLS_15;
        var __VLS_11;
    }
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "path-display-card" },
    });
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ onClick: (__VLS_ctx.openDownloadFolder) },
        ...{ class: "path-display-text" },
        title: (__VLS_ctx.form.download_dir),
    });
    const __VLS_20 = {}.ElIcon;
    /** @type {[typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, ]} */ ;
    // @ts-ignore
    const __VLS_21 = __VLS_asFunctionalComponent(__VLS_20, new __VLS_20({
        ...{ class: "folder-icon" },
    }));
    const __VLS_22 = __VLS_21({
        ...{ class: "folder-icon" },
    }, ...__VLS_functionalComponentArgsRest(__VLS_21));
    __VLS_23.slots.default;
    const __VLS_24 = {}.FolderOpened;
    /** @type {[typeof __VLS_components.FolderOpened, ]} */ ;
    // @ts-ignore
    const __VLS_25 = __VLS_asFunctionalComponent(__VLS_24, new __VLS_24({}));
    const __VLS_26 = __VLS_25({}, ...__VLS_functionalComponentArgsRest(__VLS_25));
    var __VLS_23;
    __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
        ...{ class: "path-text" },
    });
    (__VLS_ctx.form.download_dir || 'C:\\Downloads');
    const __VLS_28 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_29 = __VLS_asFunctionalComponent(__VLS_28, new __VLS_28({
        ...{ 'onClick': {} },
        type: "primary",
        plain: true,
        loading: (__VLS_ctx.selectingDir),
        ...{ class: "change-path-btn" },
    }));
    const __VLS_30 = __VLS_29({
        ...{ 'onClick': {} },
        type: "primary",
        plain: true,
        loading: (__VLS_ctx.selectingDir),
        ...{ class: "change-path-btn" },
    }, ...__VLS_functionalComponentArgsRest(__VLS_29));
    let __VLS_32;
    let __VLS_33;
    let __VLS_34;
    const __VLS_35 = {
        onClick: (__VLS_ctx.changeDownloadDirectory)
    };
    __VLS_31.slots.default;
    const __VLS_36 = {}.ElIcon;
    /** @type {[typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, ]} */ ;
    // @ts-ignore
    const __VLS_37 = __VLS_asFunctionalComponent(__VLS_36, new __VLS_36({}));
    const __VLS_38 = __VLS_37({}, ...__VLS_functionalComponentArgsRest(__VLS_37));
    __VLS_39.slots.default;
    const __VLS_40 = {}.FolderAdd;
    /** @type {[typeof __VLS_components.FolderAdd, ]} */ ;
    // @ts-ignore
    const __VLS_41 = __VLS_asFunctionalComponent(__VLS_40, new __VLS_40({}));
    const __VLS_42 = __VLS_41({}, ...__VLS_functionalComponentArgsRest(__VLS_41));
    var __VLS_39;
    (__VLS_ctx.t('settings.changeDirBtn'));
    var __VLS_31;
    var __VLS_7;
    const __VLS_44 = {}.ElFormItem;
    /** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
    // @ts-ignore
    const __VLS_45 = __VLS_asFunctionalComponent(__VLS_44, new __VLS_44({
        label: (__VLS_ctx.t('settings.maxConcurrent')),
    }));
    const __VLS_46 = __VLS_45({
        label: (__VLS_ctx.t('settings.maxConcurrent')),
    }, ...__VLS_functionalComponentArgsRest(__VLS_45));
    __VLS_47.slots.default;
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "slider-horizontal-row" },
    });
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "slider-container" },
    });
    const __VLS_48 = {}.ElSlider;
    /** @type {[typeof __VLS_components.ElSlider, typeof __VLS_components.elSlider, ]} */ ;
    // @ts-ignore
    const __VLS_49 = __VLS_asFunctionalComponent(__VLS_48, new __VLS_48({
        ...{ 'onChange': {} },
        modelValue: (__VLS_ctx.form.max_concurrent_downloads),
        min: (1),
        max: (5),
        step: (1),
        showStops: true,
    }));
    const __VLS_50 = __VLS_49({
        ...{ 'onChange': {} },
        modelValue: (__VLS_ctx.form.max_concurrent_downloads),
        min: (1),
        max: (5),
        step: (1),
        showStops: true,
    }, ...__VLS_functionalComponentArgsRest(__VLS_49));
    let __VLS_52;
    let __VLS_53;
    let __VLS_54;
    const __VLS_55 = {
        onChange: (__VLS_ctx.onSliderChange)
    };
    var __VLS_51;
    __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
        ...{ class: "slider-val-tag" },
    });
    (__VLS_ctx.form.max_concurrent_downloads);
    (__VLS_ctx.t('settings.taskCountUnit'));
    __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
        ...{ class: "slider-hint-text" },
    });
    (__VLS_ctx.t('settings.sliderHint'));
    var __VLS_47;
    const __VLS_56 = {}.ElFormItem;
    /** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
    // @ts-ignore
    const __VLS_57 = __VLS_asFunctionalComponent(__VLS_56, new __VLS_56({
        label: (__VLS_ctx.t('settings.proxy')),
    }));
    const __VLS_58 = __VLS_57({
        label: (__VLS_ctx.t('settings.proxy')),
    }, ...__VLS_functionalComponentArgsRest(__VLS_57));
    __VLS_59.slots.default;
    const __VLS_60 = {}.ElInput;
    /** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
    // @ts-ignore
    const __VLS_61 = __VLS_asFunctionalComponent(__VLS_60, new __VLS_60({
        modelValue: (__VLS_ctx.form.proxy_url),
        placeholder: (__VLS_ctx.t('settings.proxyPlaceholder')),
        clearable: true,
    }));
    const __VLS_62 = __VLS_61({
        modelValue: (__VLS_ctx.form.proxy_url),
        placeholder: (__VLS_ctx.t('settings.proxyPlaceholder')),
        clearable: true,
    }, ...__VLS_functionalComponentArgsRest(__VLS_61));
    var __VLS_59;
    const __VLS_64 = {}.ElDivider;
    /** @type {[typeof __VLS_components.ElDivider, typeof __VLS_components.elDivider, ]} */ ;
    // @ts-ignore
    const __VLS_65 = __VLS_asFunctionalComponent(__VLS_64, new __VLS_64({}));
    const __VLS_66 = __VLS_65({}, ...__VLS_functionalComponentArgsRest(__VLS_65));
    const __VLS_68 = {}.ElFormItem;
    /** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
    // @ts-ignore
    const __VLS_69 = __VLS_asFunctionalComponent(__VLS_68, new __VLS_68({
        label: (__VLS_ctx.t('settings.autoSync')),
    }));
    const __VLS_70 = __VLS_69({
        label: (__VLS_ctx.t('settings.autoSync')),
    }, ...__VLS_functionalComponentArgsRest(__VLS_69));
    __VLS_71.slots.default;
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "switch-row" },
    });
    const __VLS_72 = {}.ElSwitch;
    /** @type {[typeof __VLS_components.ElSwitch, typeof __VLS_components.elSwitch, ]} */ ;
    // @ts-ignore
    const __VLS_73 = __VLS_asFunctionalComponent(__VLS_72, new __VLS_72({
        ...{ 'onChange': {} },
        modelValue: (__VLS_ctx.form.auto_sync_enabled),
        activeText: (__VLS_ctx.t('settings.autoSyncSwitch')),
    }));
    const __VLS_74 = __VLS_73({
        ...{ 'onChange': {} },
        modelValue: (__VLS_ctx.form.auto_sync_enabled),
        activeText: (__VLS_ctx.t('settings.autoSyncSwitch')),
    }, ...__VLS_functionalComponentArgsRest(__VLS_73));
    let __VLS_76;
    let __VLS_77;
    let __VLS_78;
    const __VLS_79 = {
        onChange: (__VLS_ctx.saveSettingsSilently)
    };
    var __VLS_75;
    if (__VLS_ctx.form.auto_sync_enabled) {
        __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
            ...{ class: "switch-extra-hint" },
        });
        (__VLS_ctx.t('settings.autoSyncInterval', { hours: __VLS_ctx.form.auto_sync_interval_hours }));
    }
    var __VLS_71;
    const __VLS_80 = {}.ElFormItem;
    /** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
    // @ts-ignore
    const __VLS_81 = __VLS_asFunctionalComponent(__VLS_80, new __VLS_80({
        label: (__VLS_ctx.t('settings.safeLaunchPrompt')),
    }));
    const __VLS_82 = __VLS_81({
        label: (__VLS_ctx.t('settings.safeLaunchPrompt')),
    }, ...__VLS_functionalComponentArgsRest(__VLS_81));
    __VLS_83.slots.default;
    const __VLS_84 = {}.ElSwitch;
    /** @type {[typeof __VLS_components.ElSwitch, typeof __VLS_components.elSwitch, ]} */ ;
    // @ts-ignore
    const __VLS_85 = __VLS_asFunctionalComponent(__VLS_84, new __VLS_84({
        ...{ 'onChange': {} },
        modelValue: (__VLS_ctx.form.safe_launch_prompt),
        activeText: (__VLS_ctx.t('settings.safeLaunchSwitch')),
    }));
    const __VLS_86 = __VLS_85({
        ...{ 'onChange': {} },
        modelValue: (__VLS_ctx.form.safe_launch_prompt),
        activeText: (__VLS_ctx.t('settings.safeLaunchSwitch')),
    }, ...__VLS_functionalComponentArgsRest(__VLS_85));
    let __VLS_88;
    let __VLS_89;
    let __VLS_90;
    const __VLS_91 = {
        onChange: (__VLS_ctx.saveSettingsSilently)
    };
    var __VLS_87;
    var __VLS_83;
    var __VLS_3;
}
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "settings-card" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.h3, __VLS_intrinsicElements.h3)({
    ...{ class: "settings-title" },
});
(__VLS_ctx.t('settings.sectionWindow'));
if (__VLS_ctx.form) {
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "window-settings-grid" },
    });
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "setting-item-row" },
    });
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "item-text" },
    });
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "item-label" },
    });
    (__VLS_ctx.t('settings.rememberWindowSize'));
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "item-desc" },
    });
    (__VLS_ctx.t('settings.rememberWindowSizeDesc'));
    const __VLS_92 = {}.ElSwitch;
    /** @type {[typeof __VLS_components.ElSwitch, typeof __VLS_components.elSwitch, ]} */ ;
    // @ts-ignore
    const __VLS_93 = __VLS_asFunctionalComponent(__VLS_92, new __VLS_92({
        ...{ 'onChange': {} },
        modelValue: (__VLS_ctx.form.remember_window_size),
    }));
    const __VLS_94 = __VLS_93({
        ...{ 'onChange': {} },
        modelValue: (__VLS_ctx.form.remember_window_size),
    }, ...__VLS_functionalComponentArgsRest(__VLS_93));
    let __VLS_96;
    let __VLS_97;
    let __VLS_98;
    const __VLS_99 = {
        onChange: (__VLS_ctx.saveSettingsSilently)
    };
    var __VLS_95;
    const __VLS_100 = {}.ElDivider;
    /** @type {[typeof __VLS_components.ElDivider, typeof __VLS_components.elDivider, ]} */ ;
    // @ts-ignore
    const __VLS_101 = __VLS_asFunctionalComponent(__VLS_100, new __VLS_100({
        ...{ class: "inner-divider" },
    }));
    const __VLS_102 = __VLS_101({
        ...{ class: "inner-divider" },
    }, ...__VLS_functionalComponentArgsRest(__VLS_101));
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "setting-item-row" },
    });
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "item-text" },
    });
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "item-label" },
    });
    (__VLS_ctx.t('settings.rememberWindowPos'));
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "item-desc" },
    });
    (__VLS_ctx.t('settings.rememberWindowPosDesc'));
    const __VLS_104 = {}.ElSwitch;
    /** @type {[typeof __VLS_components.ElSwitch, typeof __VLS_components.elSwitch, ]} */ ;
    // @ts-ignore
    const __VLS_105 = __VLS_asFunctionalComponent(__VLS_104, new __VLS_104({
        ...{ 'onChange': {} },
        modelValue: (__VLS_ctx.form.remember_window_position),
    }));
    const __VLS_106 = __VLS_105({
        ...{ 'onChange': {} },
        modelValue: (__VLS_ctx.form.remember_window_position),
    }, ...__VLS_functionalComponentArgsRest(__VLS_105));
    let __VLS_108;
    let __VLS_109;
    let __VLS_110;
    const __VLS_111 = {
        onChange: (__VLS_ctx.saveSettingsSilently)
    };
    var __VLS_107;
    const __VLS_112 = {}.ElDivider;
    /** @type {[typeof __VLS_components.ElDivider, typeof __VLS_components.elDivider, ]} */ ;
    // @ts-ignore
    const __VLS_113 = __VLS_asFunctionalComponent(__VLS_112, new __VLS_112({
        ...{ class: "inner-divider" },
    }));
    const __VLS_114 = __VLS_113({
        ...{ class: "inner-divider" },
    }, ...__VLS_functionalComponentArgsRest(__VLS_113));
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "setting-item-row" },
    });
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "item-text" },
    });
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "item-label" },
    });
    (__VLS_ctx.t('settings.closeToTray'));
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "item-desc" },
    });
    (__VLS_ctx.t('settings.closeToTrayDesc'));
    const __VLS_116 = {}.ElSwitch;
    /** @type {[typeof __VLS_components.ElSwitch, typeof __VLS_components.elSwitch, ]} */ ;
    // @ts-ignore
    const __VLS_117 = __VLS_asFunctionalComponent(__VLS_116, new __VLS_116({
        ...{ 'onChange': {} },
        modelValue: (__VLS_ctx.form.close_to_tray),
    }));
    const __VLS_118 = __VLS_117({
        ...{ 'onChange': {} },
        modelValue: (__VLS_ctx.form.close_to_tray),
    }, ...__VLS_functionalComponentArgsRest(__VLS_117));
    let __VLS_120;
    let __VLS_121;
    let __VLS_122;
    const __VLS_123 = {
        onChange: (__VLS_ctx.saveSettingsSilently)
    };
    var __VLS_119;
}
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "settings-card" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.h3, __VLS_intrinsicElements.h3)({
    ...{ class: "settings-title" },
});
(__VLS_ctx.t('settings.sectionHotkeys'));
if (__VLS_ctx.form) {
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "hotkey-settings-grid" },
    });
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "setting-item-row" },
    });
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "item-text" },
    });
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "item-label" },
    });
    (__VLS_ctx.t('settings.hotkeysEnabled'));
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "item-desc" },
    });
    (__VLS_ctx.t('settings.hotkeysEnabledDesc'));
    const __VLS_124 = {}.ElSwitch;
    /** @type {[typeof __VLS_components.ElSwitch, typeof __VLS_components.elSwitch, ]} */ ;
    // @ts-ignore
    const __VLS_125 = __VLS_asFunctionalComponent(__VLS_124, new __VLS_124({
        ...{ 'onChange': {} },
        modelValue: (__VLS_ctx.form.global_hotkeys_enabled),
    }));
    const __VLS_126 = __VLS_125({
        ...{ 'onChange': {} },
        modelValue: (__VLS_ctx.form.global_hotkeys_enabled),
    }, ...__VLS_functionalComponentArgsRest(__VLS_125));
    let __VLS_128;
    let __VLS_129;
    let __VLS_130;
    const __VLS_131 = {
        onChange: (__VLS_ctx.saveSettingsSilently)
    };
    var __VLS_127;
    if (__VLS_ctx.form.global_hotkeys_enabled) {
        const __VLS_132 = {}.ElDivider;
        /** @type {[typeof __VLS_components.ElDivider, typeof __VLS_components.elDivider, ]} */ ;
        // @ts-ignore
        const __VLS_133 = __VLS_asFunctionalComponent(__VLS_132, new __VLS_132({
            ...{ class: "inner-divider" },
        }));
        const __VLS_134 = __VLS_133({
            ...{ class: "inner-divider" },
        }, ...__VLS_functionalComponentArgsRest(__VLS_133));
        __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
            ...{ class: "hotkey-inputs-row" },
        });
        __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
            ...{ class: "hotkey-field" },
        });
        __VLS_asFunctionalElement(__VLS_intrinsicElements.label, __VLS_intrinsicElements.label)({
            ...{ class: "field-label" },
        });
        (__VLS_ctx.t('settings.hotkeyToggleWindow'));
        const __VLS_136 = {}.ElInput;
        /** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
        // @ts-ignore
        const __VLS_137 = __VLS_asFunctionalComponent(__VLS_136, new __VLS_136({
            ...{ 'onBlur': {} },
            modelValue: (__VLS_ctx.form.hotkey_toggle_window),
            placeholder: "Alt+Q",
            size: "default",
        }));
        const __VLS_138 = __VLS_137({
            ...{ 'onBlur': {} },
            modelValue: (__VLS_ctx.form.hotkey_toggle_window),
            placeholder: "Alt+Q",
            size: "default",
        }, ...__VLS_functionalComponentArgsRest(__VLS_137));
        let __VLS_140;
        let __VLS_141;
        let __VLS_142;
        const __VLS_143 = {
            onBlur: (...[$event]) => {
                if (!(__VLS_ctx.form))
                    return;
                if (!(__VLS_ctx.form.global_hotkeys_enabled))
                    return;
                __VLS_ctx.validateHotkey(__VLS_ctx.form.hotkey_toggle_window || '', 'toggle');
            }
        };
        __VLS_139.slots.default;
        {
            const { append: __VLS_thisSlot } = __VLS_139.slots;
            __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
                ...{ class: "hotkey-badge" },
            });
        }
        var __VLS_139;
        __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
            ...{ class: "hotkey-field" },
        });
        __VLS_asFunctionalElement(__VLS_intrinsicElements.label, __VLS_intrinsicElements.label)({
            ...{ class: "field-label" },
        });
        (__VLS_ctx.t('settings.hotkeyOpenDownloads'));
        const __VLS_144 = {}.ElInput;
        /** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
        // @ts-ignore
        const __VLS_145 = __VLS_asFunctionalComponent(__VLS_144, new __VLS_144({
            ...{ 'onBlur': {} },
            modelValue: (__VLS_ctx.form.hotkey_open_downloads),
            placeholder: "Ctrl+J",
            size: "default",
        }));
        const __VLS_146 = __VLS_145({
            ...{ 'onBlur': {} },
            modelValue: (__VLS_ctx.form.hotkey_open_downloads),
            placeholder: "Ctrl+J",
            size: "default",
        }, ...__VLS_functionalComponentArgsRest(__VLS_145));
        let __VLS_148;
        let __VLS_149;
        let __VLS_150;
        const __VLS_151 = {
            onBlur: (...[$event]) => {
                if (!(__VLS_ctx.form))
                    return;
                if (!(__VLS_ctx.form.global_hotkeys_enabled))
                    return;
                __VLS_ctx.validateHotkey(__VLS_ctx.form.hotkey_open_downloads || '', 'downloads');
            }
        };
        __VLS_147.slots.default;
        {
            const { append: __VLS_thisSlot } = __VLS_147.slots;
            __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
                ...{ class: "hotkey-badge" },
            });
        }
        var __VLS_147;
    }
}
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "settings-card" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.h3, __VLS_intrinsicElements.h3)({
    ...{ class: "settings-title" },
});
(__VLS_ctx.t('settings.sectionNotifications'));
if (__VLS_ctx.form) {
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "window-settings-grid" },
    });
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "setting-item-row" },
    });
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "item-text" },
    });
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "item-label" },
    });
    (__VLS_ctx.t('settings.notifyDownloadComplete'));
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "item-desc" },
    });
    (__VLS_ctx.t('settings.notifyDownloadCompleteDesc'));
    const __VLS_152 = {}.ElSwitch;
    /** @type {[typeof __VLS_components.ElSwitch, typeof __VLS_components.elSwitch, ]} */ ;
    // @ts-ignore
    const __VLS_153 = __VLS_asFunctionalComponent(__VLS_152, new __VLS_152({
        ...{ 'onChange': {} },
        modelValue: (__VLS_ctx.form.notify_on_download_complete),
    }));
    const __VLS_154 = __VLS_153({
        ...{ 'onChange': {} },
        modelValue: (__VLS_ctx.form.notify_on_download_complete),
    }, ...__VLS_functionalComponentArgsRest(__VLS_153));
    let __VLS_156;
    let __VLS_157;
    let __VLS_158;
    const __VLS_159 = {
        onChange: (__VLS_ctx.saveSettingsSilently)
    };
    var __VLS_155;
    const __VLS_160 = {}.ElDivider;
    /** @type {[typeof __VLS_components.ElDivider, typeof __VLS_components.elDivider, ]} */ ;
    // @ts-ignore
    const __VLS_161 = __VLS_asFunctionalComponent(__VLS_160, new __VLS_160({
        ...{ class: "inner-divider" },
    }));
    const __VLS_162 = __VLS_161({
        ...{ class: "inner-divider" },
    }, ...__VLS_functionalComponentArgsRest(__VLS_161));
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "setting-item-row" },
    });
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "item-text" },
    });
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "item-label" },
    });
    (__VLS_ctx.t('settings.notifyDownloadFailed'));
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "item-desc" },
    });
    (__VLS_ctx.t('settings.notifyDownloadFailedDesc'));
    const __VLS_164 = {}.ElSwitch;
    /** @type {[typeof __VLS_components.ElSwitch, typeof __VLS_components.elSwitch, ]} */ ;
    // @ts-ignore
    const __VLS_165 = __VLS_asFunctionalComponent(__VLS_164, new __VLS_164({
        ...{ 'onChange': {} },
        modelValue: (__VLS_ctx.form.notify_on_download_failed),
    }));
    const __VLS_166 = __VLS_165({
        ...{ 'onChange': {} },
        modelValue: (__VLS_ctx.form.notify_on_download_failed),
    }, ...__VLS_functionalComponentArgsRest(__VLS_165));
    let __VLS_168;
    let __VLS_169;
    let __VLS_170;
    const __VLS_171 = {
        onChange: (__VLS_ctx.saveSettingsSilently)
    };
    var __VLS_167;
    const __VLS_172 = {}.ElDivider;
    /** @type {[typeof __VLS_components.ElDivider, typeof __VLS_components.elDivider, ]} */ ;
    // @ts-ignore
    const __VLS_173 = __VLS_asFunctionalComponent(__VLS_172, new __VLS_172({
        ...{ class: "inner-divider" },
    }));
    const __VLS_174 = __VLS_173({
        ...{ class: "inner-divider" },
    }, ...__VLS_functionalComponentArgsRest(__VLS_173));
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "setting-item-row" },
    });
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "item-text" },
    });
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "item-label" },
    });
    (__VLS_ctx.t('settings.notifyTrainerUpdate'));
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "item-desc" },
    });
    (__VLS_ctx.t('settings.notifyTrainerUpdateDesc'));
    const __VLS_176 = {}.ElSwitch;
    /** @type {[typeof __VLS_components.ElSwitch, typeof __VLS_components.elSwitch, ]} */ ;
    // @ts-ignore
    const __VLS_177 = __VLS_asFunctionalComponent(__VLS_176, new __VLS_176({
        ...{ 'onChange': {} },
        modelValue: (__VLS_ctx.form.notify_on_trainer_update),
    }));
    const __VLS_178 = __VLS_177({
        ...{ 'onChange': {} },
        modelValue: (__VLS_ctx.form.notify_on_trainer_update),
    }, ...__VLS_functionalComponentArgsRest(__VLS_177));
    let __VLS_180;
    let __VLS_181;
    let __VLS_182;
    const __VLS_183 = {
        onChange: (__VLS_ctx.saveSettingsSilently)
    };
    var __VLS_179;
}
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "settings-card" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.h3, __VLS_intrinsicElements.h3)({
    ...{ class: "settings-title" },
});
(__VLS_ctx.t('settings.sectionLanguage'));
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "language-selector-wrap" },
});
const __VLS_184 = {}.ElRadioGroup;
/** @type {[typeof __VLS_components.ElRadioGroup, typeof __VLS_components.elRadioGroup, typeof __VLS_components.ElRadioGroup, typeof __VLS_components.elRadioGroup, ]} */ ;
// @ts-ignore
const __VLS_185 = __VLS_asFunctionalComponent(__VLS_184, new __VLS_184({
    ...{ 'onChange': {} },
    modelValue: (__VLS_ctx.selectedLanguage),
    size: "default",
    ...{ class: "custom-radio-group" },
}));
const __VLS_186 = __VLS_185({
    ...{ 'onChange': {} },
    modelValue: (__VLS_ctx.selectedLanguage),
    size: "default",
    ...{ class: "custom-radio-group" },
}, ...__VLS_functionalComponentArgsRest(__VLS_185));
let __VLS_188;
let __VLS_189;
let __VLS_190;
const __VLS_191 = {
    onChange: (__VLS_ctx.onLanguageChange)
};
__VLS_187.slots.default;
const __VLS_192 = {}.ElRadioButton;
/** @type {[typeof __VLS_components.ElRadioButton, typeof __VLS_components.elRadioButton, typeof __VLS_components.ElRadioButton, typeof __VLS_components.elRadioButton, ]} */ ;
// @ts-ignore
const __VLS_193 = __VLS_asFunctionalComponent(__VLS_192, new __VLS_192({
    value: "zh-CN",
}));
const __VLS_194 = __VLS_193({
    value: "zh-CN",
}, ...__VLS_functionalComponentArgsRest(__VLS_193));
__VLS_195.slots.default;
(__VLS_ctx.t('settings.langZhCN'));
var __VLS_195;
const __VLS_196 = {}.ElRadioButton;
/** @type {[typeof __VLS_components.ElRadioButton, typeof __VLS_components.elRadioButton, typeof __VLS_components.ElRadioButton, typeof __VLS_components.elRadioButton, ]} */ ;
// @ts-ignore
const __VLS_197 = __VLS_asFunctionalComponent(__VLS_196, new __VLS_196({
    value: "zh-TW",
}));
const __VLS_198 = __VLS_197({
    value: "zh-TW",
}, ...__VLS_functionalComponentArgsRest(__VLS_197));
__VLS_199.slots.default;
(__VLS_ctx.t('settings.langZhTW'));
var __VLS_199;
const __VLS_200 = {}.ElRadioButton;
/** @type {[typeof __VLS_components.ElRadioButton, typeof __VLS_components.elRadioButton, typeof __VLS_components.ElRadioButton, typeof __VLS_components.elRadioButton, ]} */ ;
// @ts-ignore
const __VLS_201 = __VLS_asFunctionalComponent(__VLS_200, new __VLS_200({
    value: "en-US",
}));
const __VLS_202 = __VLS_201({
    value: "en-US",
}, ...__VLS_functionalComponentArgsRest(__VLS_201));
__VLS_203.slots.default;
(__VLS_ctx.t('settings.langEnUS'));
var __VLS_203;
var __VLS_187;
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "settings-card" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.h3, __VLS_intrinsicElements.h3)({
    ...{ class: "settings-title" },
});
(__VLS_ctx.t('settings.sectionTheme'));
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "theme-selector-wrap" },
});
const __VLS_204 = {}.ElRadioGroup;
/** @type {[typeof __VLS_components.ElRadioGroup, typeof __VLS_components.elRadioGroup, typeof __VLS_components.ElRadioGroup, typeof __VLS_components.elRadioGroup, ]} */ ;
// @ts-ignore
const __VLS_205 = __VLS_asFunctionalComponent(__VLS_204, new __VLS_204({
    ...{ 'onChange': {} },
    modelValue: (__VLS_ctx.selectedTheme),
    size: "default",
    ...{ class: "custom-radio-group" },
}));
const __VLS_206 = __VLS_205({
    ...{ 'onChange': {} },
    modelValue: (__VLS_ctx.selectedTheme),
    size: "default",
    ...{ class: "custom-radio-group" },
}, ...__VLS_functionalComponentArgsRest(__VLS_205));
let __VLS_208;
let __VLS_209;
let __VLS_210;
const __VLS_211 = {
    onChange: (__VLS_ctx.onThemeChange)
};
__VLS_207.slots.default;
const __VLS_212 = {}.ElRadioButton;
/** @type {[typeof __VLS_components.ElRadioButton, typeof __VLS_components.elRadioButton, typeof __VLS_components.ElRadioButton, typeof __VLS_components.elRadioButton, ]} */ ;
// @ts-ignore
const __VLS_213 = __VLS_asFunctionalComponent(__VLS_212, new __VLS_212({
    value: "dark",
}));
const __VLS_214 = __VLS_213({
    value: "dark",
}, ...__VLS_functionalComponentArgsRest(__VLS_213));
__VLS_215.slots.default;
const __VLS_216 = {}.ElIcon;
/** @type {[typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, ]} */ ;
// @ts-ignore
const __VLS_217 = __VLS_asFunctionalComponent(__VLS_216, new __VLS_216({
    ...{ class: "theme-btn-icon" },
}));
const __VLS_218 = __VLS_217({
    ...{ class: "theme-btn-icon" },
}, ...__VLS_functionalComponentArgsRest(__VLS_217));
__VLS_219.slots.default;
const __VLS_220 = {}.Moon;
/** @type {[typeof __VLS_components.Moon, ]} */ ;
// @ts-ignore
const __VLS_221 = __VLS_asFunctionalComponent(__VLS_220, new __VLS_220({}));
const __VLS_222 = __VLS_221({}, ...__VLS_functionalComponentArgsRest(__VLS_221));
var __VLS_219;
(__VLS_ctx.t('settings.themeDark'));
var __VLS_215;
const __VLS_224 = {}.ElRadioButton;
/** @type {[typeof __VLS_components.ElRadioButton, typeof __VLS_components.elRadioButton, typeof __VLS_components.ElRadioButton, typeof __VLS_components.elRadioButton, ]} */ ;
// @ts-ignore
const __VLS_225 = __VLS_asFunctionalComponent(__VLS_224, new __VLS_224({
    value: "light",
}));
const __VLS_226 = __VLS_225({
    value: "light",
}, ...__VLS_functionalComponentArgsRest(__VLS_225));
__VLS_227.slots.default;
const __VLS_228 = {}.ElIcon;
/** @type {[typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, ]} */ ;
// @ts-ignore
const __VLS_229 = __VLS_asFunctionalComponent(__VLS_228, new __VLS_228({
    ...{ class: "theme-btn-icon" },
}));
const __VLS_230 = __VLS_229({
    ...{ class: "theme-btn-icon" },
}, ...__VLS_functionalComponentArgsRest(__VLS_229));
__VLS_231.slots.default;
const __VLS_232 = {}.Sunny;
/** @type {[typeof __VLS_components.Sunny, ]} */ ;
// @ts-ignore
const __VLS_233 = __VLS_asFunctionalComponent(__VLS_232, new __VLS_232({}));
const __VLS_234 = __VLS_233({}, ...__VLS_functionalComponentArgsRest(__VLS_233));
var __VLS_231;
(__VLS_ctx.t('settings.themeLight'));
var __VLS_227;
var __VLS_207;
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "settings-card" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.h3, __VLS_intrinsicElements.h3)({
    ...{ class: "settings-title" },
});
(__VLS_ctx.t('settings.sectionBackup'));
if (__VLS_ctx.form) {
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "backup-section" },
    });
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "setting-item-row mb-4" },
    });
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "item-text" },
    });
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "item-label" },
    });
    (__VLS_ctx.t('settings.autoBackup'));
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "item-desc" },
    });
    (__VLS_ctx.t('settings.autoBackupDesc'));
    const __VLS_236 = {}.ElSwitch;
    /** @type {[typeof __VLS_components.ElSwitch, typeof __VLS_components.elSwitch, ]} */ ;
    // @ts-ignore
    const __VLS_237 = __VLS_asFunctionalComponent(__VLS_236, new __VLS_236({
        ...{ 'onChange': {} },
        modelValue: (__VLS_ctx.form.auto_backup_config),
    }));
    const __VLS_238 = __VLS_237({
        ...{ 'onChange': {} },
        modelValue: (__VLS_ctx.form.auto_backup_config),
    }, ...__VLS_functionalComponentArgsRest(__VLS_237));
    let __VLS_240;
    let __VLS_241;
    let __VLS_242;
    const __VLS_243 = {
        onChange: (__VLS_ctx.saveSettingsSilently)
    };
    var __VLS_239;
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "backup-actions" },
    });
    const __VLS_244 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_245 = __VLS_asFunctionalComponent(__VLS_244, new __VLS_244({
        ...{ 'onClick': {} },
        plain: true,
    }));
    const __VLS_246 = __VLS_245({
        ...{ 'onClick': {} },
        plain: true,
    }, ...__VLS_functionalComponentArgsRest(__VLS_245));
    let __VLS_248;
    let __VLS_249;
    let __VLS_250;
    const __VLS_251 = {
        onClick: (__VLS_ctx.exportConfig)
    };
    __VLS_247.slots.default;
    const __VLS_252 = {}.ElIcon;
    /** @type {[typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, ]} */ ;
    // @ts-ignore
    const __VLS_253 = __VLS_asFunctionalComponent(__VLS_252, new __VLS_252({}));
    const __VLS_254 = __VLS_253({}, ...__VLS_functionalComponentArgsRest(__VLS_253));
    __VLS_255.slots.default;
    const __VLS_256 = {}.Download;
    /** @type {[typeof __VLS_components.Download, ]} */ ;
    // @ts-ignore
    const __VLS_257 = __VLS_asFunctionalComponent(__VLS_256, new __VLS_256({}));
    const __VLS_258 = __VLS_257({}, ...__VLS_functionalComponentArgsRest(__VLS_257));
    var __VLS_255;
    (__VLS_ctx.t('settings.exportConfig'));
    var __VLS_247;
    const __VLS_260 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_261 = __VLS_asFunctionalComponent(__VLS_260, new __VLS_260({
        ...{ 'onClick': {} },
        plain: true,
    }));
    const __VLS_262 = __VLS_261({
        ...{ 'onClick': {} },
        plain: true,
    }, ...__VLS_functionalComponentArgsRest(__VLS_261));
    let __VLS_264;
    let __VLS_265;
    let __VLS_266;
    const __VLS_267 = {
        onClick: (__VLS_ctx.importConfig)
    };
    __VLS_263.slots.default;
    const __VLS_268 = {}.ElIcon;
    /** @type {[typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, ]} */ ;
    // @ts-ignore
    const __VLS_269 = __VLS_asFunctionalComponent(__VLS_268, new __VLS_268({}));
    const __VLS_270 = __VLS_269({}, ...__VLS_functionalComponentArgsRest(__VLS_269));
    __VLS_271.slots.default;
    const __VLS_272 = {}.Upload;
    /** @type {[typeof __VLS_components.Upload, ]} */ ;
    // @ts-ignore
    const __VLS_273 = __VLS_asFunctionalComponent(__VLS_272, new __VLS_272({}));
    const __VLS_274 = __VLS_273({}, ...__VLS_functionalComponentArgsRest(__VLS_273));
    var __VLS_271;
    (__VLS_ctx.t('settings.importConfig'));
    var __VLS_263;
    const __VLS_276 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_277 = __VLS_asFunctionalComponent(__VLS_276, new __VLS_276({
        ...{ 'onClick': {} },
        type: "danger",
        plain: true,
    }));
    const __VLS_278 = __VLS_277({
        ...{ 'onClick': {} },
        type: "danger",
        plain: true,
    }, ...__VLS_functionalComponentArgsRest(__VLS_277));
    let __VLS_280;
    let __VLS_281;
    let __VLS_282;
    const __VLS_283 = {
        onClick: (__VLS_ctx.confirmResetConfig)
    };
    __VLS_279.slots.default;
    const __VLS_284 = {}.ElIcon;
    /** @type {[typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, ]} */ ;
    // @ts-ignore
    const __VLS_285 = __VLS_asFunctionalComponent(__VLS_284, new __VLS_284({}));
    const __VLS_286 = __VLS_285({}, ...__VLS_functionalComponentArgsRest(__VLS_285));
    __VLS_287.slots.default;
    const __VLS_288 = {}.RefreshLeft;
    /** @type {[typeof __VLS_components.RefreshLeft, ]} */ ;
    // @ts-ignore
    const __VLS_289 = __VLS_asFunctionalComponent(__VLS_288, new __VLS_288({}));
    const __VLS_290 = __VLS_289({}, ...__VLS_functionalComponentArgsRest(__VLS_289));
    var __VLS_287;
    (__VLS_ctx.t('settings.resetConfig'));
    var __VLS_279;
}
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "settings-card" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.h3, __VLS_intrinsicElements.h3)({
    ...{ class: "settings-title" },
});
(__VLS_ctx.t('settings.sectionDiagnostics'));
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "diagnostics-actions" },
});
const __VLS_292 = {}.ElButton;
/** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
// @ts-ignore
const __VLS_293 = __VLS_asFunctionalComponent(__VLS_292, new __VLS_292({
    ...{ 'onClick': {} },
    type: "primary",
    plain: true,
}));
const __VLS_294 = __VLS_293({
    ...{ 'onClick': {} },
    type: "primary",
    plain: true,
}, ...__VLS_functionalComponentArgsRest(__VLS_293));
let __VLS_296;
let __VLS_297;
let __VLS_298;
const __VLS_299 = {
    onClick: (__VLS_ctx.openDiagnosticsModal)
};
__VLS_295.slots.default;
const __VLS_300 = {}.ElIcon;
/** @type {[typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, ]} */ ;
// @ts-ignore
const __VLS_301 = __VLS_asFunctionalComponent(__VLS_300, new __VLS_300({}));
const __VLS_302 = __VLS_301({}, ...__VLS_functionalComponentArgsRest(__VLS_301));
__VLS_303.slots.default;
const __VLS_304 = {}.Connection;
/** @type {[typeof __VLS_components.Connection, ]} */ ;
// @ts-ignore
const __VLS_305 = __VLS_asFunctionalComponent(__VLS_304, new __VLS_304({}));
const __VLS_306 = __VLS_305({}, ...__VLS_functionalComponentArgsRest(__VLS_305));
var __VLS_303;
(__VLS_ctx.t('settings.runDiagnostics'));
var __VLS_295;
const __VLS_308 = {}.ElButton;
/** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
// @ts-ignore
const __VLS_309 = __VLS_asFunctionalComponent(__VLS_308, new __VLS_308({
    ...{ 'onClick': {} },
    plain: true,
}));
const __VLS_310 = __VLS_309({
    ...{ 'onClick': {} },
    plain: true,
}, ...__VLS_functionalComponentArgsRest(__VLS_309));
let __VLS_312;
let __VLS_313;
let __VLS_314;
const __VLS_315 = {
    onClick: (__VLS_ctx.openLogsModal)
};
__VLS_311.slots.default;
const __VLS_316 = {}.ElIcon;
/** @type {[typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, ]} */ ;
// @ts-ignore
const __VLS_317 = __VLS_asFunctionalComponent(__VLS_316, new __VLS_316({}));
const __VLS_318 = __VLS_317({}, ...__VLS_functionalComponentArgsRest(__VLS_317));
__VLS_319.slots.default;
const __VLS_320 = {}.Document;
/** @type {[typeof __VLS_components.Document, ]} */ ;
// @ts-ignore
const __VLS_321 = __VLS_asFunctionalComponent(__VLS_320, new __VLS_320({}));
const __VLS_322 = __VLS_321({}, ...__VLS_functionalComponentArgsRest(__VLS_321));
var __VLS_319;
(__VLS_ctx.t('settings.viewLogs'));
var __VLS_311;
const __VLS_324 = {}.ElButton;
/** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
// @ts-ignore
const __VLS_325 = __VLS_asFunctionalComponent(__VLS_324, new __VLS_324({
    ...{ 'onClick': {} },
    plain: true,
}));
const __VLS_326 = __VLS_325({
    ...{ 'onClick': {} },
    plain: true,
}, ...__VLS_functionalComponentArgsRest(__VLS_325));
let __VLS_328;
let __VLS_329;
let __VLS_330;
const __VLS_331 = {
    onClick: (__VLS_ctx.exportDiagnosticsReport)
};
__VLS_327.slots.default;
const __VLS_332 = {}.ElIcon;
/** @type {[typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, ]} */ ;
// @ts-ignore
const __VLS_333 = __VLS_asFunctionalComponent(__VLS_332, new __VLS_332({}));
const __VLS_334 = __VLS_333({}, ...__VLS_functionalComponentArgsRest(__VLS_333));
__VLS_335.slots.default;
const __VLS_336 = {}.Share;
/** @type {[typeof __VLS_components.Share, ]} */ ;
// @ts-ignore
const __VLS_337 = __VLS_asFunctionalComponent(__VLS_336, new __VLS_336({}));
const __VLS_338 = __VLS_337({}, ...__VLS_functionalComponentArgsRest(__VLS_337));
var __VLS_335;
(__VLS_ctx.t('settings.exportDiagnostics'));
var __VLS_327;
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "actions-card" },
});
const __VLS_340 = {}.ElButton;
/** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
// @ts-ignore
const __VLS_341 = __VLS_asFunctionalComponent(__VLS_340, new __VLS_340({
    ...{ 'onClick': {} },
    type: "primary",
    size: "large",
    loading: (__VLS_ctx.saving),
}));
const __VLS_342 = __VLS_341({
    ...{ 'onClick': {} },
    type: "primary",
    size: "large",
    loading: (__VLS_ctx.saving),
}, ...__VLS_functionalComponentArgsRest(__VLS_341));
let __VLS_344;
let __VLS_345;
let __VLS_346;
const __VLS_347 = {
    onClick: (__VLS_ctx.saveSettings)
};
__VLS_343.slots.default;
(__VLS_ctx.t('settings.saveBtn'));
var __VLS_343;
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "settings-card about-card" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "about-header" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.h3, __VLS_intrinsicElements.h3)({
    ...{ class: "settings-title" },
});
(__VLS_ctx.t('settings.sectionAbout'));
const __VLS_348 = {}.ElButton;
/** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
// @ts-ignore
const __VLS_349 = __VLS_asFunctionalComponent(__VLS_348, new __VLS_348({
    ...{ 'onClick': {} },
    size: "small",
    plain: true,
    loading: (__VLS_ctx.checkingUpdate),
}));
const __VLS_350 = __VLS_349({
    ...{ 'onClick': {} },
    size: "small",
    plain: true,
    loading: (__VLS_ctx.checkingUpdate),
}, ...__VLS_functionalComponentArgsRest(__VLS_349));
let __VLS_352;
let __VLS_353;
let __VLS_354;
const __VLS_355 = {
    onClick: (__VLS_ctx.checkUpdates)
};
__VLS_351.slots.default;
const __VLS_356 = {}.ElIcon;
/** @type {[typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, ]} */ ;
// @ts-ignore
const __VLS_357 = __VLS_asFunctionalComponent(__VLS_356, new __VLS_356({}));
const __VLS_358 = __VLS_357({}, ...__VLS_functionalComponentArgsRest(__VLS_357));
__VLS_359.slots.default;
const __VLS_360 = {}.Refresh;
/** @type {[typeof __VLS_components.Refresh, ]} */ ;
// @ts-ignore
const __VLS_361 = __VLS_asFunctionalComponent(__VLS_360, new __VLS_360({}));
const __VLS_362 = __VLS_361({}, ...__VLS_functionalComponentArgsRest(__VLS_361));
var __VLS_359;
(__VLS_ctx.t('settings.checkUpdate'));
var __VLS_351;
__VLS_asFunctionalElement(__VLS_intrinsicElements.p, __VLS_intrinsicElements.p)({
    ...{ class: "about-p" },
});
(__VLS_ctx.t('settings.aboutVersion'));
__VLS_asFunctionalElement(__VLS_intrinsicElements.p, __VLS_intrinsicElements.p)({
    ...{ class: "about-p" },
});
(__VLS_ctx.t('settings.aboutSource'));
__VLS_asFunctionalElement(__VLS_intrinsicElements.p, __VLS_intrinsicElements.p)({
    ...{ class: "about-p text-gray" },
});
(__VLS_ctx.t('settings.aboutDisclaimer'));
const __VLS_364 = {}.ElDialog;
/** @type {[typeof __VLS_components.ElDialog, typeof __VLS_components.elDialog, typeof __VLS_components.ElDialog, typeof __VLS_components.elDialog, ]} */ ;
// @ts-ignore
const __VLS_365 = __VLS_asFunctionalComponent(__VLS_364, new __VLS_364({
    modelValue: (__VLS_ctx.diagDialogVisible),
    title: (__VLS_ctx.t('settings.diagnosticsTitle')),
    width: "640px",
    appendToBody: true,
}));
const __VLS_366 = __VLS_365({
    modelValue: (__VLS_ctx.diagDialogVisible),
    title: (__VLS_ctx.t('settings.diagnosticsTitle')),
    width: "640px",
    appendToBody: true,
}, ...__VLS_functionalComponentArgsRest(__VLS_365));
__VLS_367.slots.default;
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "diag-modal-body" },
});
__VLS_asFunctionalDirective(__VLS_directives.vLoading)(null, { ...__VLS_directiveBindingRestFields, value: (__VLS_ctx.runningDiag) }, null, null);
if (__VLS_ctx.diagReport) {
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "diag-report-grid" },
    });
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "diag-item" },
    });
    __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
        ...{ class: "diag-label" },
    });
    __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
        ...{ class: "diag-val" },
    });
    (__VLS_ctx.diagReport.os);
    (__VLS_ctx.diagReport.arch);
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "diag-item" },
    });
    __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
        ...{ class: "diag-label" },
    });
    const __VLS_368 = {}.ElTag;
    /** @type {[typeof __VLS_components.ElTag, typeof __VLS_components.elTag, typeof __VLS_components.ElTag, typeof __VLS_components.elTag, ]} */ ;
    // @ts-ignore
    const __VLS_369 = __VLS_asFunctionalComponent(__VLS_368, new __VLS_368({
        size: "small",
        type: (__VLS_ctx.diagReport.download_dir_writable ? 'success' : 'danger'),
    }));
    const __VLS_370 = __VLS_369({
        size: "small",
        type: (__VLS_ctx.diagReport.download_dir_writable ? 'success' : 'danger'),
    }, ...__VLS_functionalComponentArgsRest(__VLS_369));
    __VLS_371.slots.default;
    (__VLS_ctx.diagReport.download_dir_writable ? '正常' : '不可写');
    var __VLS_371;
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "diag-item" },
    });
    __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
        ...{ class: "diag-label" },
    });
    __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
        ...{ class: "diag-val" },
    });
    ((__VLS_ctx.diagReport.available_disk_bytes / 1024 / 1024 / 1024).toFixed(2));
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "diag-item" },
    });
    __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
        ...{ class: "diag-label" },
    });
    const __VLS_372 = {}.ElTag;
    /** @type {[typeof __VLS_components.ElTag, typeof __VLS_components.elTag, typeof __VLS_components.ElTag, typeof __VLS_components.elTag, ]} */ ;
    // @ts-ignore
    const __VLS_373 = __VLS_asFunctionalComponent(__VLS_372, new __VLS_372({
        size: "small",
        type: (__VLS_ctx.diagReport.db_integrity_ok ? 'success' : 'danger'),
    }));
    const __VLS_374 = __VLS_373({
        size: "small",
        type: (__VLS_ctx.diagReport.db_integrity_ok ? 'success' : 'danger'),
    }, ...__VLS_functionalComponentArgsRest(__VLS_373));
    __VLS_375.slots.default;
    (__VLS_ctx.diagReport.db_integrity_ok ? '完整正常 (OK)' : '数据库损坏');
    var __VLS_375;
    if (__VLS_ctx.diagReport.network) {
        __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
            ...{ class: "diag-item full-width" },
        });
        __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
            ...{ class: "diag-label" },
        });
        __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
            ...{ class: "diag-val" },
        });
        (__VLS_ctx.diagReport.network.dns_resolved ? '已解析' : '失败');
        (__VLS_ctx.diagReport.network.is_reachable ? '可达 (' + __VLS_ctx.diagReport.network.http_status + ')' : '异常');
        (__VLS_ctx.diagReport.network.latency_ms);
    }
}
{
    const { footer: __VLS_thisSlot } = __VLS_367.slots;
    const __VLS_376 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_377 = __VLS_asFunctionalComponent(__VLS_376, new __VLS_376({
        ...{ 'onClick': {} },
    }));
    const __VLS_378 = __VLS_377({
        ...{ 'onClick': {} },
    }, ...__VLS_functionalComponentArgsRest(__VLS_377));
    let __VLS_380;
    let __VLS_381;
    let __VLS_382;
    const __VLS_383 = {
        onClick: (...[$event]) => {
            __VLS_ctx.diagDialogVisible = false;
        }
    };
    __VLS_379.slots.default;
    (__VLS_ctx.t('library.cancel'));
    var __VLS_379;
    const __VLS_384 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_385 = __VLS_asFunctionalComponent(__VLS_384, new __VLS_384({
        ...{ 'onClick': {} },
        type: "primary",
    }));
    const __VLS_386 = __VLS_385({
        ...{ 'onClick': {} },
        type: "primary",
    }, ...__VLS_functionalComponentArgsRest(__VLS_385));
    let __VLS_388;
    let __VLS_389;
    let __VLS_390;
    const __VLS_391 = {
        onClick: (__VLS_ctx.exportDiagnosticsReport)
    };
    __VLS_387.slots.default;
    (__VLS_ctx.t('settings.exportDiagnostics'));
    var __VLS_387;
}
var __VLS_367;
const __VLS_392 = {}.ElDialog;
/** @type {[typeof __VLS_components.ElDialog, typeof __VLS_components.elDialog, typeof __VLS_components.ElDialog, typeof __VLS_components.elDialog, ]} */ ;
// @ts-ignore
const __VLS_393 = __VLS_asFunctionalComponent(__VLS_392, new __VLS_392({
    modelValue: (__VLS_ctx.logsDialogVisible),
    title: (__VLS_ctx.t('settings.logsTitle')),
    width: "780px",
    appendToBody: true,
}));
const __VLS_394 = __VLS_393({
    modelValue: (__VLS_ctx.logsDialogVisible),
    title: (__VLS_ctx.t('settings.logsTitle')),
    width: "780px",
    appendToBody: true,
}, ...__VLS_functionalComponentArgsRest(__VLS_393));
__VLS_395.slots.default;
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "logs-filter-bar" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
    ...{ class: "filter-label" },
});
(__VLS_ctx.t('settings.logFilterCategory'));
const __VLS_396 = {}.ElRadioGroup;
/** @type {[typeof __VLS_components.ElRadioGroup, typeof __VLS_components.elRadioGroup, typeof __VLS_components.ElRadioGroup, typeof __VLS_components.elRadioGroup, ]} */ ;
// @ts-ignore
const __VLS_397 = __VLS_asFunctionalComponent(__VLS_396, new __VLS_396({
    ...{ 'onChange': {} },
    modelValue: (__VLS_ctx.logCategory),
    size: "small",
}));
const __VLS_398 = __VLS_397({
    ...{ 'onChange': {} },
    modelValue: (__VLS_ctx.logCategory),
    size: "small",
}, ...__VLS_functionalComponentArgsRest(__VLS_397));
let __VLS_400;
let __VLS_401;
let __VLS_402;
const __VLS_403 = {
    onChange: (__VLS_ctx.refreshLogs)
};
__VLS_399.slots.default;
const __VLS_404 = {}.ElRadioButton;
/** @type {[typeof __VLS_components.ElRadioButton, typeof __VLS_components.elRadioButton, typeof __VLS_components.ElRadioButton, typeof __VLS_components.elRadioButton, ]} */ ;
// @ts-ignore
const __VLS_405 = __VLS_asFunctionalComponent(__VLS_404, new __VLS_404({
    value: "",
}));
const __VLS_406 = __VLS_405({
    value: "",
}, ...__VLS_functionalComponentArgsRest(__VLS_405));
__VLS_407.slots.default;
(__VLS_ctx.t('settings.logAll'));
var __VLS_407;
const __VLS_408 = {}.ElRadioButton;
/** @type {[typeof __VLS_components.ElRadioButton, typeof __VLS_components.elRadioButton, typeof __VLS_components.ElRadioButton, typeof __VLS_components.elRadioButton, ]} */ ;
// @ts-ignore
const __VLS_409 = __VLS_asFunctionalComponent(__VLS_408, new __VLS_408({
    value: "download",
}));
const __VLS_410 = __VLS_409({
    value: "download",
}, ...__VLS_functionalComponentArgsRest(__VLS_409));
__VLS_411.slots.default;
var __VLS_411;
const __VLS_412 = {}.ElRadioButton;
/** @type {[typeof __VLS_components.ElRadioButton, typeof __VLS_components.elRadioButton, typeof __VLS_components.ElRadioButton, typeof __VLS_components.elRadioButton, ]} */ ;
// @ts-ignore
const __VLS_413 = __VLS_asFunctionalComponent(__VLS_412, new __VLS_412({
    value: "launch",
}));
const __VLS_414 = __VLS_413({
    value: "launch",
}, ...__VLS_functionalComponentArgsRest(__VLS_413));
__VLS_415.slots.default;
var __VLS_415;
const __VLS_416 = {}.ElRadioButton;
/** @type {[typeof __VLS_components.ElRadioButton, typeof __VLS_components.elRadioButton, typeof __VLS_components.ElRadioButton, typeof __VLS_components.elRadioButton, ]} */ ;
// @ts-ignore
const __VLS_417 = __VLS_asFunctionalComponent(__VLS_416, new __VLS_416({
    value: "search",
}));
const __VLS_418 = __VLS_417({
    value: "search",
}, ...__VLS_functionalComponentArgsRest(__VLS_417));
__VLS_419.slots.default;
var __VLS_419;
const __VLS_420 = {}.ElRadioButton;
/** @type {[typeof __VLS_components.ElRadioButton, typeof __VLS_components.elRadioButton, typeof __VLS_components.ElRadioButton, typeof __VLS_components.elRadioButton, ]} */ ;
// @ts-ignore
const __VLS_421 = __VLS_asFunctionalComponent(__VLS_420, new __VLS_420({
    value: "system",
}));
const __VLS_422 = __VLS_421({
    value: "system",
}, ...__VLS_functionalComponentArgsRest(__VLS_421));
__VLS_423.slots.default;
var __VLS_423;
var __VLS_399;
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "log-entries-box" },
});
for (const [entry, idx] of __VLS_getVForSourceType((__VLS_ctx.logEntries))) {
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        key: (idx),
        ...{ class: "log-row" },
    });
    __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
        ...{ class: "log-time" },
    });
    (entry.timestamp);
    __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
        ...{ class: "log-level" },
        ...{ class: (`lvl-${entry.level.toLowerCase()}`) },
    });
    (entry.level);
    __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
        ...{ class: "log-cat" },
    });
    (entry.category);
    __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
        ...{ class: "log-msg" },
    });
    (entry.message);
}
if (__VLS_ctx.logEntries.length === 0) {
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "no-logs" },
    });
}
{
    const { footer: __VLS_thisSlot } = __VLS_395.slots;
    const __VLS_424 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_425 = __VLS_asFunctionalComponent(__VLS_424, new __VLS_424({
        ...{ 'onClick': {} },
    }));
    const __VLS_426 = __VLS_425({
        ...{ 'onClick': {} },
    }, ...__VLS_functionalComponentArgsRest(__VLS_425));
    let __VLS_428;
    let __VLS_429;
    let __VLS_430;
    const __VLS_431 = {
        onClick: (...[$event]) => {
            __VLS_ctx.logsDialogVisible = false;
        }
    };
    __VLS_427.slots.default;
    (__VLS_ctx.t('library.cancel'));
    var __VLS_427;
}
var __VLS_395;
/** @type {__VLS_StyleScopedClasses['settings-view']} */ ;
/** @type {__VLS_StyleScopedClasses['settings-card']} */ ;
/** @type {__VLS_StyleScopedClasses['settings-title']} */ ;
/** @type {__VLS_StyleScopedClasses['settings-form']} */ ;
/** @type {__VLS_StyleScopedClasses['field-label-with-tip']} */ ;
/** @type {__VLS_StyleScopedClasses['tip-icon']} */ ;
/** @type {__VLS_StyleScopedClasses['path-display-card']} */ ;
/** @type {__VLS_StyleScopedClasses['path-display-text']} */ ;
/** @type {__VLS_StyleScopedClasses['folder-icon']} */ ;
/** @type {__VLS_StyleScopedClasses['path-text']} */ ;
/** @type {__VLS_StyleScopedClasses['change-path-btn']} */ ;
/** @type {__VLS_StyleScopedClasses['slider-horizontal-row']} */ ;
/** @type {__VLS_StyleScopedClasses['slider-container']} */ ;
/** @type {__VLS_StyleScopedClasses['slider-val-tag']} */ ;
/** @type {__VLS_StyleScopedClasses['slider-hint-text']} */ ;
/** @type {__VLS_StyleScopedClasses['switch-row']} */ ;
/** @type {__VLS_StyleScopedClasses['switch-extra-hint']} */ ;
/** @type {__VLS_StyleScopedClasses['settings-card']} */ ;
/** @type {__VLS_StyleScopedClasses['settings-title']} */ ;
/** @type {__VLS_StyleScopedClasses['window-settings-grid']} */ ;
/** @type {__VLS_StyleScopedClasses['setting-item-row']} */ ;
/** @type {__VLS_StyleScopedClasses['item-text']} */ ;
/** @type {__VLS_StyleScopedClasses['item-label']} */ ;
/** @type {__VLS_StyleScopedClasses['item-desc']} */ ;
/** @type {__VLS_StyleScopedClasses['inner-divider']} */ ;
/** @type {__VLS_StyleScopedClasses['setting-item-row']} */ ;
/** @type {__VLS_StyleScopedClasses['item-text']} */ ;
/** @type {__VLS_StyleScopedClasses['item-label']} */ ;
/** @type {__VLS_StyleScopedClasses['item-desc']} */ ;
/** @type {__VLS_StyleScopedClasses['inner-divider']} */ ;
/** @type {__VLS_StyleScopedClasses['setting-item-row']} */ ;
/** @type {__VLS_StyleScopedClasses['item-text']} */ ;
/** @type {__VLS_StyleScopedClasses['item-label']} */ ;
/** @type {__VLS_StyleScopedClasses['item-desc']} */ ;
/** @type {__VLS_StyleScopedClasses['settings-card']} */ ;
/** @type {__VLS_StyleScopedClasses['settings-title']} */ ;
/** @type {__VLS_StyleScopedClasses['hotkey-settings-grid']} */ ;
/** @type {__VLS_StyleScopedClasses['setting-item-row']} */ ;
/** @type {__VLS_StyleScopedClasses['item-text']} */ ;
/** @type {__VLS_StyleScopedClasses['item-label']} */ ;
/** @type {__VLS_StyleScopedClasses['item-desc']} */ ;
/** @type {__VLS_StyleScopedClasses['inner-divider']} */ ;
/** @type {__VLS_StyleScopedClasses['hotkey-inputs-row']} */ ;
/** @type {__VLS_StyleScopedClasses['hotkey-field']} */ ;
/** @type {__VLS_StyleScopedClasses['field-label']} */ ;
/** @type {__VLS_StyleScopedClasses['hotkey-badge']} */ ;
/** @type {__VLS_StyleScopedClasses['hotkey-field']} */ ;
/** @type {__VLS_StyleScopedClasses['field-label']} */ ;
/** @type {__VLS_StyleScopedClasses['hotkey-badge']} */ ;
/** @type {__VLS_StyleScopedClasses['settings-card']} */ ;
/** @type {__VLS_StyleScopedClasses['settings-title']} */ ;
/** @type {__VLS_StyleScopedClasses['window-settings-grid']} */ ;
/** @type {__VLS_StyleScopedClasses['setting-item-row']} */ ;
/** @type {__VLS_StyleScopedClasses['item-text']} */ ;
/** @type {__VLS_StyleScopedClasses['item-label']} */ ;
/** @type {__VLS_StyleScopedClasses['item-desc']} */ ;
/** @type {__VLS_StyleScopedClasses['inner-divider']} */ ;
/** @type {__VLS_StyleScopedClasses['setting-item-row']} */ ;
/** @type {__VLS_StyleScopedClasses['item-text']} */ ;
/** @type {__VLS_StyleScopedClasses['item-label']} */ ;
/** @type {__VLS_StyleScopedClasses['item-desc']} */ ;
/** @type {__VLS_StyleScopedClasses['inner-divider']} */ ;
/** @type {__VLS_StyleScopedClasses['setting-item-row']} */ ;
/** @type {__VLS_StyleScopedClasses['item-text']} */ ;
/** @type {__VLS_StyleScopedClasses['item-label']} */ ;
/** @type {__VLS_StyleScopedClasses['item-desc']} */ ;
/** @type {__VLS_StyleScopedClasses['settings-card']} */ ;
/** @type {__VLS_StyleScopedClasses['settings-title']} */ ;
/** @type {__VLS_StyleScopedClasses['language-selector-wrap']} */ ;
/** @type {__VLS_StyleScopedClasses['custom-radio-group']} */ ;
/** @type {__VLS_StyleScopedClasses['settings-card']} */ ;
/** @type {__VLS_StyleScopedClasses['settings-title']} */ ;
/** @type {__VLS_StyleScopedClasses['theme-selector-wrap']} */ ;
/** @type {__VLS_StyleScopedClasses['custom-radio-group']} */ ;
/** @type {__VLS_StyleScopedClasses['theme-btn-icon']} */ ;
/** @type {__VLS_StyleScopedClasses['theme-btn-icon']} */ ;
/** @type {__VLS_StyleScopedClasses['settings-card']} */ ;
/** @type {__VLS_StyleScopedClasses['settings-title']} */ ;
/** @type {__VLS_StyleScopedClasses['backup-section']} */ ;
/** @type {__VLS_StyleScopedClasses['setting-item-row']} */ ;
/** @type {__VLS_StyleScopedClasses['mb-4']} */ ;
/** @type {__VLS_StyleScopedClasses['item-text']} */ ;
/** @type {__VLS_StyleScopedClasses['item-label']} */ ;
/** @type {__VLS_StyleScopedClasses['item-desc']} */ ;
/** @type {__VLS_StyleScopedClasses['backup-actions']} */ ;
/** @type {__VLS_StyleScopedClasses['settings-card']} */ ;
/** @type {__VLS_StyleScopedClasses['settings-title']} */ ;
/** @type {__VLS_StyleScopedClasses['diagnostics-actions']} */ ;
/** @type {__VLS_StyleScopedClasses['actions-card']} */ ;
/** @type {__VLS_StyleScopedClasses['settings-card']} */ ;
/** @type {__VLS_StyleScopedClasses['about-card']} */ ;
/** @type {__VLS_StyleScopedClasses['about-header']} */ ;
/** @type {__VLS_StyleScopedClasses['settings-title']} */ ;
/** @type {__VLS_StyleScopedClasses['about-p']} */ ;
/** @type {__VLS_StyleScopedClasses['about-p']} */ ;
/** @type {__VLS_StyleScopedClasses['about-p']} */ ;
/** @type {__VLS_StyleScopedClasses['text-gray']} */ ;
/** @type {__VLS_StyleScopedClasses['diag-modal-body']} */ ;
/** @type {__VLS_StyleScopedClasses['diag-report-grid']} */ ;
/** @type {__VLS_StyleScopedClasses['diag-item']} */ ;
/** @type {__VLS_StyleScopedClasses['diag-label']} */ ;
/** @type {__VLS_StyleScopedClasses['diag-val']} */ ;
/** @type {__VLS_StyleScopedClasses['diag-item']} */ ;
/** @type {__VLS_StyleScopedClasses['diag-label']} */ ;
/** @type {__VLS_StyleScopedClasses['diag-item']} */ ;
/** @type {__VLS_StyleScopedClasses['diag-label']} */ ;
/** @type {__VLS_StyleScopedClasses['diag-val']} */ ;
/** @type {__VLS_StyleScopedClasses['diag-item']} */ ;
/** @type {__VLS_StyleScopedClasses['diag-label']} */ ;
/** @type {__VLS_StyleScopedClasses['diag-item']} */ ;
/** @type {__VLS_StyleScopedClasses['full-width']} */ ;
/** @type {__VLS_StyleScopedClasses['diag-label']} */ ;
/** @type {__VLS_StyleScopedClasses['diag-val']} */ ;
/** @type {__VLS_StyleScopedClasses['logs-filter-bar']} */ ;
/** @type {__VLS_StyleScopedClasses['filter-label']} */ ;
/** @type {__VLS_StyleScopedClasses['log-entries-box']} */ ;
/** @type {__VLS_StyleScopedClasses['log-row']} */ ;
/** @type {__VLS_StyleScopedClasses['log-time']} */ ;
/** @type {__VLS_StyleScopedClasses['log-level']} */ ;
/** @type {__VLS_StyleScopedClasses['log-cat']} */ ;
/** @type {__VLS_StyleScopedClasses['log-msg']} */ ;
/** @type {__VLS_StyleScopedClasses['no-logs']} */ ;
var __VLS_dollars;
const __VLS_self = (await import('vue')).defineComponent({
    setup() {
        return {
            QuestionFilled: QuestionFilled,
            FolderOpened: FolderOpened,
            FolderAdd: FolderAdd,
            Moon: Moon,
            Sunny: Sunny,
            Download: Download,
            Upload: Upload,
            RefreshLeft: RefreshLeft,
            Connection: Connection,
            Document: Document,
            Share: Share,
            Refresh: Refresh,
            t: t,
            form: form,
            saving: saving,
            selectingDir: selectingDir,
            selectedLanguage: selectedLanguage,
            selectedTheme: selectedTheme,
            diagDialogVisible: diagDialogVisible,
            runningDiag: runningDiag,
            diagReport: diagReport,
            logsDialogVisible: logsDialogVisible,
            logCategory: logCategory,
            logEntries: logEntries,
            checkingUpdate: checkingUpdate,
            openDownloadFolder: openDownloadFolder,
            changeDownloadDirectory: changeDownloadDirectory,
            onSliderChange: onSliderChange,
            onLanguageChange: onLanguageChange,
            onThemeChange: onThemeChange,
            validateHotkey: validateHotkey,
            exportConfig: exportConfig,
            importConfig: importConfig,
            confirmResetConfig: confirmResetConfig,
            openDiagnosticsModal: openDiagnosticsModal,
            exportDiagnosticsReport: exportDiagnosticsReport,
            openLogsModal: openLogsModal,
            refreshLogs: refreshLogs,
            checkUpdates: checkUpdates,
            saveSettingsSilently: saveSettingsSilently,
            saveSettings: saveSettings,
        };
    },
});
export default (await import('vue')).defineComponent({
    setup() {
        return {};
    },
});
; /* PartiallyEnd: #4569/main.vue */
//# sourceMappingURL=SettingsView.vue.js.map