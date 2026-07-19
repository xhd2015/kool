# Scenario

**Feature**: tab-set run with dry-run and flag validation

```
tab-set run <name> [--dry-run] [-n] [--no-new-window]
  -> dry-run plan | flag conflict error
```

## Steps

1. Subcommand `run`.
2. Leaves set DryRun / NewWindow / NoNewWindow and fixtures.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Subcommand = "run"
	return nil
}
```
