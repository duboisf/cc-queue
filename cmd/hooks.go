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
		Args:  cobra.NoArgs,
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return nil, cobra.ShellCompDirectiveNoFileComp
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			target := queue.TargetUser
			if project {
				target = queue.TargetProject
			}

			// Check current status before installing.
			before, _, err := queue.CheckHooks(target)
			if err != nil {
				return err
			}

			if before.AllInstalled() {
				settingsPath, _ := queue.SettingsPath(target)
				fmt.Fprintf(opts.Stdout, "All hooks already installed in %s\n", settingsPath)
				return nil
			}

			if err := queue.InstallHooks(target); err != nil {
				return err
			}

			settingsPath, _ := queue.SettingsPath(target)
			fmt.Fprintf(opts.Stdout, "Hooks installed in %s\n", settingsPath)
			return nil
		},
	}

	cmd.Flags().BoolVar(&user, "user", false, "Install to ~/.claude/settings.json (default)")
	cmd.Flags().BoolVar(&project, "project", false, "Install to .claude/settings.json in cwd")
	_ = cmd.RegisterFlagCompletionFunc("user", cobra.NoFileCompletions)
	_ = cmd.RegisterFlagCompletionFunc("project", cobra.NoFileCompletions)

	return cmd
}

func newHooksUninstallCmd(opts Options) *cobra.Command {
	var user, project bool

	cmd := &cobra.Command{
		Use:   "uninstall",
		Short: "Remove cc-queue hooks from Claude Code settings",
		Args:  cobra.NoArgs,
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return nil, cobra.ShellCompDirectiveNoFileComp
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			target := queue.TargetUser
			if project {
				target = queue.TargetProject
			}

			// Check current status before uninstalling.
			before, _, err := queue.CheckHooks(target)
			if err != nil {
				return err
			}

			if !before.AnyInstalled() {
				settingsPath, _ := queue.SettingsPath(target)
				fmt.Fprintf(opts.Stdout, "No cc-queue hooks found in %s\n", settingsPath)
				return nil
			}

			if err := queue.UninstallHooks(target); err != nil {
				return err
			}

			settingsPath, _ := queue.SettingsPath(target)
			fmt.Fprintf(opts.Stdout, "Hooks removed from %s\n", settingsPath)
			return nil
		},
	}

	cmd.Flags().BoolVar(&user, "user", false, "Remove from ~/.claude/settings.json (default)")
	cmd.Flags().BoolVar(&project, "project", false, "Remove from .claude/settings.json in cwd")
	_ = cmd.RegisterFlagCompletionFunc("user", cobra.NoFileCompletions)
	_ = cmd.RegisterFlagCompletionFunc("project", cobra.NoFileCompletions)

	return cmd
}
