import { defineStore } from 'pinia';
import { ref } from 'vue';

export const useRemoteStore = defineStore('remote', () => {
  const remoteSources = ref([]);
  const loading = ref(false);
  

  function normalizeRemoteSource(source) {
    if (!source || typeof source !== 'object') {
      return null;
    }
    

    const normalized = {
      ID: source.ID || source.id || '',
      Name: source.Name || source.name || '',
      URL: source.URL || source.url || '',
      UpdateFreq: source.UpdateFreq || source.updateFreq || source.update_freq || 'manual',
      LastUpdatedAt: source.LastUpdatedAt || source.lastUpdatedAt || source.last_updated_at || '',
      LastContent: source.LastContent || source.lastContent || source.last_content || '',
      Status: source.Status || source.status || 'pending'
    };
    

    try {
      if (normalized.ID) {
      normalized.ID = normalized.ID.toString().trim();
    }
    

      if (normalized.Name) {
        normalized.Name = String(normalized.Name)
          .trim()
          .replace(/[\x00-\x1f\x7f]/g, '')
          .substring(0, 100);
      }
      
      if (normalized.URL) {
        normalized.URL = String(normalized.URL).trim();
      }
      
      const validFreqs = ['manual', 'startup'];
      if (!validFreqs.includes(normalized.UpdateFreq)) {
        normalized.UpdateFreq = 'manual';
      }
      
      const validStatuses = ['pending', 'success', 'failed'];
      if (!validStatuses.includes(normalized.Status)) {
        normalized.Status = 'pending';
      }
      

      if (!normalized.ID || !normalized.Name || !normalized.URL) {
        return null;
      }
      
      return normalized;
    } catch (error) {
      return null;
    }
  }
  

  function normalizeRemoteSourceArray(sources) {
    if (!Array.isArray(sources)) {

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
  

  function handleError(error) {
    const message = error.message || error.toString();
    throw new Error(message);
  }
  

  function validateSourceId(id) {
    if (!id || typeof id !== 'string' || id.trim() === '') {
      throw new Error('远程源ID无效');
    }
    
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
      remoteSources.value = [];
      throw error;
    } finally {
      loading.value = false;
    }
  }
  

  async function addRemoteSource(name, url, updateFreq) {
    loading.value = true;
    try {
      const newSource = await window.go.services.NetworkService.AddRemoteSource(name, url, updateFreq);
      
      // 重新加载远程源列表确保数据同步
      await loadRemoteSources();
      return newSource;
    } catch (error) {
      handleError(error);
    } finally {
      loading.value = false;
    }
  }
  

  async function updateRemoteSource(id, name, url, updateFreq) {
    loading.value = true;
    try {
      const updatedSource = await window.go.services.NetworkService.UpdateRemoteSource(id, name, url, updateFreq);

      
      await loadRemoteSources();
      return updatedSource;
    } catch (error) {
      handleError(error);
    } finally {
      loading.value = false;
    }
  }
  

  async function deleteRemoteSource(id) {

    loading.value = true;
    try {
      await window.go.services.NetworkService.DeleteRemoteSource(id);

      
      await loadRemoteSources();
    } catch (error) {
      handleError(error);
    } finally {
      loading.value = false;
    }
  }
  

  async function fetchRemoteHosts(id) {
    validateSourceId(id);
    
    loading.value = true;
    try {
      const content = await window.go.services.NetworkService.FetchRemoteHosts(id);

      return content;
    } catch (error) {
      handleError(error);
    } finally {
      loading.value = false;
    }
  }
  

  async function createConfigFromRemote(id) {
    validateSourceId(id);
    
    loading.value = true;
    try {
      const config = await window.go.services.NetworkService.CreateConfigFromRemote(id);

      return config;
    } catch (error) {
      handleError(error);
    } finally {
      loading.value = false;
    }
  }
  

  async function updateAllRemoteSources() {
    loading.value = true;
    try {
      await window.go.services.NetworkService.UpdateAllRemoteSources();
      await loadRemoteSources();
    } catch (error) {
      handleError(error);
    } finally {
      loading.value = false;
    }
  }
  

  async function applyRemoteToSystem(id) {
    validateSourceId(id);
    
    loading.value = true;
    try {
      await window.go.services.NetworkService.ApplyRemoteToSystem(id);

    } catch (error) {
      handleError(error);
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
