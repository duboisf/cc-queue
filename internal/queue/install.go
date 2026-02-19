package queue

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const (
	pushCommand         = "cc-queue push"
	popCommand          = "cc-queue pop"
	notificationMatcher = "permission_prompt|idle_prompt|elicitation_dialog"
)

// SettingsTarget represents where to install hooks.
type SettingsTarget int

const (
	TargetUser    SettingsTarget = iota // ~/.claude/settings.json
	TargetProject                       // .claude/settings.json (cwd)
)

// SettingsPath returns the file path for the given target.
func SettingsPath(target SettingsTarget) (string, error) {
	switch target {
	case TargetUser:
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(home, ".claude", "settings.json"), nil
	case TargetProject:
		cwd, err := os.Getwd()
		if err != nil {
			return "", err
		}
		return filepath.Join(cwd, ".claude", "settings.json"), nil
	default:
		return "", fmt.Errorf("unknown target: %d", target)
	}
}

// InstallHooks adds cc-queue push/pop hooks to a Claude Code settings file.
// It merges with existing hooks without clobbering them.
func InstallHooks(target SettingsTarget) error {
	path, err := SettingsPath(target)
	if err != nil {
		return err
	}

	settings, err := readSettings(path)
	if err != nil {
		return err
	}

	hooks := getOrCreateMap(settings, "hooks")

	addNotificationHook(hooks)
	addUserPromptSubmitHook(hooks)

	settings["hooks"] = hooks
	return writeSettings(path, settings)
}

func readSettings(path string) (map[string]any, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return make(map[string]any), nil
		}
		return nil, fmt.Errorf("reading %s: %w", path, err)
	}
	var settings map[string]any
	if err := json.Unmarshal(data, &settings); err != nil {
		return nil, fmt.Errorf("parsing %s: %w", path, err)
	}
	return settings, nil
}

func writeSettings(path string, settings map[string]any) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(path, data, 0644)
}

func getOrCreateMap(m map[string]any, key string) map[string]any {
	if v, ok := m[key].(map[string]any); ok {
		return v
	}
	v := make(map[string]any)
	m[key] = v
	return v
}

// addNotificationHook adds the push hook for Notification events.
func addNotificationHook(hooks map[string]any) {
	eventKey := "Notification"
	matchers := getOrCreateArray(hooks, eventKey)

	if hasHookCommand(matchers, pushCommand) {
		return // already installed
	}

	entry := map[string]any{
		"matcher": notificationMatcher,
		"hooks": []any{
			map[string]any{
				"type":    "command",
				"command": pushCommand,
			},
		},
	}
	hooks[eventKey] = append(matchers, entry)
}

// addUserPromptSubmitHook adds the pop hook for UserPromptSubmit events.
func addUserPromptSubmitHook(hooks map[string]any) {
	eventKey := "UserPromptSubmit"
	matchers := getOrCreateArray(hooks, eventKey)

	if hasHookCommand(matchers, popCommand) {
		return // already installed
	}

	// Try to append to existing empty-matcher entry, otherwise create new one.
	for _, m := range matchers {
		matcher, ok := m.(map[string]any)
		if !ok {
			continue
		}
		matcherStr, _ := matcher["matcher"].(string)
		if matcherStr != "" {
			continue
		}
		// Found the empty matcher — append our hook.
		hooksList, _ := matcher["hooks"].([]any)
		matcher["hooks"] = append(hooksList, map[string]any{
			"type":    "command",
			"command": popCommand,
		})
		return
	}

	// No empty matcher exists — create one.
	entry := map[string]any{
		"matcher": "",
		"hooks": []any{
			map[string]any{
				"type":    "command",
				"command": popCommand,
			},
		},
	}
	hooks[eventKey] = append(matchers, entry)
}

func getOrCreateArray(m map[string]any, key string) []any {
	if v, ok := m[key].([]any); ok {
		return v
	}
	return nil
}

// KittyShortcuts holds the keyboard shortcuts to install in kitty.conf.
type KittyShortcuts struct {
	Picker string // shortcut for the fzf picker overlay (e.g. "kitty_mod+shift+q")
	First  string // shortcut for jump-to-first (e.g. "kitty_mod+shift+u")
}

// KittyInstallResult describes what was written during kitty config installation.
type KittyInstallResult struct {
	ConfPath string // path to cc-queue.conf
	Content  string // content of cc-queue.conf
	MainConf string // path to kitty.conf
	Included bool   // whether an include line was added to kitty.conf
}

// kittyBlockRe matches legacy cc-queue blocks written directly into kitty.conf.
var kittyBlockRe = regexp.MustCompile(`(?m)\n?# cc-queue keyboard shortcuts\n(?:map [^\n]+ cc-queue[^\n]*\n)*`)

// BuildKittyConfig returns the content for cc-queue.conf without writing anything.
func BuildKittyConfig(shortcuts KittyShortcuts) string {
	var b strings.Builder
	b.WriteString("# cc-queue configuration for kitty\n\n")
	b.WriteString("# Enable remote control for cross-window jumping\n")
	b.WriteString("allow_remote_control socket-only\n")
	b.WriteString("listen_on unix:/tmp/kitty-{kitty_pid}\n")
	if shortcuts.Picker != "" || shortcuts.First != "" {
		b.WriteString("\n# Keyboard shortcuts\n")
		if shortcuts.Picker != "" {
			fmt.Fprintf(&b, "map %s launch --type=overlay --title cc-queue cc-queue\n", shortcuts.Picker)
		}
		if shortcuts.First != "" {
			fmt.Fprintf(&b, "map %s launch --type=overlay --title cc-queue cc-queue first\n", shortcuts.First)
		}
	}
	return b.String()
}

// InstallKittyConfig creates cc-queue.conf in the kitty config directory and
// adds an include directive to kitty.conf. Returns nil if the kitty config
// directory doesn't exist.
func InstallKittyConfig(shortcuts KittyShortcuts) (*KittyInstallResult, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	kittyDir := filepath.Join(home, ".config", "kitty")
	if _, err := os.Stat(kittyDir); os.IsNotExist(err) {
		return nil, nil // kitty config dir doesn't exist, skip
	}

	confPath := filepath.Join(kittyDir, "cc-queue.conf")
	mainConf := filepath.Join(kittyDir, "kitty.conf")
	content := BuildKittyConfig(shortcuts)

	if err := os.WriteFile(confPath, []byte(content), 0644); err != nil {
		return nil, fmt.Errorf("writing %s: %w", confPath, err)
	}

	// Add include directive to kitty.conf if not already present.
	mainContent, err := os.ReadFile(mainConf)
	if err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("reading %s: %w", mainConf, err)
	}

	included := false
	includeLine := "include cc-queue.conf"
	mainStr := string(mainContent)

	if !strings.Contains(mainStr, includeLine) {
		// Clean up any legacy cc-queue blocks from previous installs.
		cleaned := kittyBlockRe.ReplaceAllString(mainStr, "")
		cleaned = strings.TrimRight(cleaned, "\n")
		newMain := cleaned + "\n\n" + includeLine + "\n"
		if err := os.WriteFile(mainConf, []byte(newMain), 0644); err != nil {
			return nil, fmt.Errorf("writing %s: %w", mainConf, err)
		}
		included = true
	}

	return &KittyInstallResult{
		ConfPath: confPath,
		Content:  content,
		MainConf: mainConf,
		Included: included,
	}, nil
}

// hasHookCommand checks if any matcher entry already contains the given command.
func hasHookCommand(matchers []any, command string) bool {
	for _, m := range matchers {
		matcher, ok := m.(map[string]any)
		if !ok {
			continue
		}
		hooksList, ok := matcher["hooks"].([]any)
		if !ok {
			continue
		}
		for _, h := range hooksList {
			hook, ok := h.(map[string]any)
			if !ok {
				continue
			}
			if cmd, ok := hook["command"].(string); ok && cmd == command {
				return true
			}
		}
	}
	return false
}
