package main

import (
	"fmt"

	"github.com/davasorus/gitmate/internal/gitops"
	"github.com/spf13/cobra"
)

func init() { rootCmd.AddCommand(cloneCmd) }

var cloneCmd = &cobra.Command{
	Use:   "clone <url> [dir]",
	Short: "Clone a repository (into [dir], or a dir named after the repo)",
	Args:  cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		dest := ""
		if len(args) == 2 {
			dest = args[1]
		}
		path, err := gitops.Clone(args[0], dest)
		if err != nil {
			return err
		}
		fmt.Printf("cloned into %s\n", path)
		return nil
	},
}
