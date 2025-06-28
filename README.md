# 🌐 Hosts 管理工具

[![Go](https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat&logo=go&logoColor=white)](https://golang.org/)
[![Vue.js](https://img.shields.io/badge/Vue.js-3.x-4FC08D?style=flat&logo=vue.js&logoColor=white)](https://vuejs.org/)
[![Wails](https://img.shields.io/badge/Wails-2.x-FF6B6B?style=flat&logo=wails&logoColor=white)](https://wails.io/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

一个现代化的windows hosts 文件管理工具，提供直观的图形界面和强大的管理功能。支持本地配置管理、远程源同步、系统托盘集成等特性。

<p align="center">
  <img src="https://count.getloli.com/get/@sky22333.hostswitcher?theme=rule34" alt="Visitors">
</p>


## ✨ 功能特性

### 🎯 核心功能
- **可视化编辑**: 基于 Monaco Editor 的代码编辑器，支持语法高亮和智能提示
- **配置管理**: 快捷修改 hosts 配置
- **实时预览**: 实时显示文件状态和变更提示
- **权限检测**: 自动检测管理员权限，友好的权限提示

### 🌐 远程源管理
- **远程同步**: 支持从远程 URL 获取 hosts 配置
- **自动更新**: 可配置启动时自动更新远程源
- **源管理**: 添加、编辑、删除远程 hosts 源
- **状态监控**: 实时显示远程源的更新状态和时间

### 🔧 系统集成
- **托盘支持**: 最小化到系统托盘，支持快捷操作
- **快速切换**: 通过托盘菜单快速应用不同配置
- **系统文件**: 直接编辑系统 hosts 文件
- **备份恢复**: 一键管理备份、恢复多个 hosts 文件

### 🎨 用户界面
- **现代设计**: 基于 Vuetify 3 的 Material Design 界面
- **暗色主题**: 支持亮色/暗色主题切换
- **响应式布局**: 适配不同窗口大小
- **通知系统**: 统一的消息提示机制

## 🛠️ 技术栈

### 后端技术
- **[Go](https://golang.org/)**: 1.22+ - 主要开发语言
- **[Wails 2](https://wails.io/)**: 2.10+ - 跨平台桌面应用框架
- **[Systray](https://github.com/getlantern/systray)**: 系统托盘支持
- **[UUID](https://github.com/google/uuid)**: 唯一标识符生成

### 前端技术
- **[Vue.js](https://vuejs.org/)**: 3.3+ - 渐进式 JavaScript 框架
- **[Vuetify](https://vuetifyjs.com/)**: 3.5+ - Vue.js Material Design 组件库
- **[Pinia](https://pinia.vuejs.org/)**: 2.1+ - Vue 状态管理库
- **[Monaco Editor](https://microsoft.github.io/monaco-editor/)**: 0.46+ - 代码编辑器
- **[Vue Router](https://router.vuejs.org/)**: 4.2+ - Vue.js 路由管理

### 构建工具
- **[Vite](https://vitejs.dev/)**: 4.4+ - 前端构建工具
- **[Tailwind CSS](https://tailwindcss.com/)**: 3.3+ - 实用优先的 CSS 框架
- **[Sass](https://sass-lang.com/)**: 1.69+ - CSS 预处理器

## 💻 系统要求

### 运行环境
- **操作系统**: Windows 10/11
- **权限**: 修改 hosts 文件需要管理员权限，请以管理员身份运行。
- **报毒问题**: 修改hosts是敏感行为，可能会被误报病毒，可以添加白名单运行。本程序完全开源，不放心可以自行编译。


### 开发环境
- **Go**: 1.22 或更高版本
- **Node.js**: 16.0 或更高版本
- **npm**: 8.0 或更高版本
- **Git**: 版本控制


## 🔧 开发环境配置

### 1. 克隆项目


### 2. 安装 Wails CLI
```bash
go install github.com/wailsapp/wails/v2/cmd/wails@latest
```

### 3. 安装依赖
```bash
# 安装 Go 依赖
go mod tidy

# 安装前端依赖
cd frontend
npm install
cd ..
```

### 4. 启动开发服务器
```bash
wails dev
```

## 🏗️ 编译构建

### 开发构建
```bash
# 启动开发模式（热重载）
wails dev
```

### 生产构建
```bash
# 构建前端资源
cd frontend
npm run build
cd ..

# 构建应用程序
wails build

# 指定平台构建（Windows）
wails build -platform windows/amd64
```

### 构建选项
```bash
# 构建优化版本（生产环境）
wails build -clean -upx

# 构建无控制台窗口版本（Windows）
wails build -ldflags "-H windowsgui"
```

## 📁 项目结构

```
hostswitcher/
├── 📁 backend/                 # Go 后端服务
│   ├── 📁 models/             # 数据模型定义
│   │   └── 📄 config.go       # 配置和远程源模型
│   └── 📁 services/           # 业务逻辑服务
│       ├── 📁 assets/         # 嵌入的资源文件
│       │   ├── 📄 appicon.ico # Windows 图标
│       │   ├── 📄 appicon.icns# macOS 图标
│       │   └── 📄 appicon.png # Linux 图标
│       ├── 📄 config_service.go   # 配置管理服务
│       ├── 📄 network_service.go  # 网络请求服务
│       └── 📄 tray_service.go     # 系统托盘服务
├── 📁 frontend/                # Vue.js 前端界面
│   ├── 📁 public/             # 静态资源
│   ├── 📁 src/                # 源代码
│   │   ├── 📁 components/     # 可复用组件
│   │   │   ├── 📄 MonacoEditor.vue      # 代码编辑器组件
│   │   │   └── 📄 NotificationSystem.vue # 通知系统组件
│   │   ├── 📁 stores/         # Pinia 状态管理
│   │   │   ├── 📄 config.js   # 配置状态管理
│   │   │   ├── 📄 notification.js # 通知状态管理
│   │   │   └── 📄 theme.js    # 主题状态管理
│   │   ├── 📁 views/          # 页面视图
│   │   │   ├── 📄 HostsEditor.vue # Hosts 编辑器页面
│   │   │   ├── 📄 RemoteHosts.vue # 远程源管理页面
│   │   │   └── 📄 Settings.vue    # 设置页面
│   │   ├── 📄 App.vue         # 主应用组件
│   │   ├── 📄 main.js         # 应用入口文件
│   │   └── 📄 style.css       # 全局样式
│   ├── 📄 index.html          # HTML 模板
│   ├── 📄 package.json        # npm 配置文件
│   ├── 📄 tailwind.config.js  # Tailwind CSS 配置
│   └── 📄 vite.config.js      # Vite 配置文件
├── 📄 main.go                  # Go 应用程序入口
├── 📄 wails.json              # Wails 项目配置
├── 📄 go.mod                  # Go 模块定义
├── 📄 go.sum                  # Go 依赖校验
├── 📄 .gitignore              # Git 忽略文件配置
└── 📄 README.md               # 项目说明文档
```

### 核心文件说明

#### 后端核心文件
- **`main.go`**: 应用程序主入口，初始化服务和 Wails 应用
- **`config_service.go`**: 处理 hosts 配置的增删改查、系统文件读写
- **`network_service.go`**: 处理远程源的网络请求和内容同步
- **`tray_service.go`**: 处理系统托盘集成和快捷操作

#### 前端核心文件
- **`App.vue`**: 主应用框架，包含导航栏和路由视图
- **`HostsEditor.vue`**: hosts 文件编辑器
- **`RemoteHosts.vue`**: 远程源管理界面
- **`Settings.vue`**: 应用设置和配置界面

## 📄 许可证

本项目采用 MIT 许可证 - 分发请保留项目地址。

## 🙏 致谢

- [Wails](https://wails.io/) - 出色的 Go + Web 桌面应用框架
- [Vue.js](https://vuejs.org/) - 渐进式 JavaScript 框架
- [Vuetify](https://vuetifyjs.com/) - 优秀的 Vue.js 组件库
- [Monaco Editor](https://microsoft.github.io/monaco-editor/) - 强大的代码编辑器

---

如果这个项目对您有帮助，请考虑给个 ⭐ Star！

## 📸 应用预览
*Hosts 编辑器 - 基于 Monaco Editor 的代码编辑体验*
![主界面](/.github/1.jpg)

*远程源管理 - 轻松管理多个远程 hosts 源*
![远程源管理](/.github/2.jpg)

*备份管理 - 轻松管理多个备份 hosts 源*
![备份管理](/.github/3.jpg)

*设置页面 - 主题切换和系统集成功能*
![设置页面](/.github/4.jpg)


---

[![Star History Chart](https://api.star-history.com/svg?repos=sky22333/hostswitcher&type=Date)](https://www.star-history.com/#sky22333/hostswitcher&Date)
