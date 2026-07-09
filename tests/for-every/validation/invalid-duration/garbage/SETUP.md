# Scenario

**Feature**: garbage duration string is rejected

```
kool for-every notaduration --max-runs 1 true
  -> non-zero; stderr mentions duration / invalid
```

## Steps

1. Set Duration to a non-parseable token.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Duration = "notaduration"
	return nil
}
```
