package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/duboisf/cc-queue/internal/queue"
	"github.com/spf13/cobra"
)

func newInstallCmd(opts Options) *cobra.Command {
	var user, project, force, desktop bool
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
				Shell:  os.Getenv("SHELL"),
			}

			// When --force overwriting an existing file, preserve shortcuts
			// that the user didn't explicitly override via flags.
			if force {
				existing := readExistingShortcuts()
				if !cmd.Flags().Changed("picker-shortcut") {
					shortcuts.Picker = existing.Picker
				}
				if !cmd.Flags().Changed("first-shortcut") {
					shortcuts.First = existing.First
				}
			}

			// Preview the kitty config content before writing.
			content := queue.BuildKittyConfig(shortcuts)
			fmt.Fprintf(opts.Stdout, "\nCreating kitty config with:\n\n")
			for _, line := range strings.Split(content, "\n") {
				fmt.Fprintf(opts.Stdout, "  %s\n", line)
			}

			result, err := queue.InstallKittyConfig(shortcuts, force)
			if err != nil {
				return err
			}
			if result == nil {
				fmt.Fprintln(opts.Stdout, "Kitty config dir not found, skipping")
			} else if result.Skipped {
				fmt.Fprintf(opts.Stdout, "%s already exists, skipping (use --force to overwrite)\n", result.ConfPath)
			} else {
				fmt.Fprintf(opts.Stdout, "Created %s\n", result.ConfPath)
			}
			if result != nil && result.Included {
				fmt.Fprintf(opts.Stdout, "Added 'include cc-queue.conf' to %s\n", result.MainConf)
			}

			if desktop {
				shell := os.Getenv("SHELL")
				dr, err := queue.InstallDesktopEntry(shell, force)
				if err != nil {
					return err
				}
				if dr.Skipped {
					fmt.Fprintf(opts.Stdout, "%s already exists, skipping (use --force to overwrite)\n", dr.Path)
				} else {
					fmt.Fprintf(opts.Stdout, "Created %s\n", dr.Path)
				}
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&user, "user", false, "Install to ~/.claude/settings.json (default)")
	cmd.Flags().BoolVar(&project, "project", false, "Install to .claude/settings.json in cwd")
	cmd.Flags().BoolVar(&force, "force", false, "Overwrite existing cc-queue.conf (preserves shortcuts unless overridden)")
	cmd.Flags().StringVar(&pickerShortcut, "picker-shortcut", "", "Kitty shortcut for fzf picker overlay (e.g. 'kitty_mod+shift+q')")
	cmd.Flags().StringVar(&firstShortcut, "first-shortcut", "", "Kitty shortcut for jump-to-first (e.g. 'kitty_mod+shift+u')")
	_ = cmd.RegisterFlagCompletionFunc("user", cobra.NoFileCompletions)
	_ = cmd.RegisterFlagCompletionFunc("project", cobra.NoFileCompletions)
	_ = cmd.RegisterFlagCompletionFunc("force", cobra.NoFileCompletions)
	_ = cmd.RegisterFlagCompletionFunc("picker-shortcut", cobra.NoFileCompletions)
	_ = cmd.RegisterFlagCompletionFunc("first-shortcut", cobra.NoFileCompletions)
	cmd.Flags().BoolVar(&desktop, "desktop", false, "Create .desktop entry for GNOME app launcher")
	_ = cmd.RegisterFlagCompletionFunc("desktop", cobra.NoFileCompletions)

	return cmd
}

// readExistingShortcuts reads shortcuts from the existing cc-queue.conf file.
// Returns empty shortcuts if the file doesn't exist or can't be parsed.
func readExistingShortcuts() queue.KittyShortcuts {
	home, err := os.UserHomeDir()
	if err != nil {
		return queue.KittyShortcuts{}
	}
	confPath := home + "/.config/kitty/cc-queue.conf"
	data, err := os.ReadFile(confPath)
	if err != nil {
		return queue.KittyShortcuts{}
	}
	return queue.ParseKittyShortcuts(string(data))
}
