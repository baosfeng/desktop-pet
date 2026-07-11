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

// Environment 表示运行环境。
type Environment string

const (
	// EnvDevelopment 是开发环境（默认），使用 mock LLM、关闭不稳定特性。
	EnvDevelopment Environment = "development"
	// EnvProduction 是生产环境，使用真实 LLM、开启稳定特性。
	EnvProduction Environment = "production"
)

// Config 是 PetCore 的完整配置结构。
type Config struct {
	Env          Environment     `toml:"env"`
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

// DefaultConfig 返回开发环境的默认配置。
func DefaultConfig() *Config {
	return DefaultConfigFor(EnvDevelopment)
}

// DefaultConfigFor 返回指定环境的默认配置。
// 生产环境默认使用真实的 LLM Provider 并开启稳定特性；
// 开发环境使用 mock LLM 并关闭所有非必要特性。
func DefaultConfigFor(env Environment) *Config {
	cfg := &Config{
		Env: env,
		LLM: LLMConfig{
			Provider:    "mock",
			Model:       "mock-v1",
			MaxTokens:   1024,
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

	if env == EnvProduction {
		cfg.LLM.Provider = "openai"
		cfg.LLM.Model = "gpt-4o-mini"
		cfg.FeatureFlags["agent.memory_stage"] = true
		cfg.FeatureFlags["agent.tool_loop"] = true
		cfg.FeatureFlags["plugin.l1_yaml"] = true
	}

	return cfg
}

// Load 从指定路径加载 TOML 配置文件，同时自动加载环境特定覆盖文件。
// 环境由 env 参数指定；若未提供，使用 CurrentEnv() 自动检测。
// 加载顺序：默认值 → config.toml（基础配置） → config.{env}.toml（环境覆盖）
func Load(path string, env ...Environment) (*Config, error) {
	e := CurrentEnv()
	if len(env) > 0 {
		e = env[0]
	}

	cfg := DefaultConfigFor(e)

	// 1. 加载基础配置
	if _, err := os.Stat(path); err == nil {
		if _, err := toml.DecodeFile(path, cfg); err != nil {
			return nil, fmt.Errorf("config: failed to parse %s: %w", path, err)
		}
	}

	// 2. 加载环境特定覆盖文件（config.development.toml / config.production.toml）
	envPath := envConfigPath(path, e)
	if envPath != "" {
		if _, err := os.Stat(envPath); err == nil {
			if _, err := toml.DecodeFile(envPath, cfg); err != nil {
				return nil, fmt.Errorf("config: failed to parse env config %s: %w", envPath, err)
			}
		}
	}

	cfg.Env = e
	return cfg, nil
}

// CurrentEnv 返回当前运行环境。
// 从 PETCORE_ENV 环境变量读取，不区分大小写。
// 若未设置或值无效，默认返回 EnvDevelopment。
func CurrentEnv() Environment {
	env := os.Getenv("PETCORE_ENV")
	switch Environment(env) {
	case EnvProduction:
		return EnvProduction
	default:
		return EnvDevelopment
	}
}

// envConfigPath 根据基础配置文件路径和环境，返回环境特定配置文件的路径。
// 例如 /path/to/config.toml + development → /path/to/config.development.toml
func envConfigPath(basePath string, env Environment) string {
	ext := filepath.Ext(basePath)
	base := basePath[:len(basePath)-len(ext)]
	return base + "." + string(env) + ext
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
