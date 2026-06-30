# kool go update

`kool go update <dir>` updates a local Go module dependency to the latest git version tag,
dropping any replace directive and setting the require version in the current module's
go.mod.

# DSN (Domain Specific Notion)

The **user** runs `kool go update <dir>` from a consumer module that depends on a local
module via replace. The **update handler** reads the target directory's go.mod, calculates
a version tag prefix (submodule path under the git root), finds the latest matching tag,
and edits the consumer's go.mod to require that version without a replace.

## Version

0.0.1

## Decision Tree

```
target-layout
└── nested-module-no-root-gomod/     # git repo without root go.mod; module only in subdirectory
    ├── update                       # library Update() from consumer cwd
    └── cli                          # kool go update CLI from consumer cwd
```

## Test Cases

| # | Path | Description |
|---|------|-------------|
| 1 | nested-module-no-root-gomod/update | Update nested module when git root has no go.mod |
| 2 | nested-module-no-root-gomod/cli | CLI updates nested module when git root has no go.mod |

## How to Run

```sh
doctest vet ./tests/go/update
doctest test ./tests/go/update
```

```go
import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"testing"

	go_update "github.com/xhd2015/dot-pkgs/go-pkgs/gotool/update"
	"github.com/xhd2015/dot-pkgs/go-pkgs/gotool/resolve"
)

type Request struct {
	Operation   string // "update" (library) or "cli"
	TargetDir   string // populated by Setup: nested module directory
	ConsumerDir string // populated by Setup: consumer module directory
}

type Response struct {
	UpdateErr     error
	Stdout        string
	Stderr        string
	ExitCode      int
	ModuleVersion string // require version after update
}

func Run(t *testing.T, req *Request) (*Response, error) {
	resp := &Response{}

	switch req.Operation {
	case "update":
		oldWd, err := os.Getwd()
		if err != nil {
			return nil, err
		}
		if err := os.Chdir(req.ConsumerDir); err != nil {
			return nil, err
		}
		defer os.Chdir(oldWd)
		resp.UpdateErr = go_update.Update(req.TargetDir)

	case "cli":
		cmd := exec.Command("kool", "go", "update", req.TargetDir)
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
				resp.UpdateErr = runErr
			} else {
				return nil, fmt.Errorf("failed to run kool: %w", runErr)
			}
		}

	default:
		return nil, fmt.Errorf("unknown operation: %s", req.Operation)
	}

	if resp.UpdateErr == nil && resp.ExitCode == 0 {
		modInfo, err := resolve.GetModuleInfo(req.ConsumerDir)
		if err != nil {
			return nil, fmt.Errorf("failed to read consumer go.mod: %w", err)
		}
		for _, reqEntry := range modInfo.Require {
			if reqEntry.Path == "github.com/example/dot-pkgs/go-pkgs" {
				resp.ModuleVersion = reqEntry.Version
				break
			}
		}
	}

	return resp, nil
}
```