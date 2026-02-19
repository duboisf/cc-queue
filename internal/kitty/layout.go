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

// Runner abstracts kitty remote control commands for testability.
type Runner interface {
	Ls() ([]byte, error)
	GotoLayout(name string) error
}

// ExecRunner implements Runner by shelling out to kitty.
type ExecRunner struct{}

func (e *ExecRunner) Ls() ([]byte, error) {
	return exec.Command("kitty", "@", "ls").Output()
}

func (e *ExecRunner) GotoLayout(name string) error {
	out, err := exec.Command("kitty", "@", "goto-layout", name).CombinedOutput()
	if err != nil {
		return fmt.Errorf("kitty @ goto-layout %s: %w\n%s", name, err, strings.TrimSpace(string(out)))
	}
	return nil
}

type tab struct {
	IsFocused bool   `json:"is_focused"`
	Layout    string `json:"layout"`
}

type osWindow struct {
	Tabs []tab `json:"tabs"`
}

// LayoutManager implements FullTabber using a Runner for kitty remote control.
type LayoutManager struct {
	Runner Runner
}

// NewLayoutManager returns a LayoutManager that shells out to kitty.
func NewLayoutManager() *LayoutManager {
	return &LayoutManager{Runner: &ExecRunner{}}
}

// CurrentLayout returns the layout name of the focused tab.
func (l *LayoutManager) CurrentLayout() (string, error) {
	out, err := l.Runner.Ls()
	if err != nil {
		return "", fmt.Errorf("kitty @ ls: %w", err)
	}
	var windows []osWindow
	if err := json.Unmarshal(out, &windows); err != nil {
		return "", fmt.Errorf("parsing kitty @ ls: %w", err)
	}
	for _, w := range windows {
		for _, t := range w.Tabs {
			if t.IsFocused {
				return t.Layout, nil
			}
		}
	}
	return "", nil
}

// SetLayout changes the layout of the focused tab.
func (l *LayoutManager) SetLayout(name string) error {
	return l.Runner.GotoLayout(name)
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
