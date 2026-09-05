package main

import (
	"fmt"

	"github.com/davasorus/gitmate/internal/gitops"
	"github.com/spf13/cobra"
)

func init() {
	tagCreateCmd.Flags().StringVarP(&tagMessage, "message", "m", "", "annotation message (omit for a lightweight tag)")
	tagCmd.AddCommand(tagListCmd, tagCreateCmd, tagDeleteCmd, tagPushCmd)
	rootCmd.AddCommand(tagCmd)
}

var tagMessage string

var tagCmd = &cobra.Command{
	Use:   "tag",
	Short: "Manage tags (list, create, delete, push)",
	RunE:  func(cmd *cobra.Command, args []string) error { return runTagList() },
}

var tagListCmd = &cobra.Command{
	Use:   "list",
	Short: "List tags",
	RunE:  func(cmd *cobra.Command, args []string) error { return runTagList() },
}

func runTagList() error {
	tags, err := gitops.ListTags(".")
	if err != nil {
		return err
	}
	if len(tags) == 0 {
		fmt.Println("no tags")
		return nil
	}
	for _, t := range tags {
		fmt.Printf("%-20s %s\n", t.Name, t.Subject)
	}
	return nil
}

var tagCreateCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Create a tag at HEAD (use -m for an annotated tag)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := gitops.CreateTag(".", args[0], tagMessage); err != nil {
			return err
		}
		fmt.Printf("created tag %s\n", args[0])
		return nil
	},
}

var tagDeleteCmd = &cobra.Command{
	Use:   "delete <name>",
	Short: "Delete a local tag",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := gitops.DeleteTag(".", args[0]); err != nil {
			return err
		}
		fmt.Printf("deleted tag %s\n", args[0])
		return nil
	},
}

var tagPushCmd = &cobra.Command{
	Use:   "push <name>",
	Short: "Push a tag to origin (triggers tag-based release workflows)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := gitops.PushTag(".", args[0]); err != nil {
			return err
		}
		fmt.Printf("pushed tag %s to origin\n", args[0])
		return nil
	},
}
