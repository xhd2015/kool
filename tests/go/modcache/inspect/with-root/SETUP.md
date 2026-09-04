# Scenario

**Feature**: --root surfaces upgrade suggestions for local requires behind cache-newest

```
cache: foo v1.0.0 + v1.2.0
repo go.mod requires foo v1.0.0
inspect --root repo --json -> suggestion v1.0.0 -> v1.2.0
```

```go
import (
	"path/filepath"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.JSON = true
	req.NoCache = true
	writeExtracted(t, req.ModCache, "example.com/foo", "v1.0.0", "old\n")
	writeExtracted(t, req.ModCache, "example.com/foo", "v1.2.0", "new\n")

	src := filepath.Join(req.WorkingDir, "src")
	writeFile(t, filepath.Join(src, "go.mod"), ""+
		"module example.com/app\n"+
		"\n"+
		"go 1.22\n"+
		"\n"+
		"require example.com/foo v1.0.0\n")
	writeFile(t, filepath.Join(src, "go.sum"), ""+
		"example.com/foo v1.0.0 h1:abc=\n"+
		"example.com/foo v1.0.0/go.mod h1:def=\n")
	initGitRepo(t, src)
	req.Roots = []string{src}
	return nil
}
```
