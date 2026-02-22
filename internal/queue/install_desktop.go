package queue

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// DesktopInstallResult describes what was written during desktop entry installation.
type DesktopInstallResult struct {
	Path    string // path to cc-queue.desktop
	Skipped bool   // true if file already existed and force was false
}

// BuildDesktopEntry returns the content for a .desktop file that launches
// cc-queue in a new kitty OS window. Shell is the user's login shell;
// falls back to /bin/sh if empty.
func BuildDesktopEntry(shell string) string {
	if shell == "" {
		shell = "/bin/sh"
	}

	var b strings.Builder
	b.WriteString("[Desktop Entry]\n")
	b.WriteString("Name=cc-queue\n")
	b.WriteString("Comment=Claude Code input queue for kitty\n")
	fmt.Fprintf(&b, "Exec=kitty --detach --title cc-queue -- %s -ilc 'exec cc-queue'\n", shell)
	b.WriteString("Type=Application\n")
	b.WriteString("Terminal=false\n")
	b.WriteString("Categories=Development;\n")
	b.WriteString("Keywords=claude;code;queue;\n")
	return b.String()
}

// InstallDesktopEntry creates a .desktop file at
// ~/.local/share/applications/cc-queue.desktop. If force is false and the
// file already exists, it is not overwritten (Skipped is set to true).
func InstallDesktopEntry(shell string, force bool) (*DesktopInstallResult, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	appsDir := filepath.Join(home, ".local", "share", "applications")
	if err := os.MkdirAll(appsDir, 0755); err != nil {
		return nil, fmt.Errorf("creating %s: %w", appsDir, err)
	}

	desktopPath := filepath.Join(appsDir, "cc-queue.desktop")

	_, statErr := os.Stat(desktopPath)
	exists := statErr == nil

	if exists && !force {
		return &DesktopInstallResult{Path: desktopPath, Skipped: true}, nil
	}

	content := BuildDesktopEntry(shell)
	if err := os.WriteFile(desktopPath, []byte(content), 0644); err != nil {
		return nil, fmt.Errorf("writing %s: %w", desktopPath, err)
	}

	return &DesktopInstallResult{Path: desktopPath, Skipped: false}, nil
}
