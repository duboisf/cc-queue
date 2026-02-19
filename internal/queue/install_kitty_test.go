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

func TestBuildKittyConfig_BothShortcuts(t *testing.T) {
	content := BuildKittyConfig(testShortcuts)

	if !strings.Contains(content, "allow_remote_control socket-only") {
		t.Error("missing allow_remote_control")
	}
	if !strings.Contains(content, "listen_on unix:/tmp/kitty-{kitty_pid}") {
		t.Error("missing listen_on")
	}
	if !strings.Contains(content, "map kitty_mod+shift+q") {
		t.Error("missing picker shortcut")
	}
	if !strings.Contains(content, "map kitty_mod+shift+u") {
		t.Error("missing first shortcut")
	}
}

func TestBuildKittyConfig_PickerOnly(t *testing.T) {
	content := BuildKittyConfig(KittyShortcuts{Picker: "ctrl+shift+p"})

	if !strings.Contains(content, "map ctrl+shift+p") {
		t.Error("missing picker shortcut")
	}
	if strings.Contains(content, "cc-queue first") {
		t.Error("first shortcut present when not requested")
	}
}

func TestBuildKittyConfig_NoShortcuts(t *testing.T) {
	content := BuildKittyConfig(KittyShortcuts{})

	if !strings.Contains(content, "allow_remote_control") {
		t.Error("missing remote control settings")
	}
	if strings.Contains(content, "Keyboard shortcuts") {
		t.Error("shortcuts section present when none requested")
	}
}

func TestInstallKittyConfig_FreshInstall(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	kittyDir := filepath.Join(tmp, ".config", "kitty")
	if err := os.MkdirAll(kittyDir, 0755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}

	result, err := InstallKittyConfig(testShortcuts)
	if err != nil {
		t.Fatalf("InstallKittyConfig: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}

	// cc-queue.conf created with correct content.
	wantPath := filepath.Join(kittyDir, "cc-queue.conf")
	if result.ConfPath != wantPath {
		t.Errorf("ConfPath = %q, want %q", result.ConfPath, wantPath)
	}
	content, err := os.ReadFile(wantPath)
	if err != nil {
		t.Fatalf("ReadFile cc-queue.conf: %v", err)
	}
	if !strings.Contains(string(content), "map kitty_mod+shift+q") {
		t.Error("cc-queue.conf missing picker shortcut")
	}

	// include added to kitty.conf.
	if !result.Included {
		t.Error("expected Included=true on fresh install")
	}
	mainContent, err := os.ReadFile(filepath.Join(kittyDir, "kitty.conf"))
	if err != nil {
		t.Fatalf("ReadFile kitty.conf: %v", err)
	}
	if !strings.Contains(string(mainContent), "include cc-queue.conf") {
		t.Error("kitty.conf missing include directive")
	}
}

func TestInstallKittyConfig_ReplacesExisting(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	kittyDir := filepath.Join(tmp, ".config", "kitty")
	os.MkdirAll(kittyDir, 0755)

	// First install.
	_, err := InstallKittyConfig(testShortcuts)
	if err != nil {
		t.Fatalf("first install: %v", err)
	}

	// Second install with different shortcuts — cc-queue.conf overwritten.
	newShortcuts := KittyShortcuts{Picker: "ctrl+alt+p", First: "ctrl+alt+f"}
	result, err := InstallKittyConfig(newShortcuts)
	if err != nil {
		t.Fatalf("second install: %v", err)
	}

	content, _ := os.ReadFile(result.ConfPath)
	s := string(content)

	if strings.Contains(s, "kitty_mod+shift+q") {
		t.Error("old picker shortcut still present")
	}
	if !strings.Contains(s, "map ctrl+alt+p") {
		t.Error("new picker shortcut missing")
	}

	// include should not be duplicated.
	mainContent, _ := os.ReadFile(filepath.Join(kittyDir, "kitty.conf"))
	if strings.Count(string(mainContent), "include cc-queue.conf") != 1 {
		t.Error("include line duplicated in kitty.conf")
	}
}

func TestInstallKittyConfig_PreservesExistingConf(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	kittyDir := filepath.Join(tmp, ".config", "kitty")
	os.MkdirAll(kittyDir, 0755)

	existing := "font_size 12\ninclude themes/mocha.conf\n"
	mainConf := filepath.Join(kittyDir, "kitty.conf")
	os.WriteFile(mainConf, []byte(existing), 0644)

	result, err := InstallKittyConfig(testShortcuts)
	if err != nil {
		t.Fatalf("InstallKittyConfig: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}

	mainContent, _ := os.ReadFile(mainConf)
	s := string(mainContent)

	if !strings.Contains(s, "font_size 12") {
		t.Error("existing config overwritten")
	}
	if !strings.Contains(s, "include themes/mocha.conf") {
		t.Error("existing include lost")
	}
	if !strings.Contains(s, "include cc-queue.conf") {
		t.Error("cc-queue include not added")
	}
}

func TestInstallKittyConfig_CleansLegacyBlock(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	kittyDir := filepath.Join(tmp, ".config", "kitty")
	os.MkdirAll(kittyDir, 0755)

	// Simulate a legacy install that wrote directly to kitty.conf.
	legacy := "font_size 12\n\n# cc-queue keyboard shortcuts\nmap kitty_mod+q launch --type=overlay --title cc-queue cc-queue\n"
	mainConf := filepath.Join(kittyDir, "kitty.conf")
	os.WriteFile(mainConf, []byte(legacy), 0644)

	_, err := InstallKittyConfig(testShortcuts)
	if err != nil {
		t.Fatalf("InstallKittyConfig: %v", err)
	}

	mainContent, _ := os.ReadFile(mainConf)
	s := string(mainContent)

	if strings.Contains(s, "map kitty_mod+q") {
		t.Error("legacy block not cleaned from kitty.conf")
	}
	if !strings.Contains(s, "include cc-queue.conf") {
		t.Error("include not added after legacy cleanup")
	}
}

func TestInstallKittyConfig_MissingKittyDir(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	result, err := InstallKittyConfig(testShortcuts)
	if err != nil {
		t.Fatalf("InstallKittyConfig: %v", err)
	}
	if result != nil {
		t.Errorf("expected nil result when kitty dir missing, got %+v", result)
	}
}
