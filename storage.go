package gohl

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// StorageItem 存储项，包含值和过期时间
type StorageItem struct {
	Value      interface{}
	ExpireAt   time.Time
	HasTTL     bool
}

// IsExpired 检查是否已过期
func (item *StorageItem) IsExpired() bool {
	if !item.HasTTL {
		return false
	}
	return time.Now().After(item.ExpireAt)
}

// Storage 带TTL的本地KV存储
type Storage struct {
	data     map[string]*StorageItem
	mu       sync.RWMutex
	filePath string
	dirty    bool
}

// NewStorage 创建新的存储实例
func NewStorage(appName string) (*Storage, error) {
	appDataDir := os.Getenv("APPDATA")
	if appDataDir == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("无法获取用户目录: %w", err)
		}
		appDataDir = filepath.Join(homeDir, "AppData", "Roaming")
	}

	dir := filepath.Join(appDataDir, appName)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("无法创建目录: %w", err)
	}

	filePath := filepath.Join(dir, "storage.dat")
	s := &Storage{
		data:     make(map[string]*StorageItem),
		filePath: filePath,
		dirty:    false,
	}

	// 尝试加载已有数据
	if err := s.Load(); err != nil {
		// 文件不存在是正常现象，忽略错误
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("加载数据失败: %w", err)
		}
	}

	return s, nil
}

// NewStorageWithPath 使用指定路径创建存储实例
func NewStorageWithPath(filePath string) (*Storage, error) {
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("无法创建目录: %w", err)
	}

	s := &Storage{
		data:     make(map[string]*StorageItem),
		filePath: filePath,
		dirty:    false,
	}

	if err := s.Load(); err != nil {
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("加载数据失败: %w", err)
		}
	}

	return s, nil
}

// Set 设置键值对（无TTL）
func (s *Storage) Set(key string, value interface{}) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.data[key] = &StorageItem{
		Value:  value,
		HasTTL: false,
	}
	s.dirty = true

	return s.saveLocked()
}

// SetWithTTL 设置带TTL的键值对
func (s *Storage) SetWithTTL(key string, value interface{}, ttl time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.data[key] = &StorageItem{
		Value:    value,
		ExpireAt: time.Now().Add(ttl),
		HasTTL:   true,
	}
	s.dirty = true

	return s.saveLocked()
}

// Get 获取值
func (s *Storage) Get(key string) (interface{}, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	item, exists := s.data[key]
	if !exists {
		return nil, false
	}

	// 检查是否过期
	if item.IsExpired() {
		return nil, false
	}

	return item.Value, true
}

// GetString 获取字符串值
func (s *Storage) GetString(key string) (string, bool) {
	value, exists := s.Get(key)
	if !exists {
		return "", false
	}

	str, ok := value.(string)
	return str, ok
}

// GetInt 获取整数值
func (s *Storage) GetInt(key string) (int64, bool) {
	value, exists := s.Get(key)
	if !exists {
		return 0, false
	}

	switch v := value.(type) {
	case int:
		return int64(v), true
	case int8:
		return int64(v), true
	case int16:
		return int64(v), true
	case int32:
		return int64(v), true
	case int64:
		return v, true
	case uint:
		return int64(v), true
	case uint8:
		return int64(v), true
	case uint16:
		return int64(v), true
	case uint32:
		return int64(v), true
	case uint64:
		return int64(v), true
	case float32:
		return int64(v), true
	case float64:
		return int64(v), true
	default:
		return 0, false
	}
}

// GetBool 获取布尔值
func (s *Storage) GetBool(key string) (bool, bool) {
	value, exists := s.Get(key)
	if !exists {
		return false, false
	}

	b, ok := value.(bool)
	return b, ok
}

// Delete 删除键
func (s *Storage) Delete(key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.data, key)
	s.dirty = true

	return s.saveLocked()
}

// Exists 检查键是否存在（未过期）
func (s *Storage) Exists(key string) bool {
	_, exists := s.Get(key)
	return exists
}

// Keys 获取所有未过期的键
func (s *Storage) Keys() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	keys := make([]string, 0, len(s.data))
	for key, item := range s.data {
		if !item.IsExpired() {
			keys = append(keys, key)
		}
	}

	return keys
}

// TTL 获取键的剩余生存时间
func (s *Storage) TTL(key string) (time.Duration, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	item, exists := s.data[key]
	if !exists || !item.HasTTL {
		return 0, false
	}

	if item.IsExpired() {
		return 0, false
	}

	return item.ExpireAt.Sub(time.Now()), true
}

// Expire 为键设置过期时间
func (s *Storage) Expire(key string, ttl time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	item, exists := s.data[key]
	if !exists {
		return fmt.Errorf("key not found: %s", key)
	}

	item.ExpireAt = time.Now().Add(ttl)
	item.HasTTL = true
	s.dirty = true

	return s.saveLocked()
}

// Persist 移除键的过期时间
func (s *Storage) Persist(key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	item, exists := s.data[key]
	if !exists {
		return fmt.Errorf("key not found: %s", key)
	}

	item.HasTTL = false
	s.dirty = true

	return s.saveLocked()
}

// Clear 清空所有数据
func (s *Storage) Clear() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.data = make(map[string]*StorageItem)
	s.dirty = true

	return s.saveLocked()
}

// CleanExpired 清理过期数据
func (s *Storage) CleanExpired() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	changed := false
	for key, item := range s.data {
		if item.IsExpired() {
			delete(s.data, key)
			changed = true
		}
	}

	if changed {
		s.dirty = true
		return s.saveLocked()
	}

	return nil
}

// Save 保存数据到文件
func (s *Storage) Save() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.saveLocked()
}

// saveLocked 内部保存方法（已加锁）
func (s *Storage) saveLocked() error {
	if !s.dirty {
		return nil
	}

	// 清理过期数据
	for key, item := range s.data {
		if item.IsExpired() {
			delete(s.data, key)
		}
	}

	var buf bytes.Buffer
	encoder := gob.NewEncoder(&buf)

	if err := encoder.Encode(s.data); err != nil {
		return fmt.Errorf("编码数据失败: %w", err)
	}

	// 写入临时文件，然后重命名，保证原子性
	tempFile := s.filePath + ".tmp"
	if err := os.WriteFile(tempFile, buf.Bytes(), 0644); err != nil {
		return fmt.Errorf("写入临时文件失败: %w", err)
	}

	if err := os.Rename(tempFile, s.filePath); err != nil {
		os.Remove(tempFile)
		return fmt.Errorf("重命名文件失败: %w", err)
	}

	s.dirty = false
	return nil
}

// Load 从文件加载数据
func (s *Storage) Load() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := os.ReadFile(s.filePath)
	if err != nil {
		return err
	}

	decoder := gob.NewDecoder(bytes.NewReader(data))
	if err := decoder.Decode(&s.data); err != nil {
		return fmt.Errorf("解码数据失败: %w", err)
	}

	// 清理过期数据
	for key, item := range s.data {
		if item.IsExpired() {
			delete(s.data, key)
			s.dirty = true
		}
	}

	return nil
}

// Close 关闭存储，保存数据
func (s *Storage) Close() error {
	return s.Save()
}

// FilePath 获取存储文件路径
func (s *Storage) FilePath() string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.filePath
}
