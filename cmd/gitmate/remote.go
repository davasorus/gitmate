package main

import (
	"fmt"

	"github.com/davasorus/gitmate/internal/gitops"
	"github.com/spf13/cobra"
)

func init() {
	remoteCmd.AddCommand(remoteListCmd, remoteAddCmd, remoteRemoveCmd, remoteRenameCmd)
	rootCmd.AddCommand(remoteCmd)
}

var remoteCmd = &cobra.Command{
	Use:   "remote",
	Short: "Manage remotes (list, add, remove, rename)",
	RunE:  func(cmd *cobra.Command, args []string) error { return runRemoteList() },
}

var remoteListCmd = &cobra.Command{
	Use:   "list",
	Short: "List remotes",
	RunE:  func(cmd *cobra.Command, args []string) error { return runRemoteList() },
}

func runRemoteList() error {
	rs, err := gitops.ListRemotes(".")
	if err != nil {
		return err
	}
	if len(rs) == 0 {
		fmt.Println("no remotes")
		return nil
	}
	for _, r := range rs {
		fmt.Printf("%-12s %s\n", r.Name, r.URL)
	}
	return nil
}

var remoteAddCmd = &cobra.Command{
	Use:   "add <name> <url>",
	Short: "Add a remote",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := gitops.AddRemote(".", args[0], args[1]); err != nil {
			return err
		}
		fmt.Printf("added remote %s -> %s\n", args[0], args[1])
		return nil
	},
}

var remoteRemoveCmd = &cobra.Command{
	Use:   "remove <name>",
	Short: "Remove a remote",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := gitops.RemoveRemote(".", args[0]); err != nil {
			return err
		}
		fmt.Printf("removed remote %s\n", args[0])
		return nil
	},
}

var remoteRenameCmd = &cobra.Command{
	Use:   "rename <old> <new>",
	Short: "Rename a remote",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := gitops.RenameRemote(".", args[0], args[1]); err != nil {
			return err
		}
		fmt.Printf("renamed remote %s -> %s\n", args[0], args[1])
		return nil
	},
}
