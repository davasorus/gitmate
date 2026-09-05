package main

import (
	"context"
	"fmt"

	"github.com/davasorus/gitmate/internal/ghapi"
	"github.com/davasorus/gitmate/internal/gitops"
	"github.com/spf13/cobra"
)

func init() {
	prCreateCmd.Flags().StringVar(&prRepoFlag, "repo", "", "target repo as owner/name (defaults to origin remote)")
	prCreateCmd.Flags().StringVarP(&prTitle, "title", "t", "", "PR title (required)")
	prCreateCmd.Flags().StringVarP(&prBody, "body", "b", "", "PR description")
	prCreateCmd.Flags().StringVar(&prHead, "head", "", "source branch (required)")
	prCreateCmd.Flags().StringVar(&prBase, "base", "live", "target branch")

	issueCreateCmd.Flags().StringVar(&issueRepoFlag, "repo", "", "target repo as owner/name (defaults to origin remote)")
	issueCreateCmd.Flags().StringVarP(&issueTitle, "title", "t", "", "issue title (required)")
	issueCreateCmd.Flags().StringVarP(&issueBody, "body", "b", "", "issue description")
	_ = issueCreateCmd.MarkFlagRequired("title")

	prCmd.AddCommand(prCreateCmd)
	issueCmd.AddCommand(issueCreateCmd)
	rootCmd.AddCommand(prCmd, issueCmd)
}

// Parent grouping commands so we get `gitmate pr create` / `gitmate issue create`.
var prCmd = &cobra.Command{Use: "pr", Short: "Work with pull requests"}
var issueCmd = &cobra.Command{Use: "issue", Short: "Work with issues"}

var (
	prRepoFlag, prTitle, prBody, prHead, prBase string
)

var prCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Open a pull request",
	RunE: func(cmd *cobra.Command, args []string) error {
		owner, repo, err := resolveRepo(prRepoFlag)
		if err != nil {
			return err
		}
		head := prHead
		if head == "" {
			head, err = gitops.CurrentBranch(".")
			if err != nil {
				return err
			}
		}
		title := prTitle
		if title == "" {
			title, _ = gitops.LastCommitSubject(".", head)
		}
		body := prBody
		if body == "" {
			body = gitops.ReadPRTemplate(".")
		}
		ctx := context.Background()
		client, err := ghapi.New(ctx, owner, repo)
		if err != nil {
			return err
		}
		num, url, err := client.CreatePR(ctx, owner, repo, title, body, head, prBase)
		if err != nil {
			return err
		}
		fmt.Printf("opened PR #%d: %s\n", num, url)
		return nil
	},
}

var (
	issueRepoFlag, issueTitle, issueBody string
)

var issueCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Open an issue",
	RunE: func(cmd *cobra.Command, args []string) error {
		owner, repo, err := resolveRepo(issueRepoFlag)
		if err != nil {
			return err
		}
		ctx := context.Background()
		client, err := ghapi.New(ctx, owner, repo)
		if err != nil {
			return err
		}
		num, url, err := client.CreateIssue(ctx, owner, repo, issueTitle, issueBody)
		if err != nil {
			return err
		}
		fmt.Printf("opened issue #%d: %s\n", num, url)
		return nil
	},
}
