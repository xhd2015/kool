package compare_branch

import (
	"fmt"
	"strings"

	"github.com/xhd2015/gitops/git"
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

	result, err := git.CompareBranches(dir, refA, refB)
	if err != nil {
		return err
	}

	switch result.Relation {
	case git.BranchRelationSame:
		fmt.Printf("%s and %s are identical\n", refA, refB)
		return nil

	case git.BranchRelationAIsAncestorOfB:
		count := result.CommitsAheadB
		commitWord := "commit"
		if count != 1 {
			commitWord = "commits"
		}
		fmt.Printf("%s is newer(%s +%d %s -> %s)\n", refB, refA, count, commitWord, refB)
		fmt.Printf("to fast forward, on %s: \n   git merge --ff-only  %s\n", refA, refB)
		return nil

	case git.BranchRelationBIsAncestorOfA:
		count := result.CommitsAheadA
		commitWord := "commit"
		if count != 1 {
			commitWord = "commits"
		}
		fmt.Printf("%s is newer(%s +%d %s -> %s)\n", refA, refB, count, commitWord, refA)
		fmt.Printf("to fast forward, on %s: \n   git merge --ff-only  %s\n", refB, refA)
		return nil

	case git.BranchRelationDiverged:
		commitWordA := "commit"
		if result.CommitsAheadA > 1 {
			commitWordA = "commits"
		}
		commitWordB := "commit"
		if result.CommitsAheadB > 1 {
			commitWordB = "commits"
		}
		fmt.Printf("%s and %s has %d files difference\n", refA, refB, result.DiffFileCount)
		fmt.Printf("their most recent base is %s\n", result.MergeBase)
		fmt.Printf("%s has %d unique %s\n", refA, result.CommitsAheadA, commitWordA)
		fmt.Printf("%s has %d unique %s\n", refB, result.CommitsAheadB, commitWordB)
		fmt.Println("They need to be merged")
		return nil
	}

	return nil
}
