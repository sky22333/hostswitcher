<template>
  <div class="backup-manager">
    <v-container fluid class="pa-0">
      <!-- 顶部统计卡片 -->
      <v-row class="mb-4">
        <v-col cols="12" md="3">
          <v-card 
            elevation="0" 
            class="stats-card clickable-card" 
            color="primary" 
            variant="tonal"
            :class="{ 'active-filter': currentFilter === 'all' }"
            @click="setFilter('all')"
          >
            <v-card-text class="pa-4">
              <div class="d-flex align-center">
                <v-icon size="32" class="me-3">mdi-database</v-icon>
                <div>
                  <div class="text-h5 font-weight-bold">{{ backupStore.stats.total }}</div>
                  <div class="text-body-2">总备份</div>
                </div>
              </div>
            </v-card-text>
          </v-card>
        </v-col>
        <v-col cols="12" md="3">
          <v-card 
            elevation="0" 
            class="stats-card clickable-card" 
            color="success" 
            variant="tonal"
            :class="{ 'active-filter': currentFilter === 'automatic' }"
            @click="setFilter('automatic')"
          >
            <v-card-text class="pa-4">
              <div class="d-flex align-center">
                <v-icon size="32" class="me-3">mdi-robot</v-icon>
                <div>
                  <div class="text-h5 font-weight-bold">{{ backupStore.stats.automatic }}</div>
                  <div class="text-body-2">自动备份</div>
                </div>
              </div>
            </v-card-text>
          </v-card>
        </v-col>
        <v-col cols="12" md="3">
          <v-card 
            elevation="0" 
            class="stats-card clickable-card" 
            color="info" 
            variant="tonal"
            :class="{ 'active-filter': currentFilter === 'manual' }"
            @click="setFilter('manual')"
          >
            <v-card-text class="pa-4">
              <div class="d-flex align-center">
                <v-icon size="32" class="me-3">mdi-account</v-icon>
                <div>
                  <div class="text-h5 font-weight-bold">{{ backupStore.stats.manual }}</div>
                  <div class="text-body-2">手动备份</div>
                </div>
              </div>
            </v-card-text>
          </v-card>
        </v-col>
        <v-col cols="12" md="3">
          <v-card elevation="0" class="stats-card" color="warning" variant="tonal">
            <v-card-text class="pa-4">
              <div class="d-flex align-center">
                <v-icon size="32" class="me-3">mdi-harddisk</v-icon>
                <div>
                  <div class="text-h5 font-weight-bold">{{ backupStore.formatFileSize(backupStore.stats.totalSize) }}</div>
                  <div class="text-body-2">总大小</div>
                </div>
              </div>
            </v-card-text>
          </v-card>
        </v-col>
      </v-row>

      <!-- 操作工具栏 -->
      <v-card elevation="0" class="mb-4" rounded="lg">
        <v-card-text class="py-3">
          <v-row align="center" no-gutters>
            <v-col>
              <h2 class="text-h6 font-weight-bold">备份管理</h2>
            </v-col>
            <v-spacer />
            <v-col cols="auto">
              <div class="d-flex align-center gap-2">
                <v-btn
                  color="primary"
                  @click="showCreateBackupDialog = true"
                  :loading="backupStore.loading"
                  :disabled="backupStore.loading"
                  prepend-icon="mdi-plus"
                  variant="flat"
                  class="rounded-lg"
                >
                  创建备份
                </v-btn>
                <v-btn
                  color="error"
                  @click="confirmClearAutoBackups"
                  :loading="backupStore.loading"
                  :disabled="backupStore.loading || backupStore.stats.automatic === 0"
                  prepend-icon="mdi-delete-sweep"
                  variant="outlined"
                  class="rounded-lg"
                >
                  清理自动备份
                </v-btn>
                <v-btn
                  icon
                  variant="text"
                  @click="refreshBackupsSilently"
                  :loading="backupStore.loading"
                  :disabled="backupStore.loading"
                  class="icon-btn"
                >
                  <v-icon>mdi-refresh</v-icon>
                </v-btn>
              </div>
            </v-col>
          </v-row>
        </v-card-text>
      </v-card>

      <!-- 备份列表 -->
      <v-card elevation="0" rounded="lg" class="backup-list-container">
        <v-card-text class="pa-0">
          <div v-if="backupStore.loading && backupStore.sortedBackups.length === 0" class="text-center pa-8 empty-state">
            <v-progress-circular indeterminate color="primary" size="48" class="mb-4"></v-progress-circular>
            <div class="text-body-1 text-medium-emphasis">加载备份中...</div>
          </div>

          <div v-else-if="filteredBackups.length === 0" class="text-center pa-8 empty-state">
            <v-icon size="64" color="medium-emphasis" class="mb-4">mdi-database-off</v-icon>
            <div class="text-h6 mb-2">{{ getEmptyMessage() }}</div>
            <div class="text-body-2 text-medium-emphasis mb-4">{{ getEmptySubMessage() }}</div>
            <v-btn
              v-if="currentFilter === 'all'"
              color="primary"
              @click="showCreateBackupDialog = true"
              prepend-icon="mdi-plus"
              variant="flat"
            >
              创建备份
            </v-btn>
            <v-btn
              v-else
              color="primary"
              @click="setFilter('all')"
              prepend-icon="mdi-filter-off"
              variant="outlined"
            >
              查看全部
            </v-btn>
          </div>

          <!-- 备份时间线 - 添加滚动容器 -->
          <div v-else class="backup-timeline-container">
            <v-timeline side="end" class="backup-timeline">
              <v-timeline-item
                v-for="backup in filteredBackups"
                :key="backup.id"
                size="small"
                :dot-color="backup.isAutomatic ? 'success' : 'primary'"
                :icon="backup.isAutomatic ? 'mdi-robot' : 'mdi-account'"
                class="backup-item"
              >
                <template v-slot:opposite>
                  <div class="text-caption text-medium-emphasis">
                    {{ formatTimestamp(backup.timestamp) }}
                  </div>
                </template>

                <v-card class="backup-card compact" elevation="2" rounded="lg">
                  <v-card-text class="pa-3">
                    <!-- 备份头部信息 -->
                    <div class="d-flex align-center mb-2">
                      <div class="flex-1">
                        <div class="d-flex align-center mb-1">
                          <span class="text-subtitle-2 font-weight-medium me-2">{{ backup.description }}</span>
                          <v-chip
                            :color="backup.isAutomatic ? 'success' : 'primary'"
                            size="x-small"
                            variant="flat"
                          >
                            {{ backup.isAutomatic ? '自动' : '手动' }}
                          </v-chip>
                        </div>
                        <div class="text-caption text-medium-emphasis">
                          {{ backupStore.formatRelativeTime(backup.timestamp) }} • {{ backupStore.formatFileSize(backup.size) }}
                        </div>
                      </div>
                      <div class="d-flex align-center gap-1">
                        <v-btn
                          icon
                          size="small"
                          variant="text"
                          @click="showPreviewDialog(backup)"
                          title="预览内容"
                        >
                          <v-icon size="16">mdi-eye</v-icon>
                        </v-btn>
                        <v-btn
                          icon
                          size="small"
                          variant="text"
                          @click="showEditDialog(backup)"
                          :disabled="backup.isAutomatic"
                          title="编辑标签"
                        >
                          <v-icon size="16">mdi-tag</v-icon>
                        </v-btn>
                        <v-btn
                          icon
                          size="small"
                          variant="text"
                          color="error"
                          @click="confirmDeleteBackup(backup)"
                          :disabled="backup.isAutomatic"
                          title="删除备份"
                        >
                          <v-icon size="16">mdi-delete</v-icon>
                        </v-btn>
                        <v-btn
                          icon
                          size="small"
                          variant="text"
                          color="primary"
                          @click="confirmRestoreBackup(backup)"
                          title="恢复此备份"
                        >
                          <v-icon size="16">mdi-restore</v-icon>
                        </v-btn>
                      </div>
                    </div>

                    <!-- 标签 -->
                    <div v-if="backup.tags && backup.tags.length > 0" class="mb-1">
                      <v-chip
                        v-for="tag in backup.tags"
                        :key="tag"
                        size="x-small"
                        variant="outlined"
                        class="me-1 mb-1"
                      >
                        {{ tag }}
                      </v-chip>
                    </div>

                    <!-- 内容预览 - 添加点击事件 -->
                    <div class="backup-preview clickable compact" @click="showPreviewDialog(backup)">
                      <div class="text-caption mb-1 text-medium-emphasis d-flex align-center">
                        <span>内容预览:</span>
                        <v-icon size="12" class="ms-1">mdi-eye</v-icon>
                      </div>
                      <pre class="preview-content compact">{{ backupStore.getBackupPreview(backup.content, 2) }}</pre>
                    </div>
                  </v-card-text>
                </v-card>
              </v-timeline-item>
            </v-timeline>
          </div>
        </v-card-text>
      </v-card>
    </v-container>

    <!-- 创建备份对话框 - 增加边框线 -->
    <v-dialog v-model="showCreateBackupDialog" max-width="600px" persistent>
      <v-card class="rounded-xl dialog-card">
        <v-card-title class="d-flex align-center">
          <v-icon class="me-2" color="primary">mdi-plus</v-icon>
          创建手动备份
        </v-card-title>
        <v-card-text>
          <v-text-field
            v-model="newBackupDescription"
            label="备份描述"
            placeholder="输入备份描述"
            :rules="[v => !!v || '描述不能为空']"
            class="mb-3"
          />
          
          <div class="mb-3">
            <div class="text-subtitle-2 mb-2">标签 (可选)</div>
            <v-combobox
              v-model="newBackupTags"
              label="添加标签"
              multiple
              chips
              small-chips
              deletable-chips
              hint="按回车添加标签"
              persistent-hint
            />
          </div>

          <div class="mb-3">
            <div class="text-subtitle-2 mb-2">备份内容</div>
            <v-radio-group v-model="backupContentType" inline>
              <v-radio label="备份当前系统hosts" value="system"></v-radio>
              <v-radio label="备份自定义内容" value="custom"></v-radio>
            </v-radio-group>
          </div>

          <v-textarea
            v-if="backupContentType === 'custom'"
            v-model="newBackupContent"
            label="备份自定义的hosts内容"
            placeholder="输入自定义hosts文件内容"
            rows="10"
            auto-grow
            class="mb-3"
            hint="输入您想要备份的hosts文件内容"
            persistent-hint
          />
        </v-card-text>
        <v-card-actions>
          <v-spacer />
          <v-btn
            variant="text"
            @click="closeCreateBackupDialog"
            :disabled="creating"
          >
            取消
          </v-btn>
          <v-btn
            color="primary"
            @click="createBackup"
            :loading="creating"
            :disabled="!newBackupDescription || creating"
          >
            创建备份
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>

    <!-- 预览对话框 - 增加边框线 -->
    <v-dialog v-model="showContentPreviewDialog" max-width="800px">
      <v-card class="rounded-xl dialog-card">
        <v-card-title class="d-flex align-center">
          <v-icon class="me-2" color="info">mdi-eye</v-icon>
          备份内容预览
        </v-card-title>
        <v-card-text>
          <div v-if="selectedBackup" class="mb-3">
            <div class="text-subtitle-2 mb-1">{{ selectedBackup.description }}</div>
            <div class="text-caption text-medium-emphasis mb-2">
              创建时间: {{ formatTimestamp(selectedBackup.timestamp) }}
            </div>
            <v-divider class="mb-3" />
            <pre class="preview-full-content">{{ selectedBackup.content }}</pre>
          </div>
        </v-card-text>
        <v-card-actions>
          <v-spacer />
          <v-btn variant="text" @click="showContentPreviewDialog = false">
            关闭
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>

    <!-- 编辑标签对话框 - 增加边框线 -->
    <v-dialog v-model="showEditTagsDialog" max-width="500px" persistent>
      <v-card class="rounded-xl dialog-card">
        <v-card-title class="d-flex align-center">
          <v-icon class="me-2" color="primary">mdi-tag</v-icon>
          编辑备份标签
        </v-card-title>
        <v-card-text>
          <div v-if="selectedBackup" class="mb-3">
            <div class="text-subtitle-2 mb-2">{{ selectedBackup.description }}</div>
            <v-combobox
              v-model="editingTags"
              label="标签"
              multiple
              chips
              small-chips
              deletable-chips
              hint="按回车添加标签"
              persistent-hint
            />
          </div>
        </v-card-text>
        <v-card-actions>
          <v-spacer />
          <v-btn
            variant="text"
            @click="closeEditTagsDialog"
            :disabled="updating"
          >
            取消
          </v-btn>
          <v-btn
            color="primary"
            @click="updateTags"
            :loading="updating"
            :disabled="updating"
          >
            保存
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>

    <!-- 确认恢复对话框 - 增加边框线 -->
    <v-dialog v-model="showRestoreConfirmDialog" max-width="500px" persistent>
      <v-card class="rounded-xl dialog-card">
        <v-card-title class="d-flex align-center">
          <v-icon class="me-2" color="warning">mdi-restore</v-icon>
          确认恢复备份
        </v-card-title>
        <v-card-text>
          <v-alert type="warning" variant="tonal" class="mb-4">
            <div class="font-weight-bold mb-2">⚠️ 注意</div>
            <div>恢复备份将覆盖当前的hosts文件内容，当前内容将自动备份。</div>
          </v-alert>
          
          <div v-if="selectedBackup">
            <strong>备份信息:</strong>
            <ul class="mt-2">
              <li>描述: {{ selectedBackup.description }}</li>
              <li>创建时间: {{ formatTimestamp(selectedBackup.timestamp) }}</li>
              <li>大小: {{ backupStore.formatFileSize(selectedBackup.size) }}</li>
            </ul>
          </div>
        </v-card-text>
        <v-card-actions>
          <v-spacer />
          <v-btn
            variant="text"
            @click="showRestoreConfirmDialog = false"
            :disabled="restoring"
          >
            取消
          </v-btn>
          <v-btn
            color="warning"
            @click="restoreBackup"
            :loading="restoring"
            :disabled="restoring"
          >
            确认恢复
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>

    <!-- 确认删除对话框 - 增加边框线 -->
    <v-dialog v-model="showDeleteConfirmDialog" max-width="400px" persistent>
      <v-card class="rounded-xl dialog-card">
        <v-card-title class="d-flex align-center">
          <v-icon class="me-2" color="error">mdi-delete</v-icon>
          确认删除备份
        </v-card-title>
        <v-card-text>
          <div v-if="selectedBackup">
            确定要删除备份 "{{ selectedBackup.description }}" 吗？
          </div>
          <v-alert type="error" variant="tonal" class="mt-3">
            此操作不可撤销！
          </v-alert>
        </v-card-text>
        <v-card-actions>
          <v-spacer />
          <v-btn
            variant="text"
            @click="showDeleteConfirmDialog = false"
            :disabled="deleting"
          >
            取消
          </v-btn>
          <v-btn
            color="error"
            @click="deleteBackup"
            :loading="deleting"
            :disabled="deleting"
          >
            删除
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>

    <!-- 确认清理自动备份对话框 -->
    <v-dialog v-model="showClearAutoBackupsDialog" max-width="500px" persistent>
      <v-card class="rounded-xl dialog-card">
        <v-card-title class="d-flex align-center">
          <v-icon class="me-2" color="error">mdi-delete-sweep</v-icon>
          确认清理自动备份
        </v-card-title>
        <v-card-text>
          <v-alert type="warning" variant="tonal" class="mb-4">
            <div class="font-weight-bold mb-2">⚠️ 警告</div>
            <div>此操作将删除所有自动备份（共 {{ backupStore.stats.automatic }} 个），不影响手动备份。</div>
          </v-alert>
          
          <div class="text-body-2 text-medium-emphasis">
            自动备份通常由系统在应用配置时自动创建，如果您确定不再需要这些备份，可以执行此操作来释放存储空间。
          </div>
        </v-card-text>
        <v-card-actions>
          <v-spacer />
          <v-btn
            variant="text"
            @click="showClearAutoBackupsDialog = false"
            :disabled="clearingAutoBackups"
          >
            取消
          </v-btn>
          <v-btn
            color="error"
            @click="clearAutoBackups"
            :loading="clearingAutoBackups"
            :disabled="clearingAutoBackups"
          >
            确认清理
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted, computed } from 'vue';
import { useBackupStore } from '@/stores/backup';
import { useNotificationStore } from '@/stores/notification';

// Store
const backupStore = useBackupStore();
const notificationStore = useNotificationStore();

// 响应式数据
const showCreateBackupDialog = ref(false);
const showContentPreviewDialog = ref(false);
const showEditTagsDialog = ref(false);
const showRestoreConfirmDialog = ref(false);
const showDeleteConfirmDialog = ref(false);
const showClearAutoBackupsDialog = ref(false);

const newBackupDescription = ref('');
const newBackupTags = ref([]);
const newBackupContent = ref('');
const backupContentType = ref('system'); // 'system' 或 'custom'
const editingTags = ref([]);
const selectedBackup = ref(null);
const currentFilter = ref('all'); // 'all', 'automatic', 'manual'

const creating = ref(false);
const updating = ref(false);
const restoring = ref(false);
const deleting = ref(false);
const clearingAutoBackups = ref(false);

// 计算属性
const formatTimestamp = computed(() => {
  return (timestamp) => {
    try {
      const date = new Date(timestamp);
      if (isNaN(date.getTime())) {
        return '无效时间';
      }
      return date.toLocaleString('zh-CN', {
        year: 'numeric',
        month: '2-digit',
        day: '2-digit',
        hour: '2-digit',
        minute: '2-digit',
        second: '2-digit'
      });
    } catch (error) {
      console.error('时间格式化错误:', error);
      return '时间格式错误';
    }
  };
});

// 过滤后的备份列表
const filteredBackups = computed(() => {
  if (currentFilter.value === 'all') {
    return backupStore.sortedBackups;
  } else if (currentFilter.value === 'automatic') {
    return backupStore.sortedBackups.filter(backup => backup.isAutomatic);
  } else if (currentFilter.value === 'manual') {
    return backupStore.sortedBackups.filter(backup => !backup.isAutomatic);
  }
  return backupStore.sortedBackups;
});

// 方法
const refreshBackupsSilently = async () => {
  try {
    await backupStore.loadBackups();
  } catch (error) {
    console.error('刷新备份列表失败:', error);
  }
};

const createBackup = async () => {
  creating.value = true;
  try {
    if (backupContentType.value === 'custom') {
      await backupStore.createBackupWithContent(newBackupDescription.value, newBackupContent.value, newBackupTags.value);
    } else {
      await backupStore.createBackup(newBackupDescription.value, newBackupTags.value);
    }
    notificationStore.showNotification('备份创建成功', 'success');
    closeCreateBackupDialog();
  } catch (error) {
    notificationStore.showNotification('创建备份失败: ' + error.message, 'error');
  } finally {
    creating.value = false;
  }
};

const closeCreateBackupDialog = () => {
  showCreateBackupDialog.value = false;
  newBackupDescription.value = '';
  newBackupTags.value = [];
  newBackupContent.value = '';
  backupContentType.value = 'system';
};

const showPreviewDialog = (backup) => {
  selectedBackup.value = backup;
  showContentPreviewDialog.value = true;
};

const showEditDialog = (backup) => {
  selectedBackup.value = backup;
  editingTags.value = [...(backup.tags || [])];
  showEditTagsDialog.value = true;
};

const closeEditTagsDialog = () => {
  showEditTagsDialog.value = false;
  selectedBackup.value = null;
  editingTags.value = [];
};

const updateTags = async () => {
  updating.value = true;
  try {
    await backupStore.updateBackupTags(selectedBackup.value.id, editingTags.value);
    notificationStore.showNotification('标签更新成功', 'success');
    closeEditTagsDialog();
  } catch (error) {
    notificationStore.showNotification('更新标签失败: ' + error.message, 'error');
  } finally {
    updating.value = false;
  }
};

const confirmRestoreBackup = (backup) => {
  selectedBackup.value = backup;
  showRestoreConfirmDialog.value = true;
};

const restoreBackup = async () => {
  restoring.value = true;
  try {
    await backupStore.restoreBackup(selectedBackup.value.id);
    notificationStore.showNotification('备份恢复成功', 'success');
    showRestoreConfirmDialog.value = false;
  } catch (error) {
    notificationStore.showNotification('恢复备份失败: ' + error.message, 'error');
  } finally {
    restoring.value = false;
  }
};

const confirmDeleteBackup = (backup) => {
  selectedBackup.value = backup;
  showDeleteConfirmDialog.value = true;
};

const deleteBackup = async () => {
  deleting.value = true;
  try {
    await backupStore.deleteBackup(selectedBackup.value.id);
    notificationStore.showNotification('备份删除成功', 'success');
    showDeleteConfirmDialog.value = false;
  } catch (error) {
    notificationStore.showNotification('删除备份失败: ' + error.message, 'error');
  } finally {
    deleting.value = false;
  }
};

// 过滤相关方法
const setFilter = (filter) => {
  currentFilter.value = filter;
};

const getEmptyMessage = () => {
  if (currentFilter.value === 'automatic') {
    return '暂无自动备份';
  } else if (currentFilter.value === 'manual') {
    return '暂无手动备份';
  }
  return '暂无备份';
};

const getEmptySubMessage = () => {
  if (currentFilter.value === 'automatic') {
    return '系统在应用配置时会自动创建备份';
  } else if (currentFilter.value === 'manual') {
    return '点击"创建备份"按钮手动创建备份';
  }
  return '创建第一个备份来保护您的hosts配置';
};

// 清理自动备份相关方法
const confirmClearAutoBackups = () => {
  showClearAutoBackupsDialog.value = true;
};

const clearAutoBackups = async () => {
  clearingAutoBackups.value = true;
  try {
    await backupStore.clearAllAutoBackups();
    notificationStore.showNotification(`已清理 ${backupStore.stats.automatic} 个自动备份`, 'success');
    showClearAutoBackupsDialog.value = false;
  } catch (error) {
    notificationStore.showNotification('清理自动备份失败: ' + error.message, 'error');
  } finally {
    clearingAutoBackups.value = false;
  }
};

// 生命周期
onMounted(async () => {
  // 确保 Go 后端已经准备好
  if (window.go && window.go.services && window.go.services.ConfigService) {
    try {
      await refreshBackupsSilently();
    } catch (error) {
      console.error('初始化备份数据失败:', error);
      notificationStore.showNotification('初始化备份数据失败: ' + error.message, 'error');
    }
  } else {
    console.warn('Go 后端服务尚未准备好，延迟加载备份数据');
    // 延迟尝试
    setTimeout(async () => {
      if (window.go && window.go.services && window.go.services.ConfigService) {
        try {
          await refreshBackupsSilently();
        } catch (error) {
          console.error('延迟加载备份数据失败:', error);
        }
      }
    }, 1000);
  }
});
</script>

<style scoped>
.backup-manager {
  max-width: 1200px;
  margin: 0 auto;
}

.backup-list-container {
  min-height: 400px;
  transition: height 0.3s ease;
}

.empty-state {
  min-height: 300px;
  display: flex;
  flex-direction: column;
  justify-content: center;
}

.stats-card {
  transition: box-shadow 0.2s ease;
}

.backup-timeline-container {
  min-height: 200px;
  max-height: calc(100vh - 300px);
  overflow-y: auto;
  padding: 8px 8px 24px 0;
}

.backup-timeline-container::-webkit-scrollbar {
  width: 6px;
}

.backup-timeline-container::-webkit-scrollbar-track {
  background: rgba(var(--v-theme-surface-variant), 0.3);
  border-radius: 3px;
}

.backup-timeline-container::-webkit-scrollbar-thumb {
  background: rgba(var(--v-theme-primary), 0.5);
  border-radius: 3px;
}

.backup-timeline-container::-webkit-scrollbar-thumb:hover {
  background: rgba(var(--v-theme-primary), 0.7);
}

.backup-timeline {
  padding: 16px 16px 32px 16px;
}

.backup-card {
  transition: box-shadow 0.2s ease;
}

.backup-card:hover {
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1) !important;
}

.backup-card.compact {
  min-height: auto;
}

.backup-card.compact .v-card-text {
  padding-top: 12px !important;
  padding-bottom: 12px !important;
}

.backup-preview {
  background: rgba(var(--v-theme-surface-variant), 0.3);
  padding: 12px;
  border-radius: 8px;
  border-left: 3px solid rgb(var(--v-theme-primary));
}

.backup-preview.clickable {
  cursor: pointer;
  transition: background-color 0.2s ease;
}

.backup-preview.clickable:hover {
  background: rgba(var(--v-theme-surface-variant), 0.5);
}

.backup-preview.compact {
  padding: 8px;
}

.preview-content {
  font-family: 'Courier New', monospace;
  font-size: 12px;
  line-height: 1.4;
  margin: 0;
  white-space: pre-wrap;
  word-break: break-all;
  color: rgb(var(--v-theme-on-surface-variant));
  max-height: 120px;
  overflow: hidden;
}

.preview-content.compact {
  max-height: 60px;
  font-size: 11px;
}

.preview-full-content {
  font-family: 'Courier New', monospace;
  font-size: 13px;
  line-height: 1.4;
  margin: 0;
  white-space: pre-wrap;
  word-break: break-all;
  background: rgba(var(--v-theme-surface-variant), 0.3);
  padding: 16px;
  border-radius: 8px;
  max-height: 400px;
  overflow-y: auto;
}

.clickable-card {
  cursor: pointer;
  transition: box-shadow 0.2s ease;
}

.clickable-card:hover {
  box-shadow: 0 8px 20px rgba(0, 0, 0, 0.15) !important;
}

.clickable-card.active-filter {
  box-shadow: 0 6px 16px rgba(0, 0, 0, 0.2) !important;
  border: 2px solid rgba(var(--v-theme-surface), 0.3);
}

.dialog-card {
  border: 2px solid rgba(var(--v-theme-primary), 0.2);
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.15) !important;
}
</style> 