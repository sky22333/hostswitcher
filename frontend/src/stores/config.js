import { defineStore } from 'pinia';
import { ref, computed } from 'vue';
import { withLoading, withLoadingAndReload, safeAsync } from './utils';

export const useConfigStore = defineStore('config', () => {
  const configs = ref([]);
  const activeConfig = ref(null);
  const loading = ref(false);
  const systemHostsPath = ref('');
  const isAdminMode = ref(false);
  
  const sortedConfigs = computed(() => {
    return [...configs.value].sort((a, b) => {

      if (a.IsActive && !b.IsActive) return -1;
      if (!a.IsActive && b.IsActive) return 1;

      return new Date(b.UpdatedAt.Time) - new Date(a.UpdatedAt.Time);
    });
  });
  
  async function initialize() {
    try {
      systemHostsPath.value = await window.go.services.ConfigService.GetSystemHostsPath();

      const needsAdmin = await window.go.services.ConfigService.IsAdminRequired();
      isAdminMode.value = !needsAdmin;
      await loadConfigs();
    } catch (error) {
  
    }
  }
  
  async function loadConfigs() {
    loading.value = true;
    try {
      configs.value = await window.go.services.ConfigService.GetAllConfigs();
      const active = configs.value.find(config => config.IsActive);
      if (active) {
        activeConfig.value = active;
      }
    } catch (error) {
  
    } finally {
      loading.value = false;
    }
  }
  
  async function createConfig(name, description, content) {
    return withLoadingAndReload(
      () => window.go.services.ConfigService.CreateConfig(name, description, content),
      loading,
      loadConfigs,
      () => {}
    );
  }
  
  async function updateConfig(id, name, description, content) {
    return withLoadingAndReload(
      () => window.go.services.ConfigService.UpdateConfig(id, name, description, content),
      loading,
      loadConfigs,
      () => {}
    );
  }
  
  async function deleteConfig(id) {
    return withLoadingAndReload(
      () => window.go.services.ConfigService.DeleteConfig(id),
      loading,
      loadConfigs,
      () => {}
    );
  }
  
  async function applyConfig(id) {
    return withLoadingAndReload(
      () => window.go.services.ConfigService.ApplyConfig(id),
      loading,
      loadConfigs,
      () => {}
    );
  }
  
  async function readSystemHosts() {
    return safeAsync(
      () => window.go.services.ConfigService.ReadSystemHosts(),
      () => {}
    );
  }
  
  async function writeSystemHosts(content) {
    return withLoading(
      () => window.go.services.ConfigService.WriteSystemHosts(content),
      loading,
      null,
      () => {}
    );
  }
  
  async function validateHostsContent(content) {
    try {
      await window.go.services.ConfigService.ValidateHostsContent(content);
      return true;
    } catch (error) {

      throw error;
    }
  }
  
  async function isAdminRequired() {
    try {
      return await window.go.services.ConfigService.IsAdminRequired();
    } catch (error) {

      return true; // 默认返回需要管理员权限
    }
  }
  
  async function restoreDefaultHosts() {
    loading.value = true;
    try {
      await window.go.services.ConfigService.RestoreDefaultHosts();
    } catch (error) {

      throw error;
    } finally {
      loading.value = false;
    }
  }



  async function flushDNSCache() {
    try {
      await window.go.services.ConfigService.FlushDNSCache();
    } catch (error) {

      throw error;
    }
  }
  
  return {
    configs,
    activeConfig,
    loading,
    systemHostsPath,
    sortedConfigs,
    initialize,
    loadConfigs,
    createConfig,
    updateConfig,
    deleteConfig,
    applyConfig,
    readSystemHosts,
    writeSystemHosts,
    validateHostsContent,
    isAdminRequired,
    isAdminMode,
    restoreDefaultHosts,
    flushDNSCache
  };
});
