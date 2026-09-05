package main

import (
	"fmt"

	"github.com/davasorus/gitmate/internal/gitops"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(statusCmd)
}

const (
	stReset = "\033[0m"
	stGreen = "\033[32m"
	stRed   = "\033[31m"
	stDim   = "\033[2m"
)

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

		// Split into staged vs unstaged. A file can appear in both (e.g. staged
		// then edited again), so it may show in each section — matching git.
		var staged, unstaged []gitops.FileChange
		for _, c := range s.Changes {
			if c.Staged != "" {
				staged = append(staged, c)
			}
			if c.Unstaged != "" {
				unstaged = append(unstaged, c)
			}
		}

		if len(staged) == 0 && len(unstaged) == 0 && len(s.Untracked) == 0 {
			fmt.Println("working tree clean")
			return nil
		}

		if len(staged) > 0 {
			fmt.Printf("\n%sStaged:%s\n", stDim, stReset)
			for _, c := range staged {
				printChange(stGreen, c.Staged, c)
			}
		}
		if len(unstaged) > 0 {
			fmt.Printf("\n%sUnstaged:%s\n", stDim, stReset)
			for _, c := range unstaged {
				printChange(stRed, c.Unstaged, c)
			}
		}
		if len(s.Untracked) > 0 {
			fmt.Printf("\n%sUntracked:%s\n", stDim, stReset)
			for _, u := range s.Untracked {
				fmt.Printf("  %s%-10s%s %s\n", stRed, "new", stReset, u)
			}
		}
		return nil
	},
}

func printChange(color, state string, c gitops.FileChange) {
	path := c.Path
	if c.OrigPath != "" {
		path = c.OrigPath + " → " + c.Path
	}
	fmt.Printf("  %s%-10s%s %s\n", color, state, stReset, path)
}
