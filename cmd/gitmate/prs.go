package main

import (
	"context"
	"fmt"

	"github.com/davasorus/gitmate/internal/ghapi"
	"github.com/spf13/cobra"
)

var (
	prRepo  string
	prState string
)

func init() {
	prsCmd.Flags().StringVar(&prRepo, "repo", "", "target repo as owner/name (defaults to origin remote)")
	prsCmd.Flags().StringVar(&prState, "state", "open", "open, closed, or all")
	rootCmd.AddCommand(prsCmd)
}

var prsCmd = &cobra.Command{
	Use:   "prs",
	Short: "List pull requests for a repo",
	RunE: func(cmd *cobra.Command, args []string) error {
		owner, repo, err := resolveRepo(prRepo)
		if err != nil {
			return err
		}
		ctx := context.Background()
		client, err := ghapi.New(ctx, owner, repo)
		if err != nil {
			return err
		}
		prs, err := client.ListPRs(ctx, owner, repo, prState)
		if err != nil {
			return err
		}
		if len(prs) == 0 {
			fmt.Printf("No %s PRs in %s/%s\n", prState, owner, repo)
			return nil
		}
		for _, p := range prs {
			draft := ""
			if p.Draft {
				draft = " [draft]"
			}
			fmt.Printf("#%-5d %-8s %s%s  (@%s)\n", p.Number, p.State, p.Title, draft, p.Author)
		}
		return nil
	},
}
