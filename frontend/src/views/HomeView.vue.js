/// <reference types="../../node_modules/.vue-global-types/vue_3.5_0_0_0.d.ts" />
import { ref, onMounted } from 'vue';
import { useRouter } from 'vue-router';
import { Search, Clock, Refresh, StarFilled, Compass, Cpu } from '@element-plus/icons-vue';
import { ElMessage } from 'element-plus';
import { useAppStore } from '../stores/appStore';
import { api } from '../api/wails';
import { t } from '../i18n';
const router = useRouter();
const store = useAppStore();
const heroQuery = ref('');
const loading = ref(false);
const favorites = ref([]);
const recentGames = ref([]);
const recentTrainers = ref([]);
const defaultThumb = 'https://flingtrainer.com/wp-content/uploads/2020/12/header-7-200x200.jpg';
async function loadDashboardData() {
    try {
        const [favs, visited, recTrainers] = await Promise.all([
            api.listFavorites(),
            api.getRecentGames(6),
            api.getRecentTrainers(6)
        ]);
        favorites.value = favs || [];
        recentGames.value = visited || [];
        recentTrainers.value = recTrainers || [];
    }
    catch (e) {
        console.warn('Dashboard data fetch failed', e);
    }
}
onMounted(async () => {
    loading.value = true;
    await Promise.all([
        store.refreshLatest(),
        store.refreshLocalTrainers(),
        store.refreshDownloads(),
        loadDashboardData()
    ]);
    loading.value = false;
});
function handleHeroSearch() {
    if (heroQuery.value.trim()) {
        router.push({ path: '/search', query: { q: heroQuery.value.trim() } });
    }
    else {
        router.push('/search');
    }
}
function goToDetail(item) {
    router.push({ path: '/search', query: { q: item.game_name_cn || item.game_name_en } });
}
function goToGame(gameName) {
    if (gameName) {
        router.push({ path: '/search', query: { q: gameName } });
    }
}
function goToLibrary() {
    router.push('/library');
}
function formatDate(dateStr) {
    if (!dateStr)
        return '';
    try {
        const d = new Date(dateStr);
        return `${d.getFullYear()}-${(d.getMonth() + 1).toString().padStart(2, '0')}-${d.getDate().toString().padStart(2, '0')}`;
    }
    catch {
        return dateStr;
    }
}
function formatTime(dateStr) {
    if (!dateStr)
        return '';
    try {
        const d = new Date(dateStr);
        const month = (d.getMonth() + 1).toString().padStart(2, '0');
        const day = d.getDate().toString().padStart(2, '0');
        const hour = d.getHours().toString().padStart(2, '0');
        const min = d.getMinutes().toString().padStart(2, '0');
        return `${month}-${day} ${hour}:${min}`;
    }
    catch {
        return dateStr;
    }
}
async function triggerSync() {
    store.isSyncing = true;
    try {
        const rec = await api.triggerSync();
        ElMessage.success(t('home.syncSuccess', { added: rec?.added_count || 0, updated: rec?.updated_count || 0 }));
        await store.refreshLatest();
    }
    catch (e) {
        ElMessage.error(t('home.syncFailed') + (e.message || e));
    }
    finally {
        store.isSyncing = false;
    }
}
debugger; /* PartiallyEnd: #3632/scriptSetup.vue */
const __VLS_ctx = {};
let __VLS_components;
let __VLS_directives;
/** @type {__VLS_StyleScopedClasses['hero-banner']} */ ;
/** @type {__VLS_StyleScopedClasses['favorite-card']} */ ;
/** @type {__VLS_StyleScopedClasses['list-row']} */ ;
/** @type {__VLS_StyleScopedClasses['trainer-card']} */ ;
/** @type {__VLS_StyleScopedClasses['card-thumb']} */ ;
// CSS variable injection 
// CSS variable injection end 
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "home-view" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "hero-banner" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "hero-content" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.h1, __VLS_intrinsicElements.h1)({
    ...{ class: "hero-title" },
});
(__VLS_ctx.t('home.heroTitle'));
__VLS_asFunctionalElement(__VLS_intrinsicElements.p, __VLS_intrinsicElements.p)({
    ...{ class: "hero-subtitle" },
});
(__VLS_ctx.t('home.heroSubtitle'));
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "hero-search-box" },
});
const __VLS_0 = {}.ElInput;
/** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
// @ts-ignore
const __VLS_1 = __VLS_asFunctionalComponent(__VLS_0, new __VLS_0({
    ...{ 'onKeyup': {} },
    modelValue: (__VLS_ctx.heroQuery),
    size: "large",
    placeholder: (__VLS_ctx.t('home.searchPlaceholder')),
    clearable: true,
}));
const __VLS_2 = __VLS_1({
    ...{ 'onKeyup': {} },
    modelValue: (__VLS_ctx.heroQuery),
    size: "large",
    placeholder: (__VLS_ctx.t('home.searchPlaceholder')),
    clearable: true,
}, ...__VLS_functionalComponentArgsRest(__VLS_1));
let __VLS_4;
let __VLS_5;
let __VLS_6;
const __VLS_7 = {
    onKeyup: (__VLS_ctx.handleHeroSearch)
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
    }));
    const __VLS_18 = __VLS_17({
        ...{ 'onClick': {} },
        type: "primary",
    }, ...__VLS_functionalComponentArgsRest(__VLS_17));
    let __VLS_20;
    let __VLS_21;
    let __VLS_22;
    const __VLS_23 = {
        onClick: (__VLS_ctx.handleHeroSearch)
    };
    __VLS_19.slots.default;
    (__VLS_ctx.t('home.searchBtn'));
    var __VLS_19;
}
var __VLS_3;
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "stats-grid" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "stat-card" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "stat-val" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "stat-label" },
});
(__VLS_ctx.t('home.statsIndexedGames'));
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "stat-card" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "stat-val" },
});
(__VLS_ctx.store.localTrainers.length);
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "stat-label" },
});
(__VLS_ctx.t('home.statsReadyTrainers'));
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "stat-card" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "stat-val" },
});
(__VLS_ctx.store.activeDownloadsCount);
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "stat-label" },
});
(__VLS_ctx.t('home.statsActiveDownloads'));
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "stat-card" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "stat-val" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "stat-label" },
});
(__VLS_ctx.t('home.statsOfficialSource'));
if (__VLS_ctx.favorites.length > 0) {
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "home-section" },
    });
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "section-header" },
    });
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "section-title" },
    });
    const __VLS_24 = {}.ElIcon;
    /** @type {[typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, ]} */ ;
    // @ts-ignore
    const __VLS_25 = __VLS_asFunctionalComponent(__VLS_24, new __VLS_24({
        ...{ class: "text-amber" },
    }));
    const __VLS_26 = __VLS_25({
        ...{ class: "text-amber" },
    }, ...__VLS_functionalComponentArgsRest(__VLS_25));
    __VLS_27.slots.default;
    const __VLS_28 = {}.StarFilled;
    /** @type {[typeof __VLS_components.StarFilled, ]} */ ;
    // @ts-ignore
    const __VLS_29 = __VLS_asFunctionalComponent(__VLS_28, new __VLS_28({}));
    const __VLS_30 = __VLS_29({}, ...__VLS_functionalComponentArgsRest(__VLS_29));
    var __VLS_27;
    __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({});
    (__VLS_ctx.t('home.favoritesTitle'));
    __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
        ...{ class: "count-tag" },
    });
    (__VLS_ctx.favorites.length);
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "favorites-grid" },
    });
    for (const [fav] of __VLS_getVForSourceType((__VLS_ctx.favorites))) {
        __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
            ...{ onClick: (...[$event]) => {
                    if (!(__VLS_ctx.favorites.length > 0))
                        return;
                    __VLS_ctx.goToGame(fav.name_cn || fav.name_en);
                } },
            key: (fav.id),
            ...{ class: "favorite-card" },
        });
        __VLS_asFunctionalElement(__VLS_intrinsicElements.img)({
            src: (fav.cover_url || __VLS_ctx.defaultThumb),
            ...{ class: "favorite-cover" },
        });
        __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
            ...{ class: "favorite-info" },
        });
        __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
            ...{ class: "fav-name-cn" },
        });
        (fav.name_cn || fav.name_en);
        __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
            ...{ class: "fav-name-en" },
        });
        (fav.name_en);
    }
}
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "dual-columns-grid" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "column-card" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "column-header" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "column-title" },
});
const __VLS_32 = {}.ElIcon;
/** @type {[typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, ]} */ ;
// @ts-ignore
const __VLS_33 = __VLS_asFunctionalComponent(__VLS_32, new __VLS_32({}));
const __VLS_34 = __VLS_33({}, ...__VLS_functionalComponentArgsRest(__VLS_33));
__VLS_35.slots.default;
const __VLS_36 = {}.Compass;
/** @type {[typeof __VLS_components.Compass, ]} */ ;
// @ts-ignore
const __VLS_37 = __VLS_asFunctionalComponent(__VLS_36, new __VLS_36({}));
const __VLS_38 = __VLS_37({}, ...__VLS_functionalComponentArgsRest(__VLS_37));
var __VLS_35;
__VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({});
(__VLS_ctx.t('home.recentVisited'));
if (__VLS_ctx.recentGames.length > 0) {
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "item-list" },
    });
    for (const [g] of __VLS_getVForSourceType((__VLS_ctx.recentGames))) {
        __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
            ...{ onClick: (...[$event]) => {
                    if (!(__VLS_ctx.recentGames.length > 0))
                        return;
                    __VLS_ctx.goToGame(g.name_cn || g.name_en);
                } },
            key: (g.id),
            ...{ class: "list-row" },
        });
        __VLS_asFunctionalElement(__VLS_intrinsicElements.img)({
            src: (g.cover_url || __VLS_ctx.defaultThumb),
            ...{ class: "row-thumb" },
        });
        __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
            ...{ class: "row-info" },
        });
        __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
            ...{ class: "row-title" },
        });
        (g.name_cn || g.name_en);
        __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
            ...{ class: "row-sub" },
        });
        (g.name_en);
        __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
            ...{ class: "row-time" },
        });
        (__VLS_ctx.formatTime(g.updated_at));
    }
}
else {
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "column-empty" },
    });
    (__VLS_ctx.t('home.noRecentGames'));
}
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "column-card" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "column-header" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "column-title" },
});
const __VLS_40 = {}.ElIcon;
/** @type {[typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, ]} */ ;
// @ts-ignore
const __VLS_41 = __VLS_asFunctionalComponent(__VLS_40, new __VLS_40({}));
const __VLS_42 = __VLS_41({}, ...__VLS_functionalComponentArgsRest(__VLS_41));
__VLS_43.slots.default;
const __VLS_44 = {}.Cpu;
/** @type {[typeof __VLS_components.Cpu, ]} */ ;
// @ts-ignore
const __VLS_45 = __VLS_asFunctionalComponent(__VLS_44, new __VLS_44({}));
const __VLS_46 = __VLS_45({}, ...__VLS_functionalComponentArgsRest(__VLS_45));
var __VLS_43;
__VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({});
(__VLS_ctx.t('home.recentTrainers'));
if (__VLS_ctx.recentTrainers.length > 0) {
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "item-list" },
    });
    for (const [tr] of __VLS_getVForSourceType((__VLS_ctx.recentTrainers))) {
        __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
            ...{ onClick: (__VLS_ctx.goToLibrary) },
            key: (tr.local_trainer.id),
            ...{ class: "list-row" },
        });
        __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
            ...{ class: "trainer-avatar" },
        });
        __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
            ...{ class: "row-info" },
        });
        __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
            ...{ class: "row-title" },
        });
        (tr.game?.name_cn || tr.game?.name_en || tr.local_trainer.file_path.split('\\').pop());
        __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
            ...{ class: "row-sub" },
        });
        if (tr.version?.version_name) {
            __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({});
            (tr.version.version_name);
        }
        __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({});
        (tr.local_trainer.run_count);
        __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
            ...{ class: "row-time" },
        });
        (__VLS_ctx.formatTime(tr.local_trainer.last_run_at || tr.local_trainer.created_at));
    }
}
else {
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "column-empty" },
    });
    (__VLS_ctx.t('home.noRecentTrainers'));
}
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "home-section mt-6" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "section-header" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "section-title" },
});
const __VLS_48 = {}.ElIcon;
/** @type {[typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, ]} */ ;
// @ts-ignore
const __VLS_49 = __VLS_asFunctionalComponent(__VLS_48, new __VLS_48({}));
const __VLS_50 = __VLS_49({}, ...__VLS_functionalComponentArgsRest(__VLS_49));
__VLS_51.slots.default;
const __VLS_52 = {}.Clock;
/** @type {[typeof __VLS_components.Clock, ]} */ ;
// @ts-ignore
const __VLS_53 = __VLS_asFunctionalComponent(__VLS_52, new __VLS_52({}));
const __VLS_54 = __VLS_53({}, ...__VLS_functionalComponentArgsRest(__VLS_53));
var __VLS_51;
__VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({});
(__VLS_ctx.t('home.latestTitle'));
const __VLS_56 = {}.ElButton;
/** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
// @ts-ignore
const __VLS_57 = __VLS_asFunctionalComponent(__VLS_56, new __VLS_56({
    ...{ 'onClick': {} },
    size: "small",
    loading: (__VLS_ctx.store.isSyncing),
    plain: true,
}));
const __VLS_58 = __VLS_57({
    ...{ 'onClick': {} },
    size: "small",
    loading: (__VLS_ctx.store.isSyncing),
    plain: true,
}, ...__VLS_functionalComponentArgsRest(__VLS_57));
let __VLS_60;
let __VLS_61;
let __VLS_62;
const __VLS_63 = {
    onClick: (__VLS_ctx.triggerSync)
};
__VLS_59.slots.default;
const __VLS_64 = {}.ElIcon;
/** @type {[typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, ]} */ ;
// @ts-ignore
const __VLS_65 = __VLS_asFunctionalComponent(__VLS_64, new __VLS_64({}));
const __VLS_66 = __VLS_65({}, ...__VLS_functionalComponentArgsRest(__VLS_65));
__VLS_67.slots.default;
const __VLS_68 = {}.Refresh;
/** @type {[typeof __VLS_components.Refresh, ]} */ ;
// @ts-ignore
const __VLS_69 = __VLS_asFunctionalComponent(__VLS_68, new __VLS_68({}));
const __VLS_70 = __VLS_69({}, ...__VLS_functionalComponentArgsRest(__VLS_69));
var __VLS_67;
(__VLS_ctx.t('home.syncBtn'));
var __VLS_59;
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "trainers-grid" },
});
__VLS_asFunctionalDirective(__VLS_directives.vLoading)(null, { ...__VLS_directiveBindingRestFields, value: (__VLS_ctx.loading) }, null, null);
for (const [item] of __VLS_getVForSourceType((__VLS_ctx.store.latestTrainers))) {
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ onClick: (...[$event]) => {
                __VLS_ctx.goToDetail(item);
            } },
        key: (item.slug),
        ...{ class: "trainer-card" },
    });
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "card-thumb" },
    });
    __VLS_asFunctionalElement(__VLS_intrinsicElements.img)({
        src: (item.cover_url || __VLS_ctx.defaultThumb),
        alt: (item.game_name_en),
    });
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "badge-source" },
    });
    (item.source_id.toUpperCase());
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "card-info" },
    });
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "game-name-cn" },
    });
    (item.game_name_cn || item.game_name_en);
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "game-name-en" },
    });
    (item.game_name_en);
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "card-footer" },
    });
    __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
        ...{ class: "update-time" },
    });
    (__VLS_ctx.formatDate(item.updated_at));
    if (item.is_downloaded) {
        const __VLS_72 = {}.ElTag;
        /** @type {[typeof __VLS_components.ElTag, typeof __VLS_components.elTag, typeof __VLS_components.ElTag, typeof __VLS_components.elTag, ]} */ ;
        // @ts-ignore
        const __VLS_73 = __VLS_asFunctionalComponent(__VLS_72, new __VLS_72({
            size: "small",
            type: "success",
        }));
        const __VLS_74 = __VLS_73({
            size: "small",
            type: "success",
        }, ...__VLS_functionalComponentArgsRest(__VLS_73));
        __VLS_75.slots.default;
        (__VLS_ctx.t('home.tagDownloaded'));
        var __VLS_75;
    }
    else {
        const __VLS_76 = {}.ElTag;
        /** @type {[typeof __VLS_components.ElTag, typeof __VLS_components.elTag, typeof __VLS_components.ElTag, typeof __VLS_components.elTag, ]} */ ;
        // @ts-ignore
        const __VLS_77 = __VLS_asFunctionalComponent(__VLS_76, new __VLS_76({
            size: "small",
            type: "info",
        }));
        const __VLS_78 = __VLS_77({
            size: "small",
            type: "info",
        }, ...__VLS_functionalComponentArgsRest(__VLS_77));
        __VLS_79.slots.default;
        (__VLS_ctx.t('home.tagOnline'));
        var __VLS_79;
    }
}
/** @type {__VLS_StyleScopedClasses['home-view']} */ ;
/** @type {__VLS_StyleScopedClasses['hero-banner']} */ ;
/** @type {__VLS_StyleScopedClasses['hero-content']} */ ;
/** @type {__VLS_StyleScopedClasses['hero-title']} */ ;
/** @type {__VLS_StyleScopedClasses['hero-subtitle']} */ ;
/** @type {__VLS_StyleScopedClasses['hero-search-box']} */ ;
/** @type {__VLS_StyleScopedClasses['stats-grid']} */ ;
/** @type {__VLS_StyleScopedClasses['stat-card']} */ ;
/** @type {__VLS_StyleScopedClasses['stat-val']} */ ;
/** @type {__VLS_StyleScopedClasses['stat-label']} */ ;
/** @type {__VLS_StyleScopedClasses['stat-card']} */ ;
/** @type {__VLS_StyleScopedClasses['stat-val']} */ ;
/** @type {__VLS_StyleScopedClasses['stat-label']} */ ;
/** @type {__VLS_StyleScopedClasses['stat-card']} */ ;
/** @type {__VLS_StyleScopedClasses['stat-val']} */ ;
/** @type {__VLS_StyleScopedClasses['stat-label']} */ ;
/** @type {__VLS_StyleScopedClasses['stat-card']} */ ;
/** @type {__VLS_StyleScopedClasses['stat-val']} */ ;
/** @type {__VLS_StyleScopedClasses['stat-label']} */ ;
/** @type {__VLS_StyleScopedClasses['home-section']} */ ;
/** @type {__VLS_StyleScopedClasses['section-header']} */ ;
/** @type {__VLS_StyleScopedClasses['section-title']} */ ;
/** @type {__VLS_StyleScopedClasses['text-amber']} */ ;
/** @type {__VLS_StyleScopedClasses['count-tag']} */ ;
/** @type {__VLS_StyleScopedClasses['favorites-grid']} */ ;
/** @type {__VLS_StyleScopedClasses['favorite-card']} */ ;
/** @type {__VLS_StyleScopedClasses['favorite-cover']} */ ;
/** @type {__VLS_StyleScopedClasses['favorite-info']} */ ;
/** @type {__VLS_StyleScopedClasses['fav-name-cn']} */ ;
/** @type {__VLS_StyleScopedClasses['fav-name-en']} */ ;
/** @type {__VLS_StyleScopedClasses['dual-columns-grid']} */ ;
/** @type {__VLS_StyleScopedClasses['column-card']} */ ;
/** @type {__VLS_StyleScopedClasses['column-header']} */ ;
/** @type {__VLS_StyleScopedClasses['column-title']} */ ;
/** @type {__VLS_StyleScopedClasses['item-list']} */ ;
/** @type {__VLS_StyleScopedClasses['list-row']} */ ;
/** @type {__VLS_StyleScopedClasses['row-thumb']} */ ;
/** @type {__VLS_StyleScopedClasses['row-info']} */ ;
/** @type {__VLS_StyleScopedClasses['row-title']} */ ;
/** @type {__VLS_StyleScopedClasses['row-sub']} */ ;
/** @type {__VLS_StyleScopedClasses['row-time']} */ ;
/** @type {__VLS_StyleScopedClasses['column-empty']} */ ;
/** @type {__VLS_StyleScopedClasses['column-card']} */ ;
/** @type {__VLS_StyleScopedClasses['column-header']} */ ;
/** @type {__VLS_StyleScopedClasses['column-title']} */ ;
/** @type {__VLS_StyleScopedClasses['item-list']} */ ;
/** @type {__VLS_StyleScopedClasses['list-row']} */ ;
/** @type {__VLS_StyleScopedClasses['trainer-avatar']} */ ;
/** @type {__VLS_StyleScopedClasses['row-info']} */ ;
/** @type {__VLS_StyleScopedClasses['row-title']} */ ;
/** @type {__VLS_StyleScopedClasses['row-sub']} */ ;
/** @type {__VLS_StyleScopedClasses['row-time']} */ ;
/** @type {__VLS_StyleScopedClasses['column-empty']} */ ;
/** @type {__VLS_StyleScopedClasses['home-section']} */ ;
/** @type {__VLS_StyleScopedClasses['mt-6']} */ ;
/** @type {__VLS_StyleScopedClasses['section-header']} */ ;
/** @type {__VLS_StyleScopedClasses['section-title']} */ ;
/** @type {__VLS_StyleScopedClasses['trainers-grid']} */ ;
/** @type {__VLS_StyleScopedClasses['trainer-card']} */ ;
/** @type {__VLS_StyleScopedClasses['card-thumb']} */ ;
/** @type {__VLS_StyleScopedClasses['badge-source']} */ ;
/** @type {__VLS_StyleScopedClasses['card-info']} */ ;
/** @type {__VLS_StyleScopedClasses['game-name-cn']} */ ;
/** @type {__VLS_StyleScopedClasses['game-name-en']} */ ;
/** @type {__VLS_StyleScopedClasses['card-footer']} */ ;
/** @type {__VLS_StyleScopedClasses['update-time']} */ ;
var __VLS_dollars;
const __VLS_self = (await import('vue')).defineComponent({
    setup() {
        return {
            Search: Search,
            Clock: Clock,
            Refresh: Refresh,
            StarFilled: StarFilled,
            Compass: Compass,
            Cpu: Cpu,
            t: t,
            store: store,
            heroQuery: heroQuery,
            loading: loading,
            favorites: favorites,
            recentGames: recentGames,
            recentTrainers: recentTrainers,
            defaultThumb: defaultThumb,
            handleHeroSearch: handleHeroSearch,
            goToDetail: goToDetail,
            goToGame: goToGame,
            goToLibrary: goToLibrary,
            formatDate: formatDate,
            formatTime: formatTime,
            triggerSync: triggerSync,
        };
    },
});
export default (await import('vue')).defineComponent({
    setup() {
        return {};
    },
});
; /* PartiallyEnd: #4569/main.vue */
//# sourceMappingURL=HomeView.vue.js.map