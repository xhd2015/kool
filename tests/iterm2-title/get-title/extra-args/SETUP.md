# Scenario

**Feature**: get-title rejects extra positional arguments

```
# extra positional
kool iterm2 get-title foo
  -> exit 1 validation / unrecognized args
```

## Steps

1. In-session so routing reaches get-title validation.
2. ExtraArgs = `["foo"]`.

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InSession = true
	req.ExtraArgs = []string{"foo"}
	return nil
}
```
