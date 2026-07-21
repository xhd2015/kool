# Scenario

**Feature**: sealed sandbox binary unpacks, materializes, and runs a guest command

```
# host-built sealed binary
kool sandbox build -o sandbox.bin …   # host GOOS/GOARCH (no --goos linux)
  -> sandbox.bin

# run under forced materialize parent
KOOL_SANDBOX_ROOT=PARENT ./sandbox.bin [--] <command> [args...]
  -> unseal + materialize under PARENT/<session>/; exec; cleanup
```

## Steps

1. Subcommand=build with host target (clear `--goos`/`--goarch`).
2. Default `-o sandbox.bin` under WorkingDir.
3. Enable `AfterBuildRun`; resolve `SandboxRootParent` under WorkingDir.

```go
import (
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	req.HelpAtRoot = false
	req.HelpBuild = false
	req.Subcommand = "build"
	req.AfterBuildInspect = false
	req.BuildTwice = false
	// Host GOOS/GOARCH so the sealed binary can execute on this machine.
	req.Goos = ""
	req.Goarch = ""
	if !req.OutputSet && req.Output == "" {
		req.Output = "sandbox.bin"
		req.OutputSet = true
	}
	req.AfterBuildRun = true
	if req.SandboxRootParent == "" {
		req.SandboxRootParent = filepath.Join(req.WorkingDir, "kool-sandbox-root")
	}
	return nil
}
```
