package main

import (
	"fmt"

	"github.com/davasorus/gitmate/internal/gitops"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(stageCmd, commitCmd, pushCmd, unstageCmd, discardCmd)
}

var stageCmd = &cobra.Command{
	Use:   "stage [paths...]",
	Short: "Stage changes (all if no paths given)",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := gitops.Stage(".", args...); err != nil {
			return err
		}
		if len(args) == 0 {
			fmt.Println("staged all changes")
		} else {
			fmt.Printf("staged %d path(s)\n", len(args))
		}
		return nil
	},
}

var commitMessage string

var commitCmd = &cobra.Command{
	Use:   "commit",
	Short: "Create a commit from staged changes",
	RunE: func(cmd *cobra.Command, args []string) error {
		short, err := gitops.CreateCommit(".", commitMessage)
		if err != nil {
			return err
		}
		fmt.Printf("created commit %s\n", short)
		return nil
	},
}

var (
	pushRemote      string
	pushSetUpstream bool
)

var pushCmd = &cobra.Command{
	Use:   "push",
	Short: "Push the current branch to a remote",
	RunE: func(cmd *cobra.Command, args []string) error {
		branch, err := gitops.CurrentBranch(".")
		if err != nil {
			return err
		}
		if err := gitops.Push(".", pushRemote, branch, pushSetUpstream); err != nil {
			return err
		}
		fmt.Printf("pushed %s to %s\n", branch, pushRemote)
		return nil
	},
}

func init() {
	commitCmd.Flags().StringVarP(&commitMessage, "message", "m", "", "commit message (required)")
	_ = commitCmd.MarkFlagRequired("message")
	pushCmd.Flags().StringVar(&pushRemote, "remote", "origin", "remote to push to")
	pushCmd.Flags().BoolVarP(&pushSetUpstream, "set-upstream", "u", false, "set upstream tracking on push")
}

var unstageCmd = &cobra.Command{
	Use:   "unstage [paths...]",
	Short: "Unstage changes (all if no paths given)",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := gitops.Unstage(".", args...); err != nil {
			return err
		}
		if len(args) == 0 {
			fmt.Println("unstaged all changes")
		} else {
			fmt.Printf("unstaged %d path(s)\n", len(args))
		}
		return nil
	},
}

var discardForce bool

func init() {
	discardCmd.Flags().BoolVarP(&discardForce, "force", "f", false, "confirm discarding changes (required — this is destructive)")
}

var discardCmd = &cobra.Command{
	Use:   "discard <paths...>",
	Short: "Discard working-tree changes to tracked files (destructive)",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if !discardForce {
			return fmt.Errorf("refusing to discard without --force: this permanently deletes uncommitted changes to %v", args)
		}
		if err := gitops.Discard(".", args...); err != nil {
			return err
		}
		fmt.Printf("discarded changes to %d path(s)\n", len(args))
		return nil
	},
}
