package cmd_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/duboisf/cc-queue/cmd"
)

func TestHooksInstall_DefaultUser(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	opts, stdout, _ := testOptions()
	root := cmd.NewRootCmd(opts)

	_, _, err := executeCommand(root, "hooks", "install")
	if err != nil {
		t.Fatalf("hooks install returned error: %v", err)
	}

	got := stdout.String()
	if !strings.Contains(got, "Hooks installed in") {
		t.Errorf("stdout = %q, want it to contain 'Hooks installed in'", got)
	}

	// Verify all four hooks are present.
	settingsPath := filepath.Join(tmpDir, ".claude", "settings.json")
	data, err := os.ReadFile(settingsPath)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}

	var settings map[string]any
	if err := json.Unmarshal(data, &settings); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	hooks, _ := settings["hooks"].(map[string]any)
	for _, key := range []string{"Notification", "UserPromptSubmit", "SessionStart", "SessionEnd"} {
		if _, ok := hooks[key]; !ok {
			t.Errorf("missing hook: %s", key)
		}
	}
}

func TestHooksInstall_Idempotent(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	// First install.
	opts1, _, _ := testOptions()
	root1 := cmd.NewRootCmd(opts1)
	_, _, err := executeCommand(root1, "hooks", "install")
	if err != nil {
		t.Fatalf("first install: %v", err)
	}

	// Read settings after first install.
	settingsPath := filepath.Join(tmpDir, ".claude", "settings.json")
	data1, _ := os.ReadFile(settingsPath)

	// Second install.
	opts2, stdout2, _ := testOptions()
	root2 := cmd.NewRootCmd(opts2)
	_, _, err = executeCommand(root2, "hooks", "install")
	if err != nil {
		t.Fatalf("second install: %v", err)
	}

	got := stdout2.String()
	if !strings.Contains(got, "already installed") {
		t.Errorf("expected 'already installed' message, got: %q", got)
	}

	// Settings should be unchanged.
	data2, _ := os.ReadFile(settingsPath)
	if string(data1) != string(data2) {
		t.Error("settings changed on second install")
	}
}

func TestHooksInstall_ProjectFlag(t *testing.T) {
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

	_, _, err = executeCommand(root, "hooks", "install", "--project")
	if err != nil {
		t.Fatalf("hooks install --project returned error: %v", err)
	}

	settingsPath := filepath.Join(tmpDir, ".claude", "settings.json")
	if _, err := os.Stat(settingsPath); os.IsNotExist(err) {
		t.Errorf("settings file not created at %s", settingsPath)
	}
}

func TestHooks_ShowsStatus(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	// Install hooks first.
	opts1, _, _ := testOptions()
	root1 := cmd.NewRootCmd(opts1)
	_, _, err := executeCommand(root1, "hooks", "install")
	if err != nil {
		t.Fatalf("install: %v", err)
	}

	// Check status.
	opts2, stdout2, _ := testOptions()
	root2 := cmd.NewRootCmd(opts2)
	_, _, err = executeCommand(root2, "hooks")
	if err != nil {
		t.Fatalf("hooks returned error: %v", err)
	}

	got := stdout2.String()
	// All four should show as installed.
	for _, hook := range []string{"Notification", "UserPromptSubmit", "SessionStart", "SessionEnd"} {
		expected := "[v] " + hook
		if !strings.Contains(got, expected) {
			t.Errorf("missing %q in output:\n%s", expected, got)
		}
	}

	// Should NOT show the "install missing" hint.
	if strings.Contains(got, "install missing") {
		t.Errorf("should not suggest installing when all hooks present:\n%s", got)
	}
}

func TestHooksUninstall_RemovesHooks(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	// Install hooks first.
	opts1, _, _ := testOptions()
	root1 := cmd.NewRootCmd(opts1)
	_, _, err := executeCommand(root1, "hooks", "install")
	if err != nil {
		t.Fatalf("install: %v", err)
	}

	// Uninstall.
	opts2, stdout2, _ := testOptions()
	root2 := cmd.NewRootCmd(opts2)
	_, _, err = executeCommand(root2, "hooks", "uninstall")
	if err != nil {
		t.Fatalf("uninstall: %v", err)
	}

	got := stdout2.String()
	if !strings.Contains(got, "Hooks removed from") {
		t.Errorf("expected removal message, got: %q", got)
	}

	// Verify hooks are gone.
	settingsPath := filepath.Join(tmpDir, ".claude", "settings.json")
	data, _ := os.ReadFile(settingsPath)
	var settings map[string]any
	json.Unmarshal(data, &settings)

	hooks, _ := settings["hooks"].(map[string]any)
	for _, key := range []string{"Notification", "UserPromptSubmit", "SessionStart", "SessionEnd"} {
		if _, ok := hooks[key]; ok {
			t.Errorf("hook %s still present after uninstall", key)
		}
	}
}

func TestHooksUninstall_NoHooksPresent(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	// Create empty settings.
	settingsDir := filepath.Join(tmpDir, ".claude")
	os.MkdirAll(settingsDir, 0755)
	os.WriteFile(filepath.Join(settingsDir, "settings.json"), []byte("{}"), 0644)

	opts, stdout, _ := testOptions()
	root := cmd.NewRootCmd(opts)
	_, _, err := executeCommand(root, "hooks", "uninstall")
	if err != nil {
		t.Fatalf("uninstall: %v", err)
	}

	got := stdout.String()
	if !strings.Contains(got, "No cc-queue hooks found") {
		t.Errorf("expected 'no hooks' message, got: %q", got)
	}
}

func TestHooksUninstall_PreservesOtherHooks(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	// Write settings with both cc-queue and other hooks.
	settingsDir := filepath.Join(tmpDir, ".claude")
	os.MkdirAll(settingsDir, 0755)
	settings := map[string]any{
		"hooks": map[string]any{
			"UserPromptSubmit": []any{
				map[string]any{
					"matcher": "",
					"hooks": []any{
						map[string]any{
							"type":    "command",
							"command": "other-tool run",
						},
						map[string]any{
							"type":    "command",
							"command": "cc-queue pop",
						},
					},
				},
			},
			"Notification": []any{
				map[string]any{
					"matcher": "permission_prompt|idle_prompt|elicitation_dialog",
					"hooks": []any{
						map[string]any{
							"type":    "command",
							"command": "cc-queue push",
						},
					},
				},
			},
		},
	}
	data, _ := json.MarshalIndent(settings, "", "  ")
	os.WriteFile(filepath.Join(settingsDir, "settings.json"), data, 0644)

	opts, _, _ := testOptions()
	root := cmd.NewRootCmd(opts)
	_, _, err := executeCommand(root, "hooks", "uninstall")
	if err != nil {
		t.Fatalf("uninstall: %v", err)
	}

	// Read back and verify other-tool hook is preserved.
	result, _ := os.ReadFile(filepath.Join(settingsDir, "settings.json"))
	var updated map[string]any
	json.Unmarshal(result, &updated)

	hooks, _ := updated["hooks"].(map[string]any)

	// Notification should be removed entirely (only had cc-queue).
	if _, ok := hooks["Notification"]; ok {
		t.Error("Notification hook should be removed")
	}

	// UserPromptSubmit should still exist with other-tool.
	ups, ok := hooks["UserPromptSubmit"].([]any)
	if !ok || len(ups) == 0 {
		t.Fatal("UserPromptSubmit should still exist with other-tool hook")
	}
	matcher := ups[0].(map[string]any)
	hooksList := matcher["hooks"].([]any)
	if len(hooksList) != 1 {
		t.Errorf("expected 1 hook remaining, got %d", len(hooksList))
	}
	hook := hooksList[0].(map[string]any)
	if hook["command"] != "other-tool run" {
		t.Errorf("remaining hook should be other-tool, got: %v", hook["command"])
	}
}

func TestHooks_DetectsMissing(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	// Write partial settings with only Notification hook.
	settingsDir := filepath.Join(tmpDir, ".claude")
	os.MkdirAll(settingsDir, 0755)
	partial := map[string]any{
		"hooks": map[string]any{
			"Notification": []any{
				map[string]any{
					"matcher": "permission_prompt|idle_prompt|elicitation_dialog",
					"hooks": []any{
						map[string]any{
							"type":    "command",
							"command": "cc-queue push",
						},
					},
				},
			},
		},
	}
	data, _ := json.MarshalIndent(partial, "", "  ")
	os.WriteFile(filepath.Join(settingsDir, "settings.json"), data, 0644)

	opts, stdout, _ := testOptions()
	root := cmd.NewRootCmd(opts)

	_, _, err := executeCommand(root, "hooks")
	if err != nil {
		t.Fatalf("hooks returned error: %v", err)
	}

	got := stdout.String()

	// Notification should be installed.
	if !strings.Contains(got, "[v] Notification") {
		t.Errorf("Notification should be installed:\n%s", got)
	}

	// Others should be missing.
	for _, hook := range []string{"UserPromptSubmit", "SessionStart", "SessionEnd"} {
		expected := "[x] " + hook
		if !strings.Contains(got, expected) {
			t.Errorf("missing %q in output:\n%s", expected, got)
		}
	}

	// Should suggest installing.
	if !strings.Contains(got, "install missing") {
		t.Errorf("should suggest installing missing hooks:\n%s", got)
	}
}
