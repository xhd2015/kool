# Scenario

**Feature**: kool for-every / for-every-\<duration\> loop utility

```
# help (no loop)
user -> kool for-every --help -> usage (both forms + flags), exit 0

# validation (no loop)
user -> kool for-every [bad duration|flags|missing command]
  -> stderr error, non-zero exit

# loop (bounded)
user -> kool for-every[-<dur>] [OPTIONS] <cmd> …
  -> run cmd immediately, then sleep Interval, repeat until stop condition
```

## Preconditions

- Module root is `d.DOCTEST_ROOT/../..` (this tree lives at `tests/for-every/`).
- `go` is on PATH; `Run` builds `kool` once per process under an in-memory mutex
  into `os.MkdirTemp("", "kool-for-every-doctest-bin-")`.
- Loop leaves **must** pass `--max-runs` and/or failure stop flags so the child
  cannot run forever. `Run` also applies a wall-clock process timeout (default 15s).
- Prefer short intervals (`10ms`) in loop tests.

## Steps

1. Root `Setup` creates an isolated `WorkingDir` and default process timeout.
2. Grouping/leaf `Setup` narrows form, duration, flags, and child command.
3. `Run` builds/reuses process-local `kool` and executes the argv from `Request`.

## Context

- Process-local binary memo (mutex + path/err); no session disk flock.
- No durable product storage; per-leaf temp dirs only.
- Helper `intPtr` is available for optional int flags.

```go
import (
	"os"
	"testing"
	"time"
)

func Setup(t *testing.T, req *Request) error {
	if req.WorkingDir == "" {
		req.WorkingDir = t.TempDir()
	}
	if req.ProcessTimeout <= 0 {
		req.ProcessTimeout = 15 * time.Second
	}
	// Touch working dir so leaves can write counter files relative to cwd.
	if err := os.MkdirAll(req.WorkingDir, 0755); err != nil {
		return err
	}
	return nil
}

// markRootTree keeps hierarchical child packages importing this package live.
func markRootTree() {}
```
