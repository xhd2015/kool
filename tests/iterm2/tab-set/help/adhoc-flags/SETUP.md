# Scenario

**Feature**: tab-set help documents --tab, --save, --force (ad-hoc/save)

```
tab-set --help -> mentions --tab, --save, --force (and preferably --window-name)
```

## Steps

1. Help=true (same as show-usage; stronger content asserts).

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Help = true
	return nil
}
```
