# Scenario

**Feature**: bare root with no subcommand is invalid

```
kool cloudflare
  -> non-zero; stderr suggests subcommands / help / usage
```

## Steps

1. Empty Subcommand; no help flags.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Subcommand = ""
	return nil
}
```
