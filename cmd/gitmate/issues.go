package main

import (
	"context"
	"fmt"

	"github.com/davasorus/gitmate/internal/ghapi"
	"github.com/spf13/cobra"
)

var (
	issueRepo  string
	issueState string
)

func init() {
	issuesCmd.Flags().StringVar(&issueRepo, "repo", "", "target repo as owner/name (defaults to origin remote)")
	issuesCmd.Flags().StringVar(&issueState, "state", "open", "open, closed, or all")
	rootCmd.AddCommand(issuesCmd)
}

var issuesCmd = &cobra.Command{
	Use:   "issues",
	Short: "List issues for a repo",
	RunE: func(cmd *cobra.Command, args []string) error {
		owner, repo, err := resolveRepo(issueRepo)
		if err != nil {
			return err
		}
		ctx := context.Background()
		client, err := ghapi.New(ctx, owner, repo)
		if err != nil {
			return err
		}
		issues, err := client.ListIssues(ctx, owner, repo, issueState)
		if err != nil {
			return err
		}
		if len(issues) == 0 {
			fmt.Printf("No %s issues in %s/%s\n", issueState, owner, repo)
			return nil
		}
		for _, i := range issues {
			fmt.Printf("#%-5d %-8s %s  (@%s)\n", i.Number, i.State, i.Title, i.Author)
		}
		return nil
	},
}
