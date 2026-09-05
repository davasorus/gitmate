package main

import (
	"fmt"

	"github.com/davasorus/gitmate/internal/gitops"
	"github.com/spf13/cobra"
)

var pullRebase bool

func init() {
	pullCmd.Flags().BoolVar(&pullRebase, "rebase", false, "rebase local commits onto the upstream instead of merging")
	rootCmd.AddCommand(fetchCmd, pullCmd)
}

var fetchCmd = &cobra.Command{
	Use:   "fetch",
	Short: "Download remote changes without merging",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := gitops.Fetch(".", "origin"); err != nil {
			return err
		}
		fmt.Println("fetched from origin")
		return nil
	},
}

var pullCmd = &cobra.Command{
	Use:   "pull",
	Short: "Fetch and integrate remote changes (use --rebase for linear history)",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := gitops.Pull(".", pullRebase); err != nil {
			return err
		}
		if pullRebase {
			fmt.Println("pulled (rebased)")
		} else {
			fmt.Println("pulled (merged)")
		}
		return nil
	},
}
