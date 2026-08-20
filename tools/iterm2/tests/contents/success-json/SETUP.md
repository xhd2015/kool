# Scenario

**Feature**: --json includes session_id, app, contents

```
kool iterm2 contents UUID --json
```

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
	lib "github.com/xhd2015/dot-pkgs/go-pkgs/shell/iterm2"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Args = []string{"contents", "B95E6BAC-3104-43D2-ABAE-86FC02A669A2", "--json"}
	req.Hit = lib.ContentsResult{
		SessionID: "B95E6BAC-3104-43D2-ABAE-86FC02A669A2",
		App:       "~/Applications/iTerm.app",
		Contents:  "pane",
	}
	return nil
}
```
