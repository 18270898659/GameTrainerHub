/// <reference types="../../node_modules/.vue-global-types/vue_3.5_0_0_0.d.ts" />
import { ref, onMounted } from 'vue';
import { Monitor, Plus, VideoPlay, VideoPause, Delete, Check, Connection } from '@element-plus/icons-vue';
import { ElMessage, ElMessageBox } from 'element-plus';
import { useAppStore } from '../stores/appStore';
import { api } from '../api/wails';
import { t } from '../i18n';
import SecurityModal from '../components/SecurityModal.vue';
defineOptions({ name: 'LibraryView' });
const store = useAppStore();
const scanningSteam = ref(false);
const securityModalRef = ref(null);
const defaultThumb = 'https://flingtrainer.com/wp-content/uploads/2020/12/header-7-200x200.jpg';
// 完整性校验状态
const integrityDialogVisible = ref(false);
const checkingIntegrity = ref(false);
const integrityResult = ref(null);
// 删除对话框状态
const deleteDialogVisible = ref(false);
const deletingTrainerID = ref(null);
const deletePhysicalFile = ref(false);
onMounted(() => {
    store.refreshLocalTrainers();
});
async function openLaunchModal(id) {
    try {
        const card = await api.getSecurityCard(id);
        securityModalRef.value?.open(card);
    }
    catch (e) {
        ElMessage.error(e?.message || String(e));
    }
}
async function confirmStop(id) {
    try {
        await ElMessageBox.confirm(t('library.stopConfirm'), t('library.stopTrainer'), {
            confirmButtonText: t('library.stopTrainer'),
            cancelButtonText: t('library.cancel'),
            type: 'warning'
        });
        await api.stopTrainer(id);
        ElMessage.success(t('library.stopSuccess'));
        await store.refreshLocalTrainers();
    }
    catch (e) {
        if (e !== 'cancel') {
            ElMessage.error(t('library.stopFailed') + (e?.message || e));
        }
    }
}
async function openFolder(path) {
    try {
        await api.openContainingFolder(path);
    }
    catch (e) {
        ElMessage.error(e?.message || String(e));
    }
}
async function scanSteam() {
    scanningSteam.value = true;
    try {
        const count = await api.scanSteamGames();
        ElMessage.success(t('library.scanSuccess', { count }));
        await store.refreshLocalTrainers();
    }
    catch (e) {
        ElMessage.error(t('library.scanFailed') + (e.message || e));
    }
    finally {
        scanningSteam.value = false;
    }
}
async function selectAndImportExe() {
    try {
        const selected = await api.selectTrainerExecutable();
        if (selected) {
            await api.importCustomTrainer(1, selected.trim());
            ElMessage.success(t('settings.savedSuccess'));
            await store.refreshLocalTrainers();
        }
    }
    catch (e) {
        ElMessage.error('导入修改器失败: ' + (e?.message || e));
    }
}
async function relocateTrainerFile(id) {
    try {
        const newPath = await api.selectTrainerExecutable();
        if (!newPath)
            return;
        await api.relocateTrainer(id, newPath);
        ElMessage.success(t('library.relocateSuccess'));
        await store.refreshLocalTrainers();
    }
    catch (e) {
        ElMessage.error(t('library.relocateFailed') + (e?.message || e));
    }
}
async function checkIntegrity(id) {
    integrityDialogVisible.value = true;
    checkingIntegrity.value = true;
    integrityResult.value = null;
    try {
        integrityResult.value = await api.verifyTrainerIntegrity(id);
    }
    catch (e) {
        ElMessage.error('校验失败: ' + (e?.message || e));
        integrityDialogVisible.value = false;
    }
    finally {
        checkingIntegrity.value = false;
    }
}
function promptDelete(id) {
    deletingTrainerID.value = id;
    deletePhysicalFile.value = false;
    deleteDialogVisible.value = true;
}
async function confirmDelete() {
    if (!deletingTrainerID.value)
        return;
    try {
        await api.deleteLocalTrainer(deletingTrainerID.value, deletePhysicalFile.value);
        ElMessage.success(t('settings.savedSuccess'));
        deleteDialogVisible.value = false;
        await store.refreshLocalTrainers();
    }
    catch (e) {
        ElMessage.error(e?.message || String(e));
    }
}
function getFileName(path) {
    if (!path)
        return '';
    const parts = path.split(/[/\\]/);
    return parts[parts.length - 1];
}
function formatBytes(bytes) {
    if (!bytes || bytes <= 0)
        return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return (bytes / Math.pow(k, i)).toFixed(1) + ' ' + sizes[i];
}
debugger; /* PartiallyEnd: #3632/scriptSetup.vue */
const __VLS_ctx = {};
let __VLS_components;
let __VLS_directives;
/** @type {__VLS_StyleScopedClasses['lib-count']} */ ;
/** @type {__VLS_StyleScopedClasses['local-card']} */ ;
/** @type {__VLS_StyleScopedClasses['local-card']} */ ;
/** @type {__VLS_StyleScopedClasses['local-card']} */ ;
/** @type {__VLS_StyleScopedClasses['local-card-thumb']} */ ;
// CSS variable injection 
// CSS variable injection end 
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "library-view" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "lib-actions-bar" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "left-stats" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
    ...{ class: "lib-count" },
});
(__VLS_ctx.t('library.managedCount'));
__VLS_asFunctionalElement(__VLS_intrinsicElements.strong, __VLS_intrinsicElements.strong)({});
(__VLS_ctx.store.localTrainers.length);
(__VLS_ctx.t('library.managedUnit'));
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "right-buttons" },
});
const __VLS_0 = {}.ElButton;
/** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
// @ts-ignore
const __VLS_1 = __VLS_asFunctionalComponent(__VLS_0, new __VLS_0({
    ...{ 'onClick': {} },
    loading: (__VLS_ctx.scanningSteam),
    plain: true,
}));
const __VLS_2 = __VLS_1({
    ...{ 'onClick': {} },
    loading: (__VLS_ctx.scanningSteam),
    plain: true,
}, ...__VLS_functionalComponentArgsRest(__VLS_1));
let __VLS_4;
let __VLS_5;
let __VLS_6;
const __VLS_7 = {
    onClick: (__VLS_ctx.scanSteam)
};
__VLS_3.slots.default;
const __VLS_8 = {}.ElIcon;
/** @type {[typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, ]} */ ;
// @ts-ignore
const __VLS_9 = __VLS_asFunctionalComponent(__VLS_8, new __VLS_8({}));
const __VLS_10 = __VLS_9({}, ...__VLS_functionalComponentArgsRest(__VLS_9));
__VLS_11.slots.default;
const __VLS_12 = {}.Monitor;
/** @type {[typeof __VLS_components.Monitor, ]} */ ;
// @ts-ignore
const __VLS_13 = __VLS_asFunctionalComponent(__VLS_12, new __VLS_12({}));
const __VLS_14 = __VLS_13({}, ...__VLS_functionalComponentArgsRest(__VLS_13));
var __VLS_11;
(__VLS_ctx.t('library.scanSteam'));
var __VLS_3;
const __VLS_16 = {}.ElButton;
/** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
// @ts-ignore
const __VLS_17 = __VLS_asFunctionalComponent(__VLS_16, new __VLS_16({
    ...{ 'onClick': {} },
    type: "primary",
    plain: true,
}));
const __VLS_18 = __VLS_17({
    ...{ 'onClick': {} },
    type: "primary",
    plain: true,
}, ...__VLS_functionalComponentArgsRest(__VLS_17));
let __VLS_20;
let __VLS_21;
let __VLS_22;
const __VLS_23 = {
    onClick: (__VLS_ctx.selectAndImportExe)
};
__VLS_19.slots.default;
const __VLS_24 = {}.ElIcon;
/** @type {[typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, ]} */ ;
// @ts-ignore
const __VLS_25 = __VLS_asFunctionalComponent(__VLS_24, new __VLS_24({}));
const __VLS_26 = __VLS_25({}, ...__VLS_functionalComponentArgsRest(__VLS_25));
__VLS_27.slots.default;
const __VLS_28 = {}.Plus;
/** @type {[typeof __VLS_components.Plus, ]} */ ;
// @ts-ignore
const __VLS_29 = __VLS_asFunctionalComponent(__VLS_28, new __VLS_28({}));
const __VLS_30 = __VLS_29({}, ...__VLS_functionalComponentArgsRest(__VLS_29));
var __VLS_27;
(__VLS_ctx.t('library.importExe'));
var __VLS_19;
if (__VLS_ctx.store.localTrainers.length > 0) {
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "local-grid" },
    });
    for (const [item] of __VLS_getVForSourceType((__VLS_ctx.store.localTrainers))) {
        __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
            key: (item.local_trainer.id),
            ...{ class: "local-card" },
            ...{ class: ({ 'card-missing': item.file_exists === false, 'card-running': item.is_running }) },
        });
        __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
            ...{ class: "local-card-thumb" },
        });
        __VLS_asFunctionalElement(__VLS_intrinsicElements.img)({
            src: (item.game?.cover_url || __VLS_ctx.defaultThumb),
        });
        __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
            ...{ class: "badge-group-top" },
        });
        const __VLS_32 = {}.ElTag;
        /** @type {[typeof __VLS_components.ElTag, typeof __VLS_components.elTag, typeof __VLS_components.ElTag, typeof __VLS_components.elTag, ]} */ ;
        // @ts-ignore
        const __VLS_33 = __VLS_asFunctionalComponent(__VLS_32, new __VLS_32({
            type: (item.local_trainer.is_custom ? 'warning' : 'success'),
            size: "small",
            ...{ class: "source-badge" },
        }));
        const __VLS_34 = __VLS_33({
            type: (item.local_trainer.is_custom ? 'warning' : 'success'),
            size: "small",
            ...{ class: "source-badge" },
        }, ...__VLS_functionalComponentArgsRest(__VLS_33));
        __VLS_35.slots.default;
        (item.local_trainer.source_name);
        var __VLS_35;
        if (item.is_running) {
            const __VLS_36 = {}.ElTag;
            /** @type {[typeof __VLS_components.ElTag, typeof __VLS_components.elTag, typeof __VLS_components.ElTag, typeof __VLS_components.elTag, ]} */ ;
            // @ts-ignore
            const __VLS_37 = __VLS_asFunctionalComponent(__VLS_36, new __VLS_36({
                type: "danger",
                effect: "dark",
                size: "small",
                ...{ class: "status-badge" },
            }));
            const __VLS_38 = __VLS_37({
                type: "danger",
                effect: "dark",
                size: "small",
                ...{ class: "status-badge" },
            }, ...__VLS_functionalComponentArgsRest(__VLS_37));
            __VLS_39.slots.default;
            (__VLS_ctx.t('library.statusRunning'));
            var __VLS_39;
        }
        else if (item.file_exists === false) {
            const __VLS_40 = {}.ElTag;
            /** @type {[typeof __VLS_components.ElTag, typeof __VLS_components.elTag, typeof __VLS_components.ElTag, typeof __VLS_components.elTag, ]} */ ;
            // @ts-ignore
            const __VLS_41 = __VLS_asFunctionalComponent(__VLS_40, new __VLS_40({
                type: "danger",
                size: "small",
                ...{ class: "status-badge" },
            }));
            const __VLS_42 = __VLS_41({
                type: "danger",
                size: "small",
                ...{ class: "status-badge" },
            }, ...__VLS_functionalComponentArgsRest(__VLS_41));
            __VLS_43.slots.default;
            (__VLS_ctx.t('library.fileMissing'));
            var __VLS_43;
        }
        __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
            ...{ class: "local-card-info" },
        });
        __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
            ...{ class: "game-name" },
        });
        (item.game?.name_cn || item.game?.name_en || __VLS_ctx.t('library.unassociatedGame'));
        __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
            ...{ class: "file-name" },
            title: (item.local_trainer.file_path),
        });
        (__VLS_ctx.getFileName(item.local_trainer.file_path));
        if (item.target_game_version) {
            __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
                ...{ class: "target-version-tag" },
            });
            const __VLS_44 = {}.ElTag;
            /** @type {[typeof __VLS_components.ElTag, typeof __VLS_components.elTag, typeof __VLS_components.ElTag, typeof __VLS_components.elTag, ]} */ ;
            // @ts-ignore
            const __VLS_45 = __VLS_asFunctionalComponent(__VLS_44, new __VLS_44({
                size: "small",
                type: "info",
                effect: "plain",
            }));
            const __VLS_46 = __VLS_45({
                size: "small",
                type: "info",
                effect: "plain",
            }, ...__VLS_functionalComponentArgsRest(__VLS_45));
            __VLS_47.slots.default;
            (__VLS_ctx.t('library.targetGameVersion', { version: item.target_game_version }));
            var __VLS_47;
        }
        __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
            ...{ class: "stats-row" },
        });
        __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({});
        (__VLS_ctx.t('library.runCount'));
        (item.local_trainer.run_count);
        __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({});
        (__VLS_ctx.t('library.fileSize'));
        (__VLS_ctx.formatBytes(item.local_trainer.file_size));
        __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
            ...{ class: "card-btns" },
        });
        if (item.is_running) {
            const __VLS_48 = {}.ElButton;
            /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
            // @ts-ignore
            const __VLS_49 = __VLS_asFunctionalComponent(__VLS_48, new __VLS_48({
                ...{ 'onClick': {} },
                type: "danger",
                size: "small",
            }));
            const __VLS_50 = __VLS_49({
                ...{ 'onClick': {} },
                type: "danger",
                size: "small",
            }, ...__VLS_functionalComponentArgsRest(__VLS_49));
            let __VLS_52;
            let __VLS_53;
            let __VLS_54;
            const __VLS_55 = {
                onClick: (...[$event]) => {
                    if (!(__VLS_ctx.store.localTrainers.length > 0))
                        return;
                    if (!(item.is_running))
                        return;
                    __VLS_ctx.confirmStop(item.local_trainer.id);
                }
            };
            __VLS_51.slots.default;
            const __VLS_56 = {}.ElIcon;
            /** @type {[typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, ]} */ ;
            // @ts-ignore
            const __VLS_57 = __VLS_asFunctionalComponent(__VLS_56, new __VLS_56({}));
            const __VLS_58 = __VLS_57({}, ...__VLS_functionalComponentArgsRest(__VLS_57));
            __VLS_59.slots.default;
            const __VLS_60 = {}.VideoPause;
            /** @type {[typeof __VLS_components.VideoPause, ]} */ ;
            // @ts-ignore
            const __VLS_61 = __VLS_asFunctionalComponent(__VLS_60, new __VLS_60({}));
            const __VLS_62 = __VLS_61({}, ...__VLS_functionalComponentArgsRest(__VLS_61));
            var __VLS_59;
            (__VLS_ctx.t('library.stopTrainer'));
            var __VLS_51;
        }
        else if (item.file_exists !== false) {
            const __VLS_64 = {}.ElButton;
            /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
            // @ts-ignore
            const __VLS_65 = __VLS_asFunctionalComponent(__VLS_64, new __VLS_64({
                ...{ 'onClick': {} },
                type: "primary",
                size: "small",
            }));
            const __VLS_66 = __VLS_65({
                ...{ 'onClick': {} },
                type: "primary",
                size: "small",
            }, ...__VLS_functionalComponentArgsRest(__VLS_65));
            let __VLS_68;
            let __VLS_69;
            let __VLS_70;
            const __VLS_71 = {
                onClick: (...[$event]) => {
                    if (!(__VLS_ctx.store.localTrainers.length > 0))
                        return;
                    if (!!(item.is_running))
                        return;
                    if (!(item.file_exists !== false))
                        return;
                    __VLS_ctx.openLaunchModal(item.local_trainer.id);
                }
            };
            __VLS_67.slots.default;
            const __VLS_72 = {}.ElIcon;
            /** @type {[typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, ]} */ ;
            // @ts-ignore
            const __VLS_73 = __VLS_asFunctionalComponent(__VLS_72, new __VLS_72({}));
            const __VLS_74 = __VLS_73({}, ...__VLS_functionalComponentArgsRest(__VLS_73));
            __VLS_75.slots.default;
            const __VLS_76 = {}.VideoPlay;
            /** @type {[typeof __VLS_components.VideoPlay, ]} */ ;
            // @ts-ignore
            const __VLS_77 = __VLS_asFunctionalComponent(__VLS_76, new __VLS_76({}));
            const __VLS_78 = __VLS_77({}, ...__VLS_functionalComponentArgsRest(__VLS_77));
            var __VLS_75;
            (__VLS_ctx.t('library.launchBtn'));
            var __VLS_67;
        }
        if (item.file_exists === false) {
            const __VLS_80 = {}.ElButton;
            /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
            // @ts-ignore
            const __VLS_81 = __VLS_asFunctionalComponent(__VLS_80, new __VLS_80({
                ...{ 'onClick': {} },
                type: "warning",
                size: "small",
            }));
            const __VLS_82 = __VLS_81({
                ...{ 'onClick': {} },
                type: "warning",
                size: "small",
            }, ...__VLS_functionalComponentArgsRest(__VLS_81));
            let __VLS_84;
            let __VLS_85;
            let __VLS_86;
            const __VLS_87 = {
                onClick: (...[$event]) => {
                    if (!(__VLS_ctx.store.localTrainers.length > 0))
                        return;
                    if (!(item.file_exists === false))
                        return;
                    __VLS_ctx.relocateTrainerFile(item.local_trainer.id);
                }
            };
            __VLS_83.slots.default;
            const __VLS_88 = {}.ElIcon;
            /** @type {[typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, ]} */ ;
            // @ts-ignore
            const __VLS_89 = __VLS_asFunctionalComponent(__VLS_88, new __VLS_88({}));
            const __VLS_90 = __VLS_89({}, ...__VLS_functionalComponentArgsRest(__VLS_89));
            __VLS_91.slots.default;
            const __VLS_92 = {}.Connection;
            /** @type {[typeof __VLS_components.Connection, ]} */ ;
            // @ts-ignore
            const __VLS_93 = __VLS_asFunctionalComponent(__VLS_92, new __VLS_92({}));
            const __VLS_94 = __VLS_93({}, ...__VLS_functionalComponentArgsRest(__VLS_93));
            var __VLS_91;
            (__VLS_ctx.t('library.relocate'));
            var __VLS_83;
        }
        if (item.file_exists !== false) {
            const __VLS_96 = {}.ElButton;
            /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
            // @ts-ignore
            const __VLS_97 = __VLS_asFunctionalComponent(__VLS_96, new __VLS_96({
                ...{ 'onClick': {} },
                size: "small",
                plain: true,
            }));
            const __VLS_98 = __VLS_97({
                ...{ 'onClick': {} },
                size: "small",
                plain: true,
            }, ...__VLS_functionalComponentArgsRest(__VLS_97));
            let __VLS_100;
            let __VLS_101;
            let __VLS_102;
            const __VLS_103 = {
                onClick: (...[$event]) => {
                    if (!(__VLS_ctx.store.localTrainers.length > 0))
                        return;
                    if (!(item.file_exists !== false))
                        return;
                    __VLS_ctx.openFolder(item.local_trainer.file_path);
                }
            };
            __VLS_99.slots.default;
            (__VLS_ctx.t('library.folderBtn'));
            var __VLS_99;
        }
        const __VLS_104 = {}.ElButton;
        /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
        // @ts-ignore
        const __VLS_105 = __VLS_asFunctionalComponent(__VLS_104, new __VLS_104({
            ...{ 'onClick': {} },
            size: "small",
            plain: true,
            title: (__VLS_ctx.t('library.verifyIntegrity')),
        }));
        const __VLS_106 = __VLS_105({
            ...{ 'onClick': {} },
            size: "small",
            plain: true,
            title: (__VLS_ctx.t('library.verifyIntegrity')),
        }, ...__VLS_functionalComponentArgsRest(__VLS_105));
        let __VLS_108;
        let __VLS_109;
        let __VLS_110;
        const __VLS_111 = {
            onClick: (...[$event]) => {
                if (!(__VLS_ctx.store.localTrainers.length > 0))
                    return;
                __VLS_ctx.checkIntegrity(item.local_trainer.id);
            }
        };
        __VLS_107.slots.default;
        const __VLS_112 = {}.ElIcon;
        /** @type {[typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, ]} */ ;
        // @ts-ignore
        const __VLS_113 = __VLS_asFunctionalComponent(__VLS_112, new __VLS_112({}));
        const __VLS_114 = __VLS_113({}, ...__VLS_functionalComponentArgsRest(__VLS_113));
        __VLS_115.slots.default;
        const __VLS_116 = {}.Check;
        /** @type {[typeof __VLS_components.Check, ]} */ ;
        // @ts-ignore
        const __VLS_117 = __VLS_asFunctionalComponent(__VLS_116, new __VLS_116({}));
        const __VLS_118 = __VLS_117({}, ...__VLS_functionalComponentArgsRest(__VLS_117));
        var __VLS_115;
        var __VLS_107;
        const __VLS_120 = {}.ElButton;
        /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
        // @ts-ignore
        const __VLS_121 = __VLS_asFunctionalComponent(__VLS_120, new __VLS_120({
            ...{ 'onClick': {} },
            size: "small",
            type: "danger",
            link: true,
        }));
        const __VLS_122 = __VLS_121({
            ...{ 'onClick': {} },
            size: "small",
            type: "danger",
            link: true,
        }, ...__VLS_functionalComponentArgsRest(__VLS_121));
        let __VLS_124;
        let __VLS_125;
        let __VLS_126;
        const __VLS_127 = {
            onClick: (...[$event]) => {
                if (!(__VLS_ctx.store.localTrainers.length > 0))
                    return;
                __VLS_ctx.promptDelete(item.local_trainer.id);
            }
        };
        __VLS_123.slots.default;
        const __VLS_128 = {}.ElIcon;
        /** @type {[typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, ]} */ ;
        // @ts-ignore
        const __VLS_129 = __VLS_asFunctionalComponent(__VLS_128, new __VLS_128({}));
        const __VLS_130 = __VLS_129({}, ...__VLS_functionalComponentArgsRest(__VLS_129));
        __VLS_131.slots.default;
        const __VLS_132 = {}.Delete;
        /** @type {[typeof __VLS_components.Delete, ]} */ ;
        // @ts-ignore
        const __VLS_133 = __VLS_asFunctionalComponent(__VLS_132, new __VLS_132({}));
        const __VLS_134 = __VLS_133({}, ...__VLS_functionalComponentArgsRest(__VLS_133));
        var __VLS_131;
        var __VLS_123;
    }
}
else {
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "empty-wrap" },
    });
    const __VLS_136 = {}.ElEmpty;
    /** @type {[typeof __VLS_components.ElEmpty, typeof __VLS_components.elEmpty, typeof __VLS_components.ElEmpty, typeof __VLS_components.elEmpty, ]} */ ;
    // @ts-ignore
    const __VLS_137 = __VLS_asFunctionalComponent(__VLS_136, new __VLS_136({
        description: (__VLS_ctx.t('library.emptyLibrary')),
    }));
    const __VLS_138 = __VLS_137({
        description: (__VLS_ctx.t('library.emptyLibrary')),
    }, ...__VLS_functionalComponentArgsRest(__VLS_137));
    __VLS_139.slots.default;
    {
        const { extra: __VLS_thisSlot } = __VLS_139.slots;
        __VLS_asFunctionalElement(__VLS_intrinsicElements.p, __VLS_intrinsicElements.p)({
            ...{ class: "empty-tip" },
        });
        (__VLS_ctx.t('library.emptyLibraryTip'));
    }
    var __VLS_139;
}
const __VLS_140 = {}.ElDialog;
/** @type {[typeof __VLS_components.ElDialog, typeof __VLS_components.elDialog, typeof __VLS_components.ElDialog, typeof __VLS_components.elDialog, ]} */ ;
// @ts-ignore
const __VLS_141 = __VLS_asFunctionalComponent(__VLS_140, new __VLS_140({
    modelValue: (__VLS_ctx.integrityDialogVisible),
    title: (__VLS_ctx.t('library.integrityTitle')),
    width: "560px",
    appendToBody: true,
}));
const __VLS_142 = __VLS_141({
    modelValue: (__VLS_ctx.integrityDialogVisible),
    title: (__VLS_ctx.t('library.integrityTitle')),
    width: "560px",
    appendToBody: true,
}, ...__VLS_functionalComponentArgsRest(__VLS_141));
__VLS_143.slots.default;
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "integrity-dialog-body" },
});
__VLS_asFunctionalDirective(__VLS_directives.vLoading)(null, { ...__VLS_directiveBindingRestFields, value: (__VLS_ctx.checkingIntegrity) }, null, null);
if (__VLS_ctx.integrityResult) {
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "integrity-content" },
    });
    const __VLS_144 = {}.ElAlert;
    /** @type {[typeof __VLS_components.ElAlert, typeof __VLS_components.elAlert, ]} */ ;
    // @ts-ignore
    const __VLS_145 = __VLS_asFunctionalComponent(__VLS_144, new __VLS_144({
        type: (__VLS_ctx.integrityResult.is_valid ? 'success' : 'error'),
        title: (__VLS_ctx.integrityResult.is_valid ? __VLS_ctx.t('library.integrityMatch') : __VLS_ctx.t('library.integrityMismatch')),
        closable: (false),
        showIcon: true,
        ...{ class: "mb-4" },
    }));
    const __VLS_146 = __VLS_145({
        type: (__VLS_ctx.integrityResult.is_valid ? 'success' : 'error'),
        title: (__VLS_ctx.integrityResult.is_valid ? __VLS_ctx.t('library.integrityMatch') : __VLS_ctx.t('library.integrityMismatch')),
        closable: (false),
        showIcon: true,
        ...{ class: "mb-4" },
    }, ...__VLS_functionalComponentArgsRest(__VLS_145));
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "hash-field" },
    });
    __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
        ...{ class: "hash-label" },
    });
    (__VLS_ctx.t('library.expectedHash'));
    __VLS_asFunctionalElement(__VLS_intrinsicElements.code, __VLS_intrinsicElements.code)({
        ...{ class: "hash-code" },
    });
    (__VLS_ctx.integrityResult.expected_sha256);
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "hash-field mt-2" },
    });
    __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
        ...{ class: "hash-label" },
    });
    (__VLS_ctx.t('library.actualHash'));
    __VLS_asFunctionalElement(__VLS_intrinsicElements.code, __VLS_intrinsicElements.code)({
        ...{ class: "hash-code" },
    });
    (__VLS_ctx.integrityResult.actual_sha256);
}
{
    const { footer: __VLS_thisSlot } = __VLS_143.slots;
    const __VLS_148 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_149 = __VLS_asFunctionalComponent(__VLS_148, new __VLS_148({
        ...{ 'onClick': {} },
    }));
    const __VLS_150 = __VLS_149({
        ...{ 'onClick': {} },
    }, ...__VLS_functionalComponentArgsRest(__VLS_149));
    let __VLS_152;
    let __VLS_153;
    let __VLS_154;
    const __VLS_155 = {
        onClick: (...[$event]) => {
            __VLS_ctx.integrityDialogVisible = false;
        }
    };
    __VLS_151.slots.default;
    (__VLS_ctx.t('library.cancel'));
    var __VLS_151;
}
var __VLS_143;
const __VLS_156 = {}.ElDialog;
/** @type {[typeof __VLS_components.ElDialog, typeof __VLS_components.elDialog, typeof __VLS_components.ElDialog, typeof __VLS_components.elDialog, ]} */ ;
// @ts-ignore
const __VLS_157 = __VLS_asFunctionalComponent(__VLS_156, new __VLS_156({
    modelValue: (__VLS_ctx.deleteDialogVisible),
    title: (__VLS_ctx.t('library.deleteDialogTitle')),
    width: "440px",
    appendToBody: true,
}));
const __VLS_158 = __VLS_157({
    modelValue: (__VLS_ctx.deleteDialogVisible),
    title: (__VLS_ctx.t('library.deleteDialogTitle')),
    width: "440px",
    appendToBody: true,
}, ...__VLS_functionalComponentArgsRest(__VLS_157));
__VLS_159.slots.default;
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "delete-dialog-content" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.p, __VLS_intrinsicElements.p)({
    ...{ class: "delete-warning-text" },
});
(__VLS_ctx.t('library.deleteConfirm'));
const __VLS_160 = {}.ElCheckbox;
/** @type {[typeof __VLS_components.ElCheckbox, typeof __VLS_components.elCheckbox, typeof __VLS_components.ElCheckbox, typeof __VLS_components.elCheckbox, ]} */ ;
// @ts-ignore
const __VLS_161 = __VLS_asFunctionalComponent(__VLS_160, new __VLS_160({
    modelValue: (__VLS_ctx.deletePhysicalFile),
}));
const __VLS_162 = __VLS_161({
    modelValue: (__VLS_ctx.deletePhysicalFile),
}, ...__VLS_functionalComponentArgsRest(__VLS_161));
__VLS_163.slots.default;
(__VLS_ctx.t('library.deleteFileCheck'));
var __VLS_163;
{
    const { footer: __VLS_thisSlot } = __VLS_159.slots;
    const __VLS_164 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_165 = __VLS_asFunctionalComponent(__VLS_164, new __VLS_164({
        ...{ 'onClick': {} },
    }));
    const __VLS_166 = __VLS_165({
        ...{ 'onClick': {} },
    }, ...__VLS_functionalComponentArgsRest(__VLS_165));
    let __VLS_168;
    let __VLS_169;
    let __VLS_170;
    const __VLS_171 = {
        onClick: (...[$event]) => {
            __VLS_ctx.deleteDialogVisible = false;
        }
    };
    __VLS_167.slots.default;
    (__VLS_ctx.t('library.cancel'));
    var __VLS_167;
    const __VLS_172 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_173 = __VLS_asFunctionalComponent(__VLS_172, new __VLS_172({
        ...{ 'onClick': {} },
        type: "danger",
    }));
    const __VLS_174 = __VLS_173({
        ...{ 'onClick': {} },
        type: "danger",
    }, ...__VLS_functionalComponentArgsRest(__VLS_173));
    let __VLS_176;
    let __VLS_177;
    let __VLS_178;
    const __VLS_179 = {
        onClick: (__VLS_ctx.confirmDelete)
    };
    __VLS_175.slots.default;
    (__VLS_ctx.t('library.delete'));
    var __VLS_175;
}
var __VLS_159;
/** @type {[typeof SecurityModal, ]} */ ;
// @ts-ignore
const __VLS_180 = __VLS_asFunctionalComponent(SecurityModal, new SecurityModal({
    ref: "securityModalRef",
}));
const __VLS_181 = __VLS_180({
    ref: "securityModalRef",
}, ...__VLS_functionalComponentArgsRest(__VLS_180));
/** @type {typeof __VLS_ctx.securityModalRef} */ ;
var __VLS_183 = {};
var __VLS_182;
/** @type {__VLS_StyleScopedClasses['library-view']} */ ;
/** @type {__VLS_StyleScopedClasses['lib-actions-bar']} */ ;
/** @type {__VLS_StyleScopedClasses['left-stats']} */ ;
/** @type {__VLS_StyleScopedClasses['lib-count']} */ ;
/** @type {__VLS_StyleScopedClasses['right-buttons']} */ ;
/** @type {__VLS_StyleScopedClasses['local-grid']} */ ;
/** @type {__VLS_StyleScopedClasses['local-card']} */ ;
/** @type {__VLS_StyleScopedClasses['card-missing']} */ ;
/** @type {__VLS_StyleScopedClasses['card-running']} */ ;
/** @type {__VLS_StyleScopedClasses['local-card-thumb']} */ ;
/** @type {__VLS_StyleScopedClasses['badge-group-top']} */ ;
/** @type {__VLS_StyleScopedClasses['source-badge']} */ ;
/** @type {__VLS_StyleScopedClasses['status-badge']} */ ;
/** @type {__VLS_StyleScopedClasses['status-badge']} */ ;
/** @type {__VLS_StyleScopedClasses['local-card-info']} */ ;
/** @type {__VLS_StyleScopedClasses['game-name']} */ ;
/** @type {__VLS_StyleScopedClasses['file-name']} */ ;
/** @type {__VLS_StyleScopedClasses['target-version-tag']} */ ;
/** @type {__VLS_StyleScopedClasses['stats-row']} */ ;
/** @type {__VLS_StyleScopedClasses['card-btns']} */ ;
/** @type {__VLS_StyleScopedClasses['empty-wrap']} */ ;
/** @type {__VLS_StyleScopedClasses['empty-tip']} */ ;
/** @type {__VLS_StyleScopedClasses['integrity-dialog-body']} */ ;
/** @type {__VLS_StyleScopedClasses['integrity-content']} */ ;
/** @type {__VLS_StyleScopedClasses['mb-4']} */ ;
/** @type {__VLS_StyleScopedClasses['hash-field']} */ ;
/** @type {__VLS_StyleScopedClasses['hash-label']} */ ;
/** @type {__VLS_StyleScopedClasses['hash-code']} */ ;
/** @type {__VLS_StyleScopedClasses['hash-field']} */ ;
/** @type {__VLS_StyleScopedClasses['mt-2']} */ ;
/** @type {__VLS_StyleScopedClasses['hash-label']} */ ;
/** @type {__VLS_StyleScopedClasses['hash-code']} */ ;
/** @type {__VLS_StyleScopedClasses['delete-dialog-content']} */ ;
/** @type {__VLS_StyleScopedClasses['delete-warning-text']} */ ;
// @ts-ignore
var __VLS_184 = __VLS_183;
var __VLS_dollars;
const __VLS_self = (await import('vue')).defineComponent({
    setup() {
        return {
            Monitor: Monitor,
            Plus: Plus,
            VideoPlay: VideoPlay,
            VideoPause: VideoPause,
            Delete: Delete,
            Check: Check,
            Connection: Connection,
            t: t,
            SecurityModal: SecurityModal,
            store: store,
            scanningSteam: scanningSteam,
            securityModalRef: securityModalRef,
            defaultThumb: defaultThumb,
            integrityDialogVisible: integrityDialogVisible,
            checkingIntegrity: checkingIntegrity,
            integrityResult: integrityResult,
            deleteDialogVisible: deleteDialogVisible,
            deletePhysicalFile: deletePhysicalFile,
            openLaunchModal: openLaunchModal,
            confirmStop: confirmStop,
            openFolder: openFolder,
            scanSteam: scanSteam,
            selectAndImportExe: selectAndImportExe,
            relocateTrainerFile: relocateTrainerFile,
            checkIntegrity: checkIntegrity,
            promptDelete: promptDelete,
            confirmDelete: confirmDelete,
            getFileName: getFileName,
            formatBytes: formatBytes,
        };
    },
});
export default (await import('vue')).defineComponent({
    setup() {
        return {};
    },
});
; /* PartiallyEnd: #4569/main.vue */
//# sourceMappingURL=LibraryView.vue.js.map