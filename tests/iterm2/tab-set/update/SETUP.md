# Scenario

**Feature**: `kool iterm2 tab-set update` mutates named config JSON only

```
# field patch / create / rm / window_name
KOOL_ITERM2_TAB_SET_DIR=<tmp>
  -> kool iterm2 tab-set update <name> [--tab-id …] [flags]
  -> load → mutate → validate → write (never RunTabSet)

# dry-run
update … --dry-run -> plan on stdout; file unchanged
```

## Steps

1. Inherit root temp `ConfigDir` / `WorkingDir`.
2. Set `Subcommand=update`.
3. Leaves set `SetName`, fixtures, and update flags (`TabID`, `Rm`,
   `Command`, `UpdateNoSubmit`, `Force`, `DryRun`, `WindowName`, …).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Subcommand = "update"
	return nil
}
```
