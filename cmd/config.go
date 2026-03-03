package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/duboisf/cc-queue/internal/queue"
	"github.com/spf13/cobra"
)

func configViewRunE(opts Options) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		path := queue.ConfigPath()
		fmt.Fprintf(opts.Stdout, "Config: %s\n\n", path)

		data, err := os.ReadFile(path)
		if err != nil {
			// File doesn't exist — show defaults.
			fmt.Fprintln(opts.Stdout, queue.DefaultConfigJSON())
			return nil
		}
		fmt.Fprint(opts.Stdout, string(data))
		return nil
	}
}

func newConfigCmd(opts Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "View or edit cc-queue configuration",
		Args:  cobra.NoArgs,
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return nil, cobra.ShellCompDirectiveNoFileComp
		},
		RunE: configViewRunE(opts),
	}

	cmd.AddCommand(newConfigViewCmd(opts))
	cmd.AddCommand(newConfigEditCmd(opts))

	return cmd
}

func newConfigViewCmd(opts Options) *cobra.Command {
	return &cobra.Command{
		Use:   "view",
		Short: "Print config path and contents",
		Args:  cobra.NoArgs,
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return nil, cobra.ShellCompDirectiveNoFileComp
		},
		RunE: configViewRunE(opts),
	}
}

func newConfigEditCmd(opts Options) *cobra.Command {
	return &cobra.Command{
		Use:   "edit",
		Short: "Open config file in $EDITOR",
		Args:  cobra.NoArgs,
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return nil, cobra.ShellCompDirectiveNoFileComp
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			editor := os.Getenv("VISUAL")
			if editor == "" {
				editor = os.Getenv("EDITOR")
			}
			if editor == "" {
				return fmt.Errorf("$VISUAL or $EDITOR must be set")
			}

			// Ensure config file exists with defaults if missing.
			path := queue.ConfigPath()
			if _, err := os.Stat(path); os.IsNotExist(err) {
				if err := queue.WriteConfig(queue.Config{}); err != nil {
					return fmt.Errorf("creating default config: %w", err)
				}
			}

			c := exec.Command(editor, path)
			c.Stdin = os.Stdin
			c.Stdout = os.Stdout
			c.Stderr = os.Stderr
			return c.Run()
		},
	}
}
