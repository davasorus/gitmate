package main

import (
	"fmt"

	"github.com/davasorus/gitmate/internal/gitops"
	"github.com/spf13/cobra"
)

func init() { rootCmd.AddCommand(blameCmd) }

var blameCmd = &cobra.Command{
	Use:   "blame <path>",
	Short: "Show line-by-line authorship for a file",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		lines, err := gitops.Blame(".", args[0])
		if err != nil {
			return err
		}
		for _, l := range lines {
			fmt.Printf("%s %-16s %4d  %s\n", l.Short, truncate(l.Author, 16), l.Line, l.Content)
		}
		return nil
	},
}

func truncate(s string, n int) string {
	if len(s) > n {
		return s[:n-1] + "…"
	}
	return s
}
