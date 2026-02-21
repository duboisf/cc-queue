package queue

import (
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

func TestShortenPath(t *testing.T) {
	got := ShortenPath("/nonexistent/absolute/path")
	if got != "/nonexistent/absolute/path" {
		t.Errorf("non-home path changed: %q", got)
	}
}
