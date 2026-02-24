package queue

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBuildDesktopEntry_WithShell(t *testing.T) {
	got := BuildDesktopEntry("/bin/zsh")

	wantLines := []string{
		"[Desktop Entry]",
		"Name=cc-queue",
		"Comment=Claude Code input queue for kitty",
		"Exec=kitty --detach --title cc-queue -- /bin/zsh -ilc 'exec cc-queue'",
		"Type=Application",
		"Terminal=false",
		"Categories=Development;",
		"Keywords=claude;code;queue;",
	}

	for _, want := range wantLines {
		if !strings.Contains(got, want) {
			t.Errorf("BuildDesktopEntry missing line %q\ngot:\n%s", want, got)
		}
	}
}

func TestBuildDesktopEntry_BashUsesIC(t *testing.T) {
	got := BuildDesktopEntry("/bin/bash")

	want := "Exec=kitty --detach --title cc-queue -- /bin/bash -ic 'exec cc-queue'"
	if !strings.Contains(got, want) {
		t.Errorf("bash should use -ic, not -ilc\ngot:\n%s", got)
	}
	if strings.Contains(got, "-ilc") {
		t.Errorf("bash should not use -ilc\ngot:\n%s", got)
	}
}

func TestBuildDesktopEntry_FallbackShell(t *testing.T) {
	got := BuildDesktopEntry("")

	want := "Exec=kitty --detach --title cc-queue -- /bin/sh -ic 'exec cc-queue'"
	if !strings.Contains(got, want) {
		t.Errorf("expected fallback to /bin/sh with -ic\ngot:\n%s", got)
	}
}

func TestInstallDesktopEntry_FreshInstall(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	appsDir := filepath.Join(tmpDir, ".local", "share", "applications")
	if err := os.MkdirAll(appsDir, 0755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}

	result, err := InstallDesktopEntry("/bin/zsh", false)
	if err != nil {
		t.Fatalf("InstallDesktopEntry: %v", err)
	}

	if result.Skipped {
		t.Error("expected Created, got Skipped")
	}

	wantPath := filepath.Join(appsDir, "cc-queue.desktop")
	if result.Path != wantPath {
		t.Errorf("path = %q, want %q", result.Path, wantPath)
	}

	content, err := os.ReadFile(wantPath)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if !strings.Contains(string(content), "/bin/zsh") {
		t.Error("desktop file missing shell")
	}
}

func TestInstallDesktopEntry_CreatesAppsDir(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	// Don't pre-create the applications directory.
	result, err := InstallDesktopEntry("/bin/zsh", false)
	if err != nil {
		t.Fatalf("InstallDesktopEntry: %v", err)
	}

	if result.Skipped {
		t.Error("expected Created, got Skipped")
	}

	if _, err := os.Stat(result.Path); os.IsNotExist(err) {
		t.Errorf("desktop file not created at %s", result.Path)
	}
}

func TestInstallDesktopEntry_SkipsExistingWithoutForce(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	appsDir := filepath.Join(tmpDir, ".local", "share", "applications")
	os.MkdirAll(appsDir, 0755)

	// Create an existing file.
	existing := filepath.Join(appsDir, "cc-queue.desktop")
	os.WriteFile(existing, []byte("original"), 0644)

	result, err := InstallDesktopEntry("/bin/zsh", false)
	if err != nil {
		t.Fatalf("InstallDesktopEntry: %v", err)
	}

	if !result.Skipped {
		t.Error("expected Skipped when file exists without force")
	}

	// Content should be unchanged.
	content, _ := os.ReadFile(existing)
	if string(content) != "original" {
		t.Error("file was overwritten without --force")
	}
}

func TestInstallDesktopEntry_ForceOverwrites(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	appsDir := filepath.Join(tmpDir, ".local", "share", "applications")
	os.MkdirAll(appsDir, 0755)

	existing := filepath.Join(appsDir, "cc-queue.desktop")
	os.WriteFile(existing, []byte("original"), 0644)

	result, err := InstallDesktopEntry("/bin/zsh", true)
	if err != nil {
		t.Fatalf("InstallDesktopEntry: %v", err)
	}

	if result.Skipped {
		t.Error("expected overwrite with force, got Skipped")
	}

	content, _ := os.ReadFile(existing)
	if string(content) == "original" {
		t.Error("file was not overwritten with --force")
	}
	if !strings.Contains(string(content), "[Desktop Entry]") {
		t.Error("overwritten file missing desktop entry header")
	}
}
