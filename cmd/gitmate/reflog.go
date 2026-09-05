package main

import (
	"fmt"

	"github.com/davasorus/gitmate/internal/gitops"
	"github.com/spf13/cobra"
)

var reflogLimit int

func init() {
	reflogCmd.Flags().IntVarP(&reflogLimit, "number", "n", 30, "limit number of entries (0 = all)")
	rootCmd.AddCommand(reflogCmd)
}

var reflogCmd = &cobra.Command{
	Use:   "reflog",
	Short: "Show where HEAD has been (the undo safety net)",
	RunE: func(cmd *cobra.Command, args []string) error {
		entries, err := gitops.Reflog(".", reflogLimit)
		if err != nil {
			return err
		}
		if len(entries) == 0 {
			fmt.Println("no reflog entries")
			return nil
		}
		for _, e := range entries {
			fmt.Printf("%s %-10s %-8s %s\n", e.Short, e.Selector, e.Action, e.Message)
		}
		return nil
	},
}
