using CommunityToolkit.Mvvm.ComponentModel;
using CommunityToolkit.Mvvm.Input;
using HostsManager.Helpers;
using HostsManager.Services;
using Serilog;
using System;
using System.Threading.Tasks;

namespace HostsManager.ViewModels;

public partial class SyncViewModel : ObservableObject
{
    public const string DefaultSyncUrl = "https://gitee.com/if-the-wind/github-hosts/raw/main/hosts";

    private readonly SyncService _syncService;
    private readonly HostsService _hostsService;
    private readonly BackupService _backupService;
    private readonly HostsEditorViewModel _hostsEditorViewModel;

    [ObservableProperty]
    private string _syncUrl = DefaultSyncUrl;

    [ObservableProperty]
    private bool _appendMode = true;

    [ObservableProperty]
    private bool _isLoading;

    [ObservableProperty]
    private string _statusMessage = "就绪";

    public SyncViewModel(
        SyncService syncService,
        HostsService hostsService,
        BackupService backupService,
        HostsEditorViewModel hostsEditorViewModel)
    {
        _syncService = syncService;
        _hostsService = hostsService;
        _backupService = backupService;
        _hostsEditorViewModel = hostsEditorViewModel;
    }

    [RelayCommand]
    private async Task SyncFromSourceAsync()
    {
        if (string.IsNullOrWhiteSpace(SyncUrl))
        {
            StatusMessage = "请输入同步 URL";
            return;
        }

        if (!Uri.TryCreate(SyncUrl, UriKind.Absolute, out var uri) ||
            uri.Scheme is not ("http" or "https"))
        {
            StatusMessage = "URL 格式无效";
            return;
        }

        if (!AppendMode && !await DialogHelper.ConfirmAsync(
                "确认覆盖",
                "覆盖模式将完全替换当前 Hosts 内容。\n同步前会自动创建备份，是否继续？"))
            return;

        try
        {
            IsLoading = true;
            StatusMessage = "正在下载远程 Hosts...";

            var remoteHosts = await _syncService.DownloadHostsAsync(SyncUrl);

            StatusMessage = "正在备份当前 Hosts...";
            var currentHosts = await _hostsService.ReadHostsAsync();
            await _backupService.CreateBackupAsync(currentHosts);

            StatusMessage = "正在合并内容...";
            var mergedHosts = _syncService.MergeHosts(currentHosts, remoteHosts, AppendMode);

            StatusMessage = "正在保存...";
            await _hostsService.WriteHostsAsync(mergedHosts);
            await _hostsEditorViewModel.RefreshIfNeededAsync();

            StatusMessage = "同步成功，可在编辑器中查看";
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
}
