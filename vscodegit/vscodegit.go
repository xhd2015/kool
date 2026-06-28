package vscodegit

import (
	"encoding/json"
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
var stdoutWriterOverride io.Writer

// SetStderrWriterForTest redirects stderr hint output for tests.
func SetStderrWriterForTest(w io.Writer) {
	stderrWriterOverride = w
}

// SetStdoutWriterForTest redirects JSON stdout for tests.
func SetStdoutWriterForTest(w io.Writer) {
	stdoutWriterOverride = w
}

func stderrWriter() io.Writer {
	if stderrWriterOverride != nil {
		return stderrWriterOverride
	}
	return os.Stderr
}

func stdoutWriter() io.Writer {
	if stdoutWriterOverride != nil {
		return stdoutWriterOverride
	}
	return os.Stdout
}

func writeIPCFallbackHint(jsonMode bool) {
	if jsonMode {
		return
	}
	fmt.Fprint(stderrWriter(), ipcFallbackHint)
}

type openJSONPayload struct {
	IPCHandled bool   `json:"ipc_handled"`
	Path       string `json:"path"`
	Error      string `json:"error,omitempty"`
	Fallback   string `json:"fallback,omitempty"`
	OK         *bool  `json:"ok,omitempty"`
}

func writeOpenJSON(payload openJSONPayload) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(stdoutWriter(), string(data))
	return err
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

// OpenOptions configures directory open behavior.
type OpenOptions struct {
	Replace bool
	IpcOnly bool
	Json    bool
}

// OpenDir validates the path, runs precheck, tries IPC, and falls back to URI.
func OpenDir(path string, cwd string, replace bool) error {
	return OpenDirOptions(path, cwd, OpenOptions{Replace: replace})
}

// OpenDirOptions is the full open pipeline with IPC-only and JSON output options.
func OpenDirOptions(path string, cwd string, opts OpenOptions) error {
	normalized, err := ValidateDirPath(path, cwd)
	if err != nil {
		return err
	}
	if err := runPrecheck(); err != nil {
		return err
	}
	if err := sendIPC("open", normalized, opts.Replace); err != nil {
		if opts.IpcOnly {
			if opts.Json {
				_ = writeOpenJSON(openJSONPayload{
					IPCHandled: false,
					Path:       normalized,
					Error:      err.Error(),
				})
			}
			return fmt.Errorf("IPC open failed: %w", err)
		}
		if opts.Json {
			if fallbackErr := openURI(BuildOpenURI(normalized, opts.Replace)); fallbackErr != nil {
				return fallbackErr
			}
			ok := true
			return writeOpenJSON(openJSONPayload{
				IPCHandled: false,
				Path:       normalized,
				Fallback:   "uri",
				OK:         &ok,
			})
		}
		writeIPCFallbackHint(false)
		return openURI(BuildOpenURI(normalized, opts.Replace))
	}
	if opts.Json {
		return writeOpenJSON(openJSONPayload{
			IPCHandled: true,
			Path:       normalized,
		})
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
		writeIPCFallbackHint(false)
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