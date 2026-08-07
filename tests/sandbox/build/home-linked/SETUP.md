# Scenario

**Feature**: build with --home-linked seals home_linked into the pack

```
# home-linked build branch
user -> kool sandbox build -o OUT --home-linked [--file|--env]...
  -> exit 0; sealed binary at OUT (home_linked bit packed)
```

## Steps

1. Enable `HomeLinked` so build argv includes `--home-linked`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.HomeLinked = true
	return nil
}
```
