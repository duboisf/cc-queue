package cmd_test

import (
	"strings"
	"testing"

	"github.com/duboisf/cc-queue/cmd"
	"github.com/duboisf/cc-queue/internal/queue"
)

func TestConfigCmd_NoArgs_ShowsPathAndDefaults(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmp)
	setupQueueDir(t)

	opts, stdout, _ := testOptions()
	root := cmd.NewRootCmd(opts)

	_, _, err := executeCommand(root, "config")
	if err != nil {
		t.Fatalf("config: %v", err)
	}
	got := stdout.String()
	if !strings.Contains(got, "Config:") {
		t.Errorf("expected output to contain config path header, got %q", got)
	}
	if !strings.Contains(got, queue.DefaultConfigJSON()) {
		t.Errorf("expected default config JSON, got %q", got)
	}
}

func TestConfigViewCmd_ShowsPathAndDefaults(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmp)
	setupQueueDir(t)

	opts, stdout, _ := testOptions()
	root := cmd.NewRootCmd(opts)

	_, _, err := executeCommand(root, "config", "view")
	if err != nil {
		t.Fatalf("config view: %v", err)
	}
	got := stdout.String()
	if !strings.Contains(got, "Config:") {
		t.Errorf("expected output to contain config path header, got %q", got)
	}
	if !strings.Contains(got, queue.DefaultConfigJSON()) {
		t.Errorf("expected default config JSON, got %q", got)
	}
}

func TestConfigViewCmd_WithExistingConfig(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmp)
	setupQueueDir(t)

	// Write a config with debug enabled.
	if err := queue.WriteConfig(queue.Config{Debug: true}); err != nil {
		t.Fatalf("WriteConfig: %v", err)
	}

	opts, stdout, _ := testOptions()
	root := cmd.NewRootCmd(opts)

	_, _, err := executeCommand(root, "config", "view")
	if err != nil {
		t.Fatalf("config view: %v", err)
	}
	got := stdout.String()
	if !strings.Contains(got, `"debug": true`) {
		t.Errorf("expected debug: true in output, got %q", got)
	}
}

func TestConfigEditCmd_NoEditor(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmp)
	t.Setenv("VISUAL", "")
	t.Setenv("EDITOR", "")
	setupQueueDir(t)

	opts, _, _ := testOptions()
	root := cmd.NewRootCmd(opts)

	_, _, err := executeCommand(root, "config", "edit")
	if err == nil {
		t.Fatal("expected error when no editor is set")
	}
	if !strings.Contains(err.Error(), "$VISUAL or $EDITOR must be set") {
		t.Errorf("error = %q, want contains %q", err.Error(), "$VISUAL or $EDITOR must be set")
	}
}
