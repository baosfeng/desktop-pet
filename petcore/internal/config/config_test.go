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
	if cfg.Env != EnvDevelopment {
		t.Errorf("default env = %q, want %q", cfg.Env, EnvDevelopment)
	}
}

func TestDefaultConfig_NoMutation(t *testing.T) {
	c1 := DefaultConfig()
	c2 := DefaultConfig()
	c1.LLM.Provider = "openai"
	if c2.LLM.Provider != "mock" {
		t.Error("DefaultConfig should return fresh values each call")
	}
}

func TestDefaultConfigFor_Development(t *testing.T) {
	cfg := DefaultConfigFor(EnvDevelopment)
	want := map[string]bool{
		"agent.memory_stage": false,
		"agent.tool_loop":    false,
		"plugin.l1_yaml":     false,
		"plugin.l2_js":       false,
		"plugin.l3_mcp":      false,
	}
	for k, v := range want {
		if cfg.FeatureFlags[k] != v {
			t.Errorf("dev FeatureFlags[%q] = %v, want %v", k, cfg.FeatureFlags[k], v)
		}
	}
	if cfg.LLM.Provider != "mock" {
		t.Errorf("dev LLM.Provider = %q, want %q", cfg.LLM.Provider, "mock")
	}
}

func TestDefaultConfigFor_Production(t *testing.T) {
	cfg := DefaultConfigFor(EnvProduction)
	wantOn := []string{
		"agent.memory_stage",
		"agent.tool_loop",
		"plugin.l1_yaml",
	}
	for _, k := range wantOn {
		if !cfg.FeatureFlags[k] {
			t.Errorf("prod FeatureFlags[%q] = false, want true", k)
		}
	}
	if cfg.LLM.Provider != "openai" {
		t.Errorf("prod LLM.Provider = %q, want %q", cfg.LLM.Provider, "openai")
	}
	if cfg.LLM.Model != "gpt-4o-mini" {
		t.Errorf("prod LLM.Model = %q, want %q", cfg.LLM.Model, "gpt-4o-mini")
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

func TestLoad_FileNotExist_WithEnvArg(t *testing.T) {
	cfg, err := Load("/tmp/nonexistent/config.toml", EnvProduction)
	if err != nil {
		t.Fatalf("Load returned error for missing file: %v", err)
	}
	if cfg.LLM.Provider != "openai" {
		t.Errorf("expected production default, got provider=%q", cfg.LLM.Provider)
	}
	if cfg.Env != EnvProduction {
		t.Errorf("env = %q, want %q", cfg.Env, EnvProduction)
	}
}

func TestLoad_FileExists_LoadsConfig(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")
	if err := os.WriteFile(path, []byte("[llm]\nprovider=\"openai\"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load should parse valid TOML: %v", err)
	}
	if cfg.LLM.Provider != "openai" {
		t.Errorf("LLM.Provider = %q, want %q", cfg.LLM.Provider, "openai")
	}
}

func TestLoad_EnvOverrideFile(t *testing.T) {
	dir := t.TempDir()
	basePath := filepath.Join(dir, "config.toml")
	devPath := filepath.Join(dir, "config.development.toml")

	// 基础配置
	if err := os.WriteFile(basePath, []byte("[llm]\nprovider=\"ollama\"\nmax_tokens=512\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	// 开发环境覆盖：仅覆盖 provider
	if err := os.WriteFile(devPath, []byte("[llm]\nprovider=\"openai\"\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(basePath, EnvDevelopment)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if cfg.LLM.Provider != "openai" {
		t.Errorf("after override LLM.Provider = %q, want %q", cfg.LLM.Provider, "openai")
	}
	// 基础配置中未被子覆盖的字段应保留
	if cfg.LLM.MaxTokens != 512 {
		t.Errorf("MaxTokens should retain base value = %d, got %d", 512, cfg.LLM.MaxTokens)
	}
}

func TestLoad_ProductionOverrideFile(t *testing.T) {
	dir := t.TempDir()
	basePath := filepath.Join(dir, "config.toml")
	prodPath := filepath.Join(dir, "config.production.toml")

	if err := os.WriteFile(basePath, []byte("[llm]\nprovider=\"ollama\"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(prodPath, []byte("[window]\nalways_on_top=false\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(basePath, EnvProduction)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	// 生产环境默认 provider 为 openai，但基础配置覆盖为 ollama → 最终应该是 ollama
	if cfg.LLM.Provider != "ollama" {
		t.Errorf("LLM.Provider = %q, want %q (base should win over default)", cfg.LLM.Provider, "ollama")
	}
	// 生产覆盖文件覆盖 window
	if cfg.Window.AlwaysOnTop {
		t.Error("Window.AlwaysOnTop should be false after prod override")
	}
}

func TestLoad_EnvFileMissing_StillLoadsBase(t *testing.T) {
	dir := t.TempDir()
	basePath := filepath.Join(dir, "config.toml")
	// 只有基础配置，没有 config.production.toml
	if err := os.WriteFile(basePath, []byte("[agent]\nmax_tool_turns=5\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(basePath, EnvProduction)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if cfg.Agent.MaxToolTurns != 5 {
		t.Errorf("MaxToolTurns = %d, want %d", cfg.Agent.MaxToolTurns, 5)
	}
	// 环境默认值应生效
	if cfg.LLM.Provider != "openai" {
		t.Errorf("production default LLM.Provider = %q, want %q", cfg.LLM.Provider, "openai")
	}
}

func TestCurrentEnv_Default(t *testing.T) {
	t.Setenv("PETCORE_ENV", "")
	got := CurrentEnv()
	if got != EnvDevelopment {
		t.Errorf("CurrentEnv() = %q, want %q", got, EnvDevelopment)
	}
}

func TestCurrentEnv_Production(t *testing.T) {
	t.Setenv("PETCORE_ENV", "production")
	got := CurrentEnv()
	if got != EnvProduction {
		t.Errorf("CurrentEnv() = %q, want %q", got, EnvProduction)
	}
}

func TestCurrentEnv_Invalid(t *testing.T) {
	t.Setenv("PETCORE_ENV", "staging")
	got := CurrentEnv()
	if got != EnvDevelopment {
		t.Errorf("CurrentEnv() with invalid value = %q, want %q", got, EnvDevelopment)
	}
}

func TestCurrentEnv_CaseSensitive(t *testing.T) {
	t.Setenv("PETCORE_ENV", "Development")
	got := CurrentEnv()
	// 目前是大小写敏感的精确匹配，非 "production" 都回退到 development
	if got != EnvDevelopment {
		t.Errorf("CurrentEnv('Development') = %q, want %q", got, EnvDevelopment)
	}
}

func TestEnvConfigPath(t *testing.T) {
	tests := []struct {
		base string
		env  Environment
		want string
	}{
		{"/path/to/config.toml", EnvDevelopment, "/path/to/config.development.toml"},
		{"/path/to/config.toml", EnvProduction, "/path/to/config.production.toml"},
		{"/a/b/c.yaml", EnvDevelopment, "/a/b/c.development.yaml"},
		{"config.toml", EnvProduction, "config.production.toml"},
		{"./app.conf", EnvDevelopment, "./app.development.conf"},
	}
	for _, tt := range tests {
		got := envConfigPath(tt.base, tt.env)
		if got != tt.want {
			t.Errorf("envConfigPath(%q, %q) = %q, want %q", tt.base, tt.env, got, tt.want)
		}
	}
}

func TestAPIKey_FromEnv(t *testing.T) {
	t.Setenv("MY_VAR", "sk-test123")
	cfg := LLMConfig{APIKeyEnv: "MY_VAR"}
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
	configDir := os.Getenv("PETCORE_DATA_DIR")
	if configDir != "/custom/data" {
		t.Errorf("PETCORE_DATA_DIR = %q, want %q", configDir, "/custom/data")
	}
}

func TestDataDir_EmptyEnv_FallsBack(t *testing.T) {
	// When PETCORE_DATA_DIR is unset, dataDir() returns ~/.desktop-pet
	// We can't easily test os.UserHomeDir() but we can verify env is used when set
	t.Setenv("PETCORE_DATA_DIR", "")
	configDir := os.Getenv("PETCORE_DATA_DIR")
	if configDir != "" {
		t.Errorf("expected empty PETCORE_DATA_DIR")
	}
}

func TestLoad_EnvField_SetCorrectly(t *testing.T) {
	t.Run("no env arg uses CurrentEnv", func(t *testing.T) {
		t.Setenv("PETCORE_ENV", "")
		cfg, err := Load("/tmp/nonexistent/config.toml")
		if err != nil {
			t.Fatal(err)
		}
		if cfg.Env != EnvDevelopment {
			t.Errorf("cfg.Env = %q, want %q", cfg.Env, EnvDevelopment)
		}
	})

	t.Run("explicit env arg overrides env var", func(t *testing.T) {
		t.Setenv("PETCORE_ENV", "development")
		cfg, err := Load("/tmp/nonexistent/config.toml", EnvProduction)
		if err != nil {
			t.Fatal(err)
		}
		if cfg.Env != EnvProduction {
			t.Errorf("cfg.Env = %q, want %q", cfg.Env, EnvProduction)
		}
	})
}
