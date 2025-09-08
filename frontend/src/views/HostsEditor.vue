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
                
                <!-- 操作按钮组 -->
                <div class="d-flex align-center gap-1">
                  <ToolbarButton
                    icon="mdi-file-outline"
                    :tooltip="`系统hosts配置文件：${systemHostsPath}`"
                    :disabled="loading || saving || restoring || flushing"
                    @click="openSystemHostsFile"
                  />
                  
                  <ToolbarButton
                    icon="mdi-restore"
                    tooltip="恢复默认hosts文件，清除所有自定义配置。"
                    :disabled="loading || saving || restoring || flushing"
                    button-class="icon-btn warning-btn"
                    @click="showRestoreDialog = true"
                  />
                  
                  <ToolbarButton
                    icon="mdi-refresh"
                    tooltip="刷新内容，重新加载当前hosts文件。"
                    :loading="loading"
                    :disabled="loading || saving || restoring || flushing"
                    @click="refreshContent"
                  />
                  
                  <ToolbarButton
                    icon="mdi-dns"
                    tooltip="刷新DNS缓存，清除系统域名解析缓存。"
                    :loading="flushing"
                    :disabled="loading || saving || restoring || flushing"
                    button-class="icon-btn info-btn"
                    @click="flushDNSCache"
                  />
                  
                  <ToolbarButton
                    icon="mdi-check-circle-outline"
                    tooltip="验证内容格式，检查hosts文件语法。"
                    :disabled="!editorContent || loading || saving || restoring || flushing"
                    @click="validateContent"
                  />
                  

                  <ToolbarButton
                    icon="mdi-content-save"
                    tooltip="保存更改，将内容写入hosts文件。"
                    :loading="saving"
                    :disabled="!hasChanges || saving || restoring || flushing"
                    button-class="icon-btn primary-btn"
                    @click="saveContent"
                  />
                </div>
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
              v-memo="[editorContent, editorOptions]"
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
      <v-card class="rounded-xl modern-dialog">
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

    <!-- 刷新确认对话框 -->
    <v-dialog v-model="showRefreshConfirmDialog" max-width="420px" persistent>
      <v-card class="rounded-xl modern-dialog">
        <v-card-text class="pa-6 text-center">
          <div class="mb-4">
            <v-icon size="48" color="warning" class="mb-2">mdi-refresh-circle</v-icon>
          </div>
          <h3 class="text-h6 font-weight-bold mb-3">确认刷新内容</h3>
          <p class="text-body-1 text-medium-emphasis mb-0">
            您有未保存的更改，刷新将丢失这些修改。确定要继续吗？
          </p>
        </v-card-text>
        <v-card-actions class="pa-6 pt-0 d-flex gap-3">
          <v-btn
            variant="text"
            size="large"
            class="flex-1"
            @click="showRefreshConfirmDialog = false"
          >
            取消
          </v-btn>
          <v-btn
            color="warning"
            size="large"
            class="flex-1"
            @click="confirmRefresh"
          >
            确认刷新
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, watch, nextTick, shallowRef } from 'vue';
import { useConfigStore } from '@/stores/config';
import { useNotificationStore } from '@/stores/notification';
import MonacoEditor from '@/components/MonacoEditor.vue';
import { useEventManager } from '@/utils/eventManager';
import ToolbarButton from '@/components/ToolbarButton.vue';

const configStore = useConfigStore();
const notificationStore = useNotificationStore();
const { addWailsListener } = useEventManager();

const editorContent = ref('');
const originalContent = ref('');
const selectedConfigId = ref('system');
const loading = ref(false);
const saving = ref(false);
const flushing = ref(false);
const showUnsavedDialog = ref(false);
const showPermissionDialog = ref(false);
const showRestoreDialog = ref(false);
const showRefreshConfirmDialog = ref(false);
const pendingAction = ref(null);
const needsAdmin = ref(false);
const restoring = ref(false);
const editorOptions = shallowRef({
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
});


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


const loadConfigs = async () => {
  try {
    await configStore.loadConfigs();
  } catch (error) {
    notificationStore.showNotification('加载配置失败: ' + error.message, 'error');
  }
};

const loadSelectedConfig = async () => {
  try {
    if (selectedConfigId.value === 'system') {
      const content = await configStore.readSystemHosts();
      
      // 更新编辑器内容
      editorContent.value = '';
      await nextTick();
      editorContent.value = content;
      originalContent.value = content;
      
      // 处理空文件情况
      if (!content.trim()) {
        notificationStore.showNotification('系统 hosts 文件为空，已创建默认内容', 'info');
        const defaultContent = "# Copyright (c) 1993-2009 Microsoft Corp.\n#\n# This is a sample HOSTS file used by Microsoft TCP/IP for Windows.\n#\n127.0.0.1\tlocalhost\n::1\t\tlocalhost\n";
        editorContent.value = '';
        await nextTick();
        editorContent.value = defaultContent;
        originalContent.value = defaultContent;
      }
    } else {
      // 加载选中的配置
      const config = configStore.configs.find(c => c.ID === selectedConfigId.value);
      if (config) {
        editorContent.value = '';
        await nextTick();
        editorContent.value = config.Content;
        originalContent.value = config.Content;
      } else {

        selectedConfigId.value = 'system';
        await loadSelectedConfig();
        return;
      }
    }
  } catch (error) {

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
    return;
  }
  
  saving.value = true;
  try {
    if (selectedConfigId.value === 'system') {
      // 检查管理员权限
      if (needsAdmin.value) {
        showPermissionDialog.value = true;
        return;
      }
      
      // 验证内容
      await configStore.validateHostsContent(editorContent.value);
      
      await configStore.writeSystemHosts(editorContent.value);
      
      originalContent.value = editorContent.value;
      notificationStore.showNotification('系统 hosts 文件已保存', 'success');
    } else {
      // 保存到配置
      const config = selectedConfig.value;
      if (config) {
        await configStore.updateConfig(
          config.ID,
          config.Name,
          config.Description,
          editorContent.value
        );
        originalContent.value = editorContent.value;
        notificationStore.showNotification('配置已保存', 'success');
      } else {

      }
    }
  } catch (error) {

    notificationStore.showNotification('保存失败: ' + error.message, 'error');
  } finally {
    saving.value = false;
  }
};

const refreshContent = async () => {
  if (hasChanges.value) {
    showRefreshConfirmDialog.value = true;
    return;
  }
  
  await loadSelectedConfig();
  notificationStore.showNotification('内容已刷新', 'info');
};

const confirmRefresh = async () => {
  showRefreshConfirmDialog.value = false;
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
    needsAdmin.value = await configStore.isAdminRequired();
  } catch (error) {

    needsAdmin.value = true;
  }
};

const restoreDefault = async () => {
  restoring.value = true;
  try {
    await configStore.restoreDefaultHosts();
    
    await new Promise(resolve => setTimeout(resolve, 200));
    
    await loadSelectedConfig();
    
    const newContent = await configStore.readSystemHosts();
    
    if (editorContent.value !== newContent) {
      editorContent.value = '';
      await nextTick();
      editorContent.value = newContent;
      originalContent.value = newContent;
      await nextTick();
    }
    
    notificationStore.showNotification('系统 hosts 文件已恢复为默认', 'success');
  } catch (error) {

    notificationStore.showNotification('恢复默认失败: ' + error.message, 'error');
  } finally {
    restoring.value = false;
  }
};

const openSystemHostsFile = async () => {
  try {
    await window.go.services.TrayService.OpenSystemHostsFile();
    notificationStore.showNotification('已打开系统Hosts文件', 'success');
  } catch (error) {

    notificationStore.showNotification('打开系统Hosts文件失败: ' + (error.message || error), 'error');
  }
};

const confirmRestoreDefault = async () => {
  showRestoreDialog.value = false;
  await restoreDefault();
};



const flushDNSCache = async () => {
  flushing.value = true;
  try {
    await configStore.flushDNSCache();
    notificationStore.showNotification('DNS缓存已刷新', 'success');
  } catch (error) {

    notificationStore.showNotification('DNS缓存刷新失败: ' + error.message, 'error');
  } finally {
    flushing.value = false;
  }
};


onMounted(async () => {
  try {
    await configStore.initialize();
    await checkAdminPermission();
    selectedConfigId.value = 'system';
    await new Promise(resolve => setTimeout(resolve, 200));
    await loadSelectedConfig();
    
    addWailsListener('config-list-changed', () => {
      loadConfigs();
    });
    addWailsListener('config-applied', () => {
      loadConfigs();
    });
    addWailsListener('system-hosts-updated', () => {
      if (selectedConfigId.value === 'system') {
        loadSelectedConfig();
      }
    });
  } catch (error) {

    notificationStore.showNotification('初始化失败: ' + error.message, 'error');
  }
});




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
  border: none;
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

/* 图标按钮样式 - 使用全局样式，仅保留特定样式 */
.icon-btn {
  opacity: 0.7;
}

.icon-btn:hover {
  opacity: 1;
  background: rgba(var(--v-theme-on-surface), 0.08) !important;
}

/* 深色主题适配 */
.v-theme--darkTheme .editor-toolbar,
.v-theme--darkTheme .editor-main {
  border-color: rgba(255, 255, 255, 0.12);
}

.v-theme--darkTheme .editor-statusbar {
  border-bottom-color: rgba(255, 255, 255, 0.08);
}

/* 现代化对话框样式 */
.modern-dialog {
  border: none !important;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.08), 
              0 10px 30px rgba(0, 0, 0, 0.04) !important;
  backdrop-filter: blur(24px) !important;
  -webkit-backdrop-filter: blur(24px) !important;
}

.modern-dialog .v-card-text {
  border-radius: 24px 24px 0 0 !important;
}

.modern-dialog .v-card-actions {
  border-radius: 0 0 24px 24px !important;
  background: rgba(var(--v-theme-surface), 0.6) !important;
}

.modern-dialog .v-btn {
  border-radius: 12px !important;
  font-weight: 600 !important;
  text-transform: none !important;
  letter-spacing: 0.02em !important;
}

/* 暗色主题对话框 */
.v-theme--darkTheme .modern-dialog {
  box-shadow: 0 24px 72px rgba(0, 0, 0, 0.3), 
              0 12px 36px rgba(0, 0, 0, 0.15) !important;
}

.v-theme--darkTheme .modern-dialog .v-card-actions {
  background: rgba(var(--v-theme-surface), 0.4) !important;
}
</style>
