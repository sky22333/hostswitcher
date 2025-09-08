package services

import (
	"context"
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"github.com/getlantern/systray"
	"github.com/skratchdot/open-golang/open"
	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"

	"hostswitcher/backend/models"
)


//go:embed assets/appicon.ico
var iconDataWin []byte

//go:embed assets/appicon.icns
var iconDataMac []byte

//go:embed assets/appicon.png
var iconDataLinux []byte

// TrayService 托盘服务
type TrayService struct {
	ctx           context.Context
	configService *ConfigService
	configs       []*models.Config
	activeConfig  *models.Config
	isRunning     bool
	onExit        func()
	mutex         sync.RWMutex
	stopChan      chan struct{}
}

// NewTrayService 创建服务
func NewTrayService(configService *ConfigService) *TrayService {
	return &TrayService{
		configService: configService,
		configs:       []*models.Config{},
		isRunning:     false,
		mutex:         sync.RWMutex{},
		stopChan:      make(chan struct{}),
	}
}

// getEmbeddedIcon 获取图标
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
		if s.ctx != nil {
			wailsRuntime.LogInfo(s.ctx, fmt.Sprintf("未知操作系统，使用默认Windows图标，大小: %d 字节", len(iconDataWin)))
		}
		return iconDataWin
	}
}

// SetContext 设置ctx
// wails:ignore
func (s *TrayService) SetContext(ctx context.Context) {
	s.ctx = ctx
}

// SetOnExit 设置回调
// wails:ignore
func (s *TrayService) SetOnExit(onExit func()) {
	s.onExit = onExit
}

// Start 启动
func (s *TrayService) Start() {
	s.mutex.Lock()
	if s.isRunning {
		s.mutex.Unlock()
		return
	}
	s.isRunning = true
	s.mutex.Unlock()

	s.configs = s.configService.GetAllConfigs()
	s.activeConfig = s.configService.GetActiveConfig()

	systray.Run(s.onReady, func() {
		if s.onExit != nil {
			s.onExit()
		}
	})
}

// Stop 停止
func (s *TrayService) Stop() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	if !s.isRunning {
		return
	}
	
	close(s.stopChan)
	s.stopChan = make(chan struct{})
	
	s.isRunning = false
	systray.Quit()
}

// UpdateConfigs 更新配置
func (s *TrayService) UpdateConfigs() {
	s.mutex.Lock()
	if !s.isRunning {
		s.mutex.Unlock()
		return
	}
	
	s.configs = s.configService.GetAllConfigs()
	s.activeConfig = s.configService.GetActiveConfig()
	s.mutex.Unlock()

	if s.ctx != nil {
		wailsRuntime.EventsEmit(s.ctx, "tray-update-configs")
	}
}

// onReady 准备就绪
func (s *TrayService) onReady() {
	iconData := s.getEmbeddedIcon()
	
	if s.ctx != nil {
		wailsRuntime.LogInfo(s.ctx, fmt.Sprintf("正在设置托盘图标，大小: %d 字节", len(iconData)))
	}
	
	systray.SetIcon(iconData)
	systray.SetTitle("host 管理工具")
	systray.SetTooltip("host 管理工具")

	mShow := systray.AddMenuItem("显示主界面", "显示主应用界面")
	systray.AddSeparator()
	
	mRefreshRemote := systray.AddMenuItem("更新远程源", "更新所有远程 hosts 源")
	
	systray.AddSeparator()
	mExit := systray.AddMenuItem("退出", "退出应用")
	
	go s.handleMenuEvents(mShow, mRefreshRemote, mExit)
}

// handleMenuEvents 处理事件
func (s *TrayService) handleMenuEvents(mShow, mRefreshRemote, mExit *systray.MenuItem) {
	defer func() {
		if r := recover(); r != nil {
			if s.ctx != nil {
				wailsRuntime.LogError(s.ctx, fmt.Sprintf("托盘菜单处理发生错误: %v", r))
			}
			time.Sleep(1 * time.Second)
			if !s.isRunning {
				return
			}
			go s.handleMenuEvents(mShow, mRefreshRemote, mExit)
		}
	}()
	
	var lastProcessTime time.Time
	debounceInterval := 500 * time.Millisecond
	for {
		select {
		case <-s.stopChan:
			return
			
		case <-mShow.ClickedCh:
			if time.Since(lastProcessTime) < debounceInterval {
				continue
			}
			lastProcessTime = time.Now()
			
			go func() {
				if s.ctx != nil {
					wailsRuntime.WindowShow(s.ctx)
					wailsRuntime.WindowUnminimise(s.ctx)
					wailsRuntime.WindowSetAlwaysOnTop(s.ctx, false)
					wailsRuntime.WindowCenter(s.ctx)
				}
			}()
			
		case <-mRefreshRemote.ClickedCh:
			if time.Since(lastProcessTime) < debounceInterval {
				continue
			}
			lastProcessTime = time.Now()
			
			go func() {
				if s.ctx != nil {
					wailsRuntime.EventsEmit(s.ctx, "tray-refresh-remote")
				}
			}()
			
		case <-mExit.ClickedCh:
			go func() {
				if s.ctx != nil {
					wailsRuntime.LogInfo(s.ctx, "用户从托盘退出应用")
					wailsRuntime.Quit(s.ctx)
				}
				systray.Quit()
			}()
			return
		}
	}
}

// OpenSystemHostsFile 打开hosts文件
func (s *TrayService) OpenSystemHostsFile() error {
	var hostsPath string
	if GetOSType() == "windows" {
		hostsPath = filepath.Join(os.Getenv("SystemRoot"), "System32", "drivers", "etc", "hosts")
	} else {
		hostsPath = "/etc/hosts"
	}

	return open.Run(hostsPath)
}

// OpenUserDataDir 打开数据目录
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

// OpenBrowser 打开浏览器
func (s *TrayService) OpenBrowser(url string) error {
	if s.ctx != nil {
		wailsRuntime.LogInfo(s.ctx, fmt.Sprintf("尝试打开浏览器访问: %s", url))
	}
	
	err := open.Run(url)
	if err != nil && s.ctx != nil {
		wailsRuntime.LogError(s.ctx, fmt.Sprintf("打开浏览器失败: %v", err))
	}
	
	return err
}

// Cleanup 清理
func (s *TrayService) Cleanup() {
	s.Stop()
	
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	s.configs = nil
	s.activeConfig = nil
}

// GetContext 获取ctx
// wails:ignore
func (s *TrayService) GetContext() context.Context {
	return s.ctx
}
