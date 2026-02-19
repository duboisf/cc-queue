package queue

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

var testShortcuts = KittyShortcuts{
	Picker: "kitty_mod+shift+q",
	First:  "kitty_mod+shift+u",
}

func TestInstallKittyShortcut_FreshInstall(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	kittyDir := filepath.Join(tmp, ".config", "kitty")
	if err := os.MkdirAll(kittyDir, 0755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}

	path, err := InstallKittyShortcut(testShortcuts)
	if err != nil {
		t.Fatalf("InstallKittyShortcut: %v", err)
	}

	wantPath := filepath.Join(kittyDir, "kitty.conf")
	if path != wantPath {
		t.Errorf("path = %q, want %q", path, wantPath)
	}

	content, err := os.ReadFile(wantPath)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}

	if !strings.Contains(string(content), "map kitty_mod+shift+q") {
		t.Error("missing picker shortcut")
	}
	if !strings.Contains(string(content), "map kitty_mod+shift+u") {
		t.Error("missing first shortcut")
	}
	if !strings.Contains(string(content), "cc-queue first") {
		t.Error("missing cc-queue first in shortcut")
	}
}

func TestInstallKittyShortcut_PickerOnly(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	kittyDir := filepath.Join(tmp, ".config", "kitty")
	os.MkdirAll(kittyDir, 0755)

	path, err := InstallKittyShortcut(KittyShortcuts{Picker: "ctrl+shift+p"})
	if err != nil {
		t.Fatalf("InstallKittyShortcut: %v", err)
	}
	if path == "" {
		t.Fatal("expected non-empty path")
	}

	content, _ := os.ReadFile(path)
	s := string(content)
	if !strings.Contains(s, "map ctrl+shift+p") {
		t.Error("missing picker shortcut")
	}
	if strings.Contains(s, "cc-queue first") {
		t.Error("first shortcut should not be present when not requested")
	}
}

func TestInstallKittyShortcut_NoShortcuts(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	kittyDir := filepath.Join(tmp, ".config", "kitty")
	os.MkdirAll(kittyDir, 0755)

	path, err := InstallKittyShortcut(KittyShortcuts{})
	if err != nil {
		t.Fatalf("InstallKittyShortcut: %v", err)
	}
	if path != "" {
		t.Errorf("path = %q, want empty (no shortcuts requested)", path)
	}
}

func TestInstallKittyShortcut_Idempotent(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	kittyDir := filepath.Join(tmp, ".config", "kitty")
	os.MkdirAll(kittyDir, 0755)

	// First install.
	path1, err := InstallKittyShortcut(testShortcuts)
	if err != nil {
		t.Fatalf("first install: %v", err)
	}
	if path1 == "" {
		t.Fatal("first install returned empty path")
	}

	content1, _ := os.ReadFile(path1)

	// Second install should be a no-op.
	path2, err := InstallKittyShortcut(testShortcuts)
	if err != nil {
		t.Fatalf("second install: %v", err)
	}
	if path2 != "" {
		t.Errorf("second install returned path %q, want empty (already installed)", path2)
	}

	content2, _ := os.ReadFile(filepath.Join(kittyDir, "kitty.conf"))
	if string(content1) != string(content2) {
		t.Error("kitty.conf content changed on second install")
	}
}

func TestInstallKittyShortcut_MissingKittyDir(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	// No kitty dir created — should skip.
	path, err := InstallKittyShortcut(testShortcuts)
	if err != nil {
		t.Fatalf("InstallKittyShortcut: %v", err)
	}
	if path != "" {
		t.Errorf("path = %q, want empty (no kitty dir)", path)
	}
}

func TestInstallKittyShortcut_ExistingConf(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	kittyDir := filepath.Join(tmp, ".config", "kitty")
	os.MkdirAll(kittyDir, 0755)

	// Write existing config.
	existing := "font_size 12\ninclude themes/mocha.conf\n"
	confPath := filepath.Join(kittyDir, "kitty.conf")
	os.WriteFile(confPath, []byte(existing), 0644)

	path, err := InstallKittyShortcut(testShortcuts)
	if err != nil {
		t.Fatalf("InstallKittyShortcut: %v", err)
	}
	if path == "" {
		t.Fatal("expected non-empty path")
	}

	content, _ := os.ReadFile(confPath)
	s := string(content)

	// Existing content preserved.
	if !strings.Contains(s, "font_size 12") {
		t.Error("existing config was overwritten")
	}
	// Shortcuts appended.
	if !strings.Contains(s, "cc-queue keyboard shortcuts") {
		t.Error("shortcut block not appended")
	}
}
