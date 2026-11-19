/// <reference types="../../node_modules/.vue-global-types/vue_3.5_0_0_0.d.ts" />
import { ref } from 'vue';
import { ElMessage } from 'element-plus';
import { api } from '../api/wails';
import { t } from '../i18n';
const visible = ref(false);
const card = ref(null);
const runAsAdmin = ref(false);
const launching = ref(false);
function open(info) {
    card.value = info;
    runAsAdmin.value = info.requires_admin || false;
    visible.value = true;
}
function formatBytes(bytes) {
    if (bytes <= 0)
        return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return (bytes / Math.pow(k, i)).toFixed(2) + ' ' + sizes[i];
}
function copyHash() {
    if (card.value?.sha256) {
        navigator.clipboard.writeText(card.value.sha256);
        ElMessage.success(t('security.copied'));
    }
}
async function openFolder() {
    if (card.value?.file_path) {
        try {
            await api.openContainingFolder(card.value.file_path);
        }
        catch (e) {
            ElMessage.error(e?.message || String(e));
        }
    }
}
async function confirmLaunch() {
    if (!card.value)
        return;
    launching.value = true;
    try {
        await api.launchTrainer(card.value.trainer_id, runAsAdmin.value);
        ElMessage.success(t('security.launchSuccess'));
        visible.value = false;
    }
    catch (e) {
        const msg = e?.message || e || '';
        if (msg.includes('取消') || msg.includes('cancel')) {
            ElMessage.warning(msg);
        }
        else {
            ElMessage.error(t('security.launchFailed') + msg);
        }
    }
    finally {
        launching.value = false;
    }
}
const __VLS_exposed = { open };
defineExpose(__VLS_exposed);
debugger; /* PartiallyEnd: #3632/scriptSetup.vue */
const __VLS_ctx = {};
let __VLS_components;
let __VLS_directives;
/** @type {__VLS_StyleScopedClasses['info-row']} */ ;
/** @type {__VLS_StyleScopedClasses['hash-text']} */ ;
// CSS variable injection 
// CSS variable injection end 
const __VLS_0 = {}.ElDialog;
/** @type {[typeof __VLS_components.ElDialog, typeof __VLS_components.elDialog, typeof __VLS_components.ElDialog, typeof __VLS_components.elDialog, ]} */ ;
// @ts-ignore
const __VLS_1 = __VLS_asFunctionalComponent(__VLS_0, new __VLS_0({
    modelValue: (__VLS_ctx.visible),
    title: (__VLS_ctx.t('security.modalTitle')),
    width: "580px",
    ...{ class: "security-modal" },
    closeOnClickModal: (false),
}));
const __VLS_2 = __VLS_1({
    modelValue: (__VLS_ctx.visible),
    title: (__VLS_ctx.t('security.modalTitle')),
    width: "580px",
    ...{ class: "security-modal" },
    closeOnClickModal: (false),
}, ...__VLS_functionalComponentArgsRest(__VLS_1));
var __VLS_4 = {};
__VLS_3.slots.default;
if (__VLS_ctx.card) {
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "security-body" },
    });
    const __VLS_5 = {}.ElAlert;
    /** @type {[typeof __VLS_components.ElAlert, typeof __VLS_components.elAlert, typeof __VLS_components.ElAlert, typeof __VLS_components.elAlert, ]} */ ;
    // @ts-ignore
    const __VLS_6 = __VLS_asFunctionalComponent(__VLS_5, new __VLS_5({
        title: (__VLS_ctx.t('security.alertTitle')),
        type: "warning",
        closable: (false),
        showIcon: true,
        ...{ class: "mb-3" },
    }));
    const __VLS_7 = __VLS_6({
        title: (__VLS_ctx.t('security.alertTitle')),
        type: "warning",
        closable: (false),
        showIcon: true,
        ...{ class: "mb-3" },
    }, ...__VLS_functionalComponentArgsRest(__VLS_6));
    __VLS_8.slots.default;
    {
        const { default: __VLS_thisSlot } = __VLS_8.slots;
        __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
            ...{ class: "alert-desc" },
        });
        (__VLS_ctx.t('security.alertDesc'));
    }
    var __VLS_8;
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "info-table" },
    });
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "info-row" },
    });
    __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
        ...{ class: "label" },
    });
    (__VLS_ctx.t('security.targetGame'));
    __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
        ...{ class: "val font-bold" },
    });
    (__VLS_ctx.card.game_name);
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "info-row" },
    });
    __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
        ...{ class: "label" },
    });
    (__VLS_ctx.t('security.fileName'));
    __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
        ...{ class: "val" },
    });
    (__VLS_ctx.card.file_name);
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "info-row" },
    });
    __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
        ...{ class: "label" },
    });
    (__VLS_ctx.t('security.fileSize'));
    __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
        ...{ class: "val" },
    });
    (__VLS_ctx.formatBytes(__VLS_ctx.card.file_size));
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "info-row" },
    });
    __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
        ...{ class: "label" },
    });
    (__VLS_ctx.t('security.resourceSource'));
    const __VLS_9 = {}.ElTag;
    /** @type {[typeof __VLS_components.ElTag, typeof __VLS_components.elTag, typeof __VLS_components.ElTag, typeof __VLS_components.elTag, ]} */ ;
    // @ts-ignore
    const __VLS_10 = __VLS_asFunctionalComponent(__VLS_9, new __VLS_9({
        size: "small",
        type: "success",
    }));
    const __VLS_11 = __VLS_10({
        size: "small",
        type: "success",
    }, ...__VLS_functionalComponentArgsRest(__VLS_10));
    __VLS_12.slots.default;
    (__VLS_ctx.card.source_name);
    var __VLS_12;
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "info-row hash-row" },
    });
    __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
        ...{ class: "label" },
    });
    (__VLS_ctx.t('security.sha256'));
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "val hash-text" },
    });
    __VLS_asFunctionalElement(__VLS_intrinsicElements.code, __VLS_intrinsicElements.code)({});
    (__VLS_ctx.card.sha256);
    const __VLS_13 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_14 = __VLS_asFunctionalComponent(__VLS_13, new __VLS_13({
        ...{ 'onClick': {} },
        link: true,
        type: "primary",
        size: "small",
    }));
    const __VLS_15 = __VLS_14({
        ...{ 'onClick': {} },
        link: true,
        type: "primary",
        size: "small",
    }, ...__VLS_functionalComponentArgsRest(__VLS_14));
    let __VLS_17;
    let __VLS_18;
    let __VLS_19;
    const __VLS_20 = {
        onClick: (__VLS_ctx.copyHash)
    };
    __VLS_16.slots.default;
    (__VLS_ctx.t('security.copy'));
    var __VLS_16;
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "info-row" },
    });
    __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
        ...{ class: "label" },
    });
    (__VLS_ctx.t('security.localPath'));
    __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
        ...{ class: "val path-text" },
        title: (__VLS_ctx.card.file_path),
    });
    (__VLS_ctx.card.file_path);
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "admin-checkbox mt-3" },
    });
    const __VLS_21 = {}.ElCheckbox;
    /** @type {[typeof __VLS_components.ElCheckbox, typeof __VLS_components.elCheckbox, typeof __VLS_components.ElCheckbox, typeof __VLS_components.elCheckbox, ]} */ ;
    // @ts-ignore
    const __VLS_22 = __VLS_asFunctionalComponent(__VLS_21, new __VLS_21({
        modelValue: (__VLS_ctx.runAsAdmin),
    }));
    const __VLS_23 = __VLS_22({
        modelValue: (__VLS_ctx.runAsAdmin),
    }, ...__VLS_functionalComponentArgsRest(__VLS_22));
    __VLS_24.slots.default;
    (__VLS_ctx.t('security.runAsAdmin'));
    if (__VLS_ctx.card.requires_admin) {
        __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
            ...{ class: "admin-hint-alert" },
        });
        (__VLS_ctx.t('security.adminAlert'));
    }
    else {
        __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
            ...{ class: "admin-hint" },
        });
        (__VLS_ctx.t('security.adminHint'));
    }
    var __VLS_24;
}
{
    const { footer: __VLS_thisSlot } = __VLS_3.slots;
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "dialog-footer" },
    });
    const __VLS_25 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_26 = __VLS_asFunctionalComponent(__VLS_25, new __VLS_25({
        ...{ 'onClick': {} },
        plain: true,
    }));
    const __VLS_27 = __VLS_26({
        ...{ 'onClick': {} },
        plain: true,
    }, ...__VLS_functionalComponentArgsRest(__VLS_26));
    let __VLS_29;
    let __VLS_30;
    let __VLS_31;
    const __VLS_32 = {
        onClick: (__VLS_ctx.openFolder)
    };
    __VLS_28.slots.default;
    (__VLS_ctx.t('security.openFolder'));
    var __VLS_28;
    const __VLS_33 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_34 = __VLS_asFunctionalComponent(__VLS_33, new __VLS_33({
        ...{ 'onClick': {} },
    }));
    const __VLS_35 = __VLS_34({
        ...{ 'onClick': {} },
    }, ...__VLS_functionalComponentArgsRest(__VLS_34));
    let __VLS_37;
    let __VLS_38;
    let __VLS_39;
    const __VLS_40 = {
        onClick: (...[$event]) => {
            __VLS_ctx.visible = false;
        }
    };
    __VLS_36.slots.default;
    (__VLS_ctx.t('library.cancel'));
    var __VLS_36;
    const __VLS_41 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_42 = __VLS_asFunctionalComponent(__VLS_41, new __VLS_41({
        ...{ 'onClick': {} },
        type: "primary",
        loading: (__VLS_ctx.launching),
    }));
    const __VLS_43 = __VLS_42({
        ...{ 'onClick': {} },
        type: "primary",
        loading: (__VLS_ctx.launching),
    }, ...__VLS_functionalComponentArgsRest(__VLS_42));
    let __VLS_45;
    let __VLS_46;
    let __VLS_47;
    const __VLS_48 = {
        onClick: (__VLS_ctx.confirmLaunch)
    };
    __VLS_44.slots.default;
    (__VLS_ctx.t('security.confirmLaunch'));
    var __VLS_44;
}
var __VLS_3;
/** @type {__VLS_StyleScopedClasses['security-modal']} */ ;
/** @type {__VLS_StyleScopedClasses['security-body']} */ ;
/** @type {__VLS_StyleScopedClasses['mb-3']} */ ;
/** @type {__VLS_StyleScopedClasses['alert-desc']} */ ;
/** @type {__VLS_StyleScopedClasses['info-table']} */ ;
/** @type {__VLS_StyleScopedClasses['info-row']} */ ;
/** @type {__VLS_StyleScopedClasses['label']} */ ;
/** @type {__VLS_StyleScopedClasses['val']} */ ;
/** @type {__VLS_StyleScopedClasses['font-bold']} */ ;
/** @type {__VLS_StyleScopedClasses['info-row']} */ ;
/** @type {__VLS_StyleScopedClasses['label']} */ ;
/** @type {__VLS_StyleScopedClasses['val']} */ ;
/** @type {__VLS_StyleScopedClasses['info-row']} */ ;
/** @type {__VLS_StyleScopedClasses['label']} */ ;
/** @type {__VLS_StyleScopedClasses['val']} */ ;
/** @type {__VLS_StyleScopedClasses['info-row']} */ ;
/** @type {__VLS_StyleScopedClasses['label']} */ ;
/** @type {__VLS_StyleScopedClasses['info-row']} */ ;
/** @type {__VLS_StyleScopedClasses['hash-row']} */ ;
/** @type {__VLS_StyleScopedClasses['label']} */ ;
/** @type {__VLS_StyleScopedClasses['val']} */ ;
/** @type {__VLS_StyleScopedClasses['hash-text']} */ ;
/** @type {__VLS_StyleScopedClasses['info-row']} */ ;
/** @type {__VLS_StyleScopedClasses['label']} */ ;
/** @type {__VLS_StyleScopedClasses['val']} */ ;
/** @type {__VLS_StyleScopedClasses['path-text']} */ ;
/** @type {__VLS_StyleScopedClasses['admin-checkbox']} */ ;
/** @type {__VLS_StyleScopedClasses['mt-3']} */ ;
/** @type {__VLS_StyleScopedClasses['admin-hint-alert']} */ ;
/** @type {__VLS_StyleScopedClasses['admin-hint']} */ ;
/** @type {__VLS_StyleScopedClasses['dialog-footer']} */ ;
var __VLS_dollars;
const __VLS_self = (await import('vue')).defineComponent({
    setup() {
        return {
            t: t,
            visible: visible,
            card: card,
            runAsAdmin: runAsAdmin,
            launching: launching,
            formatBytes: formatBytes,
            copyHash: copyHash,
            openFolder: openFolder,
            confirmLaunch: confirmLaunch,
        };
    },
});
export default (await import('vue')).defineComponent({
    setup() {
        return {
            ...__VLS_exposed,
        };
    },
});
; /* PartiallyEnd: #4569/main.vue */
//# sourceMappingURL=SecurityModal.vue.js.map