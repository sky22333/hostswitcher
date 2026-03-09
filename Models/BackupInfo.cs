using System;

namespace HostsManager.Models;

public class BackupInfo
{
    public string FileName { get; set; } = string.Empty;
    public string FilePath { get; set; } = string.Empty;
    public DateTime CreatedTime { get; set; }
    public long FileSize { get; set; }
    public string DisplayName => $"{CreatedTime:yyyy-MM-dd HH:mm:ss} ({FormatFileSize(FileSize)})";

    private static string FormatFileSize(long bytes)
    {
        if (bytes < 1024) return $"{bytes} B";
        if (bytes < 1024 * 1024) return $"{bytes / 1024.0:F2} KB";
        return $"{bytes / (1024.0 * 1024.0):F2} MB";
    }
}
