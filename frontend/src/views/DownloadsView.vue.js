/// <reference types="../../node_modules/.vue-global-types/vue_3.5_0_0_0.d.ts" />
import { ref, onMounted } from 'vue';
import { VideoPause, VideoPlay, Delete } from '@element-plus/icons-vue';
import { ElMessage } from 'element-plus';
import { useAppStore } from '../stores/appStore';
import { api } from '../api/wails';
const store = useAppStore();
const activeTab = ref('active');
const historyList = ref([]);
onMounted(async () => {
    await store.refreshDownloads();
    loadHistory();
});
async function loadHistory() {
    try {
        historyList.value = await api.listDownloadHistory(50);
    }
    catch (e) {
        console.error('Failed to load history', e);
    }
}
async function pauseTask(id) {
    try {
        await api.pauseDownload(id);
        await store.refreshDownloads();
    }
    catch (e) {
        ElMessage.error('暂停失败: ' + (e.message || e));
    }
}
async function resumeTask(id) {
    try {
        await api.resumeDownload(id);
        await store.refreshDownloads();
    }
    catch (e) {
        ElMessage.error('恢复失败: ' + (e.message || e));
    }
}
async function cancelTask(id) {
    try {
        await api.cancelDownload(id);
        await store.refreshDownloads();
        ElMessage.success('已取消下载');
    }
    catch (e) {
        ElMessage.error('取消失败: ' + (e.message || e));
    }
}
async function openFolder(path) {
    try {
        await api.openContainingFolder(path);
    }
    catch (e) {
        ElMessage.error('定位失败: ' + (e.message || e));
    }
}
function copy(text) {
    navigator.clipboard.writeText(text);
    ElMessage.success('已复制');
}
function formatBytes(bytes) {
    if (!bytes || bytes <= 0)
        return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return (bytes / Math.pow(k, i)).toFixed(2) + ' ' + sizes[i];
}
function getStatusLabel(status) {
    switch (status) {
        case 'downloading': return '正在下载';
        case 'paused': return '已暂停';
        case 'completed': return '已完成';
        case 'failed': return '下载失败';
        default: return '等待中';
    }
}
debugger; /* PartiallyEnd: #3632/scriptSetup.vue */
const __VLS_ctx = {};
let __VLS_components;
let __VLS_directives;
// CSS variable injection 
// CSS variable injection end 
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "downloads-view" },
});
const __VLS_0 = {}.ElTabs;
/** @type {[typeof __VLS_components.ElTabs, typeof __VLS_components.elTabs, typeof __VLS_components.ElTabs, typeof __VLS_components.elTabs, ]} */ ;
// @ts-ignore
const __VLS_1 = __VLS_asFunctionalComponent(__VLS_0, new __VLS_0({
    modelValue: (__VLS_ctx.activeTab),
    ...{ class: "custom-tabs" },
}));
const __VLS_2 = __VLS_1({
    modelValue: (__VLS_ctx.activeTab),
    ...{ class: "custom-tabs" },
}, ...__VLS_functionalComponentArgsRest(__VLS_1));
__VLS_3.slots.default;
const __VLS_4 = {}.ElTabPane;
/** @type {[typeof __VLS_components.ElTabPane, typeof __VLS_components.elTabPane, typeof __VLS_components.ElTabPane, typeof __VLS_components.elTabPane, ]} */ ;
// @ts-ignore
const __VLS_5 = __VLS_asFunctionalComponent(__VLS_4, new __VLS_4({
    label: (`下载任务 (${__VLS_ctx.store.downloadTasks.length})`),
    name: "active",
}));
const __VLS_6 = __VLS_5({
    label: (`下载任务 (${__VLS_ctx.store.downloadTasks.length})`),
    name: "active",
}, ...__VLS_functionalComponentArgsRest(__VLS_5));
__VLS_7.slots.default;
if (__VLS_ctx.store.downloadTasks.length > 0) {
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "task-list" },
    });
    for (const [task] of __VLS_getVForSourceType((__VLS_ctx.store.downloadTasks))) {
        __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
            key: (task.id),
            ...{ class: "task-card" },
        });
        __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
            ...{ class: "task-header" },
        });
        __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
            ...{ class: "task-names" },
        });
        __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
            ...{ class: "task-game" },
        });
        (task.game_name || '修改器下载');
        __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
            ...{ class: "task-file" },
        });
        (task.file_name);
        __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
            ...{ class: "task-actions" },
        });
        if (task.status === 'downloading') {
            const __VLS_8 = {}.ElButton;
            /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
            // @ts-ignore
            const __VLS_9 = __VLS_asFunctionalComponent(__VLS_8, new __VLS_8({
                ...{ 'onClick': {} },
                size: "small",
                plain: true,
            }));
            const __VLS_10 = __VLS_9({
                ...{ 'onClick': {} },
                size: "small",
                plain: true,
            }, ...__VLS_functionalComponentArgsRest(__VLS_9));
            let __VLS_12;
            let __VLS_13;
            let __VLS_14;
            const __VLS_15 = {
                onClick: (...[$event]) => {
                    if (!(__VLS_ctx.store.downloadTasks.length > 0))
                        return;
                    if (!(task.status === 'downloading'))
                        return;
                    __VLS_ctx.pauseTask(task.id);
                }
            };
            __VLS_11.slots.default;
            const __VLS_16 = {}.ElIcon;
            /** @type {[typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, ]} */ ;
            // @ts-ignore
            const __VLS_17 = __VLS_asFunctionalComponent(__VLS_16, new __VLS_16({}));
            const __VLS_18 = __VLS_17({}, ...__VLS_functionalComponentArgsRest(__VLS_17));
            __VLS_19.slots.default;
            const __VLS_20 = {}.VideoPause;
            /** @type {[typeof __VLS_components.VideoPause, ]} */ ;
            // @ts-ignore
            const __VLS_21 = __VLS_asFunctionalComponent(__VLS_20, new __VLS_20({}));
            const __VLS_22 = __VLS_21({}, ...__VLS_functionalComponentArgsRest(__VLS_21));
            var __VLS_19;
            var __VLS_11;
        }
        else if (task.status === 'paused' || task.status === 'failed') {
            const __VLS_24 = {}.ElButton;
            /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
            // @ts-ignore
            const __VLS_25 = __VLS_asFunctionalComponent(__VLS_24, new __VLS_24({
                ...{ 'onClick': {} },
                size: "small",
                type: "primary",
                plain: true,
            }));
            const __VLS_26 = __VLS_25({
                ...{ 'onClick': {} },
                size: "small",
                type: "primary",
                plain: true,
            }, ...__VLS_functionalComponentArgsRest(__VLS_25));
            let __VLS_28;
            let __VLS_29;
            let __VLS_30;
            const __VLS_31 = {
                onClick: (...[$event]) => {
                    if (!(__VLS_ctx.store.downloadTasks.length > 0))
                        return;
                    if (!!(task.status === 'downloading'))
                        return;
                    if (!(task.status === 'paused' || task.status === 'failed'))
                        return;
                    __VLS_ctx.resumeTask(task.id);
                }
            };
            __VLS_27.slots.default;
            const __VLS_32 = {}.ElIcon;
            /** @type {[typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, ]} */ ;
            // @ts-ignore
            const __VLS_33 = __VLS_asFunctionalComponent(__VLS_32, new __VLS_32({}));
            const __VLS_34 = __VLS_33({}, ...__VLS_functionalComponentArgsRest(__VLS_33));
            __VLS_35.slots.default;
            const __VLS_36 = {}.VideoPlay;
            /** @type {[typeof __VLS_components.VideoPlay, ]} */ ;
            // @ts-ignore
            const __VLS_37 = __VLS_asFunctionalComponent(__VLS_36, new __VLS_36({}));
            const __VLS_38 = __VLS_37({}, ...__VLS_functionalComponentArgsRest(__VLS_37));
            var __VLS_35;
            var __VLS_27;
        }
        const __VLS_40 = {}.ElButton;
        /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
        // @ts-ignore
        const __VLS_41 = __VLS_asFunctionalComponent(__VLS_40, new __VLS_40({
            ...{ 'onClick': {} },
            size: "small",
            type: "danger",
            plain: true,
        }));
        const __VLS_42 = __VLS_41({
            ...{ 'onClick': {} },
            size: "small",
            type: "danger",
            plain: true,
        }, ...__VLS_functionalComponentArgsRest(__VLS_41));
        let __VLS_44;
        let __VLS_45;
        let __VLS_46;
        const __VLS_47 = {
            onClick: (...[$event]) => {
                if (!(__VLS_ctx.store.downloadTasks.length > 0))
                    return;
                __VLS_ctx.cancelTask(task.id);
            }
        };
        __VLS_43.slots.default;
        const __VLS_48 = {}.ElIcon;
        /** @type {[typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, ]} */ ;
        // @ts-ignore
        const __VLS_49 = __VLS_asFunctionalComponent(__VLS_48, new __VLS_48({}));
        const __VLS_50 = __VLS_49({}, ...__VLS_functionalComponentArgsRest(__VLS_49));
        __VLS_51.slots.default;
        const __VLS_52 = {}.Delete;
        /** @type {[typeof __VLS_components.Delete, ]} */ ;
        // @ts-ignore
        const __VLS_53 = __VLS_asFunctionalComponent(__VLS_52, new __VLS_52({}));
        const __VLS_54 = __VLS_53({}, ...__VLS_functionalComponentArgsRest(__VLS_53));
        var __VLS_51;
        var __VLS_43;
        __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
            ...{ class: "task-progress-wrap" },
        });
        const __VLS_56 = {}.ElProgress;
        /** @type {[typeof __VLS_components.ElProgress, typeof __VLS_components.elProgress, ]} */ ;
        // @ts-ignore
        const __VLS_57 = __VLS_asFunctionalComponent(__VLS_56, new __VLS_56({
            percentage: (Number(task.progress.toFixed(1))),
            status: (task.status === 'completed' ? 'success' : task.status === 'failed' ? 'exception' : ''),
            strokeWidth: (8),
        }));
        const __VLS_58 = __VLS_57({
            percentage: (Number(task.progress.toFixed(1))),
            status: (task.status === 'completed' ? 'success' : task.status === 'failed' ? 'exception' : ''),
            strokeWidth: (8),
        }, ...__VLS_functionalComponentArgsRest(__VLS_57));
        __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
            ...{ class: "task-footer" },
        });
        __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
            ...{ class: "task-bytes" },
        });
        (__VLS_ctx.formatBytes(task.downloaded_bytes));
        (__VLS_ctx.formatBytes(task.total_bytes));
        if (task.status === 'downloading') {
            __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
                ...{ class: "task-speed" },
            });
            (__VLS_ctx.formatBytes(task.speed_bytes_per_sec));
        }
        else {
            __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
                ...{ class: "task-status-text" },
            });
            (__VLS_ctx.getStatusLabel(task.status));
        }
    }
}
else {
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "empty-wrap" },
    });
    const __VLS_60 = {}.ElEmpty;
    /** @type {[typeof __VLS_components.ElEmpty, typeof __VLS_components.elEmpty, ]} */ ;
    // @ts-ignore
    const __VLS_61 = __VLS_asFunctionalComponent(__VLS_60, new __VLS_60({
        description: "当前暂无下载任务",
    }));
    const __VLS_62 = __VLS_61({
        description: "当前暂无下载任务",
    }, ...__VLS_functionalComponentArgsRest(__VLS_61));
}
var __VLS_7;
const __VLS_64 = {}.ElTabPane;
/** @type {[typeof __VLS_components.ElTabPane, typeof __VLS_components.elTabPane, typeof __VLS_components.ElTabPane, typeof __VLS_components.elTabPane, ]} */ ;
// @ts-ignore
const __VLS_65 = __VLS_asFunctionalComponent(__VLS_64, new __VLS_64({
    label: "下载历史",
    name: "history",
}));
const __VLS_66 = __VLS_65({
    label: "下载历史",
    name: "history",
}, ...__VLS_functionalComponentArgsRest(__VLS_65));
__VLS_67.slots.default;
const __VLS_68 = {}.ElTable;
/** @type {[typeof __VLS_components.ElTable, typeof __VLS_components.elTable, typeof __VLS_components.ElTable, typeof __VLS_components.elTable, ]} */ ;
// @ts-ignore
const __VLS_69 = __VLS_asFunctionalComponent(__VLS_68, new __VLS_68({
    data: (__VLS_ctx.historyList),
    stripe: true,
    ...{ style: {} },
    ...{ class: "custom-table" },
}));
const __VLS_70 = __VLS_69({
    data: (__VLS_ctx.historyList),
    stripe: true,
    ...{ style: {} },
    ...{ class: "custom-table" },
}, ...__VLS_functionalComponentArgsRest(__VLS_69));
__VLS_71.slots.default;
const __VLS_72 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_73 = __VLS_asFunctionalComponent(__VLS_72, new __VLS_72({
    prop: "game_name",
    label: "游戏名称",
    width: "220",
}));
const __VLS_74 = __VLS_73({
    prop: "game_name",
    label: "游戏名称",
    width: "220",
}, ...__VLS_functionalComponentArgsRest(__VLS_73));
const __VLS_76 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_77 = __VLS_asFunctionalComponent(__VLS_76, new __VLS_76({
    prop: "file_name",
    label: "文件名称",
}));
const __VLS_78 = __VLS_77({
    prop: "file_name",
    label: "文件名称",
}, ...__VLS_functionalComponentArgsRest(__VLS_77));
const __VLS_80 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_81 = __VLS_asFunctionalComponent(__VLS_80, new __VLS_80({
    label: "大小",
    width: "110",
}));
const __VLS_82 = __VLS_81({
    label: "大小",
    width: "110",
}, ...__VLS_functionalComponentArgsRest(__VLS_81));
__VLS_83.slots.default;
{
    const { default: __VLS_thisSlot } = __VLS_83.slots;
    const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
    (__VLS_ctx.formatBytes(row.file_size));
}
var __VLS_83;
const __VLS_84 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_85 = __VLS_asFunctionalComponent(__VLS_84, new __VLS_84({
    label: "SHA-256",
    width: "220",
}));
const __VLS_86 = __VLS_85({
    label: "SHA-256",
    width: "220",
}, ...__VLS_functionalComponentArgsRest(__VLS_85));
__VLS_87.slots.default;
{
    const { default: __VLS_thisSlot } = __VLS_87.slots;
    const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
    __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
        ...{ class: "hash-short" },
        title: (row.sha256),
    });
    (row.sha256.substring(0, 16));
    const __VLS_88 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_89 = __VLS_asFunctionalComponent(__VLS_88, new __VLS_88({
        ...{ 'onClick': {} },
        link: true,
        type: "primary",
        size: "small",
    }));
    const __VLS_90 = __VLS_89({
        ...{ 'onClick': {} },
        link: true,
        type: "primary",
        size: "small",
    }, ...__VLS_functionalComponentArgsRest(__VLS_89));
    let __VLS_92;
    let __VLS_93;
    let __VLS_94;
    const __VLS_95 = {
        onClick: (...[$event]) => {
            __VLS_ctx.copy(row.sha256);
        }
    };
    __VLS_91.slots.default;
    var __VLS_91;
}
var __VLS_87;
const __VLS_96 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_97 = __VLS_asFunctionalComponent(__VLS_96, new __VLS_96({
    label: "操作",
    width: "140",
}));
const __VLS_98 = __VLS_97({
    label: "操作",
    width: "140",
}, ...__VLS_functionalComponentArgsRest(__VLS_97));
__VLS_99.slots.default;
{
    const { default: __VLS_thisSlot } = __VLS_99.slots;
    const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
    const __VLS_100 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_101 = __VLS_asFunctionalComponent(__VLS_100, new __VLS_100({
        ...{ 'onClick': {} },
        size: "small",
        link: true,
        type: "primary",
    }));
    const __VLS_102 = __VLS_101({
        ...{ 'onClick': {} },
        size: "small",
        link: true,
        type: "primary",
    }, ...__VLS_functionalComponentArgsRest(__VLS_101));
    let __VLS_104;
    let __VLS_105;
    let __VLS_106;
    const __VLS_107 = {
        onClick: (...[$event]) => {
            __VLS_ctx.openFolder(row.save_path);
        }
    };
    __VLS_103.slots.default;
    var __VLS_103;
}
var __VLS_99;
var __VLS_71;
var __VLS_67;
var __VLS_3;
/** @type {__VLS_StyleScopedClasses['downloads-view']} */ ;
/** @type {__VLS_StyleScopedClasses['custom-tabs']} */ ;
/** @type {__VLS_StyleScopedClasses['task-list']} */ ;
/** @type {__VLS_StyleScopedClasses['task-card']} */ ;
/** @type {__VLS_StyleScopedClasses['task-header']} */ ;
/** @type {__VLS_StyleScopedClasses['task-names']} */ ;
/** @type {__VLS_StyleScopedClasses['task-game']} */ ;
/** @type {__VLS_StyleScopedClasses['task-file']} */ ;
/** @type {__VLS_StyleScopedClasses['task-actions']} */ ;
/** @type {__VLS_StyleScopedClasses['task-progress-wrap']} */ ;
/** @type {__VLS_StyleScopedClasses['task-footer']} */ ;
/** @type {__VLS_StyleScopedClasses['task-bytes']} */ ;
/** @type {__VLS_StyleScopedClasses['task-speed']} */ ;
/** @type {__VLS_StyleScopedClasses['task-status-text']} */ ;
/** @type {__VLS_StyleScopedClasses['empty-wrap']} */ ;
/** @type {__VLS_StyleScopedClasses['custom-table']} */ ;
/** @type {__VLS_StyleScopedClasses['hash-short']} */ ;
var __VLS_dollars;
const __VLS_self = (await import('vue')).defineComponent({
    setup() {
        return {
            VideoPause: VideoPause,
            VideoPlay: VideoPlay,
            Delete: Delete,
            store: store,
            activeTab: activeTab,
            historyList: historyList,
            pauseTask: pauseTask,
            resumeTask: resumeTask,
            cancelTask: cancelTask,
            openFolder: openFolder,
            copy: copy,
            formatBytes: formatBytes,
            getStatusLabel: getStatusLabel,
        };
    },
});
export default (await import('vue')).defineComponent({
    setup() {
        return {};
    },
});
; /* PartiallyEnd: #4569/main.vue */
//# sourceMappingURL=DownloadsView.vue.js.map