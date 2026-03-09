using Avalonia.Controls;
using Avalonia.Input;
using Avalonia.Interactivity;
using HostsManager.ViewModels;
using Serilog;
using System;
using System.Diagnostics;
using System.IO;

namespace HostsManager.Views;

public partial class HostsEditorView : UserControl
{
    public HostsEditorView()
    {
        InitializeComponent();
        DataContextChanged += OnDataContextChanged;
        Unloaded += OnUnloaded;
    }

    private void OnDataContextChanged(object? sender, EventArgs e)
    {
        if (DataContext is HostsEditorViewModel viewModel)
        {
            _ = viewModel.InitializeAsync();
        }
    }

    private void OnUnloaded(object? sender, RoutedEventArgs e)
    {
        DataContextChanged -= OnDataContextChanged;
        Unloaded -= OnUnloaded;
    }

    private void OnHostsPathClicked(object? sender, PointerPressedEventArgs e)
    {
        try
        {
            var hostsPath = @"C:\Windows\System32\drivers\etc";
            if (Directory.Exists(hostsPath))
            {
                Process.Start(new ProcessStartInfo
                {
                    FileName = "explorer.exe",
                    Arguments = hostsPath,
                    UseShellExecute = true
                });
            }
        }
        catch (Exception ex)
        {
            Log.Error(ex, "打开目录失败");
        }
    }
}
