package worktree

import (
	"fmt"
	"os"
	"strings"

	wt "github.com/xhd2015/dot-pkgs/go-pkgs/git/worktree"
)

func reclaim(args []string) error {
	var all bool
	var dryRun bool
	var path string

	for _, arg := range args {
		switch arg {
		case "--all":
			all = true
		case "--dry-run":
			dryRun = true
		default:
			if strings.HasPrefix(arg, "-") {
				return fmt.Errorf("unknown flag: %s", arg)
			}
			if path != "" {
				return fmt.Errorf("unexpected extra arguments: %s", arg)
			}
			path = arg
		}
	}

	if all && path != "" {
		return fmt.Errorf("cannot use --all with a worktree path")
	}
	if !all && path == "" {
		return fmt.Errorf("requires <worktree-dir> or --all")
	}

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get working directory: %w", err)
	}

	_, err = wt.Reclaim(wt.ReclaimOptions{
		Cwd:    cwd,
		Path:   path,
		All:    all,
		DryRun: dryRun,
		OnOutcome: func(outcome wt.ReclaimOutcome) {
			printOutcome(outcome)
			_ = os.Stdout.Sync()
		},
	})
	return err
}

func printOutcome(outcome wt.ReclaimOutcome) {
	deadSuffix := ""
	if outcome.Reason == "dead" {
		deadSuffix = " (dead)"
	}
	switch outcome.Action {
	case wt.ActionReclaimed:
		fmt.Printf("reclaimed: %s%s\n", outcome.Path, deadSuffix)
	case wt.ActionDryRun:
		fmt.Printf("dry-run: would reclaim %s%s\n", outcome.Path, deadSuffix)
	case wt.ActionSkipped:
		fmt.Printf("skipped: %s (%s)\n", outcome.Path, outcome.Reason)
	case wt.ActionError:
		if outcome.Reason != "" {
			fmt.Fprintf(os.Stderr, "error: %s: %s\n", outcome.Path, outcome.Reason)
			_ = os.Stderr.Sync()
		}
	}
}