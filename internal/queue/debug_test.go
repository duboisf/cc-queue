package queue

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDebugEnabled_ConfigFile(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmp)

	if DebugEnabled() {
		t.Error("expected disabled before config is written")
	}

	WriteConfig(Config{Debug: true})

	if !DebugEnabled() {
		t.Error("expected enabled after config Debug=true")
	}
}

func TestDebugEnabled_Disabled(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmp)

	if DebugEnabled() {
		t.Error("expected disabled when no config exists")
	}
}

func TestDebugEnabled_ExplicitlyOff(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmp)

	WriteConfig(Config{Debug: false})

	if DebugEnabled() {
		t.Error("expected disabled when config Debug=false")
	}
}

func TestDebugf_WritesToLog(t *testing.T) {
	cfgTmp := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", cfgTmp)
	stateTmp := t.TempDir()
	t.Setenv("XDG_STATE_HOME", stateTmp)

	WriteConfig(Config{Debug: true})
	EnsureDir()

	Debugf("test message %s", "hello")

	logPath := filepath.Join(Dir(), "debug.log")
	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("reading debug log: %v", err)
	}
	if len(data) == 0 {
		t.Error("expected non-empty debug log")
	}
}

func TestDebugf_NoLogWhenDisabled(t *testing.T) {
	cfgTmp := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", cfgTmp)
	stateTmp := t.TempDir()
	t.Setenv("XDG_STATE_HOME", stateTmp)

	EnsureDir()

	Debugf("should not appear")

	logPath := filepath.Join(Dir(), "debug.log")
	if _, err := os.Stat(logPath); err == nil {
		t.Error("debug log should not exist when debug is disabled")
	}
}
