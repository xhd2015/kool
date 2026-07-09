# Scenario

**Feature**: spaced for-every without a duration positional is invalid

```
kool for-every
  -> non-zero exit; stderr mentions duration or usage; no hang
```

## Steps

1. Spaced form with empty Duration and no command.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Glued = false
	req.Duration = ""
	req.Command = ""
	return nil
}
```
