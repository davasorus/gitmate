package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/davasorus/gitmate/internal/ghapi"
	"github.com/spf13/cobra"
)

func init() {
	labelCreateCmd.Flags().StringVar(&labelColor, "color", "cccccc", "label color (6-hex, no #)")
	labelCreateCmd.Flags().StringVar(&labelDesc, "desc", "", "label description")
	labelEditCmd.Flags().StringVar(&labelNewName, "name", "", "new label name (default: keep)")
	labelEditCmd.Flags().StringVar(&labelColor, "color", "", "new color (6-hex, no #)")
	labelEditCmd.Flags().StringVar(&labelDesc, "desc", "", "new description")
	labelCmd.AddCommand(labelListCmd, labelCreateCmd, labelEditCmd, labelDeleteCmd, labelAddCmd, labelRemoveCmd)
	rootCmd.AddCommand(labelCmd)
}

var (
	labelColor   string
	labelDesc    string
	labelNewName string
)

func ghClient() (*ghapi.Client, context.Context, string, string, error) {
	owner, repo, err := resolveRepo("")
	if err != nil {
		return nil, nil, "", "", err
	}
	ctx := context.Background()
	client, err := ghapi.New(ctx, owner, repo)
	return client, ctx, owner, repo, err
}

var labelCmd = &cobra.Command{
	Use:   "label",
	Short: "Manage labels (definitions + apply to issues/PRs)",
	RunE:  func(cmd *cobra.Command, args []string) error { return runLabelList() },
}

var labelListCmd = &cobra.Command{
	Use:   "list",
	Short: "List repo label definitions",
	RunE:  func(cmd *cobra.Command, args []string) error { return runLabelList() },
}

func runLabelList() error {
	client, ctx, owner, repo, err := ghClient()
	if err != nil {
		return err
	}
	labels, err := client.ListLabels(ctx, owner, repo)
	if err != nil {
		return err
	}
	if len(labels) == 0 {
		fmt.Println("no labels")
		return nil
	}
	for _, l := range labels {
		fmt.Printf("%-24s #%-7s %s\n", l.Name, l.Color, l.Description)
	}
	return nil
}

var labelCreateCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Create a label definition",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, ctx, owner, repo, err := ghClient()
		if err != nil {
			return err
		}
		if err := client.CreateLabel(ctx, owner, repo, args[0], labelColor, labelDesc); err != nil {
			return err
		}
		fmt.Printf("created label %s\n", args[0])
		return nil
	},
}

var labelEditCmd = &cobra.Command{
	Use:   "edit <name>",
	Short: "Edit a label definition",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, ctx, owner, repo, err := ghClient()
		if err != nil {
			return err
		}
		newName := labelNewName
		if newName == "" {
			newName = args[0]
		}
		if err := client.EditLabel(ctx, owner, repo, args[0], newName, labelColor, labelDesc); err != nil {
			return err
		}
		fmt.Printf("edited label %s\n", args[0])
		return nil
	},
}

var labelDeleteCmd = &cobra.Command{
	Use:   "delete <name>",
	Short: "Delete a label definition",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, ctx, owner, repo, err := ghClient()
		if err != nil {
			return err
		}
		if err := client.DeleteLabel(ctx, owner, repo, args[0]); err != nil {
			return err
		}
		fmt.Printf("deleted label %s\n", args[0])
		return nil
	},
}

var labelAddCmd = &cobra.Command{
	Use:   "add <number> <label[,label...]>",
	Short: "Add labels to an issue or PR",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		n, err := argToPRNumber(args)
		if err != nil {
			return err
		}
		client, ctx, owner, repo, err := ghClient()
		if err != nil {
			return err
		}
		labels := strings.Split(args[1], ",")
		if err := client.AddLabels(ctx, owner, repo, n, labels); err != nil {
			return err
		}
		fmt.Printf("added labels to #%d\n", n)
		return nil
	},
}

var labelRemoveCmd = &cobra.Command{
	Use:   "remove <number> <label>",
	Short: "Remove a label from an issue or PR",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		n, err := argToPRNumber(args)
		if err != nil {
			return err
		}
		client, ctx, owner, repo, err := ghClient()
		if err != nil {
			return err
		}
		if err := client.RemoveLabel(ctx, owner, repo, n, args[1]); err != nil {
			return err
		}
		fmt.Printf("removed label %s from #%d\n", args[1], n)
		return nil
	},
}
