package history

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	llsrun "github.com/xhd2015/lls/run"
	"github.com/xhd2015/kool/tools/stringtool"
	"github.com/xhd2015/less-gen/flags"
	"golang.org/x/term"
)

const help = `
kool bash history is a tool to manage your bash history.

Commands:
  merge <files>          merge history files into one
  compact,clean          compact history file
  del <cmd>              delete a command from history
  log-file list          list log files
  log-file add <file>    add a log file
  log-file rm <file>     remove a log file

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
	case "clean", "compact":
		return handleClean(args)
	case "del":
		return HandleDel(args)
	case "log-file":
		return handleLogFile(args)
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

	homeHistory, err := GetHomeHistory()
	if err != nil {
		return err
	}

	lines, err := ReadLines(homeHistory)
	if err != nil {
		return err
	}

	for _, arg := range args {
		fileLines, err := ReadLines(arg)
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

func HandleDel(args []string) error {
	args, err := flags.Help("-h,--help", help).Parse(args)
	if err != nil {
		return err
	}

	isTTY := term.IsTerminal(int(os.Stdin.Fd()))

	var delCmd string
	if len(args) > 0 {
		delCmd = args[0]
		args = args[1:]
		if len(args) > 0 {
			return fmt.Errorf("unrecognized extra args: %v", args)
		}
	}

	homeHistory, err := GetHomeHistory()
	if err != nil {
		return err
	}

	logFiles, err := readLogFiles()
	if err != nil {
		return err
	}

	allFiles := append([]string{homeHistory}, logFiles...)

	if delCmd == "" {
		if !isTTY {
			return fmt.Errorf("requires command, usage: del '<command>'")
		}
		selected, err := selectFromHistory(allFiles, "")
		if err != nil {
			return err
		}
		if selected == "" {
			return nil
		}
		if !confirmDelete(selected) {
			return nil
		}
		delCmd = selected
	} else {
		exists, err := commandExistsInFiles(allFiles, delCmd)
		if err != nil {
			return err
		}
		if !exists && isTTY {
			selected, err := selectFromHistory(allFiles, delCmd)
			if err != nil {
				return err
			}
			if selected == "" {
				return nil
			}
			if !confirmDelete(selected) {
				return nil
			}
			delCmd = selected
		} else if !exists {
			return fmt.Errorf("command not found in history: %s", delCmd)
		}
	}

	err = DeleteFromHistoryFile(homeHistory, delCmd)
	if err != nil {
		return err
	}

	for _, logFile := range logFiles {
		err = DeleteFromHistoryFile(logFile, delCmd)
		if err != nil {
			fmt.Printf("Warning: failed to delete from %s: %v\n", logFile, err)
		}
	}

	return nil
}

func selectFromHistory(files []string, query string) (string, error) {
	seen := make(map[string]bool)
	var commands []string
	for _, file := range files {
		lines, err := ReadNonEmptyLines(file)
		if err != nil {
			continue
		}
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" || seen[line] {
				continue
			}
			seen[line] = true
			commands = append(commands, line)
		}
	}

	if len(commands) == 0 {
		return "", fmt.Errorf("no history found")
	}

	return llsrun.SelectWithFzf(commands, query)
}

func commandExistsInFiles(files []string, cmd string) (bool, error) {
	for _, file := range files {
		lines, err := ReadNonEmptyLines(file)
		if err != nil {
			continue
		}
		for _, line := range lines {
			if strings.TrimSpace(line) == cmd {
				return true, nil
			}
		}
	}
	return false, nil
}

func confirmDelete(cmd string) bool {
	fmt.Fprintf(os.Stderr, "Really delete '%s'? [Y/n] ", cmd)
	reader := bufio.NewReader(os.Stdin)
	answer, err := reader.ReadString('\n')
	if err != nil {
		return false
	}
	answer = strings.TrimSpace(strings.ToLower(answer))
	return answer == "" || answer == "y" || answer == "yes"
}

func DeleteFromHistoryFile(historyFile, delCmd string) error {
	lines, err := ReadLines(historyFile)
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

	return os.WriteFile(historyFile, []byte(strings.Join(cleanedLines, "\n")), 0644)
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
		historyFile, err = GetHomeHistory()
		if err != nil {
			return err
		}
	}

	lines, err := ReadLines(historyFile)
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

func delLineFromFile(file string, line string) error {
	lines, err := ReadLines(file)
	if err != nil {
		return err
	}
	var newLines []string = make([]string, 0, len(lines))
	for _, l := range lines {
		if l == line {
			continue
		}
		newLines = append(newLines, l)
	}
	return os.WriteFile(file, []byte(strings.Join(newLines, "\n")), 0644)
}

// GetAllHistoryFiles returns all history files, including home history and log files
func GetAllHistoryFiles() ([]string, error) {
	homeHistory, err := GetHomeHistory()
	if err != nil {
		return nil, err
	}
	logFiles, err := readLogFiles()
	if err != nil {
		return nil, err
	}
	allFiles := make([]string, 0, len(logFiles)+1)
	allFiles = append(allFiles, homeHistory)
	allFiles = append(allFiles, logFiles...)
	return allFiles, nil
}

func GetHomeHistory() (string, error) {
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

func ReadNonEmptyLines(path string) ([]string, error) {
	lines, err := ReadLines(path)
	if err != nil {
		return nil, err
	}
	nonEmpty := make([]string, 0, len(lines))
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			nonEmpty = append(nonEmpty, line)
		}
	}
	return nonEmpty, nil
}

func ReadLines(path string) ([]string, error) {
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

func handleLogFile(args []string) error {
	args, err := flags.Help("-h,--help", help).Parse(args)
	if err != nil {
		return err
	}

	if len(args) == 0 {
		return fmt.Errorf("requires subcommand: add, list, rm")
	}

	subCmd := args[0]
	args = args[1:]

	switch subCmd {
	case "add":
		return handleLogFileAdd(args)
	case "list":
		return handleLogFileList(args)
	case "rm":
		return handleLogFileRm(args)
	default:
		return fmt.Errorf("unknown subcommand: %s", subCmd)
	}
}

func getLogFilesPath() (string, error) {
	userConfigDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}

	koolDir := filepath.Join(userConfigDir, "kool", "bash", "history")
	err = os.MkdirAll(koolDir, 0755)
	if err != nil {
		return "", err
	}

	return filepath.Join(koolDir, "log-files.txt"), nil
}

func readLogFiles() ([]string, error) {
	configLogFiles, err := readConfigLogFiles()
	if err != nil {
		return nil, err
	}
	// check if ~/.bash_history_log exists
	bashHistoryLog, err := GetHomeHistoryLog()
	if err != nil {
		return nil, err
	}
	if _, err := os.Stat(bashHistoryLog); err == nil {
		configLogFiles = append(configLogFiles, bashHistoryLog)
	}

	// uniq
	configLogFiles = stringtool.Uniq(configLogFiles)
	return configLogFiles, nil
}

func GetHomeHistoryLog() (string, error) {
	homeHistory, err := GetHomeHistory()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeHistory, ".bash_history_log"), nil
}

func readConfigLogFiles() ([]string, error) {
	logFilesPath, err := getLogFilesPath()
	if err != nil {
		return nil, err
	}

	if _, err := os.Stat(logFilesPath); os.IsNotExist(err) {
		return []string{}, nil
	}

	content, err := os.ReadFile(logFilesPath)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(content), "\n")
	var result []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			result = append(result, line)
		}
	}

	return result, nil
}

func writeLogFiles(files []string) error {
	logFilesPath, err := getLogFilesPath()
	if err != nil {
		return err
	}

	content := strings.Join(files, "\n")
	if len(files) > 0 {
		content += "\n"
	}

	return os.WriteFile(logFilesPath, []byte(content), 0644)
}

func handleLogFileAdd(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("requires file path")
	}

	filePath := args[0]
	if len(args) > 1 {
		return fmt.Errorf("unrecognized extra args: %v", args[1:])
	}

	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return err
	}

	if _, err := os.Stat(absPath); err != nil {
		return fmt.Errorf("file does not exist: %s", absPath)
	}

	logFiles, err := readLogFiles()
	if err != nil {
		return err
	}

	for _, existing := range logFiles {
		if existing == absPath {
			fmt.Printf("File already in list: %s\n", absPath)
			return nil
		}
	}

	logFiles = append(logFiles, absPath)
	err = writeLogFiles(logFiles)
	if err != nil {
		return err
	}

	fmt.Printf("Added file: %s\n", absPath)
	return nil
}

func handleLogFileList(args []string) error {
	if len(args) > 0 {
		return fmt.Errorf("unrecognized extra args: %v", args)
	}

	logFiles, err := readLogFiles()
	if err != nil {
		return err
	}

	for _, file := range logFiles {
		fmt.Println(file)
	}

	return nil
}

func handleLogFileRm(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("requires file path")
	}

	filePath := args[0]
	if len(args) > 1 {
		return fmt.Errorf("unrecognized extra args: %v", args[1:])
	}

	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return err
	}

	logFiles, err := readLogFiles()
	if err != nil {
		return err
	}

	var found bool
	var newLogFiles []string
	for _, existing := range logFiles {
		if existing == absPath {
			found = true
			continue
		}
		newLogFiles = append(newLogFiles, existing)
	}

	if !found {
		return fmt.Errorf("file not found in list: %s", absPath)
	}

	err = writeLogFiles(newLogFiles)
	if err != nil {
		return err
	}

	fmt.Printf("Removed file: %s\n", absPath)
	return nil
}
