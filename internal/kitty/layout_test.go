package kitty_test

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/duboisf/cc-queue/internal/kitty"
)

type mockRunner struct {
	lsOutput    []byte
	lsErr       error
	gotoLayout  string
	gotoErr     error
	gotoCalls   []string
}

func (m *mockRunner) Ls() ([]byte, error) {
	return m.lsOutput, m.lsErr
}

func (m *mockRunner) GotoLayout(name string) error {
	m.gotoCalls = append(m.gotoCalls, name)
	m.gotoLayout = name
	return m.gotoErr
}

func kittyLsJSON(tabs ...struct{ layout string; focused bool }) []byte {
	type jsonTab struct {
		IsFocused bool   `json:"is_focused"`
		Layout    string `json:"layout"`
	}
	type jsonWin struct {
		Tabs []jsonTab `json:"tabs"`
	}
	var t []jsonTab
	for _, tab := range tabs {
		t = append(t, jsonTab{IsFocused: tab.focused, Layout: tab.layout})
	}
	data, _ := json.Marshal([]jsonWin{{Tabs: t}})
	return data
}

func TestCurrentLayout_ReturnsFocusedLayout(t *testing.T) {
	r := &mockRunner{
		lsOutput: kittyLsJSON(
			struct{ layout string; focused bool }{"splits", false},
			struct{ layout string; focused bool }{"tall", true},
		),
	}
	lm := &kitty.LayoutManager{Runner: r}

	got, err := lm.CurrentLayout()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "tall" {
		t.Errorf("layout = %q, want %q", got, "tall")
	}
}

func TestCurrentLayout_NoFocusedTab(t *testing.T) {
	r := &mockRunner{
		lsOutput: kittyLsJSON(
			struct{ layout string; focused bool }{"splits", false},
		),
	}
	lm := &kitty.LayoutManager{Runner: r}

	got, err := lm.CurrentLayout()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "" {
		t.Errorf("layout = %q, want empty", got)
	}
}

func TestCurrentLayout_LsError(t *testing.T) {
	r := &mockRunner{lsErr: errors.New("no kitty")}
	lm := &kitty.LayoutManager{Runner: r}

	_, err := lm.CurrentLayout()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestCurrentLayout_InvalidJSON(t *testing.T) {
	r := &mockRunner{lsOutput: []byte("not json")}
	lm := &kitty.LayoutManager{Runner: r}

	_, err := lm.CurrentLayout()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestSetLayout(t *testing.T) {
	r := &mockRunner{}
	lm := &kitty.LayoutManager{Runner: r}

	if err := lm.SetLayout("tall"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.gotoLayout != "tall" {
		t.Errorf("gotoLayout = %q, want %q", r.gotoLayout, "tall")
	}
}

func TestSetLayout_Error(t *testing.T) {
	r := &mockRunner{gotoErr: errors.New("fail")}
	lm := &kitty.LayoutManager{Runner: r}

	if err := lm.SetLayout("tall"); err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestEnterFullTab_SwitchesAndRestores(t *testing.T) {
	r := &mockRunner{
		lsOutput: kittyLsJSON(
			struct{ layout string; focused bool }{"splits", true},
		),
	}
	lm := &kitty.LayoutManager{Runner: r}

	restore, err := lm.EnterFullTab()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(r.gotoCalls) != 1 || r.gotoCalls[0] != "stack" {
		t.Errorf("gotoCalls = %v, want [stack]", r.gotoCalls)
	}

	restore()

	if len(r.gotoCalls) != 2 || r.gotoCalls[1] != "splits" {
		t.Errorf("gotoCalls = %v, want [stack splits]", r.gotoCalls)
	}
}

func TestEnterFullTab_AlreadyStack(t *testing.T) {
	r := &mockRunner{
		lsOutput: kittyLsJSON(
			struct{ layout string; focused bool }{"stack", true},
		),
	}
	lm := &kitty.LayoutManager{Runner: r}

	restore, err := lm.EnterFullTab()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(r.gotoCalls) != 0 {
		t.Errorf("expected no goto calls, got %v", r.gotoCalls)
	}

	// Restore should be safe to call (noop).
	restore()

	if len(r.gotoCalls) != 0 {
		t.Errorf("expected no goto calls after restore, got %v", r.gotoCalls)
	}
}

func TestEnterFullTab_EmptyLayout(t *testing.T) {
	r := &mockRunner{lsOutput: []byte("[]")}
	lm := &kitty.LayoutManager{Runner: r}

	restore, err := lm.EnterFullTab()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	restore()

	if len(r.gotoCalls) != 0 {
		t.Errorf("expected no goto calls, got %v", r.gotoCalls)
	}
}

func TestEnterFullTab_LsError(t *testing.T) {
	r := &mockRunner{lsErr: errors.New("no kitty")}
	lm := &kitty.LayoutManager{Runner: r}

	restore, err := lm.EnterFullTab()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	// Restore should be safe to call.
	restore()
}

func TestEnterFullTab_GotoError(t *testing.T) {
	r := &mockRunner{
		lsOutput: kittyLsJSON(
			struct{ layout string; focused bool }{"splits", true},
		),
		gotoErr: errors.New("goto fail"),
	}
	lm := &kitty.LayoutManager{Runner: r}

	restore, err := lm.EnterFullTab()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	// Restore should be safe to call.
	restore()
}

func TestNewLayoutManager(t *testing.T) {
	lm := kitty.NewLayoutManager()
	if lm == nil {
		t.Fatal("NewLayoutManager returned nil")
	}
	if lm.Runner == nil {
		t.Fatal("Runner is nil")
	}
}
