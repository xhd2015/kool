# Scenario

**Feature**: kool iterm2 root help indexes sessions snapshot and session status

```
# user requests iterm2 root help
user -> kool iterm2 -h|--help
  -> usage lists sessions snapshot and session <id> status; exit 0; no capture
```

## Steps

1. Mode=help; HelpArgs = `["-h"]` so Run invokes root iterm2 help (not sessions -h).

## Context

- P5 polish: production root help already lists both command surfaces
  (`sessions snapshot`, `session <id> status`); this leaf locks that index.
- Current product gap (Classic RED): root help stdout may omit the trailing
  newline required of other help leaves — implementer should append `\n`.
- Flag details (`--no-stream`, `--no-enrich`, `--no-tree`, formats) remain
  asserted on sessions/session subcommand help (P3/P4 leaves).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Mode = ModeHelp
	// Root iterm2 help (not sessions -h).
	req.HelpArgs = []string{"-h"}
	return nil
}
```
