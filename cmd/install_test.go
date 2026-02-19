package cmd_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/duboisf/cc-queue/cmd"
)

func TestInstall_DefaultUser(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	opts, _, _ := testOptions()
	root := cmd.NewRootCmd(opts)

	_, _, err := executeCommand(root, "install")
	if err != nil {
		t.Fatalf("install returned error: %v", err)
	}

	settingsPath := filepath.Join(tmpDir, ".claude", "settings.json")
	if _, err := os.Stat(settingsPath); os.IsNotExist(err) {
		t.Errorf("settings file not created at %s", settingsPath)
	}
}

func TestInstall_ProjectFlag(t *testing.T) {
	tmpDir := t.TempDir()

	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd: %v", err)
	}
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Chdir: %v", err)
	}
	t.Cleanup(func() { os.Chdir(origDir) })

	opts, _, _ := testOptions()
	root := cmd.NewRootCmd(opts)

	_, _, err = executeCommand(root, "install", "--project")
	if err != nil {
		t.Fatalf("install --project returned error: %v", err)
	}

	settingsPath := filepath.Join(tmpDir, ".claude", "settings.json")
	if _, err := os.Stat(settingsPath); os.IsNotExist(err) {
		t.Errorf("settings file not created at %s", settingsPath)
	}
}

func TestInstall_OutputMessage(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	opts, stdout, _ := testOptions()
	root := cmd.NewRootCmd(opts)

	_, _, err := executeCommand(root, "install")
	if err != nil {
		t.Fatalf("install returned error: %v", err)
	}

	if got := stdout.String(); !strings.Contains(got, "Hooks installed in") {
		t.Errorf("stdout = %q, want it to contain %q", got, "Hooks installed in")
	}
}

func TestInstall_KittyConfig(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	kittyDir := filepath.Join(tmpDir, ".config", "kitty")
	if err := os.MkdirAll(kittyDir, 0755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}

	opts, stdout, _ := testOptions()
	root := cmd.NewRootCmd(opts)

	_, _, err := executeCommand(root, "install",
		"--picker-shortcut", "kitty_mod+shift+q",
		"--first-shortcut", "kitty_mod+shift+u",
	)
	if err != nil {
		t.Fatalf("install returned error: %v", err)
	}

	got := stdout.String()
	// Should show content preview.
	if !strings.Contains(got, "Creating kitty config") {
		t.Errorf("stdout missing config preview:\n%s", got)
	}
	if !strings.Contains(got, "allow_remote_control") {
		t.Errorf("stdout missing remote control in preview:\n%s", got)
	}
	// Should confirm file creation.
	if !strings.Contains(got, "Created") {
		t.Errorf("stdout missing creation confirmation:\n%s", got)
	}

	// Verify cc-queue.conf was created.
	confPath := filepath.Join(kittyDir, "cc-queue.conf")
	content, err := os.ReadFile(confPath)
	if err != nil {
		t.Fatalf("ReadFile cc-queue.conf: %v", err)
	}
	if !strings.Contains(string(content), "kitty_mod+shift+q") {
		t.Error("cc-queue.conf missing picker shortcut")
	}

	// Verify include in kitty.conf.
	mainContent, err := os.ReadFile(filepath.Join(kittyDir, "kitty.conf"))
	if err != nil {
		t.Fatalf("ReadFile kitty.conf: %v", err)
	}
	if !strings.Contains(string(mainContent), "include cc-queue.conf") {
		t.Error("kitty.conf missing include directive")
	}
}

func TestInstall_NoShortcutFlags(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	kittyDir := filepath.Join(tmpDir, ".config", "kitty")
	os.MkdirAll(kittyDir, 0755)

	opts, stdout, _ := testOptions()
	root := cmd.NewRootCmd(opts)

	_, _, err := executeCommand(root, "install")
	if err != nil {
		t.Fatalf("install returned error: %v", err)
	}

	got := stdout.String()
	// Should still show kitty config preview (remote control settings).
	if !strings.Contains(got, "Creating kitty config") {
		t.Errorf("stdout missing config preview:\n%s", got)
	}
}
