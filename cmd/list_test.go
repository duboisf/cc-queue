package cmd_test

import (
	"errors"
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

func TestListFzf_Empty(t *testing.T) {
	setupQueueDir(t)
	opts, _, _ := testOptions()
	root := cmd.NewRootCmd(opts)

	stdout, _, err := executeCommand(root, "_list-fzf")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stdout != "" {
		t.Errorf("output = %q, want empty", stdout)
	}
}

func TestListFzf_WithEntries(t *testing.T) {
	setupQueueDir(t)
	opts, _, _ := testOptions()

	seedEntry(t, "sess-fzf-1", "/home/user/proj-a", "permission_prompt", 1001)
	seedEntry(t, "sess-fzf-2", "/home/user/proj-b", "idle_prompt", 1002)

	root := cmd.NewRootCmd(opts)
	stdout, _, err := executeCommand(root, "_list-fzf")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Each line should be: session_id\tage label  path
	if !strings.Contains(stdout, "sess-fzf-1\t") {
		t.Errorf("output missing sess-fzf-1 tab-delimited:\n%s", stdout)
	}
	if !strings.Contains(stdout, "sess-fzf-2\t") {
		t.Errorf("output missing sess-fzf-2 tab-delimited:\n%s", stdout)
	}
}

func TestPreview_ExistingEntry(t *testing.T) {
	setupQueueDir(t)
	opts, _, _ := testOptions()

	seedEntryWithMessage(t, "sess-prev", "/home/user/proj", "permission_prompt", 1001, "Allow read access?")

	root := cmd.NewRootCmd(opts)
	stdout, _, err := executeCommand(root, "_preview", "sess-prev")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(stdout, "PERM") {
		t.Errorf("output missing PERM label:\n%s", stdout)
	}
	if !strings.Contains(stdout, "Allow read access?") {
		t.Errorf("output missing message:\n%s", stdout)
	}
}

func TestPreview_MissingEntry(t *testing.T) {
	setupQueueDir(t)
	opts, _, _ := testOptions()

	root := cmd.NewRootCmd(opts)
	stdout, _, err := executeCommand(root, "_preview", "nonexistent")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stdout != "" {
		t.Errorf("output = %q, want empty", stdout)
	}
}

// errFullTabber is a mock FullTabber that returns an error.
type errFullTabber struct{}

func (e *errFullTabber) EnterFullTab() (func(), error) {
	return func() {}, errors.New("kitty not available")
}

// trackingFullTabber records calls for verification.
type trackingFullTabber struct {
	entered  bool
	restored bool
}

func (t *trackingFullTabber) EnterFullTab() (func(), error) {
	t.entered = true
	return func() { t.restored = true }, nil
}

func TestFirst_FullTab_ErrorReturned(t *testing.T) {
	setupQueueDir(t)
	opts, _, _ := testOptions()
	opts.FullTabber = &errFullTabber{}

	seedEntry(t, "sess-ft", "/home/user/proj", "permission_prompt", 1001)

	root := cmd.NewRootCmd(opts)
	_, _, err := executeCommand(root, "first", "--full-tab")
	if err == nil {
		t.Fatal("expected error from FullTabber, got nil")
	}
	if !strings.Contains(err.Error(), "kitty not available") {
		t.Errorf("error = %q, want to contain 'kitty not available'", err.Error())
	}
}

func TestFirst_FullTab_Called(t *testing.T) {
	setupQueueDir(t)
	opts, _, _ := testOptions()
	ft := &trackingFullTabber{}
	opts.FullTabber = ft

	root := cmd.NewRootCmd(opts)
	_, _, _ = executeCommand(root, "first", "--full-tab")

	if !ft.entered {
		t.Error("EnterFullTab was not called")
	}
	if !ft.restored {
		t.Error("restore was not called")
	}
}

func TestRootCmd_FullTab_Error(t *testing.T) {
	setupQueueDir(t)
	opts, _, _ := testOptions()
	opts.FullTabber = &errFullTabber{}

	root := cmd.NewRootCmd(opts)
	_, _, err := executeCommand(root, "--full-tab")
	if err == nil {
		t.Fatal("expected error from FullTabber, got nil")
	}
}

func TestFirst_NoFullTab_NotCalled(t *testing.T) {
	setupQueueDir(t)
	opts, _, _ := testOptions()
	ft := &trackingFullTabber{}
	opts.FullTabber = ft

	root := cmd.NewRootCmd(opts)
	_, _, _ = executeCommand(root, "first")

	if ft.entered {
		t.Error("EnterFullTab should not be called without --full-tab")
	}
}
