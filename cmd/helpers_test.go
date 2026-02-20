package cmd_test

import (
	"bytes"
	"os"
	"testing"
	"time"

	"github.com/duboisf/cc-queue/cmd"
	"github.com/duboisf/cc-queue/internal/queue"
	"github.com/spf13/cobra"
)

// nopFullTabber is a mock FullTabber that does nothing.
type nopFullTabber struct{}

func (n *nopFullTabber) EnterFullTab() (func(), error) {
	return func() {}, nil
}

// testOptions creates cmd.Options with captured stdout/stderr and a fixed time.
func testOptions() (cmd.Options, *bytes.Buffer, *bytes.Buffer) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	opts := cmd.Options{
		TimeNow: func() time.Time {
			return time.Date(2026, 2, 18, 14, 30, 0, 0, time.UTC)
		},
		Stdin:      &bytes.Buffer{},
		Stdout:     stdout,
		Stderr:     stderr,
		FullTabber: &nopFullTabber{},
	}
	return opts, stdout, stderr
}

// testOptionsWithStdin creates cmd.Options with the given stdin content.
func testOptionsWithStdin(stdin string) (cmd.Options, *bytes.Buffer, *bytes.Buffer) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	opts := cmd.Options{
		TimeNow: func() time.Time {
			return time.Date(2026, 2, 18, 14, 30, 0, 0, time.UTC)
		},
		Stdin:      bytes.NewBufferString(stdin),
		Stdout:     stdout,
		Stderr:     stderr,
		FullTabber: &nopFullTabber{},
	}
	return opts, stdout, stderr
}

// setupQueueDir overrides XDG_STATE_HOME to isolate queue storage in a temp dir.
// Returns the temp dir path.
func setupQueueDir(t *testing.T) string {
	t.Helper()
	tmp := t.TempDir()
	t.Setenv("XDG_STATE_HOME", tmp)
	return tmp
}

// seedEntry writes a queue entry for testing.
func seedEntry(t *testing.T, sessionID, cwd, event string, pid int) {
	t.Helper()
	seedEntryWithMessage(t, sessionID, cwd, event, pid, "")
}

// seedEntryWithMessage writes a queue entry with a message for testing.
func seedEntryWithMessage(t *testing.T, sessionID, cwd, event string, pid int, message string) {
	t.Helper()
	err := queue.Write(&queue.Entry{
		Timestamp:     time.Now(),
		SessionID:     sessionID,
		KittyWindowID: "42",
		PID:           pid,
		CWD:           cwd,
		Event:         event,
		Message:       message,
	})
	if err != nil {
		t.Fatalf("seedEntry: %v", err)
	}
}

// seedEntryNoWindow writes a queue entry without a KittyWindowID for testing.
func seedEntryNoWindow(t *testing.T, sessionID, cwd, event string, pid int) {
	t.Helper()
	err := queue.Write(&queue.Entry{
		Timestamp: time.Now(),
		SessionID: sessionID,
		PID:       pid,
		CWD:       cwd,
		Event:     event,
	})
	if err != nil {
		t.Fatalf("seedEntryNoWindow: %v", err)
	}
}

// executeCommand runs a cobra command with args and captures output.
func executeCommand(root *cobra.Command, args ...string) (stdout, stderr string, err error) {
	outBuf := new(bytes.Buffer)
	errBuf := new(bytes.Buffer)
	root.SetOut(outBuf)
	root.SetErr(errBuf)
	root.SetArgs(args)
	err = root.Execute()
	return outBuf.String(), errBuf.String(), err
}

// entryCount returns the number of entries in the queue.
func entryCount(t *testing.T) int {
	t.Helper()
	entries, err := queue.List()
	if err != nil {
		t.Fatalf("entryCount: %v", err)
	}
	return len(entries)
}

// mustSetenv sets an env var and registers cleanup.
func mustSetenv(t *testing.T, key, value string) {
	t.Helper()
	old, existed := os.LookupEnv(key)
	os.Setenv(key, value)
	t.Cleanup(func() {
		if existed {
			os.Setenv(key, old)
		} else {
			os.Unsetenv(key)
		}
	})
}
