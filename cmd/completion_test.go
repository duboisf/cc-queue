package cmd_test

import (
	"strings"
	"testing"

	"github.com/duboisf/cc-queue/cmd"
)

func TestCompletion_Zsh(t *testing.T) {
	t.Parallel()
	opts, _, _ := testOptions()
	root := cmd.NewRootCmd(opts)

	stdout, _, err := executeCommand(root, "completion", "zsh")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stdout == "" {
		t.Fatal("expected non-empty output")
	}
	if !strings.Contains(stdout, "#compdef") {
		t.Errorf("expected zsh completion to contain #compdef, got:\n%s", stdout[:min(len(stdout), 200)])
	}
}

func TestCompletion_Bash(t *testing.T) {
	t.Parallel()
	opts, _, _ := testOptions()
	root := cmd.NewRootCmd(opts)

	stdout, _, err := executeCommand(root, "completion", "bash")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stdout == "" {
		t.Fatal("expected non-empty output")
	}
}

func TestCompletion_NoArgs(t *testing.T) {
	t.Parallel()
	opts, _, _ := testOptions()
	root := cmd.NewRootCmd(opts)

	_, _, err := executeCommand(root, "completion")
	if err == nil {
		t.Fatal("expected error when no args provided")
	}
}

func TestCompletion_UnsupportedShell(t *testing.T) {
	t.Parallel()
	opts, _, _ := testOptions()
	root := cmd.NewRootCmd(opts)

	_, _, err := executeCommand(root, "completion", "fish")
	if err == nil {
		t.Fatal("expected error for unsupported shell")
	}
	if !strings.Contains(err.Error(), "unsupported shell") {
		t.Errorf("expected error to contain 'unsupported shell', got: %v", err)
	}
}
