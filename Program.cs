using Avalonia;
using System;
using System.Diagnostics;
using System.IO;
using System.IO.Pipes;
using System.Security.Principal;
using System.Text;
using System.Threading;
using System.Threading.Tasks;
using Serilog;

namespace HostsManager;

class Program
{
    private static Mutex? _mutex;
    private const string MutexName = "HostsManager_SingleInstance_Mutex";
    private const string PipeName = "HostsManager_IPC_Pipe";
    private static CancellationTokenSource? _pipeCancellation;

    [STAThread]
    public static void Main(string[] args)
    {
        ConfigureLogging();
        
        bool createdNew;
        _mutex = new Mutex(true, MutexName, out createdNew);
        
        if (!createdNew)
        {
            // 程序已在运行，发送激活信号
            SendActivationSignal();
            return;
        }
        
        if (!IsAdministrator())
        {
            RestartAsAdministrator();
            _mutex?.ReleaseMutex();
            _mutex?.Dispose();
            return;
        }

        StartPipeServer();

        try
        {
            BuildAvaloniaApp().StartWithClassicDesktopLifetime(args);
        }
        catch (Exception ex)
        {
            Log.Fatal(ex, "应用程序启动失败");
            throw;
        }
        finally
        {
            _pipeCancellation?.Cancel();
            _pipeCancellation?.Dispose();
            _mutex?.ReleaseMutex();
            _mutex?.Dispose();
            Log.CloseAndFlush();
        }
    }

    public static AppBuilder BuildAvaloniaApp()
        => AppBuilder.Configure<App>()
            .UsePlatformDetect()
            .WithInterFont()
            .LogToTrace();

    private static bool IsAdministrator()
    {
        using var identity = WindowsIdentity.GetCurrent();
        var principal = new WindowsPrincipal(identity);
        return principal.IsInRole(WindowsBuiltInRole.Administrator);
    }

    private static void StartPipeServer()
    {
        _pipeCancellation = new CancellationTokenSource();
        
        Task.Run(async () =>
        {
            while (!_pipeCancellation.Token.IsCancellationRequested)
            {
                NamedPipeServerStream? pipeServer = null;
                try
                {
                    pipeServer = new NamedPipeServerStream(
                        PipeName,
                        PipeDirection.In,
                        1,
                        PipeTransmissionMode.Byte,
                        PipeOptions.Asynchronous);

                    await pipeServer.WaitForConnectionAsync(_pipeCancellation.Token);

                    var buffer = new byte[256];
                    var bytesRead = await pipeServer.ReadAsync(buffer, 0, buffer.Length, _pipeCancellation.Token);
                    var message = Encoding.UTF8.GetString(buffer, 0, bytesRead);

                    if (message == "ACTIVATE")
                    {
                        App.ActivateMainWindow();
                    }

                    pipeServer.Disconnect();
                }
                catch (OperationCanceledException)
                {
                    break;
                }
                catch (Exception ex)
                {
                    Log.Error(ex, "命名管道服务器错误");
                }
                finally
                {
                    pipeServer?.Dispose();
                }
            }
        }, _pipeCancellation.Token);
    }

    private static void SendActivationSignal()
    {
        try
        {
            using var pipeClient = new NamedPipeClientStream(".", PipeName, PipeDirection.Out);
            pipeClient.Connect(1000); // 1秒超时

            var message = Encoding.UTF8.GetBytes("ACTIVATE");
            pipeClient.Write(message, 0, message.Length);
            pipeClient.Flush();
        }
        catch (Exception ex)
        {
            Log.Error(ex, "发送激活信号失败");
        }
    }

    private static void RestartAsAdministrator()
    {
        var processInfo = new ProcessStartInfo
        {
            UseShellExecute = true,
            WorkingDirectory = Environment.CurrentDirectory,
            FileName = Environment.ProcessPath!,
            Verb = "runas"
        };

        try
        {
            Process.Start(processInfo);
        }
        catch (Exception ex)
        {
            Log.Error(ex, "无法以管理员权限重启应用程序");
        }
    }

    private static void ConfigureLogging()
    {
        var logPath = Path.Combine(
            Environment.GetFolderPath(Environment.SpecialFolder.ApplicationData),
            "HostsManager", "logs", "app-.log");

        Log.Logger = new LoggerConfiguration()
            .MinimumLevel.Information()
            .WriteTo.File(
                logPath,
                rollingInterval: RollingInterval.Day,
                retainedFileCountLimit: 7,
                fileSizeLimitBytes: 10_485_760)
            .CreateLogger();
        
        // 异步清理旧日志文件
        _ = Task.Run(() => CleanupOldLogsAsync());
    }

    private static async Task CleanupOldLogsAsync()
    {
        try
        {
            var logsDirectory = Path.Combine(
                Environment.GetFolderPath(Environment.SpecialFolder.ApplicationData),
                "HostsManager", "logs");

            if (!Directory.Exists(logsDirectory))
                return;

            var cutoffDate = DateTime.Now.AddDays(-3);
            var deletedCount = 0;

            foreach (var file in Directory.EnumerateFiles(logsDirectory, "app-*.log"))
            {
                try
                {
                    var fileInfo = new FileInfo(file);
                    if (fileInfo.LastWriteTime < cutoffDate)
                    {
                        await Task.Run(() => File.Delete(file));
                        deletedCount++;
                    }
                }
                catch
                {
                    // 忽略单个文件删除失败
                }
            }
        }
        catch (Exception ex)
        {
            Log.Error(ex, "清理旧日志失败");
        }
    }
}
