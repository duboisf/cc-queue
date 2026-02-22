# cc-queue - Project Rules

## Build

```sh
make build    # compile binary
make test     # run tests with -race
make cover    # run tests with coverage report
make install  # go install
make deps     # go mod tidy
```

## Shell Completions

Every command and subcommand MUST provide shell completions for all arguments and flag values:

- **Positional arguments**: Use `ValidArgsFunction` to provide dynamic completions (e.g., session IDs, queue entry identifiers).
- **Flag values**: Use `cmd.RegisterFlagCompletionFunc` for any flag that accepts a fixed set of values (e.g., `--scope`).
- **File suppression**: Return `cobra.ShellCompDirectiveNoFileComp` when completions should not fall back to file paths.

When adding a new command, verify completions work by running:
```
cc-queue completion zsh | source /dev/stdin
cc-queue <command> <TAB>
```

## Testing

- **TDD**: Write failing tests first, verify they fail, then implement the fix. Red → Green → Refactor.
- Every command must have corresponding tests.
- Use dependency injection via `Options` struct for testability.
- All tests must run with `-race` flag.

## Architecture

cc-queue is a Claude Code hook integration for managing an input queue across multiple kitty terminal tabs/windows.

- **Hook system**: `cc-queue push` is triggered on `Notification` hooks (permission prompts, idle, elicitation). `cc-queue pop` is triggered on `UserPromptSubmit` to remove entries when the user responds.
- **Queue storage**: JSON files in `~/.local/state/cc-queue/`, one per CC session (keyed by session ID).
- **Tab mapping**: Uses `$KITTY_WINDOW_ID` to map sessions to kitty windows.
- **Jump**: `kitty @ focus-window --match id:{wid}` to switch to the correct tab/window.
