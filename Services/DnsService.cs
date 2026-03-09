using Serilog;
using System;
using System.Runtime.InteropServices;
using System.Threading.Tasks;

namespace HostsManager.Services;

public class DnsService
{
    [DllImport("dnsapi.dll", EntryPoint = "DnsFlushResolverCache")]
    private static extern bool DnsFlushResolverCache();

    public async Task<bool> FlushDnsCacheAsync()
    {
        return await Task.Run(() =>
        {
            try
            {
                return DnsFlushResolverCache();
            }
            catch (Exception ex)
            {
                Log.Error(ex, "DNS 刷新失败");
                return false;
            }
        });
    }

    public bool IsDnsServiceAvailable()
    {
        try
        {
            return RuntimeInformation.IsOSPlatform(OSPlatform.Windows);
        }
        catch
        {
            return false;
        }
    }
}
