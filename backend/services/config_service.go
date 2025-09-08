package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"
	"unsafe"

	"github.com/google/uuid"
	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
	"golang.org/x/sys/windows"

	"hostswitcher/backend/models"
)

const defaultHostsContent = `# Copyright (c) 1993-2009 Microsoft Corp.
#
# This is a sample HOSTS file used by Microsoft TCP/IP for Windows.
#
127.0.0.1       localhost
::1             localhost
`

// ConfigService 配置服务
type ConfigService struct {
	ctx           context.Context
	configs       []*models.Config
	activeConfig  *models.Config
	appDir        string
	systemHosts   string
	backupService *BackupService
}

// NewConfigService 创建服务
func NewConfigService(ctx context.Context) *ConfigService {

	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Printf("获取用户主目录失败: %v", err)
		homeDir = "."
	}


	appDir := filepath.Join(homeDir, ".hosts-manager")


	os.MkdirAll(appDir, 0755)


	var systemHosts string
	switch runtime.GOOS {
	case "windows":
		systemHosts = filepath.Join(os.Getenv("WINDIR"), "System32", "drivers", "etc", "hosts")
	case "darwin", "linux":
		systemHosts = "/etc/hosts"
	default:
		systemHosts = "/etc/hosts"
	}


	if _, err := os.Stat(systemHosts); err != nil {
		log.Printf("系统hosts文件不存在: %s", systemHosts)
	}


	backupService := NewBackupService(appDir)
	backupService.SetContext(ctx)

	service := &ConfigService{
		ctx:           ctx,
		appDir:        appDir,
		systemHosts:   systemHosts,
		configs:       []*models.Config{},
		backupService: backupService,
	}

	return service
}

// GetOSType 获取OS
func GetOSType() string {
	return runtime.GOOS
}

// SetContext 设置ctx
// wails:ignore
func (s *ConfigService) SetContext(ctx context.Context) {
	s.ctx = ctx

	if s.backupService != nil {
		s.backupService.SetContext(ctx)
	}
}

// Initialize 初始化
func (s *ConfigService) Initialize() error {

	if err := s.loadConfigs(); err != nil {
		s.configs = []*models.Config{}
	}
	

	for _, config := range s.configs {
		if config.IsActive {
			s.activeConfig = config
			break
		}
	}

	return nil
}

// loadConfigs 加载
func (s *ConfigService) loadConfigs() error {

	configFile := filepath.Join(s.appDir, "configs.json")
	

	if _, err := os.Stat(configFile); os.IsNotExist(err) {

		s.configs = []*models.Config{}
		return nil
	}


	data, err := os.ReadFile(configFile)
	if err != nil {
		return err
	}


	err = json.Unmarshal(data, &s.configs)
	if err != nil {
		return err
	}


	if s.configs == nil {
		s.configs = []*models.Config{}
	}

	return nil
}

// GetAllConfigs 获取配置
func (s *ConfigService) GetAllConfigs() []*models.Config {
	return s.configs
}

// GetActiveConfig 获取激活
func (s *ConfigService) GetActiveConfig() *models.Config {
	return s.activeConfig
}

// GetConfigByID 获取配置
func (s *ConfigService) GetConfigByID(id string) (*models.Config, error) {
	for _, config := range s.configs {
		if config.ID == id {
			return config, nil
		}
	}
	return nil, errors.New("配置不存在")
}

// CreateConfig 创建
func (s *ConfigService) CreateConfig(name, description, content string) (*models.Config, error) {

	if strings.TrimSpace(name) == "" {
		return nil, errors.New("配置名称不能为空")
	}
	if strings.TrimSpace(content) == "" {
		return nil, errors.New("配置内容不能为空")
	}
	

	newConfig := &models.Config{
		ID:          uuid.New().String(),
		Name:        name,
		Description: description,
		Content:     content,
		IsActive:    false,
		Source:      "local",
		CreatedAt:   models.GetCurrentTimeRFC3339(),
		UpdatedAt:   models.GetCurrentTimeRFC3339(),
	}


	s.configs = append(s.configs, newConfig)


	err := s.saveConfigs()
	if err != nil {
		return nil, err
	}


	if s.ctx != nil {
		wailsRuntime.EventsEmit(s.ctx, "config-list-changed")
	}

	return newConfig, nil
}

// UpdateConfig 更新
func (s *ConfigService) UpdateConfig(id, name, description, content string) (*models.Config, error) {

	if strings.TrimSpace(id) == "" {
		return nil, errors.New("配置ID不能为空")
	}
	if strings.TrimSpace(name) == "" {
		return nil, errors.New("配置名称不能为空")
	}
	if strings.TrimSpace(content) == "" {
		return nil, errors.New("配置内容不能为空")
	}
	

	var config *models.Config
	for _, c := range s.configs {
		if c.ID == id {
			config = c
			break
		}
	}

	if config == nil {
		return nil, errors.New("配置不存在")
	}


	config.Name = name
	config.Description = description
	config.Content = content
	config.UpdatedAt = models.GetCurrentTimeRFC3339()


	err := s.saveConfigs()
	if err != nil {
		return nil, err
	}


	if s.ctx != nil {
		wailsRuntime.EventsEmit(s.ctx, "config-list-changed")
	}

	return config, nil
}

// DeleteConfig 删除
func (s *ConfigService) DeleteConfig(id string) error {

	if s.activeConfig != nil && s.activeConfig.ID == id {
		return errors.New("不能删除当前激活的配置")
	}


	index := -1
	for i, config := range s.configs {
		if config.ID == id {
			index = i
			break
		}
	}

	if index == -1 {
		return errors.New("配置不存在")
	}


	s.configs = append(s.configs[:index], s.configs[index+1:]...)


	err := s.saveConfigs()
	if err != nil {
		return err
	}


	if s.ctx != nil {
		wailsRuntime.EventsEmit(s.ctx, "config-list-changed")
	}

	return nil
}

// ApplyConfig 应用配置
func (s *ConfigService) ApplyConfig(id string) error {
	var config *models.Config
	for _, c := range s.configs {
		if c.ID == id {
			config = c
			break
		}
	}
	
	if config == nil {
		return errors.New("配置不存在")
	}
	

	currentContent, err := os.ReadFile(s.systemHosts)
	if err != nil {
		return fmt.Errorf("读取当前系统hosts失败: %v", err)
	}
	

	if s.backupService != nil {
		s.backupService.CreateBackup(string(currentContent), fmt.Sprintf("应用配置 '%s' 前的自动备份", config.Name), true, []string{"auto", "apply", config.Name})
	}
	

	if err := s.ValidateHostsContent(config.Content); err != nil {
		return fmt.Errorf("配置内容验证失败: %v", err)
	}
	

	err = os.WriteFile(s.systemHosts, []byte(config.Content), 0644)
	if err != nil {
		return fmt.Errorf("写入系统hosts失败: %v", err)
	}
	

	for _, c := range s.configs {
		c.IsActive = (c.ID == id)
		if c.IsActive {
			c.UpdatedAt = models.JSONTime{Time: time.Now()}
		}
	}
	

	err = s.saveConfigs()
	if err != nil {

		os.WriteFile(s.systemHosts, currentContent, 0644)
		return fmt.Errorf("保存配置失败: %v", err)
	}
	

	wailsRuntime.EventsEmit(s.ctx, "config-applied", id)
	wailsRuntime.EventsEmit(s.ctx, "config-list-changed")
	
	return nil
}

// ReadSystemHosts 读取hosts
func (s *ConfigService) ReadSystemHosts() (string, error) {

	hostsPath := s.GetSystemHostsPath()
	
	if s.ctx != nil {
		wailsRuntime.LogInfo(s.ctx, fmt.Sprintf("直接读取系统hosts文件: %s", hostsPath))
	}
	
	content, err := os.ReadFile(hostsPath)
	if err != nil {
		if s.ctx != nil {
			wailsRuntime.LogError(s.ctx, fmt.Sprintf("读取系统hosts文件失败: %v", err))
		}
		

		if os.IsNotExist(err) {
			if createErr := s.createDefaultHostsFile(hostsPath); createErr == nil {
	
				content, err = os.ReadFile(hostsPath)
			}
		}
		
		if err != nil {
			return "", fmt.Errorf("读取系统hosts文件失败: %v", err)
		}
	}
	
	result := string(content)
	if s.ctx != nil {
		wailsRuntime.LogInfo(s.ctx, fmt.Sprintf("成功读取系统hosts文件，内容长度: %d 字符", len(result)))

		if len(result) > 100 {
			wailsRuntime.LogInfo(s.ctx, fmt.Sprintf("文件开头内容: %s...", result[:100]))
		} else {
			wailsRuntime.LogInfo(s.ctx, fmt.Sprintf("文件完整内容: %s", result))
		}
	}
	
	return result, nil
}

// WriteSystemHosts 写入hosts
func (s *ConfigService) WriteSystemHosts(content string) error {

	if err := s.ValidateHostsContent(content); err != nil {
		return fmt.Errorf("内容验证失败: %v", err)
	}
	

	if s.backupService != nil {
		if currentContent, err := s.ReadSystemHosts(); err == nil {
			s.backupService.CreateBackup(currentContent, "系统hosts文件自动备份", true, []string{"auto", "system"})
		}
	}
	

	err := os.WriteFile(s.systemHosts, []byte(content), 0644)
	if err != nil {
		return fmt.Errorf("写入失败: %v", err)
	}
	

	wailsRuntime.EventsEmit(s.ctx, "system-hosts-updated")
	
	return nil
}

// GetSystemHostsPath 获取hosts路径
func (s *ConfigService) GetSystemHostsPath() string {
	if GetOSType() == "windows" {

		systemRoot := os.Getenv("SystemRoot")
		if systemRoot == "" {
			systemRoot = os.Getenv("WINDIR")
		}
		if systemRoot == "" {

			systemRoot = "C:\\Windows"
		}
		
		hostsPath := filepath.Join(systemRoot, "System32", "drivers", "etc", "hosts")
		

		if s.ctx != nil {
			wailsRuntime.LogInfo(s.ctx, fmt.Sprintf("系统 hosts 文件路径: %s", hostsPath))
		}
		

		if _, err := os.Stat(hostsPath); os.IsNotExist(err) {
			if s.ctx != nil {
				wailsRuntime.LogWarning(s.ctx, fmt.Sprintf("系统 hosts 文件不存在: %s，尝试创建", hostsPath))
			}

			s.createDefaultHostsFile(hostsPath)
		}
		
		return hostsPath
	} else {

		return "/etc/hosts"
	}
}

// GetUserDataDir 获取数据目录
func (s *ConfigService) GetUserDataDir() string {
	return s.appDir
}

// createDefaultHostsFile 创建默认hosts
func (s *ConfigService) createDefaultHostsFile(hostsPath string) error {

	dir := filepath.Dir(hostsPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		if s.ctx != nil {
			wailsRuntime.LogError(s.ctx, fmt.Sprintf("创建目录失败: %v", err))
		}
		return fmt.Errorf("创建目录失败: %v", err)
	}
	

	err := os.WriteFile(hostsPath, []byte(defaultHostsContent), 0644)
	if err != nil {
		if s.ctx != nil {
			wailsRuntime.LogError(s.ctx, fmt.Sprintf("创建默认hosts文件失败: %v", err))
		}
		return fmt.Errorf("创建默认hosts文件失败: %v", err)
	}
	
	if s.ctx != nil {
		wailsRuntime.LogInfo(s.ctx, fmt.Sprintf("已创建默认hosts文件: %s", hostsPath))
	}
	
	return nil
}

// IsAdminRequired 检查权限
func (s *ConfigService) IsAdminRequired() bool {
	if s.ctx != nil {
		wailsRuntime.LogInfo(s.ctx, "检查管理员权限...")
	}
	

	if GetOSType() == "windows" {

		elevated := s.isProcessElevated()
		
		if s.ctx != nil {
			if elevated {
				wailsRuntime.LogInfo(s.ctx, "检测到管理员权限")
			} else {
				wailsRuntime.LogWarning(s.ctx, "需要管理员权限")
			}
		}
		
		return !elevated
	}
	
	if s.ctx != nil {
		wailsRuntime.LogInfo(s.ctx, "非Windows系统，不需要管理员权限")
	}
	return false
}

// isProcessElevated 检查权限
func (s *ConfigService) isProcessElevated() bool {

	var token windows.Token
	err := windows.OpenProcessToken(windows.CurrentProcess(), windows.TOKEN_QUERY, &token)
	if err != nil {
		if s.ctx != nil {
			wailsRuntime.LogError(s.ctx, fmt.Sprintf("无法获取进程token: %v", err))
		}
		return false
	}
	defer token.Close()
	

	var isElevated uint32
	var returnedLen uint32
	
	err = windows.GetTokenInformation(
		token,
		windows.TokenElevation,
		(*byte)(unsafe.Pointer(&isElevated)),
		uint32(unsafe.Sizeof(isElevated)),
		&returnedLen,
	)
	
	if err != nil {
		if s.ctx != nil {
			wailsRuntime.LogError(s.ctx, fmt.Sprintf("无法查询token信息: %v", err))
		}
		return false
	}
	
	return isElevated != 0
}

// ValidateHostsContent 验证内容
func (s *ConfigService) ValidateHostsContent(content string) error {
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		line = strings.TrimSpace(line)
		

		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		

		parts := strings.Fields(line)
		if len(parts) < 2 {
			return fmt.Errorf("第 %d 行格式错误，应为: IP地址 主机名", i+1)
		}
		

	}
	return nil
}

// saveConfigs 保存
func (s *ConfigService) saveConfigs() error {

	configFile := filepath.Join(s.appDir, "configs.json")
	

	data, err := json.MarshalIndent(s.configs, "", "  ")
	if err != nil {
		return err
	}


	err = os.WriteFile(configFile, data, 0644)
	if err != nil {
		return err
	}

	return nil
}

// Cleanup 清理
func (s *ConfigService) Cleanup() {

	s.configs = nil
	s.activeConfig = nil
}

// UpdateConfigSource 更新来源
func (s *ConfigService) UpdateConfigSource(id, source, remoteURL string) error {

	var config *models.Config
	for _, c := range s.configs {
		if c.ID == id {
			config = c
			break
		}
	}

	if config == nil {
		return errors.New("配置不存在")
	}


	config.Source = source
	config.RemoteURL = remoteURL
	config.UpdatedAt = models.GetCurrentTimeRFC3339()


	err := s.saveConfigs()
	if err != nil {
		return err
	}


	if s.ctx != nil {
		wailsRuntime.EventsEmit(s.ctx, "config-list-changed")
	}

	return nil
}

// RestoreDefaultHosts 恢复默认
func (s *ConfigService) RestoreDefaultHosts() error {

	err := os.WriteFile(s.systemHosts, []byte(defaultHostsContent), 0644)
	if err != nil {
		return fmt.Errorf("恢复默认hosts文件失败: %v", err)
	}


	for _, config := range s.configs {
		config.IsActive = false
	}


	err = s.saveConfigs()
	if err != nil {
		return fmt.Errorf("保存配置失败: %v", err)
	}


	wailsRuntime.EventsEmit(s.ctx, "config-list-changed")
	wailsRuntime.EventsEmit(s.ctx, "system-hosts-updated")

	return nil
}

// FlushDNSCache 刷新DNS缓存
func (s *ConfigService) FlushDNSCache() error {
	if s.ctx != nil {
		wailsRuntime.LogInfo(s.ctx, "开始使用Windows API刷新DNS缓存")
	}
	

	err := flushDNSResolverCache()
	if err != nil {
		errorMsg := fmt.Sprintf("刷新DNS缓存失败: %v", err)
		if s.ctx != nil {
			wailsRuntime.LogError(s.ctx, errorMsg)
		}
		return fmt.Errorf(errorMsg)
	}
	
	if s.ctx != nil {
		wailsRuntime.LogInfo(s.ctx, "DNS缓存刷新成功")
	}
	
	return nil
}



// GetAllBackups 获取备份
func (s *ConfigService) GetAllBackups() ([]*models.Backup, error) {
	if s.backupService == nil {
		return []*models.Backup{}, nil
	}
	return s.backupService.GetAllBackups()
}

// CreateManualBackup 创建备份
func (s *ConfigService) CreateManualBackup(description string, tags []string) (*models.Backup, error) {
	if s.backupService == nil {
		return nil, fmt.Errorf("备份服务未初始化")
	}
	

	content, err := s.ReadSystemHosts()
	if err != nil {
		return nil, fmt.Errorf("读取系统hosts失败: %v", err)
	}
	
	if description == "" {
		description = "手动备份"
	}
	
	return s.backupService.CreateBackup(content, description, false, tags)
}

// CreateManualBackupWithContent 创建备份
func (s *ConfigService) CreateManualBackupWithContent(description, content string, tags []string) (*models.Backup, error) {
	if s.backupService == nil {
		return nil, fmt.Errorf("备份服务未初始化")
	}
	
	if description == "" {
		description = "手动备份"
	}
	
	if content == "" {

		var err error
		content, err = s.ReadSystemHosts()
		if err != nil {
			return nil, fmt.Errorf("读取系统hosts失败: %v", err)
		}
	}
	
	return s.backupService.CreateBackup(content, description, false, tags)
}

// RestoreFromBackup 恢复
func (s *ConfigService) RestoreFromBackup(backupID string) error {
	if s.backupService == nil {
		return fmt.Errorf("备份服务未初始化")
	}
	
	content, err := s.backupService.RestoreBackup(backupID)
	if err != nil {
		return err
	}
	

	if currentContent, readErr := s.ReadSystemHosts(); readErr == nil {
		s.backupService.CreateBackup(currentContent, "恢复备份前的自动备份", true, []string{"auto", "restore"})
	}
	

	return s.WriteSystemHosts(content)
}

// DeleteBackup 删除
func (s *ConfigService) DeleteBackup(backupID string) error {
	if s.backupService == nil {
		return fmt.Errorf("备份服务未初始化")
	}
	return s.backupService.DeleteBackup(backupID)
}

// UpdateBackupTags 更新标签
func (s *ConfigService) UpdateBackupTags(backupID string, tags []string) error {
	if s.backupService == nil {
		return fmt.Errorf("备份服务未初始化")
	}
	return s.backupService.UpdateBackupTags(backupID, tags)
}

// UpdateBackupDescription 更新描述
func (s *ConfigService) UpdateBackupDescription(backupID, description string) error {
	if s.backupService == nil {
		return fmt.Errorf("备份服务未初始化")
	}
	return s.backupService.UpdateBackupDescription(backupID, description)
}

// GetBackupStats 获取统计
func (s *ConfigService) GetBackupStats() (map[string]interface{}, error) {
	if s.backupService == nil {
		return map[string]interface{}{
			"total":     0,
			"automatic": 0,
			"manual":    0,
			"totalSize": int64(0),
		}, nil
	}
	return s.backupService.GetBackupStats()
}

// ClearAllAutoBackups 清理备份
func (s *ConfigService) ClearAllAutoBackups() error {
	if s.backupService == nil {
		return fmt.Errorf("备份服务未初始化")
	}
	return s.backupService.ClearAllAutoBackups()
}

// flushDNSResolverCache 清理DNS
func flushDNSResolverCache() error {

	dnsapi, err := syscall.LoadLibrary("dnsapi.dll")
	if err != nil {
		return fmt.Errorf("加载dnsapi.dll失败: %v", err)
	}
	defer syscall.FreeLibrary(dnsapi)
	

	proc, err := syscall.GetProcAddress(dnsapi, "DnsFlushResolverCache")
	if err != nil {
		return fmt.Errorf("获取DnsFlushResolverCache函数失败: %v", err)
	}
	

	_, _, callErr := syscall.Syscall(proc, 0, 0, 0, 0)
	

	if callErr != 0 {

		return fmt.Errorf("调用DnsFlushResolverCache失败: %v", callErr)
	}
	
	return nil
}
