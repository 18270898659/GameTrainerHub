/// <reference types="../../node_modules/.vue-global-types/vue_3.5_0_0_0.d.ts" />
import { ref, onMounted } from 'vue';
import { useRoute } from 'vue-router';
import { Search, Download } from '@element-plus/icons-vue';
import { ElMessage } from 'element-plus';
import { api } from '../api/wails';
const route = useRoute();
const query = ref('');
const localOnly = ref(false);
const selectedSource = ref('all');
const searching = ref(false);
const hasSearched = ref(false);
const results = ref([]);
const drawerVisible = ref(false);
const selectedGame = ref(null);
const loadingDetail = ref(false);
const detailVersions = ref([]);
const downloadingVerID = ref(null);
const defaultThumb = 'https://flingtrainer.com/wp-content/uploads/2020/12/header-7-200x200.jpg';
onMounted(() => {
    const initialQ = route.query.q;
    if (initialQ) {
        query.value = initialQ;
        handleSearch();
    }
});
async function handleSearch() {
    const q = query.value.trim();
    if (!q)
        return;
    searching.value = true;
    hasSearched.value = true;
    try {
        results.value = await api.search({
            query: q,
            local_only: localOnly.value,
            source: selectedSource.value,
            page: 1,
            page_size: 50
        });
    }
    catch (e) {
        ElMessage.error('搜索异常: ' + (e.message || e));
    }
    finally {
        searching.value = false;
    }
}
async function openDetail(item) {
    selectedGame.value = item;
    drawerVisible.value = true;
    loadingDetail.value = true;
    detailVersions.value = [];
    try {
        const res = await api.getTrainerDetail(item.detail_url);
        detailVersions.value = res.versions || [];
    }
    catch (e) {
        ElMessage.warning('获取版本详情失败: ' + (e.message || e));
    }
    finally {
        loadingDetail.value = false;
    }
}
async function downloadVersion(v) {
    downloadingVerID.value = v.id;
    try {
        await api.startDownload(v.id);
        ElMessage.success('已添加到下载中心');
    }
    catch (e) {
        ElMessage.error('创建下载失败: ' + (e.message || e));
    }
    finally {
        downloadingVerID.value = null;
    }
}
debugger; /* PartiallyEnd: #3632/scriptSetup.vue */
const __VLS_ctx = {};
let __VLS_components;
let __VLS_directives;
/** @type {__VLS_StyleScopedClasses['result-card']} */ ;
/** @type {__VLS_StyleScopedClasses['result-thumb']} */ ;
/** @type {__VLS_StyleScopedClasses['detail-title-group']} */ ;
// CSS variable injection 
// CSS variable injection end 
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "search-view" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "search-header" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "search-input-wrap" },
});
const __VLS_0 = {}.ElInput;
/** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
// @ts-ignore
const __VLS_1 = __VLS_asFunctionalComponent(__VLS_0, new __VLS_0({
    ...{ 'onKeyup': {} },
    modelValue: (__VLS_ctx.query),
    size: "large",
    placeholder: "中英双向搜索：赛博朋克2077 / Cyberpunk / sbpk / 老头环 / 荒野大镖客2...",
    clearable: true,
}));
const __VLS_2 = __VLS_1({
    ...{ 'onKeyup': {} },
    modelValue: (__VLS_ctx.query),
    size: "large",
    placeholder: "中英双向搜索：赛博朋克2077 / Cyberpunk / sbpk / 老头环 / 荒野大镖客2...",
    clearable: true,
}, ...__VLS_functionalComponentArgsRest(__VLS_1));
let __VLS_4;
let __VLS_5;
let __VLS_6;
const __VLS_7 = {
    onKeyup: (__VLS_ctx.handleSearch)
};
__VLS_3.slots.default;
{
    const { prefix: __VLS_thisSlot } = __VLS_3.slots;
    const __VLS_8 = {}.ElIcon;
    /** @type {[typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, ]} */ ;
    // @ts-ignore
    const __VLS_9 = __VLS_asFunctionalComponent(__VLS_8, new __VLS_8({}));
    const __VLS_10 = __VLS_9({}, ...__VLS_functionalComponentArgsRest(__VLS_9));
    __VLS_11.slots.default;
    const __VLS_12 = {}.Search;
    /** @type {[typeof __VLS_components.Search, ]} */ ;
    // @ts-ignore
    const __VLS_13 = __VLS_asFunctionalComponent(__VLS_12, new __VLS_12({}));
    const __VLS_14 = __VLS_13({}, ...__VLS_functionalComponentArgsRest(__VLS_13));
    var __VLS_11;
}
{
    const { append: __VLS_thisSlot } = __VLS_3.slots;
    const __VLS_16 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_17 = __VLS_asFunctionalComponent(__VLS_16, new __VLS_16({
        ...{ 'onClick': {} },
        type: "primary",
        loading: (__VLS_ctx.searching),
    }));
    const __VLS_18 = __VLS_17({
        ...{ 'onClick': {} },
        type: "primary",
        loading: (__VLS_ctx.searching),
    }, ...__VLS_functionalComponentArgsRest(__VLS_17));
    let __VLS_20;
    let __VLS_21;
    let __VLS_22;
    const __VLS_23 = {
        onClick: (__VLS_ctx.handleSearch)
    };
    __VLS_19.slots.default;
    var __VLS_19;
}
var __VLS_3;
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "filter-bar" },
});
const __VLS_24 = {}.ElCheckbox;
/** @type {[typeof __VLS_components.ElCheckbox, typeof __VLS_components.elCheckbox, typeof __VLS_components.ElCheckbox, typeof __VLS_components.elCheckbox, ]} */ ;
// @ts-ignore
const __VLS_25 = __VLS_asFunctionalComponent(__VLS_24, new __VLS_24({
    ...{ 'onChange': {} },
    modelValue: (__VLS_ctx.localOnly),
}));
const __VLS_26 = __VLS_25({
    ...{ 'onChange': {} },
    modelValue: (__VLS_ctx.localOnly),
}, ...__VLS_functionalComponentArgsRest(__VLS_25));
let __VLS_28;
let __VLS_29;
let __VLS_30;
const __VLS_31 = {
    onChange: (__VLS_ctx.handleSearch)
};
__VLS_27.slots.default;
var __VLS_27;
const __VLS_32 = {}.ElRadioGroup;
/** @type {[typeof __VLS_components.ElRadioGroup, typeof __VLS_components.elRadioGroup, typeof __VLS_components.ElRadioGroup, typeof __VLS_components.elRadioGroup, ]} */ ;
// @ts-ignore
const __VLS_33 = __VLS_asFunctionalComponent(__VLS_32, new __VLS_32({
    ...{ 'onChange': {} },
    modelValue: (__VLS_ctx.selectedSource),
    size: "small",
}));
const __VLS_34 = __VLS_33({
    ...{ 'onChange': {} },
    modelValue: (__VLS_ctx.selectedSource),
    size: "small",
}, ...__VLS_functionalComponentArgsRest(__VLS_33));
let __VLS_36;
let __VLS_37;
let __VLS_38;
const __VLS_39 = {
    onChange: (__VLS_ctx.handleSearch)
};
__VLS_35.slots.default;
const __VLS_40 = {}.ElRadioButton;
/** @type {[typeof __VLS_components.ElRadioButton, typeof __VLS_components.elRadioButton, typeof __VLS_components.ElRadioButton, typeof __VLS_components.elRadioButton, ]} */ ;
// @ts-ignore
const __VLS_41 = __VLS_asFunctionalComponent(__VLS_40, new __VLS_40({
    value: "all",
}));
const __VLS_42 = __VLS_41({
    value: "all",
}, ...__VLS_functionalComponentArgsRest(__VLS_41));
__VLS_43.slots.default;
var __VLS_43;
const __VLS_44 = {}.ElRadioButton;
/** @type {[typeof __VLS_components.ElRadioButton, typeof __VLS_components.elRadioButton, typeof __VLS_components.ElRadioButton, typeof __VLS_components.elRadioButton, ]} */ ;
// @ts-ignore
const __VLS_45 = __VLS_asFunctionalComponent(__VLS_44, new __VLS_44({
    value: "fling",
}));
const __VLS_46 = __VLS_45({
    value: "fling",
}, ...__VLS_functionalComponentArgsRest(__VLS_45));
__VLS_47.slots.default;
var __VLS_47;
var __VLS_35;
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "results-container" },
});
__VLS_asFunctionalDirective(__VLS_directives.vLoading)(null, { ...__VLS_directiveBindingRestFields, value: (__VLS_ctx.searching) }, null, null);
if (__VLS_ctx.results.length > 0) {
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "results-grid" },
    });
    for (const [item] of __VLS_getVForSourceType((__VLS_ctx.results))) {
        __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
            ...{ onClick: (...[$event]) => {
                    if (!(__VLS_ctx.results.length > 0))
                        return;
                    __VLS_ctx.openDetail(item);
                } },
            key: (item.slug),
            ...{ class: "result-card" },
        });
        __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
            ...{ class: "result-thumb" },
        });
        __VLS_asFunctionalElement(__VLS_intrinsicElements.img)({
            src: (item.cover_url || __VLS_ctx.defaultThumb),
            alt: (item.game_name_en),
        });
        if (item.is_downloaded) {
            const __VLS_48 = {}.ElTag;
            /** @type {[typeof __VLS_components.ElTag, typeof __VLS_components.elTag, typeof __VLS_components.ElTag, typeof __VLS_components.elTag, ]} */ ;
            // @ts-ignore
            const __VLS_49 = __VLS_asFunctionalComponent(__VLS_48, new __VLS_48({
                type: "success",
                size: "small",
                ...{ class: "dl-tag" },
            }));
            const __VLS_50 = __VLS_49({
                type: "success",
                size: "small",
                ...{ class: "dl-tag" },
            }, ...__VLS_functionalComponentArgsRest(__VLS_49));
            __VLS_51.slots.default;
            var __VLS_51;
        }
        else if (item.is_online_only) {
            const __VLS_52 = {}.ElTag;
            /** @type {[typeof __VLS_components.ElTag, typeof __VLS_components.elTag, typeof __VLS_components.ElTag, typeof __VLS_components.elTag, ]} */ ;
            // @ts-ignore
            const __VLS_53 = __VLS_asFunctionalComponent(__VLS_52, new __VLS_52({
                type: "info",
                size: "small",
                ...{ class: "dl-tag" },
            }));
            const __VLS_54 = __VLS_53({
                type: "info",
                size: "small",
                ...{ class: "dl-tag" },
            }, ...__VLS_functionalComponentArgsRest(__VLS_53));
            __VLS_55.slots.default;
            var __VLS_55;
        }
        __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
            ...{ class: "result-body" },
        });
        __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
            ...{ class: "result-name-cn" },
        });
        (item.game_name_cn || item.game_name_en);
        __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
            ...{ class: "result-name-en" },
        });
        (item.game_name_en);
        __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
            ...{ class: "result-meta" },
        });
        __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
            ...{ class: "source-pill" },
        });
        (item.source_id.toUpperCase());
        __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
            ...{ class: "score-tag" },
        });
        (Math.min(100, Math.round(item.score)));
    }
}
else if (!__VLS_ctx.searching && __VLS_ctx.hasSearched) {
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "empty-state" },
    });
    const __VLS_56 = {}.ElEmpty;
    /** @type {[typeof __VLS_components.ElEmpty, typeof __VLS_components.elEmpty, typeof __VLS_components.ElEmpty, typeof __VLS_components.elEmpty, ]} */ ;
    // @ts-ignore
    const __VLS_57 = __VLS_asFunctionalComponent(__VLS_56, new __VLS_56({
        description: "未检索到相关修改器资源",
    }));
    const __VLS_58 = __VLS_57({
        description: "未检索到相关修改器资源",
    }, ...__VLS_functionalComponentArgsRest(__VLS_57));
    __VLS_59.slots.default;
    {
        const { extra: __VLS_thisSlot } = __VLS_59.slots;
        __VLS_asFunctionalElement(__VLS_intrinsicElements.p, __VLS_intrinsicElements.p)({
            ...{ class: "empty-tip" },
        });
    }
    var __VLS_59;
}
const __VLS_60 = {}.ElDrawer;
/** @type {[typeof __VLS_components.ElDrawer, typeof __VLS_components.elDrawer, typeof __VLS_components.ElDrawer, typeof __VLS_components.elDrawer, ]} */ ;
// @ts-ignore
const __VLS_61 = __VLS_asFunctionalComponent(__VLS_60, new __VLS_60({
    modelValue: (__VLS_ctx.drawerVisible),
    title: "修改器详情与版本列表",
    size: "480px",
    ...{ class: "detail-drawer" },
}));
const __VLS_62 = __VLS_61({
    modelValue: (__VLS_ctx.drawerVisible),
    title: "修改器详情与版本列表",
    size: "480px",
    ...{ class: "detail-drawer" },
}, ...__VLS_functionalComponentArgsRest(__VLS_61));
__VLS_63.slots.default;
if (__VLS_ctx.selectedGame) {
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "detail-drawer-content" },
    });
    __VLS_asFunctionalDirective(__VLS_directives.vLoading)(null, { ...__VLS_directiveBindingRestFields, value: (__VLS_ctx.loadingDetail) }, null, null);
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "detail-header" },
    });
    __VLS_asFunctionalElement(__VLS_intrinsicElements.img)({
        src: (__VLS_ctx.selectedGame.cover_url || __VLS_ctx.defaultThumb),
        ...{ class: "detail-cover" },
    });
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "detail-title-group" },
    });
    __VLS_asFunctionalElement(__VLS_intrinsicElements.h3, __VLS_intrinsicElements.h3)({});
    (__VLS_ctx.selectedGame.game_name_cn || __VLS_ctx.selectedGame.game_name_en);
    __VLS_asFunctionalElement(__VLS_intrinsicElements.p, __VLS_intrinsicElements.p)({});
    (__VLS_ctx.selectedGame.game_name_en);
    const __VLS_64 = {}.ElDivider;
    /** @type {[typeof __VLS_components.ElDivider, typeof __VLS_components.elDivider, ]} */ ;
    // @ts-ignore
    const __VLS_65 = __VLS_asFunctionalComponent(__VLS_64, new __VLS_64({}));
    const __VLS_66 = __VLS_65({}, ...__VLS_functionalComponentArgsRest(__VLS_65));
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "section-sub" },
    });
    if (__VLS_ctx.detailVersions.length > 0) {
        __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
            ...{ class: "versions-list" },
        });
        for (const [v] of __VLS_getVForSourceType((__VLS_ctx.detailVersions))) {
            __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
                key: (v.version_name),
                ...{ class: "version-item" },
            });
            __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
                ...{ class: "version-name" },
            });
            (v.version_name);
            __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
                ...{ class: "version-sub" },
            });
            __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({});
            (v.file_name);
            __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
                ...{ class: "version-action" },
            });
            if (v.is_downloaded) {
                const __VLS_68 = {}.ElButton;
                /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
                // @ts-ignore
                const __VLS_69 = __VLS_asFunctionalComponent(__VLS_68, new __VLS_68({
                    type: "success",
                    size: "small",
                    disabled: true,
                }));
                const __VLS_70 = __VLS_69({
                    type: "success",
                    size: "small",
                    disabled: true,
                }, ...__VLS_functionalComponentArgsRest(__VLS_69));
                __VLS_71.slots.default;
                var __VLS_71;
            }
            else {
                const __VLS_72 = {}.ElButton;
                /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
                // @ts-ignore
                const __VLS_73 = __VLS_asFunctionalComponent(__VLS_72, new __VLS_72({
                    ...{ 'onClick': {} },
                    type: "primary",
                    size: "small",
                    loading: (__VLS_ctx.downloadingVerID === v.id),
                }));
                const __VLS_74 = __VLS_73({
                    ...{ 'onClick': {} },
                    type: "primary",
                    size: "small",
                    loading: (__VLS_ctx.downloadingVerID === v.id),
                }, ...__VLS_functionalComponentArgsRest(__VLS_73));
                let __VLS_76;
                let __VLS_77;
                let __VLS_78;
                const __VLS_79 = {
                    onClick: (...[$event]) => {
                        if (!(__VLS_ctx.selectedGame))
                            return;
                        if (!(__VLS_ctx.detailVersions.length > 0))
                            return;
                        if (!!(v.is_downloaded))
                            return;
                        __VLS_ctx.downloadVersion(v);
                    }
                };
                __VLS_75.slots.default;
                const __VLS_80 = {}.ElIcon;
                /** @type {[typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, ]} */ ;
                // @ts-ignore
                const __VLS_81 = __VLS_asFunctionalComponent(__VLS_80, new __VLS_80({}));
                const __VLS_82 = __VLS_81({}, ...__VLS_functionalComponentArgsRest(__VLS_81));
                __VLS_83.slots.default;
                const __VLS_84 = {}.Download;
                /** @type {[typeof __VLS_components.Download, ]} */ ;
                // @ts-ignore
                const __VLS_85 = __VLS_asFunctionalComponent(__VLS_84, new __VLS_84({}));
                const __VLS_86 = __VLS_85({}, ...__VLS_functionalComponentArgsRest(__VLS_85));
                var __VLS_83;
                var __VLS_75;
            }
        }
    }
    else {
        __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
            ...{ class: "text-gray text-sm" },
        });
    }
    if (__VLS_ctx.detailVersions.length > 0 && __VLS_ctx.detailVersions[0].features_text) {
        __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
            ...{ class: "mt-4" },
        });
        __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
            ...{ class: "section-sub" },
        });
        __VLS_asFunctionalElement(__VLS_intrinsicElements.pre, __VLS_intrinsicElements.pre)({
            ...{ class: "cheats-box" },
        });
        (__VLS_ctx.detailVersions[0].features_text);
    }
}
var __VLS_63;
/** @type {__VLS_StyleScopedClasses['search-view']} */ ;
/** @type {__VLS_StyleScopedClasses['search-header']} */ ;
/** @type {__VLS_StyleScopedClasses['search-input-wrap']} */ ;
/** @type {__VLS_StyleScopedClasses['filter-bar']} */ ;
/** @type {__VLS_StyleScopedClasses['results-container']} */ ;
/** @type {__VLS_StyleScopedClasses['results-grid']} */ ;
/** @type {__VLS_StyleScopedClasses['result-card']} */ ;
/** @type {__VLS_StyleScopedClasses['result-thumb']} */ ;
/** @type {__VLS_StyleScopedClasses['dl-tag']} */ ;
/** @type {__VLS_StyleScopedClasses['dl-tag']} */ ;
/** @type {__VLS_StyleScopedClasses['result-body']} */ ;
/** @type {__VLS_StyleScopedClasses['result-name-cn']} */ ;
/** @type {__VLS_StyleScopedClasses['result-name-en']} */ ;
/** @type {__VLS_StyleScopedClasses['result-meta']} */ ;
/** @type {__VLS_StyleScopedClasses['source-pill']} */ ;
/** @type {__VLS_StyleScopedClasses['score-tag']} */ ;
/** @type {__VLS_StyleScopedClasses['empty-state']} */ ;
/** @type {__VLS_StyleScopedClasses['empty-tip']} */ ;
/** @type {__VLS_StyleScopedClasses['detail-drawer']} */ ;
/** @type {__VLS_StyleScopedClasses['detail-drawer-content']} */ ;
/** @type {__VLS_StyleScopedClasses['detail-header']} */ ;
/** @type {__VLS_StyleScopedClasses['detail-cover']} */ ;
/** @type {__VLS_StyleScopedClasses['detail-title-group']} */ ;
/** @type {__VLS_StyleScopedClasses['section-sub']} */ ;
/** @type {__VLS_StyleScopedClasses['versions-list']} */ ;
/** @type {__VLS_StyleScopedClasses['version-item']} */ ;
/** @type {__VLS_StyleScopedClasses['version-name']} */ ;
/** @type {__VLS_StyleScopedClasses['version-sub']} */ ;
/** @type {__VLS_StyleScopedClasses['version-action']} */ ;
/** @type {__VLS_StyleScopedClasses['text-gray']} */ ;
/** @type {__VLS_StyleScopedClasses['text-sm']} */ ;
/** @type {__VLS_StyleScopedClasses['mt-4']} */ ;
/** @type {__VLS_StyleScopedClasses['section-sub']} */ ;
/** @type {__VLS_StyleScopedClasses['cheats-box']} */ ;
var __VLS_dollars;
const __VLS_self = (await import('vue')).defineComponent({
    setup() {
        return {
            Search: Search,
            Download: Download,
            query: query,
            localOnly: localOnly,
            selectedSource: selectedSource,
            searching: searching,
            hasSearched: hasSearched,
            results: results,
            drawerVisible: drawerVisible,
            selectedGame: selectedGame,
            loadingDetail: loadingDetail,
            detailVersions: detailVersions,
            downloadingVerID: downloadingVerID,
            defaultThumb: defaultThumb,
            handleSearch: handleSearch,
            openDetail: openDetail,
            downloadVersion: downloadVersion,
        };
    },
});
export default (await import('vue')).defineComponent({
    setup() {
        return {};
    },
});
; /* PartiallyEnd: #4569/main.vue */
//# sourceMappingURL=SearchView.vue.js.map