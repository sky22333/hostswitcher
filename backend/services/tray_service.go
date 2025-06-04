package services

import (
	"context"
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"

	"github.com/getlantern/systray"
	"github.com/skratchdot/open-golang/open"
	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"

	"hostswitcher/backend/models"
)

// 嵌入图标文件到程序中
//go:embed assets/appicon.ico
var iconDataWin []byte

//go:embed assets/appicon.icns
var iconDataMac []byte

//go:embed assets/appicon.png
var iconDataLinux []byte

// TrayService 处理系统托盘相关操作的服务
type TrayService struct {
	ctx           context.Context
	configService *ConfigService
	configs       []*models.Config
	activeConfig  *models.Config
	isRunning     bool
	onExit        func()
	mutex         sync.RWMutex // 添加互斥锁保护并发访问
	stopChan      chan struct{} // 用于停止后台goroutine的通道
}

// NewTrayService 创建一个新的托盘服务实例
func NewTrayService(configService *ConfigService) *TrayService {
	return &TrayService{
		configService: configService,
		configs:       []*models.Config{},
		isRunning:     false,
		mutex:         sync.RWMutex{}, // 初始化互斥锁
		stopChan:      make(chan struct{}), // 初始化stop channel
	}
}

// getEmbeddedIcon 根据操作系统获取对应的嵌入图标
func (s *TrayService) getEmbeddedIcon() []byte {
	switch runtime.GOOS {
	case "windows":
		if s.ctx != nil {
			wailsRuntime.LogInfo(s.ctx, fmt.Sprintf("使用Windows嵌入图标，大小: %d 字节", len(iconDataWin)))
		}
		return iconDataWin
	case "darwin":
		if s.ctx != nil {
			wailsRuntime.LogInfo(s.ctx, fmt.Sprintf("使用macOS嵌入图标，大小: %d 字节", len(iconDataMac)))
		}
		return iconDataMac
	case "linux":
		if s.ctx != nil {
			wailsRuntime.LogInfo(s.ctx, fmt.Sprintf("使用Linux嵌入图标，大小: %d 字节", len(iconDataLinux)))
		}
		return iconDataLinux
	default:
		// 默认使用Windows图标
		if s.ctx != nil {
			wailsRuntime.LogInfo(s.ctx, fmt.Sprintf("未知操作系统，使用默认Windows图标，大小: %d 字节", len(iconDataWin)))
		}
		return iconDataWin
	}
}

// SetContext 设置上下文
// wails:ignore
func (s *TrayService) SetContext(ctx context.Context) {
	s.ctx = ctx
}

// SetOnExit 设置退出回调
// wails:ignore
func (s *TrayService) SetOnExit(onExit func()) {
	s.onExit = onExit
}

// Start 启动托盘图标
func (s *TrayService) Start() {
	s.mutex.Lock()
	if s.isRunning {
		s.mutex.Unlock()
		return
	}
	s.isRunning = true
	s.mutex.Unlock()

	// 加载配置
	s.configs = s.configService.GetAllConfigs()
	s.activeConfig = s.configService.GetActiveConfig()

	// 启动托盘图标 - 在goroutine中运行以避免阻塞
	systray.Run(s.onReady, func() {
		// 退出时的清理
		if s.onExit != nil {
			s.onExit()
		}
	})
}

// Stop 停止托盘图标
func (s *TrayService) Stop() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	if !s.isRunning {
		return
	}
	
	// 发送停止信号并重新创建channel
	close(s.stopChan)
	s.stopChan = make(chan struct{})
	
	s.isRunning = false
	systray.Quit()
}

// UpdateConfigs 更新配置列表
func (s *TrayService) UpdateConfigs() {
	s.mutex.Lock()
	if !s.isRunning {
		s.mutex.Unlock()
		return
	}
	
	s.configs = s.configService.GetAllConfigs()
	s.activeConfig = s.configService.GetActiveConfig()
	s.mutex.Unlock()

	// 通知托盘图标更新
	if s.ctx != nil {
		wailsRuntime.EventsEmit(s.ctx, "tray-update-configs")
	}
}

// onReady 托盘图标准备就绪
func (s *TrayService) onReady() {
	// 获取嵌入的图标数据
	iconData := s.getEmbeddedIcon()
	
	// 记录图标加载信息
	if s.ctx != nil {
		wailsRuntime.LogInfo(s.ctx, fmt.Sprintf("正在设置托盘图标，大小: %d 字节", len(iconData)))
	}
	
	// 设置托盘图标
	systray.SetIcon(iconData)
	systray.SetTitle("host 管理工具")
	systray.SetTooltip("host 管理工具 - 双击显示主窗口")

	// 创建托盘菜单
	mShow := systray.AddMenuItem("显示主界面", "显示主应用界面")
	systray.AddSeparator()
	
	mRestoreDefault := systray.AddMenuItem("恢复默认", "恢复默认系统 hosts 文件")
	mRefreshRemote := systray.AddMenuItem("更新远程源", "更新所有远程 hosts 源")
	
	systray.AddSeparator()
	mExit := systray.AddMenuItem("退出", "退出应用")
	
	// 处理菜单事件
	go s.handleMenuEvents(mShow, mRestoreDefault, mRefreshRemote, mExit)
}

// handleMenuEvents 处理菜单事件
func (s *TrayService) handleMenuEvents(mShow, mRestoreDefault, mRefreshRemote, mExit *systray.MenuItem) {
	defer func() {
		if r := recover(); r != nil {
			if s.ctx != nil {
				wailsRuntime.LogError(s.ctx, fmt.Sprintf("托盘菜单处理发生错误: %v", r))
			}
		}
	}()
	
	// 启动独立的goroutine处理每个菜单项
	go func() {
		for {
			select {
			case <-s.stopChan:
				return
			case <-mShow.ClickedCh:
				// 显示主窗口
				if s.ctx != nil {
					wailsRuntime.WindowShow(s.ctx)
					wailsRuntime.WindowUnminimise(s.ctx)
					wailsRuntime.WindowSetAlwaysOnTop(s.ctx, false)
					wailsRuntime.WindowCenter(s.ctx)
				}
			}
		}
	}()
	
	go func() {
		for {
			select {
			case <-s.stopChan:
				return
			case <-mRestoreDefault.ClickedCh:
				// 恢复默认hosts文件
				if s.ctx != nil {
					wailsRuntime.EventsEmit(s.ctx, "tray-restore-default")
				}
			}
		}
	}()
	
	go func() {
		for {
			select {
			case <-s.stopChan:
				return
			case <-mRefreshRemote.ClickedCh:
				// 更新远程源
				if s.ctx != nil {
					wailsRuntime.EventsEmit(s.ctx, "tray-refresh-remote")
				}
			}
		}
	}()
	
	go func() {
		for {
			select {
			case <-s.stopChan:
				return
			case <-mExit.ClickedCh:
				// 退出应用
				if s.ctx != nil {
					wailsRuntime.LogInfo(s.ctx, "用户从托盘退出应用")
					wailsRuntime.Quit(s.ctx)
				}
				systray.Quit()
			}
		}
	}()
}

// ShowNotification 显示系统通知
func (s *TrayService) ShowNotification(title, message string) {
	if s.ctx != nil {
		wailsRuntime.EventsEmit(s.ctx, "show-notification", title, message)
	}
}

// OpenSystemHostsFile 打开系统 hosts 文件
func (s *TrayService) OpenSystemHostsFile() error {
	var hostsPath string
	if GetOSType() == "windows" {
		hostsPath = filepath.Join(os.Getenv("SystemRoot"), "System32", "drivers", "etc", "hosts")
	} else {
		hostsPath = "/etc/hosts"
	}

	return open.Run(hostsPath)
}

// OpenUserDataDir 打开用户数据目录
func (s *TrayService) OpenUserDataDir() error {
	dataDir := s.configService.GetUserDataDir()
	
	if s.ctx != nil {
		wailsRuntime.LogInfo(s.ctx, fmt.Sprintf("尝试打开用户数据目录: %s", dataDir))
	}
	
	err := open.Run(dataDir)
	if err != nil && s.ctx != nil {
		wailsRuntime.LogError(s.ctx, fmt.Sprintf("打开用户数据目录失败: %v", err))
	}
	
	return err
}

// Cleanup 清理资源
func (s *TrayService) Cleanup() {
	s.Stop() // 确保停止所有后台goroutine
	
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	// 清空内存中的配置，帮助GC回收
	s.configs = nil
	s.activeConfig = nil
}

// GetContext 获取上下文
// wails:ignore
func (s *TrayService) GetContext() context.Context {
	return s.ctx
}
