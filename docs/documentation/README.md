# Documentation

How to maintain the cc-queue agent documentation system.

## Quick rules

- Every file under `docs/` must be reachable from `CLAUDE.md` via markdown links.
- Run `make lint-docs` before committing doc changes — it fails on orphaned files and broken links.
- One topic per file, split files exceeding ~100 lines.
- Update this doc in the same PR when changing the conventions it describes.

## Structure

```
CLAUDE.md          → docs gate + project overview + topic index
docs/
├── <topic>/
│   ├── README.md  → quick rules + links to detailed files
│   └── *.md       → one focused topic per file
└── rfcs/
    └── README.md  → RFC index
```

`CLAUDE.md` is the single source of truth for agent instructions. Claude Code loads it automatically at session start.

## Adding a new doc file

1. Create the file under the appropriate `docs/<topic>/` directory.
2. Add a link to it from the topic's `README.md` under "Detailed guides".
3. Run `make lint-docs` to verify the file is reachable and all links resolve.

## Adding a new docs topic

1. Create `docs/<topic>/README.md` following the README pattern (title, description, quick rules, detailed guides).
2. Add a `## <Topic>` section in `CLAUDE.md` with a sentence-level description and link to the README.
3. Run `make lint-docs` to verify.

## Removing a doc file

1. Delete the file.
2. Remove all links to it from README files.
3. If the directory is now empty, remove it and its section from `CLAUDE.md`.
4. Run `make lint-docs` to verify no broken links remain.

## README pattern

Every `docs/<topic>/README.md` follows this structure:

```markdown
# <Topic Name>

<One sentence describing what this area covers.>

## Quick rules

- Rule 1
- Rule 2
- Update this doc in the same PR when changing the conventions it describes.

## Detailed guides

- [Sub-topic A](./sub-topic-a.md) — brief description
```

## Testing the docs gate

Spawn a non-interactive Claude session with read-only tools and ask questions whose correct answers require reading the linked docs. For example:

- "What file format does the queue use?" (answer requires reading queue-storage.md)
- "How do I add shell completions?" (answer requires reading shell-completions.md)

If the agent answers correctly without being told where to look, the docs gate is working.
