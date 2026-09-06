package main

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

func init() {
	actionsCmd.AddCommand(runsListCmd, runViewCmd, runCancelCmd, runRerunCmd, dispatchCmd)
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

func parseRunID(a string) (int64, error) {
	var id int64
	if _, err := fmt.Sscan(a, &id); err != nil {
		return 0, fmt.Errorf("invalid run id %q", a)
	}
	return id, nil
}

var runCancelCmd = &cobra.Command{
	Use:   "cancel <runID>",
	Short: "Cancel an in-progress run",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := parseRunID(args[0])
		if err != nil {
			return err
		}
		client, ctx, owner, repo, err := ghClient()
		if err != nil {
			return err
		}
		if err := client.CancelRun(ctx, owner, repo, id); err != nil {
			return err
		}
		fmt.Printf("cancelled run %d\n", id)
		return nil
	},
}

var rerunFailedOnly bool

var runRerunCmd = &cobra.Command{
	Use:   "rerun <runID>",
	Short: "Re-run a completed run (--failed for failed jobs only)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := parseRunID(args[0])
		if err != nil {
			return err
		}
		client, ctx, owner, repo, err := ghClient()
		if err != nil {
			return err
		}
		if rerunFailedOnly {
			if err := client.RerunFailed(ctx, owner, repo, id); err != nil {
				return err
			}
			fmt.Printf("re-running failed jobs of %d\n", id)
		} else {
			if err := client.RerunRun(ctx, owner, repo, id); err != nil {
				return err
			}
			fmt.Printf("re-running %d\n", id)
		}
		return nil
	},
}

var dispatchRef string

var dispatchCmd = &cobra.Command{
	Use:   "dispatch <workflow.yml> [key=value...]",
	Short: "Trigger a workflow_dispatch (with optional inputs)",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, ctx, owner, repo, err := ghClient()
		if err != nil {
			return err
		}
		inputs := map[string]interface{}{}
		for _, kv := range args[1:] {
			parts := strings.SplitN(kv, "=", 2)
			if len(parts) == 2 {
				inputs[parts[0]] = parts[1]
			}
		}
		ref := dispatchRef
		if ref == "" {
			ref = "live"
		}
		if err := client.TriggerDispatch(ctx, owner, repo, args[0], ref, inputs); err != nil {
			return err
		}
		fmt.Printf("dispatched %s on %s\n", args[0], ref)
		return nil
	},
}

func init() {
	runRerunCmd.Flags().BoolVar(&rerunFailedOnly, "failed", false, "re-run only failed jobs")
	dispatchCmd.Flags().StringVar(&dispatchRef, "ref", "", "git ref to run on (default: live)")
}
