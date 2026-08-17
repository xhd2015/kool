# Scenario

**Feature**: HandleWith with no args returns an example error mentioning kool with-go

```
# empty argv after with-go
HandleWith([]) -> error containing "kool with-go"
```

## Steps

1. Set `req.Args` empty.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.Args = nil
	return nil
}
```
