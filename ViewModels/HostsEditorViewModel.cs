using CommunityToolkit.Mvvm.ComponentModel;
using CommunityToolkit.Mvvm.Input;
using HostsManager.Helpers;
using HostsManager.Services;
using Serilog;
using System;
using System.Text;
using System.Threading.Tasks;

namespace HostsManager.ViewModels;

public partial class HostsEditorViewModel : ObservableObject
{
    private readonly HostsService _hostsService;
    private readonly BackupService _backupService;
    private readonly DnsService _dnsService;

    [ObservableProperty]
    private string _hostsContent = string.Empty;

    [ObservableProperty]
    private bool _isModified;

    [ObservableProperty]
    private bool _isLoading;

    [ObservableProperty]
    private string _statusMessage = "就绪";
    
    private bool _isInitialLoad = true;
    private bool _isInitialized;
    private string? _lastLoadedHash;

    public HostsEditorViewModel(HostsService hostsService, BackupService backupService, DnsService dnsService)
    {
        _hostsService = hostsService;
        _backupService = backupService;
        _dnsService = dnsService;
    }

    public async Task InitializeAsync()
    {
        if (_isInitialized) return;
        _isInitialized = true;
        await LoadHostsAsync();
    }

    public async Task RefreshIfNeededAsync()
    {
        try
        {
            var currentFileContent = await _hostsService.ReadHostsAsync();
            var currentHash = HashHelper.ComputeHash(currentFileContent);
            
            if (currentHash != _lastLoadedHash)
            {
                await LoadHostsAsync();
            }
        }
        catch (Exception ex)
        {
            Log.Error(ex, "刷新检查失败");
        }
    }

    [RelayCommand]
    private async Task LoadHostsAsync()
    {
        try
        {
            IsLoading = true;
            StatusMessage = "正在加载...";

            var content = await _hostsService.ReadHostsAsync();
            
            _isInitialLoad = true;
            HostsContent = content;
            _lastLoadedHash = HashHelper.ComputeHash(content);
            
            await Task.Delay(50);
            _isInitialLoad = false;
            
            IsModified = false;
            StatusMessage = "加载成功";
        }
        catch (Exception ex)
        {
            StatusMessage = $"加载失败: {ex.Message}";
            Log.Error(ex, "加载失败");
        }
        finally
        {
            IsLoading = false;
        }
    }

    [RelayCommand]
    private async Task SaveHostsAsync()
    {
        try
        {
            IsLoading = true;
            StatusMessage = "正在保存...";

            if (!await _hostsService.ValidateHostsContentAsync(HostsContent))
            {
                StatusMessage = "内容验证失败";
                return;
            }

            await _backupService.CreateBackupAsync(HostsContent);
            await _hostsService.WriteHostsAsync(HostsContent);
            
            _lastLoadedHash = HashHelper.ComputeHash(HostsContent);
            IsModified = false;
            StatusMessage = "保存成功";
        }
        catch (Exception ex)
        {
            StatusMessage = $"保存失败: {ex.Message}";
            Log.Error(ex, "保存失败");
        }
        finally
        {
            IsLoading = false;
        }
    }

    [RelayCommand]
    private async Task FlushDnsAsync()
    {
        try
        {
            IsLoading = true;
            StatusMessage = "正在刷新 DNS 缓存...";

            if (!_dnsService.IsDnsServiceAvailable())
            {
                StatusMessage = "DNS 服务不可用";
                return;
            }

            bool result = await _dnsService.FlushDnsCacheAsync();
            StatusMessage = result ? "DNS 缓存刷新成功" : "DNS 缓存刷新失败";
        }
        catch (Exception ex)
        {
            StatusMessage = $"刷新 DNS 失败: {ex.Message}";
            Log.Error(ex, "DNS 刷新失败");
        }
        finally
        {
            IsLoading = false;
        }
    }

    partial void OnHostsContentChanged(string value)
    {
        if (!_isInitialLoad)
        {
            IsModified = true;
        }
    }
}
