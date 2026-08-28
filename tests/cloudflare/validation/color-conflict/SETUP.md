# Scenario

**Feature**: `--color` and `--no-color` conflict

```
kool cloudflare serve --domain a.example.com --url http://127.0.0.1:9 --color --no-color
  -> non-zero; conflict message; no StartSession
```

## Steps

1. Serve with both color flags.

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Subcommand = "serve"
	req.DomainSet = true
	req.Domain = "a.example.com"
	req.URLSet = true
	req.URL = "http://127.0.0.1:9"
	req.Color = true
	req.NoColor = true
	req.AllowStart = false
	return nil
}
```
