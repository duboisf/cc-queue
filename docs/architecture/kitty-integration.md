# Kitty Integration

cc-queue uses kitty's remote control protocol to map sessions to windows and switch focus.

## Window mapping

Each queue entry records `$KITTY_WINDOW_ID` and `$KITTY_LISTEN_ON` from the environment at push time. These map a CC session to a specific kitty window.

## Focus and jump

`kitty @ focus-window --match id:{wid}` switches to the tab/window for a given session. The `first` command jumps to the first session needing attention. The `list` command launches fzf and jumps on selection.

## Layout switching

The `FullTabber` interface (`internal/kitty/layout.go`) handles switching a kitty window to full-tab layout and back:

```go
type FullTabber interface {
    FullTab(windowID string) (restore func(), err error)
}
```

Commands use the `--full-tab` flag to temporarily maximize a window when launching overlays.

## Kitty config requirements

cc-queue requires kitty's remote control:

```
allow_remote_control  socket-only
listen_on             unix:/tmp/kitty-{kitty_pid}
```

`cc-queue install --kitty` generates `~/.config/kitty/cc-queue.conf` with these settings and optional keyboard shortcuts.

## Parsing kitty state

`internal/kitty/windows.go` parses the JSON output of `kitty @ ls` to enumerate OS windows, tabs, and individual windows with their IDs, titles, and foreground processes.
