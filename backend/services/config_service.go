package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/google/uuid"
	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
	"golang.org/x/text/encoding/simplifiedchinese"

	"hostswitcher/backend/models"
)

// ConfigService 处理 hosts 配置的服务
type ConfigService struct {
	ctx          context.Context
	configs      []*models.Config
	activeConfig *models.Config
	appDir       string
	configsDir   string
	systemHosts  string
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

	service := &ConfigService{
		ctx:         ctx,
		appDir:      appDir,
		configsDir:  configsDir,
		systemHosts: systemHosts,
		configs:     []*models.Config{},
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
	
	// 默认的hosts文件内容
	defaultContent := `# Copyright (c) 1993-2009 Microsoft Corp.
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

	// 写入文件
	err := os.WriteFile(hostsPath, []byte(defaultContent), 0644)
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
		wailsRuntime.LogInfo(s.ctx, "开始检查管理员权限...")
	}
	
	// 在Windows上修改系统hosts文件通常需要管理员权限
	if GetOSType() == "windows" {
		// 尝试写入测试，检查是否有权限
		testFile := s.systemHosts + ".test"
		
		if s.ctx != nil {
			wailsRuntime.LogInfo(s.ctx, fmt.Sprintf("尝试创建测试文件: %s", testFile))
		}
		
		err := os.WriteFile(testFile, []byte("test"), 0644)
		if err != nil {
			if s.ctx != nil {
				wailsRuntime.LogWarning(s.ctx, fmt.Sprintf("创建测试文件失败: %v，需要管理员权限", err))
			}
			return true
		}
		
		// 清理测试文件
		removeErr := os.Remove(testFile)
		if removeErr != nil && s.ctx != nil {
			wailsRuntime.LogWarning(s.ctx, fmt.Sprintf("删除测试文件失败: %v", removeErr))
		}
		
		if s.ctx != nil {
			wailsRuntime.LogInfo(s.ctx, "权限检查通过，有足够权限修改hosts文件")
		}
		return false
	}
	
	if s.ctx != nil {
		wailsRuntime.LogInfo(s.ctx, "非Windows系统，不需要管理员权限")
	}
	return false
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
	// 默认的hosts文件内容
	defaultContent := `# Copyright (c) 1993-2009 Microsoft Corp.
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

	// 写入默认内容
	err := os.WriteFile(s.systemHosts, []byte(defaultContent), 0644)
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

// WriteSystemHostsWithANSI 以ANSI编码写入系统hosts文件，提高兼容性
func (s *ConfigService) WriteSystemHostsWithANSI(content string) error {
	// 验证内容
	if err := s.ValidateHostsContent(content); err != nil {
		return fmt.Errorf("内容验证失败: %v", err)
	}
	
	// 转换为ANSI编码 (GBK)
	encoder := simplifiedchinese.GBK.NewEncoder()
	ansiContent, err := encoder.String(content)
	if err != nil {
		// 如果转换失败，使用原始内容
		if s.ctx != nil {
			wailsRuntime.LogWarning(s.ctx, fmt.Sprintf("ANSI编码转换失败，使用原始内容: %v", err))
		}
		ansiContent = content
	}
	
	// 写入文件
	err = os.WriteFile(s.systemHosts, []byte(ansiContent), 0644)
	if err != nil {
		return fmt.Errorf("写入失败: %v", err)
	}
	
	// 发出系统hosts更新事件
	if s.ctx != nil {
		wailsRuntime.EventsEmit(s.ctx, "system-hosts-updated")
		wailsRuntime.LogInfo(s.ctx, "已使用ANSI编码保存hosts文件")
	}
	
	return nil
}

// FlushDNSCache 刷新系统DNS缓存
func (s *ConfigService) FlushDNSCache() error {
	var cmd *exec.Cmd
	var cmdDesc string
	
	switch runtime.GOOS {
	case "windows":
		// 获取Windows系统目录
		systemRoot := os.Getenv("SystemRoot")
		if systemRoot == "" {
			systemRoot = os.Getenv("WINDIR")
		}
		if systemRoot == "" {
			systemRoot = "C:\\Windows"
		}
		
		// 尝试多个可能的ipconfig路径
		ipConfigPaths := []string{
			filepath.Join(systemRoot, "System32", "ipconfig.exe"),
			filepath.Join(systemRoot, "SysWOW64", "ipconfig.exe"),
			"ipconfig.exe", // 最后尝试相对路径
		}
		
		var ipConfigPath string
		for _, path := range ipConfigPaths {
			if _, err := os.Stat(path); err == nil {
				ipConfigPath = path
				break
			}
		}
		
		if ipConfigPath == "" {
			// 如果找不到ipconfig.exe，尝试直接使用命令名
			ipConfigPath = "ipconfig"
		}
		
		cmd = exec.Command(ipConfigPath, "/flushdns")
		cmdDesc = fmt.Sprintf("%s /flushdns", ipConfigPath)
		
		if s.ctx != nil {
			wailsRuntime.LogInfo(s.ctx, fmt.Sprintf("使用ipconfig路径: %s", ipConfigPath))
		}
		
	case "darwin":
		cmd = exec.Command("sudo", "dscacheutil", "-flushcache")
		cmdDesc = "sudo dscacheutil -flushcache"
	case "linux":
		// 尝试多种Linux发行版的DNS缓存刷新命令
		if _, err := exec.LookPath("systemctl"); err == nil {
			cmd = exec.Command("sudo", "systemctl", "restart", "systemd-resolved")
			cmdDesc = "sudo systemctl restart systemd-resolved"
		} else if _, err := exec.LookPath("service"); err == nil {
			cmd = exec.Command("sudo", "service", "network-manager", "restart")
			cmdDesc = "sudo service network-manager restart"
		} else {
			return fmt.Errorf("未找到合适的DNS缓存刷新命令")
		}
	default:
		return fmt.Errorf("不支持的操作系统: %s", runtime.GOOS)
	}
	
	if s.ctx != nil {
		wailsRuntime.LogInfo(s.ctx, fmt.Sprintf("正在执行DNS缓存刷新命令: %s", cmdDesc))
	}
	
	// 执行命令
	output, err := cmd.CombinedOutput()
	outputStr := string(output)
	
	if s.ctx != nil {
		wailsRuntime.LogInfo(s.ctx, fmt.Sprintf("命令输出: %s", outputStr))
		if err != nil {
			wailsRuntime.LogError(s.ctx, fmt.Sprintf("命令执行错误: %v", err))
		}
	}
	
	if err != nil {
		// 改进错误处理，提供更详细的错误信息
		errorDetails := ""
		if err.Error() != "" {
			errorDetails = err.Error()
		} else {
			errorDetails = "未知错误"
		}
		
		if outputStr != "" {
			errorDetails += fmt.Sprintf(", 输出: %s", outputStr)
		}
		
		errorMsg := fmt.Sprintf("刷新DNS缓存失败: %s", errorDetails)
		if s.ctx != nil {
			wailsRuntime.LogError(s.ctx, errorMsg)
		}
		return fmt.Errorf(errorMsg)
	}
	
	// 在Windows上，即使成功也可能没有输出，这是正常的
	if runtime.GOOS == "windows" && outputStr == "" {
		if s.ctx != nil {
			wailsRuntime.LogInfo(s.ctx, "DNS缓存刷新命令执行成功（无输出是正常的）")
		}
	} else if s.ctx != nil {
		wailsRuntime.LogInfo(s.ctx, fmt.Sprintf("DNS缓存刷新成功，输出: %s", outputStr))
	}
	
	return nil
}
