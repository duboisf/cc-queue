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

func TestInstall_KittyShortcuts(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	// Create kitty config dir so shortcuts get installed.
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
	if !strings.Contains(got, "Kitty shortcuts installed in") {
		t.Errorf("stdout = %q, want it to contain kitty install message", got)
	}

	// Verify kitty.conf was created with shortcuts.
	confPath := filepath.Join(kittyDir, "kitty.conf")
	content, err := os.ReadFile(confPath)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if !strings.Contains(string(content), "cc-queue") {
		t.Error("kitty.conf missing cc-queue shortcuts")
	}
}

func TestInstall_NoShortcutFlags(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	// Create kitty dir — but don't pass shortcut flags.
	kittyDir := filepath.Join(tmpDir, ".config", "kitty")
	os.MkdirAll(kittyDir, 0755)

	opts, stdout, _ := testOptions()
	root := cmd.NewRootCmd(opts)

	_, _, err := executeCommand(root, "install")
	if err != nil {
		t.Fatalf("install returned error: %v", err)
	}

	got := stdout.String()
	if strings.Contains(got, "Kitty shortcuts") {
		t.Errorf("stdout = %q, should not mention kitty shortcuts when no flags provided", got)
	}
}
