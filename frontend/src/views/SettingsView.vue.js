/// <reference types="../../node_modules/.vue-global-types/vue_3.5_0_0_0.d.ts" />
import { ref, onMounted } from 'vue';
import { ElMessage } from 'element-plus';
import { QuestionFilled, FolderOpened, FolderAdd, Moon, Sunny } from '@element-plus/icons-vue';
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
    // 前端优先快速检查活跃下载任务
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
// 并发滑块发生改变时即时保存
async function onSliderChange() {
    await saveSettingsSilently();
}
// 语言切换处理
async function onLanguageChange(val) {
    const lang = val;
    setLanguage(lang);
    selectedLanguage.value = lang;
    if (form.value) {
        form.value.language = lang;
        await saveSettingsSilently();
    }
}
// 主题切换处理 (即时全局响应并持久化)
async function onThemeChange(val) {
    const mode = val;
    setTheme(mode);
    selectedTheme.value = mode;
    if (form.value) {
        form.value.theme = mode;
        await saveSettingsSilently();
    }
}
// 静默自动保存配置
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
// 手动点击保存配置
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
/** @type {__VLS_StyleScopedClasses['tip-icon']} */ ;
/** @type {__VLS_StyleScopedClasses['path-display-text']} */ ;
/** @type {__VLS_StyleScopedClasses['path-text']} */ ;
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
(__VLS_ctx.t('settings.sectionLanguage'));
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "language-selector-wrap" },
});
const __VLS_124 = {}.ElRadioGroup;
/** @type {[typeof __VLS_components.ElRadioGroup, typeof __VLS_components.elRadioGroup, typeof __VLS_components.ElRadioGroup, typeof __VLS_components.elRadioGroup, ]} */ ;
// @ts-ignore
const __VLS_125 = __VLS_asFunctionalComponent(__VLS_124, new __VLS_124({
    ...{ 'onChange': {} },
    modelValue: (__VLS_ctx.selectedLanguage),
    size: "default",
    ...{ class: "custom-radio-group" },
}));
const __VLS_126 = __VLS_125({
    ...{ 'onChange': {} },
    modelValue: (__VLS_ctx.selectedLanguage),
    size: "default",
    ...{ class: "custom-radio-group" },
}, ...__VLS_functionalComponentArgsRest(__VLS_125));
let __VLS_128;
let __VLS_129;
let __VLS_130;
const __VLS_131 = {
    onChange: (__VLS_ctx.onLanguageChange)
};
__VLS_127.slots.default;
const __VLS_132 = {}.ElRadioButton;
/** @type {[typeof __VLS_components.ElRadioButton, typeof __VLS_components.elRadioButton, typeof __VLS_components.ElRadioButton, typeof __VLS_components.elRadioButton, ]} */ ;
// @ts-ignore
const __VLS_133 = __VLS_asFunctionalComponent(__VLS_132, new __VLS_132({
    value: "zh-CN",
}));
const __VLS_134 = __VLS_133({
    value: "zh-CN",
}, ...__VLS_functionalComponentArgsRest(__VLS_133));
__VLS_135.slots.default;
(__VLS_ctx.t('settings.langZhCN'));
var __VLS_135;
const __VLS_136 = {}.ElRadioButton;
/** @type {[typeof __VLS_components.ElRadioButton, typeof __VLS_components.elRadioButton, typeof __VLS_components.ElRadioButton, typeof __VLS_components.elRadioButton, ]} */ ;
// @ts-ignore
const __VLS_137 = __VLS_asFunctionalComponent(__VLS_136, new __VLS_136({
    value: "zh-TW",
}));
const __VLS_138 = __VLS_137({
    value: "zh-TW",
}, ...__VLS_functionalComponentArgsRest(__VLS_137));
__VLS_139.slots.default;
(__VLS_ctx.t('settings.langZhTW'));
var __VLS_139;
const __VLS_140 = {}.ElRadioButton;
/** @type {[typeof __VLS_components.ElRadioButton, typeof __VLS_components.elRadioButton, typeof __VLS_components.ElRadioButton, typeof __VLS_components.elRadioButton, ]} */ ;
// @ts-ignore
const __VLS_141 = __VLS_asFunctionalComponent(__VLS_140, new __VLS_140({
    value: "en-US",
}));
const __VLS_142 = __VLS_141({
    value: "en-US",
}, ...__VLS_functionalComponentArgsRest(__VLS_141));
__VLS_143.slots.default;
(__VLS_ctx.t('settings.langEnUS'));
var __VLS_143;
var __VLS_127;
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
const __VLS_144 = {}.ElRadioGroup;
/** @type {[typeof __VLS_components.ElRadioGroup, typeof __VLS_components.elRadioGroup, typeof __VLS_components.ElRadioGroup, typeof __VLS_components.elRadioGroup, ]} */ ;
// @ts-ignore
const __VLS_145 = __VLS_asFunctionalComponent(__VLS_144, new __VLS_144({
    ...{ 'onChange': {} },
    modelValue: (__VLS_ctx.selectedTheme),
    size: "default",
    ...{ class: "custom-radio-group" },
}));
const __VLS_146 = __VLS_145({
    ...{ 'onChange': {} },
    modelValue: (__VLS_ctx.selectedTheme),
    size: "default",
    ...{ class: "custom-radio-group" },
}, ...__VLS_functionalComponentArgsRest(__VLS_145));
let __VLS_148;
let __VLS_149;
let __VLS_150;
const __VLS_151 = {
    onChange: (__VLS_ctx.onThemeChange)
};
__VLS_147.slots.default;
const __VLS_152 = {}.ElRadioButton;
/** @type {[typeof __VLS_components.ElRadioButton, typeof __VLS_components.elRadioButton, typeof __VLS_components.ElRadioButton, typeof __VLS_components.elRadioButton, ]} */ ;
// @ts-ignore
const __VLS_153 = __VLS_asFunctionalComponent(__VLS_152, new __VLS_152({
    value: "dark",
}));
const __VLS_154 = __VLS_153({
    value: "dark",
}, ...__VLS_functionalComponentArgsRest(__VLS_153));
__VLS_155.slots.default;
const __VLS_156 = {}.ElIcon;
/** @type {[typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, ]} */ ;
// @ts-ignore
const __VLS_157 = __VLS_asFunctionalComponent(__VLS_156, new __VLS_156({
    ...{ class: "theme-btn-icon" },
}));
const __VLS_158 = __VLS_157({
    ...{ class: "theme-btn-icon" },
}, ...__VLS_functionalComponentArgsRest(__VLS_157));
__VLS_159.slots.default;
const __VLS_160 = {}.Moon;
/** @type {[typeof __VLS_components.Moon, ]} */ ;
// @ts-ignore
const __VLS_161 = __VLS_asFunctionalComponent(__VLS_160, new __VLS_160({}));
const __VLS_162 = __VLS_161({}, ...__VLS_functionalComponentArgsRest(__VLS_161));
var __VLS_159;
(__VLS_ctx.t('settings.themeDark'));
var __VLS_155;
const __VLS_164 = {}.ElRadioButton;
/** @type {[typeof __VLS_components.ElRadioButton, typeof __VLS_components.elRadioButton, typeof __VLS_components.ElRadioButton, typeof __VLS_components.elRadioButton, ]} */ ;
// @ts-ignore
const __VLS_165 = __VLS_asFunctionalComponent(__VLS_164, new __VLS_164({
    value: "light",
}));
const __VLS_166 = __VLS_165({
    value: "light",
}, ...__VLS_functionalComponentArgsRest(__VLS_165));
__VLS_167.slots.default;
const __VLS_168 = {}.ElIcon;
/** @type {[typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, ]} */ ;
// @ts-ignore
const __VLS_169 = __VLS_asFunctionalComponent(__VLS_168, new __VLS_168({
    ...{ class: "theme-btn-icon" },
}));
const __VLS_170 = __VLS_169({
    ...{ class: "theme-btn-icon" },
}, ...__VLS_functionalComponentArgsRest(__VLS_169));
__VLS_171.slots.default;
const __VLS_172 = {}.Sunny;
/** @type {[typeof __VLS_components.Sunny, ]} */ ;
// @ts-ignore
const __VLS_173 = __VLS_asFunctionalComponent(__VLS_172, new __VLS_172({}));
const __VLS_174 = __VLS_173({}, ...__VLS_functionalComponentArgsRest(__VLS_173));
var __VLS_171;
(__VLS_ctx.t('settings.themeLight'));
var __VLS_167;
var __VLS_147;
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "actions-card" },
});
const __VLS_176 = {}.ElButton;
/** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
// @ts-ignore
const __VLS_177 = __VLS_asFunctionalComponent(__VLS_176, new __VLS_176({
    ...{ 'onClick': {} },
    type: "primary",
    size: "large",
    loading: (__VLS_ctx.saving),
}));
const __VLS_178 = __VLS_177({
    ...{ 'onClick': {} },
    type: "primary",
    size: "large",
    loading: (__VLS_ctx.saving),
}, ...__VLS_functionalComponentArgsRest(__VLS_177));
let __VLS_180;
let __VLS_181;
let __VLS_182;
const __VLS_183 = {
    onClick: (__VLS_ctx.saveSettings)
};
__VLS_179.slots.default;
(__VLS_ctx.t('settings.saveBtn'));
var __VLS_179;
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "settings-card about-card" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.h3, __VLS_intrinsicElements.h3)({
    ...{ class: "settings-title" },
});
(__VLS_ctx.t('settings.sectionAbout'));
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
/** @type {__VLS_StyleScopedClasses['language-selector-wrap']} */ ;
/** @type {__VLS_StyleScopedClasses['custom-radio-group']} */ ;
/** @type {__VLS_StyleScopedClasses['settings-card']} */ ;
/** @type {__VLS_StyleScopedClasses['settings-title']} */ ;
/** @type {__VLS_StyleScopedClasses['theme-selector-wrap']} */ ;
/** @type {__VLS_StyleScopedClasses['custom-radio-group']} */ ;
/** @type {__VLS_StyleScopedClasses['theme-btn-icon']} */ ;
/** @type {__VLS_StyleScopedClasses['theme-btn-icon']} */ ;
/** @type {__VLS_StyleScopedClasses['actions-card']} */ ;
/** @type {__VLS_StyleScopedClasses['settings-card']} */ ;
/** @type {__VLS_StyleScopedClasses['about-card']} */ ;
/** @type {__VLS_StyleScopedClasses['settings-title']} */ ;
/** @type {__VLS_StyleScopedClasses['about-p']} */ ;
/** @type {__VLS_StyleScopedClasses['about-p']} */ ;
/** @type {__VLS_StyleScopedClasses['about-p']} */ ;
/** @type {__VLS_StyleScopedClasses['text-gray']} */ ;
var __VLS_dollars;
const __VLS_self = (await import('vue')).defineComponent({
    setup() {
        return {
            QuestionFilled: QuestionFilled,
            FolderOpened: FolderOpened,
            FolderAdd: FolderAdd,
            Moon: Moon,
            Sunny: Sunny,
            t: t,
            form: form,
            saving: saving,
            selectingDir: selectingDir,
            selectedLanguage: selectedLanguage,
            selectedTheme: selectedTheme,
            openDownloadFolder: openDownloadFolder,
            changeDownloadDirectory: changeDownloadDirectory,
            onSliderChange: onSliderChange,
            onLanguageChange: onLanguageChange,
            onThemeChange: onThemeChange,
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