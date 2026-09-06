package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	actionsCmd.AddCommand(runsListCmd, runViewCmd)
	rootCmd.AddCommand(actionsCmd)
}

var actionsCmd = &cobra.Command{
	Use:   "actions",
	Short: "GitHub Actions — list and inspect workflow runs",
	RunE:  func(cmd *cobra.Command, args []string) error { return runRunsList() },
}

var runsListCmd = &cobra.Command{
	Use:   "runs",
	Short: "List recent workflow runs",
	RunE:  func(cmd *cobra.Command, args []string) error { return runRunsList() },
}

func runRunsList() error {
	client, ctx, owner, repo, err := ghClient()
	if err != nil {
		return err
	}
	runs, err := client.ListRuns(ctx, owner, repo, 30)
	if err != nil {
		return err
	}
	if len(runs) == 0 {
		fmt.Println("no runs")
		return nil
	}
	for _, r := range runs {
		st := r.Status
		if r.Conclusion != "" {
			st = r.Conclusion
		}
		fmt.Printf("%-10d %-12s %-10s %s (%s)\n", r.ID, st, r.Branch, r.Name, r.Event)
	}
	return nil
}

var runViewCmd = &cobra.Command{
	Use:   "view <runID>",
	Short: "Show a run's jobs and steps",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var id int64
		if _, err := fmt.Sscan(args[0], &id); err != nil {
			return fmt.Errorf("invalid run id %q", args[0])
		}
		client, ctx, owner, repo, err := ghClient()
		if err != nil {
			return err
		}
		jobs, err := client.ListRunJobs(ctx, owner, repo, id)
		if err != nil {
			return err
		}
		for _, j := range jobs {
			jst := j.Status
			if j.Conclusion != "" {
				jst = j.Conclusion
			}
			fmt.Printf("● %s [%s]\n", j.Name, jst)
			for _, st := range j.Steps {
				sst := st.Status
				if st.Conclusion != "" {
					sst = st.Conclusion
				}
				fmt.Printf("    %s [%s]\n", st.Name, sst)
			}
		}
		return nil
	},
}
