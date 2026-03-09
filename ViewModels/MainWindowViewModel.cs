using CommunityToolkit.Mvvm.ComponentModel;
using CommunityToolkit.Mvvm.Input;

namespace HostsManager.ViewModels;

public partial class MainWindowViewModel : ObservableObject
{
    [ObservableProperty]
    private object? _currentView;

    [ObservableProperty]
    private int _selectedIndex;

    private readonly HostsEditorViewModel _hostsEditorViewModel;
    private readonly BackupViewModel _backupViewModel;
    private readonly SyncViewModel _syncViewModel;

    public MainWindowViewModel(
        HostsEditorViewModel hostsEditorViewModel,
        BackupViewModel backupViewModel,
        SyncViewModel syncViewModel)
    {
        _hostsEditorViewModel = hostsEditorViewModel;
        _backupViewModel = backupViewModel;
        _syncViewModel = syncViewModel;

        CurrentView = _hostsEditorViewModel;
    }

    async partial void OnSelectedIndexChanged(int value)
    {
        var previousView = CurrentView;
        
        CurrentView = value switch
        {
            0 => _hostsEditorViewModel,
            1 => _backupViewModel,
            2 => _syncViewModel,
            _ => _hostsEditorViewModel
        };

        if (CurrentView == _hostsEditorViewModel && previousView != _hostsEditorViewModel)
        {
            await _hostsEditorViewModel.RefreshIfNeededAsync();
        }
    }

    [RelayCommand]
    private void NavigateToEditor() => SelectedIndex = 0;

    [RelayCommand]
    private void NavigateToBackup() => SelectedIndex = 1;

    [RelayCommand]
    private void NavigateToSync() => SelectedIndex = 2;
}
