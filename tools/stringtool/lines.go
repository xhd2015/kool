package stringtool

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/xhd2015/kool/pkgs/jsondecode"
	"github.com/xhd2015/kool/pkgs/terminal"
	"github.com/xhd2015/less-gen/flags"
	"github.com/xhd2015/xgo/support/cmd"
	"golang.org/x/term"
)

func HandleLines(args []string) error {
	if len(args) > 0 && args[0] == "diff" {
		return handleDiff(args[1:])
	}
	isTTY := term.IsTerminal(int(os.Stdin.Fd()))

	var actions []string
	n := len(args)
	i := 0
	for ; i < n; i++ {
		ok := true
		switch args[i] {
		case "sort", "reverse", "uniq":
			actions = append(actions, args[i])
		default:
			ok = false
		}
		if !ok {
			break
		}
	}

	var useArgs bool
	var inputLines []string
	if !isTTY {
		// not tty, try read from stdin
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			return err
		}
		if len(data) == 0 {
			useArgs = true
		} else {
			inputLines = strings.Split(string(data), "\n")
		}
	} else {
		useArgs = true
	}
	if useArgs {
		for _, arg := range args[i:] {
			argLines := strings.Split(arg, "\n")
			inputLines = append(inputLines, argLines...)
		}
	}
	lines := trimSpace(inputLines)
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		// trim trailing empty line
		lines = lines[:len(lines)-1]
	}
	for _, action := range actions {
		switch action {
		case "sort":
			lines = sortLines(lines)
		case "reverse":
			lines = Reverse(lines)
		case "uniq":
			lines = UniqTail(lines)
		default:
			return fmt.Errorf("unknown line operation: %s", action)
		}
	}
	for _, line := range lines {
		fmt.Println(line)
	}
	return nil
}
func handleDiff(args []string) error {
	var jsonFlag bool

	var wordDiff bool
	args, err := flags.Bool("--json", &jsonFlag).
		Bool("--word-diff", &wordDiff).
		Parse(args)
	if err != nil {
		return err
	}

	content, err := terminal.ReadOrTerminalDataOrFile(args)
	if err != nil {
		return err
	}
	content = strings.TrimSuffix(content, "\n")
	lines := strings.Split(content, "\n")
	if len(lines) != 2 {
		return fmt.Errorf("requires 2 lines to compare, given: %d", len(lines))
	}

	line1 := lines[0]
	line2 := lines[1]
	if jsonFlag {
		line1 = jsondecode.MustPrettyString(line1)
		line2 = jsondecode.MustPrettyString(line2)
	}

	diff, err := diff(line1, line2, wordDiff)
	if err != nil {
		return err
	}
	fmt.Println(diff)
	return nil
}

type SortType int

const (
	SortTypeNone SortType = iota
	SortTypeAsc
	SortTypeDesc
)

func trimSpace(lines []string) []string {
	trimmedLines := make([]string, len(lines))
	for i, line := range lines {
		trimmedLines[i] = strings.TrimSpace(line)
	}
	return trimmedLines
}

func UniqTail(lines []string) []string {
	mapping := make(map[string]int, len(lines))
	n := len(lines)
	uniqLines := make([]string, 0, len(lines))
	for i := n - 1; i >= 0; i-- {
		line := lines[i]
		if _, ok := mapping[line]; ok {
			continue
		}
		mapping[line] = i
		uniqLines = append(uniqLines, line)
	}
	return Reverse(uniqLines)
}

func Uniq(lines []string) []string {
	mapping := make(map[string]bool, len(lines))
	uniqLines := make([]string, 0, len(lines))
	for _, line := range lines {
		if mapping[line] {
			continue
		}
		mapping[line] = true
		uniqLines = append(uniqLines, line)
	}
	return uniqLines
}

func sortLines(lines []string) []string {
	sortedLines := make([]string, len(lines))
	copy(sortedLines, lines)
	sort.Strings(sortedLines)
	return sortedLines
}

func Reverse(lines []string) []string {
	n := len(lines)
	reversedLines := make([]string, n)
	for i, line := range lines {
		reversedLines[n-1-i] = line
	}
	return reversedLines
}

func diff(line1 string, line2 string, wordDiff bool) (string, error) {
	tmpDir, err := os.MkdirTemp("", "stringtool-diff")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(tmpDir)

	err = os.WriteFile(filepath.Join(tmpDir, "line1.txt"), []byte(line1), 0644)
	if err != nil {
		return "", err
	}
	err = os.WriteFile(filepath.Join(tmpDir, "line2.txt"), []byte(line2), 0644)
	if err != nil {
		return "", err
	}

	cmdFlags := []string{
		"diff",
		"--no-index",
	}
	if wordDiff {
		cmdFlags = append(cmdFlags, "--word-diff")
	}
	cmdFlags = append(cmdFlags, "line1.txt", "line2.txt")
	diffOutput, err := cmd.Dir(tmpDir).Output("git", cmdFlags...)
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			if exitErr.ExitCode() == 1 {
				return diffOutput, nil
			}
		}
		return "", err
	}
	return diffOutput, nil
}
