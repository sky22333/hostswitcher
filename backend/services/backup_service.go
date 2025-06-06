package services

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/google/uuid"
	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"

	"hostswitcher/backend/models"
)

// BackupService 处理备份相关操作的服务
type BackupService struct {
	ctx        context.Context
	appDir     string
	backupDir  string
	backupFile string
	maxBackups int
}

// NewBackupService 创建一个新的备份服务实例
func NewBackupService(appDir string) *BackupService {
	backupDir := filepath.Join(appDir, "backups")
	backupFile := filepath.Join(appDir, "backups.json")
	
	// 创建应用目录和备份目录
	err := os.MkdirAll(appDir, 0755)
	if err != nil {
		fmt.Printf("创建应用目录失败: %v\n", err)
	}
	
	err = os.MkdirAll(backupDir, 0755)
	if err != nil {
		fmt.Printf("创建备份目录失败: %v\n", err)
	}

	return &BackupService{
		appDir:     appDir,
		backupDir:  backupDir,
		backupFile: backupFile,
		maxBackups: 10, // 最多保留10个自动备份
	}
}

// SetContext 设置上下文
func (s *BackupService) SetContext(ctx context.Context) {
	s.ctx = ctx
}

// calculateHash 计算内容的MD5哈希
func (s *BackupService) calculateHash(content string) string {
	hash := md5.Sum([]byte(content))
	return hex.EncodeToString(hash[:])
}

// loadBackups 从文件加载备份列表
func (s *BackupService) loadBackups() ([]*models.Backup, error) {
	if _, err := os.Stat(s.backupFile); os.IsNotExist(err) {
		return []*models.Backup{}, nil
	}

	data, err := os.ReadFile(s.backupFile)
	if err != nil {
		return nil, err
	}

	var backupList models.BackupList
	err = json.Unmarshal(data, &backupList)
	if err != nil {
		return nil, err
	}

	if backupList.Backups == nil {
		return []*models.Backup{}, nil
	}

	return backupList.Backups, nil
}

// saveBackups 保存备份列表到文件
func (s *BackupService) saveBackups(backups []*models.Backup) error {
	backupList := models.BackupList{Backups: backups}
	
	data, err := json.MarshalIndent(backupList, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(s.backupFile, data, 0644)
}

// CreateBackup 创建备份
func (s *BackupService) CreateBackup(content, description string, isAutomatic bool, tags []string) (*models.Backup, error) {
	hash := s.calculateHash(content)
	
	// 如果是自动备份，检查是否已存在相同内容的备份
	if isAutomatic {
		backups, err := s.loadBackups()
		if err == nil {
			for _, backup := range backups {
				if backup.Hash == hash && backup.IsAutomatic {
					// 相同内容的自动备份已存在，不创建新备份
					return nil, nil
				}
			}
		}
	}

	backup := &models.Backup{
		ID:          uuid.New().String(),
		Timestamp:   models.JSONTime{Time: time.Now()},
		Description: description,
		Content:     content,
		Size:        int64(len(content)),
		IsAutomatic: isAutomatic,
		Tags:        tags,
		Hash:        hash,
	}

	// 加载现有备份
	backups, err := s.loadBackups()
	if err != nil {
		return nil, err
	}

	// 添加新备份
	backups = append(backups, backup)

	// 如果是自动备份，清理旧的自动备份
	if isAutomatic {
		backups = s.cleanupAutoBackups(backups)
	}

	// 保存备份列表
	err = s.saveBackups(backups)
	if err != nil {
		return nil, err
	}

	// 发出事件通知
	if s.ctx != nil {
		wailsRuntime.EventsEmit(s.ctx, "backup-created", backup.ID)
	}

	return backup, nil
}

// cleanupAutoBackups 清理旧的自动备份，只保留最近的maxBackups个
func (s *BackupService) cleanupAutoBackups(backups []*models.Backup) []*models.Backup {
	// 分离自动备份和手动备份
	var autoBackups []*models.Backup
	var manualBackups []*models.Backup

	for _, backup := range backups {
		if backup.IsAutomatic {
			autoBackups = append(autoBackups, backup)
		} else {
			manualBackups = append(manualBackups, backup)
		}
	}

	// 按时间排序自动备份（最新的在前）
	sort.Slice(autoBackups, func(i, j int) bool {
		return autoBackups[i].Timestamp.Time.After(autoBackups[j].Timestamp.Time)
	})

	// 只保留最近的maxBackups个自动备份
	if len(autoBackups) > s.maxBackups {
		autoBackups = autoBackups[:s.maxBackups]
	}

	// 合并自动备份和手动备份
	result := make([]*models.Backup, 0, len(autoBackups)+len(manualBackups))
	result = append(result, autoBackups...)
	result = append(result, manualBackups...)

	return result
}

// GetAllBackups 获取所有备份
func (s *BackupService) GetAllBackups() ([]*models.Backup, error) {
	backups, err := s.loadBackups()
	if err != nil {
		return nil, err
	}

	// 按时间排序（最新的在前）
	sort.Slice(backups, func(i, j int) bool {
		return backups[i].Timestamp.Time.After(backups[j].Timestamp.Time)
	})

	return backups, nil
}

// GetBackupByID 根据ID获取备份
func (s *BackupService) GetBackupByID(id string) (*models.Backup, error) {
	backups, err := s.loadBackups()
	if err != nil {
		return nil, err
	}

	for _, backup := range backups {
		if backup.ID == id {
			return backup, nil
		}
	}

	return nil, fmt.Errorf("备份不存在: %s", id)
}

// DeleteBackup 删除备份
func (s *BackupService) DeleteBackup(id string) error {
	backups, err := s.loadBackups()
	if err != nil {
		return err
	}

	// 查找并删除备份
	for i, backup := range backups {
		if backup.ID == id {
			// 不允许删除自动备份（保护机制）
			if backup.IsAutomatic {
				return fmt.Errorf("不能删除自动备份")
			}

			backups = append(backups[:i], backups[i+1:]...)
			
			// 保存更新后的备份列表
			err = s.saveBackups(backups)
			if err != nil {
				return err
			}

			// 发出事件通知
			if s.ctx != nil {
				wailsRuntime.EventsEmit(s.ctx, "backup-deleted", id)
			}

			return nil
		}
	}

	return fmt.Errorf("备份不存在: %s", id)
}

// UpdateBackupTags 更新备份标签
func (s *BackupService) UpdateBackupTags(id string, tags []string) error {
	backups, err := s.loadBackups()
	if err != nil {
		return err
	}

	for _, backup := range backups {
		if backup.ID == id {
			backup.Tags = tags
			
			// 保存更新后的备份列表
			err = s.saveBackups(backups)
			if err != nil {
				return err
			}

			// 发出事件通知
			if s.ctx != nil {
				wailsRuntime.EventsEmit(s.ctx, "backup-updated", id)
			}

			return nil
		}
	}

	return fmt.Errorf("备份不存在: %s", id)
}

// UpdateBackupDescription 更新备份描述
func (s *BackupService) UpdateBackupDescription(id, description string) error {
	backups, err := s.loadBackups()
	if err != nil {
		return err
	}

	for _, backup := range backups {
		if backup.ID == id {
			backup.Description = description
			
			// 保存更新后的备份列表
			err = s.saveBackups(backups)
			if err != nil {
				return err
			}

			// 发出事件通知
			if s.ctx != nil {
				wailsRuntime.EventsEmit(s.ctx, "backup-updated", id)
			}

			return nil
		}
	}

	return fmt.Errorf("备份不存在: %s", id)
}

// RestoreBackup 恢复备份
func (s *BackupService) RestoreBackup(id string) (string, error) {
	backup, err := s.GetBackupByID(id)
	if err != nil {
		return "", err
	}

	// 发出事件通知
	if s.ctx != nil {
		wailsRuntime.EventsEmit(s.ctx, "backup-restored", id)
	}

	return backup.Content, nil
}

// GetBackupStats 获取备份统计信息
func (s *BackupService) GetBackupStats() (map[string]interface{}, error) {
	backups, err := s.loadBackups()
	if err != nil {
		return nil, err
	}

	stats := map[string]interface{}{
		"total":     len(backups),
		"automatic": 0,
		"manual":    0,
		"totalSize": int64(0),
	}

	for _, backup := range backups {
		stats["totalSize"] = stats["totalSize"].(int64) + backup.Size
		if backup.IsAutomatic {
			stats["automatic"] = stats["automatic"].(int) + 1
		} else {
			stats["manual"] = stats["manual"].(int) + 1
		}
	}

	return stats, nil
} 