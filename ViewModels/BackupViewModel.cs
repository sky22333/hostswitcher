using CommunityToolkit.Mvvm.ComponentModel;
using CommunityToolkit.Mvvm.Input;
using HostsManager.Models;
using HostsManager.Services;
using Serilog;
using System;
using System.Collections.ObjectModel;
using System.Diagnostics;
using System.Threading.Tasks;

namespace HostsManager.ViewModels;

public partial class BackupViewModel : ObservableObject
{
    private readonly BackupService _backupService;
    private readonly HostsService _hostsService;

    [ObservableProperty]
    private ObservableCollection<BackupInfo> _backups = new();

    [ObservableProperty]
    private BackupInfo? _selectedBackup;

    [ObservableProperty]
    private bool _isLoading;

    [ObservableProperty]
    private string _statusMessage = "就绪";

    public BackupViewModel(BackupService backupService, HostsService hostsService)
    {
        _backupService = backupService;
        _hostsService = hostsService;
        _ = LoadBackupsAsync();
    }

    [RelayCommand]
    private async Task LoadBackupsAsync()
    {
        try
        {
            IsLoading = true;
            StatusMessage = "正在加载备份列表...";

            var backups = await _backupService.GetBackupsAsync();
            Backups.Clear();
            foreach (var backup in backups)
            {
                Backups.Add(backup);
            }

            StatusMessage = $"找到 {Backups.Count} 个备份";
        }
        catch (Exception ex)
        {
            StatusMessage = $"加载失败: {ex.Message}";
            Log.Error(ex, "加载备份失败");
        }
        finally
        {
            IsLoading = false;
        }
    }

    [RelayCommand]
    private async Task CreateBackupAsync()
    {
        try
        {
            IsLoading = true;
            StatusMessage = "正在创建备份...";

            var currentHosts = await _hostsService.ReadHostsAsync();
            await _backupService.CreateBackupAsync(currentHosts);
            await LoadBackupsAsync();

            StatusMessage = "备份创建成功";
        }
        catch (Exception ex)
        {
            StatusMessage = $"创建备份失败: {ex.Message}";
            Log.Error(ex, "创建备份失败");
        }
        finally
        {
            IsLoading = false;
        }
    }

    [RelayCommand]
    private async Task RestoreBackupAsync()
    {
        if (SelectedBackup == null)
        {
            StatusMessage = "请选择要恢复的备份";
            return;
        }

        try
        {
            IsLoading = true;
            StatusMessage = "正在恢复备份...";

            var currentHosts = await _hostsService.ReadHostsAsync();
            await _backupService.CreateBackupAsync(currentHosts);

            var content = await _backupService.RestoreBackupAsync(SelectedBackup.FilePath);
            await _hostsService.WriteHostsAsync(content);

            StatusMessage = "恢复成功";
        }
        catch (Exception ex)
        {
            StatusMessage = $"恢复失败: {ex.Message}";
            Log.Error(ex, "恢复备份失败");
        }
        finally
        {
            IsLoading = false;
        }
    }

    [RelayCommand]
    private async Task DeleteBackupAsync()
    {
        if (SelectedBackup == null)
        {
            StatusMessage = "请选择要删除的备份";
            return;
        }

        try
        {
            IsLoading = true;
            StatusMessage = "正在删除备份...";

            await _backupService.DeleteBackupAsync(SelectedBackup.FilePath);
            await LoadBackupsAsync();

            StatusMessage = "删除成功";
        }
        catch (Exception ex)
        {
            StatusMessage = $"删除失败: {ex.Message}";
            Log.Error(ex, "删除备份失败");
        }
        finally
        {
            IsLoading = false;
        }
    }

    [RelayCommand]
    private void PreviewBackup()
    {
        if (SelectedBackup == null)
        {
            StatusMessage = "请选择要预览的备份";
            return;
        }

        try
        {
            Process.Start(new ProcessStartInfo
            {
                FileName = SelectedBackup.FilePath,
                UseShellExecute = true
            });
            
            StatusMessage = "已打开预览";
        }
        catch (Exception ex)
        {
            StatusMessage = $"预览失败: {ex.Message}";
            Log.Error(ex, "预览备份失败");
        }
    }
}
