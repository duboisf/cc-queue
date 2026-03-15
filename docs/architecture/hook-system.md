# Hook System

cc-queue integrates with Claude Code via its hook system. Hooks are shell commands that Claude Code runs at specific lifecycle events.

## Hook events

| Hook Event         | Matcher                                              | Command         | Effect                        |
|--------------------|------------------------------------------------------|-----------------|-------------------------------|
| `Notification`     | `permission_prompt\|idle_prompt\|elicitation_dialog` | `cc-queue push` | Add entry to queue            |
| `UserPromptSubmit` | *(none)*                                             | `cc-queue pop`  | Mark session as "working"     |
| `SessionStart`     | *(none)*                                             | `cc-queue push` | Track session startup         |
| `SessionEnd`       | *(none)*                                             | `cc-queue end`  | Track session termination     |

## HookInput

Every hook receives JSON on stdin. `queue.ParseHookInput()` extracts:

```go
type HookInput struct {
    SessionID     string         // Unique CC session identifier
    CWD           string         // Working directory
    HookEventName string         // Event name (e.g., "Notification")
    Raw           map[string]any // Full JSON for event-specific fields
}
```

`EventType()` resolves the most specific event by checking `notification_type`, `type`, `matcher` fields in order, falling back to `HookEventName`.

## Adding a new hook

1. Define the hook event and matcher in `internal/queue/install.go`.
2. Create or modify the command in `cmd/` that handles the event.
3. Write tests using `testOptionsWithStdin()` to inject hook JSON.
4. Update `install` command if the hook should be auto-installed.
5. Update this doc.

## Installation

`cc-queue install` writes hook definitions to Claude Code's `settings.json` (user or project scope) and optionally generates kitty config.
