package cmd

import (
	"fmt"

	"github.com/duboisf/cc-queue/internal/queue"
	"github.com/spf13/cobra"
)

func newInstallCmd(opts Options) *cobra.Command {
	var user, project bool
	var pickerShortcut, firstShortcut string

	cmd := &cobra.Command{
		Use:   "install",
		Short: "Install CC hooks and optional kitty shortcuts",
		Args:  cobra.NoArgs,
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return nil, cobra.ShellCompDirectiveNoFileComp
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			target := queue.TargetUser
			if project {
				target = queue.TargetProject
			}

			if err := queue.InstallHooks(target); err != nil {
				return err
			}

			settingsPath, _ := queue.SettingsPath(target)
			fmt.Fprintf(opts.Stdout, "Hooks installed in %s\n", settingsPath)

			shortcuts := queue.KittyShortcuts{
				Picker: pickerShortcut,
				First:  firstShortcut,
			}
			kittyPath, err := queue.InstallKittyShortcut(shortcuts)
			if err != nil {
				return err
			}
			if kittyPath != "" {
				fmt.Fprintf(opts.Stdout, "Kitty shortcuts installed in %s\n", kittyPath)
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&user, "user", false, "Install to ~/.claude/settings.json (default)")
	cmd.Flags().BoolVar(&project, "project", false, "Install to .claude/settings.json in cwd")
	cmd.Flags().StringVar(&pickerShortcut, "picker-shortcut", "", "Kitty shortcut for fzf picker overlay (e.g. 'kitty_mod+shift+q')")
	cmd.Flags().StringVar(&firstShortcut, "first-shortcut", "", "Kitty shortcut for jump-to-first (e.g. 'kitty_mod+shift+u')")
	_ = cmd.RegisterFlagCompletionFunc("user", cobra.NoFileCompletions)
	_ = cmd.RegisterFlagCompletionFunc("project", cobra.NoFileCompletions)
	_ = cmd.RegisterFlagCompletionFunc("picker-shortcut", cobra.NoFileCompletions)
	_ = cmd.RegisterFlagCompletionFunc("first-shortcut", cobra.NoFileCompletions)

	return cmd
}
