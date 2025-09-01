package staged

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

func Handle(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: kool git staged <backup|restore> <file>")
	}

	cmd := args[0]
	switch cmd {
	case "backup":
		if len(args) != 2 {
			return fmt.Errorf("usage: kool git staged backup <file.txt>")
		}
		return HandleBackup(args[1])
	case "restore":
		if len(args) != 2 {
			return fmt.Errorf("usage: kool git staged restore <file.txt>")
		}
		return HandleRestore(args[1])
	default:
		return fmt.Errorf("unknown command: %s, available commands: backup, restore", cmd)
	}
}

func HandleBackup(backupFile string) error {
	// Get staged files
	stagedFiles, err := getStagedFiles()
	if err != nil {
		return fmt.Errorf("failed to get staged files: %w", err)
	}

	if len(stagedFiles) == 0 {
		return fmt.Errorf("no staged files found")
	}

	// Ensure we're in the git root directory
	gitRoot, err := getGitRoot()
	if err != nil {
		return fmt.Errorf("failed to get git root: %w", err)
	}

	// Create backup file
	file, err := os.Create(backupFile)
	if err != nil {
		return fmt.Errorf("failed to create backup file: %w", err)
	}
	defer file.Close()

	for _, stagedFile := range stagedFiles {
		// Write file header
		header := fmt.Sprintf("===== %s =====\n", stagedFile.Path)
		if stagedFile.IsDeleted {
			header = fmt.Sprintf("===== %s [Deleted] =====\n", stagedFile.Path)
		} else if stagedFile.IsRenamed {
			header = fmt.Sprintf("===== %s -> %s =====\n", stagedFile.OldPath, stagedFile.Path)
		}

		if _, err := file.WriteString(header); err != nil {
			return fmt.Errorf("failed to write header for %s: %w", stagedFile.Path, err)
		}

		// Write file content if not deleted
		if !stagedFile.IsDeleted {
			content, err := getStagedFileContent(gitRoot, stagedFile.Path)
			if err != nil {
				return fmt.Errorf("failed to get content for %s: %w", stagedFile.Path, err)
			}
			if _, err := file.WriteString(content); err != nil {
				return fmt.Errorf("failed to write content for %s: %w", stagedFile.Path, err)
			}
		}

		if _, err := file.WriteString("\n"); err != nil {
			return fmt.Errorf("failed to write newline: %w", err)
		}
	}

	fmt.Printf("Backup created: %s (%d files)\n", backupFile, len(stagedFiles))
	return nil
}

func HandleRestore(backupFile string) error {
	// Check if worktree is clean
	if !isWorktreeClean() {
		return fmt.Errorf("worktree is not clean, please commit or stash your changes first")
	}

	// Ensure we're in the git root directory
	gitRoot, err := getGitRoot()
	if err != nil {
		return fmt.Errorf("failed to get git root: %w", err)
	}

	// Change to git root
	if err := os.Chdir(gitRoot); err != nil {
		return fmt.Errorf("failed to change to git root: %w", err)
	}

	// Parse backup file
	files, err := parseBackupFile(backupFile)
	if err != nil {
		return fmt.Errorf("failed to parse backup file: %w", err)
	}

	if len(files) == 0 {
		return fmt.Errorf("no files found in backup")
	}

	// Restore files
	for _, file := range files {
		if err := restoreFile(file); err != nil {
			return fmt.Errorf("failed to restore %s: %w", file.Path, err)
		}
	}

	fmt.Printf("Restored %d files from backup\n", len(files))
	return nil
}

type StagedFile struct {
	Path      string
	OldPath   string // for renames
	Content   string
	IsDeleted bool
	IsRenamed bool
}

func getStagedFiles() ([]StagedFile, error) {
	// Use git diff --cached --name-status to get staged files
	cmd := exec.Command("git", "diff", "--cached", "--name-status")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var files []StagedFile
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")

	for _, line := range lines {
		if line == "" {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}

		status := parts[0]
		path := parts[1]

		file := StagedFile{Path: path}

		// Handle different status codes
		switch status[0] {
		case 'D':
			file.IsDeleted = true
		case 'R':
			file.IsRenamed = true
			// For renames, git shows status like "R100" and we have oldpath newpath
			if len(parts) >= 3 {
				file.OldPath = parts[1]
				file.Path = parts[2]
			}
		}

		files = append(files, file)
	}

	return files, nil
}

func getStagedFileContent(gitRoot, filePath string) (string, error) {
	// Use git show :filename to get staged content
	cmd := exec.Command("git", "show", ":"+filePath)
	cmd.Dir = gitRoot
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(output), nil
}

func getGitRoot() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

func isWorktreeClean() bool {
	cmd := exec.Command("git", "status", "--porcelain")
	output, err := cmd.Output()
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(output)) == ""
}

func parseBackupFile(backupFile string) ([]StagedFile, error) {
	file, err := os.Open(backupFile)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var files []StagedFile
	var currentFile *StagedFile
	var contentLines []string

	scanner := bufio.NewScanner(file)
	headerRegex := regexp.MustCompile(`^===== (.+) =====\s*$`)
	deletedRegex := regexp.MustCompile(`^===== (.+) \[Deleted\] =====\s*$`)
	renamedRegex := regexp.MustCompile(`^===== (.+) -> (.+) =====\s*$`)

	for scanner.Scan() {
		line := scanner.Text()

		// Check for file headers
		if matches := renamedRegex.FindStringSubmatch(line); matches != nil {
			// Save previous file if exists
			if currentFile != nil {
				currentFile.Content = strings.Join(contentLines, "\n")
				files = append(files, *currentFile)
			}

			// Start new renamed file
			currentFile = &StagedFile{
				OldPath:   matches[1],
				Path:      matches[2],
				IsRenamed: true,
			}
			contentLines = []string{}
		} else if matches := deletedRegex.FindStringSubmatch(line); matches != nil {
			// Save previous file if exists
			if currentFile != nil {
				currentFile.Content = strings.Join(contentLines, "\n")
				files = append(files, *currentFile)
			}

			// Start new deleted file
			currentFile = &StagedFile{
				Path:      matches[1],
				IsDeleted: true,
			}
			contentLines = []string{}
		} else if matches := headerRegex.FindStringSubmatch(line); matches != nil {
			// Save previous file if exists
			if currentFile != nil {
				currentFile.Content = strings.Join(contentLines, "\n")
				files = append(files, *currentFile)
			}

			// Start new regular file
			currentFile = &StagedFile{
				Path: matches[1],
			}
			contentLines = []string{}
		} else {
			// Content line
			if currentFile != nil {
				contentLines = append(contentLines, line)
			}
		}
	}

	// Save last file
	if currentFile != nil {
		currentFile.Content = strings.Join(contentLines, "\n")
		files = append(files, *currentFile)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return files, nil
}

func restoreFile(file StagedFile) error {
	if file.IsDeleted {
		// For deleted files, remove them from filesystem and stage the deletion
		if err := os.Remove(file.Path); err != nil && !os.IsNotExist(err) {
			return err
		}
		cmd := exec.Command("git", "add", file.Path)
		return cmd.Run()
	}

	if file.IsRenamed {
		// For renamed files, first create the new file, then remove the old one
		if err := writeFileContent(file.Path, file.Content); err != nil {
			return err
		}
		if err := os.Remove(file.OldPath); err != nil && !os.IsNotExist(err) {
			return err
		}
		// Stage both operations
		cmd := exec.Command("git", "add", file.Path, file.OldPath)
		return cmd.Run()
	}

	// Regular file - write content and stage
	if err := writeFileContent(file.Path, file.Content); err != nil {
		return err
	}
	cmd := exec.Command("git", "add", file.Path)
	return cmd.Run()
}

func writeFileContent(filePath, content string) error {
	// Ensure directory exists
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	return os.WriteFile(filePath, []byte(content), 0644)
}
