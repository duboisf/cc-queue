package cmd_test

import (
	"strings"
	"testing"

	"github.com/duboisf/cc-queue/cmd"
)

func TestList_Empty(t *testing.T) {
	setupQueueDir(t)
	opts, stdout, _ := testOptions()
	root := cmd.NewRootCmd(opts)

	_, _, err := executeCommand(root, "list")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := stdout.String()
	want := "No pending items\n"
	if got != want {
		t.Errorf("output = %q, want %q", got, want)
	}
}

func TestList_WithEntries(t *testing.T) {
	setupQueueDir(t)
	opts, stdout, _ := testOptions()

	seedEntry(t, "sess-1", "/home/user/project-a", "permission_prompt", 1001)
	seedEntry(t, "sess-2", "/home/user/project-b", "idle_prompt", 1002)

	root := cmd.NewRootCmd(opts)
	_, _, err := executeCommand(root, "list")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := stdout.String()

	if !strings.Contains(got, "PERM") {
		t.Errorf("output missing PERM label:\n%s", got)
	}
	if !strings.Contains(got, "IDLE") {
		t.Errorf("output missing IDLE label:\n%s", got)
	}
	if !strings.Contains(got, "project-a") {
		t.Errorf("output missing cwd project-a:\n%s", got)
	}
	if !strings.Contains(got, "project-b") {
		t.Errorf("output missing cwd project-b:\n%s", got)
	}
}

func TestList_OutputFormat(t *testing.T) {
	setupQueueDir(t)
	opts, stdout, _ := testOptions()

	seedEntry(t, "sess-fmt", "/tmp/myproject", "permission_prompt", 2001)

	root := cmd.NewRootCmd(opts)
	_, _, err := executeCommand(root, "list")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := stdout.String()
	lines := strings.Split(strings.TrimRight(got, "\n"), "\n")
	if len(lines) != 1 {
		t.Fatalf("expected 1 line, got %d: %q", len(lines), got)
	}

	line := lines[0]
	// Format is "%-5s %-4s  %s\n" -> age(5) space label(4) two-spaces path
	// The line should contain the PERM label and the path
	if !strings.Contains(line, "PERM") {
		t.Errorf("line missing PERM label: %q", line)
	}
	if !strings.Contains(line, "/tmp/myproject") {
		t.Errorf("line missing path: %q", line)
	}

	// Verify column alignment: label starts after age field (5 chars + space)
	// Age is left-aligned in 5 chars, then label in 4 chars, then two spaces, then path
	parts := strings.Fields(line)
	if len(parts) < 3 {
		t.Fatalf("expected at least 3 fields, got %d: %q", len(parts), line)
	}
	if parts[1] != "PERM" {
		t.Errorf("second field = %q, want %q", parts[1], "PERM")
	}
}
