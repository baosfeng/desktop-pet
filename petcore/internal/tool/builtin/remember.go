// Package builtin 提供 PetCore 内置工具。
//
// 这些工具直接由 Agent 注册，不需要通过 MCP 等外部服务。
package builtin

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/desktop-pet/petcore/internal/memory"
)

// RememberTool 允许 LLM 保存用户事实到记忆系统。
type RememberTool struct {
	mem memory.Manager
}

// NewRemember 创建一个新的 RememberTool 实例。
// mem 是记忆系统管理器，用于持久化用户事实。
func NewRemember(mem memory.Manager) *RememberTool {
	return &RememberTool{mem: mem}
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

// Execute 执行 remember 工具调用，将事实保存到记忆系统。
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

	// 持久化到记忆系统
	if t.mem != nil {
		if err := t.mem.Remember(params.Key, params.Value, params.Importance); err != nil {
			return "", fmt.Errorf("remember: save failed: %w", err)
		}

		// 同时保存为长期记忆（通过事实抽取机制）
		_ = t.mem.Store(context.TODO(), memory.Fact{
			Key:        params.Key,
			Value:      params.Value,
			Category:   "preference",
			Importance: params.Importance,
			Timestamp:  timestamp(),
		})
	}

	result := fmt.Sprintf("Saved fact: %s = %s (importance: %d)", params.Key, params.Value, params.Importance)
	return result, nil
}

func timestamp() int64 {
	ts := os.Getenv("MOCK_NOW_MS")
	if ts == "" {
		return 0 // real timestamp handled by caller
	}
	v, _ := strconv.ParseInt(ts, 10, 64)
	return v
}
