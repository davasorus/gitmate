package main

import (
	"context"
	"fmt"

	"github.com/davasorus/gitmate/internal/ghapi"
	"github.com/spf13/cobra"
)

func init() {
	prCmd.AddCommand(prCloseCmd, prReopenCmd)
	issueCmd.AddCommand(issueCloseCmd, issueReopenCmd)
}

func setPRState(number int, state string) error {
	owner, repo, err := resolveRepo("")
	if err != nil {
		return err
	}
	ctx := context.Background()
	client, err := ghapi.New(ctx, owner, repo)
	if err != nil {
		return err
	}
	return client.SetPRState(ctx, owner, repo, number, state)
}

func setIssueState(number int, state string) error {
	owner, repo, err := resolveRepo("")
	if err != nil {
		return err
	}
	ctx := context.Background()
	client, err := ghapi.New(ctx, owner, repo)
	if err != nil {
		return err
	}
	return client.SetIssueState(ctx, owner, repo, number, state)
}

var prCloseCmd = &cobra.Command{
	Use:   "close <number>",
	Short: "Close a pull request",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		n, err := argToPRNumber(args)
		if err != nil {
			return err
		}
		if err := setPRState(n, "closed"); err != nil {
			return err
		}
		fmt.Printf("closed PR #%d\n", n)
		return nil
	},
}

var prReopenCmd = &cobra.Command{
	Use:   "reopen <number>",
	Short: "Reopen a pull request",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		n, err := argToPRNumber(args)
		if err != nil {
			return err
		}
		if err := setPRState(n, "open"); err != nil {
			return err
		}
		fmt.Printf("reopened PR #%d\n", n)
		return nil
	},
}

var issueCloseCmd = &cobra.Command{
	Use:   "close <number>",
	Short: "Close an issue",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		n, err := argToPRNumber(args)
		if err != nil {
			return err
		}
		if err := setIssueState(n, "closed"); err != nil {
			return err
		}
		fmt.Printf("closed issue #%d\n", n)
		return nil
	},
}

var issueReopenCmd = &cobra.Command{
	Use:   "reopen <number>",
	Short: "Reopen an issue",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		n, err := argToPRNumber(args)
		if err != nil {
			return err
		}
		if err := setIssueState(n, "open"); err != nil {
			return err
		}
		fmt.Printf("reopened issue #%d\n", n)
		return nil
	},
}
