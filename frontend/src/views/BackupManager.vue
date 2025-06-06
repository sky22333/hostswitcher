<template>
  <div class="backup-manager">
    <v-container fluid class="pa-0">
      <!-- 顶部统计卡片 -->
      <v-row class="mb-4">
        <v-col cols="12" md="3">
          <v-card elevation="0" class="stats-card" color="primary" variant="tonal">
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
          <v-card elevation="0" class="stats-card" color="success" variant="tonal">
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
          <v-card elevation="0" class="stats-card" color="info" variant="tonal">
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
                  icon
                  variant="text"
                  @click="refreshBackups"
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
      <v-card elevation="0" rounded="lg">
        <v-card-text class="pa-0">
          <div v-if="backupStore.loading && backupStore.sortedBackups.length === 0" class="text-center pa-8">
            <v-progress-circular indeterminate color="primary" size="48" class="mb-4"></v-progress-circular>
            <div class="text-body-1 text-medium-emphasis">加载备份中...</div>
          </div>

          <div v-else-if="backupStore.sortedBackups.length === 0" class="text-center pa-8">
            <v-icon size="64" color="medium-emphasis" class="mb-4">mdi-database-off</v-icon>
            <div class="text-h6 mb-2">暂无备份</div>
            <div class="text-body-2 text-medium-emphasis mb-4">创建第一个备份来保护您的hosts配置</div>
            <v-btn
              color="primary"
              @click="showCreateBackupDialog = true"
              prepend-icon="mdi-plus"
              variant="flat"
            >
              创建备份
            </v-btn>
          </div>

          <!-- 备份时间线 -->
          <v-timeline v-else side="end" class="backup-timeline">
            <v-timeline-item
              v-for="backup in backupStore.sortedBackups"
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

              <v-card class="backup-card" elevation="2" rounded="lg">
                <v-card-text class="pa-4">
                  <!-- 备份头部信息 -->
                  <div class="d-flex align-center mb-3">
                    <div class="flex-1">
                      <div class="d-flex align-center mb-1">
                        <span class="text-subtitle-1 font-weight-medium me-2">{{ backup.description }}</span>
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
                  <div v-if="backup.tags && backup.tags.length > 0" class="mb-2">
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

                  <!-- 内容预览 -->
                  <div class="backup-preview">
                    <div class="text-caption mb-1 text-medium-emphasis">内容预览:</div>
                    <pre class="preview-content">{{ backupStore.getBackupPreview(backup.content, 3) }}</pre>
                  </div>
                </v-card-text>
              </v-card>
            </v-timeline-item>
          </v-timeline>
        </v-card-text>
      </v-card>
    </v-container>

    <!-- 创建备份对话框 -->
    <v-dialog v-model="showCreateBackupDialog" max-width="500px" persistent>
      <v-card class="rounded-xl">
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

    <!-- 预览对话框 -->
    <v-dialog v-model="showContentPreviewDialog" max-width="800px">
      <v-card class="rounded-xl">
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

    <!-- 编辑标签对话框 -->
    <v-dialog v-model="showEditTagsDialog" max-width="500px" persistent>
      <v-card class="rounded-xl">
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

    <!-- 确认恢复对话框 -->
    <v-dialog v-model="showRestoreConfirmDialog" max-width="500px" persistent>
      <v-card class="rounded-xl">
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

    <!-- 确认删除对话框 -->
    <v-dialog v-model="showDeleteConfirmDialog" max-width="400px" persistent>
      <v-card class="rounded-xl">
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

const newBackupDescription = ref('');
const newBackupTags = ref([]);
const editingTags = ref([]);
const selectedBackup = ref(null);

const creating = ref(false);
const updating = ref(false);
const restoring = ref(false);
const deleting = ref(false);

// 计算属性
const formatTimestamp = computed(() => {
  return (timestamp) => {
    return new Date(timestamp).toLocaleString('zh-CN', {
      year: 'numeric',
      month: '2-digit',
      day: '2-digit',
      hour: '2-digit',
      minute: '2-digit',
      second: '2-digit'
    });
  };
});

// 方法
const refreshBackups = async () => {
  try {
    await backupStore.loadBackups();
    notificationStore.showNotification('备份列表已刷新', 'success');
  } catch (error) {
    notificationStore.showNotification('刷新备份列表失败: ' + error.message, 'error');
  }
};

const createBackup = async () => {
  creating.value = true;
  try {
    await backupStore.createBackup(newBackupDescription.value, newBackupTags.value);
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

// 生命周期
onMounted(async () => {
  await refreshBackups();
});
</script>

<style scoped>
.backup-manager {
  max-width: 1200px;
  margin: 0 auto;
}

.stats-card {
  transition: transform 0.2s ease;
}

.stats-card:hover {
  transform: translateY(-2px);
}

.backup-timeline {
  padding: 16px;
}

.backup-card {
  transition: transform 0.2s ease, box-shadow 0.2s ease;
}

.backup-card:hover {
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1) !important;
}

.backup-preview {
  background: rgba(var(--v-theme-surface-variant), 0.3);
  padding: 12px;
  border-radius: 8px;
  border-left: 3px solid rgb(var(--v-theme-primary));
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

.icon-btn {
  transition: transform 0.2s ease;
}

.icon-btn:hover {
  transform: scale(1.1);
}
</style> 