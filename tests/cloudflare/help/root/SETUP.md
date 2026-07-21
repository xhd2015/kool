# Scenario

**Feature**: root --help lists serve and tunnel flags

```
kool cloudflare --help
  -> exit 0; stdout mentions serve, --domain, --url, --tunnel
```

## Steps

1. HelpAtRoot=true.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.HelpAtRoot = true
	req.Subcommand = ""
	return nil
}
```
