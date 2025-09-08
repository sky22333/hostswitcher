## 🌐 Hosts 管理工具

[![Go](https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat&logo=go&logoColor=white)](https://golang.org/)
[![Vue.js](https://img.shields.io/badge/Vue.js-3.x-4FC08D?style=flat&logo=vue.js&logoColor=white)](https://vuejs.org/)
[![Wails](https://img.shields.io/badge/Wails-2.x-FF6B6B?style=flat&logo=wails&logoColor=white)](https://wails.io/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

一个现代化的windows hosts 文件管理工具，提供直观的图形界面和强大的管理功能。支持本地配置管理、远程源同步、系统托盘集成等特性。使用`Wails v2`开发。

<p align="center">
  <img src="https://count.getloli.com/get/@sky22333.hostswitcher?theme=rule34" alt="Visitors">
</p>

## ✨ 功能特性

### 🎯 核心功能
- **可视化编辑**: 基于 Monaco Editor 的代码编辑器
- **配置管理**: 快捷修改 hosts 配置
- **实时预览**: 实时显示文件状态和变更提示
- **权限检测**: 自动检测管理员权限，友好的权限提示
- **备份恢复**: 一键管理备份、恢复多个 hosts 文件
- **远程同步**: 支持从远程 URL 获取 hosts 配置

## 💻 系统要求

### 运行环境
- Windows 10 +
- Windows Server 2016 +
- WebView2 Runtime (通常已预装)
- **权限**: 修改 hosts 文件需要管理员权限，请以管理员身份运行。
- **报毒问题**: 修改hosts是敏感行为，可能会被误报病毒，可以添加白名单运行。

## 🔧 开发环境配置

### 1. 克隆项目

### 2. 安装 Wails CLI
```bash
go install github.com/wailsapp/wails/v2/cmd/wails@latest
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

---

如果这个项目对您有帮助，请考虑给个 ⭐ Star！

## 📸 应用预览
*Host编辑器*
![主界面](/.github/demo/1.png)

*远程源管理*
![远程源管理](/.github/demo/2.png)

*备份管理*
![备份管理](/.github/demo/3.png)

*应用设置*
![设置页面](/.github/demo/4.png)


---

## Stargazers over time
[![Stargazers over time](https://starchart.cc/sky22333/hostswitcher.svg?variant=adaptive)](https://starchart.cc/sky22333/hostswitcher)

