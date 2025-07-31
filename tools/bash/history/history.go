package history

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/xhd2015/less-gen/flags"
)

const help = `
kool bash history is a tool to manage your bash history.

Commands:
  merge <files>       merge history files into one
  clean               clean history file
  del <cmd>           delete a command from history

Options:
  -w        write back to history file
  -o        output file

Examples:
  kool bash history merge <files> -w
  kool bash history clean -w
`

func Handle(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("requires command: merge, clean, del")
	}

	cmd := args[0]
	args = args[1:]

	switch cmd {
	case "merge":
		return handleMerge(args)
	case "clean":
		return handleClean(args)
	case "del":
		return handleDel(args)
	}

	return fmt.Errorf("unknown command: %s", cmd)
}

func handleMerge(args []string) error {
	var writeBack bool
	args, err := flags.Bool("-w", &writeBack).Help("-h,--help", help).Parse(args)
	if err != nil {
		return err
	}

	if len(args) == 0 {
		return fmt.Errorf("requires input files")
	}

	homeHistory, err := getHomeHistory()
	if err != nil {
		return err
	}

	lines, err := readLines(homeHistory)
	if err != nil {
		return err
	}

	for _, arg := range args {
		fileLines, err := readLines(arg)
		if err != nil {
			return err
		}

		lines = append(lines, fileLines...)
	}

	lines = cleanLines(lines)

	if !writeBack {
		for _, line := range lines {
			fmt.Println(line)
		}
		return nil
	}

	err = os.WriteFile(homeHistory, []byte(strings.Join(lines, "\n")), 0644)
	if err != nil {
		return err
	}

	return nil
}

func handleDel(args []string) error {
	args, err := flags.Help("-h,--help", help).Parse(args)
	if err != nil {
		return err
	}

	if len(args) == 0 {
		return fmt.Errorf("requires command, usage: del '<command>'")
	}

	delCmd := args[0]
	args = args[1:]

	if len(args) > 0 {
		return fmt.Errorf("unrecognized extra args: %v", args)
	}

	homeHistory, err := getHomeHistory()
	if err != nil {
		return err
	}

	lines, err := readLines(homeHistory)
	if err != nil {
		return err
	}

	var removedDel bool
	cleanedLines := make([]string, 0, len(lines))
	for i := len(lines) - 1; i >= 0; i-- {
		line := lines[i]

		trimLine := strings.TrimSpace(line)
		if !removedDel && strings.HasPrefix(trimLine, "kool bash history del ") {
			removedDel = true
			continue
		}

		if trimLine == delCmd {
			continue
		}

		cleanedLines = append(cleanedLines, line)
	}

	err = os.WriteFile(homeHistory, []byte(strings.Join(cleanedLines, "\n")), 0644)
	if err != nil {
		return err
	}

	return nil
}

func handleClean(args []string) error {
	var writeBack bool
	args, err := flags.Bool("-w", &writeBack).Help("-h,--help", help).Parse(args)
	if err != nil {
		return err
	}

	var historyFile string
	if len(args) > 0 {
		historyFile = args[0]
		args = args[1:]
		if len(args) > 0 {
			return fmt.Errorf("unrecognized extra args: %v", args)
		}
	} else {
		historyFile, err = getHomeHistory()
		if err != nil {
			return err
		}
	}

	lines, err := readLines(historyFile)
	if err != nil {
		return err
	}

	lines = cleanLines(lines)

	if !writeBack {
		for _, line := range lines {
			fmt.Println(line)
		}
		return nil
	}

	err = os.WriteFile(historyFile, []byte(strings.Join(lines, "\n")), 0644)
	if err != nil {
		return fmt.Errorf("failed to write history file: %w", err)
	}

	return nil
}

func getHomeHistory() (string, error) {
	homeHistory, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	homeHistory = filepath.Join(homeHistory, ".bash_history")

	stat, err := os.Stat(homeHistory)
	if err != nil {
		return "", err
	}

	if stat.IsDir() {
		return "", fmt.Errorf("history file is a directory: %s", homeHistory)
	}

	return homeHistory, nil
}
func readLines(path string) ([]string, error) {
	lines, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	return strings.Split(string(lines), "\n"), nil
}

func cleanLines(lines []string) []string {
	cleaned := make([]string, 0, len(lines))
	seen := make(map[string]bool, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || line == "exit" || line == "ls" || line == "pwd" || strings.HasPrefix(line, "kool bash history del ") {
			continue
		}
		if seen[line] {
			continue
		}
		seen[line] = true
		cleaned = append(cleaned, line)
	}

	return cleaned
}
