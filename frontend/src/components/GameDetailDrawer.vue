<template>
  <el-drawer
    v-model="store.drawerVisible"
    :title="t('search.drawerTitle')"
    size="480px"
    class="detail-drawer"
    :append-to-body="true"
  >
    <div v-if="store.selectedGame" v-loading="store.loadingDetail" class="detail-drawer-content">
      <div class="detail-header">
        <img :src="store.selectedGame.cover_url || DEFAULT_GAME_COVER" class="detail-cover" />
        <div class="detail-title-group">
          <div class="drawer-title-row">
            <h3>{{ store.selectedGame.game_name_cn || store.selectedGame.game_name_en }}</h3>
            <el-button
              circle
              size="small"
              :type="store.selectedGame.is_favorite ? 'warning' : 'default'"
              :title="store.selectedGame.is_favorite ? t('search.unfavorite') : t('search.favorite')"
              @click="store.toggleFavorite(store.selectedGame)"
            >
              <el-icon>
                <StarFilled v-if="store.selectedGame.is_favorite" />
                <Star v-else />
              </el-icon>
            </el-button>
          </div>
          <p>{{ store.selectedGame.game_name_en }}</p>
        </div>
      </div>

      <el-divider />

      <!-- Versions List -->
      <div class="section-sub">{{ t('search.versionList') }}</div>
      <div v-if="store.detailVersions.length > 0" class="versions-list">
        <div
          v-for="v in store.detailVersions"
          :key="v.version_name || v.id"
          class="version-item"
        >
          <div class="version-name">{{ v.version_name }}</div>
          <div class="version-sub">
            <span>{{ v.file_name }}</span>
          </div>
          <div class="version-action">
            <el-button
              v-if="v.is_downloaded"
              type="success"
              size="small"
              disabled
            >
              {{ t('search.downloaded') }}
            </el-button>
            <el-button
              v-else
              type="primary"
              size="small"
              :loading="store.downloadingVerID === v.id"
              @click="store.downloadVersion(v)"
            >
              <el-icon><Download /></el-icon> {{ t('search.downloadBtn') }}
            </el-button>
          </div>
        </div>
      </div>
      <div v-else class="text-gray text-sm">{{ t('search.noFeatures') }}</div>

      <!-- Features Cheats List -->
      <div v-if="store.detailVersions.length > 0 && store.detailVersions[0]?.features_text" class="mt-4">
        <div class="section-sub">{{ t('search.featuresList') }}</div>
        <pre class="cheats-box">{{ store.detailVersions[0]?.features_text }}</pre>
      </div>
    </div>
  </el-drawer>
</template>

<script setup lang="ts">
import { Download, Star, StarFilled } from '@element-plus/icons-vue'
import { useSearchStore } from '../stores/searchStore'
import { t } from '../i18n'
import { DEFAULT_GAME_COVER } from '../utils/assets'

defineOptions({ name: 'GameDetailDrawer' })

const store = useSearchStore()
</script>

<style scoped>
.detail-drawer-content {
  padding: 0 4px;
}
.detail-header {
  display: flex;
  gap: 16px;
  align-items: center;
}
.detail-cover {
  width: 90px;
  height: 90px;
  border-radius: 6px;
  object-fit: cover;
}
.detail-title-group {
  flex: 1;
  min-width: 0;
}
.drawer-title-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  margin-bottom: 4px;
}
.drawer-title-row h3 {
  margin: 0;
  font-size: 16px;
  color: var(--text-main);
}
.detail-title-group p {
  margin: 0;
  font-size: 12px;
  color: var(--text-muted);
}
.section-sub {
  font-size: 13px;
  font-weight: 600;
  color: var(--text-primary);
  margin-bottom: 10px;
}
.versions-list {
  display: flex;
  flex-direction: column;
  gap: 10px;
}
.version-item {
  background: var(--bg-card);
  border: 1px solid var(--border-color);
  border-radius: 6px;
  padding: 10px 12px;
}
.version-name {
  font-size: 13px;
  font-weight: 600;
  color: var(--accent-cyan);
  margin-bottom: 4px;
}
.version-sub {
  font-size: 11px;
  color: var(--text-muted);
  margin-bottom: 8px;
}
.version-action {
  display: flex;
  justify-content: flex-end;
}
.cheats-box {
  background: var(--code-bg);
  padding: 12px;
  border-radius: 6px;
  font-family: monospace;
  font-size: 12px;
  color: var(--code-text);
  max-height: 240px;
  overflow-y: auto;
  white-space: pre-wrap;
  border: 1px solid var(--border-color);
}
.text-gray {
  color: var(--text-muted);
}
.mt-4 {
  margin-top: 16px;
}
</style>
