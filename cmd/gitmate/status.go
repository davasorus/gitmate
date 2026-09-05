package main

import (
	"fmt"

	"github.com/davasorus/gitmate/internal/gitops"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(statusCmd)
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show working tree status",
	RunE: func(cmd *cobra.Command, args []string) error {
		s, err := gitops.GetStatus(".")
		if err != nil {
			return err
		}

		if s.Detached {
			fmt.Println("HEAD detached")
		} else {
			line := "On branch " + s.Branch
			if s.Upstream != "" {
				line += fmt.Sprintf(" (tracking %s, ahead %d, behind %d)", s.Upstream, s.Ahead, s.Behind)
			}
			fmt.Println(line)
		}

		if len(s.Changes) == 0 && len(s.Untracked) == 0 {
			fmt.Println("working tree clean")
			return nil
		}

		for _, c := range s.Changes {
			label := c.Staged
			if c.Unstaged != "" {
				if label != "" {
					label += "/"
				}
				label += c.Unstaged
			}
			if c.OrigPath != "" {
				fmt.Printf("  %-18s %s → %s\n", label, c.OrigPath, c.Path)
			} else {
				fmt.Printf("  %-18s %s\n", label, c.Path)
			}
		}
		for _, u := range s.Untracked {
			fmt.Printf("  %-18s %s\n", "untracked", u)
		}
		return nil
	},
}
