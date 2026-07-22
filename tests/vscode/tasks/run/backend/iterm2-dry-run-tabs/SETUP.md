# Scenario

**Feature**: iterm2 backend dry-run prints multi-leaf tab plan offline (never RunTabSet)

```
run "Both Steps" --backend=iterm2 --dry-run
  + mock env installed (out path empty after run)
  -> exit 0; plan maps Step One / Step Two to tabs
  -> never calls RunTabSet / never writes mock JSON
```

## Steps

1. echoCompositeJSONC; Backend=iterm2; DryRun=true; Query=`Both Steps`.
2. enableITerm2Mock so Assert can prove mock out file was **not** written.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	writeTasksJSON(t, req.WorkingDir, echoCompositeJSONC)
	req.Dir = req.WorkingDir
	req.Query = "Both Steps"
	req.Backend = "iterm2"
	req.DryRun = true
	// Install mock seam: dry-run must still never invoke / record RunTabSet.
	enableITerm2Mock(t, req)
	return nil
}
```
