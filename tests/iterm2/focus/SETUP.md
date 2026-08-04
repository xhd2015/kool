# Scenario

**Feature**: focus an existing iTerm2 directory session through an injected boundary

```text
kool iterm2 focus <dir> [--index N] -> focus service -> FocusFake
FocusFake -> candidate inventory / selected session -> CLI output
```

## Preconditions

- Each leaf receives a directory under its `d.DOCTEST_CASE`; no process cwd or environment is changed.
- `FocusFake` is the required writer-injected L2 boundary: it returns `Candidates`, records discovery, and records focus selections.

## Steps

1. Build focus arguments and an isolated target path.
2. Invoke `RunFocusForTest` with stdout/stderr buffers and `FocusFake`.
3. Assert rendered output and fake-boundary calls.

## Context

- Candidate ordering is the service order; displayed indexes must remain 0-based and stable.

```go
import (
    "os"
    "testing"

    "github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
    if req.TargetKind == "" { req.TargetKind = "directory" }
    return nil
}
```
