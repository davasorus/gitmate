package main

import (
	"fmt"

	"github.com/davasorus/gitmate/internal/gitops"
	"github.com/spf13/cobra"
)

var switchCreate bool

func init() {
	switchCmd.Flags().BoolVarP(&switchCreate, "create", "c", false, "create the branch and switch to it")
	rootCmd.AddCommand(switchCmd)
}

var switchCmd = &cobra.Command{
	Use:   "switch <branch>",
	Short: "Switch to a branch (use -c to create it)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		branch := args[0]
		var err error
		if switchCreate {
			err = gitops.SwitchNew(".", branch)
		} else {
			err = gitops.Switch(".", branch)
		}
		if err != nil {
			return err
		}
		fmt.Printf("switched to %s\n", branch)
		return nil
	},
}
