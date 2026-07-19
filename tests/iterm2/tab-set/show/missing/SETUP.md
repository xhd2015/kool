# Scenario

**Feature**: show unknown set name errors

```
show no-such-set -> Error, exit ≠ 0
```

## Steps

1. SetName that has no JSON file.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.SetName = "no-such-set"
	return nil
}
```
