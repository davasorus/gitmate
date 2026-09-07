package main

import (
	"fmt"
	"os"

	"github.com/davasorus/gitmate/internal/gitops"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "gitmate",
	Short: "A git and GitHub companion tool",
}

func main() {
	// If git invoked us as its rebase sequence/message editor, do that and exit
	// before touching Cobra. (Interactive-rebase driving — see internal/gitops.)
	if handled, code := gitops.MaybeRunRebaseEditor(os.Args); handled {
		os.Exit(code)
	}
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
