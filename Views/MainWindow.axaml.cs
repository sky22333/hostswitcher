using Avalonia.Controls;
using HostsManager.Helpers;

namespace HostsManager.Views;

public partial class MainWindow : Window
{
    public MainWindow()
    {
        InitializeComponent();
        WindowChrome.EnsureOpaqueFallback(this);
        
        Closing += (s, e) =>
        {
            e.Cancel = true;
            Hide();
        };
    }
}
