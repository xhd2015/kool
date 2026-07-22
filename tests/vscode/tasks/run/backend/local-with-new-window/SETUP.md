# Scenario

**Feature**: window flags with `--backend=local` error (not silently ignored)

```
run "Say Hello" --backend=local -n
  -> error: -n / --new-window only meaningful for iterm2 (or not valid with local)
```

## Steps

1. echoLeaf; Backend=local; NewWindow=true.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	writeTasksJSON(t, req.WorkingDir, echoLeafJSONC)
	req.Dir = req.WorkingDir
	req.Query = "Say Hello"
	req.Backend = "local"
	req.NewWindow = true
	req.DryRun = false
	return nil
}
```
