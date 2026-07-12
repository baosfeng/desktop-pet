// Package openai 提供 DeepSeek 对话模型的便捷注册。
//
// DeepSeek 使用完全兼容 OpenAI 的 API 格式，只需更换 base_url 和模型名。
// 此处注册为 "deepseek" 提供默认配置，实际复用 openai.Provider 的实现。
package openai

import (
	"fmt"
	"strings"

	"github.com/desktop-pet/petcore/internal/llm"
)

func init() {
	llm.Register("deepseek", func(cfg map[string]any) (llm.Provider, error) {
		p := &Provider{
			model:   "deepseek-chat",
			baseURL: "https://api.deepseek.com/v1",
		}
		if cfg != nil {
			if v, ok := cfg["model"].(string); ok && v != "" {
				p.model = v
			}
			if v, ok := cfg["base_url"].(string); ok && v != "" {
				p.baseURL = strings.TrimRight(v, "/")
			}
			if v, ok := cfg["api_key"].(string); ok && v != "" {
				p.apiKey = v
			}
		}
		if p.apiKey == "" {
			return nil, fmt.Errorf("deepseek: API key is required (set PETCORE_API_KEY env or configure api_key_env)")
		}

		return p, nil
	})
}
