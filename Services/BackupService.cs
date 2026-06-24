using HostsManager.Helpers;
using HostsManager.Models;
using Serilog;
using System;
using System.Collections.Generic;
using System.IO;
using System.Linq;
using System.Text;
using System.Threading.Tasks;

namespace HostsManager.Services;

public class BackupService
{
    private readonly string _backupDirectory;
    private const int MaxBackupCount = 50;
    private string? _lastBackupHash;
    private string? _lastBackupPath;

    public BackupService()
    {
        _backupDirectory = Path.Combine(
            Environment.GetFolderPath(Environment.SpecialFolder.ApplicationData),
            "HostsManager", "backups");

        Directory.CreateDirectory(_backupDirectory);
    }

    public async Task<string> CreateBackupAsync(string content)
    {
        try
        {
            var contentHash = HashHelper.ComputeHash(content);

            if (_lastBackupHash == contentHash && !string.IsNullOrEmpty(_lastBackupPath) && File.Exists(_lastBackupPath))
                return _lastBackupPath;

            var existingBackups = GetBackups();
            if (existingBackups.Count > 0)
            {
                var latestBackup = existingBackups[0];
                if (latestBackup.FileSize == Encoding.UTF8.GetByteCount(content))
                {
                    var latestContent = await File.ReadAllTextAsync(latestBackup.FilePath, Encoding.UTF8);
                    if (HashHelper.ComputeHash(latestContent) == contentHash)
                    {
                        _lastBackupHash = contentHash;
                        _lastBackupPath = latestBackup.FilePath;
                        return latestBackup.FilePath;
                    }
                }
            }

            var timestamp = DateTime.Now.ToString("yyyyMMdd_HHmmss");
            var filePath = Path.Combine(_backupDirectory, $"hosts_backup_{timestamp}.txt");

            await File.WriteAllTextAsync(filePath, content, Encoding.UTF8);

            _lastBackupHash = contentHash;
            _lastBackupPath = filePath;

            CleanupOldBackups(existingBackups.Count + 1);

            return filePath;
        }
        catch (Exception ex)
        {
            Log.Error(ex, "创建备份失败");
            throw;
        }
    }

    public Task<List<BackupInfo>> GetBackupsAsync()
    {
        return Task.FromResult(GetBackups());
    }

    public async Task<string> ReadBackupAsync(string backupPath)
    {
        try
        {
            if (!File.Exists(backupPath))
                throw new FileNotFoundException("备份文件不存在", backupPath);

            return await File.ReadAllTextAsync(backupPath, Encoding.UTF8);
        }
        catch (Exception ex)
        {
            Log.Error(ex, "读取备份失败");
            throw;
        }
    }

    public async Task<string> PrepareRestoreAsync(string backupPath)
    {
        var content = await ReadBackupAsync(backupPath);
        _lastBackupHash = HashHelper.ComputeHash(content);
        _lastBackupPath = backupPath;
        return content;
    }

    public Task DeleteBackupAsync(string backupPath)
    {
        try
        {
            if (File.Exists(backupPath))
            {
                File.Delete(backupPath);

                if (_lastBackupPath == backupPath)
                {
                    _lastBackupHash = null;
                    _lastBackupPath = null;
                }
            }
        }
        catch (Exception ex)
        {
            Log.Error(ex, "删除备份失败");
            throw;
        }

        return Task.CompletedTask;
    }

    private List<BackupInfo> GetBackups()
    {
        var backups = new List<BackupInfo>();

        foreach (var file in Directory.EnumerateFiles(_backupDirectory, "hosts_backup_*.txt"))
        {
            var fileInfo = new FileInfo(file);
            backups.Add(new BackupInfo
            {
                FileName = fileInfo.Name,
                FilePath = file,
                CreatedTime = fileInfo.CreationTime,
                FileSize = fileInfo.Length
            });
        }

        return backups.OrderByDescending(b => b.CreatedTime).ToList();
    }

    private void CleanupOldBackups(int totalCount)
    {
        if (totalCount <= MaxBackupCount)
            return;

        var backups = GetBackups();
        foreach (var backup in backups.Skip(MaxBackupCount))
            File.Delete(backup.FilePath);
    }
}
