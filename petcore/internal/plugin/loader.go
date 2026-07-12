// Package plugin 提供三层插件系统。
//
// L1 YAML 动作包加载器：扫描指定目录中的 *.yaml 文件，
// 解析为声明式动作包并注册为插件。
package plugin

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"gopkg.in/yaml.v3"
)

// ─── YAML 动作包格式 ─────────────────────────

// ActionPack 是 L1 YAML 动作包的文件格式。
type ActionPack struct {
	Name        string   `yaml:"name"`
	Description string   `yaml:"description,omitempty"`
	Actions     []Action `yaml:"actions"`
}

// Action 定义单个动作及其触发条件。
type Action struct {
	Trigger string   `yaml:"trigger"`            // on_startup / on_user_message / on_idle
	Speak   string   `yaml:"speak,omitempty"`    // 宠物说的话
	Emotion string   `yaml:"emotion,omitempty"`  // 情绪变化
	Action  string   `yaml:"action,omitempty"`   // 动作名称（wave / jump / blink）
	DelayMS int      `yaml:"delay_ms,omitempty"` // 执行延迟
	Tags    []string `yaml:"tags,omitempty"`     // 匹配标签
}

// ─── YAML 动作包插件 ─────────────────────────

// YAMLPlugin 是一个 L1 YAML 动作包插件。
type YAMLPlugin struct {
	pack *ActionPack
	mu   sync.RWMutex
}

// NewYAMLPlugin 创建一个 YAML 动作包插件。
func NewYAMLPlugin(pack *ActionPack) *YAMLPlugin {
	return &YAMLPlugin{pack: pack}
}

// Name 返回插件名称。
func (p *YAMLPlugin) Name() string { return "yaml:" + p.pack.Name }

// Type 返回插件类型（L1 YAML）。
func (p *YAMLPlugin) Type() Type { return TypeL1YAML }

// Start 启动插件（YAML 插件无需启动操作）。
func (p *YAMLPlugin) Start(_ context.Context) error { return nil }

// Stop 停止插件。
func (p *YAMLPlugin) Stop() error { return nil }

// Pack 返回插件的动作包。
func (p *YAMLPlugin) Pack() *ActionPack {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.pack
}

// ─── 加载器 ──────────────────────────────────

// LoadYAMLDir 扫描指定目录中的所有 *.yaml 文件，
// 解析为 YAMLPlugin 并注册到 Registry。
func LoadYAMLDir(reg Registry, dir string) (int, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil // 目录不存在不是错误
		}
		return 0, fmt.Errorf("plugin: read dir %s: %w", dir, err)
	}

	count := 0
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		ext := filepath.Ext(entry.Name())
		if ext != ".yaml" && ext != ".yml" {
			continue
		}

		fullPath := filepath.Join(dir, entry.Name())
		pack, err := loadYAMLFile(fullPath)
		if err != nil {
			return count, fmt.Errorf("plugin: load %s: %w", fullPath, err)
		}

		plugin := NewYAMLPlugin(pack)
		if err := reg.Register(plugin); err != nil {
			return count, fmt.Errorf("plugin: register %s: %w", entry.Name(), err)
		}
		count++
	}

	return count, nil
}

func loadYAMLFile(path string) (*ActionPack, error) {
	// G304: path comes from filepath.Join(dir, entry.Name()) which is controlled input
	data, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}

	var pack ActionPack
	if err := yaml.Unmarshal(data, &pack); err != nil {
		return nil, fmt.Errorf("parse yaml: %w", err)
	}

	if pack.Name == "" {
		return nil, fmt.Errorf("action pack name is required")
	}

	return &pack, nil
}

// LoadYAMLBytes 从字节数据加载 YAML 动作包（用于测试）。
func LoadYAMLBytes(data []byte) (*ActionPack, error) {
	var pack ActionPack
	if err := yaml.Unmarshal(data, &pack); err != nil {
		return nil, fmt.Errorf("parse yaml: %w", err)
	}
	if pack.Name == "" {
		return nil, fmt.Errorf("action pack name is required")
	}
	return &pack, nil
}
