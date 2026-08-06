# Scenario

**Feature**: package-table with non-existent profile file → Error exit 1

```
user -> package-table /abs/path/does-not-exist.out
  -> exit 1; stderr contains Error: and profile/path signal
```

## Steps

1. Point ProfilePath at an absolute path that does not exist under WorkingDir.

```go
import (
	"path/filepath"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Subcommand = "package-table"
	req.ProfilePath = filepath.Join(req.WorkingDir, "missing-coverage.out")
	req.ProfileSet = true
	// Ensure parent exists but file does not.
	return nil
}
```
