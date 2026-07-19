# Scenario

**Feature**: run --dry-run prints plan without requiring iTerm

```
run bots --dry-run -> exit 0; plan mentions tabs/commands; no iTerm failure
```

## Steps

1. Write bots.json; DryRun=true; SetName=bots.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	writeBotsConfig(t, req.ConfigDir)
	req.SetName = "bots"
	req.DryRun = true
	return nil
}
```
