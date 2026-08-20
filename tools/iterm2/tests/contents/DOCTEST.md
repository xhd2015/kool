# kool iterm2 contents

CLI for `kool iterm2 contents <session-id>`.

## Version

0.0.1

## DSN

- User invokes `kool iterm2 contents <id> [--json] [-o FILE]`
- Handler calls `lib.Contents` (injectable)
- Human stdout is raw pane text; `--json` is `{session_id,app,contents}`

## How to run

```sh
doctest test ./tools/iterm2/tests/contents
```

```go
import (
	"bytes"
	"fmt"
	"testing"

	"github.com/xhd2015/doctest/session"
	lib "github.com/xhd2015/dot-pkgs/go-pkgs/shell/iterm2"
	iterm2cmd "github.com/xhd2015/kool/tools/iterm2"
)

type Request struct {
	Args []string
	// Hit is the fake Contents result when not NotFound.
	Hit lib.ContentsResult
	NotFound bool
}

type Response struct {
	ExitCode int
	Stdout   string
	Stderr   string
}

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	_ = t
	_ = d
	var stdoutBuf, stderrBuf bytes.Buffer
	exit := iterm2cmd.RunForTestEnv(req.Args, &stdoutBuf, &stderrBuf, iterm2cmd.TestRun{
		GOOS: "darwin",
		Contents: func(sessionID string, cfg *lib.ContentsConfig) (lib.ContentsResult, error) {
			if req.NotFound {
				return lib.ContentsResult{}, fmt.Errorf("%w: %s", lib.ErrSessionNotFound, sessionID)
			}
			got := req.Hit
			if got.SessionID == "" {
				got.SessionID = sessionID
			}
			return got, nil
		},
	})
	return &Response{
		ExitCode: exit,
		Stdout:   stdoutBuf.String(),
		Stderr:   stderrBuf.String(),
	}, nil
}
```
