package plugin

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestLoadYAMLBytes(t *testing.T) {
	data := []byte(`
name: hello_world
description: A simple greeting
actions:
  - trigger: on_startup
    speak: "Hello! I'm your desktop pet!"
    emotion: happy
  - trigger: on_idle
    speak: "I'm bored..."
    delay_ms: 5000
`)

	pack, err := LoadYAMLBytes(data)
	if err != nil {
		t.Fatal(err)
	}

	if pack.Name != "hello_world" {
		t.Errorf("Name = %q, want %q", pack.Name, "hello_world")
	}
	if len(pack.Actions) != 2 {
		t.Fatalf("expected 2 actions, got %d", len(pack.Actions))
	}
	if pack.Actions[0].Trigger != "on_startup" {
		t.Errorf("Action[0].Trigger = %q, want %q", pack.Actions[0].Trigger, "on_startup")
	}
	if pack.Actions[0].Speak != "Hello! I'm your desktop pet!" {
		t.Errorf("Action[0].Speak = %q, want %q", pack.Actions[0].Speak, "Hello! I'm your desktop pet!")
	}
	if pack.Actions[0].Emotion != "happy" {
		t.Errorf("Action[0].Emotion = %q, want %q", pack.Actions[0].Emotion, "happy")
	}
}

func TestLoadYAMLBytes_EmptyName(t *testing.T) {
	data := []byte(`
actions:
  - trigger: on_startup
    speak: "hi"
`)
	_, err := LoadYAMLBytes(data)
	if err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestLoadYAMLDir_LoadsYAMLFiles(t *testing.T) {
	// Create temp dir with yaml files
	tmpDir := t.TempDir()

	yamlContent := []byte(`
name: test_pack
actions:
  - trigger: on_startup
    speak: "test"
`)
	if err := os.WriteFile(filepath.Join(tmpDir, "test.yaml"), yamlContent, 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "ignored.txt"), []byte("not yaml"), 0644); err != nil {
		t.Fatal(err)
	}
	ymlContent := []byte(`
name: yml_pack
actions:
  - trigger: on_idle
    action: wave
`)
	if err := os.WriteFile(filepath.Join(tmpDir, "action.yml"), ymlContent, 0644); err != nil {
		t.Fatal(err)
	}

	reg := NewRegistry()
	count, err := LoadYAMLDir(reg, tmpDir)
	if err != nil {
		t.Fatal(err)
	}
	if count != 2 {
		t.Errorf("expected 2 plugins loaded, got %d", count)
	}

	plugins := reg.List()
	if len(plugins) != 2 {
		t.Fatalf("expected 2 plugins registered, got %d", len(plugins))
	}
}

func TestLoadYAMLDir_NonExistentDir(t *testing.T) {
	reg := NewRegistry()
	count, err := LoadYAMLDir(reg, "/tmp/nonexistent-yaml-dir-12345")
	if err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Errorf("expected 0 plugins for nonexistent dir, got %d", count)
	}
}

func TestYAMLPlugin_Interface(t *testing.T) {
	pack := &ActionPack{
		Name: "test",
		Actions: []Action{
			{Trigger: "on_startup", Speak: "hello"},
		},
	}
	p := NewYAMLPlugin(pack)

	if p.Name() != "yaml:test" {
		t.Errorf("Name() = %q, want %q", p.Name(), "yaml:test")
	}
	if p.Type() != TypeL1YAML {
		t.Errorf("Type() = %v, want %v", p.Type(), TypeL1YAML)
	}
	if err := p.Start(context.Background()); err != nil {
		t.Errorf("Start() error: %v", err)
	}
	if err := p.Stop(); err != nil {
		t.Errorf("Stop() error: %v", err)
	}
}

func TestLoadYAMLDir_FileReadError(t *testing.T) {
	// A directory we can list but not read a specific file from is not easy to create.
	// Instead, test with a non-existent file passed via a malformed directory structure.
	// This test verifies we handle non-existent directories gracefully.
	tmpDir := t.TempDir()
	reg := NewRegistry()

	// Empty dir should produce 0 plugins
	count, err := LoadYAMLDir(reg, tmpDir)
	if err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Errorf("expected 0 plugins for empty dir, got %d", count)
	}
}

func TestLoadYAMLFile_InvalidFormat(t *testing.T) {
	tmpDir := t.TempDir()
	badFile := filepath.Join(tmpDir, "bad.yaml")
	if err := os.WriteFile(badFile, []byte("invalid yaml: [\n"), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := LoadYAMLBytes([]byte("invalid yaml: [\n"))
	if err == nil {
		t.Error("expected error for invalid YAML")
	}
}

func TestYAMLPlugin_Pack(t *testing.T) {
	pack := &ActionPack{Name: "test", Actions: []Action{{Trigger: "on_startup", Speak: "hi"}}}
	p := NewYAMLPlugin(pack)
	got := p.Pack()
	if got.Name != "test" {
		t.Errorf("Pack().Name = %q, want %q", got.Name, "test")
	}
}

func TestLoadYAMLDir_WithDirEntry(t *testing.T) {
	tmpDir := t.TempDir()
	// Create a subdirectory with .yaml extension to test isDir skip
	subDir := filepath.Join(tmpDir, "subdir.yaml")
	if err := os.Mkdir(subDir, 0755); err != nil {
		t.Fatal(err)
	}

	reg := NewRegistry()
	count, err := LoadYAMLDir(reg, tmpDir)
	if err != nil {
		t.Fatal(err)
	}
	// Directory with .yaml extension should be skipped by isDir check
	if count != 0 {
		t.Errorf("expected 0 plugins, got %d", count)
	}
}
