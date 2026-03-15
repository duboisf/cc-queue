// Package lintdocs verifies that all files under docs/ are reachable
// from CLAUDE.md and that all relative markdown links resolve to existing files.
package lintdocs

import (
	"bufio"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

var linkRe = regexp.MustCompile(`\[.*?\]\(([^)]+)\)`)

// Run performs orphan detection and broken link detection starting from root.
// It returns a non-nil error if any problems are found or if setup fails.
func Run(root string) error {
	agentsFile := filepath.Join(root, "CLAUDE.md")
	docsDir := filepath.Join(root, "docs")

	if _, err := os.Stat(agentsFile); err != nil {
		return fmt.Errorf("%s not found", agentsFile)
	}
	if _, err := os.Stat(docsDir); err != nil {
		return fmt.Errorf("%s not found", docsDir)
	}

	// Collect all files under docs/ (not just .md — images, diagrams, etc. must also be linked)
	var allDocs []string
	if err := filepath.WalkDir(docsDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			abs, err := filepath.Abs(path)
			if err != nil {
				return err
			}
			allDocs = append(allDocs, abs)
		}
		return nil
	}); err != nil {
		return fmt.Errorf("walking docs/: %w", err)
	}
	sort.Strings(allDocs)

	// BFS from CLAUDE.md to discover reachable files
	visited := map[string]bool{}
	agentsAbs, _ := filepath.Abs(agentsFile)
	visited[agentsAbs] = true

	queue := extractLinks(agentsFile)
	for i := 0; i < len(queue); i++ {
		link := queue[i]
		if visited[link] {
			continue
		}
		visited[link] = true
		if info, err := os.Stat(link); err == nil && !info.IsDir() {
			queue = append(queue, extractLinks(link)...)
		}
	}

	// Orphan detection
	var orphans []string
	for _, doc := range allDocs {
		if !visited[doc] {
			rel, _ := filepath.Rel(root, doc)
			orphans = append(orphans, rel)
		}
	}

	// Broken link detection — check CLAUDE.md and all .md files under docs/
	type brokenLink struct {
		source string
		target string
	}
	var broken []brokenLink

	mdFiles := []string{agentsFile}
	for _, f := range allDocs {
		if strings.HasSuffix(f, ".md") {
			mdFiles = append(mdFiles, f)
		}
	}
	for _, file := range mdFiles {
		for _, link := range extractRawLinks(file) {
			dir := filepath.Dir(file)
			resolved := filepath.Join(dir, link)
			if _, err := os.Stat(resolved); err != nil {
				srcRel, _ := filepath.Rel(root, file)
				broken = append(broken, brokenLink{source: srcRel, target: link})
			}
		}
	}

	// Report
	hasErrors := false

	if len(orphans) > 0 {
		hasErrors = true
		fmt.Println("ORPHANED DOCS (not reachable from CLAUDE.md):")
		for _, o := range orphans {
			fmt.Printf("  %s\n", o)
		}
	}

	if len(broken) > 0 {
		hasErrors = true
		fmt.Println("BROKEN LINKS:")
		for _, b := range broken {
			fmt.Printf("  %s -> %s\n", b.source, b.target)
		}
	}

	if hasErrors {
		return fmt.Errorf("lint-docs failed")
	}

	fmt.Printf("lint-docs: OK (%d docs, all reachable, no broken links)\n", len(allDocs))
	return nil
}

// extractLinks returns absolute paths of all relative markdown link targets
// found outside fenced code blocks in file.
func extractLinks(file string) []string {
	raw := extractRawLinks(file)
	dir := filepath.Dir(file)
	var out []string
	for _, link := range raw {
		resolved := filepath.Join(dir, link)
		abs, err := filepath.Abs(resolved)
		if err != nil {
			continue
		}
		out = append(out, abs)
	}
	return out
}

// extractRawLinks returns the raw relative paths from markdown links in file,
// skipping fenced code blocks, URLs, anchors, and absolute paths.
func extractRawLinks(file string) []string {
	f, err := os.Open(file)
	if err != nil {
		return nil
	}
	defer f.Close()

	var links []string
	inCodeBlock := false
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(strings.TrimSpace(line), "```") {
			inCodeBlock = !inCodeBlock
			continue
		}
		if inCodeBlock {
			continue
		}
		for _, match := range linkRe.FindAllStringSubmatch(line, -1) {
			link := match[1]
			if strings.HasPrefix(link, "http://") || strings.HasPrefix(link, "https://") {
				continue
			}
			if strings.HasPrefix(link, "#") {
				continue
			}
			if strings.HasPrefix(link, "/") {
				continue
			}
			// Strip anchor fragment
			if idx := strings.Index(link, "#"); idx >= 0 {
				link = link[:idx]
			}
			if link == "" {
				continue
			}
			links = append(links, link)
		}
	}
	return links
}
