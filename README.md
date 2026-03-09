## Hosts 管理器

一个现代化的 Windows Hosts 文件管理工具，基于 .NET 8 和 Avalonia UI 开发。

## 功能特性

- 📝 **可视化编辑** - 实时编辑
- 💾 **智能备份** - 自动备份与恢复，最多保留 50 个版本
- 🔄 **远程同步** - 从 URL 下载 Hosts，支持追加和覆盖模式
- 🎨 **现代化 UI** - Windows 11 Fluent Design 风格

## 快速开始

### 系统要求

- Windows 10 / Windows 11
- .NET 8 SDK

### 运行项目

```bash
# 还原依赖
dotnet restore

# 运行项目
dotnet run
```

## 使用说明

### 编辑 Hosts 文件
1. 启动程序（自动请求管理员权限）
2. 在编辑器中修改内容
3. 点击"保存更改"（自动创建备份）

### 备份管理
1. 点击左侧"备份管理"
2. 查看所有历史备份
3. 选择备份后点击"恢复备份"

### 远程同步
1. 点击左侧"远程同步"
2. 选择同步模式（追加/覆盖）
3. 选择同步源（预置 GitHub Hosts）
4. 点击"开始同步"

## 技术栈

- .NET 8
- C#
- Avalonia UI 11.0
- MVVM 架构（CommunityToolkit.Mvvm）
- AvaloniaEdit（代码编辑器）
- Serilog（日志系统）