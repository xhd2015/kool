# Scenario

**Feature**: tab-set --help prints usage and config path

```
tab-set --help -> exit 0; stdout has tab-set, list, run, config path
```

## Steps

1. Request Help on tab-set.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Help = true
	return nil
}
```
