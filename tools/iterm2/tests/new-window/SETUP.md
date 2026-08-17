# Scenario

**Feature**: kool iterm2 new-window flag

```
# user specifies mode via flags
kool iterm2 <dir> [-n | --new-window | -r | --reuse | --reuse-window] -> CLI handler -> lib.OpenConfig(cfg)
# library selects script builder based on Mode
cfg.Mode -> {ModeSmart, ModeReuseCurrent, ModeForceNew} -> BuildScript
# osascript runs the generated AppleScript
BuildScript -> osascript -> iTerm2
```

## Preconditions

- Root Run uses `RunForTestEnv` with `GOOS=darwin`, `Installed`, and `Osascript` hooks
- A temp dir is created for the test and passed as the CLI dir argument

## Steps

1. Build full args from Request.Args + temp dir
2. Invoke `iterm2cmd.RunForTestEnv` with hooks
3. Capture AppleScript from the `Osascript` hook

## Context

- Temp dir is created via `t.TempDir()` — automatically cleaned up
- Library and handler are from the local workspace (replace directives in go.mod)
