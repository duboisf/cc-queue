# Development

Build commands, testing workflow, shell completions, and how to add new commands.

## Quick rules

- Run `make test` before every commit — all tests must pass with `-race`.
- Follow TDD: write a failing test first, verify it fails, then implement.
- Every command gets its own file in `cmd/` with a corresponding `*_test.go`.
- Every command MUST provide shell completions (see [Shell Completions](./shell-completions.md)).
- Inject dependencies via the `Options` struct — never use global state in commands.
- Update this doc in the same PR when changing the conventions it describes.

## Build commands

```sh
make build    # compile binary to ./cc-queue
make test     # go test -race ./...
make cover    # test with coverage report
make install  # go install to $GOPATH/bin
make deps     # go mod tidy
```

## Adding a new command

1. Create `cmd/<name>.go` with a `newXxxCmd(opts *Options)` constructor.
2. Create `cmd/<name>_test.go` with table-driven tests.
3. Register the command in `cmd/root.go`'s `NewRootCmd()`.
4. Add `ValidArgsFunction` and flag completions.
5. Update the source tree in `CLAUDE.md` if the file adds a new directory.

## Detailed guides

- [Testing](./testing.md) — TDD workflow, test helpers, isolation patterns
- [Shell Completions](./shell-completions.md) — completion rules, how to add and verify
