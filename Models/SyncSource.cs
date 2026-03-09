using CommunityToolkit.Mvvm.ComponentModel;

namespace HostsManager.Models;

public partial class SyncSource : ObservableObject
{
    [ObservableProperty]
    private string _name = string.Empty;

    [ObservableProperty]
    private string _url = string.Empty;

    [ObservableProperty]
    private string _description = string.Empty;
}
