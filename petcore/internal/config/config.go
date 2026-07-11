// Package config 提供 PetCore 配置加载与管理。
//
// 仅依赖 TOML 解析器和环境变量，不依赖其他 petcore 内部模块。
// 可在单元测试中用临时文件或内存数据测试。
package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// Config 是 PetCore 的完整配置结构。
type Config struct {
	LLM          LLMConfig       `toml:"llm"`
	Agent        AgentConfig     `toml:"agent"`
	Window       WindowConfig    `toml:"window"`
	Update       UpdateConfig    `toml:"update"`
	Memory       MemoryConfig    `toml:"memory"`
	Plugin       PluginConfig    `toml:"plugin"`
	FeatureFlags map[string]bool `toml:"feature_flags"`
}

// LLMConfig 是 LLM Provider 的配置。
type LLMConfig struct {
	Provider    string  `toml:"provider"` // openai / ollama / mock
	Model       string  `toml:"model"`
	BaseURL     string  `toml:"base_url"`
	APIKeyEnv   string  `toml:"api_key_env"` // 从环境变量读取 API Key
	MaxTokens   int     `toml:"max_tokens"`
	Temperature float64 `toml:"temperature"`
}

// AgentConfig 是 AI Agent 的配置。
type AgentConfig struct {
	SystemPrompt string  `toml:"system_prompt"`
	Temperature  float64 `toml:"temperature"`
	MaxToolTurns int     `toml:"max_tool_turns"`
}

// WindowConfig 是窗口相关的配置（桌面壳使用）。
type WindowConfig struct {
	AlwaysOnTop bool    `toml:"always_on_top"`
	Transparent bool    `toml:"transparent"`
	Opacity     float64 `toml:"opacity"`
}

// UpdateConfig 是自动更新相关的配置。
type UpdateConfig struct {
	AutoCheck bool   `toml:"auto_check"`
	Channel   string `toml:"channel"` // stable / beta
}

// MemoryConfig 是记忆系统配置。
type MemoryConfig struct {
	ShortTermSize int `toml:"short_term_size"` // L1 短期记忆轮数
}

// PluginConfig 是插件系统配置。
type PluginConfig struct {
	ActionsDir string   `toml:"actions_dir"` // L1 YAML 目录
	ScriptsDir string   `toml:"scripts_dir"` // L2 JS 目录
	MCPConfigs []MCPDef `toml:"mcps"`        // L3 MCP 定义
}

// MCPDef 定义了一个 MCP 子进程插件。
type MCPDef struct {
	Name    string   `toml:"name"`
	Command string   `toml:"command"`
	Args    []string `toml:"args"`
	Enabled bool     `toml:"enabled"`
}

// DefaultConfig 返回带默认值的配置。
func DefaultConfig() *Config {
	return &Config{
		LLM: LLMConfig{
			Provider:    "mock",
			Model:       "mock-v1",
			MaxTokens:   4096,
			Temperature: 0.7,
		},
		Agent: AgentConfig{
			SystemPrompt: "You are a cute desktop pet.",
			Temperature:  0.7,
			MaxToolTurns: 10,
		},
		Window: WindowConfig{
			AlwaysOnTop: true,
			Transparent: true,
			Opacity:     1.0,
		},
		Update: UpdateConfig{
			AutoCheck: true,
			Channel:   "stable",
		},
		Memory: MemoryConfig{
			ShortTermSize: 20,
		},
		Plugin: PluginConfig{
			ActionsDir: filepath.Join(dataDir(), "actions"),
			ScriptsDir: filepath.Join(dataDir(), "scripts"),
		},
		FeatureFlags: map[string]bool{
			"agent.memory_stage": false,
			"agent.tool_loop":    false,
			"plugin.l1_yaml":     false,
			"plugin.l2_js":       false,
			"plugin.l3_mcp":      false,
		},
	}
}

// Load 从指定路径加载 TOML 配置文件。
// 如果文件不存在，返回默认配置。
func Load(path string) (*Config, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return DefaultConfig(), nil
	}

	cfg := DefaultConfig()
	if _, err := toml.DecodeFile(path, cfg); err != nil {
		return nil, fmt.Errorf("config: failed to parse %s: %w", path, err)
	}
	return cfg, nil
}

// APIKey 从环境变量读取 LLM API Key。
func (c *LLMConfig) APIKey() string {
	if c.APIKeyEnv == "" {
		return ""
	}
	return os.Getenv(c.APIKeyEnv)
}

func dataDir() string {
	if d := os.Getenv("PETCORE_DATA_DIR"); d != "" {
		return d
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "~/.desktop-pet"
	}
	return filepath.Join(home, ".desktop-pet")
}
