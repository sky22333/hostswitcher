using Serilog;
using System;
using System.Net.Http;
using System.Text;
using System.Threading.Tasks;

namespace HostsManager.Services;

public class SyncService : IDisposable
{
    private readonly HttpClient _httpClient;
    private bool _disposed;

    public SyncService()
    {
        _httpClient = new HttpClient
        {
            Timeout = TimeSpan.FromSeconds(30)
        };
        _httpClient.DefaultRequestHeaders.Add("User-Agent", "HostsManager/1.0");
    }

    public async Task<string> DownloadHostsAsync(string url)
    {
        ObjectDisposedException.ThrowIf(_disposed, this);
        
        try
        {
            var response = await _httpClient.GetAsync(url);
            response.EnsureSuccessStatusCode();

            var content = await response.Content.ReadAsStringAsync();
            
            if (content.Length > 5 * 1024 * 1024)
            {
                throw new InvalidOperationException("远程 Hosts 文件过大");
            }

            return content;
        }
        catch (Exception ex)
        {
            Log.Error(ex, "下载远程 Hosts 失败");
            throw;
        }
    }

    public string MergeHosts(string currentHosts, string remoteHosts, bool appendMode)
    {
        if (appendMode)
        {
            const string marker = "# === 远程同步内容 ===";
            var markerIndex = currentHosts.IndexOf(marker, StringComparison.Ordinal);
            
            if (markerIndex >= 0)
            {
                currentHosts = currentHosts.Substring(0, markerIndex).TrimEnd();
            }
            
            var sb = new StringBuilder(currentHosts.Length + remoteHosts.Length + 50);
            sb.Append(currentHosts);
            sb.Append("\r\n\r\n");
            sb.Append(marker);
            sb.Append("\r\n");
            sb.Append(remoteHosts);
            return sb.ToString();
        }
        else
        {
            return remoteHosts;
        }
    }

    public void Dispose()
    {
        if (_disposed) return;
        
        _httpClient?.Dispose();
        _disposed = true;
        GC.SuppressFinalize(this);
    }
}
