package main

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var reviewBody string

func init() {
	reviewCmd.Flags().StringVarP(&reviewBody, "body", "b", "", "review comment body")
	prCmd.AddCommand(reviewCmd, reviewsListCmd, reviewersCmd)
}

var reviewCmd = &cobra.Command{
	Use:   "review <number> <approve|request-changes|comment>",
	Short: "Submit a review on a pull request",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		n, err := argToPRNumber(args)
		if err != nil {
			return err
		}
		var event string
		switch strings.ToLower(args[1]) {
		case "approve":
			event = "APPROVE"
		case "request-changes":
			event = "REQUEST_CHANGES"
		case "comment":
			event = "COMMENT"
		default:
			return fmt.Errorf("verdict must be approve, request-changes, or comment")
		}
		if event != "APPROVE" && reviewBody == "" {
			return fmt.Errorf("--body is required for %s", args[1])
		}
		client, ctx, owner, repo, err := ghClient()
		if err != nil {
			return err
		}
		if err := client.SubmitReview(ctx, owner, repo, n, event, reviewBody); err != nil {
			return err
		}
		fmt.Printf("submitted %s review on #%d\n", args[1], n)
		return nil
	},
}

var reviewsListCmd = &cobra.Command{
	Use:   "reviews <number>",
	Short: "List reviews on a pull request",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		n, err := argToPRNumber(args)
		if err != nil {
			return err
		}
		client, ctx, owner, repo, err := ghClient()
		if err != nil {
			return err
		}
		rs, err := client.ListReviews(ctx, owner, repo, n)
		if err != nil {
			return err
		}
		if len(rs) == 0 {
			fmt.Println("no reviews")
			return nil
		}
		for _, r := range rs {
			fmt.Printf("%-18s %s\n", r.Author, r.State)
		}
		return nil
	},
}

var reviewersCmd = &cobra.Command{
	Use:   "reviewers <number> [add|remove <login>]",
	Short: "List, request, or remove PR reviewers",
	Args:  cobra.RangeArgs(1, 3),
	RunE: func(cmd *cobra.Command, args []string) error {
		n, err := argToPRNumber(args)
		if err != nil {
			return err
		}
		client, ctx, owner, repo, err := ghClient()
		if err != nil {
			return err
		}
		if len(args) == 1 {
			rs, err := client.ListRequestedReviewers(ctx, owner, repo, n)
			if err != nil {
				return err
			}
			if len(rs) == 0 {
				fmt.Println("no requested reviewers")
				return nil
			}
			for _, r := range rs {
				fmt.Println(r.Login)
			}
			return nil
		}
		if len(args) != 3 {
			return fmt.Errorf("usage: reviewers <number> add|remove <login>")
		}
		switch args[1] {
		case "add":
			if err := client.RequestReviewers(ctx, owner, repo, n, []string{args[2]}); err != nil {
				return err
			}
			fmt.Printf("requested review from %s on #%d\n", args[2], n)
		case "remove":
			if err := client.RemoveReviewer(ctx, owner, repo, n, args[2]); err != nil {
				return err
			}
			fmt.Printf("removed %s from #%d\n", args[2], n)
		default:
			return fmt.Errorf("second arg must be add or remove")
		}
		return nil
	},
}
