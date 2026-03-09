using Avalonia.Controls;

namespace HostsManager.Views;

public partial class MainWindow : Window
{
    public MainWindow()
    {
        InitializeComponent();
        
        Closing += (s, e) =>
        {
            e.Cancel = true;
            Hide();
        };
    }
}
