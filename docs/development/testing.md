# Testing

## TDD workflow

1. **Red** — Write a failing test that describes the expected behavior.
2. **Green** — Write the minimum code to make the test pass.
3. **Refactor** — Clean up while keeping tests green.

Always verify the test fails before implementing the fix. This confirms the test actually exercises the code path.

## Running tests

```sh
make test     # go test -race ./...
make cover    # test + coverage report
```

All tests run with `-race`. If a test is flaky under the race detector, fix the race — don't disable the flag.

## Test helpers

Shared helpers live in `cmd/helpers_test.go`:

| Helper                     | Purpose                                      |
|----------------------------|----------------------------------------------|
| `testOptions()`            | Options with fixed time, captured stdout/stderr |
| `testOptionsWithStdin(s)`  | Same + custom stdin content                  |
| `setupQueueDir(t)`         | Isolate queue in `t.TempDir()` via `XDG_STATE_HOME` |
| `seedEntry(t, ...)`        | Create a test entry with session, cwd, event, window ID |
| `seedEntryAtTime(t, ...)`  | Same with explicit timestamp                 |
| `seedEntryWithMessage(t, ...)`| Same with event message                   |
| `seedEntryWithHistory(t, ...)` | Entry with historical events              |
| `seedEntryNoWindow(t, ...)` | Entry without kitty window ID               |
| `executeCommand(root, args)` | Run command, return stdout/stderr/error     |
| `entryCount(t)`            | Count entries in queue directory             |

## Isolation patterns

- **Filesystem**: `setupQueueDir(t)` sets `XDG_STATE_HOME` to a temp dir — tests never touch the real queue.
- **Time**: `testOptions()` fixes `TimeNow` to `2026-02-18T14:30:00Z` — tests produce deterministic output.
- **I/O**: `Options.Stdout` and `Options.Stderr` capture output in `bytes.Buffer` — no console side effects.
- **Kitty**: `nopFullTabber` returns no-op functions — tests don't require a running kitty instance.
- **File locking**: The `Locker` interface can be replaced for tests that need to simulate lock contention.

## Writing command tests

```go
func TestMyCommand(t *testing.T) {
    setupQueueDir(t)
    seedEntry(t, "sess-1", "/tmp/project", "permission_prompt", 1234)

    opts, _, _ := testOptions()
    root := cmd.NewRootCmd(opts)
    stdout, stderr, err := executeCommand(root, "mycommand", "--flag")

    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    // Assert on stdout, stderr, or queue state
}
```
