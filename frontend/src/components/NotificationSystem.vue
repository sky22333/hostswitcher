<template>
  <v-snackbar
    v-model="notificationStore.show"
    v-memo="[notificationStore.show, notificationStore.color, notificationStore.text]"
    :color="notificationStore.color"
    :timeout="1000"
    location="bottom right"
    :multi-line="notificationStore.text.length > 50"
    max-width="350"
    rounded="lg"
    elevation="4"
    variant="elevated"
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
/* 简洁的通知样式 */
.notification-content {
  display: flex;
  align-items: center;
  gap: 12px;
}

.notification-icon {
  flex-shrink: 0;
}

.notification-text {
  font-weight: 500;
  line-height: 1.4;
}

.notification-close {
  min-width: auto !important;
  width: 36px !important;
  height: 36px !important;
}
</style>