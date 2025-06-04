<template>
  <v-snackbar
    v-model="notificationStore.show"
    :color="notificationStore.color"
    :timeout="notificationStore.timeout"
    location="top right"
    class="notification-snackbar"
    :multi-line="notificationStore.text.length > 50"
    :z-index="2000"
    elevation="0"
    no-timeout-on-hover
    variant="text"
    max-width="300"
  >
    <div class="notification-content">
      <v-icon 
        :icon="getIcon(notificationStore.color)" 
        class="notification-icon"
        size="small"
      />
      <span class="notification-text">{{ notificationStore.text }}</span>
    </div>
    
    <template #actions>
      <v-btn
        variant="text"
        size="small"
        @click="notificationStore.show = false"
        icon="mdi-close"
        class="notification-close"
      />
    </template>
  </v-snackbar>
</template>

<script setup>
import { useNotificationStore } from '@/stores/notification';

const notificationStore = useNotificationStore();

const getIcon = (color) => {
  const iconMap = {
    'success': 'mdi-check-circle',
    'error': 'mdi-alert-circle',
    'warning': 'mdi-alert',
    'info': 'mdi-information'
  };
  return iconMap[color] || 'mdi-information';
};
</script>

<style scoped>
.notification-snackbar :deep(.v-snackbar__wrapper) {
  background: transparent !important;
  box-shadow: none !important;
  border: none !important;
  border-radius: 0 !important;
  backdrop-filter: none !important;
  -webkit-backdrop-filter: none !important;
  margin-top: 80px !important;
  margin-right: 12px !important;
  width: 300px !important;
  max-width: 300px !important;
}

.notification-snackbar :deep(.v-snackbar__content) {
  background: transparent !important;
  padding: 8px !important;
  display: flex !important;
  align-items: center !important;
  gap: 12px !important;
}

.notification-snackbar :deep(.v-overlay__scrim) {
  display: none !important;
}

/* 通知内容样式 */
.notification-content {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 0;
  background: transparent;
  flex: 1;
  max-width: calc(100% - 40px);
}

.notification-icon {
  flex-shrink: 0;
}

.notification-text {
  font-size: 14px;
  font-weight: 500;
  line-height: 1.4;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.notification-close {
  opacity: 0.7;
  min-width: auto !important;
  width: 24px;
  height: 24px;
}

.notification-close:hover {
  opacity: 1;
}

/* 白天模式 */
.v-theme--lightTheme .notification-text {
  color: rgba(0, 0, 0, 0.87) !important;
}

.v-theme--lightTheme .notification-close {
  color: rgba(0, 0, 0, 0.6) !important;
}

/* 黑夜模式 */
.v-theme--darkTheme .notification-text {
  color: rgba(255, 255, 255, 0.87) !important;
}

.v-theme--darkTheme .notification-close {
  color: rgba(255, 255, 255, 0.6) !important;
}

/* 图标颜色适配 */
.notification-icon {
  color: inherit !important;
}
</style> 