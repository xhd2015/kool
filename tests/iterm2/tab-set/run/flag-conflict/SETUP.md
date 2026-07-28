# Scenario

**Feature**: -n and --no-new-window together is an error

```
run bots -n --no-new-window -> Error, exit 1
```

## Steps

1. bots fixture present; NewWindow + NoNewWindow both true.
2. DryRun optional (conflict should fail before run).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d

	writeBotsConfig(t, req.ConfigDir)
	req.SetName = "bots"
	req.NewWindow = true
	req.NoNewWindow = true
	req.DryRun = true // fail at flag parse even without iTerm
	return nil
}
```
