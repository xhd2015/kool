package vscodegit

import (
	"fmt"
	"io"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

const ipcFallbackHint = "Note: extension not reachable via IPC; opening via vscode:// URI."

var stderrWriterOverride io.Writer

// SetStderrWriterForTest redirects stderr hint output for tests.
func SetStderrWriterForTest(w io.Writer) {
	stderrWriterOverride = w
}

func stderrWriter() io.Writer {
	if stderrWriterOverride != nil {
		return stderrWriterOverride
	}
	return os.Stderr
}

func writeIPCFallbackHint() {
	fmt.Fprint(stderrWriter(), ipcFallbackHint)
}

const extensionID = "xhd2015.open-in-new-window"

var execCommandHook func(name string, arg ...string) *exec.Cmd
var goosOverride string

func SetExecCommandHook(hook func(name string, arg ...string) *exec.Cmd) {
	execCommandHook = hook
}

func SetGOOSForTest(goos string) {
	goosOverride = goos
}

func effectiveGOOS() string {
	if goosOverride != "" {
		return goosOverride
	}
	return runtime.GOOS
}

func execCommand(name string, arg ...string) *exec.Cmd {
	if execCommandHook != nil {
		return execCommandHook(name, arg...)
	}
	return exec.Command(name, arg...)
}

// ValidateGitRepoPath resolves path against cwd, verifies it is an existing
// directory containing .git (file or directory), and returns the normalized
// absolute path.
func ValidateGitRepoPath(path string, cwd string) (string, error) {
	trimmed := strings.TrimSpace(path)
	if trimmed == "" {
		return "", fmt.Errorf("path is required")
	}

	joined := trimmed
	if !filepath.IsAbs(trimmed) {
		joined = filepath.Join(cwd, trimmed)
	}

	absPath, err := filepath.Abs(filepath.Clean(joined))
	if err != nil {
		return "", err
	}
	absPath = strings.TrimRight(absPath, "/\\")

	info, err := os.Stat(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("path does not exist: %s", absPath)
		}
		return "", err
	}
	if !info.IsDir() {
		return "", fmt.Errorf("not a directory: %s", absPath)
	}

	gitPath := filepath.Join(absPath, ".git")
	if _, err := os.Stat(gitPath); err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("not a git repository: %s", absPath)
		}
		return "", err
	}

	return absPath, nil
}

// ValidateDirPath resolves path against cwd, verifies it is an existing directory,
// and returns the normalized absolute path.
func ValidateDirPath(path string, cwd string) (string, error) {
	trimmed := strings.TrimSpace(path)
	if trimmed == "" {
		return "", fmt.Errorf("path is required")
	}

	joined := trimmed
	if !filepath.IsAbs(trimmed) {
		joined = filepath.Join(cwd, trimmed)
	}

	absPath, err := filepath.Abs(filepath.Clean(joined))
	if err != nil {
		return "", err
	}
	absPath = strings.TrimRight(absPath, "/\\")

	info, err := os.Stat(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("exist: path does not exist: %s", absPath)
		}
		return "", err
	}
	if !info.IsDir() {
		return "", fmt.Errorf("directory: not a directory: %s", absPath)
	}

	return absPath, nil
}

// BuildOpenURI constructs the vscode:// deep link for opening a directory.
func BuildOpenURI(absPath string, replace bool) string {
	encoded := strings.ReplaceAll(absPath, " ", "%20")
	uri := fmt.Sprintf("vscode://%s/open?path=%s", extensionID, encoded)
	if replace {
		uri += "&replace=true"
	}
	return uri
}

// BuildGitOpenRepoURI constructs the vscode:// deep link for opening a git repo.
func BuildGitOpenRepoURI(absPath string) string {
	return fmt.Sprintf("vscode://%s/git-open?path=%s", extensionID, url.QueryEscape(absPath))
}

// OpenDir validates the path, runs precheck, tries IPC, and falls back to URI.
func OpenDir(path string, cwd string, replace bool) error {
	normalized, err := ValidateDirPath(path, cwd)
	if err != nil {
		return err
	}
	if err := runPrecheck(); err != nil {
		return err
	}
	if err := sendIPC("open", normalized, replace); err != nil {
		writeIPCFallbackHint()
		return openURI(BuildOpenURI(normalized, replace))
	}
	return nil
}

// OpenGitRepo validates the path, runs precheck, tries IPC, and falls back to URI.
func OpenGitRepo(path string, cwd string) error {
	normalized, err := ValidateGitRepoPath(path, cwd)
	if err != nil {
		return err
	}
	if err := runPrecheck(); err != nil {
		return err
	}
	if err := sendIPC("git-open", normalized, false); err != nil {
		writeIPCFallbackHint()
		return openURI(BuildGitOpenRepoURI(normalized))
	}
	return nil
}

func openURI(uri string) error {
	goos := effectiveGOOS()

	var cmd *exec.Cmd
	switch goos {
	case "windows":
		cmd = execCommand("cmd", "/c", "start", uri)
	case "darwin":
		cmd = execCommand("open", uri)
	case "linux":
		cmd = execCommand("xdg-open", uri)
	default:
		return fmt.Errorf("unsupported platform: %s", goos)
	}

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to open URI: %w", err)
	}
	return nil
}