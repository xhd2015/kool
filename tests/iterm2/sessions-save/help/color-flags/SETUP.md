# Scenario

**Feature**: save and restore help document `--color` and `--no-color`

```
Caller
  -> sessions save -h
  -> sessions restore -h
  <- both mention --color and --no-color
```

## Steps

1. ModeHelp with HelpCombinedSaveRestore (concatenated save + restore help).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Mode = ModeHelp
	req.HelpCombinedSaveRestore = true
	return nil
}
```
