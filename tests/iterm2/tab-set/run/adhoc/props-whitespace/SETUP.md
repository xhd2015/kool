# Scenario

**Feature**: --tab props block allows arbitrary spaces around [ ] and key=value

```
run scratch --tab "  [ id = a ]  echo hi" --dry-run
  -> id=a command=echo hi; exit 0
```

## Steps

1. Single --tab with leading/trailing spaces and spaces inside props.
2. DryRun=true.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.SetName = "scratch"
	req.DryRun = true
	req.Tabs = []string{"  [ id = a ]  echo hi"}
	return nil
}
```
