package main

import (
	"fmt"

	"github.com/davasorus/gitmate/internal/gitops"
	"github.com/spf13/cobra"
)

var resetSoft, resetHard, resetForce bool

func init() {
	resetCmd.Flags().BoolVar(&resetSoft, "soft", false, "move HEAD only, keep index and working tree")
	resetCmd.Flags().BoolVar(&resetHard, "hard", false, "move HEAD and DISCARD index + working tree changes (destructive)")
	resetCmd.Flags().BoolVarP(&resetForce, "force", "f", false, "required to confirm a --hard reset")
	rootCmd.AddCommand(resetCmd)
}

var resetCmd = &cobra.Command{
	Use:   "reset <rev>",
	Short: "Move HEAD to a revision (default mixed; --soft / --hard)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		mode := gitops.ResetMixed
		switch {
		case resetHard:
			mode = gitops.ResetHard
			if !resetForce {
				return fmt.Errorf("refusing --hard reset without --force: this discards uncommitted changes and orphans commits after %s (recoverable via reflog)", args[0])
			}
		case resetSoft:
			mode = gitops.ResetSoft
		}
		if err := gitops.Reset(".", args[0], mode); err != nil {
			return err
		}
		fmt.Printf("reset (%s) to %s\n", mode, args[0])
		return nil
	},
}
