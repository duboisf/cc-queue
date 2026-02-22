package queue

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

func TestFormatAge(t *testing.T) {
	tests := []struct {
		age  time.Duration
		want string
	}{
		{5 * time.Second, "5s"},
		{45 * time.Second, "45s"},
		{3 * time.Minute, "3m"},
		{59 * time.Minute, "59m"},
		{2 * time.Hour, "2h"},
		{25 * time.Hour, "25h"},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			ts := time.Now().Add(-tt.age)
			got := FormatAge(ts)
			if got != tt.want {
				t.Errorf("FormatAge(%v ago) = %q, want %q", tt.age, got, tt.want)
			}
		})
	}
}

func TestEventLabel(t *testing.T) {
	tests := []struct {
		event string
		want  string
	}{
		{"permission_prompt", "PERM"},
		{"elicitation_dialog", "ASK"},
		{"idle_prompt", "IDLE"},
		{"working", "WORK"},
		{"SessionStart", "START"},
		{"SessionEnd", "END"},
		{"unknown_thing", "UNKNOWN_THING"},
	}
	for _, tt := range tests {
		t.Run(tt.event, func(t *testing.T) {
			got := EventLabel(tt.event)
			if got != tt.want {
				t.Errorf("EventLabel(%q) = %q, want %q", tt.event, got, tt.want)
			}
		})
	}
}

func TestNeedsAttention(t *testing.T) {
	tests := []struct {
		event string
		want  bool
	}{
		{"permission_prompt", true},
		{"elicitation_dialog", true},
		{"idle_prompt", true},
		{"working", false},
		{"SessionStart", false},
		{"SessionEnd", false},
		{"", false},
	}
	for _, tt := range tests {
		t.Run(tt.event, func(t *testing.T) {
			got := NeedsAttention(tt.event)
			if got != tt.want {
				t.Errorf("NeedsAttention(%q) = %v, want %v", tt.event, got, tt.want)
			}
		})
	}
}

func TestGitBranch_InGitRepo(t *testing.T) {
	dir := t.TempDir()
	// Init a git repo with a commit so HEAD exists.
	for _, args := range [][]string{
		{"init"},
		{"commit", "--allow-empty", "-m", "init"},
	} {
		cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
		cmd.Env = append(os.Environ(), "GIT_CONFIG_GLOBAL=/dev/null", "GIT_AUTHOR_NAME=test", "GIT_AUTHOR_EMAIL=test@test", "GIT_COMMITTER_NAME=test", "GIT_COMMITTER_EMAIL=test@test")
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	got := GitBranch(dir)
	// Default branch is usually "main" or "master" depending on git config.
	if got == "" {
		t.Error("GitBranch returned empty for a git repo")
	}
}

func TestGitBranch_NotGitRepo(t *testing.T) {
	dir := t.TempDir()
	got := GitBranch(dir)
	if got != "" {
		t.Errorf("GitBranch returned %q for non-git directory", got)
	}
}

func TestGitBranch_NonexistentDir(t *testing.T) {
	got := GitBranch(filepath.Join(t.TempDir(), "nonexistent"))
	if got != "" {
		t.Errorf("GitBranch returned %q for nonexistent directory", got)
	}
}

func TestShortenPath(t *testing.T) {
	got := ShortenPath("/nonexistent/absolute/path")
	if got != "/nonexistent/absolute/path" {
		t.Errorf("non-home path changed: %q", got)
	}
}
