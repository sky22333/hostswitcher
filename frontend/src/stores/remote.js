import { defineStore } from 'pinia';
import { ref } from 'vue';

/**
 * 远程源管理存储
 * 负责管理远程hosts源的状态和操作
 */
export const useRemoteStore = defineStore('remote', () => {
  // 远程源列表
  const remoteSources = ref([]);
  // 加载状态
  const loading = ref(false);
  
  /**
   * 标准化远程源数据 - 兼容大小写字段名和各种格式
   */
  function normalizeRemoteSource(source) {
    if (!source || typeof source !== 'object') {
      console.warn('RemoteStore: 无效的远程源数据:', source);
      return null;
    }
    
    // 创建标准化的数据对象
    const normalized = {
      // ID字段兼容：id, ID
      ID: source.ID || source.id || '',
      // Name字段兼容：name, Name
      Name: source.Name || source.name || '',
      // URL字段兼容：url, URL
      URL: source.URL || source.url || '',
      // UpdateFreq字段兼容：updateFreq, UpdateFreq, update_freq
      UpdateFreq: source.UpdateFreq || source.updateFreq || source.update_freq || 'manual',
      // LastUpdatedAt字段兼容：lastUpdatedAt, LastUpdatedAt, last_updated_at
      LastUpdatedAt: source.LastUpdatedAt || source.lastUpdatedAt || source.last_updated_at || '',
      // LastContent字段兼容：lastContent, LastContent, last_content
      LastContent: source.LastContent || source.lastContent || source.last_content || '',
      // Status字段兼容：status, Status
      Status: source.Status || source.status || 'pending'
    };
    
    // 数据清洗和验证
    try {
      // 清理ID字段，移除特殊字符，但保留UUID格式
      if (normalized.ID) {
        normalized.ID = String(normalized.ID).trim();

        // 验证UUID格式（可选）
        if (!/^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i.test(normalized.ID)) {
          // ID格式可能不是标准UUID，但仍然可以使用
        }
      }
      
      // 清理Name字段，支持中文和特殊符号，但移除危险字符
      if (normalized.Name) {
        normalized.Name = String(normalized.Name)
          .trim()
          .replace(/[\x00-\x1f\x7f]/g, '') // 移除控制字符
          .substring(0, 100); // 限制长度
      }
      
      // 验证和清理URL字段
      if (normalized.URL) {
        normalized.URL = String(normalized.URL).trim();
        // 基本URL格式验证
        if (!/^https?:\/\/.+/i.test(normalized.URL)) {
          // URL格式可能不正确，但仍然可以使用
        }
      }
      
      // 标准化UpdateFreq字段
      const validFreqs = ['manual', 'startup'];
      if (!validFreqs.includes(normalized.UpdateFreq)) {
        normalized.UpdateFreq = 'manual';
      }
      
      // 标准化Status字段
      const validStatuses = ['pending', 'success', 'failed'];
      if (!validStatuses.includes(normalized.Status)) {
        normalized.Status = 'pending';
      }
      
      // 验证必需字段
      if (!normalized.ID || !normalized.Name || !normalized.URL) {
        return null;
      }
      
      return normalized;
    } catch (error) {
      console.error('RemoteStore: 数据标准化失败:', error, source);
      return null;
    }
  }
  
  /**
   * 标准化远程源数组
   */
  function normalizeRemoteSourceArray(sources) {
    if (!Array.isArray(sources)) {
      // 尝试转换单个对象为数组
      if (sources && typeof sources === 'object') {
        sources = [sources];
      } else {
        return [];
      }
    }
    
    const normalized = [];
    for (const source of sources) {
      const normalizedSource = normalizeRemoteSource(source);
      if (normalizedSource) {
        normalized.push(normalizedSource);
      }
    }
    
    return normalized;
  }
  
  /**
   * 通用错误处理
   */
  function handleError(operation, error) {
    console.error(`RemoteStore: ${operation}失败:`, error);
    const message = error.message || error.toString();
    throw new Error(message);
  }
  
  /**
   * 验证远程源ID - 兼容各种格式
   */
  function validateSourceId(id, operation = '操作') {
    if (!id || typeof id !== 'string' || id.trim() === '') {
      throw new Error('远程源ID无效');
    }
    
    // 标准化ID
    const normalizedId = String(id).trim();
    
    const source = remoteSources.value.find(s => 
      s.ID === normalizedId || 
      s.id === normalizedId ||
      String(s.ID || s.id || '').trim() === normalizedId
    );
    
    if (!source) {
      throw new Error('本地列表中未找到指定的远程源');
    }
    
    return source;
  }
  
  /**
   * 加载所有远程源
   */
  async function loadRemoteSources() {
    loading.value = true;
    try {
      const rawSources = await window.go.services.NetworkService.GetAllRemoteSources();
      
      // 使用标准化函数处理数据，兼容各种格式
      const normalizedSources = normalizeRemoteSourceArray(rawSources);
      
      // 按名称排序（支持中文排序）
      normalizedSources.sort((a, b) => {
        try {
          return a.Name.localeCompare(b.Name, 'zh-CN', { 
            numeric: true, 
            sensitivity: 'base' 
          });
        } catch (error) {
          // 降级排序
          return String(a.Name || '').localeCompare(String(b.Name || ''));
        }
      });
      
      remoteSources.value = normalizedSources;
    } catch (error) {
      console.error('RemoteStore: 加载远程源失败:', error);
      remoteSources.value = [];
      throw error;
    } finally {
      loading.value = false;
    }
  }
  
  /**
   * 添加远程源
   */
  async function addRemoteSource(name, url, updateFreq) {
    loading.value = true;
    try {
      const newSource = await window.go.services.NetworkService.AddRemoteSource(name, url, updateFreq);
      
      // 重新加载远程源列表确保数据同步
      await loadRemoteSources();
      return newSource;
    } catch (error) {
      handleError('添加远程源', error);
    } finally {
      loading.value = false;
    }
  }
  
  /**
   * 更新远程源
   */
  async function updateRemoteSource(id, name, url, updateFreq) {
    loading.value = true;
    try {
      const updatedSource = await window.go.services.NetworkService.UpdateRemoteSource(id, name, url, updateFreq);
      console.log('RemoteStore: 远程源更新成功:', updatedSource);
      
      await loadRemoteSources();
      return updatedSource;
    } catch (error) {
      handleError('更新远程源', error);
    } finally {
      loading.value = false;
    }
  }
  
  /**
   * 删除远程源
   */
  async function deleteRemoteSource(id) {
    console.log('RemoteStore: 开始删除远程源...', id);
    loading.value = true;
    try {
      await window.go.services.NetworkService.DeleteRemoteSource(id);
      console.log('RemoteStore: 远程源删除成功');
      
      await loadRemoteSources();
    } catch (error) {
      handleError('删除远程源', error);
    } finally {
      loading.value = false;
    }
  }
  
  /**
   * 获取远程hosts内容
   */
  async function fetchRemoteHosts(id) {
    console.log('RemoteStore: 开始获取远程hosts内容，ID:', id);
    
    const source = validateSourceId(id, '获取远程内容');
    console.log('RemoteStore: 找到远程源:', source.Name, source.URL);
    
    loading.value = true;
    try {
      const content = await window.go.services.NetworkService.FetchRemoteHosts(id);
      console.log('RemoteStore: 获取远程内容成功，长度:', content.length);
      return content;
    } catch (error) {
      handleError('获取远程hosts内容', error);
    } finally {
      loading.value = false;
    }
  }
  
  /**
   * 从远程源创建配置
   */
  async function createConfigFromRemote(id) {
    console.log('RemoteStore: 开始从远程源创建配置，ID:', id);
    
    const source = validateSourceId(id, '创建配置');
    console.log('RemoteStore: 找到远程源:', source.Name, source.URL);
    
    loading.value = true;
    try {
      const config = await window.go.services.NetworkService.CreateConfigFromRemote(id);
      console.log('RemoteStore: 从远程源创建配置成功');
      return config;
    } catch (error) {
      handleError('从远程源创建配置', error);
    } finally {
      loading.value = false;
    }
  }
  
  /**
   * 更新所有远程源
   */
  async function updateAllRemoteSources() {
    loading.value = true;
    try {
      await window.go.services.NetworkService.UpdateAllRemoteSources();
      await loadRemoteSources();
    } catch (error) {
      handleError('更新所有远程源', error);
    } finally {
      loading.value = false;
    }
  }
  
  /**
   * 将远程源直接应用到系统hosts文件
   */
  async function applyRemoteToSystem(id) {
    console.log('RemoteStore: 开始将远程源应用到系统，ID:', id);
    
    const source = validateSourceId(id, '应用到系统');
    console.log('RemoteStore: 找到远程源:', source.Name, source.URL);
    
    loading.value = true;
    try {
      await window.go.services.NetworkService.ApplyRemoteToSystem(id);
      console.log('RemoteStore: 应用远程源到系统成功');
    } catch (error) {
      handleError('应用远程源到系统hosts', error);
    } finally {
      loading.value = false;
    }
  }
  
  return {
    remoteSources,
    loading,
    loadRemoteSources,
    addRemoteSource,
    updateRemoteSource,
    deleteRemoteSource,
    fetchRemoteHosts,
    createConfigFromRemote,
    updateAllRemoteSources,
    applyRemoteToSystem
  };
});
