package queue

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// DebugEnabled returns true when CC_QUEUE_DEBUG is set to a non-empty value.
func DebugEnabled() bool {
	return os.Getenv("CC_QUEUE_DEBUG") != ""
}

// Debugf appends a timestamped line to the debug log file in the queue directory.
// Silently does nothing if CC_QUEUE_DEBUG is not set or the log cannot be opened.
func Debugf(format string, args ...any) {
	if !DebugEnabled() {
		return
	}
	logPath := filepath.Join(Dir(), "debug.log")
	f, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer f.Close()
	msg := fmt.Sprintf(format, args...)
	fmt.Fprintf(f, "%s  %s\n", time.Now().Format("2006-01-02T15:04:05.000"), msg)
}
