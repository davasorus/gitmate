package main

import (
	"context"
	"fmt"

	"github.com/davasorus/gitmate/internal/ghapi"
	"github.com/davasorus/gitmate/internal/gitops"
	"github.com/spf13/cobra"
)

var (
	initRepoName string
	initDesc     string
	initPrivate  bool
)

func init() {
	initRemoteCmd.Flags().StringVar(&initRepoName, "name", "", "repository name to create (required)")
	initRemoteCmd.Flags().StringVar(&initDesc, "description", "", "repository description")
	initRemoteCmd.Flags().BoolVar(&initPrivate, "private", false, "create as a private repo")
	initRemoteCmd.MarkFlagRequired("name")
	rootCmd.AddCommand(initRemoteCmd)
}

var initRemoteCmd = &cobra.Command{
	Use:   "init-remote",
	Short: "Create a GitHub repo and add it as the origin remote",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		client, err := ghapi.New(ctx, "", "")
		if err != nil {
			return err
		}

		cloneURL, err := client.CreateRepo(ctx, initRepoName, initDesc, initPrivate)
		if err != nil {
			return err
		}
		fmt.Printf("Created %s\n", cloneURL)

		if err := gitops.AddRemote(".", "origin", cloneURL); err != nil {
			return fmt.Errorf("repo created, but adding remote failed: %w", err)
		}
		fmt.Println("Added as origin remote")
		fmt.Println("Now push with: git push -u origin live")
		return nil
	},
}
