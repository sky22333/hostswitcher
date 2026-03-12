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
            
            // 检查最近一次备份
            if (_lastBackupHash == contentHash && !string.IsNullOrEmpty(_lastBackupPath) && File.Exists(_lastBackupPath))
            {
                return _lastBackupPath;
            }

            // 检查所有现有备份
            var existingBackups = await GetBackupsAsync();
            if (existingBackups.Count > 0)
            {
                // 读取最新的备份文件进行比较
                var latestBackup = existingBackups.First();
                if (latestBackup.FileSize == Encoding.UTF8.GetByteCount(content)) // 简单的长度预检查
                {
                    var latestContent = await File.ReadAllTextAsync(latestBackup.FilePath, Encoding.UTF8);
                    var latestHash = HashHelper.ComputeHash(latestContent);
                    
                    if (latestHash == contentHash)
                    {
                        _lastBackupHash = contentHash;
                        _lastBackupPath = latestBackup.FilePath;
                        return latestBackup.FilePath;
                    }
                }
            }

            var timestamp = DateTime.Now.ToString("yyyyMMdd_HHmmss");
            var fileName = $"hosts_backup_{timestamp}.txt";
            var filePath = Path.Combine(_backupDirectory, fileName);

            await File.WriteAllTextAsync(filePath, content, Encoding.UTF8);
            
            _lastBackupHash = contentHash;
            _lastBackupPath = filePath;
            
            await CleanupOldBackupsAsync();

            return filePath;
        }
        catch (Exception ex)
        {
            Log.Error(ex, "创建备份失败");
            throw;
        }
    }

    public async Task<List<BackupInfo>> GetBackupsAsync()
    {
        await Task.CompletedTask;
        
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

    public async Task<string> RestoreBackupAsync(string backupPath)
    {
        try
        {
            if (!File.Exists(backupPath))
            {
                throw new FileNotFoundException("备份文件不存在", backupPath);
            }

            var content = await File.ReadAllTextAsync(backupPath, Encoding.UTF8);
            
            _lastBackupHash = HashHelper.ComputeHash(content);
            _lastBackupPath = backupPath;
            
            return content;
        }
        catch (Exception ex)
        {
            Log.Error(ex, "恢复备份失败");
            throw;
        }
    }

    public async Task DeleteBackupAsync(string backupPath)
    {
        await Task.CompletedTask;
        
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
    }

    private async Task CleanupOldBackupsAsync()
    {
        var backups = await GetBackupsAsync();
        
        if (backups.Count > MaxBackupCount)
        {
            var toDelete = backups.Skip(MaxBackupCount).ToList();
            foreach (var backup in toDelete)
            {
                await DeleteBackupAsync(backup.FilePath);
            }
        }
    }
}
