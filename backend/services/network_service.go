package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"

	"hostswitcher/backend/models"
)

// NetworkService 网络服务
type NetworkService struct {
	ctx           context.Context
	configService *ConfigService
	remoteSources []*models.RemoteSource
	remoteFile    string
	httpClient    *http.Client
}

// NewNetworkService 创建网络服务
func NewNetworkService(configService *ConfigService) *NetworkService {

	userDir, err := os.UserHomeDir()
	if err != nil {
		userDir = "."
	}


	appDir := filepath.Join(userDir, ".hosts-manager")
	remoteFile := filepath.Join(appDir, "remote_sources.json")


	os.MkdirAll(appDir, 0755)


	httpClient := &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 10,
			IdleConnTimeout:     90 * time.Second,
		},
	}

	return &NetworkService{
		configService: configService,
		remoteFile:    remoteFile,
		remoteSources: []*models.RemoteSource{},
		httpClient:    httpClient,
	}
}

// SetContext 设置上下文
// wails:ignore
func (s *NetworkService) SetContext(ctx context.Context) {
	s.ctx = ctx
}

// Initialize 初始化
func (s *NetworkService) Initialize() error {

	if err := s.loadRemoteSources(); err != nil {

		s.remoteSources = []*models.RemoteSource{}
	}
	

	go func() {

		time.Sleep(3 * time.Second)
		s.updateStartupSources()
	}()
	
	return nil
}

// findRemoteSource 查找远程源
func (s *NetworkService) findRemoteSource(id string) (*models.RemoteSource, error) {

	if strings.TrimSpace(id) == "" {
		return nil, errors.New("远程源ID不能为空")
	}
	
	for _, source := range s.remoteSources {
		if source.ID == id {
			return source, nil
		}
	}
	
	return nil, fmt.Errorf("未找到ID为 %s 的远程源", id)
}

// validateRemoteSourceParams 验证远程源参数的通用方法
func (s *NetworkService) validateRemoteSourceParams(name, url, updateFreq string) error {
	if strings.TrimSpace(name) == "" {
		return errors.New("名称不能为空")
	}
	if strings.TrimSpace(url) == "" {
		return errors.New("URL不能为空")
	}
	
	// 验证URL
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		return errors.New("URL必须以http://或https://开头")
	}

	// 验证更新频率
	if updateFreq != "manual" && updateFreq != "startup" {
		return errors.New("更新频率必须是manual或startup")
	}
	
	return nil
}

// loadRemoteSources 从文件加载远程源
func (s *NetworkService) loadRemoteSources() error {
	// 检查远程源文件是否存在
	if _, err := os.Stat(s.remoteFile); os.IsNotExist(err) {
		// 文件不存在，初始化为空列表
		s.remoteSources = []*models.RemoteSource{}
		return nil
	}

	// 读取远程源文件
	data, err := os.ReadFile(s.remoteFile)
	if err != nil {
		return err
	}

	// 如果文件为空，初始化为空列表
	if len(data) == 0 {
		s.remoteSources = []*models.RemoteSource{}
		return nil
	}

	// 尝试解析为数组格式（标准格式）
	var sourcesArray []*models.RemoteSource
	err = json.Unmarshal(data, &sourcesArray)
	if err == nil {
		// 成功解析为数组
		s.remoteSources = sourcesArray
		if s.remoteSources == nil {
			s.remoteSources = []*models.RemoteSource{}
		}
		// 验证并补全必需字段
		s.validateAndFixRemoteSources()
		return nil
	}

	if s.ctx != nil {
		wailsRuntime.LogInfo(s.ctx, "尝试解析为数组格式失败，尝试解析为单个对象格式")
	}

	// 尝试解析为单个对象格式（兼容用户现有数据）
	var singleSource models.RemoteSource
	err = json.Unmarshal(data, &singleSource)
	if err == nil {
		// 成功解析为单个对象，转换为数组
		s.remoteSources = []*models.RemoteSource{&singleSource}
		if s.ctx != nil {
			wailsRuntime.LogInfo(s.ctx, "成功解析单个远程源对象，已转换为数组格式")
		}
		// 验证并补全必需字段
		s.validateAndFixRemoteSources()
		// 立即保存为标准数组格式
		return s.saveRemoteSources()
	}

	if s.ctx != nil {
		wailsRuntime.LogError(s.ctx, fmt.Sprintf("JSON解析失败: %v", err))
	}

	// 都解析失败，初始化为空列表
	s.remoteSources = []*models.RemoteSource{}
	return nil
}

// validateAndFixRemoteSources 验证并修复远程源数据
func (s *NetworkService) validateAndFixRemoteSources() {
	for _, source := range s.remoteSources {
		// 补全缺失的必需字段
		if source.ID == "" {
			source.ID = uuid.New().String()
			if s.ctx != nil {
				wailsRuntime.LogInfo(s.ctx, fmt.Sprintf("为远程源 '%s' 生成新ID: %s", source.Name, source.ID))
			}
		}
		if source.URL == "" {
			source.Status = "failed"
			if s.ctx != nil {
				wailsRuntime.LogWarning(s.ctx, fmt.Sprintf("远程源 '%s' 缺少URL，已标记为失败状态，需要用户手动配置", source.Name))
			}
		}
		if source.UpdateFreq == "" {
			source.UpdateFreq = "manual"
		}
		if source.Status == "" {
			source.Status = "pending"
		}
		// 如果时间为零值，设置为当前时间
		if source.LastUpdatedAt.Time.IsZero() {
			source.LastUpdatedAt = models.GetCurrentTimeRFC3339()
		}
	}
}

// GetAllRemoteSources 获取所有远程源
func (s *NetworkService) GetAllRemoteSources() []*models.RemoteSource {
	if s.ctx != nil {
		wailsRuntime.LogInfo(s.ctx, fmt.Sprintf("GetAllRemoteSources: 当前远程源数量: %d", len(s.remoteSources)))
		for i, src := range s.remoteSources {
			wailsRuntime.LogInfo(s.ctx, fmt.Sprintf("远程源[%d]: ID=%s, Name=%s, Status=%s", i, src.ID, src.Name, src.Status))
		}
	}
	
	return s.remoteSources
}

// GetRemoteSourceByID 根据ID获取远程源
func (s *NetworkService) GetRemoteSourceByID(id string) (*models.RemoteSource, error) {
	return s.findRemoteSource(id)
}

// AddRemoteSource 添加远程源
func (s *NetworkService) AddRemoteSource(name, url, updateFreq string) (*models.RemoteSource, error) {
	if s.ctx != nil {
		wailsRuntime.LogInfo(s.ctx, fmt.Sprintf("AddRemoteSource: 开始添加远程源 - Name: %s, URL: %s", name, url))
	}
	
	// 验证参数
	if err := s.validateRemoteSourceParams(name, url, updateFreq); err != nil {
		return nil, err
	}

	// 创建新远程源
	newSource := &models.RemoteSource{
		ID:            uuid.New().String(),
		Name:          name,
		URL:           url,
		UpdateFreq:    updateFreq,
		LastUpdatedAt: models.GetCurrentTimeRFC3339(),
		Status:        "pending",
	}

	if s.ctx != nil {
		wailsRuntime.LogInfo(s.ctx, fmt.Sprintf("创建的新远程源ID: %s", newSource.ID))
		wailsRuntime.LogInfo(s.ctx, fmt.Sprintf("添加前远程源数量: %d", len(s.remoteSources)))
	}

	// 添加到远程源列表
	s.remoteSources = append(s.remoteSources, newSource)

	if s.ctx != nil {
		wailsRuntime.LogInfo(s.ctx, fmt.Sprintf("添加后远程源数量: %d", len(s.remoteSources)))
	}

	// 保存远程源
	err := s.saveRemoteSources()
	if err != nil {
		if s.ctx != nil {
			wailsRuntime.LogError(s.ctx, fmt.Sprintf("保存远程源失败: %v", err))
		}
		return nil, err
	}

	if s.ctx != nil {
		wailsRuntime.LogInfo(s.ctx, "远程源保存成功")
	}

	// 通知前端远程源列表已更新
	if s.ctx != nil {
		wailsRuntime.EventsEmit(s.ctx, "remote-source-list-changed")
		wailsRuntime.LogInfo(s.ctx, "已发送remote-source-list-changed事件")
	}

	return newSource, nil
}

// UpdateRemoteSource 更新远程源
func (s *NetworkService) UpdateRemoteSource(id, name, url, updateFreq string) (*models.RemoteSource, error) {
	// 验证ID
	if strings.TrimSpace(id) == "" {
		return nil, errors.New("ID不能为空")
	}
	
	// 验证其他参数
	if err := s.validateRemoteSourceParams(name, url, updateFreq); err != nil {
		return nil, err
	}

	// 查找远程源
	source, err := s.findRemoteSource(id)
	if err != nil {
		return nil, err
	}

	// 更新远程源
	source.Name = name
	source.URL = url
	source.UpdateFreq = updateFreq

	// 保存远程源
	err = s.saveRemoteSources()
	if err != nil {
		return nil, err
	}

	// 通知前端远程源列表已更新
	if s.ctx != nil {
		wailsRuntime.EventsEmit(s.ctx, "remote-source-list-changed")
	}

	return source, nil
}

// DeleteRemoteSource 删除远程源
func (s *NetworkService) DeleteRemoteSource(id string) error {
	// 查找远程源
	var sourceToDelete *models.RemoteSource
	index := -1
	for i, source := range s.remoteSources {
		if source.ID == id {
			sourceToDelete = source
			index = i
			break
		}
	}

	if index == -1 {
		return errors.New("远程源不存在")
	}

	// 从系统hosts文件中清理该远程源的内容
	currentContent, err := s.configService.ReadSystemHosts()
	if err == nil {
		// 清理指定源的内容
		cleanedContent := s.cleanOldRemoteContent(currentContent, sourceToDelete.Name)
		
		// 如果内容有变化，更新系统hosts文件
		if cleanedContent != currentContent {
			err = s.configService.WriteSystemHosts(cleanedContent)
			if err != nil {
				// 记录警告但不阻止删除操作
				if s.ctx != nil {
					wailsRuntime.LogWarning(s.ctx, "删除远程源时清理hosts文件失败: "+err.Error())
				}
			}
		}
	}

	// 删除所有与此远程源相关的配置
	configs := s.configService.GetAllConfigs()
	for _, config := range configs {
		if config.Source == "remote" && config.RemoteURL == sourceToDelete.URL {
			s.configService.DeleteConfig(config.ID)
		}
	}

	// 从远程源列表中删除
	s.remoteSources = append(s.remoteSources[:index], s.remoteSources[index+1:]...)

	// 保存远程源
	err = s.saveRemoteSources()
	if err != nil {
		return err
	}

	// 通知前端远程源列表已更新
	if s.ctx != nil {
		wailsRuntime.EventsEmit(s.ctx, "remote-source-list-changed")
		// 通知已清理系统hosts文件
		wailsRuntime.EventsEmit(s.ctx, "remote-source-cleaned-from-system", sourceToDelete.Name)
	}

	return nil
}

// FetchRemoteHosts 获取远程 hosts 内容
func (s *NetworkService) FetchRemoteHosts(id string) (string, error) {
	// 添加调试日志
	if s.ctx != nil {
		wailsRuntime.LogInfo(s.ctx, fmt.Sprintf("FetchRemoteHosts: 开始获取远程源 %s", id))
		wailsRuntime.LogInfo(s.ctx, fmt.Sprintf("当前远程源数量: %d", len(s.remoteSources)))
		for i, src := range s.remoteSources {
			wailsRuntime.LogInfo(s.ctx, fmt.Sprintf("远程源[%d]: ID=%s, Name=%s, URL=%s", i, src.ID, src.Name, src.URL))
		}
	}
	
	// 查找远程源
	source, err := s.findRemoteSource(id)
	if err != nil {
		if s.ctx != nil {
			wailsRuntime.LogError(s.ctx, fmt.Sprintf("未找到ID为 %s 的远程源: %v", id, err))
		}
		return "", fmt.Errorf("未找到远程源: %v", err)
	}

	if s.ctx != nil {
		wailsRuntime.LogInfo(s.ctx, fmt.Sprintf("找到远程源: %s (%s)", source.Name, source.URL))
	}

	// 更新状态
	source.Status = "pending"
	s.saveRemoteSources()

	// 通知前端远程源状态已更新
	if s.ctx != nil {
		wailsRuntime.EventsEmit(s.ctx, "remote-source-status-changed", id)
	}

	// 创建带有自定义User-Agent的请求
	req, err := http.NewRequest("GET", source.URL, nil)
	if err != nil {
		source.Status = "failed"
		s.saveRemoteSources()
		if s.ctx != nil {
			wailsRuntime.EventsEmit(s.ctx, "remote-source-status-changed", id)
			wailsRuntime.LogError(s.ctx, fmt.Sprintf("创建HTTP请求失败: %v", err))
		}
		return "", fmt.Errorf("创建HTTP请求失败: %v", err)
	}
	
	// 设置User-Agent和其他请求头
	req.Header.Set("User-Agent", "HostSwitcher/1.0")
	req.Header.Set("Accept", "text/plain, */*")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")

	// 发送请求
	resp, err := s.httpClient.Do(req)
	if err != nil {
		source.Status = "failed"
		s.saveRemoteSources()
		
		if s.ctx != nil {
			wailsRuntime.EventsEmit(s.ctx, "remote-source-status-changed", id)
			wailsRuntime.LogError(s.ctx, fmt.Sprintf("HTTP请求失败: %v", err))
		}
		return "", fmt.Errorf("网络请求失败: %v", err)
	}
	defer resp.Body.Close() // 确保响应体被关闭

	if resp.StatusCode != http.StatusOK {
		source.Status = "failed"
		s.saveRemoteSources()
		
		if s.ctx != nil {
			wailsRuntime.EventsEmit(s.ctx, "remote-source-status-changed", id)
			wailsRuntime.LogError(s.ctx, fmt.Sprintf("HTTP响应状态错误: %s", resp.Status))
		}
		return "", fmt.Errorf("服务器响应错误: %s", resp.Status)
	}

	// 限制读取大小，防止内存溢出
	body, err := io.ReadAll(io.LimitReader(resp.Body, 50*1024*1024)) // 限制最大读取50MB
	if err != nil {
		source.Status = "failed"
		s.saveRemoteSources()
		
		if s.ctx != nil {
			wailsRuntime.EventsEmit(s.ctx, "remote-source-status-changed", id)
			wailsRuntime.LogError(s.ctx, fmt.Sprintf("读取响应内容失败: %v", err))
		}
		return "", fmt.Errorf("读取响应内容失败: %v", err)
	}

	remoteContent := string(body)
	
	// 更新状态和内容
	source.Status = "success"
	source.LastUpdatedAt = models.GetCurrentTimeRFC3339()
	source.LastContent = remoteContent // 保存内容用于比对
	s.saveRemoteSources()

	// 通知前端远程源状态已更新
	if s.ctx != nil {
		wailsRuntime.EventsEmit(s.ctx, "remote-source-status-changed", id)
		wailsRuntime.LogInfo(s.ctx, fmt.Sprintf("成功获取远程源内容，大小: %d 字节", len(body)))
	}

	return remoteContent, nil
}

// CreateConfigFromRemote 从远程源创建配置
func (s *NetworkService) CreateConfigFromRemote(id string) (*models.Config, error) {
	// 获取远程内容
	remoteContent, err := s.FetchRemoteHosts(id)
	if err != nil {
		return nil, err
	}

	// 查找远程源
	source, err := s.findRemoteSource(id)
	if err != nil {
		return nil, err
	}

	// 获取当前系统hosts内容
	currentContent, err := s.configService.ReadSystemHosts()
	if err != nil {
		// 如果无法读取当前hosts，仅使用远程内容
		currentContent = ""
	}

	// 合并内容：保留当前hosts内容，并追加远程内容
	mergedContent := s.mergeHostsContent(currentContent, remoteContent, source.Name)

	// 创建配置
	config, err := s.configService.CreateConfig(
		source.Name+" (远程)",
		"从 "+source.URL+" 获取的远程配置，已合并到本地hosts",
		mergedContent,
	)
	if err != nil {
		return nil, err
	}

	// 更新配置来源
	err = s.configService.UpdateConfigSource(config.ID, "remote", source.URL)
	if err != nil {
		return nil, err
	}

	// 重新获取更新后的配置
	config, err = s.configService.GetConfigByID(config.ID)
	if err != nil {
		return nil, err
	}

	return config, nil
}

// UpdateConfigFromRemote 从远程源更新配置
func (s *NetworkService) UpdateConfigFromRemote(configID string) (*models.Config, error) {
	// 查找配置
	config, err := s.configService.GetConfigByID(configID)
	if err != nil {
		return nil, err
	}

	// 检查配置来源
	if config.Source != "remote" || config.RemoteURL == "" {
		return nil, errors.New("配置不是远程配置")
	}

	// 查找对应的远程源
	var source *models.RemoteSource
	for _, src := range s.remoteSources {
		if src.URL == config.RemoteURL {
			source = src
			break
		}
	}

	var remoteContent string
	var sourceName string

	if source != nil {
		// 使用现有的远程源获取内容
		remoteContent, err = s.FetchRemoteHosts(source.ID)
		if err != nil {
			return nil, err
		}
		sourceName = source.Name
	} else {
		// 直接从URL获取内容（远程源可能已被删除）
		resp, err := s.httpClient.Get(config.RemoteURL)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, errors.New("远程服务器返回错误: " + resp.Status)
		}

		body, err := io.ReadAll(io.LimitReader(resp.Body, 10*1024*1024))
		if err != nil {
			return nil, err
		}
		
		remoteContent = string(body)
		sourceName = "远程源"
	}

	// 获取当前系统hosts内容
	currentContent, err := s.configService.ReadSystemHosts()
	if err != nil {
		// 如果无法读取当前hosts，使用配置中的内容
		currentContent = config.Content
	}

	// 合并内容
	mergedContent := s.mergeHostsContent(currentContent, remoteContent, sourceName)

	// 更新配置
	config, err = s.configService.UpdateConfig(
		config.ID,
		config.Name,
		config.Description,
		mergedContent,
	)
	if err != nil {
		return nil, err
	}

	// 更新配置来源
	err = s.configService.UpdateConfigSource(config.ID, "remote", config.RemoteURL)
	if err != nil {
		return nil, err
	}

	// 重新获取更新后的配置
	config, err = s.configService.GetConfigByID(config.ID)
	if err != nil {
		return nil, err
	}

	return config, nil
}

// mergeHostsContent 合并hosts内容
// 将远程内容追加到当前内容，并添加分隔注释
func (s *NetworkService) mergeHostsContent(currentContent, remoteContent, sourceName string) string {
	// 确保当前内容以换行符结尾
	if currentContent != "" && !strings.HasSuffix(currentContent, "\n") {
		currentContent += "\n"
	}

	// 首先清理该源名称的所有旧内容
	currentContent = s.cleanOldRemoteContent(currentContent, sourceName)

	// 添加分隔注释
	separator := "\n\n# ===== 以下是从 " + sourceName + " 获取的远程内容 =====\n"
	endSeparator := "\n# ===== " + sourceName + " 远程内容结束 =====\n\n"
	
	// 追加新的远程内容
	currentContent += separator + remoteContent + endSeparator

	return currentContent
}

// cleanOldRemoteContent 清理指定源名称的旧远程内容
func (s *NetworkService) cleanOldRemoteContent(content, sourceName string) string {
	separator := "\n\n# ===== 以下是从 " + sourceName + " 获取的远程内容 =====\n"
	endSeparator := "\n# ===== " + sourceName + " 远程内容结束 =====\n\n"
	
	for {
		startIndex := strings.Index(content, separator)
		if startIndex == -1 {
			break // 没有找到该源的内容
		}
		
		endIndex := strings.Index(content[startIndex:], endSeparator)
		if endIndex == -1 {
			// 没找到结束标记，尝试查找下一个开始标记
			nextSeparatorIndex := strings.Index(content[startIndex+len(separator):], "\n\n# ===== 以下是从 ")
			if nextSeparatorIndex > 0 {
				nextSeparatorIndex = startIndex + len(separator) + nextSeparatorIndex
				content = content[:startIndex] + content[nextSeparatorIndex:]
			} else {
				// 没有下一个分隔符，删除到末尾
				content = content[:startIndex]
			}
		} else {
			// 找到了完整的内容块，删除它
			endIndex = startIndex + endIndex + len(endSeparator)
			content = content[:startIndex] + content[endIndex:]
		}
	}
	
	return content
}

// saveRemoteSources 保存远程源
func (s *NetworkService) saveRemoteSources() error {
	// 序列化远程源
	data, err := json.MarshalIndent(s.remoteSources, "", "  ")
	if err != nil {
		return err
	}

	// 写入远程源文件
	err = os.WriteFile(s.remoteFile, data, 0644)
	if err != nil {
		return err
	}

	return nil
}

// UpdateAllRemoteSources 手动更新所有远程源
func (s *NetworkService) UpdateAllRemoteSources() error {
	if s.ctx != nil {
		wailsRuntime.LogInfo(s.ctx, "手动触发更新所有远程源")
	}
	
	for _, source := range s.remoteSources {
		if s.ctx != nil {
			wailsRuntime.LogInfo(s.ctx, fmt.Sprintf("正在更新远程源: %s", source.Name))
		}
		
		// 直接应用到系统hosts文件
		err := s.ApplyRemoteToSystem(source.ID)
		if err != nil {
			if s.ctx != nil {
				wailsRuntime.LogError(s.ctx, fmt.Sprintf("更新远程源 %s 失败: %v", source.Name, err))
			}
		} else {
			if s.ctx != nil {
				wailsRuntime.LogInfo(s.ctx, fmt.Sprintf("成功更新远程源: %s", source.Name))
			}
		}
	}
	
	return nil
}

// Cleanup 清理资源
func (s *NetworkService) Cleanup() {
	// 关闭HTTP客户端连接
	if s.httpClient != nil && s.httpClient.Transport != nil {
		if transport, ok := s.httpClient.Transport.(*http.Transport); ok {
			transport.CloseIdleConnections()
		}
	}
	
	// 清空内存中的远程源，帮助GC回收
	s.remoteSources = nil
}

// ApplyRemoteToSystem 直接将远程hosts内容应用到系统hosts文件
func (s *NetworkService) ApplyRemoteToSystem(id string) error {
	// 获取远程内容
	remoteContent, err := s.FetchRemoteHosts(id)
	if err != nil {
		return fmt.Errorf("获取远程内容失败: %v", err)
	}

	// 查找远程源
	source, err := s.findRemoteSource(id)
	if err != nil {
		return err
	}

	// 获取当前系统hosts内容
	currentContent, err := s.configService.ReadSystemHosts()
	if err != nil {
		return fmt.Errorf("读取系统hosts文件失败: %v", err)
	}

	// 合并内容：保留当前hosts内容，并追加远程内容
	mergedContent := s.mergeHostsContent(currentContent, remoteContent, source.Name)

	// 验证合并后的内容
	err = s.configService.ValidateHostsContent(mergedContent)
	if err != nil {
		return fmt.Errorf("合并后的hosts内容格式无效: %v", err)
	}

	// 直接写入系统hosts文件
	err = s.configService.WriteSystemHosts(mergedContent)
	if err != nil {
		return fmt.Errorf("写入系统hosts文件失败: %v", err)
	}

	// 通知前端
	if s.ctx != nil {
		wailsRuntime.EventsEmit(s.ctx, "remote-applied-to-system", source.Name)
	}

	return nil
}

// updateStartupSources 更新启动时自动更新的远程源
func (s *NetworkService) updateStartupSources() {
	if s.ctx != nil {
		wailsRuntime.LogInfo(s.ctx, "开始检查启动时自动更新的远程源")
	}
	
	// 检查是否有远程源
	if len(s.remoteSources) == 0 {
		if s.ctx != nil {
			wailsRuntime.LogInfo(s.ctx, "没有配置任何远程源，跳过启动时更新")
		}
		return
	}
	
	for _, source := range s.remoteSources {
		if source.UpdateFreq == "startup" {
			if s.ctx != nil {
				wailsRuntime.LogInfo(s.ctx, fmt.Sprintf("正在检查启动源: %s (URL: %s)", source.Name, source.URL))
			}
			
			// 获取最新内容
			newContent, err := s.FetchRemoteHosts(source.ID)
			if err != nil {
				if s.ctx != nil {
					wailsRuntime.LogError(s.ctx, fmt.Sprintf("启动时获取远程源 %s 失败: %v", source.Name, err))
				}
				// 更新状态为失败
				source.Status = "failed"
				s.saveRemoteSources()
				continue
			}
			
			// 比对内容是否有变化
			if source.LastContent != "" && source.LastContent == newContent {
				if s.ctx != nil {
					wailsRuntime.LogInfo(s.ctx, fmt.Sprintf("远程源 %s 内容无变化，跳过更新", source.Name))
				}
				continue
			}
			
			if s.ctx != nil {
				wailsRuntime.LogInfo(s.ctx, fmt.Sprintf("远程源 %s 内容有变化，正在应用更新", source.Name))
			}
			
			// 内容有变化，直接应用到系统hosts文件
			err = s.ApplyRemoteToSystem(source.ID)
			if err != nil {
				if s.ctx != nil {
					wailsRuntime.LogError(s.ctx, fmt.Sprintf("启动时更新远程源 %s 失败: %v", source.Name, err))
				}
				// 更新状态为失败
				source.Status = "failed"
				s.saveRemoteSources()
			} else {
				if s.ctx != nil {
					wailsRuntime.LogInfo(s.ctx, fmt.Sprintf("启动时成功更新远程源: %s", source.Name))
				}
				// 更新状态为成功
				source.Status = "success"
				source.LastUpdatedAt = models.GetCurrentTimeRFC3339()
				s.saveRemoteSources()
			}
		}
	}
	
	// 通知前端更新状态
	if s.ctx != nil {
		wailsRuntime.EventsEmit(s.ctx, "startup-sources-updated")
	}
}
