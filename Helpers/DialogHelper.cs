using Avalonia;
using Avalonia.Controls;
using Avalonia.Controls.ApplicationLifetimes;
using Avalonia.Controls.Primitives;
using Avalonia.Layout;
using Avalonia.Media;
using System.Threading.Tasks;

namespace HostsManager.Helpers;

public static class DialogHelper
{
    private static Window? GetMainWindow()
    {
        if (Application.Current?.ApplicationLifetime is IClassicDesktopStyleApplicationLifetime desktop)
            return desktop.MainWindow;
        return null;
    }

    public static async Task<bool> ConfirmAsync(string title, string message)
    {
        var owner = GetMainWindow();
        if (owner == null)
            return false;

        var dialog = CreateDialogWindow(title, 420, resizable: false);

        var cancelButton = CreateDialogButton("取消");
        cancelButton.Click += (_, _) => dialog.Close(false);

        var confirmButton = CreateDialogButton("确定");
        confirmButton.Click += (_, _) => dialog.Close(true);

        dialog.Content = new Border
        {
            Padding = new Thickness(24),
            Child = new StackPanel
            {
                Spacing = 20,
                Children =
                {
                    new TextBlock
                    {
                        Text = message,
                        TextWrapping = TextWrapping.Wrap
                    },
                    new StackPanel
                    {
                        Orientation = Orientation.Horizontal,
                        Spacing = 8,
                        HorizontalAlignment = HorizontalAlignment.Right,
                        Children = { cancelButton, confirmButton }
                    }
                }
            }
        };

        return await dialog.ShowDialog<bool>(owner);
    }

    public static async Task ShowPreviewAsync(string title, string content)
    {
        var owner = GetMainWindow();
        if (owner == null)
            return;

        var dialog = CreateDialogWindow(title, 720, resizable: true, height: 520);

        var editor = new TextBox
        {
            Text = content,
            IsReadOnly = true,
            Classes = { "editor" }
        };
        ScrollViewer.SetHorizontalScrollBarVisibility(editor, ScrollBarVisibility.Auto);
        ScrollViewer.SetVerticalScrollBarVisibility(editor, ScrollBarVisibility.Auto);

        var closeButton = CreateDialogButton("关闭");
        closeButton.Click += (_, _) => dialog.Close();

        var root = new Grid
        {
            RowDefinitions = new RowDefinitions("*,Auto"),
            Margin = new Thickness(16),
            RowSpacing = 12
        };

        var card = new Border { Classes = { "card" }, Child = editor };
        Grid.SetRow(card, 0);

        var buttonRow = new StackPanel
        {
            Orientation = Orientation.Horizontal,
            HorizontalAlignment = HorizontalAlignment.Right,
            Children = { closeButton }
        };
        Grid.SetRow(buttonRow, 1);

        root.Children.Add(card);
        root.Children.Add(buttonRow);

        dialog.Content = root;
        await dialog.ShowDialog(owner);
    }

    private static Window CreateDialogWindow(string title, double width, bool resizable, double? height = null)
    {
        var window = new Window
        {
            Title = title,
            Width = width,
            WindowStartupLocation = WindowStartupLocation.CenterOwner,
            CanResize = resizable,
            ShowInTaskbar = false
        };

        if (resizable && height.HasValue)
        {
            window.Height = height.Value;
            window.SizeToContent = SizeToContent.Manual;
        }
        else
        {
            window.SizeToContent = SizeToContent.Height;
        }

        return window;
    }

    private static Button CreateDialogButton(string text)
    {
        return new Button
        {
            Classes = { "dialog" },
            Content = text,
            MinWidth = 96,
            HorizontalAlignment = HorizontalAlignment.Left,
            VerticalAlignment = VerticalAlignment.Center
        };
    }
}
