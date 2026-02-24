package cmd_test

import (
	"os"
	"path/filepath"
	"runtime"
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

func TestInstall_SkipsExistingWithoutForce(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	kittyDir := filepath.Join(tmpDir, ".config", "kitty")
	os.MkdirAll(kittyDir, 0755)

	// First install creates cc-queue.conf with shortcuts.
	opts1, _, _ := testOptions()
	root1 := cmd.NewRootCmd(opts1)
	_, _, err := executeCommand(root1, "install",
		"--picker-shortcut", "kitty_mod+shift+q",
	)
	if err != nil {
		t.Fatalf("first install: %v", err)
	}

	// Second install without --force — should skip cc-queue.conf.
	opts2, stdout2, _ := testOptions()
	root2 := cmd.NewRootCmd(opts2)
	_, _, err = executeCommand(root2, "install")
	if err != nil {
		t.Fatalf("second install: %v", err)
	}

	got := stdout2.String()
	if !strings.Contains(got, "already exists, skipping") {
		t.Errorf("expected skip message, got:\n%s", got)
	}

	// Original shortcuts should be preserved.
	content, _ := os.ReadFile(filepath.Join(kittyDir, "cc-queue.conf"))
	if !strings.Contains(string(content), "kitty_mod+shift+q") {
		t.Error("original shortcuts were overwritten")
	}
}

func TestInstall_ForceOverwritePreservesShortcuts(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	kittyDir := filepath.Join(tmpDir, ".config", "kitty")
	os.MkdirAll(kittyDir, 0755)

	// First install with shortcuts.
	opts1, _, _ := testOptions()
	root1 := cmd.NewRootCmd(opts1)
	_, _, err := executeCommand(root1, "install",
		"--picker-shortcut", "kitty_mod+shift+q",
		"--first-shortcut", "kitty_mod+shift+u",
	)
	if err != nil {
		t.Fatalf("first install: %v", err)
	}

	// Second install with --force but no shortcut flags — should preserve.
	opts2, stdout2, _ := testOptions()
	root2 := cmd.NewRootCmd(opts2)
	_, _, err = executeCommand(root2, "install", "--force")
	if err != nil {
		t.Fatalf("second install: %v", err)
	}

	got := stdout2.String()
	if strings.Contains(got, "already exists, skipping") {
		t.Error("should not skip when --force is used")
	}
	if !strings.Contains(got, "Created") {
		t.Errorf("expected Created message, got:\n%s", got)
	}

	// Shortcuts should be preserved since flags weren't passed.
	content, _ := os.ReadFile(filepath.Join(kittyDir, "cc-queue.conf"))
	s := string(content)
	if !strings.Contains(s, "kitty_mod+shift+q") {
		t.Error("picker shortcut lost during --force without flag")
	}
	if !strings.Contains(s, "kitty_mod+shift+u") {
		t.Error("first shortcut lost during --force without flag")
	}
}

func TestInstall_DesktopEntry(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	t.Setenv("SHELL", "/bin/zsh")

	opts, stdout, _ := testOptions()
	root := cmd.NewRootCmd(opts)

	_, _, err := executeCommand(root, "install", "--desktop")
	if err != nil {
		t.Fatalf("install --desktop returned error: %v", err)
	}

	got := stdout.String()
	if !strings.Contains(got, "Created") || !strings.Contains(got, "cc-queue.desktop") {
		t.Errorf("stdout missing desktop creation message:\n%s", got)
	}

	desktopPath := filepath.Join(tmpDir, ".local", "share", "applications", "cc-queue.desktop")
	content, err := os.ReadFile(desktopPath)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if !strings.Contains(string(content), "Exec=kitty") {
		t.Error("desktop file missing Exec line")
	}
	if !strings.Contains(string(content), "/bin/zsh") {
		t.Error("desktop file missing shell")
	}
}

func TestInstall_DesktopEntryAutoOnLinux(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("desktop auto-install only runs on Linux")
	}
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	t.Setenv("SHELL", "/bin/zsh")

	opts, stdout, _ := testOptions()
	root := cmd.NewRootCmd(opts)

	// No --desktop flag — should auto-install on Linux.
	_, _, err := executeCommand(root, "install")
	if err != nil {
		t.Fatalf("install returned error: %v", err)
	}

	got := stdout.String()
	if !strings.Contains(got, "cc-queue.desktop") {
		t.Errorf("expected desktop entry creation message on Linux:\n%s", got)
	}

	desktopPath := filepath.Join(tmpDir, ".local", "share", "applications", "cc-queue.desktop")
	if _, err := os.Stat(desktopPath); os.IsNotExist(err) {
		t.Error("desktop file not auto-created on Linux")
	}
}

func TestInstall_DesktopEntrySkipsWithoutForce(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	// First install.
	opts1, _, _ := testOptions()
	root1 := cmd.NewRootCmd(opts1)
	_, _, err := executeCommand(root1, "install", "--desktop")
	if err != nil {
		t.Fatalf("first install: %v", err)
	}

	// Second install without --force.
	opts2, stdout2, _ := testOptions()
	root2 := cmd.NewRootCmd(opts2)
	_, _, err = executeCommand(root2, "install", "--desktop")
	if err != nil {
		t.Fatalf("second install: %v", err)
	}

	got := stdout2.String()
	if !strings.Contains(got, "cc-queue.desktop already exists, skipping") {
		t.Errorf("expected skip message, got:\n%s", got)
	}
}

func TestInstall_ForceOverrideOneShortcut(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	kittyDir := filepath.Join(tmpDir, ".config", "kitty")
	os.MkdirAll(kittyDir, 0755)

	// First install with both shortcuts.
	opts1, _, _ := testOptions()
	root1 := cmd.NewRootCmd(opts1)
	_, _, err := executeCommand(root1, "install",
		"--picker-shortcut", "kitty_mod+shift+q",
		"--first-shortcut", "kitty_mod+shift+u",
	)
	if err != nil {
		t.Fatalf("first install: %v", err)
	}

	// Second install with --force, only changing picker.
	opts2, _, _ := testOptions()
	root2 := cmd.NewRootCmd(opts2)
	_, _, err = executeCommand(root2, "install", "--force",
		"--picker-shortcut", "ctrl+alt+p",
	)
	if err != nil {
		t.Fatalf("second install: %v", err)
	}

	content, _ := os.ReadFile(filepath.Join(kittyDir, "cc-queue.conf"))
	s := string(content)

	// Picker should be updated.
	if !strings.Contains(s, "ctrl+alt+p") {
		t.Error("picker shortcut not updated")
	}
	if strings.Contains(s, "kitty_mod+shift+q") {
		t.Error("old picker shortcut still present")
	}
	// First shortcut should be preserved.
	if !strings.Contains(s, "kitty_mod+shift+u") {
		t.Error("first shortcut lost when only picker was overridden")
	}
}
