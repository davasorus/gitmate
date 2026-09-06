package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var (
	relName       string
	relBody       string
	relDraft      bool
	relPrerelease bool
	relGenNotes   bool
)

func init() {
	releaseCreateCmd.Flags().StringVar(&relName, "name", "", "release title")
	releaseCreateCmd.Flags().StringVar(&relBody, "body", "", "release notes")
	releaseCreateCmd.Flags().BoolVar(&relDraft, "draft", false, "create as draft")
	releaseCreateCmd.Flags().BoolVar(&relPrerelease, "prerelease", false, "mark as prerelease")
	releaseCreateCmd.Flags().BoolVar(&relGenNotes, "generate-notes", false, "auto-generate notes from PRs/commits")
	releaseCmd.AddCommand(releaseListCmd, releaseCreateCmd, releaseDeleteCmd, releaseNotesCmd, assetsCmd)
	rootCmd.AddCommand(releaseCmd)
}

var releaseCmd = &cobra.Command{
	Use:   "release",
	Short: "Manage GitHub releases",
	RunE:  func(cmd *cobra.Command, args []string) error { return runReleaseList() },
}

var releaseListCmd = &cobra.Command{
	Use:   "list",
	Short: "List releases",
	RunE:  func(cmd *cobra.Command, args []string) error { return runReleaseList() },
}

func runReleaseList() error {
	client, ctx, owner, repo, err := ghClient()
	if err != nil {
		return err
	}
	rels, err := client.ListReleases(ctx, owner, repo)
	if err != nil {
		return err
	}
	if len(rels) == 0 {
		fmt.Println("no releases")
		return nil
	}
	for _, r := range rels {
		flags := ""
		if r.Draft {
			flags += " [draft]"
		}
		if r.Prerelease {
			flags += " [prerelease]"
		}
		fmt.Printf("%-16s %s%s\n", r.TagName, r.Name, flags)
	}
	return nil
}

var releaseCreateCmd = &cobra.Command{
	Use:   "create <tag>",
	Short: "Create a release on a tag",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, ctx, owner, repo, err := ghClient()
		if err != nil {
			return err
		}
		name, body := relName, relBody
		if relGenNotes {
			gn, gb, gerr := client.GenerateReleaseNotes(ctx, owner, repo, args[0])
			if gerr != nil {
				return gerr
			}
			if name == "" {
				name = gn
			}
			if body == "" {
				body = gb
			}
		}
		r, err := client.CreateRelease(ctx, owner, repo, args[0], name, body, relDraft, relPrerelease)
		if err != nil {
			return err
		}
		fmt.Printf("created release %s: %s\n", r.TagName, r.URL)
		return nil
	},
}

var releaseDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a release by ID",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var id int64
		if _, err := fmt.Sscan(args[0], &id); err != nil {
			return fmt.Errorf("invalid release id %q", args[0])
		}
		client, ctx, owner, repo, err := ghClient()
		if err != nil {
			return err
		}
		if err := client.DeleteRelease(ctx, owner, repo, id); err != nil {
			return err
		}
		fmt.Printf("deleted release %d\n", id)
		return nil
	},
}

var releaseNotesCmd = &cobra.Command{
	Use:   "notes <tag>",
	Short: "Preview auto-generated release notes for a tag",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, ctx, owner, repo, err := ghClient()
		if err != nil {
			return err
		}
		name, body, err := client.GenerateReleaseNotes(ctx, owner, repo, args[0])
		if err != nil {
			return err
		}
		fmt.Printf("# %s\n\n%s\n", name, body)
		return nil
	},
}

func parseAssetID(a string) (int64, error) {
	var id int64
	if _, err := fmt.Sscan(a, &id); err != nil {
		return 0, fmt.Errorf("invalid id %q", a)
	}
	return id, nil
}

var assetsCmd = &cobra.Command{
	Use:   "assets",
	Short: "Manage release assets (list/upload/download/delete)",
}

func init() {
	assetListCmd := &cobra.Command{
		Use: "list <releaseID>", Short: "List assets", Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := parseAssetID(args[0])
			if err != nil {
				return err
			}
			client, ctx, owner, repo, err := ghClient()
			if err != nil {
				return err
			}
			as, err := client.ListAssets(ctx, owner, repo, id)
			if err != nil {
				return err
			}
			for _, a := range as {
				fmt.Printf("%d\t%s\t%d bytes\n", a.ID, a.Name, a.Size)
			}
			return nil
		},
	}
	assetUploadCmd := &cobra.Command{
		Use: "upload <releaseID> <file>", Short: "Upload a file as an asset", Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := parseAssetID(args[0])
			if err != nil {
				return err
			}
			data, err := os.ReadFile(args[1])
			if err != nil {
				return err
			}
			client, ctx, owner, repo, err := ghClient()
			if err != nil {
				return err
			}
			a, err := client.UploadAsset(ctx, owner, repo, id, filepath.Base(args[1]), data)
			if err != nil {
				return err
			}
			fmt.Printf("uploaded %s (asset %d)\n", a.Name, a.ID)
			return nil
		},
	}
	assetDownloadCmd := &cobra.Command{
		Use: "download <assetID> <dest>", Short: "Download an asset", Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := parseAssetID(args[0])
			if err != nil {
				return err
			}
			client, ctx, owner, repo, err := ghClient()
			if err != nil {
				return err
			}
			data, err := client.DownloadAsset(ctx, owner, repo, id)
			if err != nil {
				return err
			}
			if err := os.WriteFile(args[1], data, 0o644); err != nil {
				return err
			}
			fmt.Printf("downloaded to %s\n", args[1])
			return nil
		},
	}
	assetDeleteCmd := &cobra.Command{
		Use: "delete <assetID>", Short: "Delete an asset", Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := parseAssetID(args[0])
			if err != nil {
				return err
			}
			client, ctx, owner, repo, err := ghClient()
			if err != nil {
				return err
			}
			if err := client.DeleteAsset(ctx, owner, repo, id); err != nil {
				return err
			}
			fmt.Printf("deleted asset %d\n", id)
			return nil
		},
	}
	assetsCmd.AddCommand(assetListCmd, assetUploadCmd, assetDownloadCmd, assetDeleteCmd)
}
