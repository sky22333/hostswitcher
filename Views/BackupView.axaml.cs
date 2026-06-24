using Avalonia.Controls;
using Avalonia.Input;
using HostsManager.ViewModels;

namespace HostsManager.Views;

public partial class BackupView : UserControl
{
    public BackupView()
    {
        InitializeComponent();
    }

    private async void OnBackupDoubleTapped(object? sender, TappedEventArgs e)
    {
        if (DataContext is BackupViewModel viewModel && viewModel.PreviewBackupCommand.CanExecute(null))
            await viewModel.PreviewBackupCommand.ExecuteAsync(null);
    }
}
