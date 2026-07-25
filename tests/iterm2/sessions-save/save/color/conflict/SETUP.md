# Scenario

**Feature**: `--color` and `--no-color` together are rejected

```
Caller
  -> sessions save --dry-run --color --no-color
  <- non-zero; message “cannot be specified together”
```

## Steps

1. Color=true and NoColor=true.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Color = true
	req.NoColor = true
	// No fixture required for flag parse conflict, but parent sets dry-run fixture;
	// conflict should fail before capture matters.
	req.UseCriticalFixture = false
	return nil
}
```
