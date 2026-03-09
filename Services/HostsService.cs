using Serilog;
using System;
using System.IO;
using System.Text;
using System.Threading;
using System.Threading.Tasks;

namespace HostsManager.Services;

public class HostsService
{
    private const string HostsPath = @"C:\Windows\System32\drivers\etc\hosts";
    private readonly SemaphoreSlim _fileLock = new(1, 1);

    public async Task<string> ReadHostsAsync()
    {
        await _fileLock.WaitAsync();
        try
        {
            if (!File.Exists(HostsPath))
            {
                await File.WriteAllTextAsync(HostsPath, "# Hosts file\r\n", Encoding.UTF8);
            }

            return await File.ReadAllTextAsync(HostsPath, Encoding.UTF8);
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
        await _fileLock.WaitAsync();
        try
        {
            await File.WriteAllTextAsync(HostsPath, content, Encoding.UTF8);
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

    public async Task<bool> ValidateHostsContentAsync(string content)
    {
        await Task.CompletedTask;
        
        if (string.IsNullOrWhiteSpace(content))
            return true;

        if (content.Length > 10 * 1024 * 1024)
        {
            return false;
        }

        return true;
    }
}
