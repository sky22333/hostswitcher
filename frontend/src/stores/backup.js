import { ref, computed } from 'vue';
import { defineStore } from 'pinia';

export const useBackupStore = defineStore('backup', () => {
  // 状态
  const backups = ref([]);
  const loading = ref(false);
  const stats = ref({
    total: 0,
    automatic: 0,
    manual: 0,
    totalSize: 0
  });

  // 计算属性
  const sortedBackups = computed(() => {
    return [...backups.value].sort((a, b) => {
      const dateA = new Date(a.timestamp);
      const dateB = new Date(b.timestamp);
      return dateB.getTime() - dateA.getTime();
    });
  });

  const automaticBackups = computed(() => {
    return backups.value.filter(backup => backup.isAutomatic);
  });

  const manualBackups = computed(() => {
    return backups.value.filter(backup => !backup.isAutomatic);
  });

  // 格式化文件大小
  function formatFileSize(bytes) {
    if (bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
  }

  // 格式化相对时间
  function formatRelativeTime(timestamp) {
    const now = new Date();
    const time = new Date(timestamp);
    const diffInMinutes = Math.floor((now.getTime() - time.getTime()) / (1000 * 60));
    
    if (diffInMinutes < 1) return '刚刚';
    if (diffInMinutes < 60) return `${diffInMinutes}分钟前`;
    
    const diffInHours = Math.floor(diffInMinutes / 60);
    if (diffInHours < 24) return `${diffInHours}小时前`;
    
    const diffInDays = Math.floor(diffInHours / 24);
    if (diffInDays < 30) return `${diffInDays}天前`;
    
    const diffInMonths = Math.floor(diffInDays / 30);
    return `${diffInMonths}个月前`;
  }

  // 操作方法

  /**
   * 加载所有备份
   */
  async function loadBackups() {
    if (!window.go || !window.go.services || !window.go.services.ConfigService) {
      console.warn('Go 后端服务尚未准备好');
      return;
    }
    
    loading.value = true;
    try {
      const result = await window.go.services.ConfigService.GetAllBackups();
      backups.value = result || [];
      await loadStats();
    } catch (error) {
      console.error('加载备份失败:', error);
      // 重置为默认值而不是抛出错误
      backups.value = [];
      stats.value = {
        total: 0,
        automatic: 0,
        manual: 0,
        totalSize: 0
      };
      throw error;
    } finally {
      loading.value = false;
    }
  }

  /**
   * 加载备份统计信息
   */
  async function loadStats() {
    try {
      const result = await window.go.services.ConfigService.GetBackupStats();
      stats.value = result || {
        total: 0,
        automatic: 0,
        manual: 0,
        totalSize: 0
      };
    } catch (error) {
      console.error('加载备份统计失败:', error);
    }
  }

  /**
   * 创建手动备份
   */
  async function createBackup(description, tags = []) {
    loading.value = true;
    try {
      const backup = await window.go.services.ConfigService.CreateManualBackup(description, tags);
      if (backup) {
        await loadBackups(); // 重新加载备份列表
      }
      return backup;
    } catch (error) {
      console.error('创建备份失败:', error);
      throw error;
    } finally {
      loading.value = false;
    }
  }

  /**
   * 创建带自定义内容的手动备份
   */
  async function createBackupWithContent(description, content, tags = []) {
    loading.value = true;
    try {
      const backup = await window.go.services.ConfigService.CreateManualBackupWithContent(description, content, tags);
      if (backup) {
        await loadBackups(); // 重新加载备份列表
      }
      return backup;
    } catch (error) {
      console.error('创建备份失败:', error);
      throw error;
    } finally {
      loading.value = false;
    }
  }

  /**
   * 从备份恢复
   */
  async function restoreBackup(backupId) {
    loading.value = true;
    try {
      await window.go.services.ConfigService.RestoreFromBackup(backupId);
      await loadBackups(); // 重新加载备份列表（可能有新的自动备份）
    } catch (error) {
      console.error('恢复备份失败:', error);
      throw error;
    } finally {
      loading.value = false;
    }
  }

  /**
   * 删除备份
   */
  async function deleteBackup(backupId) {
    loading.value = true;
    try {
      await window.go.services.ConfigService.DeleteBackup(backupId);
      await loadBackups(); // 重新加载备份列表
    } catch (error) {
      console.error('删除备份失败:', error);
      throw error;
    } finally {
      loading.value = false;
    }
  }

  /**
   * 更新备份标签
   */
  async function updateBackupTags(backupId, tags) {
    try {
      await window.go.services.ConfigService.UpdateBackupTags(backupId, tags);
      await loadBackups(); // 重新加载备份列表
    } catch (error) {
      console.error('更新备份标签失败:', error);
      throw error;
    }
  }

  /**
   * 更新备份描述
   */
  async function updateBackupDescription(backupId, description) {
    try {
      await window.go.services.ConfigService.UpdateBackupDescription(backupId, description);
      await loadBackups(); // 重新加载备份列表
    } catch (error) {
      console.error('更新备份描述失败:', error);
      throw error;
    }
  }

  /**
   * 获取备份内容预览
   */
  function getBackupPreview(content, maxLines = 10) {
    const lines = content.split('\n');
    if (lines.length <= maxLines) {
      return content;
    }
    return lines.slice(0, maxLines).join('\n') + '\n... (还有 ' + (lines.length - maxLines) + ' 行)';
  }

  return {
    // 状态
    backups,
    loading,
    stats,
    
    // 计算属性
    sortedBackups,
    automaticBackups,
    manualBackups,
    
    // 工具方法
    formatFileSize,
    formatRelativeTime,
    getBackupPreview,
    
    // 操作方法
    loadBackups,
    loadStats,
    createBackup,
    createBackupWithContent,
    restoreBackup,
    deleteBackup,
    updateBackupTags,
    updateBackupDescription
  };
}); 