package cmd

import (
	"fmt"

	"github.com/duboisf/cc-queue/internal/queue"
	"github.com/spf13/cobra"
)

func newDebugCmd(opts Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "debug [on|off]",
		Short: "Show or toggle debug logging",
		Long: `Show or toggle debug logging for cc-queue.

When debug is enabled, all cc-queue commands log to a debug.log file
in the queue directory (~/.local/state/cc-queue/debug.log).

Without arguments, shows the current debug status.
With "on" or "off", enables or disables debug logging.`,
		Args: cobra.MaximumNArgs(1),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) == 0 {
				return []string{"on", "off"}, cobra.ShellCompDirectiveNoFileComp
			}
			return nil, cobra.ShellCompDirectiveNoFileComp
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				cfg := queue.ReadConfig()
				if cfg.Debug {
					fmt.Fprintln(opts.Stdout, "debug: on")
				} else {
					fmt.Fprintln(opts.Stdout, "debug: off")
				}
				return nil
			}

			switch args[0] {
			case "on":
				cfg := queue.ReadConfig()
				cfg.Debug = true
				if err := queue.WriteConfig(cfg); err != nil {
					return fmt.Errorf("enabling debug: %w", err)
				}
				fmt.Fprintln(opts.Stdout, "debug: on")
			case "off":
				cfg := queue.ReadConfig()
				cfg.Debug = false
				if err := queue.WriteConfig(cfg); err != nil {
					return fmt.Errorf("disabling debug: %w", err)
				}
				fmt.Fprintln(opts.Stdout, "debug: off")
			default:
				return fmt.Errorf("unknown argument %q (use \"on\" or \"off\")", args[0])
			}
			return nil
		},
	}
	return cmd
}
