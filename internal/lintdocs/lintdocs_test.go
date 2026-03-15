package lintdocs

import (
	"os"
	"path/filepath"
	"testing"
)

func setupTree(t *testing.T, files map[string]string) string {
	t.Helper()
	root := t.TempDir()
	for name, content := range files {
		path := filepath.Join(root, name)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

func TestRun_AllLinked(t *testing.T) {
	root := setupTree(t, map[string]string{
		"CLAUDE.md": "See [docs/topic/](docs/topic/README.md) for details.",
		"docs/topic/README.md": "# Topic\n\n- [Detail](./detail.md)\n",
		"docs/topic/detail.md": "# Detail\n",
	})

	if err := Run(root); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

func TestRun_OrphanDetected(t *testing.T) {
	root := setupTree(t, map[string]string{
		"CLAUDE.md":            "See [docs/topic/](docs/topic/README.md).",
		"docs/topic/README.md": "# Topic\n",
		"docs/topic/orphan.md": "# Orphan\n",
	})

	err := Run(root)
	if err == nil {
		t.Fatal("expected error for orphaned doc")
	}
}

func TestRun_BrokenLinkDetected(t *testing.T) {
	root := setupTree(t, map[string]string{
		"CLAUDE.md":            "See [docs/topic/](docs/topic/README.md).",
		"docs/topic/README.md": "# Topic\n\n- [Missing](./missing.md)\n",
	})

	err := Run(root)
	if err == nil {
		t.Fatal("expected error for broken link")
	}
}

func TestRun_CodeBlockLinksIgnored(t *testing.T) {
	root := setupTree(t, map[string]string{
		"CLAUDE.md": "See [docs/topic/](docs/topic/README.md).",
		"docs/topic/README.md": "# Topic\n\n```markdown\n[Example](./nonexistent.md)\n```\n",
	})

	if err := Run(root); err != nil {
		t.Fatalf("expected code block links to be ignored, got: %v", err)
	}
}

func TestRun_NonMdOrphanDetected(t *testing.T) {
	root := setupTree(t, map[string]string{
		"CLAUDE.md":            "See [docs/topic/](docs/topic/README.md).",
		"docs/topic/README.md": "# Topic\n",
		"docs/topic/diagram.png": "fake image data",
	})

	err := Run(root)
	if err == nil {
		t.Fatal("expected error for orphaned non-.md file")
	}
}

func TestRun_NonMdFileLinked(t *testing.T) {
	root := setupTree(t, map[string]string{
		"CLAUDE.md":            "See [docs/topic/](docs/topic/README.md).",
		"docs/topic/README.md": "# Topic\n\n![Diagram](./diagram.png)\n",
		"docs/topic/diagram.png": "fake image data",
	})

	if err := Run(root); err != nil {
		t.Fatalf("expected linked non-.md file to pass, got: %v", err)
	}
}

func TestRun_MissingClaudeFile(t *testing.T) {
	root := t.TempDir()
	os.MkdirAll(filepath.Join(root, "docs"), 0o755)

	err := Run(root)
	if err == nil {
		t.Fatal("expected error for missing CLAUDE.md")
	}
}

func TestRun_MissingDocsDir(t *testing.T) {
	root := t.TempDir()
	os.WriteFile(filepath.Join(root, "CLAUDE.md"), []byte("# Agents"), 0o644)

	err := Run(root)
	if err == nil {
		t.Fatal("expected error for missing docs/")
	}
}

func TestRun_TransitiveReachability(t *testing.T) {
	root := setupTree(t, map[string]string{
		"CLAUDE.md":            "See [docs/a/](docs/a/README.md).",
		"docs/a/README.md":     "# A\n\n- [B](./b.md)\n",
		"docs/a/b.md":          "# B\n\n- [C](./c.md)\n",
		"docs/a/c.md":          "# C\n",
	})

	if err := Run(root); err != nil {
		t.Fatalf("expected transitive links to be followed, got: %v", err)
	}
}
