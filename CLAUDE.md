# cc-queue

CLI tool and Claude Code hook integration for managing an input queue across multiple kitty terminal tabs/windows.

When running multiple Claude Code sessions in different kitty tabs, CC may need user input (permission prompts, questions, idle waiting) while you're in another tab. cc-queue captures these events via CC hooks, stores them in a local queue, and provides an fzf picker to jump to the right kitty window.

## Architecture

- **Hook scripts**: `cc-queue push` (on `Notification` hook) and `cc-queue pop` (on `UserPromptSubmit` hook) manage queue entries
- **Queue storage**: JSON files in `~/.local/state/cc-queue/`, one per CC session (keyed by session ID)
- **Tab mapping**: Uses `$KITTY_WINDOW_ID` (inherited from kitty) to map sessions to windows
- **Jump**: `kitty @ focus-window --match id:{wid}` to switch to the correct tab/window

## Project layout

```
cmd/cc-queue/       CLI entry point
internal/queue/     Core queue logic (entry CRUD, formatting, hook parsing, install)
```

## Commands

| Command | Description |
|---------|-------------|
| `cc-queue` | fzf picker (default) — select and jump to kitty window |
| `cc-queue list` | Plain text list of pending items |
| `cc-queue push` | Hook handler: add entry (reads CC hook JSON from stdin) |
| `cc-queue pop` | Hook handler: remove entry (reads CC hook JSON from stdin) |
| `cc-queue clear` | Remove all entries |
| `cc-queue clean` | Remove stale entries (dead PIDs) |
| `cc-queue install` | Install CC hooks (`--user` or `--project`) |

## Build & install

```sh
go build -o cc-queue ./cmd/cc-queue/
# Move binary somewhere in PATH, then:
cc-queue install --user
```

## Hook events captured

- `permission_prompt` — CC needs tool permission (label: PERM)
- `idle_prompt` — CC finished its turn, waiting for input (label: IDLE)
- `elicitation_dialog` — CC is asking a question (label: ASK)
