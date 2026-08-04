# `kool iterm2 focus`

`kool iterm2 focus <dir> [--index N]` selects an already-open iTerm2 session
whose `path` or `user.koolTargetDir` exactly equals the canonical target
directory. It never creates a window, tab, or session.

## Version

0.0.2

# DSN (Domain Specific Notion)

## Participants

- **User** invokes the `focus` subcommand with a directory and optional candidate index.
- **Kool iTerm2 CLI** validates arguments and renders success or error output.
- **Focus service** canonicalizes the target, discovers candidate sessions, and selects one.
- **Injected iTerm boundary** supplies deterministic candidates and records the focus request.

## Behaviors

- User -> Kool CLI -> focus service -> injected boundary discovers all exact matches.
- One match -> focus service -> boundary selects the session -> CLI prints success.
- No match, ambiguous match, or invalid input -> CLI prints a fatal error; boundary does not focus or create anything.
- An explicit valid `--index` resolves an ambiguous result; an invalid index preserves the candidate list for retry.

## Decision Tree

```text
focus/
├── help/
│   ├── root-lists-focus/
│   └── focus-usage/
├── validation/
│   ├── missing-directory/
│   ├── not-a-directory/
│   └── non-integer-index/
└── matches/
    ├── none/
    ├── exactly-one/
    └── multiple/
        ├── requires-index/
        ├── selects-explicit-index/
        └── rejects-out-of-range-index/
```

## Test Index

| Leaf | Contract |
|---|---|
| `help/root-lists-focus/` | Parent help advertises `focus`. |
| `help/focus-usage/` | `focus --help` is successful and documents `<dir>` plus `--index`. |
| `validation/missing-directory/` | No target is a fatal usage error without discovery. |
| `validation/not-a-directory/` | Existing file target is rejected before discovery. |
| `validation/non-integer-index/` | A non-numeric index is a fatal parse error without discovery. |
| `matches/none/` | No exact canonical match is fatal and never focuses. |
| `matches/exactly-one/` | One match (including `user.koolTargetDir`) is focused and reported. |
| `matches/multiple/requires-index/` | Ambiguity is fatal and lists stable candidates with retry guidance. |
| `matches/multiple/selects-explicit-index/` | `--index 1` focuses exactly candidate 1. |
| `matches/multiple/rejects-out-of-range-index/` | Invalid index is fatal, lists valid choices, and never focuses. |

## How to Run

```sh
doctest vet ./tests/iterm2/focus
doctest test ./tests/iterm2/focus
```

```go
import (
    "bytes"
    "fmt"
    "os"
    "path/filepath"
    "testing"

    "github.com/xhd2015/doctest/session"
    iterm2cmd "github.com/xhd2015/kool/tools/iterm2"
)

type Request struct {
    Args       []string
    Target     bool
    TargetKind string // "directory" (default) or "file"
    Candidates []iterm2cmd.FocusCandidate
}

type Response struct {
    ExitCode int
    Stdout   string
    Stderr   string
    Focused  []string
    DiscoverCalls int
}

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
    target := d.DOCTEST_CASE
    if req.TargetKind == "file" {
        target = filepath.Join(d.DOCTEST_CASE, "not-a-directory")
        info, err := os.Stat(target)
        if err != nil { return nil, err }
        if info.IsDir() { return nil, fmt.Errorf("fixture %q must be a file", target) }
    }

    args := append([]string(nil), req.Args...)
    if req.Target { args = append(args, target) }
    fake := &iterm2cmd.FocusFake{Candidates: req.Candidates}
    var stdout, stderr bytes.Buffer
    code := iterm2cmd.RunFocusForTest(args, &stdout, &stderr, fake)
    return &Response{ExitCode: code, Stdout: stdout.String(), Stderr: stderr.String(), Focused: fake.Focused, DiscoverCalls: fake.DiscoverCalls}, nil
}
```
