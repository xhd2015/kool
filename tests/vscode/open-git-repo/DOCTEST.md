# kool vscode open-git-repo

`kool vscode open-git-repo <path>` validates a local git repository, builds a
`vscode://xhd2015.open-in-new-window/git-open?path=<encoded>` URI, and opens it
via the OS handler (`open` / `xdg-open` / `cmd /c start`).

## Version

0.0.2

## DSN (Domain Specific Notion)

### Participants
- **kool CLI** — `vscode.go` subcommand `open-git-repo`; orchestrates validate →
  build URI → exec open.
- **`validateGitRepoPath`** — resolves path against cwd, checks exists, directory,
  and `.git` (file or directory); returns normalized absolute path.
- **`buildGitOpenRepoURI`** — constructs `vscode://xhd2015.open-in-new-window/git-open?path=...`
  with URL-encoded absolute path.
- **OS opener** — `open` (darwin), `xdg-open` (linux), `cmd /c start` (windows);
  injectable `execCommand` for tests.
- **VS Code extension** — receives URI and calls `git.openRepository` (tested separately
  in `tests/git-open-cli/`).

### Behaviors
- **Validation failure** — missing arg, nonexistent path, not a directory, or no `.git`
  → stderr error, non-zero exit, no OS open.
- **URI building** — absolute/relative/trailing-slash/spaces paths normalize to encoded URI.
- **Exec** — on success, OS opener invoked with built `vscode://` URI.

## Decision Tree

```
open-git-repo/
├── validation/
│   ├── missing-arg/        → stderr usage error; no open
│   ├── nonexistent-path/   → error before open
│   ├── not-directory/      → error before open
│   └── no-git/             → "not a git repository"
├── uri/
│   ├── absolute-path/      → correct vscode:// URI
│   ├── relative-path/      → cwd-resolved absolute in URI
│   ├── trailing-slash/     → normalized path in URI
│   ├── spaces-in-path/     → URL-encoded spaces
│   └── worktree-git-file/  → .git file worktree accepted
└── exec/
    └── invokes-open/       → mock exec; opener called with URI
```

## Test Index

| # | Path | Description |
|---|------|-------------|
| 1 | `validation/missing-arg/` | No path argument shows usage error |
| 2 | `validation/nonexistent-path/` | Nonexistent path fails before open |
| 3 | `validation/not-directory/` | File path fails before open |
| 4 | `validation/no-git/` | Directory without `.git` fails |
| 5 | `uri/absolute-path/` | Absolute path produces correct URI |
| 6 | `uri/relative-path/` | Relative path resolved in URI |
| 7 | `uri/trailing-slash/` | Trailing slash stripped in URI |
| 8 | `uri/spaces-in-path/` | Spaces URL-encoded in URI |
| 9 | `uri/worktree-git-file/` | Worktree `.git` file accepted |
| 10 | `exec/invokes-open/` | OS opener invoked with built URI |

## How to Run

```sh
cd kool-vscode
doctest vet ./tests/vscode/open-git-repo
doctest test ./tests/vscode/open-git-repo
go test ./...
```

```go
import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	vscodegit "github.com/xhd2015/kool/vscodegit"
)

type Request struct {
	Phase      string
	RepoPath   string
	WorkingDir string
	GoOS       string
}

type Response struct {
	NormalizedPath string
	VSCodeURI      string
	Stdout         string
	Stderr         string
	ExitCode       int
	ExecCalled     bool
	ExecCommand    string
	ExecArgs       []string
	ValidateErr    string
}

func resolveKoolBinary() (string, error) {
	moduleRoot := filepath.Join(DOCTEST_ROOT, "..", "..", "..")
	candidates := []string{
		filepath.Join(moduleRoot, "kool"),
		filepath.Join(moduleRoot, "bin", "kool"),
	}
	for _, candidate := range candidates {
		if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
			return candidate, nil
		}
	}
	if path, err := exec.LookPath("kool"); err == nil {
		return path, nil
	}
	return "", fmt.Errorf("kool binary not found in PATH or %s; build with: go build -o kool .", moduleRoot)
}

func Run(t *testing.T, req *Request) (*Response, error) {
	switch req.Phase {
	case "cli":
		return runCLI(t, req)
	case "validate":
		return runValidate(t, req)
	case "build-uri":
		return runBuildURI(t, req)
	case "exec":
		return runExec(t, req)
	default:
		return nil, fmt.Errorf("unknown phase %q", req.Phase)
	}
}

func runCLI(t *testing.T, req *Request) (*Response, error) {
	koolBin, err := resolveKoolBinary()
	if err != nil {
		return nil, err
	}
	args := []string{"vscode", "open-git-repo"}
	if req.RepoPath != "" {
		args = append(args, req.RepoPath)
	}
	cmd := exec.Command(koolBin, args...)
	if req.WorkingDir != "" {
		cmd.Dir = req.WorkingDir
	}
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	runErr := cmd.Run()
	exitCode := 0
	if runErr != nil {
		if exitErr, ok := runErr.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			return nil, fmt.Errorf("failed to run kool: %w", runErr)
		}
	}
	return &Response{
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
		ExitCode: exitCode,
	}, nil
}

func runValidate(t *testing.T, req *Request) (*Response, error) {
	cwd := req.WorkingDir
	if cwd == "" {
		cwd, _ = os.Getwd()
	}
	normalized, err := vscodegit.ValidateGitRepoPath(req.RepoPath, cwd)
	resp := &Response{NormalizedPath: normalized}
	if err != nil {
		resp.ValidateErr = err.Error()
	}
	return resp, nil
}

func runBuildURI(t *testing.T, req *Request) (*Response, error) {
	cwd := req.WorkingDir
	if cwd == "" {
		var err error
		cwd, err = os.Getwd()
		if err != nil {
			return nil, err
		}
	}
	normalized, err := vscodegit.ValidateGitRepoPath(req.RepoPath, cwd)
	if err != nil {
		return &Response{ValidateErr: err.Error()}, nil
	}
	uri := vscodegit.BuildGitOpenRepoURI(normalized)
	return &Response{
		NormalizedPath: normalized,
		VSCodeURI:      uri,
	}, nil
}

func runExec(t *testing.T, req *Request) (*Response, error) {
	cwd := req.WorkingDir
	if cwd == "" {
		var err error
		cwd, err = os.Getwd()
		if err != nil {
			return nil, err
		}
	}
	var execCalled bool
	var execCommand string
	var execArgs []string
	vscodegit.SetExecCommandHook(func(name string, arg ...string) *exec.Cmd {
		execCalled = true
		execCommand = name
		execArgs = append([]string{}, arg...)
		return exec.Command("true")
	})
	t.Cleanup(func() { vscodegit.SetExecCommandHook(nil) })
	if req.GoOS != "" {
		vscodegit.SetGOOSForTest(req.GoOS)
		t.Cleanup(func() { vscodegit.SetGOOSForTest("") })
	}

	err := vscodegit.OpenGitRepo(req.RepoPath, cwd)
	resp := &Response{
		ExecCalled:  execCalled,
		ExecCommand: execCommand,
		ExecArgs:    execArgs,
	}
	if err != nil {
		resp.ValidateErr = err.Error()
		return resp, nil
	}
	if execCalled && len(execArgs) > 0 {
		resp.VSCodeURI = execArgs[len(execArgs)-1]
	}
	return resp, nil
}
```