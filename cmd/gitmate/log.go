package main

import (
	"fmt"
	"time"

	"github.com/davasorus/gitmate/internal/gitops"
	"github.com/spf13/cobra"
)

var logLimit int

func init() {
	logCmd.Flags().IntVarP(&logLimit, "number", "n", 10, "limit number of commits")
	rootCmd.AddCommand(logCmd)
}

var logCmd = &cobra.Command{
	Use:   "log",
	Short: "Show commit history",
	RunE: func(cmd *cobra.Command, args []string) error {
		commits, err := gitops.GetLog(".", logLimit)
		if err != nil {
			return err
		}
		for _, c := range commits {
			fmt.Printf("%s  %s  %s\n", c.Short, relTime(c.When), c.Author)
			fmt.Printf("        %s\n", c.Subject)
		}
		return nil
	},
}

// relTime renders a coarse "N ago" like git's %ar, but computed in Go
// so you see how the timestamp round-trips.
func relTime(t time.Time) string {
	d := time.Since(t)
	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		return fmt.Sprintf("%dm ago", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh ago", int(d.Hours()))
	default:
		return fmt.Sprintf("%dd ago", int(d.Hours()/24))
	}
}

var _ = time.Now // keep import if you trim relTime later
