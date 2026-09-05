package main

import (
	"fmt"

	"github.com/davasorus/gitmate/internal/gitops"
	"github.com/spf13/cobra"
)

var (
	stashMessage   string
	stashUntracked bool
	stashDropForce bool
)

func init() {
	stashSaveCmd.Flags().StringVarP(&stashMessage, "message", "m", "", "stash description")
	stashSaveCmd.Flags().BoolVarP(&stashUntracked, "include-untracked", "u", false, "also stash untracked files")
	stashDropCmd.Flags().BoolVarP(&stashDropForce, "force", "f", false, "confirm dropping the stash (destructive)")
	stashCmd.AddCommand(stashSaveCmd, stashListCmd, stashPopCmd, stashDropCmd)
	rootCmd.AddCommand(stashCmd)
}

var stashCmd = &cobra.Command{
	Use:   "stash",
	Short: "Save, list, pop, or drop stashed changes",
	// bare `gitmate stash` lists
	RunE: func(cmd *cobra.Command, args []string) error {
		return runStashList()
	},
}

var stashSaveCmd = &cobra.Command{
	Use:   "save",
	Short: "Stash current changes",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := gitops.StashSave(".", stashMessage, stashUntracked); err != nil {
			return err
		}
		fmt.Println("stashed changes")
		return nil
	},
}

var stashListCmd = &cobra.Command{
	Use:   "list",
	Short: "List stashes",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runStashList()
	},
}

func runStashList() error {
	stashes, err := gitops.StashList(".")
	if err != nil {
		return err
	}
	if len(stashes) == 0 {
		fmt.Println("no stashes")
		return nil
	}
	for _, s := range stashes {
		fmt.Printf("%-12s %-16s %s\n", s.Ref, s.Branch, s.Message)
	}
	return nil
}

var stashPopCmd = &cobra.Command{
	Use:   "pop [ref]",
	Short: "Apply and remove a stash (default: newest)",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ref := ""
		if len(args) == 1 {
			ref = args[0]
		}
		if err := gitops.StashPop(".", ref); err != nil {
			return err
		}
		fmt.Println("popped stash")
		return nil
	},
}

var stashDropCmd = &cobra.Command{
	Use:   "drop [ref]",
	Short: "Discard a stash without applying (destructive)",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if !stashDropForce {
			return fmt.Errorf("refusing to drop without --force: this permanently discards the stashed changes")
		}
		ref := ""
		if len(args) == 1 {
			ref = args[0]
		}
		if err := gitops.StashDrop(".", ref); err != nil {
			return err
		}
		fmt.Println("dropped stash")
		return nil
	},
}
