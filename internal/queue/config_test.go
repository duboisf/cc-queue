package queue

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConfigDir_XDGConfigHome(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "/tmp/test-xdg-config")
	got := ConfigDir()
	want := "/tmp/test-xdg-config/cc-queue"
	if got != want {
		t.Errorf("ConfigDir() = %q, want %q", got, want)
	}
}

func TestConfigDir_Default(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "")
	home, _ := os.UserHomeDir()
	got := ConfigDir()
	want := filepath.Join(home, ".config", "cc-queue")
	if got != want {
		t.Errorf("ConfigDir() = %q, want %q", got, want)
	}
}

func TestReadConfig_NoFile(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmp)

	cfg := ReadConfig()
	if cfg.Debug {
		t.Error("expected Debug=false when no config file exists")
	}
}

func TestReadConfig_InvalidJSON(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmp)

	dir := filepath.Join(tmp, "cc-queue")
	os.MkdirAll(dir, 0755)
	os.WriteFile(filepath.Join(dir, "config.json"), []byte("{invalid"), 0644)

	cfg := ReadConfig()
	if cfg.Debug {
		t.Error("expected Debug=false for invalid JSON")
	}
}

func TestWriteAndReadConfig(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmp)

	cfg := Config{Debug: true}
	if err := WriteConfig(cfg); err != nil {
		t.Fatalf("WriteConfig: %v", err)
	}

	got := ReadConfig()
	if !got.Debug {
		t.Error("expected Debug=true after writing")
	}
}

func TestWriteConfig_CreatesDirectory(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmp)

	if err := WriteConfig(Config{Debug: true}); err != nil {
		t.Fatalf("WriteConfig: %v", err)
	}

	dir := filepath.Join(tmp, "cc-queue")
	if _, err := os.Stat(dir); err != nil {
		t.Errorf("config directory not created: %v", err)
	}
}

func TestDefaultConfigJSON(t *testing.T) {
	got := DefaultConfigJSON()
	want := "{\n  \"debug\": false\n}"
	if got != want {
		t.Errorf("DefaultConfigJSON() = %q, want %q", got, want)
	}
}

func TestWriteConfig_ToggleOff(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmp)

	WriteConfig(Config{Debug: true})
	WriteConfig(Config{Debug: false})

	got := ReadConfig()
	if got.Debug {
		t.Error("expected Debug=false after toggling off")
	}
}
