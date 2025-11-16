package rules

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/xhd2015/kool/tools/config"
)

// Variable to allow mocking exec.Command in tests
var execCommand = exec.Command

func Handle(args []string) error {
	if len(args) == 0 {
		// No arguments provided, enter sub-shell mode
		return EnterSubshell()
	}

	cmd := args[0]
	args = args[1:]

	switch cmd {
	case "add":
		if len(args) == 0 {
			return fmt.Errorf("usage: kool rule add <file>")
		}
		rulePath := args[0]
		resultName, err := Add(rulePath)
		if err != nil {
			return err
		}

		inputName := filepath.Base(rulePath)
		if resultName != inputName {
			fmt.Fprintf(os.Stderr, "file with name '%s' already exists, added as '%s'\n", inputName, resultName)
		}
		return nil
	case "get":
		if len(args) == 0 {
			return fmt.Errorf("usage: kool rule get <file>")
		}
		ruleName := args[0]
		return Get(ruleName)
	case "list":
		files, err := List()
		if err != nil {
			return err
		}
		for _, file := range files {
			fmt.Println(file)
		}
		return nil
	case "mv":
		if len(args) < 2 {
			return fmt.Errorf("usage: kool rule mv <source_file> <dest_file>")
		}
		srcName := args[0]
		destName := args[1]
		return Move(srcName, destName)
	case "use":
		if len(args) == 0 {
			return fmt.Errorf("usage: kool rule use <file>")
		}
		ruleName := args[0]
		return Use(ruleName)
	case "dir":
		rulesDir, err := GetRulesDir()
		if err != nil {
			return err
		}
		fmt.Println(rulesDir)
		return nil
	case "rm":
		if len(args) == 0 {
			return fmt.Errorf("usage: kool rule rm <file>")
		}
		ruleName := args[0]
		return Remove(ruleName)
	default:
		return fmt.Errorf("unknown rule command: %s", cmd)
	}
}

// EnterSubshell creates and enters a bash sub-shell with a custom prompt
// and special 'use' command for rule management
func EnterSubshell() error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	rulesDir, err := GetRulesDir()
	if err != nil {
		return err
	}

	// Create custom bash initialization script
	initScript := fmt.Sprintf(`
# Source user's bashrc if available
if [ -f ~/.bash_profile ]; then
	. ~/.bash_profile
fi

# Set custom prompt to show current directory
PS1="(kool rules)$PS1"

# Define special use command
function use {
	if [ $# -eq 1 ]; then
		(cd "$OLD_PWD" && command kool rule use "$1")
	else
		echo "Usage: use <file>"
	fi
}

# Define aliases for common rule commands
alias list='kool rule list'
alias dir='kool rule dir'
alias add='kool rule add'
alias rm='kool rule rm'
alias mv='kool rule mv'
alias get='kool rule get'

# Show welcome message
echo ""
echo "Welcome to kool rules shell!"
echo "Type 'list' to see available rules, 'use <file>' to use a rule."
echo "Type 'exit' or press Ctrl+D to return to normal shell."
echo "Rules directory is at: %s"
echo "Previous working directory is: %s"
echo ""
`, rulesDir, cwd)

	// Create a temporary file for the initialization script
	tmpFile, err := os.CreateTemp("", "kool-rule-init-*.sh")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(initScript); err != nil {
		return fmt.Errorf("failed to write to temp file: %w", err)
	}
	tmpFile.Close()

	// Launch bash with the initialization script
	cmd := execCommand("bash", "--rcfile", tmpFile.Name(), "-i")
	// cmd.Dir = rulesDir
	// cmd.Env = append(os.Environ(), "OLD_PWD="+cwd)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// GetRulesDir returns the path to the .kool/rules directory in the user's home directory
func GetRulesDir() (string, error) {
	koolConfigDir, err := config.GetKoolConfigDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(koolConfigDir, "rules", "files")

	// Ensure the directory exists
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("failed to create rules directory: %w", err)
	}

	return dir, nil
}

// GetCursorRulesDir returns the path to the .cursor/rules directory in the current working directory
func GetCursorRulesDir() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get current working directory: %w", err)
	}

	dir := filepath.Join(cwd, ".cursor", "rules")

	// Ensure the directory exists
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("failed to create .cursor/rules directory: %w", err)
	}

	return dir, nil
}

// Add copies a rule file to the ~/.kool/rules/files directory
// If a file with the same name already exists, a timestamp will be appended to make it unique
func Add(rulePath string) (string, error) {
	if !fileExists(rulePath) {
		return "", fmt.Errorf("file not found: %s", rulePath)
	}

	rulesDir, err := GetRulesDir()
	if err != nil {
		return "", err
	}

	fileName := filepath.Base(rulePath)
	destPath := filepath.Join(rulesDir, fileName)
	resultName := fileName

	// If file already exists, append a date-time suffix
	if fileExists(destPath) {
		// Generate timestamp suffix in format YYYYMMDD-HHMMSS
		timestamp := time.Now().Format("20060102-150405")

		// Split filename and extension to insert timestamp
		ext := filepath.Ext(fileName)
		baseName := fileName[:len(fileName)-len(ext)]

		// Create new filename with timestamp
		resultName = fmt.Sprintf("%s-%s%s", baseName, timestamp, ext)
		destPath = filepath.Join(rulesDir, resultName)
	}

	err = copyFile(rulePath, destPath)
	return resultName, err
}

// List returns a list of rule files in the ~/.kool/rules/files directory
func List() ([]string, error) {
	rulesDir, err := GetRulesDir()
	if err != nil {
		return nil, err
	}

	return listFilesInDir(rulesDir)
}

// Use copies a rule file from ~/.kool/rules/files to .cursor/rules if it doesn't already exist
func Use(ruleName string) error {
	cursorRulesDir, err := GetCursorRulesDir()
	if err != nil {
		return err
	}

	cursorRulePath := filepath.Join(cursorRulesDir, ruleName)

	// Check if file already exists in .cursor/rules
	if fileExists(cursorRulePath) {
		return fmt.Errorf("rule already exists in .cursor/rules: %s", ruleName)
	}

	// Find the rule in ~/.kool/rules/files
	rulesDir, err := GetRulesDir()
	if err != nil {
		return err
	}

	rulePath := filepath.Join(rulesDir, ruleName)
	if !fileExists(rulePath) {
		return fmt.Errorf("rule not found in ~/.kool/rules/files: %s", ruleName)
	}

	// Copy the rule to .cursor/rules
	return copyFile(rulePath, cursorRulePath)
}

// Remove deletes a rule file from the ~/.kool/rules/files directory
func Remove(ruleName string) error {
	rulesDir, err := GetRulesDir()
	if err != nil {
		return err
	}

	rulePath := filepath.Join(rulesDir, ruleName)
	if !fileExists(rulePath) {
		return fmt.Errorf("rule not found: %s", ruleName)
	}

	return os.Remove(rulePath)
}

// Move renames a rule file in the ~/.kool/rules/files directory
func Move(srcName, destName string) error {
	rulesDir, err := GetRulesDir()
	if err != nil {
		return err
	}

	// Check if source file exists
	srcPath := filepath.Join(rulesDir, srcName)
	if !fileExists(srcPath) {
		return fmt.Errorf("source file not found: %s", srcName)
	}

	// Check if destination file already exists
	destPath := filepath.Join(rulesDir, destName)
	if fileExists(destPath) {
		return fmt.Errorf("destination file already exists: %s", destName)
	}

	// Backup the source file before moving
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get user home directory: %w", err)
	}

	// Create backup directory with today's date
	now := time.Now()
	backupDir := filepath.Join(home, "CursorDeleted", now.Format("2006-01-02"))
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return fmt.Errorf("failed to create backup directory: %w", err)
	}

	backupPath := filepath.Join(backupDir, srcName)
	if err := copyFile(srcPath, backupPath); err != nil {
		return fmt.Errorf("failed to backup file: %w", err)
	}

	// Rename the file
	if err := os.Rename(srcPath, destPath); err != nil {
		return fmt.Errorf("failed to rename file: %w", err)
	}

	fmt.Printf("Renamed '%s' to '%s'\n", srcName, destName)
	fmt.Printf("Backup saved to '%s'\n", backupPath)
	return nil
}

// Get prints the content of a rule file from the ~/.kool/rules/files directory
func Get(ruleName string) error {
	rulesDir, err := GetRulesDir()
	if err != nil {
		return err
	}

	rulePath := filepath.Join(rulesDir, ruleName)
	if !fileExists(rulePath) {
		return fmt.Errorf("rule not found: %s", ruleName)
	}

	content, err := os.ReadFile(rulePath)
	if err != nil {
		return fmt.Errorf("failed to read rule file: %w", err)
	}

	fmt.Print(string(content))
	return nil
}

// Helper functions

// fileExists checks if a file exists
func fileExists(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return fmt.Errorf("failed to copy file content: %w", err)
	}

	return nil
}

// listFilesInDir returns a list of files in a directory
func listFilesInDir(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	var files []string
	for _, entry := range entries {
		if !entry.IsDir() {
			files = append(files, entry.Name())
		}
	}

	return files, nil
}
