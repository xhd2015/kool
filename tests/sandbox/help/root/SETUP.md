# Scenario

**Feature**: root --help lists build and principal pack flags

```
kool sandbox --help
  -> exit 0; stdout mentions build, -o/--output, -i/--input, --file, --env, --goos, --goarch
```

## Steps

1. HelpAtRoot=true.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.HelpAtRoot = true
	req.HelpBuild = false
	req.Subcommand = ""
	return nil
}
```
