package main

import (
	"fmt"

	"github.com/davasorus/gitmate/internal/gitops"
	"github.com/spf13/cobra"
)

var (
	diffStaged bool
	diffRev    string
)

func init() {
	diffCmd.Flags().BoolVar(&diffStaged, "staged", false, "show staged changes (index vs HEAD)")
	diffCmd.Flags().StringVar(&diffRev, "rev", "", "diff a revision or range (e.g. HEAD~1 or a..b)")
	rootCmd.AddCommand(diffCmd)
}

const (
	ansiReset = "\033[0m"
	ansiGreen = "\033[32m"
	ansiRed   = "\033[31m"
	ansiCyan  = "\033[36m"
	ansiDim   = "\033[2m"
)

var diffCmd = &cobra.Command{
	Use:   "diff [path]",
	Short: "Show changes (working tree, staged, or a revision)",
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := gitops.DiffOptions{Staged: diffStaged, Rev: diffRev}
		if len(args) > 0 {
			opts.Path = args[0]
		}
		files, err := gitops.Diff(".", opts)
		if err != nil {
			return err
		}
		if len(files) == 0 {
			fmt.Println("no changes")
			return nil
		}
		for _, f := range files {
			path := f.NewPath
			if path == "" {
				path = f.OldPath
			}
			fmt.Printf("%s%s%s\n", ansiCyan, path, ansiReset)
			if f.Binary {
				fmt.Printf("%s  (binary file)%s\n", ansiDim, ansiReset)
				continue
			}
			for _, h := range f.Hunks {
				fmt.Printf("%s%s%s\n", ansiDim, h.Header, ansiReset)
				for _, ln := range h.Lines {
					switch ln.Kind {
					case gitops.LineAdd:
						fmt.Printf("%s+%s%s\n", ansiGreen, ln.Content, ansiReset)
					case gitops.LineRemove:
						fmt.Printf("%s-%s%s\n", ansiRed, ln.Content, ansiReset)
					default:
						fmt.Printf(" %s\n", ln.Content)
					}
				}
			}
		}
		return nil
	},
}
