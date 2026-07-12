// Package memory 提供三层记忆系统。
//
// Manager 是统一入口，各层可以独立实现和测试。
// 内置 InMemoryManager 用于测试。
package memory

import (
	"context"
	"fmt"
	"strings"
	"sync"
)

// Message 表示一条对话消息，用于 L1 短期记忆。
type Message struct {
	Role    string // system / user / assistant / tool
	Content string
}

// Fact 表示一条事实记录，用于 L2 核心和 L3 长期记忆。
type Fact struct {
	Key        string
	Value      string
	Category   string // preference / fact / event
	Importance int    // 1-5
	Timestamp  int64
}

// Manager 是记忆系统的统一接口。
type Manager interface {
	// L1 短期记忆
	AddShortTerm(msg Message)
	GetShortTerm() []Message
	ClearShortTerm()

	// L2 核心记忆
	Remember(key, value string, importance int) error
	Recall(key string) (string, error)
	GetAllCore() (map[string]string, error)

	// L3 长期记忆
	Store(ctx context.Context, fact Fact) error
	Search(ctx context.Context, query string, limit int) ([]Fact, error)
}

// Ensure InMemoryManager implements Manager.
var _ Manager = (*InMemoryManager)(nil)

// InMemoryManager 是 Manager 的内存实现，用于测试。
type InMemoryManager struct {
	mu        sync.RWMutex
	shortTerm []Message
	core      map[string]Fact
	longTerm  []Fact
}

// NewInMemoryManager 创建一个空的 InMemoryManager 实例。
func NewInMemoryManager() *InMemoryManager {
	return &InMemoryManager{
		core: make(map[string]Fact),
	}
}

// AddShortTerm 添加一条消息到短期记忆。
func (m *InMemoryManager) AddShortTerm(msg Message) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.shortTerm = append(m.shortTerm, msg)
}

// GetShortTerm 返回短期记忆中的所有消息。
func (m *InMemoryManager) GetShortTerm() []Message {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]Message, len(m.shortTerm))
	copy(out, m.shortTerm)
	return out
}

// ClearShortTerm 清空短期记忆。
func (m *InMemoryManager) ClearShortTerm() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.shortTerm = nil
}

// Remember 保存一条核心记忆。
func (m *InMemoryManager) Remember(key, value string, importance int) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.core[key] = Fact{Key: key, Value: value, Category: "core", Importance: importance}
	return nil
}

// Recall 根据键名检索核心记忆。
func (m *InMemoryManager) Recall(key string) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	fact, ok := m.core[key]
	if !ok {
		return "", fmt.Errorf("memory: key %q not found", key)
	}
	return fact.Value, nil
}

// GetAllCore 返回所有核心记忆的键值对。
func (m *InMemoryManager) GetAllCore() (map[string]string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make(map[string]string, len(m.core))
	for k, v := range m.core {
		out[k] = v.Value
	}
	return out, nil
}

// Store 保存一条长期记忆事实。
func (m *InMemoryManager) Store(_ context.Context, fact Fact) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.longTerm = append(m.longTerm, fact)
	return nil
}

// Search 根据查询字符串搜索长期记忆。
func (m *InMemoryManager) Search(_ context.Context, query string, limit int) ([]Fact, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	// 简单实现：按关键词前缀匹配（测试用）
	var results []Fact
	for _, fact := range m.longTerm {
		if limit > 0 && len(results) >= limit {
			break
		}
		if contains(fact.Value, query) || contains(fact.Key, query) {
			results = append(results, fact)
		}
	}
	return results, nil
}

// NewMockManager 返回一个空的 InMemoryManager，供其他模块测试使用。
func NewMockManager() *InMemoryManager {
	return NewInMemoryManager()
}

func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}
