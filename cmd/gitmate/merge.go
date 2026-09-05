package main

import (
	"fmt"

	"github.com/davasorus/gitmate/internal/gitops"
	"github.com/spf13/cobra"
)

func init() { rootCmd.AddCommand(mergeCmd, mergeAbortCmd) }

var mergeCmd = &cobra.Command{
	Use:   "merge <branch>",
	Short: "Merge a branch into the current branch",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := gitops.Merge(".", args[0]); err != nil {
			files, ferr := gitops.ConflictedFiles(".")
			if ferr == nil && len(files) > 0 {
				fmt.Printf("merge of %s hit conflicts in %d file(s):\n", args[0], len(files))
				for _, f := range files {
					fmt.Printf("  %s\n", f)
				}
				fmt.Println("resolve them and commit, or run `gitmate merge-abort`")
				return nil
			}
			return err
		}
		fmt.Printf("merged %s\n", args[0])
		return nil
	},
}

var mergeAbortCmd = &cobra.Command{
	Use:   "merge-abort",
	Short: "Abort an in-progress merge",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := gitops.MergeAbort("."); err != nil {
			return err
		}
		fmt.Println("merge aborted")
		return nil
	},
}
