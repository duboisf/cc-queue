# cc-queue — Agent Rules

## Pre-Work Docs Gate (MANDATORY)

**STOP. Before writing any code, tests, or config changes, complete these steps:**

1. Identify which `docs/` area(s) this task touches (see sections below).
2. Read the linked README(s) using the Read tool.
3. Load ALL matching skills via the Skill tool.
4. Output `Docs consulted:` and `Skills loaded:` lines BEFORE any other work.
5. If no docs or skills apply, output `Docs consulted: none (<reason>)` and `Skills loaded: none (<reason>)`.

### Response Contract

Every response that involves code, test, or config changes MUST include:

```
Docs consulted: <list of docs read, or "none (reason)">
Skills loaded: <list of skills loaded, or "none (reason)">
```

### Plan Mode

When writing a plan file (Claude Code plan mode), include `Docs consulted:` and `Skills loaded:` in the plan file itself — not just in chat output that scrolls away.

---

## Documentation Linking Rule

Every file under `docs/` must be reachable from this file via markdown links. Run `make lint-docs` to verify — it checks for orphaned files and broken links. See [docs/documentation/](docs/documentation/README.md) for how to maintain this system.

---

## Project Overview

**cc-queue** is a Go CLI tool that integrates with Claude Code hooks to manage an input queue across multiple kitty terminal tabs/windows. When Claude Code needs attention (permission prompts, idle, elicitation), `cc-queue push` adds an entry. When the user responds, `cc-queue pop` removes it. An interactive picker (`cc-queue list`) lets users jump between sessions.

| Component      | Detail                              |
|----------------|-------------------------------------|
| Language       | Go 1.25+                           |
| CLI framework  | Cobra                              |
| Terminal       | kitty (remote control via `kitty @`)|
| Storage        | JSON files in `~/.local/state/cc-queue/` |
| Build          | `make build`, `make test`, `make install` |
| Testing        | Go standard library, `-race` flag   |

---

## Source Tree

```
cc-queue/
├── main.go                     # Entry point — calls cmd.Execute()
├── cmd/                        # Cobra command implementations
│   ├── root.go                 # Root command, Options DI container
│   ├── push.go                 # Notification hook handler — adds queue entry
│   ├── pop.go                  # UserPromptSubmit hook handler — removes entry
│   ├── list.go                 # Interactive picker via fzf
│   ├── first.go                # Jump to first session needing attention
│   ├── clear.go                # Remove all entries
│   ├── clean.go                # Remove stale entries
│   ├── completion.go           # Shell completion generation
│   ├── config.go               # Config view/edit subcommands
│   ├── debug.go                # Debug utilities
│   ├── hooks.go                # Hooks view subcommand
│   ├── install.go              # Install CC hooks + kitty shortcuts
│   ├── shell.go                # Launch shell in queue dir
│   ├── end.go                  # End session marker
│   ├── version.go              # Version display
│   └── *_test.go               # Command tests + helpers_test.go
├── internal/
│   ├── queue/                  # Core queue logic
│   │   ├── entry.go            # Entry struct, Read/Write/List/Remove
│   │   ├── hook.go             # HookInput parsing from CC hooks
│   │   ├── format.go           # Display formatting (age, labels, git branch)
│   │   ├── config.go           # Config struct and persistence
│   │   ├── debug.go            # Debug logging
│   │   ├── install.go          # Hook + kitty config installation
│   │   └── *_test.go           # Package tests
│   ├── conversation/           # Claude Code conversation file parsing
│   │   ├── conversation.go     # Parse JSONL conversation files
│   │   └── conversation_test.go
│   ├── kitty/                  # Kitty terminal integration
│   │   ├── windows.go          # Parse kitty @ ls JSON output
│   │   ├── layout.go           # FullTabber interface, layout switching
│   │   └── *_test.go           # Package tests
│   └── lintdocs/               # Docs linter
│       ├── lintdocs.go         # Orphan + broken link detection
│       └── lintdocs_test.go    # Linter tests
├── docs/                       # Agent documentation (you are here)
├── Makefile                    # Build targets
├── go.mod / go.sum             # Go module definition
└── mise.toml                   # Tool version management
```

---

## Architecture

Rules for the hook system, queue storage model, kitty terminal integration, and dependency injection patterns. Read these before modifying core data flow or adding new hook events.

[docs/architecture/](docs/architecture/README.md)

## Development

Build commands, TDD workflow, test patterns, shell completion requirements, and how to add new commands. Read these before writing any code or tests.

[docs/development/](docs/development/README.md)

## Documentation

How to maintain this docs system: adding files, adding topics, running lint, and keeping everything linked. Read this before modifying any docs.

[docs/documentation/](docs/documentation/README.md)

## RFCs

Design documents for significant changes. Each RFC has a date-prefixed filename and is indexed in the README.

[docs/rfcs/](docs/rfcs/README.md)
