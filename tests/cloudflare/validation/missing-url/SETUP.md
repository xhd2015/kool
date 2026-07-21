# Scenario

**Feature**: serve without --url is invalid

```
kool cloudflare serve --domain a.example.com
  -> non-zero; message mentions url; no StartSession
```

## Steps

1. serve with Domain only.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Subcommand = "serve"
	req.DomainSet = true
	req.Domain = "a.example.com"
	req.URLSet = false
	return nil
}
```
