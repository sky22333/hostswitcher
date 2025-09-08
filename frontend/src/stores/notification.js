import { defineStore } from 'pinia';
import { ref } from 'vue';


export const useNotificationStore = defineStore('notification', () => {

  const show = ref(false);
  const text = ref('');
  const color = ref('info');
  const timeout = ref(3000);
  

  function showNotification(message, type = 'info', duration = 3000) {
    text.value = message;
    

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
