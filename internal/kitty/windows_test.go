package kitty_test

import (
	"testing"

	"github.com/duboisf/cc-queue/internal/kitty"
)

func TestParseWindowIDs(t *testing.T) {
	// Realistic kitty @ ls output with 2 OS windows, multiple tabs and windows.
	lsOutput := []byte(`[
		{
			"id": 1,
			"tabs": [
				{
					"id": 1,
					"windows": [
						{"id": 10},
						{"id": 11}
					]
				},
				{
					"id": 2,
					"windows": [
						{"id": 20}
					]
				}
			]
		},
		{
			"id": 2,
			"tabs": [
				{
					"id": 3,
					"windows": [
						{"id": 30},
						{"id": 31}
					]
				}
			]
		}
	]`)

	ids, err := kitty.ParseWindowIDs(lsOutput)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := map[string]bool{
		"10": true, "11": true,
		"20": true,
		"30": true, "31": true,
	}
	if len(ids) != len(want) {
		t.Fatalf("got %d IDs, want %d", len(ids), len(want))
	}
	for id := range want {
		if !ids[id] {
			t.Errorf("missing window ID %q", id)
		}
	}
}

func TestParseWindowIDs_Empty(t *testing.T) {
	ids, err := kitty.ParseWindowIDs([]byte(`[]`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ids) != 0 {
		t.Errorf("got %d IDs, want 0", len(ids))
	}
}

func TestParseWindowIDs_InvalidJSON(t *testing.T) {
	_, err := kitty.ParseWindowIDs([]byte(`not json`))
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestParseWindowIDs_NoWindows(t *testing.T) {
	lsOutput := []byte(`[{"id": 1, "tabs": [{"id": 1, "windows": []}]}]`)
	ids, err := kitty.ParseWindowIDs(lsOutput)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ids) != 0 {
		t.Errorf("got %d IDs, want 0", len(ids))
	}
}
