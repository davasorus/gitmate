package main

import (
	"fmt"

	"github.com/davasorus/gitmate/internal/gitops"
	"github.com/spf13/cobra"
)

var branchDeleteForce bool

func init() {
	branchDeleteCmd.Flags().BoolVarP(&branchDeleteForce, "force", "f", false, "force delete even if the branch has unmerged commits (-D)")
	branchCmd.AddCommand(branchListCmd, branchDeleteCmd, branchRenameCmd)
	rootCmd.AddCommand(branchCmd)
}

var branchCmd = &cobra.Command{
	Use:     "branch",
	Aliases: []string{"branches", "br"},
	Short:   "Manage branches (list, delete, rename)",
	// bare `gitmate branch` lists, like git
	RunE: func(cmd *cobra.Command, args []string) error {
		return runBranchList()
	},
}

var branchListCmd = &cobra.Command{
	Use:   "list",
	Short: "List local branches",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runBranchList()
	},
}

func runBranchList() error {
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
}

var branchDeleteCmd = &cobra.Command{
	Use:   "delete <name>",
	Short: "Delete a branch (safe -d; use --force for -D)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := gitops.DeleteBranch(".", args[0], branchDeleteForce); err != nil {
			return err
		}
		fmt.Printf("deleted branch %s\n", args[0])
		return nil
	},
}

var branchRenameCmd = &cobra.Command{
	Use:   "rename <old> <new>",
	Short: "Rename a branch",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := gitops.RenameBranch(".", args[0], args[1]); err != nil {
			return err
		}
		fmt.Printf("renamed %s → %s\n", args[0], args[1])
		return nil
	},
}
