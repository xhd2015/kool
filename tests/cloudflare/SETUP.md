# Scenario

**Feature**: kool cloudflare serve — local origin → public Cloudflare hostname

```
# help
user -> kool cloudflare [--help | serve --help]
  -> usage on stdout, exit 0

# validation
user -> kool cloudflare [bad/missing args]
  -> stderr error, non-zero; no StartSession

# serve (injected)
user -> kool cloudflare serve --domain HOST --url URL [--tunnel NAME]
  -> StartSession(Domain, LocalURL, TunnelName) → print public URL → WaitSignal → Stop
```

## Preconditions

- Module root is `DOCTEST_ROOT/../..` (this tree lives at `tests/cloudflare/`).
- Product package `github.com/xhd2015/kool/tools/cloudflare` exports `Handle` /
  `HandleWith` with injectable `StartSession` and `WaitSignal` (see root
  `DOCTEST.md` DSN). Until implemented, suite is RED.
- No `cloudflared` binary, network, or `~/.cloudflared` required for doctests:
  `Run` always injects fakes.

## Steps

1. Root `Setup` creates an isolated `WorkingDir` (optional isolation).
2. Grouping/leaf `Setup` sets help/subcommand/flags and inject expectations.
3. `Run` calls `cloudflare.HandleWith` with capture buffers and inject hooks.

## Context

- Default tunnel prefix: `kool-lb-` + leftmost domain label (slugified).
- Public URL printed as `https://<domain>` (session `PublicBaseURL()`).
- Errors on stderr; help on stdout ending with `\n`.

```go
import (
	"os"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	if req.WorkingDir == "" {
		req.WorkingDir = t.TempDir()
	}
	if err := os.MkdirAll(req.WorkingDir, 0755); err != nil {
		return err
	}
	// Default: StartSession must not run unless a serve leaf opts in.
	// Leaves set AllowStart=true for lifecycle cases.
	return nil
}
```
