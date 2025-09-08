package models

import (
	"encoding/json"
	"fmt"
	"time"
)

// Config 表示一个 hosts 配置
type Config struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Content     string   `json:"content"`
	IsActive    bool     `json:"is_active"`
	Source      string   `json:"source"`
	RemoteURL   string   `json:"remoteUrl,omitempty"`
	CreatedAt   JSONTime `json:"created_at"`
	UpdatedAt   JSONTime `json:"updated_at"`
}

// RemoteSource 表示一个远程 hosts 源
type RemoteSource struct {
	ID            string   `json:"id"`
	Name          string   `json:"name"`
	URL           string   `json:"url"`
	UpdateFreq    string   `json:"updateFreq"`
	LastUpdatedAt JSONTime `json:"lastUpdatedAt"`
	LastContent   string   `json:"lastContent,omitempty"`
	Status        string   `json:"status"`
}

// JSONTime 用于JSON序列化的时间类型
type JSONTime struct {
	time.Time
}

// MarshalJSON 实现JSON序列化
func (jt JSONTime) MarshalJSON() ([]byte, error) {
	return json.Marshal(jt.Time.Format(time.RFC3339))
}

// UnmarshalJSON 实现JSON反序列化
func (jt *JSONTime) UnmarshalJSON(data []byte) error {
	var timeStr string
	err := json.Unmarshal(data, &timeStr)
	if err != nil {
		return err
	}

	// 处理空字符串
	if timeStr == "" {
		jt.Time = time.Time{}
		return nil
	}

	// 尝试多种时间格式
	formats := []string{
		time.RFC3339,                // 2006-01-02T15:04:05Z07:00
		time.RFC3339Nano,            // 2006-01-02T15:04:05.999999999Z07:00
		"2006-01-02T15:04:05+07:00", // 支持用户数据中的时区格式 +08:00
		"2006-01-02T15:04:05-07:00", // 支持负时区
		"2006-01-02T15:04:05Z",      // UTC时间
		"2006-01-02T15:04:05",       // 无时区
		"2006-01-02 15:04:05",       // 空格分隔
	}

	var lastErr error
	for _, format := range formats {
		if t, err := time.Parse(format, timeStr); err == nil {
			jt.Time = t
			return nil
		} else {
			lastErr = err
		}
	}

	// 如果所有格式都失败，返回最后一个错误
	return fmt.Errorf("无法解析时间格式 '%s': %v", timeStr, lastErr)
}

// GetCurrentTimeRFC3339 获取当前时间的RFC3339格式
func GetCurrentTimeRFC3339() JSONTime {
	return JSONTime{Time: time.Now()}
}

// Backup 表示一个备份记录
type Backup struct {
	ID          string   `json:"id"`
	Timestamp   JSONTime `json:"timestamp"`
	Description string   `json:"description"`
	Content     string   `json:"content"`
	Size        int64    `json:"size"`
	IsAutomatic bool     `json:"isAutomatic"`
	Tags        []string `json:"tags,omitempty"`
	Hash        string   `json:"hash"` // 内容哈希，用于去重
}

// BackupList 备份列表的包装器，便于序列化
type BackupList struct {
	Backups []*Backup `json:"backups"`
}
