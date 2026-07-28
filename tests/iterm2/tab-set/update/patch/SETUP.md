# Scenario

**Feature**: update field-patch on an existing tab (no confirm; write immediately)

```
existing bots.json
  -> update bots --tab-id <id> [--no-submit|--submit|--command|--name|--cwd|--clear-cwd]
  -> short change summary; file patched; other tabs unchanged
```

## Steps

1. Leaves write fixtures (bots or staged-with-no_submit).
2. Set TabID + the single patch flag(s) under test.
3. No Force (field patch does not confirm).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	// Leaves set SetName / TabID / patch flags and fixtures.
	return nil
}
```
