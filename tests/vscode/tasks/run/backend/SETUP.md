# Scenario

**Feature**: run backend selection (auto / iterm2 dry-run / live mock / window flags)

```
--backend=auto     -> single fg leaf => local exec (offline-safe)
--backend=iterm2 --dry-run multi leaves -> tab plan (no RunTabSet)
--backend=iterm2 live + KOOL_VSCODE_TASKS_ITERM2_MOCK*
  -> RunTabSet record JSON; modes smart|new-window|no-new-window
  -> mock err => fail closed (no local multi fallback)
-n + --no-new-window -> error
--backend=local -n -> error
```

## Steps

1. Default DryRun=false for real-run backend leaves; iterm2 dry-run leaf overrides.
2. Leaves set Backend, window flags, fixtures; live leaves enable mock env.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	// Real-run default for backend leaves; iterm2-dry-run-tabs sets DryRun=true.
	req.DryRun = false
	return nil
}
```
