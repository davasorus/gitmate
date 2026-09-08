package main

import (
	"fmt"

	"github.com/davasorus/gitmate/internal/version"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.Version = version.Resolve()
	rootCmd.SetVersionTemplate("gitmate {{.Version}}\n")
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show the gitmate version",
	Run: func(_ *cobra.Command, _ []string) {
		fmt.Printf("gitmate %s\n", version.Resolve())
	},
}
