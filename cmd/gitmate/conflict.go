package main

import (
	"fmt"

	"github.com/davasorus/gitmate/internal/gitops"
	"github.com/spf13/cobra"
)

func init() { rootCmd.AddCommand(conflictsCmd, resolveCmd) }

var conflictsCmd = &cobra.Command{
	Use:   "conflicts",
	Short: "List files with merge conflicts",
	RunE: func(cmd *cobra.Command, args []string) error {
		files, err := gitops.ConflictedFiles(".")
		if err != nil {
			return err
		}
		if len(files) == 0 {
			fmt.Println("no conflicts")
			return nil
		}
		for _, f := range files {
			fmt.Println(" ", f)
		}
		return nil
	},
}

var resolveSideFlag string

func init() {
	resolveCmd.Flags().StringVar(&resolveSideFlag, "side", "", "resolve taking 'ours' or 'theirs' (omit to just mark a hand-edited file resolved)")
}

var resolveCmd = &cobra.Command{
	Use:   "resolve <path>",
	Short: "Resolve a conflicted file (--side ours|theirs, or mark hand-edited as resolved)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		path := args[0]
		var err error
		switch resolveSideFlag {
		case "ours":
			err = gitops.ResolveOurs(".", path)
		case "theirs":
			err = gitops.ResolveTheirs(".", path)
		case "":
			err = gitops.MarkResolved(".", path)
		default:
			return fmt.Errorf("--side must be 'ours' or 'theirs'")
		}
		if err != nil {
			return err
		}
		fmt.Printf("resolved %s\n", path)
		return nil
	},
}
