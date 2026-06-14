package compare_branch

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

func Handle(args []string) error {
	var dir string
	var refs []string
	for i := 0; i < len(args); i++ {
		if args[i] == "-C" {
			if i+1 >= len(args) {
				return fmt.Errorf("-C requires a directory argument")
			}
			dir = args[i+1]
			i++
		} else if !strings.HasPrefix(args[i], "-") {
			refs = append(refs, args[i])
		} else {
			return fmt.Errorf("unknown flag: %s", args[i])
		}
	}
	if len(refs) < 2 {
		return fmt.Errorf("usage: kool git compare-branch <refA> <refB> [-C <dir>]")
	}
	if len(refs) > 2 {
		return fmt.Errorf("unexpected arguments: %v", refs[2:])
	}
	refA, refB := refs[0], refs[1]

	revA, err := revParse(dir, refA)
	if err != nil {
		return err
	}
	revB, err := revParse(dir, refB)
	if err != nil {
		return err
	}

	if revA == revB {
		fmt.Printf("%s and %s are identical\n", refA, refB)
		return nil
	}

	aIsAncestorOfB, err := isAncestor(dir, refA, refB)
	if err != nil {
		return err
	}
	bIsAncestorOfA, err := isAncestor(dir, refB, refA)
	if err != nil {
		return err
	}

	if aIsAncestorOfB {
		count, err := revListCount(dir, refA, refB)
		if err != nil {
			return err
		}
		commitWord := "commit"
		if count != 1 {
			commitWord = "commits"
		}
		fmt.Printf("%s is newer(%s +%d %s -> %s)\n", refB, refA, count, commitWord, refB)
		fmt.Printf("to fast forward, on %s: \n   git merge --ff-only  %s\n", refA, refB)
		return nil
	}

	if bIsAncestorOfA {
		count, err := revListCount(dir, refB, refA)
		if err != nil {
			return err
		}
		commitWord := "commit"
		if count != 1 {
			commitWord = "commits"
		}
		fmt.Printf("%s is newer(%s +%d %s -> %s)\n", refA, refB, count, commitWord, refA)
		fmt.Printf("to fast forward, on %s: \n   git merge --ff-only  %s\n", refB, refA)
		return nil
	}

	base, err := mergeBase(dir, refA, refB)
	if err != nil {
		return err
	}

	fileCount, err := diffFileCount(dir, refA, refB)
	if err != nil {
		return err
	}

	countA, err := revListCount(dir, refB, refA)
	if err != nil {
		return err
	}
	countB, err := revListCount(dir, refA, refB)
	if err != nil {
		return err
	}

	commitWordA := "commit"
	if countA > 1 {
		commitWordA = "commits"
	}
	commitWordB := "commit"
	if countB > 1 {
		commitWordB = "commits"
	}

	fmt.Printf("%s and %s has %d files difference\n", refA, refB, fileCount)
	fmt.Printf("their most recent base is %s\n", base)
	fmt.Printf("%s has %d unique %s\n", refA, countA, commitWordA)
	fmt.Printf("%s has %d unique %s\n", refB, countB, commitWordB)
	fmt.Println("They need to be merged")
	return nil
}

func revParse(dir, ref string) (string, error) {
	cmd := exec.Command("git", "rev-parse", "--verify", ref)
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			gitErr := strings.TrimSpace(string(exitErr.Stderr))
			if gitErr != "" {
				return "", fmt.Errorf("failed to resolve '%s': %s", ref, gitErr)
			}
		}
		return "", fmt.Errorf("failed to resolve '%s': %v", ref, err)
	}
	return strings.TrimSpace(string(out)), nil
}

func isAncestor(dir, ancestor, descendant string) (bool, error) {
	cmd := exec.Command("git", "merge-base", "--is-ancestor", ancestor, descendant)
	cmd.Dir = dir
	err := cmd.Run()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			if exitErr.ExitCode() == 1 {
				return false, nil
			}
		}
		return false, fmt.Errorf("failed to check if %s is ancestor of %s: %w", ancestor, descendant, err)
	}
	return true, nil
}

func mergeBase(dir, refA, refB string) (string, error) {
	cmd := exec.Command("git", "merge-base", refA, refB)
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			gitErr := strings.TrimSpace(string(exitErr.Stderr))
			if gitErr != "" {
				return "", fmt.Errorf("%s", gitErr)
			}
		}
		return "", fmt.Errorf("failed to find merge base: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

func diffFileCount(dir, refA, refB string) (int, error) {
	cmd := exec.Command("git", "diff", "--name-only", refA, refB)
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("failed to diff: %w", err)
	}
	trimmed := strings.TrimSpace(string(out))
	if trimmed == "" {
		return 0, nil
	}
	return len(strings.Split(trimmed, "\n")), nil
}

func revListCount(dir, from, to string) (int, error) {
	cmd := exec.Command("git", "rev-list", "--count", from+".."+to)
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("failed to count commits: %w", err)
	}
	count, err := strconv.Atoi(strings.TrimSpace(string(out)))
	if err != nil {
		return 0, fmt.Errorf("failed to parse commit count: %w", err)
	}
	return count, nil
}
