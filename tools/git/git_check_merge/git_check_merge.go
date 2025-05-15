package git_check_merge

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/xhd2015/kool/pkgs/errs"
)

// MergeStatus represents the merge status of a reference relative to HEAD
type MergeStatus struct {
	// IsMerged indicates if the ref is merged into HEAD
	IsMerged bool

	// MergeCommit is the commit where the ref is merged (if IsMerged is true)
	MergeCommit string

	// CommitCount is the number of commits ahead of HEAD (if IsMerged is false)
	CommitCount int
}

// Handle processes the git check-merge command
func Handle(args []string) error {
	// Parse flags
	var dir string
	var refs []string
	n := len(args)
	for i := 0; i < n; i++ {
		arg := args[i]
		if !strings.HasPrefix(arg, "-") {
			refs = append(refs, arg)
			continue
		}
		if arg == "--dir" {
			if i+1 >= n {
				return fmt.Errorf("--dir requires a directory path")
			}
			dir = args[i+1]
			i++ // Skip the next argument since it's the directory
			continue
		} else if strings.HasPrefix(arg, "--dir=") {
			dir = strings.TrimPrefix(arg, "--dir=")
			continue
		}
		return fmt.Errorf("unrecognized %s", arg)
	}

	if len(refs) == 0 {
		return fmt.Errorf("usage: kool git check-merge <refs...>")
	}

	exitCode := 0
	for _, ref := range refs {
		status, err := checkMergeStatus(dir, ref)
		if err != nil {
			return fmt.Errorf("failed to check merge status for %s: %w", ref, err)
		}

		if status.IsMerged {
			fmt.Printf("%s is merged into HEAD at %s\n", ref, status.MergeCommit)
		} else {
			fmt.Printf("%s has %d commits ahead of HEAD\n", ref, status.CommitCount)
			exitCode = 1
		}
	}

	if exitCode != 0 {
		// Return a custom error to signal non-zero exit code
		return errs.NewSilenceExitCode(exitCode)
	}
	return nil
}

// checkMergeStatus checks if a ref is merged into HEAD
// Returns:
// - MergeStatus: struct containing merge status information
// - error: any error that occurred
func checkMergeStatus(dir string, ref string) (*MergeStatus, error) {
	// Normalize the ref
	normalizeCmd := exec.Command("git", "rev-parse", ref)
	normalizeCmd.Dir = dir
	normalizedBytes, err := normalizeCmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to normalize ref: %w", err)
	}
	normalizedRef := strings.TrimSpace(string(normalizedBytes))

	// Check if the ref is an ancestor of HEAD
	// git merge-base --is-ancestor <ref> HEAD
	isAncestorCmd := exec.Command("git", "merge-base", "--is-ancestor", normalizedRef, "HEAD")
	isAncestorCmd.Dir = dir
	if err := isAncestorCmd.Run(); err != nil {
		// Not an ancestor, count commits ahead
		revListCmd := exec.Command("git", "rev-list", "--count", "HEAD.."+normalizedRef)
		revListCmd.Dir = dir
		countBytes, countErr := revListCmd.Output()
		if countErr != nil {
			return nil, fmt.Errorf("failed to count commits: %w", countErr)
		}
		count := strings.TrimSpace(string(countBytes))
		countNum := 0
		fmt.Sscanf(count, "%d", &countNum)
		return &MergeStatus{
			IsMerged:    false,
			CommitCount: countNum,
		}, nil
	}

	// If we get here, the ref is an ancestor of HEAD
	// Find the merge commit
	// git log --first-parent --grep="Merge.*<ref>" --pretty=format:%H
	mergeCommitCmd := exec.Command("git", "log", "--first-parent", "--grep=Merge.*"+normalizedRef, "--pretty=format:%H", "-n", "1")
	mergeCommitCmd.Dir = dir
	mergeCommitBytes, err := mergeCommitCmd.Output()
	if err != nil {
		// If we can't find a merge commit with this pattern, try another approach
		// Look for a commit that has the ref as one of its parents
		mergeCommitCmd = exec.Command("git", "log", "--all", "--merges", "--pretty=format:%H", "-n", "1", normalizedRef+"...HEAD")
		mergeCommitCmd.Dir = dir
		mergeCommitBytes, err = mergeCommitCmd.Output()
		if err != nil {
			// If we still can't find it, just use the ref itself
			return &MergeStatus{
				IsMerged:    true,
				MergeCommit: normalizedRef,
			}, nil
		}
	}

	mergeCommit := strings.TrimSpace(string(mergeCommitBytes))
	if mergeCommit == "" {
		// If we still can't find a merge commit, just use the ref itself
		// This can happen if the ref was directly applied to the main branch
		return &MergeStatus{
			IsMerged:    true,
			MergeCommit: normalizedRef,
		}, nil
	}

	return &MergeStatus{
		IsMerged:    true,
		MergeCommit: mergeCommit,
	}, nil
}
