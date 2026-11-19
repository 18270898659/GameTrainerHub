/// <reference types="../../node_modules/.vue-global-types/vue_3.5_0_0_0.d.ts" />
import { onMounted, watch } from 'vue';
import { useRoute } from 'vue-router';
import { Search, Download, Star, StarFilled } from '@element-plus/icons-vue';
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
function clickVisited(vg) {
    store.executeSearch(vg.name_cn || vg.name_en);
}
debugger; /* PartiallyEnd: #3632/scriptSetup.vue */
const __VLS_ctx = {};
let __VLS_components;
let __VLS_directives;
/** @type {__VLS_StyleScopedClasses['history-tag']} */ ;
/** @type {__VLS_StyleScopedClasses['history-tag']} */ ;
/** @type {__VLS_StyleScopedClasses['visited-tag']} */ ;
/** @type {__VLS_StyleScopedClasses['clear-btn']} */ ;
/** @type {__VLS_StyleScopedClasses['result-card']} */ ;
/** @type {__VLS_StyleScopedClasses['result-thumb']} */ ;
/** @type {__VLS_StyleScopedClasses['star-btn']} */ ;
/** @type {__VLS_StyleScopedClasses['star-btn']} */ ;
/** @type {__VLS_StyleScopedClasses['drawer-title-row']} */ ;
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
if (__VLS_ctx.store.recentVisited.length > 0) {
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "history-bar visited-bar" },
    });
    __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
        ...{ class: "history-label" },
    });
    (__VLS_ctx.t('search.visitedGames'));
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "history-tags" },
    });
    for (const [vg] of __VLS_getVForSourceType((__VLS_ctx.store.recentVisited))) {
        const __VLS_41 = {}.ElTag;
        /** @type {[typeof __VLS_components.ElTag, typeof __VLS_components.elTag, typeof __VLS_components.ElTag, typeof __VLS_components.elTag, ]} */ ;
        // @ts-ignore
        const __VLS_42 = __VLS_asFunctionalComponent(__VLS_41, new __VLS_41({
            ...{ 'onClick': {} },
            key: (vg.id),
            type: "warning",
            effect: "plain",
            size: "small",
            ...{ class: "visited-tag" },
        }));
        const __VLS_43 = __VLS_42({
            ...{ 'onClick': {} },
            key: (vg.id),
            type: "warning",
            effect: "plain",
            size: "small",
            ...{ class: "visited-tag" },
        }, ...__VLS_functionalComponentArgsRest(__VLS_42));
        let __VLS_45;
        let __VLS_46;
        let __VLS_47;
        const __VLS_48 = {
            onClick: (...[$event]) => {
                if (!(__VLS_ctx.store.recentVisited.length > 0))
                    return;
                __VLS_ctx.clickVisited(vg);
            }
        };
        __VLS_44.slots.default;
        (vg.name_cn || vg.name_en);
        var __VLS_44;
    }
}
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "filter-bar" },
});
const __VLS_49 = {}.ElCheckbox;
/** @type {[typeof __VLS_components.ElCheckbox, typeof __VLS_components.elCheckbox, typeof __VLS_components.ElCheckbox, typeof __VLS_components.elCheckbox, ]} */ ;
// @ts-ignore
const __VLS_50 = __VLS_asFunctionalComponent(__VLS_49, new __VLS_49({
    ...{ 'onChange': {} },
    modelValue: (__VLS_ctx.store.localOnly),
}));
const __VLS_51 = __VLS_50({
    ...{ 'onChange': {} },
    modelValue: (__VLS_ctx.store.localOnly),
}, ...__VLS_functionalComponentArgsRest(__VLS_50));
let __VLS_53;
let __VLS_54;
let __VLS_55;
const __VLS_56 = {
    onChange: (__VLS_ctx.handleSearch)
};
__VLS_52.slots.default;
(__VLS_ctx.t('search.localOnly'));
var __VLS_52;
const __VLS_57 = {}.ElRadioGroup;
/** @type {[typeof __VLS_components.ElRadioGroup, typeof __VLS_components.elRadioGroup, typeof __VLS_components.ElRadioGroup, typeof __VLS_components.elRadioGroup, ]} */ ;
// @ts-ignore
const __VLS_58 = __VLS_asFunctionalComponent(__VLS_57, new __VLS_57({
    ...{ 'onChange': {} },
    modelValue: (__VLS_ctx.store.selectedSource),
    size: "small",
}));
const __VLS_59 = __VLS_58({
    ...{ 'onChange': {} },
    modelValue: (__VLS_ctx.store.selectedSource),
    size: "small",
}, ...__VLS_functionalComponentArgsRest(__VLS_58));
let __VLS_61;
let __VLS_62;
let __VLS_63;
const __VLS_64 = {
    onChange: (__VLS_ctx.handleSearch)
};
__VLS_60.slots.default;
const __VLS_65 = {}.ElRadioButton;
/** @type {[typeof __VLS_components.ElRadioButton, typeof __VLS_components.elRadioButton, typeof __VLS_components.ElRadioButton, typeof __VLS_components.elRadioButton, ]} */ ;
// @ts-ignore
const __VLS_66 = __VLS_asFunctionalComponent(__VLS_65, new __VLS_65({
    value: "all",
}));
const __VLS_67 = __VLS_66({
    value: "all",
}, ...__VLS_functionalComponentArgsRest(__VLS_66));
__VLS_68.slots.default;
(__VLS_ctx.t('search.allSources'));
var __VLS_68;
const __VLS_69 = {}.ElRadioButton;
/** @type {[typeof __VLS_components.ElRadioButton, typeof __VLS_components.elRadioButton, typeof __VLS_components.ElRadioButton, typeof __VLS_components.elRadioButton, ]} */ ;
// @ts-ignore
const __VLS_70 = __VLS_asFunctionalComponent(__VLS_69, new __VLS_69({
    value: "fling",
}));
const __VLS_71 = __VLS_70({
    value: "fling",
}, ...__VLS_functionalComponentArgsRest(__VLS_70));
__VLS_72.slots.default;
(__VLS_ctx.t('search.flingSource'));
var __VLS_72;
var __VLS_60;
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
            key: (item.slug || item.game_id),
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
            const __VLS_73 = {}.ElTag;
            /** @type {[typeof __VLS_components.ElTag, typeof __VLS_components.elTag, typeof __VLS_components.ElTag, typeof __VLS_components.elTag, ]} */ ;
            // @ts-ignore
            const __VLS_74 = __VLS_asFunctionalComponent(__VLS_73, new __VLS_73({
                type: "success",
                size: "small",
                ...{ class: "dl-tag" },
            }));
            const __VLS_75 = __VLS_74({
                type: "success",
                size: "small",
                ...{ class: "dl-tag" },
            }, ...__VLS_functionalComponentArgsRest(__VLS_74));
            __VLS_76.slots.default;
            (__VLS_ctx.t('search.downloaded'));
            var __VLS_76;
        }
        else if (item.is_online_only) {
            const __VLS_77 = {}.ElTag;
            /** @type {[typeof __VLS_components.ElTag, typeof __VLS_components.elTag, typeof __VLS_components.ElTag, typeof __VLS_components.elTag, ]} */ ;
            // @ts-ignore
            const __VLS_78 = __VLS_asFunctionalComponent(__VLS_77, new __VLS_77({
                type: "info",
                size: "small",
                ...{ class: "dl-tag" },
            }));
            const __VLS_79 = __VLS_78({
                type: "info",
                size: "small",
                ...{ class: "dl-tag" },
            }, ...__VLS_functionalComponentArgsRest(__VLS_78));
            __VLS_80.slots.default;
            (__VLS_ctx.t('home.tagOnline'));
            var __VLS_80;
        }
        __VLS_asFunctionalElement(__VLS_intrinsicElements.button, __VLS_intrinsicElements.button)({
            ...{ onClick: (...[$event]) => {
                    if (!(__VLS_ctx.store.results.length > 0))
                        return;
                    __VLS_ctx.store.toggleFavorite(item);
                } },
            ...{ class: "star-btn" },
            ...{ class: ({ 'is-active': item.is_favorite }) },
            title: (item.is_favorite ? __VLS_ctx.t('search.unfavorite') : __VLS_ctx.t('search.favorite')),
        });
        const __VLS_81 = {}.ElIcon;
        /** @type {[typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, ]} */ ;
        // @ts-ignore
        const __VLS_82 = __VLS_asFunctionalComponent(__VLS_81, new __VLS_81({
            size: (16),
        }));
        const __VLS_83 = __VLS_82({
            size: (16),
        }, ...__VLS_functionalComponentArgsRest(__VLS_82));
        __VLS_84.slots.default;
        if (item.is_favorite) {
            const __VLS_85 = {}.StarFilled;
            /** @type {[typeof __VLS_components.StarFilled, ]} */ ;
            // @ts-ignore
            const __VLS_86 = __VLS_asFunctionalComponent(__VLS_85, new __VLS_85({}));
            const __VLS_87 = __VLS_86({}, ...__VLS_functionalComponentArgsRest(__VLS_86));
        }
        else {
            const __VLS_89 = {}.Star;
            /** @type {[typeof __VLS_components.Star, ]} */ ;
            // @ts-ignore
            const __VLS_90 = __VLS_asFunctionalComponent(__VLS_89, new __VLS_89({}));
            const __VLS_91 = __VLS_90({}, ...__VLS_functionalComponentArgsRest(__VLS_90));
        }
        var __VLS_84;
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
        ((item.source_id || 'FLiNG').toUpperCase());
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
    const __VLS_93 = {}.ElEmpty;
    /** @type {[typeof __VLS_components.ElEmpty, typeof __VLS_components.elEmpty, typeof __VLS_components.ElEmpty, typeof __VLS_components.elEmpty, ]} */ ;
    // @ts-ignore
    const __VLS_94 = __VLS_asFunctionalComponent(__VLS_93, new __VLS_93({
        description: (__VLS_ctx.t('search.emptyTitle')),
    }));
    const __VLS_95 = __VLS_94({
        description: (__VLS_ctx.t('search.emptyTitle')),
    }, ...__VLS_functionalComponentArgsRest(__VLS_94));
    __VLS_96.slots.default;
    {
        const { extra: __VLS_thisSlot } = __VLS_96.slots;
        __VLS_asFunctionalElement(__VLS_intrinsicElements.p, __VLS_intrinsicElements.p)({
            ...{ class: "empty-tip" },
        });
        (__VLS_ctx.t('search.emptyTip'));
    }
    var __VLS_96;
}
const __VLS_97 = {}.ElDrawer;
/** @type {[typeof __VLS_components.ElDrawer, typeof __VLS_components.elDrawer, typeof __VLS_components.ElDrawer, typeof __VLS_components.elDrawer, ]} */ ;
// @ts-ignore
const __VLS_98 = __VLS_asFunctionalComponent(__VLS_97, new __VLS_97({
    modelValue: (__VLS_ctx.store.drawerVisible),
    title: (__VLS_ctx.t('search.drawerTitle')),
    size: "480px",
    ...{ class: "detail-drawer" },
}));
const __VLS_99 = __VLS_98({
    modelValue: (__VLS_ctx.store.drawerVisible),
    title: (__VLS_ctx.t('search.drawerTitle')),
    size: "480px",
    ...{ class: "detail-drawer" },
}, ...__VLS_functionalComponentArgsRest(__VLS_98));
__VLS_100.slots.default;
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
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "drawer-title-row" },
    });
    __VLS_asFunctionalElement(__VLS_intrinsicElements.h3, __VLS_intrinsicElements.h3)({});
    (__VLS_ctx.store.selectedGame.game_name_cn || __VLS_ctx.store.selectedGame.game_name_en);
    const __VLS_101 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_102 = __VLS_asFunctionalComponent(__VLS_101, new __VLS_101({
        ...{ 'onClick': {} },
        circle: true,
        size: "small",
        type: (__VLS_ctx.store.selectedGame.is_favorite ? 'warning' : 'default'),
        title: (__VLS_ctx.store.selectedGame.is_favorite ? __VLS_ctx.t('search.unfavorite') : __VLS_ctx.t('search.favorite')),
    }));
    const __VLS_103 = __VLS_102({
        ...{ 'onClick': {} },
        circle: true,
        size: "small",
        type: (__VLS_ctx.store.selectedGame.is_favorite ? 'warning' : 'default'),
        title: (__VLS_ctx.store.selectedGame.is_favorite ? __VLS_ctx.t('search.unfavorite') : __VLS_ctx.t('search.favorite')),
    }, ...__VLS_functionalComponentArgsRest(__VLS_102));
    let __VLS_105;
    let __VLS_106;
    let __VLS_107;
    const __VLS_108 = {
        onClick: (...[$event]) => {
            if (!(__VLS_ctx.store.selectedGame))
                return;
            __VLS_ctx.store.toggleFavorite(__VLS_ctx.store.selectedGame);
        }
    };
    __VLS_104.slots.default;
    const __VLS_109 = {}.ElIcon;
    /** @type {[typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, ]} */ ;
    // @ts-ignore
    const __VLS_110 = __VLS_asFunctionalComponent(__VLS_109, new __VLS_109({}));
    const __VLS_111 = __VLS_110({}, ...__VLS_functionalComponentArgsRest(__VLS_110));
    __VLS_112.slots.default;
    if (__VLS_ctx.store.selectedGame.is_favorite) {
        const __VLS_113 = {}.StarFilled;
        /** @type {[typeof __VLS_components.StarFilled, ]} */ ;
        // @ts-ignore
        const __VLS_114 = __VLS_asFunctionalComponent(__VLS_113, new __VLS_113({}));
        const __VLS_115 = __VLS_114({}, ...__VLS_functionalComponentArgsRest(__VLS_114));
    }
    else {
        const __VLS_117 = {}.Star;
        /** @type {[typeof __VLS_components.Star, ]} */ ;
        // @ts-ignore
        const __VLS_118 = __VLS_asFunctionalComponent(__VLS_117, new __VLS_117({}));
        const __VLS_119 = __VLS_118({}, ...__VLS_functionalComponentArgsRest(__VLS_118));
    }
    var __VLS_112;
    var __VLS_104;
    __VLS_asFunctionalElement(__VLS_intrinsicElements.p, __VLS_intrinsicElements.p)({});
    (__VLS_ctx.store.selectedGame.game_name_en);
    const __VLS_121 = {}.ElDivider;
    /** @type {[typeof __VLS_components.ElDivider, typeof __VLS_components.elDivider, ]} */ ;
    // @ts-ignore
    const __VLS_122 = __VLS_asFunctionalComponent(__VLS_121, new __VLS_121({}));
    const __VLS_123 = __VLS_122({}, ...__VLS_functionalComponentArgsRest(__VLS_122));
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
                const __VLS_125 = {}.ElButton;
                /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
                // @ts-ignore
                const __VLS_126 = __VLS_asFunctionalComponent(__VLS_125, new __VLS_125({
                    type: "success",
                    size: "small",
                    disabled: true,
                }));
                const __VLS_127 = __VLS_126({
                    type: "success",
                    size: "small",
                    disabled: true,
                }, ...__VLS_functionalComponentArgsRest(__VLS_126));
                __VLS_128.slots.default;
                (__VLS_ctx.t('search.downloaded'));
                var __VLS_128;
            }
            else {
                const __VLS_129 = {}.ElButton;
                /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
                // @ts-ignore
                const __VLS_130 = __VLS_asFunctionalComponent(__VLS_129, new __VLS_129({
                    ...{ 'onClick': {} },
                    type: "primary",
                    size: "small",
                    loading: (__VLS_ctx.store.downloadingVerID === v.id),
                }));
                const __VLS_131 = __VLS_130({
                    ...{ 'onClick': {} },
                    type: "primary",
                    size: "small",
                    loading: (__VLS_ctx.store.downloadingVerID === v.id),
                }, ...__VLS_functionalComponentArgsRest(__VLS_130));
                let __VLS_133;
                let __VLS_134;
                let __VLS_135;
                const __VLS_136 = {
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
                __VLS_132.slots.default;
                const __VLS_137 = {}.ElIcon;
                /** @type {[typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, ]} */ ;
                // @ts-ignore
                const __VLS_138 = __VLS_asFunctionalComponent(__VLS_137, new __VLS_137({}));
                const __VLS_139 = __VLS_138({}, ...__VLS_functionalComponentArgsRest(__VLS_138));
                __VLS_140.slots.default;
                const __VLS_141 = {}.Download;
                /** @type {[typeof __VLS_components.Download, ]} */ ;
                // @ts-ignore
                const __VLS_142 = __VLS_asFunctionalComponent(__VLS_141, new __VLS_141({}));
                const __VLS_143 = __VLS_142({}, ...__VLS_functionalComponentArgsRest(__VLS_142));
                var __VLS_140;
                (__VLS_ctx.t('search.downloadBtn'));
                var __VLS_132;
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
var __VLS_100;
/** @type {__VLS_StyleScopedClasses['search-view']} */ ;
/** @type {__VLS_StyleScopedClasses['search-header']} */ ;
/** @type {__VLS_StyleScopedClasses['search-input-wrap']} */ ;
/** @type {__VLS_StyleScopedClasses['history-bar']} */ ;
/** @type {__VLS_StyleScopedClasses['history-label']} */ ;
/** @type {__VLS_StyleScopedClasses['history-tags']} */ ;
/** @type {__VLS_StyleScopedClasses['history-tag']} */ ;
/** @type {__VLS_StyleScopedClasses['clear-btn']} */ ;
/** @type {__VLS_StyleScopedClasses['history-bar']} */ ;
/** @type {__VLS_StyleScopedClasses['visited-bar']} */ ;
/** @type {__VLS_StyleScopedClasses['history-label']} */ ;
/** @type {__VLS_StyleScopedClasses['history-tags']} */ ;
/** @type {__VLS_StyleScopedClasses['visited-tag']} */ ;
/** @type {__VLS_StyleScopedClasses['filter-bar']} */ ;
/** @type {__VLS_StyleScopedClasses['results-container']} */ ;
/** @type {__VLS_StyleScopedClasses['results-grid']} */ ;
/** @type {__VLS_StyleScopedClasses['result-card']} */ ;
/** @type {__VLS_StyleScopedClasses['result-thumb']} */ ;
/** @type {__VLS_StyleScopedClasses['dl-tag']} */ ;
/** @type {__VLS_StyleScopedClasses['dl-tag']} */ ;
/** @type {__VLS_StyleScopedClasses['star-btn']} */ ;
/** @type {__VLS_StyleScopedClasses['is-active']} */ ;
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
/** @type {__VLS_StyleScopedClasses['drawer-title-row']} */ ;
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
            Star: Star,
            StarFilled: StarFilled,
            t: t,
            store: store,
            defaultThumb: defaultThumb,
            handleSearch: handleSearch,
            clickHistory: clickHistory,
            clickVisited: clickVisited,
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