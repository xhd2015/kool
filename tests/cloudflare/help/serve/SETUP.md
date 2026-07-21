# Scenario

**Feature**: serve --help documents serve flags

```
kool cloudflare serve --help
  -> exit 0; stdout mentions --domain, --url, --tunnel
```

## Steps

1. HelpServe=true.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.HelpServe = true
	return nil
}
```
