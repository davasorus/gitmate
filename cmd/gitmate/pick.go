package main

import (
	"fmt"

	"github.com/davasorus/gitmate/internal/gitops"
	"github.com/spf13/cobra"
)

func init() { rootCmd.AddCommand(cherryPickCmd, revertCmd) }

func reportSeqConflicts(op, rev string) error {
	files, _ := gitops.ConflictedFiles(".")
	if len(files) > 0 {
		fmt.Printf("%s of %s hit conflicts in %d file(s):\n", op, rev, len(files))
		for _, f := range files {
			fmt.Printf("  %s\n", f)
		}
		fmt.Printf("resolve them, then `gitmate %s --continue` (or --abort)\n", op)
		return nil
	}
	return fmt.Errorf("%s failed", op)
}

var cpContinue, cpAbort bool
var rvContinue, rvAbort bool

func init() {
	cherryPickCmd.Flags().BoolVar(&cpContinue, "continue", false, "resume after resolving conflicts")
	cherryPickCmd.Flags().BoolVar(&cpAbort, "abort", false, "abort the cherry-pick")
	revertCmd.Flags().BoolVar(&rvContinue, "continue", false, "resume after resolving conflicts")
	revertCmd.Flags().BoolVar(&rvAbort, "abort", false, "abort the revert")
}

var cherryPickCmd = &cobra.Command{
	Use:   "cherry-pick [rev]",
	Short: "Apply a specific commit onto the current branch",
	RunE: func(cmd *cobra.Command, args []string) error {
		switch {
		case cpAbort:
			if err := gitops.CherryPickAbort("."); err != nil {
				return err
			}
			fmt.Println("cherry-pick aborted")
			return nil
		case cpContinue:
			if err := gitops.CherryPickContinue("."); err != nil {
				return reportSeqConflicts("cherry-pick", "")
			}
			fmt.Println("cherry-pick continued")
			return nil
		}
		if len(args) != 1 {
			return fmt.Errorf("a commit is required (or --continue/--abort)")
		}
		if err := gitops.CherryPick(".", args[0]); err != nil {
			return reportSeqConflicts("cherry-pick", args[0])
		}
		fmt.Printf("cherry-picked %s\n", args[0])
		return nil
	},
}

var revertCmd = &cobra.Command{
	Use:   "revert [rev]",
	Short: "Create a new commit that undoes a previous commit",
	RunE: func(cmd *cobra.Command, args []string) error {
		switch {
		case rvAbort:
			if err := gitops.RevertAbort("."); err != nil {
				return err
			}
			fmt.Println("revert aborted")
			return nil
		case rvContinue:
			if err := gitops.RevertContinue("."); err != nil {
				return reportSeqConflicts("revert", "")
			}
			fmt.Println("revert continued")
			return nil
		}
		if len(args) != 1 {
			return fmt.Errorf("a commit is required (or --continue/--abort)")
		}
		if err := gitops.Revert(".", args[0]); err != nil {
			return reportSeqConflicts("revert", args[0])
		}
		fmt.Printf("reverted %s\n", args[0])
		return nil
	},
}
