# Scenario

**Feature**: unknown subcommand is rejected

```
kool cloudflare nosuch
  -> non-zero; stderr indicates unknown / unrecognized command
```

## Steps

1. Subcommand = nosuch.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Subcommand = "nosuch"
	return nil
}
```
