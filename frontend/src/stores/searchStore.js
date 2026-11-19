import { defineStore } from 'pinia';
import { ref } from 'vue';
import { ElMessage } from 'element-plus';
import { api } from '../api/wails';
import { t } from '../i18n';
export const useSearchStore = defineStore('search', () => {
    const query = ref('');
    const localOnly = ref(false);
    const selectedSource = ref('all');
    const searching = ref(false);
    const hasSearched = ref(false);
    const results = ref([]);
    // 详情抽屉状态
    const drawerVisible = ref(false);
    const selectedGame = ref(null);
    const loadingDetail = ref(false);
    const detailVersions = ref([]);
    const downloadingVerID = ref(null);
    // 搜索历史与最近访问
    const searchHistory = ref([]);
    const recentVisited = ref([]);
    async function loadSearchHistory() {
        try {
            const history = await api.getSearchHistory(15);
            searchHistory.value = history || [];
        }
        catch (e) {
            console.warn('Failed to load search history from backend', e);
        }
    }
    async function loadRecentVisited() {
        try {
            const visited = await api.getRecentGames(10);
            recentVisited.value = visited || [];
        }
        catch (e) {
            console.warn('Failed to load recent games', e);
        }
    }
    async function addSearchHistory(keyword) {
        const kw = keyword.trim();
        if (!kw)
            return;
        const filtered = searchHistory.value.filter(k => k.toLowerCase() !== kw.toLowerCase());
        filtered.unshift(kw);
        if (filtered.length > 15) {
            filtered.length = 15;
        }
        searchHistory.value = filtered;
        try {
            await api.addSearchHistory(kw);
        }
        catch (e) {
            console.warn('Failed to persist search history', e);
        }
    }
    async function removeSearchHistory(keyword) {
        searchHistory.value = searchHistory.value.filter(k => k !== keyword);
        try {
            await api.removeSearchHistory(keyword);
        }
        catch (e) {
            console.warn('Failed to remove search history item', e);
        }
    }
    async function clearSearchHistory() {
        searchHistory.value = [];
        try {
            await api.clearSearchHistory();
        }
        catch (e) {
            console.warn('Failed to clear search history', e);
        }
    }
    async function executeSearch(targetQuery) {
        if (targetQuery !== undefined) {
            query.value = targetQuery;
        }
        const q = query.value.trim();
        if (!q)
            return;
        searching.value = true;
        hasSearched.value = true;
        await addSearchHistory(q);
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
            ElMessage.error('搜索异常: ' + (e?.message || e || '网络请求失败'));
        }
        finally {
            searching.value = false;
        }
    }
    async function toggleFavorite(item) {
        if (!item || !item.game_id)
            return;
        try {
            const nextStatus = await api.toggleFavorite(item.game_id);
            item.is_favorite = nextStatus;
            if (selectedGame.value && selectedGame.value.game_id === item.game_id) {
                selectedGame.value.is_favorite = nextStatus;
            }
            ElMessage.success(t(nextStatus ? 'search.favoriteAdded' : 'search.favoriteRemoved'));
        }
        catch (e) {
            ElMessage.error('收藏操作失败: ' + (e?.message || e));
        }
    }
    async function openDetail(item) {
        if (!item)
            return;
        selectedGame.value = item;
        drawerVisible.value = true;
        loadingDetail.value = true;
        detailVersions.value = [];
        // 记录最近访问
        if (item.game_id) {
            api.recordGameVisit(item.game_id).then(() => {
                loadRecentVisited();
            }).catch(err => console.warn('Record visit failed', err));
        }
        if (!item.detail_url) {
            loadingDetail.value = false;
            ElMessage.info('该资源暂无在线详情链接');
            return;
        }
        try {
            const res = await api.getTrainerDetail(item.detail_url);
            detailVersions.value = res?.versions || [];
        }
        catch (e) {
            ElMessage.warning('获取版本详情失败: ' + (e?.message || e || '未知错误'));
        }
        finally {
            loadingDetail.value = false;
        }
    }
    function closeDetail() {
        drawerVisible.value = false;
    }
    async function downloadVersion(v) {
        downloadingVerID.value = v.id;
        try {
            await api.startDownload(v.id);
            ElMessage.success(t('search.downloadStarted'));
        }
        catch (e) {
            ElMessage.error(e?.message || String(e));
        }
        finally {
            downloadingVerID.value = null;
        }
    }
    // 初始化加载历史与访问记录
    loadSearchHistory();
    loadRecentVisited();
    return {
        query,
        localOnly,
        selectedSource,
        searching,
        hasSearched,
        results,
        drawerVisible,
        selectedGame,
        loadingDetail,
        detailVersions,
        downloadingVerID,
        searchHistory,
        recentVisited,
        loadSearchHistory,
        loadRecentVisited,
        addSearchHistory,
        removeSearchHistory,
        clearSearchHistory,
        executeSearch,
        toggleFavorite,
        openDetail,
        closeDetail,
        downloadVersion
    };
});
//# sourceMappingURL=searchStore.js.map