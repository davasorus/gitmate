package main

import (
	"context"
	"fmt"
	"strconv"

	"github.com/davasorus/gitmate/internal/ghapi"
	"github.com/spf13/cobra"
)

var (
	mergeRepoFlag   string
	mergeMethodFlag string
	commentRepoFlag string
	commentBodyFlag string
	checksRepoFlag  string
)

func init() {
	prMergeCmd.Flags().StringVar(&mergeRepoFlag, "repo", "", "target repo as owner/name (defaults to origin remote)")
	prMergeCmd.Flags().StringVar(&mergeMethodFlag, "method", "merge", "merge, squash, or rebase")

	prCommentCmd.Flags().StringVar(&commentRepoFlag, "repo", "", "target repo as owner/name (defaults to origin remote)")
	prCommentCmd.Flags().StringVarP(&commentBodyFlag, "body", "b", "", "comment body (required)")
	_ = prCommentCmd.MarkFlagRequired("body")

	prChecksCmd.Flags().StringVar(&checksRepoFlag, "repo", "", "target repo as owner/name (defaults to origin remote)")

	prCmd.AddCommand(prMergeCmd, prCommentCmd, prChecksCmd)
}

// argToPRNumber parses the first positional arg as a PR number.
func argToPRNumber(args []string) (int, error) {
	if len(args) < 1 {
		return 0, fmt.Errorf("a PR number is required")
	}
	n, err := strconv.Atoi(args[0])
	if err != nil {
		return 0, fmt.Errorf("invalid PR number %q", args[0])
	}
	return n, nil
}

var prMergeCmd = &cobra.Command{
	Use:   "merge <number>",
	Short: "Merge a pull request",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		num, err := argToPRNumber(args)
		if err != nil {
			return err
		}
		owner, repo, err := resolveRepo(mergeRepoFlag)
		if err != nil {
			return err
		}
		ctx := context.Background()
		client, err := ghapi.New(ctx, owner, repo)
		if err != nil {
			return err
		}
		sha, err := client.MergePR(ctx, owner, repo, num, mergeMethodFlag)
		if err != nil {
			return err
		}
		fmt.Printf("merged PR #%d (%s) via %s\n", num, sha[:min(7, len(sha))], mergeMethodFlag)
		return nil
	},
}

var prCommentCmd = &cobra.Command{
	Use:   "comment <number>",
	Short: "Comment on a pull request",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		num, err := argToPRNumber(args)
		if err != nil {
			return err
		}
		owner, repo, err := resolveRepo(commentRepoFlag)
		if err != nil {
			return err
		}
		ctx := context.Background()
		client, err := ghapi.New(ctx, owner, repo)
		if err != nil {
			return err
		}
		url, err := client.CommentPR(ctx, owner, repo, num, commentBodyFlag)
		if err != nil {
			return err
		}
		fmt.Printf("commented on #%d: %s\n", num, url)
		return nil
	},
}

var prChecksCmd = &cobra.Command{
	Use:   "checks <number>",
	Short: "Show CI check status for a pull request",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		num, err := argToPRNumber(args)
		if err != nil {
			return err
		}
		owner, repo, err := resolveRepo(checksRepoFlag)
		if err != nil {
			return err
		}
		ctx := context.Background()
		client, err := ghapi.New(ctx, owner, repo)
		if err != nil {
			return err
		}
		runs, err := client.PRChecks(ctx, owner, repo, num)
		if err != nil {
			return err
		}
		if len(runs) == 0 {
			fmt.Printf("no checks reported for #%d\n", num)
			return nil
		}
		for _, r := range runs {
			state := r.Status
			if r.Conclusion != "" {
				state = r.Conclusion
			}
			fmt.Printf("  %-12s %s\n", state, r.Name)
		}
		return nil
	},
}
