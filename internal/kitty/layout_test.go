package kitty

import (
	"testing"
)

func TestFocusedLayout(t *testing.T) {
	input := `[{"id":1,"tabs":[{"id":1,"is_focused":false,"layout":"splits"},{"id":2,"is_focused":true,"layout":"tall"}]}]`

	var windows []osWindow
	if err := parseJSON(input, &windows); err != nil {
		t.Fatalf("parse: %v", err)
	}

	got := focusedLayout(windows)
	if got != "tall" {
		t.Errorf("layout = %q, want %q", got, "tall")
	}
}

func TestFocusedLayout_NoFocused(t *testing.T) {
	input := `[{"id":1,"tabs":[{"id":1,"is_focused":false,"layout":"splits"}]}]`

	var windows []osWindow
	if err := parseJSON(input, &windows); err != nil {
		t.Fatalf("parse: %v", err)
	}

	got := focusedLayout(windows)
	if got != "" {
		t.Errorf("layout = %q, want empty", got)
	}
}

func TestFocusedLayout_MultipleOSWindows(t *testing.T) {
	input := `[
		{"id":1,"tabs":[{"id":1,"is_focused":false,"layout":"splits"}]},
		{"id":2,"tabs":[{"id":2,"is_focused":true,"layout":"fat"}]}
	]`

	var windows []osWindow
	if err := parseJSON(input, &windows); err != nil {
		t.Fatalf("parse: %v", err)
	}

	got := focusedLayout(windows)
	if got != "fat" {
		t.Errorf("layout = %q, want %q", got, "fat")
	}
}
