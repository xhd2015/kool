# Scenario

**Feature**: CLI help output

```
kool iterm2 --help -> usage on stdout -> exit 0
```

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Phase = "cli"
	return nil
}
```