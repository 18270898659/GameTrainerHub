/// <reference types="../../node_modules/.vue-global-types/vue_3.5_0_0_0.d.ts" />
import { onMounted, watch } from 'vue';
import { useRoute } from 'vue-router';
import { Search, Download } from '@element-plus/icons-vue';
import { useSearchStore } from '../stores/searchStore';
import { t } from '../i18n';
defineOptions({ name: 'SearchView' });
const route = useRoute();
const store = useSearchStore();
const defaultThumb = 'https://flingtrainer.com/wp-content/uploads/2020/12/header-7-200x200.jpg';
onMounted(() => {
    const initialQ = route.query.q;
    if (initialQ && initialQ !== store.query) {
        store.executeSearch(initialQ);
    }
});
watch(() => route.query.q, (newQ) => {
    if (newQ && typeof newQ === 'string' && newQ !== store.query) {
        store.executeSearch(newQ);
    }
});
function handleSearch() {
    store.executeSearch();
}
function clickHistory(kw) {
    store.executeSearch(kw);
}
debugger; /* PartiallyEnd: #3632/scriptSetup.vue */
const __VLS_ctx = {};
let __VLS_components;
let __VLS_directives;
/** @type {__VLS_StyleScopedClasses['history-tag']} */ ;
/** @type {__VLS_StyleScopedClasses['clear-btn']} */ ;
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
    modelValue: (__VLS_ctx.store.query),
    size: "large",
    placeholder: (__VLS_ctx.t('search.placeholder')),
    clearable: true,
}));
const __VLS_2 = __VLS_1({
    ...{ 'onKeyup': {} },
    modelValue: (__VLS_ctx.store.query),
    size: "large",
    placeholder: (__VLS_ctx.t('search.placeholder')),
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
        loading: (__VLS_ctx.store.searching),
    }));
    const __VLS_18 = __VLS_17({
        ...{ 'onClick': {} },
        type: "primary",
        loading: (__VLS_ctx.store.searching),
    }, ...__VLS_functionalComponentArgsRest(__VLS_17));
    let __VLS_20;
    let __VLS_21;
    let __VLS_22;
    const __VLS_23 = {
        onClick: (__VLS_ctx.handleSearch)
    };
    __VLS_19.slots.default;
    (__VLS_ctx.t('search.searchBtn'));
    var __VLS_19;
}
var __VLS_3;
if (__VLS_ctx.store.searchHistory.length > 0) {
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "history-bar" },
    });
    __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
        ...{ class: "history-label" },
    });
    (__VLS_ctx.t('search.recentSearch'));
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "history-tags" },
    });
    for (const [item] of __VLS_getVForSourceType((__VLS_ctx.store.searchHistory))) {
        const __VLS_24 = {}.ElTag;
        /** @type {[typeof __VLS_components.ElTag, typeof __VLS_components.elTag, typeof __VLS_components.ElTag, typeof __VLS_components.elTag, ]} */ ;
        // @ts-ignore
        const __VLS_25 = __VLS_asFunctionalComponent(__VLS_24, new __VLS_24({
            ...{ 'onClick': {} },
            ...{ 'onClose': {} },
            key: (item),
            closable: true,
            size: "small",
            ...{ class: "history-tag" },
        }));
        const __VLS_26 = __VLS_25({
            ...{ 'onClick': {} },
            ...{ 'onClose': {} },
            key: (item),
            closable: true,
            size: "small",
            ...{ class: "history-tag" },
        }, ...__VLS_functionalComponentArgsRest(__VLS_25));
        let __VLS_28;
        let __VLS_29;
        let __VLS_30;
        const __VLS_31 = {
            onClick: (...[$event]) => {
                if (!(__VLS_ctx.store.searchHistory.length > 0))
                    return;
                __VLS_ctx.clickHistory(item);
            }
        };
        const __VLS_32 = {
            onClose: (...[$event]) => {
                if (!(__VLS_ctx.store.searchHistory.length > 0))
                    return;
                __VLS_ctx.store.removeSearchHistory(item);
            }
        };
        __VLS_27.slots.default;
        (item);
        var __VLS_27;
    }
    const __VLS_33 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_34 = __VLS_asFunctionalComponent(__VLS_33, new __VLS_33({
        ...{ 'onClick': {} },
        link: true,
        type: "info",
        size: "small",
        ...{ class: "clear-btn" },
    }));
    const __VLS_35 = __VLS_34({
        ...{ 'onClick': {} },
        link: true,
        type: "info",
        size: "small",
        ...{ class: "clear-btn" },
    }, ...__VLS_functionalComponentArgsRest(__VLS_34));
    let __VLS_37;
    let __VLS_38;
    let __VLS_39;
    const __VLS_40 = {
        onClick: (__VLS_ctx.store.clearSearchHistory)
    };
    __VLS_36.slots.default;
    (__VLS_ctx.t('search.clear'));
    var __VLS_36;
}
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "filter-bar" },
});
const __VLS_41 = {}.ElCheckbox;
/** @type {[typeof __VLS_components.ElCheckbox, typeof __VLS_components.elCheckbox, typeof __VLS_components.ElCheckbox, typeof __VLS_components.elCheckbox, ]} */ ;
// @ts-ignore
const __VLS_42 = __VLS_asFunctionalComponent(__VLS_41, new __VLS_41({
    ...{ 'onChange': {} },
    modelValue: (__VLS_ctx.store.localOnly),
}));
const __VLS_43 = __VLS_42({
    ...{ 'onChange': {} },
    modelValue: (__VLS_ctx.store.localOnly),
}, ...__VLS_functionalComponentArgsRest(__VLS_42));
let __VLS_45;
let __VLS_46;
let __VLS_47;
const __VLS_48 = {
    onChange: (__VLS_ctx.handleSearch)
};
__VLS_44.slots.default;
(__VLS_ctx.t('search.localOnly'));
var __VLS_44;
const __VLS_49 = {}.ElRadioGroup;
/** @type {[typeof __VLS_components.ElRadioGroup, typeof __VLS_components.elRadioGroup, typeof __VLS_components.ElRadioGroup, typeof __VLS_components.elRadioGroup, ]} */ ;
// @ts-ignore
const __VLS_50 = __VLS_asFunctionalComponent(__VLS_49, new __VLS_49({
    ...{ 'onChange': {} },
    modelValue: (__VLS_ctx.store.selectedSource),
    size: "small",
}));
const __VLS_51 = __VLS_50({
    ...{ 'onChange': {} },
    modelValue: (__VLS_ctx.store.selectedSource),
    size: "small",
}, ...__VLS_functionalComponentArgsRest(__VLS_50));
let __VLS_53;
let __VLS_54;
let __VLS_55;
const __VLS_56 = {
    onChange: (__VLS_ctx.handleSearch)
};
__VLS_52.slots.default;
const __VLS_57 = {}.ElRadioButton;
/** @type {[typeof __VLS_components.ElRadioButton, typeof __VLS_components.elRadioButton, typeof __VLS_components.ElRadioButton, typeof __VLS_components.elRadioButton, ]} */ ;
// @ts-ignore
const __VLS_58 = __VLS_asFunctionalComponent(__VLS_57, new __VLS_57({
    value: "all",
}));
const __VLS_59 = __VLS_58({
    value: "all",
}, ...__VLS_functionalComponentArgsRest(__VLS_58));
__VLS_60.slots.default;
(__VLS_ctx.t('search.allSources'));
var __VLS_60;
const __VLS_61 = {}.ElRadioButton;
/** @type {[typeof __VLS_components.ElRadioButton, typeof __VLS_components.elRadioButton, typeof __VLS_components.ElRadioButton, typeof __VLS_components.elRadioButton, ]} */ ;
// @ts-ignore
const __VLS_62 = __VLS_asFunctionalComponent(__VLS_61, new __VLS_61({
    value: "fling",
}));
const __VLS_63 = __VLS_62({
    value: "fling",
}, ...__VLS_functionalComponentArgsRest(__VLS_62));
__VLS_64.slots.default;
(__VLS_ctx.t('search.flingSource'));
var __VLS_64;
var __VLS_52;
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "results-container" },
});
__VLS_asFunctionalDirective(__VLS_directives.vLoading)(null, { ...__VLS_directiveBindingRestFields, value: (__VLS_ctx.store.searching) }, null, null);
if (__VLS_ctx.store.results.length > 0) {
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "results-grid" },
    });
    for (const [item] of __VLS_getVForSourceType((__VLS_ctx.store.results))) {
        __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
            ...{ onClick: (...[$event]) => {
                    if (!(__VLS_ctx.store.results.length > 0))
                        return;
                    __VLS_ctx.store.openDetail(item);
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
            const __VLS_65 = {}.ElTag;
            /** @type {[typeof __VLS_components.ElTag, typeof __VLS_components.elTag, typeof __VLS_components.ElTag, typeof __VLS_components.elTag, ]} */ ;
            // @ts-ignore
            const __VLS_66 = __VLS_asFunctionalComponent(__VLS_65, new __VLS_65({
                type: "success",
                size: "small",
                ...{ class: "dl-tag" },
            }));
            const __VLS_67 = __VLS_66({
                type: "success",
                size: "small",
                ...{ class: "dl-tag" },
            }, ...__VLS_functionalComponentArgsRest(__VLS_66));
            __VLS_68.slots.default;
            (__VLS_ctx.t('search.downloaded'));
            var __VLS_68;
        }
        else if (item.is_online_only) {
            const __VLS_69 = {}.ElTag;
            /** @type {[typeof __VLS_components.ElTag, typeof __VLS_components.elTag, typeof __VLS_components.ElTag, typeof __VLS_components.elTag, ]} */ ;
            // @ts-ignore
            const __VLS_70 = __VLS_asFunctionalComponent(__VLS_69, new __VLS_69({
                type: "info",
                size: "small",
                ...{ class: "dl-tag" },
            }));
            const __VLS_71 = __VLS_70({
                type: "info",
                size: "small",
                ...{ class: "dl-tag" },
            }, ...__VLS_functionalComponentArgsRest(__VLS_70));
            __VLS_72.slots.default;
            (__VLS_ctx.t('home.tagOnline'));
            var __VLS_72;
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
        (__VLS_ctx.t('search.matchScore', { score: Math.min(100, Math.round(item.score)) }));
    }
}
else if (!__VLS_ctx.store.searching && __VLS_ctx.store.hasSearched) {
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "empty-state" },
    });
    const __VLS_73 = {}.ElEmpty;
    /** @type {[typeof __VLS_components.ElEmpty, typeof __VLS_components.elEmpty, typeof __VLS_components.ElEmpty, typeof __VLS_components.elEmpty, ]} */ ;
    // @ts-ignore
    const __VLS_74 = __VLS_asFunctionalComponent(__VLS_73, new __VLS_73({
        description: (__VLS_ctx.t('search.emptyTitle')),
    }));
    const __VLS_75 = __VLS_74({
        description: (__VLS_ctx.t('search.emptyTitle')),
    }, ...__VLS_functionalComponentArgsRest(__VLS_74));
    __VLS_76.slots.default;
    {
        const { extra: __VLS_thisSlot } = __VLS_76.slots;
        __VLS_asFunctionalElement(__VLS_intrinsicElements.p, __VLS_intrinsicElements.p)({
            ...{ class: "empty-tip" },
        });
        (__VLS_ctx.t('search.emptyTip'));
    }
    var __VLS_76;
}
const __VLS_77 = {}.ElDrawer;
/** @type {[typeof __VLS_components.ElDrawer, typeof __VLS_components.elDrawer, typeof __VLS_components.ElDrawer, typeof __VLS_components.elDrawer, ]} */ ;
// @ts-ignore
const __VLS_78 = __VLS_asFunctionalComponent(__VLS_77, new __VLS_77({
    modelValue: (__VLS_ctx.store.drawerVisible),
    title: (__VLS_ctx.t('search.drawerTitle')),
    size: "480px",
    ...{ class: "detail-drawer" },
}));
const __VLS_79 = __VLS_78({
    modelValue: (__VLS_ctx.store.drawerVisible),
    title: (__VLS_ctx.t('search.drawerTitle')),
    size: "480px",
    ...{ class: "detail-drawer" },
}, ...__VLS_functionalComponentArgsRest(__VLS_78));
__VLS_80.slots.default;
if (__VLS_ctx.store.selectedGame) {
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "detail-drawer-content" },
    });
    __VLS_asFunctionalDirective(__VLS_directives.vLoading)(null, { ...__VLS_directiveBindingRestFields, value: (__VLS_ctx.store.loadingDetail) }, null, null);
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "detail-header" },
    });
    __VLS_asFunctionalElement(__VLS_intrinsicElements.img)({
        src: (__VLS_ctx.store.selectedGame.cover_url || __VLS_ctx.defaultThumb),
        ...{ class: "detail-cover" },
    });
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "detail-title-group" },
    });
    __VLS_asFunctionalElement(__VLS_intrinsicElements.h3, __VLS_intrinsicElements.h3)({});
    (__VLS_ctx.store.selectedGame.game_name_cn || __VLS_ctx.store.selectedGame.game_name_en);
    __VLS_asFunctionalElement(__VLS_intrinsicElements.p, __VLS_intrinsicElements.p)({});
    (__VLS_ctx.store.selectedGame.game_name_en);
    const __VLS_81 = {}.ElDivider;
    /** @type {[typeof __VLS_components.ElDivider, typeof __VLS_components.elDivider, ]} */ ;
    // @ts-ignore
    const __VLS_82 = __VLS_asFunctionalComponent(__VLS_81, new __VLS_81({}));
    const __VLS_83 = __VLS_82({}, ...__VLS_functionalComponentArgsRest(__VLS_82));
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "section-sub" },
    });
    (__VLS_ctx.t('search.versionList'));
    if (__VLS_ctx.store.detailVersions.length > 0) {
        __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
            ...{ class: "versions-list" },
        });
        for (const [v] of __VLS_getVForSourceType((__VLS_ctx.store.detailVersions))) {
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
                const __VLS_85 = {}.ElButton;
                /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
                // @ts-ignore
                const __VLS_86 = __VLS_asFunctionalComponent(__VLS_85, new __VLS_85({
                    type: "success",
                    size: "small",
                    disabled: true,
                }));
                const __VLS_87 = __VLS_86({
                    type: "success",
                    size: "small",
                    disabled: true,
                }, ...__VLS_functionalComponentArgsRest(__VLS_86));
                __VLS_88.slots.default;
                (__VLS_ctx.t('search.downloaded'));
                var __VLS_88;
            }
            else {
                const __VLS_89 = {}.ElButton;
                /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
                // @ts-ignore
                const __VLS_90 = __VLS_asFunctionalComponent(__VLS_89, new __VLS_89({
                    ...{ 'onClick': {} },
                    type: "primary",
                    size: "small",
                    loading: (__VLS_ctx.store.downloadingVerID === v.id),
                }));
                const __VLS_91 = __VLS_90({
                    ...{ 'onClick': {} },
                    type: "primary",
                    size: "small",
                    loading: (__VLS_ctx.store.downloadingVerID === v.id),
                }, ...__VLS_functionalComponentArgsRest(__VLS_90));
                let __VLS_93;
                let __VLS_94;
                let __VLS_95;
                const __VLS_96 = {
                    onClick: (...[$event]) => {
                        if (!(__VLS_ctx.store.selectedGame))
                            return;
                        if (!(__VLS_ctx.store.detailVersions.length > 0))
                            return;
                        if (!!(v.is_downloaded))
                            return;
                        __VLS_ctx.store.downloadVersion(v);
                    }
                };
                __VLS_92.slots.default;
                const __VLS_97 = {}.ElIcon;
                /** @type {[typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, ]} */ ;
                // @ts-ignore
                const __VLS_98 = __VLS_asFunctionalComponent(__VLS_97, new __VLS_97({}));
                const __VLS_99 = __VLS_98({}, ...__VLS_functionalComponentArgsRest(__VLS_98));
                __VLS_100.slots.default;
                const __VLS_101 = {}.Download;
                /** @type {[typeof __VLS_components.Download, ]} */ ;
                // @ts-ignore
                const __VLS_102 = __VLS_asFunctionalComponent(__VLS_101, new __VLS_101({}));
                const __VLS_103 = __VLS_102({}, ...__VLS_functionalComponentArgsRest(__VLS_102));
                var __VLS_100;
                (__VLS_ctx.t('search.downloadBtn'));
                var __VLS_92;
            }
        }
    }
    else {
        __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
            ...{ class: "text-gray text-sm" },
        });
        (__VLS_ctx.t('search.noFeatures'));
    }
    if (__VLS_ctx.store.detailVersions.length > 0 && __VLS_ctx.store.detailVersions[0]?.features_text) {
        __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
            ...{ class: "mt-4" },
        });
        __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
            ...{ class: "section-sub" },
        });
        (__VLS_ctx.t('search.featuresList'));
        __VLS_asFunctionalElement(__VLS_intrinsicElements.pre, __VLS_intrinsicElements.pre)({
            ...{ class: "cheats-box" },
        });
        (__VLS_ctx.store.detailVersions[0]?.features_text);
    }
}
var __VLS_80;
/** @type {__VLS_StyleScopedClasses['search-view']} */ ;
/** @type {__VLS_StyleScopedClasses['search-header']} */ ;
/** @type {__VLS_StyleScopedClasses['search-input-wrap']} */ ;
/** @type {__VLS_StyleScopedClasses['history-bar']} */ ;
/** @type {__VLS_StyleScopedClasses['history-label']} */ ;
/** @type {__VLS_StyleScopedClasses['history-tags']} */ ;
/** @type {__VLS_StyleScopedClasses['history-tag']} */ ;
/** @type {__VLS_StyleScopedClasses['clear-btn']} */ ;
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
            t: t,
            store: store,
            defaultThumb: defaultThumb,
            handleSearch: handleSearch,
            clickHistory: clickHistory,
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