import { defineStore } from 'pinia';
import { ref } from 'vue';

/**
 * 通知管理存储
 * 负责管理应用的全局通知
 */
export const useNotificationStore = defineStore('notification', () => {
  // 通知状态
  const show = ref(false);
  const text = ref('');
  const color = ref('info');
  const timeout = ref(3000);
  
  /**
   * 显示通知
   * @param {string} message - 通知消息
   * @param {string} type - 通知类型 ('success', 'info', 'warning', 'error')
   * @param {number} duration - 显示时长(毫秒)
   */
  function showNotification(message, type = 'info', duration = 3000) {
    text.value = message;
    
    // 根据类型设置颜色
    switch (type) {
      case 'success':
        color.value = 'success';
        break;
      case 'warning':
        color.value = 'warning';
        break;
      case 'error':
        color.value = 'error';
        break;
      default:
        color.value = 'info';
    }
    
    timeout.value = duration;
    show.value = true;
  }
  
  return {
    show,
    text,
    color,
    timeout,
    showNotification
  };
});
