# Scenario

**Feature**: dry-run for ahead branch prints planned commands without mutations

```
user -> merge-back --dry-run -> planned git -C commands, no changes
```

## Steps

1. Run merge-back with `--dry-run` only

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	req.DryRun = true
	req.Remove = false
	req.ConfirmFromStdin = false
	return nil
}
```