# Scenario

**Feature**: vscode tasks run — dry-run plan, match rules, local/iterm2/auto backends (live RunTabSet mock)

```
run <label> --dry-run              -> plan | validation error
run <label> [--backend=…]          -> execute via local | iterm2 | auto
run <label> --backend=iterm2 --dry-run -> offline tab plan (no RunTabSet)
run <label> --backend=iterm2       -> RunTabSet (or mock env in CI)
```

## Steps

1. Subcommand `run`.
2. Leaves set Query, DryRun, Backend, window flags, fixtures, mock env.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Subcommand = "run"
	return nil
}
```
