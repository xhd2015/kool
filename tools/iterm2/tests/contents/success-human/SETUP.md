# Scenario

**Feature**: human stdout is raw pane text

```
kool iterm2 contents UUID -> pane + newline
```

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
	lib "github.com/xhd2015/dot-pkgs/go-pkgs/shell/iterm2"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Args = []string{"contents", "B95E6BAC-3104-43D2-ABAE-86FC02A669A2"}
	req.Hit = lib.ContentsResult{
		SessionID: "B95E6BAC-3104-43D2-ABAE-86FC02A669A2",
		App:       lib.CanonicalITermAppHome,
		Contents:  "❯ hello",
	}
	return nil
}
```
