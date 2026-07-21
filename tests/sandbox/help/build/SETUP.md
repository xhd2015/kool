# Scenario

**Feature**: build --help documents pack flags

```
kool sandbox build --help
  -> exit 0; stdout documents -o/-i/--file/--env/--goos/--goarch
```

## Steps

1. HelpBuild=true.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.HelpAtRoot = false
	req.HelpBuild = true
	req.Subcommand = ""
	return nil
}
```
