# Scenario

**Feature**: config-mode --dry-run marks no_submit tabs in the plan

```
run staged --dry-run
  -> exit 0; plan marks tab with no_submit; default tab not marked
```

## Steps

1. Write fixture with tab a (omit no_submit) and tab b (`no_submit: true`).
2. SetName=staged; DryRun=true.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

const dryRunNoSubmitJSON = `{
  "version": 1,
  "window_name": "local-staged",
  "tabs": [
    {"id": "a", "name": "a", "command": "echo a"},
    {"id": "b", "name": "b", "command": "echo staged", "no_submit": true}
  ]
}
`

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d

	writeConfigFile(t, req.ConfigDir, "staged", dryRunNoSubmitJSON)
	req.SetName = "staged"
	req.DryRun = true
	return nil
}
```
