package kitty

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

// FullTabber switches the kitty tab to a full-screen layout and provides
// a function to restore the previous layout.
type FullTabber interface {
	EnterFullTab() (restore func(), err error)
}

type tab struct {
	IsFocused bool   `json:"is_focused"`
	Layout    string `json:"layout"`
}

type osWindow struct {
	Tabs []tab `json:"tabs"`
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

// LayoutManager implements FullTabber using kitty remote control.
type LayoutManager struct{}

// CurrentLayout returns the layout name of the focused tab.
func (l *LayoutManager) CurrentLayout() (string, error) {
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
func (l *LayoutManager) SetLayout(name string) error {
	out, err := exec.Command("kitty", "@", "goto-layout", name).CombinedOutput()
	if err != nil {
		return fmt.Errorf("kitty @ goto-layout %s: %w\n%s", name, err, strings.TrimSpace(string(out)))
	}
	return nil
}

// EnterFullTab switches to stack layout and returns a function that restores
// the previous layout. If already in stack layout or kitty is unavailable,
// the returned function is a no-op.
func (l *LayoutManager) EnterFullTab() (restore func(), err error) {
	noop := func() {}

	layout, err := l.CurrentLayout()
	if err != nil {
		return noop, err
	}
	if layout == "" || layout == "stack" {
		return noop, nil
	}

	if err := l.SetLayout("stack"); err != nil {
		return noop, err
	}

	return func() {
		l.SetLayout(layout)
	}, nil
}
