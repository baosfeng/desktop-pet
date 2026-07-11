// Package openai 提供 OpenAI 兼容 API 的 LLM Provider。
//
// 在 init() 中自动注册为 "openai" Provider，通过 blank import 启用：
//
//	import _ "github.com/desktop-pet/petcore/internal/llm/openai"
package openai

import (
	"fmt"

	"github.com/desktop-pet/petcore/internal/llm"
)

func init() {
	llm.Register("openai", func(_ map[string]any) (llm.Provider, error) {
		return nil, fmt.Errorf("openai: not yet implemented")
	})
}
