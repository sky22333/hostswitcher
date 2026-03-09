using CommunityToolkit.Mvvm.ComponentModel;
using CommunityToolkit.Mvvm.Input;
using HostsManager.Models;
using HostsManager.Services;
using Serilog;
using System;
using System.Collections.ObjectModel;
using System.Threading.Tasks;

namespace HostsManager.ViewModels;

public partial class SyncViewModel : ObservableObject, IDisposable
{
    private readonly SyncService _syncService;
    private readonly HostsService _hostsService;
    private readonly BackupService _backupService;
    private bool _disposed;

    [ObservableProperty]
    private ObservableCollection<SyncSource> _syncSources = new();

    [ObservableProperty]
    private SyncSource? _selectedSource;

    [ObservableProperty]
    private bool _appendMode = true;

    [ObservableProperty]
    private bool _isLoading;

    [ObservableProperty]
    private string _statusMessage = "就绪";

    [ObservableProperty]
    private string _newSourceName = string.Empty;

    [ObservableProperty]
    private string _newSourceUrl = string.Empty;

    public SyncViewModel(SyncService syncService, HostsService hostsService, BackupService backupService)
    {
        _syncService = syncService;
        _hostsService = hostsService;
        _backupService = backupService;

        LoadDefaultSources();
    }

    private void LoadDefaultSources()
    {
        SyncSources.Add(new SyncSource
        {
            Name = "GitHub Hosts",
            Url = "https://gitee.com/if-the-wind/github-hosts/raw/main/hosts",
            Description = "GitHub 加速"
        });
    }

    [RelayCommand]
    private async Task SyncFromSourceAsync()
    {
        if (SelectedSource == null)
        {
            StatusMessage = "请选择同步源";
            return;
        }

        try
        {
            IsLoading = true;
            StatusMessage = "正在下载远程 Hosts...";

            var remoteHosts = await _syncService.DownloadHostsAsync(SelectedSource.Url);
            
            StatusMessage = "正在备份当前 Hosts...";
            var currentHosts = await _hostsService.ReadHostsAsync();
            await _backupService.CreateBackupAsync(currentHosts);

            StatusMessage = "正在合并内容...";
            var mergedHosts = _syncService.MergeHosts(currentHosts, remoteHosts, AppendMode);

            StatusMessage = "正在保存...";
            await _hostsService.WriteHostsAsync(mergedHosts);

            StatusMessage = "同步成功";
        }
        catch (Exception ex)
        {
            StatusMessage = $"同步失败: {ex.Message}";
            Log.Error(ex, "同步 Hosts 失败");
        }
        finally
        {
            IsLoading = false;
        }
    }

    [RelayCommand]
    private void AddSource()
    {
        if (string.IsNullOrWhiteSpace(NewSourceName) || string.IsNullOrWhiteSpace(NewSourceUrl))
        {
            StatusMessage = "请输入名称和 URL";
            return;
        }

        SyncSources.Add(new SyncSource
        {
            Name = NewSourceName,
            Url = NewSourceUrl
        });

        NewSourceName = string.Empty;
        NewSourceUrl = string.Empty;
        StatusMessage = "添加成功";
    }

    [RelayCommand]
    private void RemoveSource()
    {
        if (SelectedSource != null)
        {
            SyncSources.Remove(SelectedSource);
            StatusMessage = "删除成功";
        }
    }

    public void Dispose()
    {
        if (_disposed) return;
        
        _syncService?.Dispose();
        _disposed = true;
        GC.SuppressFinalize(this);
    }
}
