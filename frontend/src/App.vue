<template>
  <v-app>
    <!-- 侧边栏 -->
    <v-navigation-drawer
      permanent
      width="240"
      class="sidebar-custom"
    >
      <!-- 导航菜单 -->
      <v-list nav density="comfortable" class="nav-list">
        <v-list-item
          v-for="(item, index) in menuItems"
          :key="index"
          :value="item.value"
          :active="activeTab === item.value"
          @click="activeTab = item.value"
          :title="item.title"
          :prepend-icon="item.icon"
          class="nav-item"
          rounded="xl"
        />
      </v-list>
    </v-navigation-drawer>

    <!-- 主内容区域 -->
    <v-main>
      <div class="main-content">
        <!-- 顶部应用栏 -->
        <v-app-bar 
          flat 
          class="main-app-bar"
          height="64"
          color="transparent"
        >
          <v-app-bar-title class="text-h6 font-weight-medium">
            {{ currentPageTitle }}
          </v-app-bar-title>
          
          <v-spacer />
          
          <!-- 应用状态指示器 -->
          <v-chip
            v-if="configStore.isAdminMode"
            color="success"
            variant="flat"
            size="small"
            prepend-icon="mdi-shield-check"
            class="mr-3"
          >
            管理员模式
          </v-chip>
          
          <v-chip
            v-else
            color="warning"
            variant="flat"
            size="small"
            prepend-icon="mdi-shield-alert"
            class="mr-3"
          >
            受限模式
          </v-chip>
        </v-app-bar>

        <!-- 页面内容 -->
        <div class="page-content">
          <Transition name="page" mode="out-in">
            <HostsEditor v-if="activeTab === 'editor'" key="editor" />
            <RemoteHosts v-else-if="activeTab === 'remote'" key="remote" />
            <BackupManager v-else-if="activeTab === 'backup'" key="backup" />
            <Settings v-else-if="activeTab === 'settings'" key="settings" />
          </Transition>
        </div>
      </div>
    </v-main>

    <!-- 通知系统 -->
    <NotificationSystem />
  </v-app>
</template>

<script setup>
import { ref, computed, onMounted, watch, nextTick } from 'vue';
import { useConfigStore } from '@/stores/config';
import { useNotificationStore } from '@/stores/notification';
import { useEventManager } from '@/utils/eventManager';

// 组件导入
import HostsEditor from '@/views/HostsEditor.vue';
import RemoteHosts from '@/views/RemoteHosts.vue';
import BackupManager from '@/views/BackupManager.vue';
import Settings from '@/views/Settings.vue';
import NotificationSystem from '@/components/NotificationSystem.vue';

// Store
const configStore = useConfigStore();
const notificationStore = useNotificationStore();

// 事件管理器
const { addWailsListener } = useEventManager();

// 响应式数据
const activeTab = ref('editor');

// 监听activeTab变化
watch(activeTab, (newTab, oldTab) => {
  nextTick(() => {
    // Tab切换完成
  });
});

// 菜单项
const menuItems = ref([
  {
    title: 'Host编辑器',
    icon: 'mdi-file-document-edit',
    value: 'editor'
  },
  {
    title: '远程源管理',
    icon: 'mdi-cloud-download',
    value: 'remote'
  },
  {
    title: '备份管理',
    icon: 'mdi-backup-restore',
    value: 'backup'
  },
  {
    title: '应用设置',
    icon: 'mdi-cog',
    value: 'settings'
  }
]);

// 计算属性
const currentPageTitle = computed(() => {
  const currentItem = menuItems.value.find(item => item.value === activeTab.value);
  return currentItem ? currentItem.title : 'Hosts 管理器';
});

// 生命周期
onMounted(async () => {
  try {
    await configStore.initialize();
    
    // 监听托盘事件 - 使用统一事件管理器
    addWailsListener('tray-refresh-remote', () => {
      // 切换到远程源页面
      activeTab.value = 'remote';
      notificationStore.showNotification('正在更新远程源...', 'info');
    });
    
    addWailsListener('tray-apply-config', (configId) => {
      if (configId) {
        configStore.applyConfig(configId).then(() => {
          notificationStore.showNotification('配置已应用', 'success');
        }).catch((error) => {
          notificationStore.showNotification('应用配置失败: ' + error.message, 'error');
        });
      }
    });
  } catch (error) {

    notificationStore.showNotification('初始化失败: ' + error.message, 'error');
  }
});
</script>

<style scoped>
/* 侧边栏样式 */
.sidebar-custom {
  border-right: 1px solid rgba(var(--v-theme-on-surface), 0.12);
  background: rgb(var(--v-theme-surface));
}

.nav-list {
  padding: 16px 8px;
}

.nav-item {
  margin-bottom: 4px;
}

.nav-item:hover {
  background: rgba(var(--v-theme-primary), 0.08);
}

.nav-item.v-list-item--active {
  background: rgba(var(--v-theme-primary), 0.12);
  color: rgb(var(--v-theme-primary));
}

/* 主内容区样式 */
.main-content {
  height: 100vh;
  display: flex;
  flex-direction: column;
}

.main-app-bar {
  background: rgba(var(--v-theme-surface), 0.8) !important;
  border-bottom: 1px solid rgba(var(--v-theme-on-surface), 0.08);
}

.page-content {
  flex: 1;
  overflow: hidden;
  padding: 16px;
}


</style>
