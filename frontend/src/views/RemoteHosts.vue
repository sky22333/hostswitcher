<template>
  <div style="height: 100%">
    <v-container fluid style="height: 100%">
      <v-row>
        <v-col cols="12">
          <v-card class="app-card mb-4">
            <v-card-title class="d-flex align-center">
              <span class="text-h5">远程 Hosts 管理</span>
              <v-spacer></v-spacer>
              <v-btn
                color="primary"
                variant="text"
                icon="mdi-refresh"
                @click="refreshRemoteSources"
                :loading="remoteStore.loading"
                :disabled="remoteStore.loading"
                title="刷新远程源"
              ></v-btn>
              <v-btn
                color="primary"
                prepend-icon="mdi-plus"
                @click="showAddDialog = true"
              >
                添加远程源
              </v-btn>
            </v-card-title>
            
            <v-card-text>
              <v-alert
                v-if="remoteStore.remoteSources.length === 0"
                type="info"
                variant="tonal"
                class="mb-4"
              >
                暂无远程源，请点击"添加远程源"按钮添加。
              </v-alert>
              
              <v-row>
                <v-col
                  v-for="source in remoteStore.remoteSources"
                  :key="source.ID"
                  cols="12"
                  md="6"
                  lg="4"
                >
                  <v-card class="app-card h-100">
                    <v-card-item>
                      <v-card-title>{{ safeDisplayText(source.Name, 30) }}</v-card-title>
                      <v-card-subtitle>
                        <v-chip
                          :color="getStatusColor(source.Status)"
                          size="small"
                          class="mr-2"
                        >
                          {{ getStatusText(source.Status) }}
                        </v-chip>
                        {{ getUpdateFreqText(source.UpdateFreq) }}
                      </v-card-subtitle>
                    </v-card-item>
                    
                    <v-card-text>
                      <p class="text-truncate">{{ safeDisplayUrl(source.URL) }}</p>
                      <p class="text-caption">
                        上次更新: {{ formatDate(source.LastUpdatedAt) }}
                      </p>
                    </v-card-text>
                    
                    <v-card-actions>
                      <v-btn
                        color="primary"
                        variant="text"
                        @click="fetchRemoteContent(source)"
                        :loading="loadingSourceId === source.ID"
                        :disabled="remoteStore.loading"
                        size="small"
                      >
                        查看内容
                      </v-btn>
                      <v-btn
                        color="warning"
                        variant="text"
                        @click="applyDirectlyToSystem(source)"
                        :loading="applyingSourceId === source.ID"
                        :disabled="remoteStore.loading"
                        title="直接应用到系统hosts文件"
                        size="small"
                      >
                        直接应用
                      </v-btn>
                      <v-spacer></v-spacer>
                      <v-menu>
                        <template v-slot:activator="{ props }">
                          <v-btn
                            icon="mdi-dots-vertical"
                            variant="text"
                            v-bind="props"
                            size="small"
                          ></v-btn>
                        </template>
                        <v-list>
                          <v-list-item
                            @click="editRemoteSource(source)"
                            prepend-icon="mdi-pencil"
                            title="编辑"
                          ></v-list-item>
                          <v-list-item
                            @click="confirmDelete(source)"
                            prepend-icon="mdi-delete"
                            title="删除"
                          ></v-list-item>
                        </v-list>
                      </v-menu>
                    </v-card-actions>
                  </v-card>
                </v-col>
              </v-row>
            </v-card-text>
            
            <v-card-actions>
              <v-spacer></v-spacer>
              <v-btn
                color="primary"
                prepend-icon="mdi-refresh"
                @click="updateAllRemoteSources"
                :loading="updatingAll"
                :disabled="updatingAll || remoteStore.remoteSources.length === 0"
              >
                更新所有远程源
              </v-btn>
            </v-card-actions>
          </v-card>
        </v-col>
      </v-row>
    </v-container>
    
    <!-- 添加/编辑远程源对话框 -->
    <v-dialog v-model="showAddDialog" max-width="500px" persistent no-click-animation>
      <v-card class="app-card">
        <v-card-title class="text-h5">
          {{ isEditing ? '编辑远程源' : '添加远程源' }}
        </v-card-title>
        <v-card-text>
          <v-form ref="form" @submit.prevent="saveRemoteSource">
            <v-container>
              <v-row>
                <v-col cols="12">
                  <v-text-field
                    v-model="sourceForm.name"
                    label="名称"
                    :rules="[
                      v => !!v || '名称不能为空',
                      v => (v && v.trim().length >= 1) || '名称不能为空',
                      v => (v && v.length <= 100) || '名称不能超过100个字符'
                    ]"
                    required
                  ></v-text-field>
                </v-col>
                <v-col cols="12">
                  <v-text-field
                    v-model="sourceForm.url"
                    label="URL"
                    :rules="[
                      v => !!v || 'URL不能为空',
                      v => (v && /^https?:\/\/.+/i.test(v.trim())) || 'URL格式不正确，必须以http://或https://开头',
                      v => (v && v.length <= 500) || 'URL不能超过500个字符'
                    ]"
                    required
                    placeholder="https://example.com/hosts"
                  ></v-text-field>
                </v-col>
                <v-col cols="12">
                  <v-select
                    v-model="sourceForm.updateFreq"
                    :items="updateFreqOptions"
                    label="更新频率"
                    item-title="text"
                    item-value="value"
                  ></v-select>
                </v-col>
              </v-row>
            </v-container>
          </v-form>
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn
            color="default"
            variant="text"
            @click="showAddDialog = false; resetForm();"
            :disabled="remoteStore.loading"
          >
            取消
          </v-btn>
          <v-btn
            color="primary"
            @click="saveRemoteSource"
            :loading="remoteStore.loading"
            :disabled="remoteStore.loading"
          >
            {{ isEditing ? '更新' : '添加' }}
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    
    <!-- 远程内容预览对话框 -->
    <v-dialog v-model="showPreviewDialog" max-width="800px">
      <v-card class="app-card">
        <v-card-title class="d-flex align-center">
          <span>远程内容预览: {{ safeDisplayText(currentSource?.Name, 40) }}</span>
          <v-spacer></v-spacer>
          <v-btn
            icon="mdi-close"
            variant="text"
            @click="showPreviewDialog = false"
          ></v-btn>
        </v-card-title>
        <v-card-text>
          <v-textarea
            v-model="remoteContent"
            readonly
            rows="15"
            auto-grow
            variant="outlined"
            class="font-monospace"
          ></v-textarea>
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn
            color="primary"
            @click="showPreviewDialog = false"
          >
            关闭
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    
    <!-- 删除确认对话框 -->
    <v-dialog v-model="showDeleteDialog" max-width="400px" persistent no-click-animation>
      <v-card class="app-card">
        <v-card-title class="text-h5">
          确认删除
        </v-card-title>
        <v-card-text>
          您确定要删除远程源 "{{ safeDisplayText(sourceToDelete?.Name, 50) }}" 吗？此操作无法撤销。
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn
            color="default"
            variant="text"
            @click="showDeleteDialog = false; sourceToDelete = null;"
            :disabled="remoteStore.loading"
          >
            取消
          </v-btn>
          <v-btn
            color="error"
            @click="deleteRemoteSource"
            :loading="remoteStore.loading"
            :disabled="remoteStore.loading"
          >
            删除
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted, onBeforeUnmount, nextTick } from 'vue';
import { useRemoteStore } from '@/stores/remote';
import { useNotificationStore } from '@/stores/notification';

// 状态管理
const remoteStore = useRemoteStore();
const notificationStore = useNotificationStore();

// 表单状态
const showAddDialog = ref(false);
const showPreviewDialog = ref(false);
const showDeleteDialog = ref(false);
const isEditing = ref(false);
const sourceForm = ref({
  id: '',
  name: '',
  url: '',
  updateFreq: 'startup',
});
const sourceToDelete = ref(null);
const form = ref(null);

// 远程内容状态
const remoteContent = ref('');
const currentSource = ref(null);
const loadingSourceId = ref(null);
const updatingAll = ref(false);
const applyingSourceId = ref(null);

// 更新频率选项
const updateFreqOptions = [
  { text: '手动', value: 'manual' },
  { text: '软件启动时', value: 'startup' },
];

// 生命周期钩子
onMounted(async () => {
  console.log('RemoteHosts组件已挂载，开始初始化...');
  
  // 强制刷新远程源列表
  await refreshRemoteSources();
  
  // 监听远程源列表变化事件
  window.runtime.EventsOn('remote-source-list-changed', () => {
    console.log('收到remote-source-list-changed事件，刷新列表');
    refreshRemoteSources();
  });
  
  // 监听远程源状态变化事件
  window.runtime.EventsOn('remote-source-status-changed', (id) => {
    console.log('收到remote-source-status-changed事件，ID:', id);
    refreshRemoteSources();
  });
  
  // 监听远程源需要更新事件
  window.runtime.EventsOn('remote-source-need-update', (id) => {
    const source = remoteStore.remoteSources.find(s => s.ID === id);
    if (source) {
      notificationStore.showNotification(`远程源 "${source.Name}" 需要更新`, 'info');
    }
  });
  
  // 监听远程源应用到系统事件
  window.runtime.EventsOn('remote-applied-to-system', (sourceName) => {
    notificationStore.showNotification(`远程源 "${sourceName}" 已成功应用到系统hosts文件`, 'success');
  });
  
  // 监听远程内容合并到系统事件
  window.runtime.EventsOn('remote-merged-to-system', (sourceName) => {
    notificationStore.showNotification(`远程内容 "${sourceName}" 已成功合并到系统hosts文件`, 'success');
  });
  
  // 监听远程源从系统清理事件
  window.runtime.EventsOn('remote-source-cleaned-from-system', (sourceName) => {
    notificationStore.showNotification(`远程源 "${sourceName}" 的内容已从系统hosts文件中清理`, 'info');
  });
  
  console.log('RemoteHosts组件初始化完成，当前远程源数量:', remoteStore.remoteSources.length);
});

onBeforeUnmount(() => {
  console.log('RemoteHosts组件即将卸载，清理事件监听');
  // 移除事件监听
  window.runtime.EventsOff('remote-source-list-changed');
  window.runtime.EventsOff('remote-source-status-changed');
  window.runtime.EventsOff('remote-source-need-update');
  window.runtime.EventsOff('remote-applied-to-system');
  window.runtime.EventsOff('remote-merged-to-system');
  window.runtime.EventsOff('remote-source-cleaned-from-system');
});

// 刷新远程源列表
async function refreshRemoteSources() {
  console.log('开始刷新远程源列表...');
  try {
    await remoteStore.loadRemoteSources();
    console.log('远程源列表刷新成功，数量:', remoteStore.remoteSources.length);
    // 强制触发响应式更新
    await nextTick();
  } catch (error) {
    console.error('刷新远程源列表失败:', error);
    notificationStore.showNotification('加载远程源失败: ' + error, 'error');
  }
}

// 获取远程内容
async function fetchRemoteContent(source) {
  console.log('获取远程内容:', source);
  if (!source || !source.ID) {
    console.error('远程源数据无效:', source);
    notificationStore.showNotification('远程源数据无效', 'error');
    return;
  }
  
  loadingSourceId.value = source.ID;
  try {
    remoteContent.value = await remoteStore.fetchRemoteHosts(source.ID);
    currentSource.value = source;
    showPreviewDialog.value = true;
  } catch (error) {
    console.error('获取远程内容失败:', error);
    notificationStore.showNotification('获取远程内容失败: ' + (error.message || error), 'error');
  } finally {
    loadingSourceId.value = null;
  }
}

// 编辑远程源
function editRemoteSource(source) {
  console.log('编辑远程源:', source);
  if (!source || !source.ID) {
    console.error('远程源数据无效:', source);
    notificationStore.showNotification('远程源数据无效', 'error');
    return;
  }
  
  isEditing.value = true;
  sourceForm.value = {
    id: source.ID,
    name: source.Name,
    url: source.URL,
    updateFreq: source.UpdateFreq,
  };
  showAddDialog.value = true;
}

// 确认删除远程源
function confirmDelete(source) {
  console.log('确认删除远程源:', source);
  if (!source || !source.ID) {
    console.error('远程源数据无效:', source);
    notificationStore.showNotification('远程源数据无效', 'error');
    return;
  }
  
  sourceToDelete.value = source;
  showDeleteDialog.value = true;
}

// 删除远程源
async function deleteRemoteSource() {
  if (!sourceToDelete.value) return;
  
  try {
    console.log('正在删除远程源:', sourceToDelete.value.ID);
    await remoteStore.deleteRemoteSource(sourceToDelete.value.ID);
    notificationStore.showNotification('远程源已成功删除', 'success');
    showDeleteDialog.value = false;
    sourceToDelete.value = null;
    
    // 强制刷新UI
    await nextTick();
    await refreshRemoteSources();
  } catch (error) {
    console.error('删除远程源失败:', error);
    notificationStore.showNotification('删除远程源失败: ' + (error.message || error), 'error');
  }
}

// 保存远程源
async function saveRemoteSource() {
  if (!form.value) return;
  
  const { valid } = await form.value.validate();
  if (!valid) return;
  
  try {
    if (isEditing.value) {
      // 确保编辑时ID不为空
      if (!sourceForm.value.id || sourceForm.value.id.trim() === '') {
        notificationStore.showNotification('编辑远程源时ID不能为空', 'error');
        return;
      }
      
      console.log('正在更新远程源:', sourceForm.value);
      await remoteStore.updateRemoteSource(
        sourceForm.value.id,
        sourceForm.value.name,
        sourceForm.value.url,
        sourceForm.value.updateFreq
      );
      notificationStore.showNotification('远程源已成功更新', 'success');
    } else {
      console.log('正在添加远程源:', sourceForm.value);
      await remoteStore.addRemoteSource(
        sourceForm.value.name,
        sourceForm.value.url,
        sourceForm.value.updateFreq
      );
      notificationStore.showNotification('远程源已成功添加', 'success');
    }
    
    showAddDialog.value = false;
    resetForm();
    
    // 强制刷新UI
    await nextTick();
    await refreshRemoteSources();
  } catch (error) {
    console.error('保存远程源失败:', error);
    notificationStore.showNotification(
      (isEditing.value ? '更新' : '添加') + '远程源失败: ' + (error.message || error),
      'error'
    );
  }
}

// 更新所有远程源
async function updateAllRemoteSources() {
  updatingAll.value = true;
  try {
    await remoteStore.updateAllRemoteSources();
    notificationStore.showNotification('所有远程源已成功更新', 'success');
  } catch (error) {
    notificationStore.showNotification('更新远程源失败: ' + error, 'error');
  } finally {
    updatingAll.value = false;
  }
}

// 重置表单
function resetForm() {
  sourceForm.value = {
    id: '',
    name: '',
    url: '',
    updateFreq: 'startup',
  };
  isEditing.value = false;
}

// 获取状态颜色
function getStatusColor(status) {
  switch (status) {
    case 'success': return 'success';
    case 'failed': return 'error';
    case 'pending': return 'warning';
    default: return 'default';
  }
}

// 获取状态文本
function getStatusText(status) {
  switch (status) {
    case 'success': return '成功';
    case 'failed': return '失败';
    case 'pending': return '等待中';
    default: return '未知';
  }
}

// 获取更新频率文本
function getUpdateFreqText(freq) {
  switch (freq) {
    case 'startup': return '软件启动时更新';
    case 'manual': return '手动更新';
    default: return '未知';
  }
}

// 安全显示文本 - 防止XSS和处理特殊字符
function safeDisplayText(text, maxLength = 50) {
  if (!text) return '';
  
  try {
    // 转换为字符串并清理
    const cleaned = String(text)
      .trim()
      .replace(/[\x00-\x08\x0B\x0C\x0E-\x1F\x7F]/g, '') // 移除控制字符，保留换行符
      .substring(0, maxLength);
    
    return cleaned;
  } catch (error) {
    console.warn('RemoteHosts: 文本处理失败:', error, text);
    return String(text || '').substring(0, maxLength);
  }
}

// 安全显示URL
function safeDisplayUrl(url) {
  if (!url) return '';
  
  try {
    const cleaned = String(url).trim();
    // 截断超长URL但保留协议部分
    if (cleaned.length > 60) {
      const protocolEnd = cleaned.indexOf('://') + 3;
      const start = cleaned.substring(0, protocolEnd + 15);
      const end = cleaned.substring(cleaned.length - 20);
      return `${start}...${end}`;
    }
    return cleaned;
  } catch (error) {
    console.warn('RemoteHosts: URL处理失败:', error, url);
    return String(url || '');
  }
}

// 格式化日期 - 支持各种时间格式
function formatDate(dateString) {
  if (!dateString) return '从未更新';
  
  try {
    // 处理各种可能的日期格式
    let date;
    if (dateString instanceof Date) {
      date = dateString;
    } else if (typeof dateString === 'string') {
      // 标准化时间字符串
      const cleanDateString = String(dateString).trim();
      if (!cleanDateString) return '从未更新';
      
      date = new Date(cleanDateString);
    } else {
      return '时间格式错误';
    }
    
    // 验证日期有效性
    if (isNaN(date.getTime())) {
      console.warn('RemoteHosts: 无效的日期格式:', dateString);
      return '时间格式错误';
    }
    
    // 使用中文格式
    return date.toLocaleString('zh-CN', {
      year: 'numeric',
      month: '2-digit',
      day: '2-digit',
      hour: '2-digit',
      minute: '2-digit',
      second: '2-digit',
      timeZone: 'Asia/Shanghai'
    });
  } catch (error) {
    console.warn('RemoteHosts: 日期格式化失败:', error, dateString);
    return '时间格式错误';
  }
}

// 直接应用到系统hosts文件
async function applyDirectlyToSystem(source) {
  console.log('直接应用到系统:', source);
  if (!source || !source.ID) {
    console.error('远程源数据无效:', source);
    notificationStore.showNotification('远程源数据无效', 'error');
    return;
  }
  
  applyingSourceId.value = source.ID;
  try {
    await remoteStore.applyRemoteToSystem(source.ID);
    notificationStore.showNotification(`远程源 "${source.Name}" 已成功应用到系统hosts文件`, 'success');
  } catch (error) {
    console.error('应用到系统hosts文件失败:', error);
    notificationStore.showNotification('应用到系统hosts文件失败: ' + (error.message || error), 'error');
  } finally {
    applyingSourceId.value = null;
  }
}
</script>
