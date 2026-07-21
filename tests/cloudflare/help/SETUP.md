# Scenario

**Feature**: cloudflare help mode (no tunnel)

```
# user asks for help at root or serve
user -> kool cloudflare [--help | serve --help]
  -> handler prints usage, exit 0 (no StartSession)
```

## Steps

1. Mark help branch; StartSession must not be called (AllowStart remains false).

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.AllowStart = false
	return nil
}
```
