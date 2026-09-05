package main

import (
	"fmt"

	"github.com/davasorus/gitmate/internal/gitops"
	"github.com/spf13/cobra"
)

func init() { rootCmd.AddCommand(rebaseCmd) }

var rebaseContinue, rebaseAbort bool

func init() {
	rebaseCmd.Flags().BoolVar(&rebaseContinue, "continue", false, "resume after resolving conflicts")
	rebaseCmd.Flags().BoolVar(&rebaseAbort, "abort", false, "abort the in-progress rebase")
}

var rebaseCmd = &cobra.Command{
	Use:   "rebase [base]",
	Short: "Rebase the current branch onto base (--continue / --abort mid-rebase)",
	RunE: func(cmd *cobra.Command, args []string) error {
		switch {
		case rebaseAbort:
			if err := gitops.RebaseAbort("."); err != nil {
				return err
			}
			fmt.Println("rebase aborted")
			return nil
		case rebaseContinue:
			if err := gitops.RebaseContinue("."); err != nil {
				files, _ := gitops.ConflictedFiles(".")
				if len(files) > 0 {
					fmt.Printf("still conflicts in %d file(s):\n", len(files))
					for _, f := range files {
						fmt.Printf("  %s\n", f)
					}
					return nil
				}
				return err
			}
			fmt.Println("rebase continued")
			return nil
		}
		if len(args) != 1 {
			return fmt.Errorf("a base branch is required (or use --continue/--abort)")
		}
		if err := gitops.Rebase(".", args[0]); err != nil {
			files, ferr := gitops.ConflictedFiles(".")
			if ferr == nil && len(files) > 0 {
				fmt.Printf("rebase onto %s hit conflicts in %d file(s):\n", args[0], len(files))
				for _, f := range files {
					fmt.Printf("  %s\n", f)
				}
				fmt.Println("resolve them, then `gitmate rebase --continue` (or --abort)")
				return nil
			}
			return err
		}
		fmt.Printf("rebased onto %s\n", args[0])
		return nil
	},
}
