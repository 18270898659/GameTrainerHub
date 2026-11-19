/// <reference types="../../node_modules/.vue-global-types/vue_3.5_0_0_0.d.ts" />
import { ref, onMounted } from 'vue';
import { Monitor, Plus, VideoPlay, Delete } from '@element-plus/icons-vue';
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
async function promptImport() {
    try {
        const { value: filePath } = await ElMessageBox.prompt(t('library.importExe'), t('library.importExe'), {
            confirmButtonText: t('library.importExe'),
            cancelButtonText: t('library.cancel'),
            inputPattern: /^.+\.exe$/i,
            inputErrorMessage: '.exe'
        });
        if (filePath) {
            await api.importCustomTrainer(1, filePath.trim());
            ElMessage.success(t('settings.savedSuccess'));
            await store.refreshLocalTrainers();
        }
    }
    catch {
        // 用户取消输入
    }
}
async function deleteTrainer(id) {
    try {
        await api.deleteLocalTrainer(id, false);
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
    onClick: (__VLS_ctx.promptImport)
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
        });
        __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
            ...{ class: "local-card-thumb" },
        });
        __VLS_asFunctionalElement(__VLS_intrinsicElements.img)({
            src: (item.game?.cover_url || __VLS_ctx.defaultThumb),
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
        const __VLS_36 = {}.ElButton;
        /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
        // @ts-ignore
        const __VLS_37 = __VLS_asFunctionalComponent(__VLS_36, new __VLS_36({
            ...{ 'onClick': {} },
            type: "primary",
            size: "small",
        }));
        const __VLS_38 = __VLS_37({
            ...{ 'onClick': {} },
            type: "primary",
            size: "small",
        }, ...__VLS_functionalComponentArgsRest(__VLS_37));
        let __VLS_40;
        let __VLS_41;
        let __VLS_42;
        const __VLS_43 = {
            onClick: (...[$event]) => {
                if (!(__VLS_ctx.store.localTrainers.length > 0))
                    return;
                __VLS_ctx.openLaunchModal(item.local_trainer.id);
            }
        };
        __VLS_39.slots.default;
        const __VLS_44 = {}.ElIcon;
        /** @type {[typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, ]} */ ;
        // @ts-ignore
        const __VLS_45 = __VLS_asFunctionalComponent(__VLS_44, new __VLS_44({}));
        const __VLS_46 = __VLS_45({}, ...__VLS_functionalComponentArgsRest(__VLS_45));
        __VLS_47.slots.default;
        const __VLS_48 = {}.VideoPlay;
        /** @type {[typeof __VLS_components.VideoPlay, ]} */ ;
        // @ts-ignore
        const __VLS_49 = __VLS_asFunctionalComponent(__VLS_48, new __VLS_48({}));
        const __VLS_50 = __VLS_49({}, ...__VLS_functionalComponentArgsRest(__VLS_49));
        var __VLS_47;
        (__VLS_ctx.t('library.launchBtn'));
        var __VLS_39;
        const __VLS_52 = {}.ElButton;
        /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
        // @ts-ignore
        const __VLS_53 = __VLS_asFunctionalComponent(__VLS_52, new __VLS_52({
            ...{ 'onClick': {} },
            size: "small",
            plain: true,
        }));
        const __VLS_54 = __VLS_53({
            ...{ 'onClick': {} },
            size: "small",
            plain: true,
        }, ...__VLS_functionalComponentArgsRest(__VLS_53));
        let __VLS_56;
        let __VLS_57;
        let __VLS_58;
        const __VLS_59 = {
            onClick: (...[$event]) => {
                if (!(__VLS_ctx.store.localTrainers.length > 0))
                    return;
                __VLS_ctx.openFolder(item.local_trainer.file_path);
            }
        };
        __VLS_55.slots.default;
        (__VLS_ctx.t('library.folderBtn'));
        var __VLS_55;
        const __VLS_60 = {}.ElPopconfirm;
        /** @type {[typeof __VLS_components.ElPopconfirm, typeof __VLS_components.elPopconfirm, typeof __VLS_components.ElPopconfirm, typeof __VLS_components.elPopconfirm, ]} */ ;
        // @ts-ignore
        const __VLS_61 = __VLS_asFunctionalComponent(__VLS_60, new __VLS_60({
            ...{ 'onConfirm': {} },
            title: (__VLS_ctx.t('library.deleteConfirm')),
            confirmButtonText: (__VLS_ctx.t('library.delete')),
            cancelButtonText: (__VLS_ctx.t('library.cancel')),
        }));
        const __VLS_62 = __VLS_61({
            ...{ 'onConfirm': {} },
            title: (__VLS_ctx.t('library.deleteConfirm')),
            confirmButtonText: (__VLS_ctx.t('library.delete')),
            cancelButtonText: (__VLS_ctx.t('library.cancel')),
        }, ...__VLS_functionalComponentArgsRest(__VLS_61));
        let __VLS_64;
        let __VLS_65;
        let __VLS_66;
        const __VLS_67 = {
            onConfirm: (...[$event]) => {
                if (!(__VLS_ctx.store.localTrainers.length > 0))
                    return;
                __VLS_ctx.deleteTrainer(item.local_trainer.id);
            }
        };
        __VLS_63.slots.default;
        {
            const { reference: __VLS_thisSlot } = __VLS_63.slots;
            const __VLS_68 = {}.ElButton;
            /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
            // @ts-ignore
            const __VLS_69 = __VLS_asFunctionalComponent(__VLS_68, new __VLS_68({
                size: "small",
                type: "danger",
                link: true,
            }));
            const __VLS_70 = __VLS_69({
                size: "small",
                type: "danger",
                link: true,
            }, ...__VLS_functionalComponentArgsRest(__VLS_69));
            __VLS_71.slots.default;
            const __VLS_72 = {}.ElIcon;
            /** @type {[typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, ]} */ ;
            // @ts-ignore
            const __VLS_73 = __VLS_asFunctionalComponent(__VLS_72, new __VLS_72({}));
            const __VLS_74 = __VLS_73({}, ...__VLS_functionalComponentArgsRest(__VLS_73));
            __VLS_75.slots.default;
            const __VLS_76 = {}.Delete;
            /** @type {[typeof __VLS_components.Delete, ]} */ ;
            // @ts-ignore
            const __VLS_77 = __VLS_asFunctionalComponent(__VLS_76, new __VLS_76({}));
            const __VLS_78 = __VLS_77({}, ...__VLS_functionalComponentArgsRest(__VLS_77));
            var __VLS_75;
            var __VLS_71;
        }
        var __VLS_63;
    }
}
else {
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "empty-wrap" },
    });
    const __VLS_80 = {}.ElEmpty;
    /** @type {[typeof __VLS_components.ElEmpty, typeof __VLS_components.elEmpty, typeof __VLS_components.ElEmpty, typeof __VLS_components.elEmpty, ]} */ ;
    // @ts-ignore
    const __VLS_81 = __VLS_asFunctionalComponent(__VLS_80, new __VLS_80({
        description: (__VLS_ctx.t('library.emptyLibrary')),
    }));
    const __VLS_82 = __VLS_81({
        description: (__VLS_ctx.t('library.emptyLibrary')),
    }, ...__VLS_functionalComponentArgsRest(__VLS_81));
    __VLS_83.slots.default;
    {
        const { extra: __VLS_thisSlot } = __VLS_83.slots;
        __VLS_asFunctionalElement(__VLS_intrinsicElements.p, __VLS_intrinsicElements.p)({
            ...{ class: "empty-tip" },
        });
        (__VLS_ctx.t('library.emptyLibraryTip'));
    }
    var __VLS_83;
}
/** @type {[typeof SecurityModal, ]} */ ;
// @ts-ignore
const __VLS_84 = __VLS_asFunctionalComponent(SecurityModal, new SecurityModal({
    ref: "securityModalRef",
}));
const __VLS_85 = __VLS_84({
    ref: "securityModalRef",
}, ...__VLS_functionalComponentArgsRest(__VLS_84));
/** @type {typeof __VLS_ctx.securityModalRef} */ ;
var __VLS_87 = {};
var __VLS_86;
/** @type {__VLS_StyleScopedClasses['library-view']} */ ;
/** @type {__VLS_StyleScopedClasses['lib-actions-bar']} */ ;
/** @type {__VLS_StyleScopedClasses['left-stats']} */ ;
/** @type {__VLS_StyleScopedClasses['lib-count']} */ ;
/** @type {__VLS_StyleScopedClasses['right-buttons']} */ ;
/** @type {__VLS_StyleScopedClasses['local-grid']} */ ;
/** @type {__VLS_StyleScopedClasses['local-card']} */ ;
/** @type {__VLS_StyleScopedClasses['local-card-thumb']} */ ;
/** @type {__VLS_StyleScopedClasses['source-badge']} */ ;
/** @type {__VLS_StyleScopedClasses['local-card-info']} */ ;
/** @type {__VLS_StyleScopedClasses['game-name']} */ ;
/** @type {__VLS_StyleScopedClasses['file-name']} */ ;
/** @type {__VLS_StyleScopedClasses['stats-row']} */ ;
/** @type {__VLS_StyleScopedClasses['card-btns']} */ ;
/** @type {__VLS_StyleScopedClasses['empty-wrap']} */ ;
/** @type {__VLS_StyleScopedClasses['empty-tip']} */ ;
// @ts-ignore
var __VLS_88 = __VLS_87;
var __VLS_dollars;
const __VLS_self = (await import('vue')).defineComponent({
    setup() {
        return {
            Monitor: Monitor,
            Plus: Plus,
            VideoPlay: VideoPlay,
            Delete: Delete,
            t: t,
            SecurityModal: SecurityModal,
            store: store,
            scanningSteam: scanningSteam,
            securityModalRef: securityModalRef,
            defaultThumb: defaultThumb,
            openLaunchModal: openLaunchModal,
            openFolder: openFolder,
            scanSteam: scanSteam,
            promptImport: promptImport,
            deleteTrainer: deleteTrainer,
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