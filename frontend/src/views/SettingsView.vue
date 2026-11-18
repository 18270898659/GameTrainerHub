<template>
  <div class="settings-view">
    <div class="settings-card">
      <h3 class="settings-title">客户端与下载设置</h3>

      <el-form v-if="form" label-position="top" class="settings-form">
        <!-- 下载目录 -->
        <el-form-item label="修改器默认下载目录">
          <el-input v-model="form.download_dir" placeholder="输入下载绝对路径" />
        </el-form-item>

        <!-- 最大并发数 -->
        <el-form-item label="最大并发下载任务数">
          <div class="slider-row">
            <el-slider v-model="form.max_concurrent_downloads" :min="1" :max="5" show-stops />
            <span class="slider-val">{{ form.max_concurrent_downloads }} 个任务</span>
          </div>
        </el-form-item>

        <!-- 网络代理 -->
        <el-form-item label="网络代理设置 (用于加速访问海外资源网)">
          <el-input v-model="form.proxy_url" placeholder="如: http://127.0.0.1:7890 (留空表示不使用代理)" />
        </el-form-item>

        <!-- 自动同步设置 -->
        <el-divider />
        <el-form-item label="原网站自动后台同步">
          <div class="sync-switch-row">
            <el-switch v-model="form.auto_sync_enabled" active-text="开启定期同步" />
            <span v-if="form.auto_sync_enabled" class="sync-interval-text">
              每 {{ form.auto_sync_interval_hours }} 小时检查一次
            </span>
          </div>
        </el-form-item>

        <!-- 安全启动弹窗 -->
        <el-form-item label="启动修改器前显示安全警示卡片 (推荐开启)">
          <el-switch v-model="form.safe_launch_prompt" active-text="强制校验哈希与杀软说明弹窗" />
        </el-form-item>

        <el-form-item>
          <el-button type="primary" :loading="saving" @click="saveSettings">
            保存配置
          </el-button>
        </el-form-item>
      </el-form>
    </div>

    <!-- 关于与声明 -->
    <div class="settings-card about-card">
      <h3 class="settings-title">关于 Game Trainer Hub</h3>
      <p class="about-p">版本号: v1.0.0 (Windows x64 Release)</p>
      <p class="about-p">资源索引源: FLiNG Trainer (https://flingtrainer.com/)</p>
      <p class="about-p text-gray">
        免责声明：本客户端所索引下载之修改器均来源于原作者官方公开发布入口，仅供单机单人离线娱乐学习研究。严禁用于任何多人在线竞技作弊，客户端遵守计算机软件保护条例与原网访问礼节。
      </p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { useAppStore } from '../stores/appStore'
import { api } from '../api/wails'
import type { AppConfig } from '../types'

const store = useAppStore()
const form = ref<AppConfig | null>(null)
const saving = ref(false)

onMounted(async () => {
  await store.refreshConfig()
  if (store.config) {
    form.value = { ...store.config }
  }
})

async function saveSettings() {
  if (!form.value) return
  saving.value = true
  try {
    await api.saveConfig(form.value)
    ElMessage.success('设置已保存并生效')
    await store.refreshConfig()
  } catch (e: any) {
    ElMessage.error('保存失败: ' + (e.message || e))
  } finally {
    saving.value = false
  }
}
</script>

<style scoped>
.settings-view {
  max-width: 720px;
  display: flex;
  flex-direction: column;
  gap: 20px;
}
.settings-card {
  background: rgba(30, 41, 59, 0.5);
  border: 1px solid rgba(255, 255, 255, 0.08);
  border-radius: 8px;
  padding: 24px;
}
.settings-title {
  margin: 0 0 20px 0;
  font-size: 16px;
  font-weight: 600;
  color: #f8fafc;
}
.slider-row {
  display: flex;
  align-items: center;
  gap: 20px;
}
.slider-val {
  font-size: 13px;
  color: #38bdf8;
  font-weight: 600;
  width: 60px;
}
.sync-switch-row {
  display: flex;
  align-items: center;
  gap: 16px;
}
.sync-interval-text {
  font-size: 13px;
  color: #94a3b8;
}
.about-card {
  background: rgba(15, 23, 42, 0.4);
}
.about-p {
  font-size: 13px;
  color: #94a3b8;
  line-height: 1.6;
  margin: 6px 0;
}
.text-gray {
  color: #64748b;
  font-size: 12px;
}
</style>

