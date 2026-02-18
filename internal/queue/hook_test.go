package queue

import (
	"strings"
	"testing"
)

func TestParseHookInput(t *testing.T) {
	input := `{
		"session_id": "abc-123",
		"cwd": "/home/fred/project",
		"hook_event_name": "Notification",
		"notification_type": "permission_prompt",
		"transcript_path": "/tmp/transcript.jsonl"
	}`

	hi, err := ParseHookInput(strings.NewReader(input))
	if err != nil {
		t.Fatalf("ParseHookInput: %v", err)
	}

	if hi.SessionID != "abc-123" {
		t.Errorf("SessionID = %q, want %q", hi.SessionID, "abc-123")
	}
	if hi.CWD != "/home/fred/project" {
		t.Errorf("CWD = %q, want %q", hi.CWD, "/home/fred/project")
	}
	if hi.HookEventName != "Notification" {
		t.Errorf("HookEventName = %q, want %q", hi.HookEventName, "Notification")
	}
}

func TestEventType_NotificationType(t *testing.T) {
	input := `{
		"session_id": "s1",
		"hook_event_name": "Notification",
		"notification_type": "permission_prompt"
	}`
	hi, _ := ParseHookInput(strings.NewReader(input))
	if got := hi.EventType(); got != "permission_prompt" {
		t.Errorf("EventType() = %q, want %q", got, "permission_prompt")
	}
}

func TestEventType_FallbackToHookEventName(t *testing.T) {
	input := `{
		"session_id": "s2",
		"hook_event_name": "Stop"
	}`
	hi, _ := ParseHookInput(strings.NewReader(input))
	if got := hi.EventType(); got != "Stop" {
		t.Errorf("EventType() = %q, want %q", got, "Stop")
	}
}

func TestEventType_TypeField(t *testing.T) {
	input := `{
		"session_id": "s3",
		"hook_event_name": "Notification",
		"type": "idle_prompt"
	}`
	hi, _ := ParseHookInput(strings.NewReader(input))
	if got := hi.EventType(); got != "idle_prompt" {
		t.Errorf("EventType() = %q, want %q", got, "idle_prompt")
	}
}
