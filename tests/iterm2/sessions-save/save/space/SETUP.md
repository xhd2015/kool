# Scenario

**Feature**: sessions save records macOS Space fields (or omits with ignore)

```
Caller
  -> sessions save [--dry-run] [--ignore-macos-space] --file …
  -> critical fixture (+ SpaceIndexForWindow unless ignore)
  <- plan space label / JSON space fields / resolve warning
```

## Steps

1. ModeSave; UseCriticalFixture (shared by space save leaves).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Mode = ModeSave
	req.UseCriticalFixture = true
	return nil
}
```
