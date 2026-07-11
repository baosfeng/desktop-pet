// Package builtin 提供 PetCore 内置工具。
//
// 这些工具直接由 Agent 注册，不需要通过 MCP 等外部服务。
package builtin

import (
	"context"
	"encoding/json"
	"fmt"
)

// RememberTool 允许 LLM 保存用户事实到记忆系统。
//
// Phase 1：仅提供工具定义（Name / Description / JSON Schema）。
// Phase 2：在 Execute 中通过 ToolRegistry 或 Agent 回调保存到 memory.Manager。
type RememberTool struct{}

// NewRemember 创建一个新的 RememberTool 实例。
func NewRemember() *RememberTool {
	return &RememberTool{}
}

// Name 返回工具名称。
func (t *RememberTool) Name() string {
	return "remember"
}

// Description 返回工具描述，LLM 通过此描述了解工具用途。
func (t *RememberTool) Description() string {
	return `Save a user fact to long-term memory. The LLM should call this when the user mentions something worth remembering — a preference, a personal detail, a goal, or a recurring need.

Arguments (JSON):
  - key: string (required) — a concise label for the fact, e.g. "favorite_food"
  - value: string (required) — the fact content, e.g. "user loves pizza"
  - importance: int (optional, default 5) — importance level 1-10

Example:
  {"key": "coffee_preference", "value": "user drinks black coffee every morning", "importance": 6}`
}

// RememberArgs 是 RememberTool.Execute 的参数结构。
type RememberArgs struct {
	Key        string `json:"key"`
	Value      string `json:"value"`
	Importance int    `json:"importance"`
}

// Execute 执行 remember 工具调用。
// Phase 1：仅解析参数并返回确认信息——实际持久化在 Phase 4 实现。
func (t *RememberTool) Execute(_ context.Context, args string) (string, error) {
	var params RememberArgs
	if err := json.Unmarshal([]byte(args), &params); err != nil {
		return "", fmt.Errorf("remember: invalid args: %w", err)
	}

	if params.Key == "" {
		return "", fmt.Errorf("remember: key is required")
	}
	if params.Value == "" {
		return "", fmt.Errorf("remember: value is required")
	}
	if params.Importance <= 0 {
		params.Importance = 5
	}
	if params.Importance > 10 {
		params.Importance = 10
	}

	result := fmt.Sprintf("Saved fact: %s = %s (importance: %d)", params.Key, params.Value, params.Importance)
	return result, nil
}
