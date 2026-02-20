package cmd_test

import (
	"sort"
	"testing"

	"github.com/duboisf/cc-queue/cmd"
)

func TestNewRootCmd_HasExpectedSubcommands(t *testing.T) {
	opts, _, _ := testOptions()
	root := cmd.NewRootCmd(opts)

	expected := []string{
		"push", "pop", "list", "clear", "clean", "first",
		"install", "completion", "version", "_list-fzf", "_preview", "_jump",
	}
	sort.Strings(expected)

	var got []string
	for _, c := range root.Commands() {
		got = append(got, c.Name())
	}
	sort.Strings(got)

	if len(got) != len(expected) {
		t.Fatalf("subcommand count = %d, want %d\ngot:  %v\nwant: %v", len(got), len(expected), got, expected)
	}
	for i := range expected {
		if got[i] != expected[i] {
			t.Errorf("subcommand[%d] = %q, want %q", i, got[i], expected[i])
		}
	}
}

func TestNewRootCmd_UnknownSubcommand(t *testing.T) {
	setupQueueDir(t)
	opts, _, _ := testOptions()
	root := cmd.NewRootCmd(opts)

	_, _, err := executeCommand(root, "nonexistent")
	if err == nil {
		t.Fatal("expected error for unknown subcommand, got nil")
	}
}

func TestDefaultOptions_NonNil(t *testing.T) {
	opts := cmd.DefaultOptions()

	if opts.TimeNow == nil {
		t.Error("TimeNow is nil")
	}
	if opts.Stdin == nil {
		t.Error("Stdin is nil")
	}
	if opts.Stdout == nil {
		t.Error("Stdout is nil")
	}
	if opts.Stderr == nil {
		t.Error("Stderr is nil")
	}
	if opts.CleanStaleWindowsFn == nil {
		t.Error("CleanStaleWindowsFn is nil")
	}
}
