package cmd_test

import (
	"strings"
	"testing"

	"github.com/duboisf/cc-queue/cmd"
)

func TestDebugCmd_ShowsOff(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmp)
	setupQueueDir(t)

	opts, stdout, _ := testOptions()
	root := cmd.NewRootCmd(opts)

	_, _, err := executeCommand(root, "debug")
	if err != nil {
		t.Fatalf("debug: %v", err)
	}
	got := stdout.String()
	if !strings.Contains(got, "debug: off") {
		t.Errorf("got %q, want contains %q", got, "debug: off")
	}
}

func TestDebugCmd_TurnOn(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmp)
	setupQueueDir(t)

	opts, stdout, _ := testOptions()
	root := cmd.NewRootCmd(opts)

	_, _, err := executeCommand(root, "debug", "on")
	if err != nil {
		t.Fatalf("debug on: %v", err)
	}
	got := stdout.String()
	if !strings.Contains(got, "debug: on") {
		t.Errorf("got %q, want contains %q", got, "debug: on")
	}

	// Verify it persists by reading status.
	stdout.Reset()
	root2 := cmd.NewRootCmd(opts)
	_, _, err = executeCommand(root2, "debug")
	if err != nil {
		t.Fatalf("debug (check): %v", err)
	}
	got = stdout.String()
	if !strings.Contains(got, "debug: on") {
		t.Errorf("got %q, want contains %q", got, "debug: on")
	}
}

func TestDebugCmd_TurnOff(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmp)
	setupQueueDir(t)

	opts, stdout, _ := testOptions()

	// Turn on first.
	root := cmd.NewRootCmd(opts)
	executeCommand(root, "debug", "on")
	stdout.Reset()

	// Turn off.
	root2 := cmd.NewRootCmd(opts)
	_, _, err := executeCommand(root2, "debug", "off")
	if err != nil {
		t.Fatalf("debug off: %v", err)
	}
	got := stdout.String()
	if !strings.Contains(got, "debug: off") {
		t.Errorf("got %q, want contains %q", got, "debug: off")
	}
}

func TestDebugCmd_InvalidArg(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmp)
	setupQueueDir(t)

	opts, _, _ := testOptions()
	root := cmd.NewRootCmd(opts)

	_, _, err := executeCommand(root, "debug", "maybe")
	if err == nil {
		t.Fatal("expected error for invalid argument")
	}
	if !strings.Contains(err.Error(), "unknown argument") {
		t.Errorf("error = %q, want contains %q", err.Error(), "unknown argument")
	}
}
