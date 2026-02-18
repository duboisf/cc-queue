package queue

import (
	"encoding/json"
	"io"
)

// HookInput captures common fields from Claude Code hook JSON.
type HookInput struct {
	SessionID     string `json:"session_id"`
	CWD           string `json:"cwd"`
	HookEventName string `json:"hook_event_name"`
	// Raw holds the full parsed JSON for extracting event-specific fields.
	Raw map[string]any `json:"-"`
}

// ParseHookInput reads and parses hook JSON from a reader.
func ParseHookInput(r io.Reader) (*HookInput, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	var input HookInput
	if err := json.Unmarshal(data, &input); err != nil {
		return nil, err
	}

	var raw map[string]any
	json.Unmarshal(data, &raw)
	input.Raw = raw

	return &input, nil
}

// EventType returns the most specific event type available.
// It checks known field names in the raw JSON, falling back to HookEventName.
func (h *HookInput) EventType() string {
	for _, key := range []string{"notification_type", "type", "matcher"} {
		if v, ok := h.Raw[key].(string); ok && v != "" {
			return v
		}
	}
	return h.HookEventName
}
