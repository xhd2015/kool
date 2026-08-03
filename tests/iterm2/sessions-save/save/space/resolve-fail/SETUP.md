# Scenario

**Feature**: save Space resolve fail / unknown window → space 0 + warning; still write

```
Caller
  -> sessions save --file resolve-fail.json
  -> critical fixture without resolvable Space (no WindowID / inject fail)
  <- FileJSON "space": 0; stderr warning about space; exit 0
```

## Steps

1. FilePath=resolve-fail.json (live write). Fixture has no injectable
   WindowID → implementer treats missing/failed resolve as space 0 + warn.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.FilePath = "resolve-fail.json"
	return nil
}
```
