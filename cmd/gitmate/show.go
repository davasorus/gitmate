package main

import (
	"fmt"

	"github.com/davasorus/gitmate/internal/gitops"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(showCmd)
}

var showCmd = &cobra.Command{
	Use:   "show [rev]",
	Short: "Show a commit's metadata and diff (default: HEAD)",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		rev := "HEAD"
		if len(args) == 1 {
			rev = args[0]
		}
		d, err := gitops.Show(".", rev)
		if err != nil {
			return err
		}
		fmt.Printf("%scommit %s%s\n", ansiCyan, d.Hash, ansiReset)
		fmt.Printf("Author: %s <%s>\n", d.Author, d.Email)
		fmt.Printf("Date:   %s\n", d.Date)
		fmt.Printf("\n    %s\n", d.Subject)
		if d.Body != "" {
			fmt.Printf("\n    %s\n", d.Body)
		}
		fmt.Println()
		for _, f := range d.Files {
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
