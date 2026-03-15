# Architecture

How cc-queue is structured: hook integration, queue storage, kitty terminal control, and dependency injection.

## Quick rules

- All external I/O is injected via the `Options` struct in `cmd/root.go` — never call `os.Stdin`, `os.Stdout`, or `time.Now()` directly in commands.
- One JSON file per CC session in `~/.local/state/cc-queue/` — never store multiple sessions in one file.
- Entry deduplication is by `(kitty_window_id, cwd)` tuple — newer entries replace older ones.
- Kitty remote control requires `allow_remote_control socket-only` and `listen_on` in kitty config.
- Update this doc in the same PR when changing the conventions it describes.

## Detailed guides

- [Hook System](./hook-system.md) — how Claude Code hooks trigger cc-queue commands
- [Queue Storage](./queue-storage.md) — JSON file format, session files, history, locking
- [Kitty Integration](./kitty-integration.md) — window mapping, focus, layout switching
