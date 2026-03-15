# Queue Storage

Queue entries are stored as JSON files, one per Claude Code session.

## File layout

```
~/.local/state/cc-queue/
├── <session-id-1>.json
├── <session-id-2>.json
└── ...
```

The directory respects `$XDG_STATE_HOME` (defaults to `~/.local/state`).

Session IDs containing `/` or `\` are sanitized to `_` for safe filenames.

## SessionFile format

```go
type SessionFile struct {
    Current *Entry   // Most recent entry for this session
    History []*Entry // Up to 10 previous entries
}
```

## Entry struct

```go
type Entry struct {
    Timestamp     time.Time
    SessionID     string
    KittyWindowID string // from $KITTY_WINDOW_ID
    KittyListenOn string // from $KITTY_LISTEN_ON
    PID           int    // Claude Code process PID
    CWD           string
    Event         string // permission_prompt, idle_prompt, working, etc.
    Message       string // Optional event message
}
```

## History

When a new entry is written for an existing session, the previous `Current` is pushed to `History`. Maximum 10 historical entries. Duplicate events (same event + message) skip the history push.

## Deduplication

On every `List()` call, entries are deduplicated by `(KittyWindowID, CWD)`. When duplicates exist, the newest entry wins and older entries are removed. Entries with empty `KittyWindowID` are never deduplicated.

## File locking

All reads and writes use `syscall.Flock` for exclusive file-level locking. The `Locker` interface in `entry.go` abstracts this for testing.

## Stale entry cleanup

`CleanStale()` removes entries whose PID is no longer running (checked via `syscall.Kill(pid, 0)`). `CleanStaleWindows()` removes entries whose kitty window no longer exists.
