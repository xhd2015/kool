# Scenario

**Feature**: kool iterm2 install `--via-open` (CLI → InstallViaUserOpen)

```
# help
user -> kool iterm2 install --help
  -> usage includes --via-open + open/Gatekeeper wording; exit 0

# dry-run (fake HTTP)
user -> kool iterm2 install --dry-run [--via-open] --download-dir DIR
  -> resolve latest via inject; print plan; exit 0; no zip

# validation
user -> kool iterm2 install --via-open --download-only …
  -> Error: on stderr; non-zero; no download
```

## Preconditions

- Module root is `DOCTEST_ROOT/../..` (this tree lives at `tests/iterm2-install/`).
- Package under test: `github.com/xhd2015/kool/tools/iterm2` (`RunForTest`).
- L2 inject: implementer exports `SetInstallHTTPForTest` wrapping existing
  package vars `installHTTPClient` / `installLatestURL` (see unit tests in
  `tools/iterm2/install_test.go`). Until export + `--via-open` land, suite is RED.
- No real iterm2.com / `open` / `xattr` in this suite.
- Package inject is process-global: root `Run` holds `installInjectMu` around
  inject + call (leaves must not assume parallel isolation of those vars).
- No `t.Setenv` / `t.Chdir` in leaves; `RunForTest` workingDir is always `""`.

## Steps

1. Root `Setup` ensures a per-leaf `DownloadDir` under `t.TempDir()` when the
   leaf will resolve or might write a zip.
2. Grouping/leaf `Setup` sets Help / DryRun / ViaOpen / DownloadOnly / UseFakeHTTP.
3. Root `Run` builds `install …` args, optionally injects fake latest HTTP, calls
   `RunForTest`, returns stdout/stderr/exit.

## Context

- Fake zip name default: `iTerm2-3_6_11.zip` → version string `3.6.11`.
- Dry-run must never create the zip under `DownloadDir`.
- Help stdout must end with trailing `\n`.

```go
import (
	"path/filepath"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	// Per-leaf download dir for isolation (dry-run + validation no-zip asserts).
	// Help leaves leave DownloadDir empty.
	if req.DownloadDir == "" && !req.Help {
		req.DownloadDir = filepath.Join(t.TempDir(), "dl")
	}
	if req.FinalZipName == "" {
		req.FinalZipName = "iTerm2-3_6_11.zip"
	}
	return nil
}
```
