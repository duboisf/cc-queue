package queue

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestInstallHooks_FreshFile(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	if err := InstallHooks(TargetUser); err != nil {
		t.Fatalf("InstallHooks: %v", err)
	}

	path := filepath.Join(tmp, ".claude", "settings.json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}

	var settings map[string]any
	if err := json.Unmarshal(data, &settings); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	hooks, ok := settings["hooks"].(map[string]any)
	if !ok {
		t.Fatal("hooks key missing or not a map")
	}

	// Check Notification hook exists.
	notif, ok := hooks["Notification"].([]any)
	if !ok || len(notif) == 0 {
		t.Fatal("Notification hooks missing")
	}
	if !hasHookCommand(notif, pushCommand) {
		t.Error("push hook not found in Notification")
	}

	// Check UserPromptSubmit hook exists.
	ups, ok := hooks["UserPromptSubmit"].([]any)
	if !ok || len(ups) == 0 {
		t.Fatal("UserPromptSubmit hooks missing")
	}
	if !hasHookCommand(ups, popCommand) {
		t.Error("pop hook not found in UserPromptSubmit")
	}
}

func TestInstallHooks_Idempotent(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	InstallHooks(TargetUser)
	InstallHooks(TargetUser) // second call should not duplicate

	path := filepath.Join(tmp, ".claude", "settings.json")
	data, _ := os.ReadFile(path)

	var settings map[string]any
	json.Unmarshal(data, &settings)
	hooks := settings["hooks"].(map[string]any)

	notif := hooks["Notification"].([]any)
	if len(notif) != 1 {
		t.Errorf("Notification matchers = %d, want 1 (idempotent)", len(notif))
	}
}

func TestInstallHooks_PreservesExisting(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	// Write existing settings with a hook already present.
	dir := filepath.Join(tmp, ".claude")
	os.MkdirAll(dir, 0755)

	existing := map[string]any{
		"permissions": map[string]any{"allow": []any{"Read"}},
		"hooks": map[string]any{
			"UserPromptSubmit": []any{
				map[string]any{
					"matcher": "",
					"hooks": []any{
						map[string]any{
							"type":    "command",
							"command": "existing-hook.sh",
						},
					},
				},
			},
		},
	}
	data, _ := json.MarshalIndent(existing, "", "  ")
	os.WriteFile(filepath.Join(dir, "settings.json"), data, 0644)

	if err := InstallHooks(TargetUser); err != nil {
		t.Fatalf("InstallHooks: %v", err)
	}

	result, _ := os.ReadFile(filepath.Join(dir, "settings.json"))
	var settings map[string]any
	json.Unmarshal(result, &settings)

	// permissions should still exist.
	if _, ok := settings["permissions"]; !ok {
		t.Error("existing permissions key was lost")
	}

	// Existing hook should still be present.
	hooks := settings["hooks"].(map[string]any)
	ups := hooks["UserPromptSubmit"].([]any)
	if !hasHookCommand(ups, "existing-hook.sh") {
		t.Error("existing hook was removed")
	}
	if !hasHookCommand(ups, popCommand) {
		t.Error("pop hook was not added")
	}
}

func TestInstallHooks_NoDuplicateWhenCommandPrefixed(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	// Simulate user having modified the command with an env var prefix.
	dir := filepath.Join(tmp, ".claude")
	os.MkdirAll(dir, 0755)

	existing := map[string]any{
		"hooks": map[string]any{
			"Notification": []any{
				map[string]any{
					"matcher": notificationMatcher,
					"hooks": []any{
						map[string]any{
							"type":    "command",
							"command": "CC_QUEUE_DEBUG=1 cc-queue push",
						},
					},
				},
			},
			"UserPromptSubmit": []any{
				map[string]any{
					"matcher": "",
					"hooks": []any{
						map[string]any{
							"type":    "command",
							"command": "CC_QUEUE_DEBUG=1 cc-queue pop",
						},
					},
				},
			},
		},
	}
	data, _ := json.MarshalIndent(existing, "", "  ")
	os.WriteFile(filepath.Join(dir, "settings.json"), data, 0644)

	if err := InstallHooks(TargetUser); err != nil {
		t.Fatalf("InstallHooks: %v", err)
	}

	result, _ := os.ReadFile(filepath.Join(dir, "settings.json"))
	var settings map[string]any
	json.Unmarshal(result, &settings)
	hooks := settings["hooks"].(map[string]any)

	notif := hooks["Notification"].([]any)
	if len(notif) != 1 {
		t.Errorf("Notification matchers = %d, want 1 (should not duplicate prefixed command)", len(notif))
	}

	ups := hooks["UserPromptSubmit"].([]any)
	// Should still be 1 matcher group with 1 hook (no duplicate pop added).
	if len(ups) != 1 {
		t.Errorf("UserPromptSubmit matchers = %d, want 1", len(ups))
	}
	matcher := ups[0].(map[string]any)
	hooksList := matcher["hooks"].([]any)
	if len(hooksList) != 1 {
		t.Errorf("UserPromptSubmit hooks = %d, want 1 (should not duplicate prefixed command)", len(hooksList))
	}
}

func TestInstallHooks_ProjectTarget(t *testing.T) {
	tmp := t.TempDir()
	// Change to temp dir to test project-level install.
	oldWd, _ := os.Getwd()
	os.Chdir(tmp)
	defer os.Chdir(oldWd)

	if err := InstallHooks(TargetProject); err != nil {
		t.Fatalf("InstallHooks(project): %v", err)
	}

	path := filepath.Join(tmp, ".claude", "settings.json")
	if _, err := os.Stat(path); err != nil {
		t.Errorf("project settings file not created: %v", err)
	}
}
