# kool go replace

`kool go replace <dir>` adds a `replace` directive in the current module's go.mod,
mapping the target directory's module path to its absolute filesystem path.

# DSN (Domain Specific Notion)

The **user** runs `kool go replace <dir>` from a consumer module that depends on a local
module via require. The **replace handler** reads the target directory's go.mod, resolves
its module path, and edits the consumer's go.mod to add `replace modulePath => absDir`.

## Version

0.0.1

## Decision Tree

```
target-layout
└── nested-module-no-root-gomod/     # git repo without root go.mod; module only in subdirectory
    ├── replace                      # library Replace() from consumer cwd
    └── cli                          # kool go replace CLI from consumer cwd
```

## Test Cases

| # | Path | Description |
|---|------|-------------|
| 1 | nested-module-no-root-gomod/replace | Replace nested module when git root has no go.mod |
| 2 | nested-module-no-root-gomod/cli | CLI replaces nested module when git root has no go.mod |

## How to Run

```sh
doctest vet ./tests/go/replace
doctest test ./tests/go/replace
```

```go
import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	go_replace "github.com/xhd2015/dot-pkgs/go-pkgs/gotool/replace"
	"github.com/xhd2015/dot-pkgs/go-pkgs/gotool/resolve"
)

type Request struct {
	Operation   string // "replace" (library) or "cli"
	TargetDir   string // populated by Setup: nested module directory
	ConsumerDir string // populated by Setup: consumer module directory
}

type Response struct {
	ReplaceErr error
	Stdout     string
	Stderr     string
	ExitCode   int
	ModulePath string
	AbsDir     string
	HasReplace bool
}

func Run(t *testing.T, req *Request) (*Response, error) {
	resp := &Response{}

	switch req.Operation {
	case "replace":
		oldWd, err := os.Getwd()
		if err != nil {
			return nil, err
		}
		if err := os.Chdir(req.ConsumerDir); err != nil {
			return nil, err
		}
		defer os.Chdir(oldWd)
		resp.AbsDir, resp.ModulePath, resp.ReplaceErr = go_replace.Replace(req.TargetDir)

	case "cli":
		cmd := exec.Command("kool", "go", "replace", req.TargetDir)
		cmd.Dir = req.ConsumerDir
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		runErr := cmd.Run()
		resp.Stdout = stdout.String()
		resp.Stderr = stderr.String()
		if runErr != nil {
			if exitErr, ok := runErr.(*exec.ExitError); ok {
				resp.ExitCode = exitErr.ExitCode()
				resp.ReplaceErr = runErr
			} else {
				return nil, fmt.Errorf("failed to run kool: %w", runErr)
			}
		}

	default:
		return nil, fmt.Errorf("unknown operation: %s", req.Operation)
	}

	if resp.ReplaceErr == nil && resp.ExitCode == 0 {
		modInfo, err := resolve.GetModuleInfo(req.ConsumerDir)
		if err != nil {
			return nil, fmt.Errorf("failed to read consumer go.mod: %w", err)
		}
		if resp.ModulePath == "" {
			for _, reqEntry := range modInfo.Require {
				if reqEntry.Path == "github.com/example/dot-pkgs/go-pkgs" {
					resp.ModulePath = reqEntry.Path
					break
				}
			}
		}
		if resp.AbsDir == "" {
			absDir, absErr := filepath.Abs(req.TargetDir)
			if absErr != nil {
				return nil, fmt.Errorf("failed to resolve target abs path: %w", absErr)
			}
			resp.AbsDir = absDir
		}
		resp.HasReplace = hasReplaceFor(modInfo, resp.ModulePath, resp.AbsDir)
	}

	return resp, nil
}

func hasReplaceFor(modInfo *resolve.ModuleInfo, modulePath, absDir string) bool {
	for _, repl := range modInfo.Replace {
		if repl.Old.Path == modulePath && repl.New.Path == absDir {
			return true
		}
	}
	return false
}
```