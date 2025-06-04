import { defineStore } from 'pinia';
import { ref, watch } from 'vue';
import { useTheme } from 'vuetify';

/**
 * 主题管理存储
 * 负责管理应用的主题切换和持久化
 */
export const useThemeStore = defineStore('theme', () => {
  // 当前主题
  const currentTheme = ref('lightTheme');
  let vuetifyTheme = null;
  
  /**
   * 初始化主题
   * 从本地存储加载用户偏好的主题设置
   */
  function initTheme() {
    // 获取Vuetify主题实例
    try {
      vuetifyTheme = useTheme();
    } catch (error) {
      console.warn('Vuetify theme not available during initialization');
    }
    
    const savedTheme = localStorage.getItem('theme');
    if (savedTheme && (savedTheme === 'lightTheme' || savedTheme === 'darkTheme')) {
      currentTheme.value = savedTheme;
    } else {
      // 如果没有保存的主题，则根据系统偏好设置
      const prefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches;
      currentTheme.value = prefersDark ? 'darkTheme' : 'lightTheme';
    }
    
    // 应用主题到Vuetify
    applyThemeToVuetify();
    
    // 监听主题变化
    watch(currentTheme, (newTheme) => {
      applyThemeToVuetify();
      // 应用到document的class
      updateDocumentTheme(newTheme);
    }, { immediate: true });
  }
  
  /**
   * 应用主题到Vuetify
   */
  function applyThemeToVuetify() {
    if (vuetifyTheme) {
      vuetifyTheme.global.name.value = currentTheme.value;
    }
  }
  
  /**
   * 更新document的主题class
   */
  function updateDocumentTheme(theme) {
    const html = document.documentElement;
    const body = document.body;
    
    // 移除所有主题class
    html.classList.remove('v-theme--lightTheme', 'v-theme--darkTheme');
    body.classList.remove('v-theme--lightTheme', 'v-theme--darkTheme');
    
    // 添加当前主题class
    html.classList.add(`v-theme--${theme}`);
    body.classList.add(`v-theme--${theme}`);
    
    // 设置CSS变量 - 确保正确的颜色值
    if (theme === 'darkTheme') {
      html.style.setProperty('--v-theme-background', '26, 26, 26');
      html.style.setProperty('--v-theme-surface', '42, 42, 42');
      html.style.setProperty('--v-theme-surface-bright', '58, 58, 58');
      html.style.setProperty('--v-theme-surface-variant', '58, 58, 58');
      html.style.setProperty('--v-theme-on-surface', '255, 255, 255');
      html.style.setProperty('--v-theme-primary', '96, 205, 255');
    } else {
      html.style.setProperty('--v-theme-background', '245, 245, 245');
      html.style.setProperty('--v-theme-surface', '255, 255, 255');
      html.style.setProperty('--v-theme-surface-bright', '255, 255, 255');
      html.style.setProperty('--v-theme-surface-variant', '243, 243, 243');
      html.style.setProperty('--v-theme-on-surface', '31, 31, 31');
      html.style.setProperty('--v-theme-primary', '0, 120, 212');
    }
  }
  
  /**
   * 切换主题
   * 在亮色和暗色主题之间切换
   */
  function toggleTheme() {
    const newTheme = currentTheme.value === 'lightTheme' ? 'darkTheme' : 'lightTheme';
    setTheme(newTheme);
  }
  
  /**
   * 设置特定主题
   * @param {string} theme - 主题名称 ('lightTheme' 或 'darkTheme')
   */
  function setTheme(theme) {
    if (theme === 'lightTheme' || theme === 'darkTheme') {
      currentTheme.value = theme;
      localStorage.setItem('theme', theme);
      
      // 确保Vuetify主题也更新
      if (!vuetifyTheme) {
        try {
          vuetifyTheme = useTheme();
        } catch (error) {
          console.warn('Vuetify theme not available');
        }
      }
      applyThemeToVuetify();
    }
  }
  
  return {
    currentTheme,
    initTheme,
    toggleTheme,
    setTheme
  };
});
