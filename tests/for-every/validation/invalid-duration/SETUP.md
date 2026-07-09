# Scenario

**Feature**: duration string fails pkgs/duration.Parse rules

```
# same rules as kool timeout: ParseDuration, else bare int seconds; must be > 0
user -> kool for-every <bad-duration> …
  -> invalid duration error, non-zero, no loop
```

## Steps

1. Spaced form with a command present so the failure is duration, not missing command.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Glued = false
	req.Command = "true"
	req.MaxRuns = intPtr(1)
	return nil
}
```
