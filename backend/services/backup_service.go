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

// BackupService 备份
type BackupService struct {
	ctx        context.Context
	appDir     string
	backupFile string
	maxBackups int
}

// NewBackupService 创建服务
func NewBackupService(appDir string) *BackupService {
	backupFile := filepath.Join(appDir, "backups.json")

	err := os.MkdirAll(appDir, 0755)
	if err != nil {
		fmt.Printf("创建应用目录失败: %v\n", err)
	}

	return &BackupService{
		appDir:     appDir,
		backupFile: backupFile,
		maxBackups: 99,
	}
}

// SetContext 设置ctx
func (s *BackupService) SetContext(ctx context.Context) {
	s.ctx = ctx
}

// calculateHash 计算哈希
func (s *BackupService) calculateHash(content string) string {
	hash := md5.Sum([]byte(content))
	return hex.EncodeToString(hash[:])
}

// loadBackups 加载备份
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

// saveBackups 保存备份
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

	if isAutomatic {
		backups, err := s.loadBackups()
		if err == nil {
			for _, backup := range backups {
				if backup.Hash == hash && backup.IsAutomatic {
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

	backups, err := s.loadBackups()
	if err != nil {
		return nil, err
	}

	backups = append(backups, backup)

	if isAutomatic {
		backups = s.cleanupAutoBackups(backups)
	}

	err = s.saveBackups(backups)
	if err != nil {
		return nil, err
	}

	if s.ctx != nil {
		wailsRuntime.EventsEmit(s.ctx, "backup-created", backup.ID)
	}

	return backup, nil
}

// cleanupAutoBackups 清理备份
func (s *BackupService) cleanupAutoBackups(backups []*models.Backup) []*models.Backup {
	var autoBackups []*models.Backup
	var manualBackups []*models.Backup

	for _, backup := range backups {
		if backup.IsAutomatic {
			autoBackups = append(autoBackups, backup)
		} else {
			manualBackups = append(manualBackups, backup)
		}
	}

	sort.Slice(autoBackups, func(i, j int) bool {
		return autoBackups[i].Timestamp.Time.After(autoBackups[j].Timestamp.Time)
	})

	if len(autoBackups) > s.maxBackups {
		autoBackups = autoBackups[:s.maxBackups]
	}

	result := make([]*models.Backup, 0, len(autoBackups)+len(manualBackups))
	result = append(result, autoBackups...)
	result = append(result, manualBackups...)

	return result
}

// GetAllBackups 获取备份
func (s *BackupService) GetAllBackups() ([]*models.Backup, error) {
	backups, err := s.loadBackups()
	if err != nil {
		return nil, err
	}

	sort.Slice(backups, func(i, j int) bool {
		return backups[i].Timestamp.Time.After(backups[j].Timestamp.Time)
	})

	return backups, nil
}

// GetBackupByID 获取备份
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

	for i, backup := range backups {
		if backup.ID == id {
			if backup.IsAutomatic {
				return fmt.Errorf("不能删除自动备份")
			}

			backups = append(backups[:i], backups[i+1:]...)

			err = s.saveBackups(backups)
			if err != nil {
				return err
			}

			if s.ctx != nil {
				wailsRuntime.EventsEmit(s.ctx, "backup-deleted", id)
			}

			return nil
		}
	}

	return fmt.Errorf("备份不存在: %s", id)
}

// UpdateBackupTags 更新标签
func (s *BackupService) UpdateBackupTags(id string, tags []string) error {
	backups, err := s.loadBackups()
	if err != nil {
		return err
	}

	for _, backup := range backups {
		if backup.ID == id {
			backup.Tags = tags

			err = s.saveBackups(backups)
			if err != nil {
				return err
			}

			if s.ctx != nil {
				wailsRuntime.EventsEmit(s.ctx, "backup-updated", id)
			}

			return nil
		}
	}

	return fmt.Errorf("备份不存在: %s", id)
}

// UpdateBackupDescription 更新描述
func (s *BackupService) UpdateBackupDescription(id, description string) error {
	backups, err := s.loadBackups()
	if err != nil {
		return err
	}

	for _, backup := range backups {
		if backup.ID == id {
			backup.Description = description

			err = s.saveBackups(backups)
			if err != nil {
				return err
			}

			if s.ctx != nil {
				wailsRuntime.EventsEmit(s.ctx, "backup-updated", id)
			}

			return nil
		}
	}

	return fmt.Errorf("备份不存在: %s", id)
}

// RestoreBackup 恢复
func (s *BackupService) RestoreBackup(id string) (string, error) {
	backup, err := s.GetBackupByID(id)
	if err != nil {
		return "", err
	}

	if s.ctx != nil {
		wailsRuntime.EventsEmit(s.ctx, "backup-restored", id)
	}

	return backup.Content, nil
}

// GetBackupStats 获取统计
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

// ClearAllAutoBackups 清理自动备份
func (s *BackupService) ClearAllAutoBackups() error {
	backups, err := s.loadBackups()
	if err != nil {
		return err
	}

	var manualBackups []*models.Backup
	for _, backup := range backups {
		if !backup.IsAutomatic {
			manualBackups = append(manualBackups, backup)
		}
	}

	err = s.saveBackups(manualBackups)
	if err != nil {
		return err
	}

	if s.ctx != nil {
		wailsRuntime.EventsEmit(s.ctx, "backup-list-changed")
	}

	return nil
}
