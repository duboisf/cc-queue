package cmd

import (
	"fmt"

	"github.com/duboisf/cc-queue/internal/queue"
	"github.com/spf13/cobra"
)

type hookEntry struct {
	name      string
	installed bool
}

func newHooksCmd(opts Options) *cobra.Command {
	var user, project bool

	cmd := &cobra.Command{
		Use:   "hooks",
		Short: "Show cc-queue hook status in Claude Code settings",
		Args:  cobra.NoArgs,
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return nil, cobra.ShellCompDirectiveNoFileComp
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			target := queue.TargetUser
			if project {
				target = queue.TargetProject
			}

			status, path, err := queue.CheckHooks(target)
			if err != nil {
				return err
			}

			fmt.Fprintf(opts.Stdout, "Settings: %s\n\n", path)

			hooks := []hookEntry{
				{"Notification", status.Notification},
				{"UserPromptSubmit", status.UserPromptSubmit},
				{"SessionStart", status.SessionStart},
				{"SessionEnd", status.SessionEnd},
			}

			for _, h := range hooks {
				mark := "x"
				if h.installed {
					mark = "v"
				}
				fmt.Fprintf(opts.Stdout, "  [%s] %s\n", mark, h.name)
			}

			if !status.AllInstalled() {
				fmt.Fprintf(opts.Stdout, "\nRun 'cc-queue hooks install' to install missing hooks.\n")
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&user, "user", false, "Check ~/.claude/settings.json (default)")
	cmd.Flags().BoolVar(&project, "project", false, "Check .claude/settings.json in cwd")
	cmd.MarkFlagsMutuallyExclusive("user", "project")
	_ = cmd.RegisterFlagCompletionFunc("user", cobra.NoFileCompletions)
	_ = cmd.RegisterFlagCompletionFunc("project", cobra.NoFileCompletions)

	cmd.AddCommand(newHooksInstallCmd(opts))
	cmd.AddCommand(newHooksUninstallCmd(opts))

	return cmd
}

func newHooksInstallCmd(opts Options) *cobra.Command {
	var user, project bool

	cmd := &cobra.Command{
		Use:   "install",
		Short: "Install cc-queue hooks into Claude Code settings",
		Long: `Install all four cc-queue hooks into Claude Code settings.

This is idempotent — running it multiple times is safe and will not
duplicate hooks. Existing hooks from other tools are preserved.

The following hooks are installed:

  Notification       Triggers "cc-queue push" on permission prompts,
                     idle prompts, and elicitation dialogs so they
                     appear in the queue.
  UserPromptSubmit   Triggers "cc-queue pop" when you respond to a
                     prompt, clearing the entry from the queue.
  SessionStart       Triggers "cc-queue push" to register new sessions.
  SessionEnd         Triggers "cc-queue end" to clean up finished sessions.

By default hooks are written to ~/.claude/settings.json (user-level).
Use --project to write to .claude/settings.json in the current directory.`,
		Args: cobra.NoArgs,
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return nil, cobra.ShellCompDirectiveNoFileComp
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			target := queue.TargetUser
			if project {
				target = queue.TargetProject
			}

			// Check current status before installing.
			before, path, err := queue.CheckHooks(target)
			if err != nil {
				return err
			}

			if before.AllInstalled() {
				fmt.Fprintf(opts.Stdout, "All hooks already installed in %s\n", path)
				return nil
			}

			if err := queue.InstallHooks(target); err != nil {
				return err
			}

			fmt.Fprintf(opts.Stdout, "Hooks installed in %s\n", path)
			return nil
		},
	}

	cmd.Flags().BoolVar(&user, "user", false, "Install to ~/.claude/settings.json (default)")
	cmd.Flags().BoolVar(&project, "project", false, "Install to .claude/settings.json in cwd")
	cmd.MarkFlagsMutuallyExclusive("user", "project")
	_ = cmd.RegisterFlagCompletionFunc("user", cobra.NoFileCompletions)
	_ = cmd.RegisterFlagCompletionFunc("project", cobra.NoFileCompletions)

	return cmd
}

func newHooksUninstallCmd(opts Options) *cobra.Command {
	var user, project bool

	cmd := &cobra.Command{
		Use:   "uninstall",
		Short: "Remove cc-queue hooks from Claude Code settings",
		Long: `Remove all cc-queue hooks from Claude Code settings.

This removes the Notification, UserPromptSubmit, SessionStart, and
SessionEnd hooks that were installed by "cc-queue hooks install".

Only cc-queue entries are removed — hooks from other tools sharing the
same event keys are left intact. If a matcher contains both a cc-queue
hook and another tool's hook, only the cc-queue hook is removed.

By default hooks are removed from ~/.claude/settings.json (user-level).
Use --project to target .claude/settings.json in the current directory.`,
		Args: cobra.NoArgs,
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return nil, cobra.ShellCompDirectiveNoFileComp
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			target := queue.TargetUser
			if project {
				target = queue.TargetProject
			}

			// Check current status before uninstalling.
			before, path, err := queue.CheckHooks(target)
			if err != nil {
				return err
			}

			if !before.AnyInstalled() {
				fmt.Fprintf(opts.Stdout, "No cc-queue hooks found in %s\n", path)
				return nil
			}

			if err := queue.UninstallHooks(target); err != nil {
				return err
			}

			fmt.Fprintf(opts.Stdout, "Hooks removed from %s\n", path)
			return nil
		},
	}

	cmd.Flags().BoolVar(&user, "user", false, "Remove from ~/.claude/settings.json (default)")
	cmd.Flags().BoolVar(&project, "project", false, "Remove from .claude/settings.json in cwd")
	cmd.MarkFlagsMutuallyExclusive("user", "project")
	_ = cmd.RegisterFlagCompletionFunc("user", cobra.NoFileCompletions)
	_ = cmd.RegisterFlagCompletionFunc("project", cobra.NoFileCompletions)

	return cmd
}
