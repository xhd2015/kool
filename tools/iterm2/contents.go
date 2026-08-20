package iterm2

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	lib "github.com/xhd2015/dot-pkgs/go-pkgs/shell/iterm2"
	"github.com/xhd2015/kool/pkgs/errs"
	lessflags "github.com/xhd2015/less-flags"
)

const contentsHelp = `iterm2 contents <session-id> [options]

Print currently visible iTerm2 pane text. Does not switch Desktop or select
the pane. Tries ITERM2_APP_PATH (if set), then ~/Applications/iTerm.app, then
/Applications/iTerm.app, skipping installs that are not running.

  <session-id>     iTerm2 session unique ID (UUID; ITERM_SESSION_ID ok)
  --json           emit {"session_id","app","contents"} (no ANSI)
  -o, --output FILE  write output to FILE instead of stdout
  -h, --help       show this help

Examples:
  kool iterm2 contents B95E6BAC-3104-43D2-ABAE-86FC02A669A2
  kool iterm2 contents B95E6BAC-3104-43D2-ABAE-86FC02A669A2 --json
`

type contentsJSON struct {
	SessionID string `json:"session_id"`
	App       string `json:"app"`
	Contents  string `json:"contents"`
}

func runContents(args []string, stdout, stderr io.Writer, env TestRun) error {
	var asJSON bool
	var outPath string
	remain, err := lessflags.Bool("--json", &asJSON).
		String("-o,--output", &outPath).
		HelpFunc("-h,--help", func() {}).
		HelpNoExit().
		Parse(args)
	if err != nil {
		if err == lessflags.ErrHelp {
			fmt.Fprint(stdout, strings.TrimSpace(contentsHelp)+"\n")
			return nil
		}
		WriteError(stderr, err.Error())
		return errs.NewSilenceExitCode(1)
	}
	if len(remain) == 0 {
		WriteError(stderr, "contents: missing session-id")
		fmt.Fprint(stderr, "\n"+strings.TrimSpace(contentsHelp)+"\n")
		return errs.NewSilenceExitCode(1)
	}
	if len(remain) > 1 {
		WriteError(stderr, fmt.Sprintf("contents: unexpected arguments: %s", strings.Join(remain[1:], " ")))
		return errs.NewSilenceExitCode(1)
	}
	sessionID := strings.TrimSpace(remain[0])
	if sessionID == "" {
		WriteError(stderr, "contents: missing session-id")
		return errs.NewSilenceExitCode(1)
	}

	fn := env.Contents
	if fn == nil {
		fn = lib.Contents
	}
	res, err := fn(sessionID, nil)
	if err != nil {
		msg := err.Error()
		if errors.Is(err, lib.ErrSessionNotFound) || strings.Contains(strings.ToLower(msg), "session not found") {
			WriteError(stderr, fmt.Sprintf("session not found: %s", sessionID))
			return errs.NewSilenceExitCode(1)
		}
		if strings.Contains(strings.ToLower(msg), "session id is required") {
			WriteError(stderr, "contents: missing session-id")
			return errs.NewSilenceExitCode(1)
		}
		WriteError(stderr, strings.TrimPrefix(msg, "Error: "))
		return errs.NewSilenceExitCode(1)
	}

	var body []byte
	if asJSON {
		var buf bytes.Buffer
		enc := json.NewEncoder(&buf)
		enc.SetEscapeHTML(false)
		if err := enc.Encode(contentsJSON{
			SessionID: res.SessionID,
			App:       res.App,
			Contents:  res.Contents,
		}); err != nil {
			WriteError(stderr, err.Error())
			return errs.NewSilenceExitCode(1)
		}
		body = buf.Bytes()
	} else {
		body = []byte(res.Contents)
		if len(body) > 0 && body[len(body)-1] != '\n' {
			body = append(body, '\n')
		}
	}

	if strings.TrimSpace(outPath) != "" {
		if err := os.WriteFile(outPath, body, 0644); err != nil {
			WriteError(stderr, err.Error())
			return errs.NewSilenceExitCode(1)
		}
		return nil
	}
	_, err = stdout.Write(body)
	return err
}
