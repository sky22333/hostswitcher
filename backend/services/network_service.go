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

// NetworkService 网络
type NetworkService struct {
	ctx           context.Context
	configService *ConfigService
	remoteSources []*models.RemoteSource
	remoteFile    string
	httpClient    *http.Client
}

// NewNetworkService 创建服务
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

// SetContext 设置ctx
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

// findRemoteSource 查找源
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

// validateRemoteSourceParams 验证参数
func (s *NetworkService) validateRemoteSourceParams(name, url, updateFreq string) error {
	if strings.TrimSpace(name) == "" {
		return errors.New("名称不能为空")
	}
	if strings.TrimSpace(url) == "" {
		return errors.New("URL不能为空")
	}
	

	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		return errors.New("URL必须以http://或https://开头")
	}


	if updateFreq != "manual" && updateFreq != "startup" {
		return errors.New("更新频率必须是manual或startup")
	}
	
	return nil
}

// loadRemoteSources 加载源
func (s *NetworkService) loadRemoteSources() error {

	if _, err := os.Stat(s.remoteFile); os.IsNotExist(err) {

		s.remoteSources = []*models.RemoteSource{}
		return nil
	}


	data, err := os.ReadFile(s.remoteFile)
	if err != nil {
		return err
	}


	if len(data) == 0 {
		s.remoteSources = []*models.RemoteSource{}
		return nil
	}


	var sourcesArray []*models.RemoteSource
	err = json.Unmarshal(data, &sourcesArray)
	if err == nil {

		s.remoteSources = sourcesArray
		if s.remoteSources == nil {
			s.remoteSources = []*models.RemoteSource{}
		}

		s.validateAndFixRemoteSources()
		return nil
	}

	if s.ctx != nil {
		wailsRuntime.LogInfo(s.ctx, "尝试解析为数组格式失败，尝试解析为单个对象格式")
	}


	var singleSource models.RemoteSource
	err = json.Unmarshal(data, &singleSource)
	if err == nil {

		s.remoteSources = []*models.RemoteSource{&singleSource}
		if s.ctx != nil {
			wailsRuntime.LogInfo(s.ctx, "成功解析单个远程源对象，已转换为数组格式")
		}

		s.validateAndFixRemoteSources()

		return s.saveRemoteSources()
	}

	if s.ctx != nil {
		wailsRuntime.LogError(s.ctx, fmt.Sprintf("JSON解析失败: %v", err))
	}


	s.remoteSources = []*models.RemoteSource{}
	return nil
}

// validateAndFixRemoteSources 验证修复
func (s *NetworkService) validateAndFixRemoteSources() {
	for _, source := range s.remoteSources {

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

		if source.LastUpdatedAt.Time.IsZero() {
			source.LastUpdatedAt = models.GetCurrentTimeRFC3339()
		}
	}
}

// GetAllRemoteSources 获取源
func (s *NetworkService) GetAllRemoteSources() []*models.RemoteSource {
	if s.ctx != nil {
		wailsRuntime.LogInfo(s.ctx, fmt.Sprintf("GetAllRemoteSources: 当前远程源数量: %d", len(s.remoteSources)))
		for i, src := range s.remoteSources {
			wailsRuntime.LogInfo(s.ctx, fmt.Sprintf("远程源[%d]: ID=%s, Name=%s, Status=%s", i, src.ID, src.Name, src.Status))
		}
	}
	
	return s.remoteSources
}

// GetRemoteSourceByID 获取源
func (s *NetworkService) GetRemoteSourceByID(id string) (*models.RemoteSource, error) {
	return s.findRemoteSource(id)
}

// AddRemoteSource 添加源
func (s *NetworkService) AddRemoteSource(name, url, updateFreq string) (*models.RemoteSource, error) {
	if s.ctx != nil {
		wailsRuntime.LogInfo(s.ctx, fmt.Sprintf("AddRemoteSource: 开始添加远程源 - Name: %s, URL: %s", name, url))
	}
	

	if err := s.validateRemoteSourceParams(name, url, updateFreq); err != nil {
		return nil, err
	}


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


	s.remoteSources = append(s.remoteSources, newSource)

	if s.ctx != nil {
		wailsRuntime.LogInfo(s.ctx, fmt.Sprintf("添加后远程源数量: %d", len(s.remoteSources)))
	}


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


	if s.ctx != nil {
		wailsRuntime.EventsEmit(s.ctx, "remote-source-list-changed")
		wailsRuntime.LogInfo(s.ctx, "已发送remote-source-list-changed事件")
	}

	return newSource, nil
}

// UpdateRemoteSource 更新源
func (s *NetworkService) UpdateRemoteSource(id, name, url, updateFreq string) (*models.RemoteSource, error) {

	if strings.TrimSpace(id) == "" {
		return nil, errors.New("ID不能为空")
	}
	

	if err := s.validateRemoteSourceParams(name, url, updateFreq); err != nil {
		return nil, err
	}


	source, err := s.findRemoteSource(id)
	if err != nil {
		return nil, err
	}


	source.Name = name
	source.URL = url
	source.UpdateFreq = updateFreq


	err = s.saveRemoteSources()
	if err != nil {
		return nil, err
	}


	if s.ctx != nil {
		wailsRuntime.EventsEmit(s.ctx, "remote-source-list-changed")
	}

	return source, nil
}

// DeleteRemoteSource 删除源
func (s *NetworkService) DeleteRemoteSource(id string) error {

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


	currentContent, err := s.configService.ReadSystemHosts()
	if err == nil {

		cleanedContent := s.cleanOldRemoteContent(currentContent, sourceToDelete.Name)
		

		if cleanedContent != currentContent {
			err = s.configService.WriteSystemHosts(cleanedContent)
			if err != nil {
	
				if s.ctx != nil {
					wailsRuntime.LogWarning(s.ctx, "删除远程源时清理hosts文件失败: "+err.Error())
				}
			}
		}
	}


	configs := s.configService.GetAllConfigs()
	for _, config := range configs {
		if config.Source == "remote" && config.RemoteURL == sourceToDelete.URL {
			s.configService.DeleteConfig(config.ID)
		}
	}


	s.remoteSources = append(s.remoteSources[:index], s.remoteSources[index+1:]...)


	err = s.saveRemoteSources()
	if err != nil {
		return err
	}


	if s.ctx != nil {
		wailsRuntime.EventsEmit(s.ctx, "remote-source-list-changed")

		wailsRuntime.EventsEmit(s.ctx, "remote-source-cleaned-from-system", sourceToDelete.Name)
	}

	return nil
}

// FetchRemoteHosts 获取内容
func (s *NetworkService) FetchRemoteHosts(id string) (string, error) {

	if s.ctx != nil {
		wailsRuntime.LogInfo(s.ctx, fmt.Sprintf("FetchRemoteHosts: 开始获取远程源 %s", id))
		wailsRuntime.LogInfo(s.ctx, fmt.Sprintf("当前远程源数量: %d", len(s.remoteSources)))
		for i, src := range s.remoteSources {
			wailsRuntime.LogInfo(s.ctx, fmt.Sprintf("远程源[%d]: ID=%s, Name=%s, URL=%s", i, src.ID, src.Name, src.URL))
		}
	}
	

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


	source.Status = "pending"
	s.saveRemoteSources()


	if s.ctx != nil {
		wailsRuntime.EventsEmit(s.ctx, "remote-source-status-changed", id)
	}


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
	

	req.Header.Set("User-Agent", "HostSwitcher/1.0")
	req.Header.Set("Accept", "text/plain, */*")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")


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
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		source.Status = "failed"
		s.saveRemoteSources()
		
		if s.ctx != nil {
			wailsRuntime.EventsEmit(s.ctx, "remote-source-status-changed", id)
			wailsRuntime.LogError(s.ctx, fmt.Sprintf("HTTP响应状态错误: %s", resp.Status))
		}
		return "", fmt.Errorf("服务器响应错误: %s", resp.Status)
	}


	body, err := io.ReadAll(io.LimitReader(resp.Body, 50*1024*1024))
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
	

	source.Status = "success"
	source.LastUpdatedAt = models.GetCurrentTimeRFC3339()
	source.LastContent = remoteContent
	s.saveRemoteSources()


	if s.ctx != nil {
		wailsRuntime.EventsEmit(s.ctx, "remote-source-status-changed", id)
		wailsRuntime.LogInfo(s.ctx, fmt.Sprintf("成功获取远程源内容，大小: %d 字节", len(body)))
	}

	return remoteContent, nil
}

// CreateConfigFromRemote 创建配置
func (s *NetworkService) CreateConfigFromRemote(id string) (*models.Config, error) {

	remoteContent, err := s.FetchRemoteHosts(id)
	if err != nil {
		return nil, err
	}

	source, err := s.findRemoteSource(id)
	if err != nil {
		return nil, err
	}

	currentContent, err := s.configService.ReadSystemHosts()
	if err != nil {
		currentContent = ""
	}

	mergedContent := s.mergeHostsContent(currentContent, remoteContent, source.Name)

	config, err := s.configService.CreateConfig(
		source.Name+" (远程)",
		"从 "+source.URL+" 获取的远程配置，已合并到本地hosts",
		mergedContent,
	)
	if err != nil {
		return nil, err
	}

	err = s.configService.UpdateConfigSource(config.ID, "remote", source.URL)
	if err != nil {
		return nil, err
	}

	config, err = s.configService.GetConfigByID(config.ID)
	if err != nil {
		return nil, err
	}

	return config, nil
}

// UpdateConfigFromRemote 更新配置
func (s *NetworkService) UpdateConfigFromRemote(configID string) (*models.Config, error) {
	config, err := s.configService.GetConfigByID(configID)
	if err != nil {
		return nil, err
	}

	if config.Source != "remote" || config.RemoteURL == "" {
		return nil, errors.New("配置不是远程配置")
	}

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
		remoteContent, err = s.FetchRemoteHosts(source.ID)
		if err != nil {
			return nil, err
		}
		sourceName = source.Name
	} else {
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

	currentContent, err := s.configService.ReadSystemHosts()
	if err != nil {
		currentContent = config.Content
	}

	mergedContent := s.mergeHostsContent(currentContent, remoteContent, sourceName)

	config, err = s.configService.UpdateConfig(
		config.ID,
		config.Name,
		config.Description,
		mergedContent,
	)
	if err != nil {
		return nil, err
	}

	err = s.configService.UpdateConfigSource(config.ID, "remote", config.RemoteURL)
	if err != nil {
		return nil, err
	}

	config, err = s.configService.GetConfigByID(config.ID)
	if err != nil {
		return nil, err
	}

	return config, nil
}

// mergeHostsContent 合并内容
func (s *NetworkService) mergeHostsContent(currentContent, remoteContent, sourceName string) string {
	if currentContent != "" && !strings.HasSuffix(currentContent, "\n") {
		currentContent += "\n"
	}

	currentContent = s.cleanOldRemoteContent(currentContent, sourceName)
	separator := "\n\n# ===== 以下是从 " + sourceName + " 获取的远程内容 =====\n"
	endSeparator := "\n# ===== " + sourceName + " 远程内容结束 =====\n\n"
	
	currentContent += separator + remoteContent + endSeparator

	return currentContent
}

// cleanOldRemoteContent 清理内容
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

// saveRemoteSources 保存源
func (s *NetworkService) saveRemoteSources() error {
	data, err := json.MarshalIndent(s.remoteSources, "", "  ")
	if err != nil {
		return err
	}

	err = os.WriteFile(s.remoteFile, data, 0644)
	if err != nil {
		return err
	}

	return nil
}

// UpdateAllRemoteSources 更新所有源
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

// Cleanup 清理
func (s *NetworkService) Cleanup() {
	if s.httpClient != nil && s.httpClient.Transport != nil {
		if transport, ok := s.httpClient.Transport.(*http.Transport); ok {
			transport.CloseIdleConnections()
		}
	}
	
	s.remoteSources = nil
}

// ApplyRemoteToSystem 应用到系统
func (s *NetworkService) ApplyRemoteToSystem(id string) error {
	remoteContent, err := s.FetchRemoteHosts(id)
	if err != nil {
		return fmt.Errorf("获取远程内容失败: %v", err)
	}

	source, err := s.findRemoteSource(id)
	if err != nil {
		return err
	}

	currentContent, err := s.configService.ReadSystemHosts()
	if err != nil {
		return fmt.Errorf("读取系统hosts文件失败: %v", err)
	}

	mergedContent := s.mergeHostsContent(currentContent, remoteContent, source.Name)

	err = s.configService.ValidateHostsContent(mergedContent)
	if err != nil {
		return fmt.Errorf("合并后的hosts内容格式无效: %v", err)
	}

	err = s.configService.WriteSystemHosts(mergedContent)
	if err != nil {
		return fmt.Errorf("写入系统hosts文件失败: %v", err)
	}

	if s.ctx != nil {
		wailsRuntime.EventsEmit(s.ctx, "remote-applied-to-system", source.Name)
	}

	return nil
}

// updateStartupSources 更新启动源
func (s *NetworkService) updateStartupSources() {
	if s.ctx != nil {
		wailsRuntime.LogInfo(s.ctx, "开始检查启动时自动更新的远程源")
	}
	
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
			
			newContent, err := s.FetchRemoteHosts(source.ID)
			if err != nil {
				if s.ctx != nil {
					wailsRuntime.LogError(s.ctx, fmt.Sprintf("启动时获取远程源 %s 失败: %v", source.Name, err))
				}
				source.Status = "failed"
				s.saveRemoteSources()
				continue
			}
			
			if source.LastContent != "" && source.LastContent == newContent {
				if s.ctx != nil {
					wailsRuntime.LogInfo(s.ctx, fmt.Sprintf("远程源 %s 内容无变化，跳过更新", source.Name))
				}
				continue
			}
			
			if s.ctx != nil {
				wailsRuntime.LogInfo(s.ctx, fmt.Sprintf("远程源 %s 内容有变化，正在应用更新", source.Name))
			}
			
			err = s.ApplyRemoteToSystem(source.ID)
			if err != nil {
				if s.ctx != nil {
					wailsRuntime.LogError(s.ctx, fmt.Sprintf("启动时更新远程源 %s 失败: %v", source.Name, err))
				}
				source.Status = "failed"
				s.saveRemoteSources()
			} else {
				if s.ctx != nil {
					wailsRuntime.LogInfo(s.ctx, fmt.Sprintf("启动时成功更新远程源: %s", source.Name))
				}
				source.Status = "success"
				source.LastUpdatedAt = models.GetCurrentTimeRFC3339()
				s.saveRemoteSources()
			}
		}
	}
	
	if s.ctx != nil {
		wailsRuntime.EventsEmit(s.ctx, "startup-sources-updated")
	}
}
