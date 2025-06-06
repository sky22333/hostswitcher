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

// 默认的hosts文件内容常量
const defaultHostsContent = `# Copyright (c) 1993-2009 Microsoft Corp.
#
# This is a sample HOSTS file used by Microsoft TCP/IP for Windows.
#
# This file contains the mappings of IP addresses to host names. Each
# entry should be kept on an individual line. The IP address should
# be placed in the first column followed by the corresponding host name.
# The IP address and the host name should be separated by at least one
# space.
#
# Additionally, comments (such as these) may be inserted on individual
# lines or following the machine name denoted by a '#' symbol.
#
# For example:
#
#      102.54.94.97     rhino.acme.com          # source server
#       38.25.63.10     x.acme.com              # x client host

# localhost name resolution is handled within DNS itself.

127.0.0.1       localhost
::1             localhost
`

// ConfigService 处理 hosts 配置的服务
type ConfigService struct {
	ctx           context.Context
	configs       []*models.Config
	activeConfig  *models.Config
	appDir        string
	configsDir    string
	systemHosts   string
	backupService *BackupService
}

// NewConfigService 创建一个新的配置服务实例
func NewConfigService(ctx context.Context) *ConfigService {
	// 获取用户主目录
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Printf("获取用户主目录失败: %v", err)
		homeDir = "."
	}

	// 创建应用目录
	appDir := filepath.Join(homeDir, ".hosts-manager")
	configsDir := filepath.Join(appDir, "configs")

	// 创建必要的目录
	os.MkdirAll(configsDir, 0755)

	// 获取系统hosts文件路径
	var systemHosts string
	switch runtime.GOOS {
	case "windows":
		systemHosts = filepath.Join(os.Getenv("WINDIR"), "System32", "drivers", "etc", "hosts")
	case "darwin", "linux":
		systemHosts = "/etc/hosts"
	default:
		systemHosts = "/etc/hosts"
	}

	// 验证系统hosts文件是否存在
	if _, err := os.Stat(systemHosts); err != nil {
		log.Printf("系统hosts文件不存在: %s", systemHosts)
	}

	// 创建备份服务
	backupService := NewBackupService(appDir)
	backupService.SetContext(ctx)

	service := &ConfigService{
		ctx:           ctx,
		appDir:        appDir,
		configsDir:    configsDir,
		systemHosts:   systemHosts,
		configs:       []*models.Config{},
		backupService: backupService,
	}

	return service
}

// GetOSType 获取操作系统类型
func GetOSType() string {
	return runtime.GOOS
}

// SetContext 设置上下文
// wails:ignore
func (s *ConfigService) SetContext(ctx context.Context) {
	s.ctx = ctx
	// 同时更新备份服务的上下文
	if s.backupService != nil {
		s.backupService.SetContext(ctx)
	}
}

// Initialize 初始化服务
func (s *ConfigService) Initialize() error {
	// 尝试加载已保存的配置
	if err := s.loadConfigs(); err != nil {
		s.configs = []*models.Config{}
	}
	
	// 查找激活的配置
	for _, config := range s.configs {
		if config.IsActive {
			s.activeConfig = config
			break
		}
	}

	return nil
}

// loadConfigs 从文件加载配置
func (s *ConfigService) loadConfigs() error {
	// 创建配置文件路径
	configFile := filepath.Join(s.appDir, "configs.json")
	
	// 检查配置文件是否存在
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		// 文件不存在，初始化为空列表
		s.configs = []*models.Config{}
		return nil
	}

	// 读取配置文件
	data, err := os.ReadFile(configFile)
	if err != nil {
		return err
	}

	// 反序列化配置
	err = json.Unmarshal(data, &s.configs)
	if err != nil {
		return err
	}

	// 如果配置为nil，初始化为空列表
	if s.configs == nil {
		s.configs = []*models.Config{}
	}

	return nil
}

// GetAllConfigs 获取所有配置
func (s *ConfigService) GetAllConfigs() []*models.Config {
	return s.configs
}

// GetActiveConfig 获取当前激活的配置
func (s *ConfigService) GetActiveConfig() *models.Config {
	return s.activeConfig
}

// GetConfigByID 根据ID获取配置
func (s *ConfigService) GetConfigByID(id string) (*models.Config, error) {
	for _, config := range s.configs {
		if config.ID == id {
			return config, nil
		}
	}
	return nil, errors.New("配置不存在")
}

// CreateConfig 创建新配置
func (s *ConfigService) CreateConfig(name, description, content string) (*models.Config, error) {
	// 验证必要参数
	if strings.TrimSpace(name) == "" {
		return nil, errors.New("配置名称不能为空")
	}
	if strings.TrimSpace(content) == "" {
		return nil, errors.New("配置内容不能为空")
	}
	
	// 创建新配置
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

	// 添加到配置列表
	s.configs = append(s.configs, newConfig)

	// 保存配置
	err := s.saveConfigs()
	if err != nil {
		return nil, err
	}

	// 通知前端配置列表已更新
	if s.ctx != nil {
		wailsRuntime.EventsEmit(s.ctx, "config-list-changed")
	}

	return newConfig, nil
}

// UpdateConfig 更新配置
func (s *ConfigService) UpdateConfig(id, name, description, content string) (*models.Config, error) {
	// 验证必要参数
	if strings.TrimSpace(id) == "" {
		return nil, errors.New("配置ID不能为空")
	}
	if strings.TrimSpace(name) == "" {
		return nil, errors.New("配置名称不能为空")
	}
	if strings.TrimSpace(content) == "" {
		return nil, errors.New("配置内容不能为空")
	}
	
	// 查找配置
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

	// 更新配置
	config.Name = name
	config.Description = description
	config.Content = content
	config.UpdatedAt = models.GetCurrentTimeRFC3339()

	// 保存配置
	err := s.saveConfigs()
	if err != nil {
		return nil, err
	}

	// 通知前端配置列表已更新
	if s.ctx != nil {
		wailsRuntime.EventsEmit(s.ctx, "config-list-changed")
	}

	return config, nil
}

// DeleteConfig 删除配置
func (s *ConfigService) DeleteConfig(id string) error {
	// 不能删除激活的配置
	if s.activeConfig != nil && s.activeConfig.ID == id {
		return errors.New("不能删除当前激活的配置")
	}

	// 查找配置索引
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

	// 删除配置
	s.configs = append(s.configs[:index], s.configs[index+1:]...)

	// 保存配置
	err := s.saveConfigs()
	if err != nil {
		return err
	}

	// 通知前端配置列表已更新
	if s.ctx != nil {
		wailsRuntime.EventsEmit(s.ctx, "config-list-changed")
	}

	return nil
}

// ApplyConfig 应用指定的配置
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
	
	// 读取当前系统hosts文件内容
	currentContent, err := os.ReadFile(s.systemHosts)
	if err != nil {
		return fmt.Errorf("读取当前系统hosts失败: %v", err)
	}
	
	// 在应用配置前创建自动备份
	if s.backupService != nil {
		s.backupService.CreateBackup(string(currentContent), fmt.Sprintf("应用配置 '%s' 前的自动备份", config.Name), true, []string{"auto", "apply", config.Name})
	}
	
	// 验证配置内容
	if err := s.ValidateHostsContent(config.Content); err != nil {
		return fmt.Errorf("配置内容验证失败: %v", err)
	}
	
	// 应用配置
	err = os.WriteFile(s.systemHosts, []byte(config.Content), 0644)
	if err != nil {
		return fmt.Errorf("写入系统hosts失败: %v", err)
	}
	
	// 更新所有配置的激活状态
	for _, c := range s.configs {
		c.IsActive = (c.ID == id)
		if c.IsActive {
			c.UpdatedAt = models.JSONTime{Time: time.Now()}
		}
	}
	
	// 保存配置
	err = s.saveConfigs()
	if err != nil {
		// 如果保存失败，尝试恢复原内容
		os.WriteFile(s.systemHosts, currentContent, 0644)
		return fmt.Errorf("保存配置失败: %v", err)
	}
	
	// 发出配置应用事件
	wailsRuntime.EventsEmit(s.ctx, "config-applied", id)
	wailsRuntime.EventsEmit(s.ctx, "config-list-changed")
	
	return nil
}

// ReadSystemHosts 读取系统 hosts 文件内容
func (s *ConfigService) ReadSystemHosts() (string, error) {
	// 直接读取系统hosts文件，不通过任何配置
	hostsPath := s.GetSystemHostsPath()
	
	if s.ctx != nil {
		wailsRuntime.LogInfo(s.ctx, fmt.Sprintf("直接读取系统hosts文件: %s", hostsPath))
	}
	
	content, err := os.ReadFile(hostsPath)
	if err != nil {
		if s.ctx != nil {
			wailsRuntime.LogError(s.ctx, fmt.Sprintf("读取系统hosts文件失败: %v", err))
		}
		
		// 如果文件不存在，尝试创建默认文件
		if os.IsNotExist(err) {
			if createErr := s.createDefaultHostsFile(hostsPath); createErr == nil {
				// 重新尝试读取
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
		// 记录文件开头内容用于调试
		if len(result) > 100 {
			wailsRuntime.LogInfo(s.ctx, fmt.Sprintf("文件开头内容: %s...", result[:100]))
		} else {
			wailsRuntime.LogInfo(s.ctx, fmt.Sprintf("文件完整内容: %s", result))
		}
	}
	
	return result, nil
}

// WriteSystemHosts 写入系统hosts文件
func (s *ConfigService) WriteSystemHosts(content string) error {
	// 验证内容
	if err := s.ValidateHostsContent(content); err != nil {
		return fmt.Errorf("内容验证失败: %v", err)
	}
	
	// 在修改前创建自动备份
	if s.backupService != nil {
		if currentContent, err := s.ReadSystemHosts(); err == nil {
			s.backupService.CreateBackup(currentContent, "系统hosts文件自动备份", true, []string{"auto", "system"})
		}
	}
	
	// 写入文件
	err := os.WriteFile(s.systemHosts, []byte(content), 0644)
	if err != nil {
		return fmt.Errorf("写入失败: %v", err)
	}
	
	// 发出系统hosts更新事件
	wailsRuntime.EventsEmit(s.ctx, "system-hosts-updated")
	
	return nil
}

// GetSystemHostsPath 获取系统 hosts 文件路径
func (s *ConfigService) GetSystemHostsPath() string {
	if GetOSType() == "windows" {
		// 尝试多种环境变量来确定 Windows 系统目录
		systemRoot := os.Getenv("SystemRoot")
		if systemRoot == "" {
			systemRoot = os.Getenv("WINDIR")
		}
		if systemRoot == "" {
			// 最后回退到默认路径
			systemRoot = "C:\\Windows"
		}
		
		hostsPath := filepath.Join(systemRoot, "System32", "drivers", "etc", "hosts")
		
		// 记录日志
		if s.ctx != nil {
			wailsRuntime.LogInfo(s.ctx, fmt.Sprintf("系统 hosts 文件路径: %s", hostsPath))
		}
		
		// 检查文件是否存在
		if _, err := os.Stat(hostsPath); os.IsNotExist(err) {
			if s.ctx != nil {
				wailsRuntime.LogWarning(s.ctx, fmt.Sprintf("系统 hosts 文件不存在: %s，尝试创建", hostsPath))
			}
			// 尝试创建默认的hosts文件
			s.createDefaultHostsFile(hostsPath)
		}
		
		return hostsPath
	} else {
		// Unix/Linux/macOS
		return "/etc/hosts"
	}
}

// GetUserDataDir 获取用户数据目录
func (s *ConfigService) GetUserDataDir() string {
	return s.appDir
}

// createDefaultHostsFile 创建默认的hosts文件
func (s *ConfigService) createDefaultHostsFile(hostsPath string) error {
	// 确保目录存在
	dir := filepath.Dir(hostsPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		if s.ctx != nil {
			wailsRuntime.LogError(s.ctx, fmt.Sprintf("创建目录失败: %v", err))
		}
		return fmt.Errorf("创建目录失败: %v", err)
	}
	
	// 写入文件
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

// IsAdminRequired 检查是否需要管理员权限
func (s *ConfigService) IsAdminRequired() bool {
	if s.ctx != nil {
		wailsRuntime.LogInfo(s.ctx, "检查管理员权限...")
	}
	
	// 只在Windows上需要检查管理员权限
	if GetOSType() == "windows" {
		// 使用Windows标准API检查进程提升状态
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

// 使用Windows标准API检查进程是否已提升权限
func (s *ConfigService) isProcessElevated() bool {
	// 获取当前进程token（Windows Vista+标准方式）
	var token windows.Token
	err := windows.OpenProcessToken(windows.CurrentProcess(), windows.TOKEN_QUERY, &token)
	if err != nil {
		if s.ctx != nil {
			wailsRuntime.LogError(s.ctx, fmt.Sprintf("无法获取进程token: %v", err))
		}
		return false
	}
	defer token.Close()
	
	// 查询token提升状态（官方推荐方式）
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

// ValidateHostsContent 验证hosts文件内容格式
func (s *ConfigService) ValidateHostsContent(content string) error {
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		line = strings.TrimSpace(line)
		
		// 跳过空行和注释行
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		
		// 简单验证hosts条目格式
		parts := strings.Fields(line)
		if len(parts) < 2 {
			return fmt.Errorf("第 %d 行格式错误，应为: IP地址 主机名", i+1)
		}
		
		// 可以添加更多验证逻辑，比如IP地址格式验证
	}
	return nil
}

// saveConfigs 保存配置
func (s *ConfigService) saveConfigs() error {
	// 创建配置文件路径
	configFile := filepath.Join(s.appDir, "configs.json")
	
	// 序列化配置
	data, err := json.MarshalIndent(s.configs, "", "  ")
	if err != nil {
		return err
	}

	// 写入配置文件
	err = os.WriteFile(configFile, data, 0644)
	if err != nil {
		return err
	}

	return nil
}

// Cleanup 清理资源
func (s *ConfigService) Cleanup() {
	// 清空内存中的配置，帮助GC回收
	s.configs = nil
	s.activeConfig = nil
}

// UpdateConfigSource 更新配置的来源信息
func (s *ConfigService) UpdateConfigSource(id, source, remoteURL string) error {
	// 查找配置
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

	// 更新来源信息
	config.Source = source
	config.RemoteURL = remoteURL
	config.UpdatedAt = models.GetCurrentTimeRFC3339()

	// 保存配置
	err := s.saveConfigs()
	if err != nil {
		return err
	}

	// 通知前端配置列表已更新
	if s.ctx != nil {
		wailsRuntime.EventsEmit(s.ctx, "config-list-changed")
	}

	return nil
}

// RestoreDefaultHosts 恢复默认的hosts文件
func (s *ConfigService) RestoreDefaultHosts() error {
	// 写入默认内容
	err := os.WriteFile(s.systemHosts, []byte(defaultHostsContent), 0644)
	if err != nil {
		return fmt.Errorf("恢复默认hosts文件失败: %v", err)
	}

	// 清除所有配置的激活状态
	for _, config := range s.configs {
		config.IsActive = false
	}

	// 保存配置
	err = s.saveConfigs()
	if err != nil {
		return fmt.Errorf("保存配置失败: %v", err)
	}

	// 发出事件通知
	wailsRuntime.EventsEmit(s.ctx, "config-list-changed")
	wailsRuntime.EventsEmit(s.ctx, "system-hosts-updated")

	return nil
}



// FlushDNSCache 刷新系统DNS缓存
// 使用Windows API实现，兼容Win10及以上版本
func (s *ConfigService) FlushDNSCache() error {
	if s.ctx != nil {
		wailsRuntime.LogInfo(s.ctx, "开始使用Windows API刷新DNS缓存")
	}
	
	// 调用Windows API DnsFlushResolverCache
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

// 备份相关方法

// GetAllBackups 获取所有备份
func (s *ConfigService) GetAllBackups() ([]*models.Backup, error) {
	if s.backupService == nil {
		return []*models.Backup{}, nil
	}
	return s.backupService.GetAllBackups()
}

// CreateManualBackup 创建手动备份
func (s *ConfigService) CreateManualBackup(description string, tags []string) (*models.Backup, error) {
	if s.backupService == nil {
		return nil, fmt.Errorf("备份服务未初始化")
	}
	
	// 读取当前系统hosts内容
	content, err := s.ReadSystemHosts()
	if err != nil {
		return nil, fmt.Errorf("读取系统hosts失败: %v", err)
	}
	
	if description == "" {
		description = "手动备份"
	}
	
	return s.backupService.CreateBackup(content, description, false, tags)
}

// RestoreFromBackup 从备份恢复
func (s *ConfigService) RestoreFromBackup(backupID string) error {
	if s.backupService == nil {
		return fmt.Errorf("备份服务未初始化")
	}
	
	content, err := s.backupService.RestoreBackup(backupID)
	if err != nil {
		return err
	}
	
	// 在恢复前创建自动备份
	if currentContent, readErr := s.ReadSystemHosts(); readErr == nil {
		s.backupService.CreateBackup(currentContent, "恢复备份前的自动备份", true, []string{"auto", "restore"})
	}
	
	// 写入恢复的内容
	return s.WriteSystemHosts(content)
}

// DeleteBackup 删除备份
func (s *ConfigService) DeleteBackup(backupID string) error {
	if s.backupService == nil {
		return fmt.Errorf("备份服务未初始化")
	}
	return s.backupService.DeleteBackup(backupID)
}

// UpdateBackupTags 更新备份标签
func (s *ConfigService) UpdateBackupTags(backupID string, tags []string) error {
	if s.backupService == nil {
		return fmt.Errorf("备份服务未初始化")
	}
	return s.backupService.UpdateBackupTags(backupID, tags)
}

// UpdateBackupDescription 更新备份描述
func (s *ConfigService) UpdateBackupDescription(backupID, description string) error {
	if s.backupService == nil {
		return fmt.Errorf("备份服务未初始化")
	}
	return s.backupService.UpdateBackupDescription(backupID, description)
}

// GetBackupStats 获取备份统计信息
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

// flushDNSResolverCache 调用Windows API DnsFlushResolverCache清理DNS缓存
func flushDNSResolverCache() error {
	// 加载dnsapi.dll
	dnsapi, err := syscall.LoadLibrary("dnsapi.dll")
	if err != nil {
		return fmt.Errorf("加载dnsapi.dll失败: %v", err)
	}
	defer syscall.FreeLibrary(dnsapi)
	
	// 获取DnsFlushResolverCache函数
	proc, err := syscall.GetProcAddress(dnsapi, "DnsFlushResolverCache")
	if err != nil {
		return fmt.Errorf("获取DnsFlushResolverCache函数失败: %v", err)
	}
	
	// 调用DnsFlushResolverCache函数
	_, _, callErr := syscall.Syscall(proc, 0, 0, 0, 0)
	
	// 因为该API没有返回值，Windows文档说明该函数总是成功的
	if callErr != 0 {
		// 如果有系统调用错误，将其转换为Go错误
		return fmt.Errorf("调用DnsFlushResolverCache失败: %v", callErr)
	}
	
	return nil
}
