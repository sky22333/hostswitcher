<template>
  <div class="hosts-editor">
    <v-container fluid class="pa-0">
      <!-- 顶部工具栏 -->
      <v-card class="editor-toolbar" elevation="0" rounded="lg">
        <v-card-text class="py-3">
          <v-row align="center" no-gutters>

            <v-spacer />
            
            <v-col cols="auto">
              <div class="d-flex align-center gap-2">
                <!-- 权限提示 -->
                <v-chip
                  v-if="needsAdmin"
                  size="small"
                  color="warning"
                  variant="tonal"
                  prepend-icon="mdi-shield-account"
                >
                  需要管理员权限
                </v-chip>
                
                <!-- 系统配置信息 -->
                <v-tooltip location="bottom">
                  <template #activator="{ props }">
                    <v-btn
                      v-bind="props"
                      size="small"
                      variant="tonal"
                      prepend-icon="mdi-cog-outline"
                      @click="openSystemHostsFile"
                      :disabled="loading || saving || restoring"
                    >
                      系统host配置
                    </v-btn>
                  </template>
                  <span>{{ systemHostsPath }}</span>
                </v-tooltip>
                
                <!-- 操作按钮 -->
                <v-btn-group variant="outlined" density="compact">
                  <v-btn
                    @click="showRestoreDialog = true"
                    :disabled="loading || saving || restoring"
                    title="恢复默认hosts文件"
                    color="warning"
                  >
                    <v-icon>mdi-restore</v-icon>
                  </v-btn>
                  
                  <v-btn
                    @click="refreshContent"
                    :loading="loading"
                    :disabled="loading || saving || restoring"
                    title="刷新内容"
                  >
                    <v-icon>mdi-refresh</v-icon>
                  </v-btn>
                  
                  <v-btn
                    @click="validateContent"
                    :disabled="!editorContent || loading || saving || restoring"
                    title="验证内容"
                  >
                    <v-icon>mdi-check-circle-outline</v-icon>
                  </v-btn>
                  
                  <v-btn
                    @click="saveContent"
                    :loading="saving"
                    :disabled="!hasChanges || saving || restoring"
                    color="primary"
                    title="保存更改"
                  >
                    <v-icon>mdi-content-save</v-icon>
                  </v-btn>
                </v-btn-group>
              </div>
            </v-col>
          </v-row>
        </v-card-text>
      </v-card>
      
      <!-- 编辑器主体 -->
      <v-card class="editor-main mt-4" elevation="0" rounded="lg">
        <v-card-text class="pa-0">
          <!-- 编辑器状态栏 -->
          <div class="editor-statusbar">
            <div class="d-flex align-center justify-space-between">
              <div class="d-flex align-center gap-4">
                <v-chip
                  size="small"
                  :color="hasChanges ? 'warning' : 'success'"
                  variant="tonal"
                >
                  {{ hasChanges ? '未保存' : '已保存' }}
                </v-chip>
                
                <span class="text-caption text-medium-emphasis">
                  {{ lineCount }} 行 | {{ editorContent.length }} 字符
                </span>
              </div>
            </div>
          </div>
          
          <!-- Monaco编辑器 -->
          <div class="editor-container">
            <MonacoEditor
              v-model="editorContent"
              language="hosts"
              :options="editorOptions"
              @change="onEditorChange"
              height="calc(100vh - 280px)"
            />
          </div>
        </v-card-text>
      </v-card>
    </v-container>
    
    <!-- 未保存更改提示对话框 -->
    <v-dialog v-model="showUnsavedDialog" max-width="400px" persistent>
      <v-card class="rounded-lg">
        <v-card-title class="d-flex align-center">
          <v-icon color="warning" class="mr-2">mdi-alert</v-icon>
          未保存的更改
        </v-card-title>
        <v-card-text>
          您有未保存的更改，是否保存后再切换？
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn
            variant="text"
            @click="handleUnsavedChanges(false)"
          >
            不保存
          </v-btn>
          <v-btn
            color="primary"
            @click="handleUnsavedChanges(true)"
            :loading="saving"
            :disabled="saving"
          >
            保存
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    
    <!-- 权限警告对话框 -->
    <v-dialog v-model="showPermissionDialog" max-width="500px" persistent>
      <v-card class="rounded-lg">
        <v-card-title class="d-flex align-center">
          <v-icon color="error" class="mr-2">mdi-shield-alert</v-icon>
          权限不足
        </v-card-title>
        <v-card-text>
          <p>修改系统 hosts 文件需要管理员权限。</p>
          <p class="mt-2">请以管理员身份重新启动此应用程序。</p>
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn
            color="primary"
            @click="showPermissionDialog = false"
          >
            确定
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    
    <!-- 恢复默认hosts确认对话框 -->
    <v-dialog v-model="showRestoreDialog" max-width="600px" persistent>
      <v-card class="rounded-lg">
        <v-card-title class="d-flex align-center">
          <v-icon color="warning" class="mr-2">mdi-alert-circle</v-icon>
          确认恢复默认hosts文件
        </v-card-title>
        <v-card-text>
          <v-alert
            type="warning"
            variant="tonal"
            class="mb-4"
          >
            <div class="font-weight-bold mb-2">⚠️ 重要警告</div>
            <div>此操作将完全覆盖您当前的hosts文件配置！</div>
          </v-alert>
          
          <div class="mb-3">
            <strong>此操作将会：</strong>
            <ul class="mt-2">
              <li>清除所有自定义的hosts映射</li>
              <li>删除所有远程源应用的配置</li>
              <li>恢复为Windows系统默认的hosts文件</li>
              <li>可能影响您当前的网络访问配置</li>
            </ul>
          </div>
          
          <div class="mb-3">
            <strong>恢复后的默认内容将包含：</strong>
            <ul class="mt-2">
              <li>127.0.0.1 → localhost</li>
              <li>::1 → localhost (IPv6)</li>
              <li>基本的系统注释信息</li>
            </ul>
          </div>
          
          <v-alert
            type="info"
            variant="tonal"
            class="mt-4"
          >
            <strong>建议：</strong>如果您有重要的hosts配置，请先手动备份当前内容。
          </v-alert>
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn
            color="default"
            variant="text"
            @click="showRestoreDialog = false"
            :disabled="restoring"
          >
            取消
          </v-btn>
          <v-btn
            color="warning"
            @click="confirmRestoreDefault"
            :loading="restoring"
            :disabled="restoring"
          >
            确认恢复默认
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onBeforeUnmount, watch, nextTick } from 'vue';
import { useConfigStore } from '@/stores/config';
import { useNotificationStore } from '@/stores/notification';
import MonacoEditor from '@/components/MonacoEditor.vue';

// 状态管理
const configStore = useConfigStore();
const notificationStore = useNotificationStore();

// 编辑器状态
const editorContent = ref('');
const originalContent = ref('');
const selectedConfigId = ref('system');
const loading = ref(false);
const saving = ref(false);
const showUnsavedDialog = ref(false);
const showPermissionDialog = ref(false);
const showRestoreDialog = ref(false);
const pendingAction = ref(null);
const needsAdmin = ref(false);
const restoring = ref(false);

// 编辑器配置
const editorOptions = {
  fontSize: 14,
  wordWrap: 'on',
  automaticLayout: true,
  minimap: { enabled: true },
  scrollBeyondLastLine: false,
  lineNumbers: 'on',
  renderLineHighlight: 'all',
  tabSize: 4,
  insertSpaces: true,
  detectIndentation: false,
  folding: true,
  foldingStrategy: 'indentation',
  showFoldingControls: 'always',
  rulers: [80],
  renderWhitespace: 'selection',
};

// 计算属性
const configItems = computed(() => {
  return [
    { 
      title: '系统 hosts 文件', 
      value: 'system',
      icon: 'mdi-file-cog'
    }
  ];
});

const selectedConfig = computed(() => {
  if (selectedConfigId.value === 'system') {
    return null;
  }
  return configStore.configs.find(config => config.ID === selectedConfigId.value);
});

const hasChanges = computed(() => {
  return editorContent.value !== originalContent.value;
});

const lineCount = computed(() => {
  return editorContent.value.split('\n').length;
});

const currentFilePath = computed(() => {
  if (selectedConfigId.value === 'system') {
    return 'hosts';
  }
  return selectedConfig.value?.Name || '';
});

const systemHostsPath = computed(() => {
  return configStore.systemHostsPath;
});

// 方法
const loadConfigs = async () => {
  try {
    await configStore.loadConfigs();
  } catch (error) {
    notificationStore.showNotification('加载配置失败: ' + error.message, 'error');
  }
};

const loadSelectedConfig = async () => {
  try {
    console.log('=== 开始加载配置 ===');
    console.log('当前选择的配置ID:', selectedConfigId.value);
    
    if (selectedConfigId.value === 'system') {
      console.log('正在直接读取系统hosts文件...');
      
      const content = await configStore.readSystemHosts();
      console.log('读取到的系统hosts内容长度:', content.length);
      
      // 更新编辑器内容
      editorContent.value = '';
      await nextTick();
      editorContent.value = content;
      originalContent.value = content;
      
      // 处理空文件情况
      if (content.trim()) {
        console.log('✅ 成功加载系统hosts文件');
      } else {
        notificationStore.showNotification('系统 hosts 文件为空，已创建默认内容', 'info');
        const defaultContent = "# Copyright (c) 1993-2009 Microsoft Corp.\n#\n# This is a sample HOSTS file used by Microsoft TCP/IP for Windows.\n#\n127.0.0.1\tlocalhost\n::1\t\tlocalhost\n";
        editorContent.value = '';
        await nextTick();
        editorContent.value = defaultContent;
        originalContent.value = defaultContent;
        console.log('⚠️ 系统hosts文件为空，使用默认内容');
      }
    } else {
      // 加载选中的配置
      const config = configStore.configs.find(c => c.ID === selectedConfigId.value);
      if (config) {
        console.log('正在加载配置:', config.Name);
        editorContent.value = '';
        await nextTick();
        editorContent.value = config.Content;
        originalContent.value = config.Content;
      } else {
        console.error('找不到配置:', selectedConfigId.value);
        // 如果配置不存在，强制回退到系统hosts
        selectedConfigId.value = 'system';
        await loadSelectedConfig();
        return;
      }
    }
    
    console.log('=== 加载配置完成 ===');
  } catch (error) {
    console.error('❌ 加载内容失败:', error);
    notificationStore.showNotification('加载内容失败: ' + error.message, 'error');
    
  }
};

const handleConfigChange = async () => {
  if (hasChanges.value) {
    pendingAction.value = loadSelectedConfig;
    showUnsavedDialog.value = true;
  } else {
    await loadSelectedConfig();
  }
};

const saveContent = async () => {
  if (!hasChanges.value) {
    console.log('⚠️ 没有变化，跳过保存');
    return;
  }
  
  console.log('=== 开始保存内容 ===');
  console.log('当前选择配置:', selectedConfigId.value);
  console.log('editorContent长度:', editorContent.value.length);
  console.log('needsAdmin:', needsAdmin.value);
  
  saving.value = true;
  try {
    if (selectedConfigId.value === 'system') {
      console.log('保存到系统hosts文件...');
      
      // 检查管理员权限
      if (needsAdmin.value) {
        console.log('❌ 需要管理员权限，显示权限对话框');
        showPermissionDialog.value = true;
        return;
      }
      
      console.log('✅ 有管理员权限，开始保存');
      
      // 验证内容
      console.log('验证hosts内容...');
      await configStore.validateHostsContent(editorContent.value);
      console.log('✅ 内容验证通过');
      
      // 保存到系统hosts文件
      console.log('写入系统hosts文件...');
      await configStore.writeSystemHosts(editorContent.value);
      console.log('✅ 写入成功');
      
      originalContent.value = editorContent.value;
      notificationStore.showNotification('系统 hosts 文件已保存', 'success');
      console.log('✅ 保存完成');
    } else {
      // 保存到配置
      const config = selectedConfig.value;
      if (config) {
        console.log('保存到配置:', config.Name);
        await configStore.updateConfig(
          config.ID,
          config.Name,
          config.Description,
          editorContent.value
        );
        originalContent.value = editorContent.value;
        notificationStore.showNotification('配置已保存', 'success');
        console.log('✅ 配置保存完成');
      } else {
        console.error('❌ 找不到要保存的配置');
      }
    }
  } catch (error) {
    console.error('❌ 保存失败:', error);
    console.error('错误详情:', error.message);
    notificationStore.showNotification('保存失败: ' + error.message, 'error');
  } finally {
    saving.value = false;
    console.log('=== 保存状态重置 ===');
  }
};

const refreshContent = async () => {
  if (hasChanges.value) {
    const confirmed = confirm('您有未保存的更改，确定要刷新吗？');
    if (!confirmed) return;
  }
  
  await loadSelectedConfig();
  notificationStore.showNotification('内容已刷新', 'info');
};

const validateContent = async () => {
  try {
    await configStore.validateHostsContent(editorContent.value);
    notificationStore.showNotification('内容格式正确', 'success');
  } catch (error) {
    notificationStore.showNotification('验证失败: ' + error.message, 'error');
  }
};

const handleUnsavedChanges = async (save) => {
  if (save) {
    await saveContent();
  }
  
  showUnsavedDialog.value = false;
  
  if (pendingAction.value) {
    await pendingAction.value();
    pendingAction.value = null;
  }
};

const onEditorChange = () => {
  // 编辑器内容变化时的处理
};

const checkAdminPermission = async () => {
  try {
    console.log('检查管理员权限...');
    needsAdmin.value = await configStore.isAdminRequired();
    console.log('IsAdminRequired返回:', needsAdmin.value);
    console.log('needsAdmin.value设置为:', needsAdmin.value);
    
    if (needsAdmin.value) {
      console.log('⚠️ 需要管理员权限');
    } else {
      console.log('✅ 有足够权限');
    }
  } catch (error) {
    console.error('检查权限失败:', error);
    needsAdmin.value = true;
  }
};

const restoreDefault = async () => {
  console.log('=== 开始恢复默认hosts ===');
  restoring.value = true;
  try {
    console.log('调用后端恢复默认hosts...');
    // 恢复默认hosts文件
    await configStore.restoreDefaultHosts();
    console.log('✅ 后端恢复成功');
    
    // 等待一下确保文件系统同步
    await new Promise(resolve => setTimeout(resolve, 200));
    
    // 强制重新加载内容，确保显示最新状态
    console.log('强制重新加载内容...');
    await loadSelectedConfig();
    console.log('✅ 内容重新加载完成');
    
    // 再次检查内容是否正确更新
    const newContent = await configStore.readSystemHosts();
    console.log('重新读取的内容长度:', newContent.length);
    console.log('重新读取的内容开头:', newContent.substring(0, 100));
    
    // 确保编辑器内容也正确更新
    if (editorContent.value !== newContent) {
      console.log('编辑器内容不一致，再次更新...');
      editorContent.value = '';
      await nextTick();
      editorContent.value = newContent;
      originalContent.value = newContent;
      await nextTick();
    }
    
    notificationStore.showNotification('系统 hosts 文件已恢复为默认', 'success');
    console.log('✅ 恢复默认完成');
  } catch (error) {
    console.error('❌ 恢复默认失败:', error);
    console.error('错误详情:', error.message);
    console.error('错误堆栈:', error.stack);
    notificationStore.showNotification('恢复默认失败: ' + error.message, 'error');
  } finally {
    restoring.value = false;
    console.log('=== 恢复状态重置 ===');
  }
};

const openSystemHostsFile = async () => {
  try {
    console.log('正在打开系统hosts文件...');
    await window.go.services.TrayService.OpenSystemHostsFile();
    notificationStore.showNotification('已打开系统Hosts文件', 'success');
  } catch (error) {
    console.error('打开系统hosts文件失败:', error);
    notificationStore.showNotification('打开系统Hosts文件失败: ' + (error.message || error), 'error');
  }
};

const confirmRestoreDefault = async () => {
  showRestoreDialog.value = false;
  await restoreDefault();
};

// 生命周期
onMounted(async () => {
  try {
    // 首先清理所有配置数据，确保干净状态
    console.log('正在清理配置数据...');
    
    // 初始化配置store
    await configStore.initialize();
    
    // 检查管理员权限
    await checkAdminPermission();
    
    // 强制选择系统hosts文件
    selectedConfigId.value = 'system';
    
    // 延迟确保所有初始化完成
    await new Promise(resolve => setTimeout(resolve, 200));
    
    // 强制加载系统hosts文件
    console.log('强制加载系统hosts文件...');
    await loadSelectedConfig();
    
    // 监听配置列表变化事件
    if (window.runtime && window.runtime.EventsOn) {
      window.runtime.EventsOn('config-list-changed', () => {
        console.log('配置列表已变化，重新加载...');
        loadConfigs();
      });
      window.runtime.EventsOn('config-applied', () => {
        console.log('配置已应用，重新加载...');
        loadConfigs();
      });
      window.runtime.EventsOn('system-hosts-updated', () => {
        console.log('系统hosts已更新，重新加载...');
        if (selectedConfigId.value === 'system') {
          loadSelectedConfig();
        }
      });
    }
  } catch (error) {
    console.error('初始化失败:', error);
    notificationStore.showNotification('初始化失败: ' + error.message, 'error');
  }
});

onBeforeUnmount(() => {
  // 移除事件监听
  if (window.runtime && window.runtime.EventsOff) {
    window.runtime.EventsOff('config-list-changed');
    window.runtime.EventsOff('config-applied');
    window.runtime.EventsOff('system-hosts-updated');
  }
});

// 监听配置选择变化
watch(selectedConfigId, (newVal, oldVal) => {
  if (newVal !== oldVal && !loading.value) {
    handleConfigChange();
  }
});
</script>

<style scoped>
.hosts-editor {
  height: 100%;
}

.editor-toolbar,
.editor-main {
  border: 1px solid rgba(var(--v-theme-on-surface), 0.12);
  background: rgba(var(--v-theme-surface), 0.95);
}

.editor-statusbar {
  padding: 12px 16px;
  border-bottom: 1px solid rgba(var(--v-theme-on-surface), 0.08);
  background: rgba(var(--v-theme-surface-variant), 0.3);
}

.editor-container {
  overflow: visible;
  position: relative;
}

.gap-2 {
  gap: 8px;
}

.gap-4 {
  gap: 16px;
}

/* 深色主题适配 */
.v-theme--darkTheme .editor-toolbar,
.v-theme--darkTheme .editor-main {
  border-color: rgba(255, 255, 255, 0.12);
}

.v-theme--darkTheme .editor-statusbar {
  border-bottom-color: rgba(255, 255, 255, 0.08);
}
</style>
