# kool create Template Anatomy

## How scaffolding works

```
kool create <template> <dir>
       │
       ▼
  create.go dispatch
       │
       ▼
  HandleCreate*() in create_<name>.go
       │
       ├── prepareProjectDir(dir)
       ├── resolve placeholders (__PROJECT_NAME__, __MODULE_NAME__, …)
       ├── fs.WalkDir(embed.FS, templateRoot, …)
       ├── substitute placeholders in paths + contents
       ├── post-process (go.mod rename, go mod tidy, npm/bun install)
       └── initGitRepo(dir)
```

Templates are **embedded** via `//go:embed` — they ship inside the kool binary.

## Existing templates

| CLI name | Template dir | Handler |
|----------|--------------|---------|
| `frontend` | `frontend_template/` | `create.go` (inline) |
| `server` | `server_template/` | `create.go` (inline) |
| `electron` | `electron_app_template/` | `create.go` (inline) |
| `react` | (via create_react.go) | `create_react.go` |
| `go-cli` | `go_cli_template/` | `create_go_cli.go` |
| `go-react` | `go_react/` | `create_go_react.go` |
| `macos-app-go-daemon` | `macos_app_go_daemon_template/` | `create_macos_app_go_daemon.go` |

## Naming conventions

| Concept | Convention | Example |
|---------|------------|---------|
| CLI template name | kebab-case | `macos-app-go-daemon` |
| Template directory | snake_case + `_template` | `macos_app_go_daemon_template` |
| Handler file | `create_<snake_case>.go` | `create_macos_app_go_daemon.go` |
| Test file | `create_<snake_case>_test.go` | `create_macos_app_go_daemon_test.go` |
| Embed variable | `<camelCase>TemplateFS` | `macOSAppGoDaemonTemplateFS` |

## Standard placeholders

| Placeholder | Set from | Used in |
|-------------|----------|---------|
| `__PROJECT_NAME__` | `filepath.Base(projectDir)` | dirs, Package.swift, scripts, README |
| `__MODULE_NAME__` | `suggestGoModPath` + suffix | go.mod, imports |

Add project-specific placeholders as needed (`__DAEMON_NAME__`, `__BUNDLE_ID__`,
`__DEFAULT_PORT__`, `__STATE_SUBPATH__`, …). All use `__KEY__` format and are
replaced by `applyPlaceholders()` in `placeholders.go`.

## Files that must NOT be in templates

- Build artifacts: `.build/`, `dist/`, `node_modules/`
- Bundled apps: `*.app`, `*.dmg`, `*.zip`
- VCS: `.git/`
- Secrets: `.env`, credentials, API keys
- Large binaries unless essential and small

## Go templates with placeholder imports

Template `.go` files cannot compile inside kool's module. Use:

```go
//go:build ignore

package main

import "__MODULE_NAME__/server"
```

Strip `//go:build ignore` when copying to the user's project.

## Post-scaffold mutations

| Template type | Typical post-steps |
|---------------|-------------------|
| Go server/cli | `go.mod.template` → `go.mod`, `go mod tidy` |
| Frontend | `bun install` |
| Electron | `npm install` |
| Shell scripts | `chmod 0755 script/*.sh` |
| All | `git init` (if no `.git`) |

## Test patterns

See `create_go_cli_test.go` and `create_macos_app_go_daemon_test.go`.

Minimum assertions:

1. Expected file tree exists
2. No raw placeholders remain
3. Generated Go code compiles
4. `prepareProjectDir` edge cases (non-empty, git-only)