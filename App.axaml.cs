using Avalonia;
using Avalonia.Controls;
using Avalonia.Controls.ApplicationLifetimes;
using Avalonia.Markup.Xaml;
using Avalonia.Threading;
using HostsManager.Services;
using HostsManager.ViewModels;
using HostsManager.Views;
using Microsoft.Extensions.DependencyInjection;
using System;

namespace HostsManager;

public partial class App : Application
{
    public static IServiceProvider Services { get; private set; } = null!;
    private static IClassicDesktopStyleApplicationLifetime? _desktop;

    public override void Initialize()
    {
        AvaloniaXamlLoader.Load(this);
        ConfigureServices();
    }

    public override void OnFrameworkInitializationCompleted()
    {
        if (ApplicationLifetime is IClassicDesktopStyleApplicationLifetime desktop)
        {
            _desktop = desktop;
            
            desktop.MainWindow = new MainWindow
            {
                DataContext = Services.GetRequiredService<MainWindowViewModel>()
            };

            desktop.ShutdownMode = ShutdownMode.OnExplicitShutdown;
        }

        base.OnFrameworkInitializationCompleted();
    }

    public static void ActivateMainWindow()
    {
        Dispatcher.UIThread.Post(() =>
        {
            if (_desktop?.MainWindow != null)
            {
                _desktop.MainWindow.Show();
                _desktop.MainWindow.WindowState = WindowState.Normal;
                _desktop.MainWindow.Activate();
                _desktop.MainWindow.Topmost = true;
                _desktop.MainWindow.Topmost = false;
            }
        });
    }

    private void ConfigureServices()
    {
        var services = new ServiceCollection();

        services.AddSingleton<HostsService>();
        services.AddSingleton<BackupService>();
        services.AddTransient<SyncService>();
        services.AddSingleton<DnsService>();
        
        services.AddSingleton<MainWindowViewModel>();
        services.AddSingleton<HostsEditorViewModel>();
        services.AddSingleton<BackupViewModel>();
        services.AddTransient<SyncViewModel>();

        Services = services.BuildServiceProvider();
    }

    private void ShowMainWindow(object? sender, EventArgs e)
    {
        ActivateMainWindow();
    }

    private void ExitApplication(object? sender, EventArgs e)
    {
        if (ApplicationLifetime is IClassicDesktopStyleApplicationLifetime desktop)
        {
            desktop.Shutdown();
        }
    }
}
