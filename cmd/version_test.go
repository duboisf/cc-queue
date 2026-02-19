package cmd_test

import (
	"strings"
	"testing"

	"github.com/duboisf/cc-queue/cmd"
)

func TestVersion_Output(t *testing.T) {
	t.Parallel()
	opts, _, _ := testOptions()
	root := cmd.NewRootCmd(opts)

	stdout, _, err := executeCommand(root, "version")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.HasPrefix(stdout, "cc-queue ") {
		t.Errorf("expected output to start with 'cc-queue ', got: %q", stdout)
	}
}

func TestVersion_ContainsVersion(t *testing.T) {
	t.Parallel()
	opts, _, _ := testOptions()
	root := cmd.NewRootCmd(opts)

	stdout, _, err := executeCommand(root, "version")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stdout == "" {
		t.Fatal("expected non-empty output")
	}
	if !strings.HasSuffix(stdout, "\n") {
		t.Errorf("expected output to end with newline, got: %q", stdout)
	}
}
