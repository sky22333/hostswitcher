package main

import (
	"context"
	"embed"
	"fmt"
	"log"
	"syscall"
	"unsafe"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
	"github.com/wailsapp/wails/v2/pkg/runtime"

	"hostswitcher/backend/services"
)

//go:embed all:frontend/dist
var assets embed.FS

var (
	kernel32                = syscall.NewLazyDLL("kernel32.dll")
	user32                  = syscall.NewLazyDLL("user32.dll")
	procCreateMutex         = kernel32.NewProc("CreateMutexW")
	procCloseHandle         = kernel32.NewProc("CloseHandle")
	procFindWindow          = user32.NewProc("FindWindowW")
	procShowWindow          = user32.NewProc("ShowWindow")
	procSetForegroundWindow = user32.NewProc("SetForegroundWindow")
)

func activateExistingWindow() bool {
	windowTitle, _ := syscall.UTF16PtrFromString("host 管理工具")
	hwnd, _, _ := procFindWindow.Call(0, uintptr(unsafe.Pointer(windowTitle)))

	if hwnd != 0 {
		procShowWindow.Call(hwnd, 9)
		procSetForegroundWindow.Call(hwnd)
		return true
	}
	return false
}

func createSingleInstanceMutex() (syscall.Handle, error) {
	mutexName, _ := syscall.UTF16PtrFromString("Global\\HostSwitcher_SingleInstance_Mutex")

	ret, _, err := procCreateMutex.Call(
		uintptr(0),
		uintptr(0),
		uintptr(unsafe.Pointer(mutexName)),
	)

	if ret == 0 {
		return 0, err
	}

	if err.(syscall.Errno) == 183 {
		return 0, fmt.Errorf("应用程序已在运行")
	}

	return syscall.Handle(ret), nil
}

func releaseMutex(handle syscall.Handle) {
	if handle != 0 {
		procCloseHandle.Call(uintptr(handle))
	}
}

func main() {
	mutexHandle, err := createSingleInstanceMutex()
	if err != nil {
		if activateExistingWindow() {
			log.Println("程序已在运行，已激活主窗口")
		} else {
			log.Println("程序已在运行，请检查系统托盘")
		}
		return
	}

	defer releaseMutex(mutexHandle)

	log.Println("启动 Host 管理工具...")

	tempCtx := context.Background()
	configService := services.NewConfigService(tempCtx)
	networkService := services.NewNetworkService(configService)
	trayService := services.NewTrayService(configService)

	err = configService.Initialize()
	if err != nil {
		log.Fatalf("初始化配置服务失败: %v", err)
	}

	err = networkService.Initialize()
	if err != nil {
		log.Fatalf("初始化网络服务失败: %v", err)
	}

	trayService.SetOnExit(func() {
		configService.Cleanup()
		networkService.Cleanup()
		trayService.Cleanup()

		runtime.Quit(trayService.GetContext())
	})

	app := wails.Run(&options.App{
		Title:             "host 管理工具",
		Width:             1020,
		Height:            800,
		MinWidth:          900,
		MinHeight:         600,
		DisableResize:     false,
		Fullscreen:        false,
		Frameless:         false,
		StartHidden:       false,
		HideWindowOnClose: true,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 32, G: 32, B: 32, A: 1},
		OnStartup: func(ctx context.Context) {
			configService.SetContext(ctx)
			networkService.SetContext(ctx)
			trayService.SetContext(ctx)

			go trayService.Start()
		},
		OnShutdown: func(ctx context.Context) {
			trayService.Stop()
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
