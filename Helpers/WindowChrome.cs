using Avalonia;
using Avalonia.Controls;
using Avalonia.Media;

namespace HostsManager.Helpers;

public static class WindowChrome
{
    public static void EnsureOpaqueFallback(Window window)
    {
        window.Opened += OnOpened;

        void OnOpened(object? sender, System.EventArgs e)
        {
            window.Opened -= OnOpened;
            if (window.ActualTransparencyLevel != WindowTransparencyLevel.None)
                return;

            if (Application.Current?.TryFindResource(
                    "SystemControlBackgroundChromeMediumBrush", out var resource) == true
                && resource is IBrush brush)
            {
                window.Background = brush;
            }
        }
    }
}
