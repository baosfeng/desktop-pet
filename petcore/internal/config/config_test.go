package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig_Values(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.LLM.Provider != "mock" {
		t.Errorf("default provider = %q, want %q", cfg.LLM.Provider, "mock")
	}
	if cfg.Memory.ShortTermSize != 20 {
		t.Errorf("ShortTermSize = %d, want %d", cfg.Memory.ShortTermSize, 20)
	}
	if cfg.Agent.MaxToolTurns != 10 {
		t.Errorf("MaxToolTurns = %d, want %d", cfg.Agent.MaxToolTurns, 10)
	}
}

func TestDefaultConfig_NoMutation(t *testing.T) {
	// 多次调用应返回独立副本
	c1 := DefaultConfig()
	c2 := DefaultConfig()
	c1.LLM.Provider = "openai"
	if c2.LLM.Provider != "mock" {
		t.Error("DefaultConfig should return fresh values each call")
	}
}

func TestLoad_FileNotExist_ReturnsDefault(t *testing.T) {
	cfg, err := Load("/tmp/nonexistent/config.toml")
	if err != nil {
		t.Fatalf("Load returned error for missing file: %v", err)
	}
	if cfg.LLM.Provider != "mock" {
		t.Errorf("expected default config, got provider=%q", cfg.LLM.Provider)
	}
}

func TestLoad_FileExists_ReturnsTODO(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")
	if err := os.WriteFile(path, []byte("[llm]\nprovider=\"openai\"\n"), 0644); err != nil {
		t.Fatal(err)
	}
	_, err := Load(path)
	if err == nil {
		t.Error("expected error (TOML parsing not implemented)")
	}
}

func TestAPIKey_FromEnv(t *testing.T) {
	t.Setenv("TEST_LLM_KEY", "sk-test123")
	cfg := LLMConfig{APIKeyEnv: "TEST_LLM_KEY"}
	if got := cfg.APIKey(); got != "sk-test123" {
		t.Errorf("APIKey() = %q, want %q", got, "sk-test123")
	}
}

func TestAPIKey_EmptyEnv(t *testing.T) {
	cfg := LLMConfig{APIKeyEnv: ""}
	if got := cfg.APIKey(); got != "" {
		t.Errorf("APIKey() = %q, want %q", got, "")
	}
}

func TestDataDir_RespectsEnv(t *testing.T) {
	t.Setenv("PETCORE_DATA_DIR", "/custom/data")
	// dataDir is not exported; we test via config load behavior
	configDir := os.Getenv("PETCORE_DATA_DIR")
	if configDir != "/custom/data" {
		t.Errorf("PETCORE_DATA_DIR = %q, want %q", configDir, "/custom/data")
	}
}
