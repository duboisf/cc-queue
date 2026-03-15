//go:build ignore

package main

import (
	"fmt"
	"os"

	"github.com/duboisf/cc-queue/internal/lintdocs"
)

func main() {
	root := "."
	if len(os.Args) > 1 {
		root = os.Args[1]
	}
	if err := lintdocs.Run(root); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
