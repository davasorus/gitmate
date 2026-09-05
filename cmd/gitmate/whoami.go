package main

import (
	"context"
	"fmt"

	"github.com/davasorus/gitmate/internal/ghapi"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(whoamiCmd)
}

var whoamiCmd = &cobra.Command{
	Use:   "whoami",
	Short: "Verify GitHub auth and show rate limit",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		// owner/repo don't matter for these two calls — pass your handle.
		client, err := ghapi.New(ctx, "davasorus", "gitmate")
		if err != nil {
			return err
		}
		login, err := client.Whoami(ctx)
		if err != nil {
			return err
		}
		remaining, limit, err := client.RateLimit(ctx)
		if err != nil {
			return err
		}
		fmt.Printf("Authenticated as: %s\n", login)
		fmt.Printf("Rate limit: %d/%d core requests remaining\n", remaining, limit)
		return nil
	},
}
