package cmd

import (
	"fmt"

	"github.com/duboisf/cc-queue/internal/queue"
	"github.com/spf13/cobra"
)

func newInstallCmd(opts Options) *cobra.Command {
	var user, project bool

	cmd := &cobra.Command{
		Use:   "install",
		Short: "Install CC hooks (--user or --project)",
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
			return nil
		},
	}

	cmd.Flags().BoolVar(&user, "user", false, "Install to ~/.claude/settings.json (default)")
	cmd.Flags().BoolVar(&project, "project", false, "Install to .claude/settings.json in cwd")
	_ = cmd.RegisterFlagCompletionFunc("user", cobra.NoFileCompletions)
	_ = cmd.RegisterFlagCompletionFunc("project", cobra.NoFileCompletions)

	return cmd
}
