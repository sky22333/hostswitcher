<template>
  <div>
    <v-container fluid>
      <v-row>
        <v-col cols="12">
          <v-card class="app-card mb-4">
            <v-card-title class="text-h5">
              设置
            </v-card-title>
            
            <v-card-text>
              <v-list>
                <!-- 主题设置 -->
                <v-list-subheader>外观</v-list-subheader>
                <v-list-item>
                  <template v-slot:prepend>
                    <v-icon icon="mdi-theme-light-dark"></v-icon>
                  </template>
                  <v-list-item-title>主题</v-list-item-title>
                  <template v-slot:append>
                    <v-btn-toggle
                      v-model="themeMode"
                      color="primary"
                      rounded="lg"
                      density="comfortable"
                    >
                      <v-btn value="light">
                        <v-icon>mdi-weather-sunny</v-icon>
                        亮色
                      </v-btn>
                      <v-btn value="dark">
                        <v-icon>mdi-weather-night</v-icon>
                        暗色
                      </v-btn>
                      <v-btn value="system">
                        <v-icon>mdi-desktop-tower-monitor</v-icon>
                        跟随系统
                      </v-btn>
                    </v-btn-toggle>
                  </template>
                </v-list-item>
                
                <v-divider></v-divider>
                
                <!-- 系统 Hosts 文件 -->
                <v-list-subheader>系统 Hosts 文件</v-list-subheader>
                <v-list-item>
                  <template v-slot:prepend>
                    <v-icon icon="mdi-file-cog"></v-icon>
                  </template>
                  <v-list-item-title>打开系统Hosts文件</v-list-item-title>
                  <v-list-item-subtitle>在系统默认编辑器中打开</v-list-item-subtitle>
                  <template v-slot:append>
                    <v-btn
                      color="primary"
                      variant="tonal"
                      @click="openSystemHostsFile"
                      :loading="openingFile"
                      :disabled="openingFile"
                    >
                      打开
                    </v-btn>
                  </template>
                </v-list-item>
                
                <v-list-item>
                  <template v-slot:prepend>
                    <v-icon icon="mdi-folder-open"></v-icon>
                  </template>
                  <v-list-item-title>打开用户本地数据目录</v-list-item-title>
                  <v-list-item-subtitle>打开应用数据存储目录</v-list-item-subtitle>
                  <template v-slot:append>
                    <v-btn
                      color="primary"
                      variant="tonal"
                      @click="openUserDataDir"
                      :loading="openingDataDir"
                      :disabled="openingDataDir"
                    >
                      打开
                    </v-btn>
                  </template>
                </v-list-item>
                
                <v-divider></v-divider>
                
                <!-- 关于 -->
                <v-list-subheader>关于</v-list-subheader>
                <v-list-item>
                  <template v-slot:prepend>
                    <v-icon icon="mdi-information"></v-icon>
                  </template>
                  <v-list-item-title>版本信息</v-list-item-title>
                  <v-list-item-subtitle>host 管理工具 v1.2</v-list-item-subtitle>
                </v-list-item>
                
                <v-list-item>
                  <template v-slot:prepend>
                    <v-icon icon="mdi-github"></v-icon>
                  </template>
                  <v-list-item-title>GitHub</v-list-item-title>
                  <v-list-item-subtitle>查看源代码和贡献</v-list-item-subtitle>
                  <template v-slot:append>
                    <v-btn
                      color="primary"
                      variant="tonal"
                      @click="openGithub"
                    >
                      打开
                    </v-btn>
                  </template>
                </v-list-item>
              </v-list>
            </v-card-text>
          </v-card>
        </v-col>
      </v-row>
    </v-container>
  </div>
</template>

<script setup>
import { ref, onMounted, watch } from 'vue';
import { useThemeStore } from '@/stores/theme';
import { useNotificationStore } from '@/stores/notification';

// 状态管理
const themeStore = useThemeStore();
const notificationStore = useNotificationStore();

// 设置状态
const themeMode = ref('system');
const openingFile = ref(false);
const openingDataDir = ref(false);

// 生命周期钩子
onMounted(async () => {
  // 初始化主题设置
  initThemeMode();
});

// 监听主题模式变化
watch(themeMode, (newMode) => {
  applyThemeMode(newMode);
});

// 初始化主题模式
function initThemeMode() {
  const savedMode = localStorage.getItem('themeMode');
  if (savedMode) {
    themeMode.value = savedMode;
  } else {
    // 默认跟随系统
    themeMode.value = 'system';
  }
  applyThemeMode(themeMode.value);
}

// 应用主题模式
function applyThemeMode(mode) {
  localStorage.setItem('themeMode', mode);
  
  if (mode === 'system') {
    // 检测系统主题
    const prefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches;
    themeStore.setTheme(prefersDark ? 'darkTheme' : 'lightTheme');
  } else if (mode === 'dark') {
    themeStore.setTheme('darkTheme');
  } else {
    themeStore.setTheme('lightTheme');
  }
}

// 打开系统Hosts文件
async function openSystemHostsFile() {
  openingFile.value = true;
  try {
    await window.go.services.TrayService.OpenSystemHostsFile();
    notificationStore.showNotification('已打开系统Hosts文件', 'success');
  } catch (error) {
    notificationStore.showNotification('打开系统Hosts文件失败: ' + error, 'error');
  } finally {
    openingFile.value = false;
  }
}

// 打开用户数据目录
async function openUserDataDir() {
  openingDataDir.value = true;
  try {
    await window.go.services.TrayService.OpenUserDataDir();
    notificationStore.showNotification('已打开用户数据目录', 'success');
  } catch (error) {
    notificationStore.showNotification('打开用户数据目录失败: ' + error, 'error');
  } finally {
    openingDataDir.value = false;
  }
}

// 打开GitHub页面
async function openGithub() {
  try {
    await window.go.services.TrayService.OpenBrowser('https://github.com/sky22333/hosts');
    notificationStore.showNotification('已打开GitHub页面', 'success');
  } catch (error) {
    notificationStore.showNotification('打开GitHub页面失败: ' + error, 'error');
  }
}
</script>

<style scoped>
.v-list-subheader {
  font-weight: 600;
  color: rgb(var(--v-theme-primary));
  text-transform: uppercase;
  letter-spacing: 0.5px;
  font-size: 0.75rem;
}

.v-list-item {
  margin: 4px 8px;
}

.v-list-item:hover {
  background: rgba(var(--v-theme-primary), 0.04);
}

.v-btn-toggle {
  overflow: hidden;
}

.v-btn-toggle .v-btn {
  font-size: 0.75rem;
  padding: 8px 12px;
  min-width: auto;
}

.v-btn-toggle .v-btn .v-icon {
  margin-right: 4px;
  font-size: 1rem;
}

/* 深色主题适配 */
.v-theme--darkTheme .v-list-item:hover {
  background: rgba(var(--v-theme-primary), 0.08);
}
</style>
