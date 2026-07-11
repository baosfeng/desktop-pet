// Package ollama 提供 Ollama 本地 LLM 的 Provider 实现。
//
// 在 init() 中自动注册为 "ollama" Provider，通过 blank import 启用：
//
//	import _ "github.com/desktop-pet/petcore/internal/llm/ollama"
package ollama

import (
	"fmt"

	"github.com/desktop-pet/petcore/internal/llm"
)

func init() {
	llm.Register("ollama", func(_ map[string]any) (llm.Provider, error) {
		return nil, fmt.Errorf("ollama: not yet implemented")
	})
}
