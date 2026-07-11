// Package feature 提供功能开关 (Feature Flags) 系统。
//
// 用于控制 Phase 2+ 的渐进式功能开放。
// 所有 flag 通过 TOML 配置驱动，无需修改代码即可开关功能。
package feature

import "fmt"

// Flags 是功能开关集合。
type Flags struct {
	flags map[string]bool
}

// New 根据配置创建一个 Flags 实例。
// cfg 是 feature_flags 段的配置映射。
func New(cfg map[string]bool) *Flags {
	f := &Flags{flags: make(map[string]bool, len(cfg))}
	for k, v := range cfg {
		f.flags[k] = v
	}
	return f
}

// IsEnabled 查询指定功能是否开启。
// 未注册的 flag 返回 false。
func (f *Flags) IsEnabled(name string) bool {
	v, ok := f.flags[name]
	if !ok {
		return false
	}
	return v
}

// AllEnabled 返回所有已开启的 flag 列表。
func (f *Flags) AllEnabled() map[string]bool {
	out := make(map[string]bool, len(f.flags))
	for k, v := range f.flags {
		if v {
			out[k] = true
		}
	}
	return out
}

// Register 注册一个 flag 及其默认值。
// 如果该 flag 尚未在配置中定义，则使用默认值。
func (f *Flags) Register(name string, defaultValue bool) {
	if _, ok := f.flags[name]; !ok {
		f.flags[name] = defaultValue
	}
}

// ─── 预定义 Flag 名称 ────────────────────────

// 预定义的 Flag 名称常量。
const (
	FlagMemoryStage = "agent.memory_stage" // Phase 2: 注入记忆到上下文
	FlagToolLoop    = "agent.tool_loop"    // Phase 4: 工具调用循环
	FlagL1YAML      = "plugin.l1_yaml"     // Phase 4: L1 YAML 动作包
	FlagL2JS        = "plugin.l2_js"       // Phase 4: L2 JS 脚本引擎
	FlagL3MCP       = "plugin.l3_mcp"      // Phase 4: L3 MCP 子进程
)

// RegisteredFlags 返回所有预定义 flag 及其默认值。
func RegisteredFlags() map[string]bool {
	return map[string]bool{
		FlagMemoryStage: false,
		FlagToolLoop:    false,
		FlagL1YAML:      false,
		FlagL2JS:        false,
		FlagL3MCP:       false,
	}
}

// RegisterDefaults 将所有预定义 flag 注册到实例中（如配置未定义则使用默认值）。
func (f *Flags) RegisterDefaults() {
	for name, defaultValue := range RegisteredFlags() {
		f.Register(name, defaultValue)
	}
}

// String 返回所有 flag 的状态摘要。
func (f *Flags) String() string {
	out := "Feature Flags:\n"
	for k, v := range f.flags {
		status := "off"
		if v {
			status = "on"
		}
		out += fmt.Sprintf("  %s = %s\n", k, status)
	}
	return out
}
