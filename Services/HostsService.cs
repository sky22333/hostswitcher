using Serilog;
using System;
using System.IO;
using System.Text;
using System.Threading;
using System.Threading.Tasks;

namespace HostsManager.Services;

public class HostsService
{
    public const string HostsFilePath = @"C:\Windows\System32\drivers\etc\hosts";
    public static readonly string HostsDirectory = Path.GetDirectoryName(HostsFilePath)!;

    private readonly SemaphoreSlim _fileLock = new(1, 1);

    public async Task<string> ReadHostsAsync()
    {
        await _fileLock.WaitAsync();
        try
        {
            if (!File.Exists(HostsFilePath))
            {
                await File.WriteAllTextAsync(HostsFilePath, "# Hosts file\r\n", Encoding.UTF8);
            }

            return await File.ReadAllTextAsync(HostsFilePath, Encoding.UTF8);
        }
        catch (Exception ex)
        {
            Log.Error(ex, "读取 Hosts 失败");
            throw;
        }
        finally
        {
            _fileLock.Release();
        }
    }

    public async Task WriteHostsAsync(string content)
    {
        if (content.Length > 10 * 1024 * 1024)
            throw new InvalidOperationException("Hosts 文件内容过大");

        await _fileLock.WaitAsync();
        try
        {
            await File.WriteAllTextAsync(HostsFilePath, content, Encoding.UTF8);
        }
        catch (Exception ex)
        {
            Log.Error(ex, "写入 Hosts 失败");
            throw;
        }
        finally
        {
            _fileLock.Release();
        }
    }
}
