# Scenario

**Feature**: `--help` lists `--via-open` and user-driven open / Gatekeeper wording

```
kool iterm2 install --help
  -> exit 0
  -> stdout contains --via-open
  -> stdout mentions open and/or Gatekeeper and/or user-driven
  -> stdout ends with newline
```

## Steps

1. Inherit Help=true from parent (no extra flags).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	// Leaf: install --help only (restate for clarity in leaf SETUP).
	req.Help = true
	req.UseFakeHTTP = false
	req.DownloadDir = ""
	return nil
}
```
