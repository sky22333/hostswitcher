package main

import (
	"context"
	"embed"
	"log"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
	"github.com/wailsapp/wails/v2/pkg/runtime"

	"hostswitcher/backend/services"
)

//go:embed all:frontend/dist
var assets embed.FS

// 主函数：应用程序入口点
func main() {
	// 创建服务实例
	tempCtx := context.Background()
	configService := services.NewConfigService(tempCtx)
	networkService := services.NewNetworkService(configService)
	trayService := services.NewTrayService(configService)

	// 初始化服务
	err := configService.Initialize()
	if err != nil {
		log.Fatalf("初始化配置服务失败: %v", err)
	}

	err = networkService.Initialize()
	if err != nil {
		log.Fatalf("初始化网络服务失败: %v", err)
	}

	// 设置托盘服务退出回调
	trayService.SetOnExit(func() {
		// 应用退出时清理资源
		configService.Cleanup()
		networkService.Cleanup()
		trayService.Cleanup()
		
		// 退出应用
		runtime.Quit(trayService.GetContext())
	})

	// 创建应用
	app := wails.Run(&options.App{
		Title:             "host 管理工具",
		Width:             1200,
		Height:            800,
		MinWidth:          900,
		MinHeight:         600,
		DisableResize:     false,
		Fullscreen:        false,
		Frameless:         false,
		StartHidden:       false,
		HideWindowOnClose: true,  // 关闭时隐藏到托盘
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 32, G: 32, B: 32, A: 1},
		OnStartup: func(ctx context.Context) {
			// 设置服务上下文
			configService.SetContext(ctx)
			networkService.SetContext(ctx)
			trayService.SetContext(ctx)
			
			// 在应用启动后再启动托盘服务
			go trayService.Start()
		},
		OnShutdown: func(ctx context.Context) {
			// 停止托盘图标
			trayService.Stop()
			
			// 清理资源
			configService.Cleanup()
			networkService.Cleanup()
			trayService.Cleanup()
		},
		Bind: []interface{}{
			configService,
			networkService,
			trayService,
		},
		Windows: &windows.Options{
			WebviewIsTransparent:              false,
			WindowIsTranslucent:               false,
			BackdropType:                      windows.Mica,
			DisableWindowIcon:                 false,
			DisableFramelessWindowDecorations: false,
			ResizeDebounceMS:                  10,
			OnSuspend:                         func() {},
			OnResume:                          func() {},
		},
	})

	if app != nil {
		log.Fatal(app)
	}
}
