# Shell Completions

Every command and subcommand MUST provide shell completions for all arguments and flag values.

## Rules

- **Positional arguments**: Use `ValidArgsFunction` to provide dynamic completions (e.g., session IDs).
- **Flag values**: Use `cmd.RegisterFlagCompletionFunc` for any flag with a fixed set of values.
- **File suppression**: Return `cobra.ShellCompDirectiveNoFileComp` when completions should not fall back to file paths.

## Example: positional argument completion

```go
cmd.ValidArgsFunction = func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
    entries, _ := queue.List()
    var ids []string
    for _, e := range entries {
        if strings.HasPrefix(e.SessionID, toComplete) {
            ids = append(ids, e.SessionID)
        }
    }
    return ids, cobra.ShellCompDirectiveNoFileComp
}
```

## Example: flag value completion

```go
cmd.RegisterFlagCompletionFunc("scope", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
    return []string{"user", "project"}, cobra.ShellCompDirectiveNoFileComp
})
```

## Verification

After adding a command, verify completions work:

```sh
cc-queue completion zsh | source /dev/stdin
cc-queue <command> <TAB>
```
