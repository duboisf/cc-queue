package cmd

import (
	"io"
	"os"
	"os/exec"
	"time"

	"github.com/duboisf/cc-queue/internal/kitty"
	"github.com/duboisf/cc-queue/internal/queue"
	"github.com/spf13/cobra"
)

// Options holds injectable dependencies for all commands.
type Options struct {
	// TimeNow returns the current time. Defaults to time.Now.
	TimeNow func() time.Time
	// Stdin for command input.
	Stdin io.Reader
	// Stdout for command output.
	Stdout io.Writer
	// Stderr for error output.
	Stderr io.Writer
	// FullTabber manages kitty tab layout for full-tab overlays.
	FullTabber kitty.FullTabber
	// CleanStaleWindowsFn removes entries with dead kitty windows. Nil to skip.
	CleanStaleWindowsFn func()
}

// NewRootCmd creates the root cobra command with all subcommands wired up.
func NewRootCmd(opts Options) *cobra.Command {
	root := &cobra.Command{
		Use:           "cc-queue",
		Short:         "Claude Code input queue for kitty terminal",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE:          jumpRunE(opts),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return nil, cobra.ShellCompDirectiveNoFileComp
		},
	}
	root.Flags().Bool("full-tab", false, "Use stack layout to cover the entire tab, restore on exit")
	_ = root.RegisterFlagCompletionFunc("full-tab", cobra.NoFileCompletions)
	root.SetOut(opts.Stdout)
	root.SetErr(opts.Stderr)

	root.AddGroup(
		&cobra.Group{ID: "core", Title: "Core Commands:"},
		&cobra.Group{ID: "setup", Title: "Setup Commands:"},
	)

	pushCmd := newPushCmd(opts)
	pushCmd.GroupID = "core"
	popCmd := newPopCmd(opts)
	popCmd.GroupID = "core"
	listCmd := newListCmd(opts)
	listCmd.GroupID = "core"
	clearCmd := newClearCmd(opts)
	clearCmd.GroupID = "core"
	cleanCmd := newCleanCmd(opts)
	cleanCmd.GroupID = "core"
	firstCmd := newFirstCmd(opts)
	firstCmd.GroupID = "core"

	installCmd := newInstallCmd(opts)
	installCmd.GroupID = "setup"
	completionCmd := newCompletionCmd()
	completionCmd.GroupID = "setup"
	versionCmd := newVersionCmd()
	versionCmd.GroupID = "setup"

	root.SetHelpCommand(&cobra.Command{Hidden: true})

	root.AddCommand(
		pushCmd,
		popCmd,
		listCmd,
		clearCmd,
		cleanCmd,
		firstCmd,
		installCmd,
		completionCmd,
		versionCmd,
		newEndCmd(opts),
		newListFzfCmd(opts),
		newPreviewCmd(),
		newJumpInternalCmd(),
	)
	return root
}

// Execute creates the root command with default options and runs it.
func Execute() error {
	opts := DefaultOptions()
	return NewRootCmd(opts).Execute()
}

// DefaultOptions returns production-ready Options with standard I/O.
func DefaultOptions() Options {
	return Options{
		TimeNow:    time.Now,
		Stdin:      os.Stdin,
		Stdout:     os.Stdout,
		Stderr:     os.Stderr,
		FullTabber: kitty.NewLayoutManager(),
		CleanStaleWindowsFn: func() {
			out, err := exec.Command("kitty", "@", "ls").Output()
			if err != nil {
				return
			}
			ids, err := kitty.ParseWindowIDs(out)
			if err != nil {
				return
			}
			queue.CleanStaleWindows(ids)
		},
	}
}
