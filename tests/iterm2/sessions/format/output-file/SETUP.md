# Scenario

**Feature**: -o writes buffered snapshot to file; stderr reports Wrote

```
sessions snapshot -o snap.json
  -> stdout empty
  -> stderr contains Wrote
  -> file contains JSON source iterm2 / fixture id
```

## Steps

1. OutputPath=snap.json (relative; Run joins WorkingDir). Format inferred from .json.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.OutputPath = "snap.json"
	// No explicit --json: suffix .json infers FormatJSON (existing ResolveFormat).
	return nil
}
```
