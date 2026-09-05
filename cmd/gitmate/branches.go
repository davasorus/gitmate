package main

import (
	"fmt"

	"github.com/davasorus/gitmate/internal/gitops"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(branchesCmd)
}

var branchesCmd = &cobra.Command{
	Use:     "branches",
	Aliases: []string{"branch", "br"},
	Short:   "List local branches",
	RunE: func(cmd *cobra.Command, args []string) error {
		branches, err := gitops.GetBranches(".")
		if err != nil {
			return err
		}
		for _, b := range branches {
			marker := " "
			if b.IsCurrent {
				marker = "*"
			}

			track := ""
			switch {
			case b.Upstream == "":
				track = "(no upstream)"
			case b.Ahead > 0 && b.Behind > 0:
				track = fmt.Sprintf("[ahead %d, behind %d]", b.Ahead, b.Behind)
			case b.Ahead > 0:
				track = fmt.Sprintf("[ahead %d]", b.Ahead)
			case b.Behind > 0:
				track = fmt.Sprintf("[behind %d]", b.Behind)
			}

			fmt.Printf("%s %-20s %-8s %s %s\n", marker, b.Name, b.LastHash, track, b.LastSubject)
		}
		return nil
	},
}
