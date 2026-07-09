# Scenario

**Feature**: --help lists set-title and get-title

```
kool iterm2 --help
  -> exit 0; stdout mentions set-title and get-title (and open-dir usage)
```

## Steps

1. Help=true from parent; no session required.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Help = true
	req.InSession = false
	return nil
}
```
