package worktree

import (
	"fmt"
	"os"
	"strings"

	wt "github.com/xhd2015/dot-pkgs/go-pkgs/git/worktree"
)

func mergeBack(args []string) error {
	var to string
	var dryRun bool
	var remove bool
	var confirmFromStdin bool

	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch arg {
		case "--to":
			if i+1 >= len(args) {
				return fmt.Errorf("--to requires a path argument")
			}
			i++
			to = args[i]
		case "--dry-run":
			dryRun = true
		case "--rm":
			remove = true
		case "--confirm-from-stdin":
			confirmFromStdin = true
		default:
			if strings.HasPrefix(arg, "-") {
				return fmt.Errorf("unknown flag: %s", arg)
			}
			return fmt.Errorf("unexpected argument: %s", arg)
		}
	}

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get working directory: %w", err)
	}

	result, err := wt.MergeBack(wt.MergeBackOptions{
		SourcePath: cwd,
		TargetPath: to,
		DryRun:     dryRun,
		Remove:     remove,
		Confirm: func(plan wt.MergeBackPlan) (bool, error) {
			return wt.PromptConfirmPlan(plan, confirmFromStdin, false)
		},
	})
	if err != nil {
		return err
	}

	if result.Message != "" {
		fmt.Println(result.Message)
	}
	return nil
}