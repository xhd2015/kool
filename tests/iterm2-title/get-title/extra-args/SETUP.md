# Scenario

**Feature**: get-title rejects extra positional arguments

```
# extra positional
kool iterm2 get-title foo
  -> exit 1 validation / unrecognized args
```

## Steps

1. In-session so routing reaches get-title validation.
2. ExtraArgs = `["foo"]`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.InSession = true
	req.ExtraArgs = []string{"foo"}
	return nil
}
```
