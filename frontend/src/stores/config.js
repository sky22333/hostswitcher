import { defineStore } from 'pinia';
import { ref, computed } from 'vue';

/**
 * 配置管理存储
 * 负责管理hosts配置的状态和操作
 */
export const useConfigStore = defineStore('config', () => {
  // 配置列表
  const configs = ref([]);
  // 当前激活的配置
  const activeConfig = ref(null);
  // 加载状态
  const loading = ref(false);
  // 系统hosts文件路径
  const systemHostsPath = ref('');
  // 管理员模式状态
  const isAdminMode = ref(false);
  
  // 计算属性：获取排序后的配置列表
  const sortedConfigs = computed(() => {
    return [...configs.value].sort((a, b) => {
      // 激活的配置排在最前面
      if (a.IsActive && !b.IsActive) return -1;
      if (!a.IsActive && b.IsActive) return 1;
      // 其次按更新时间排序
      return new Date(b.UpdatedAt.Time) - new Date(a.UpdatedAt.Time);
    });
  });
  
  /**
   * 初始化store
   */
  async function initialize() {
    try {
      systemHostsPath.value = await window.go.services.ConfigService.GetSystemHostsPath();
      // 检查管理员权限
      const needsAdmin = await window.go.services.ConfigService.IsAdminRequired();
      isAdminMode.value = !needsAdmin;
      await loadConfigs();
    } catch (error) {
      console.error('初始化配置store失败:', error);
    }
  }
  
  /**
   * 加载所有配置
   */
  async function loadConfigs() {
    loading.value = true;
    try {
      configs.value = await window.go.services.ConfigService.GetAllConfigs();
      const active = configs.value.find(config => config.IsActive);
      if (active) {
        activeConfig.value = active;
      }
    } catch (error) {
      console.error('加载配置失败:', error);
    } finally {
      loading.value = false;
    }
  }
  
  /**
   * 创建新配置
   * @param {string} name - 配置名称
   * @param {string} description - 配置描述
   * @param {string} content - 配置内容
   */
  async function createConfig(name, description, content) {
    loading.value = true;
    try {
      const newConfig = await window.go.services.ConfigService.CreateConfig(name, description, content);
      await loadConfigs(); // 重新加载配置列表
      return newConfig;
    } catch (error) {
      console.error('创建配置失败:', error);
      throw error;
    } finally {
      loading.value = false;
    }
  }
  
  /**
   * 更新配置
   * @param {string} id - 配置ID
   * @param {string} name - 配置名称
   * @param {string} description - 配置描述
   * @param {string} content - 配置内容
   */
  async function updateConfig(id, name, description, content) {
    loading.value = true;
    try {
      const updatedConfig = await window.go.services.ConfigService.UpdateConfig(id, name, description, content);
      await loadConfigs(); // 重新加载配置列表
      return updatedConfig;
    } catch (error) {
      console.error('更新配置失败:', error);
      throw error;
    } finally {
      loading.value = false;
    }
  }
  
  /**
   * 删除配置
   * @param {string} id - 配置ID
   */
  async function deleteConfig(id) {
    loading.value = true;
    try {
      await window.go.services.ConfigService.DeleteConfig(id);
      await loadConfigs(); // 重新加载配置列表
    } catch (error) {
      console.error('删除配置失败:', error);
      throw error;
    } finally {
      loading.value = false;
    }
  }
  
  /**
   * 应用配置
   * @param {string} id - 配置ID
   */
  async function applyConfig(id) {
    loading.value = true;
    try {
      await window.go.services.ConfigService.ApplyConfig(id);
      await loadConfigs(); // 重新加载配置列表
    } catch (error) {
      console.error('应用配置失败:', error);
      throw error;
    } finally {
      loading.value = false;
    }
  }
  
  /**
   * 读取系统hosts文件
   */
  async function readSystemHosts() {
    try {
      return await window.go.services.ConfigService.ReadSystemHosts();
    } catch (error) {
      console.error('读取系统hosts文件失败:', error);
      throw error;
    }
  }
  
  /**
   * 写入系统hosts文件
   * @param {string} content - hosts文件内容
   */
  async function writeSystemHosts(content) {
    loading.value = true;
    try {
      await window.go.services.ConfigService.WriteSystemHosts(content);
    } catch (error) {
      console.error('写入系统hosts文件失败:', error);
      throw error;
    } finally {
      loading.value = false;
    }
  }
  
  /**
   * 验证hosts文件内容
   * @param {string} content - hosts文件内容
   */
  async function validateHostsContent(content) {
    try {
      await window.go.services.ConfigService.ValidateHostsContent(content);
      return true;
    } catch (error) {
      console.error('验证hosts文件内容失败:', error);
      throw error;
    }
  }
  
  /**
   * 检查是否需要管理员权限
   */
  async function isAdminRequired() {
    try {
      return await window.go.services.ConfigService.IsAdminRequired();
    } catch (error) {
      console.error('检查管理员权限失败:', error);
      return true; // 默认返回需要管理员权限
    }
  }
  
  /**
   * 恢复默认的系统hosts文件
   */
  async function restoreDefaultHosts() {
    loading.value = true;
    try {
      await window.go.services.ConfigService.RestoreDefaultHosts();
    } catch (error) {
      console.error('恢复默认hosts文件失败:', error);
      throw error;
    } finally {
      loading.value = false;
    }
  }



  /**
   * 刷新系统DNS缓存
   */
  async function flushDNSCache() {
    try {
      await window.go.services.ConfigService.FlushDNSCache();
    } catch (error) {
      console.error('刷新DNS缓存失败:', error);
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
