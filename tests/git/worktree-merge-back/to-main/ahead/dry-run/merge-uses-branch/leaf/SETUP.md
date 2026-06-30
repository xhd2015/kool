# Scenario

**Feature**: dry-run leaf asserts branch name in planned merge

```
user -> merge-back --dry-run -> merge --ff-only <branch>
```

## Steps

1. Run merge-back with `--dry-run` (configured by ancestor setup)

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	if !req.DryRun {
		t.Fatal("expected dry-run from ancestor setup")
	}
	return nil
}
```