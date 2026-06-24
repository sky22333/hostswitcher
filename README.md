## Hosts 管理器

一个现代化的 Windows Hosts 文件管理工具，基于 .NET 8 和 Avalonia UI 12 开发。

## 功能特性

- **可视化编辑** - 实时编辑系统 Hosts 文件
- **智能备份** - 自动备份与恢复，相同内容去重，最多保留 50 个版本
- **远程同步** - 从 URL 下载 Hosts，支持追加和覆盖模式
- **DNS 刷新** - 一键清除系统 DNS 缓存
- **现代化 UI** - Fluent Design 风格，矢量图标，系统托盘驻留

## 快速开始

### 系统要求

- Windows 10 / Windows 11 (x64)
- .NET 8 SDK（开发）或直接使用发布包

### 运行项目

```bash
dotnet restore
dotnet run
```

## 使用说明

### 编辑 Hosts 文件

1. 启动程序（自动请求管理员权限）
2. 在编辑器中修改内容
3. 点击「保存更改」（自动创建备份）

### 备份管理

1. 点击左侧「备份管理」
2. 双击备份项预览内容
3. 需要恢复时点击「恢复备份」（恢复前会自动备份当前内容）

### 远程同步

1. 点击左侧「远程同步」
2. 选择同步模式（追加 / 覆盖）
3. 确认或修改同步 URL
4. 点击「开始同步」

## 技术栈

- .NET 8
- Avalonia UI 12
- MVVM（CommunityToolkit.Mvvm）
- Serilog

### 预览

![主界面](/.github/demo/demo.png)
