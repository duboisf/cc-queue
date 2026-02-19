package kitty

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

type tab struct {
	IsFocused bool   `json:"is_focused"`
	Layout    string `json:"layout"`
}

type osWindow struct {
	Tabs []tab `json:"tabs"`
}

func parseJSON(data string, v any) error {
	return json.Unmarshal([]byte(data), v)
}

func focusedLayout(windows []osWindow) string {
	for _, w := range windows {
		for _, t := range w.Tabs {
			if t.IsFocused {
				return t.Layout
			}
		}
	}
	return ""
}

// CurrentLayout returns the layout name of the focused tab.
func CurrentLayout() (string, error) {
	out, err := exec.Command("kitty", "@", "ls").Output()
	if err != nil {
		return "", fmt.Errorf("kitty @ ls: %w", err)
	}
	var windows []osWindow
	if err := json.Unmarshal(out, &windows); err != nil {
		return "", fmt.Errorf("parsing kitty @ ls: %w", err)
	}
	return focusedLayout(windows), nil
}

// SetLayout changes the layout of the focused tab.
func SetLayout(name string) error {
	out, err := exec.Command("kitty", "@", "goto-layout", name).CombinedOutput()
	if err != nil {
		return fmt.Errorf("kitty @ goto-layout %s: %w\n%s", name, err, strings.TrimSpace(string(out)))
	}
	return nil
}

// EnterFullTab switches to stack layout and returns a function that restores
// the previous layout. If already in stack layout or kitty is unavailable,
// the returned function is a no-op.
func EnterFullTab() (restore func(), err error) {
	noop := func() {}

	layout, err := CurrentLayout()
	if err != nil {
		return noop, err
	}
	if layout == "" || layout == "stack" {
		return noop, nil
	}

	if err := SetLayout("stack"); err != nil {
		return noop, err
	}

	return func() {
		SetLayout(layout)
	}, nil
}
