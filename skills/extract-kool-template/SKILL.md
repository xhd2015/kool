---
name: extract-kool-template
description: >-
  Extract a kool create scaffold template from an existing excellent project.
  Use when the user wants to turn a reference repo into a reusable kool template,
  add a new `kool create <name>` command, extract a template inside this project,
  or says "extract template", "scaffold template", "kool create template",
  or "/extract-kool-template". Always brainstorm scope with the user before
  implementing unless they explicitly say "go ahead".
---

# Extract kool create Template

Turn an excellent reference project into an embedded `kool create` template
inside **this repo** (`tools/create/`).

## When to use

- User points at an external project path and wants it scaffoldable via `kool create`
- User wants a new template name (e.g. `macos-app-go-daemon`) added to kool
- User says the source project is "excellent" and worth generalizing

## Phase 1 — Brainstorm (required unless user says "go ahead")

Discuss with the user **before writing code**. Cover:

1. **Template name** — kebab-case CLI arg (`kool create <name> <dir>`)
2. **Scope** — minimal skeleton vs full fidelity; what to keep vs strip
3. **Placeholders** — what must be parameterized at scaffold time
4. **Interactive prompts** — anything to ask on TTY (ports, bundle IDs, etc.)
5. **Post-scaffold steps** — `go mod tidy`, `bun install`, `git init`, chmod scripts
6. **Tests** — unit tests in `tools/create/create_<name>_test.go`; expected smoke command

Explicitly tell the user:

- Data models and storage layout (if any)
- Test scenarios and expected output
- How tests run (prefer rerunnable `go test ./tools/create`)

Do **not** implement until the user confirms with **"go ahead"**.

For Go templates, follow `go-best-practice kool-create` and consider tests.

## Phase 2 — Analyze source project

Explore the reference project at the path the user gives:

```
<reference-project>/
├── build system (Makefile, Package.swift, go.mod, package.json, …)
├── app/runtime code
├── scripts (dev, bundle, install, …)
├── tests
└── generated artifacts to EXCLUDE
```

Produce a keep/strip table:

| Include in template | Exclude |
|---------------------|---------|
| Source files users need | `.build/`, `node_modules/`, `*.app`, `*.dmg`, `.git/` |
| Scripts referenced in README | Test fixtures, large binaries, secrets |
| Minimal README for scaffolded project | Domain-specific integrations (unless in scope) |

Identify **hardcoded values** that become placeholders:

| Placeholder | Typical source | Example |
|-------------|----------------|---------|
| `__PROJECT_NAME__` | directory / app name | `my-cool-app` |
| `__MODULE_NAME__` | go module path | `github.com/you/repo/my-cool-app/go-pkgs` |
| Custom (`__DAEMON_NAME__`, `__BUNDLE_ID__`, `__DEFAULT_PORT__`, …) | app-specific constants | per brainstorm |

## Phase 3 — Create template directory

Add under `tools/create/<snake_case>_template/`:

```
tools/create/
├── <snake_case>_template/     # embedded source tree
│   ├── .gitignore
│   ├── README.md              # scaffolded-project docs
│   ├── go.mod.template        # if Go (renamed → go.mod on scaffold)
│   └── …                      # trimmed source files
├── create_<snake_case>.go     # handler
├── create_<snake_case>_test.go
└── create.go                  # wire dispatch + help text
```

### Placeholder rules

- Use `__PROJECT_NAME__` and `__MODULE_NAME__` for generic substitution (existing convention)
- Add custom placeholders as `__KEY__` — uppercase snake case between double underscores
- Replace via `applyPlaceholders()` in `placeholders.go`
- Replace placeholders in **both** file contents **and** directory names during copy
- Template Go files with invalid module paths: prefix with `//go:build ignore` and
  **strip that line** when copying to the user's project (`stripTemplateBuildIgnore`)

### Go module path

```go
modulePath, _ := suggestGoModPath(projectDir)
if modulePath == "" {
    modulePath = filepath.Join(projectName, "go-pkgs") // adjust per layout
} else {
    modulePath = filepath.Join(modulePath, "go-pkgs")
}
```

See `create_macos_app_go_daemon.go` and `create_go_cli.go` for patterns.

### Reuse existing helpers

| Helper | File | Purpose |
|--------|------|---------|
| `prepareProjectDir` | `create.go` | empty dir or `.git`-only |
| `suggestGoModPath` | `create.go` | derive module from git remote |
| `initGitRepo` | `create.go` | `git init` if needed |
| `copyTemplateDir` | `create_go_react.go` | `__PROJECT_NAME__` / `__MODULE_NAME__` via `applyPlaceholders` |
| `applyPlaceholders` | `placeholders.go` | shared `__KEY__` → value replacement |

Prefer a dedicated `copy*Template()` when placeholders or renames differ.

## Phase 4 — Write the handler

Create `create_<snake_case>.go`:

```go
//go:embed all:<snake_case>_template
var <camelCase>TemplateFS embed.FS

func HandleCreate<Name>(args []string) error {
    // 1. parse args, --help
    // 2. prepareProjectDir
    // 3. resolve dynamic values (ports, IDs, …)
    // 4. copyTemplate with placeholder substitution
    // 5. post-processing (rename go.mod.template, go mod tidy, chmod scripts)
    // 6. initGitRepo
    // 7. print success + next steps
}
```

Wire in `create.go`:

1. Add template to help text and usage error string
2. Add dispatch branch: `if template == "<kebab-name>" { return HandleCreate… }`

### TTY prompts (optional)

Pattern from `resolveDaemonPort()` in `create_macos_app_go_daemon.go`:

- non-TTY → sensible default (random available port, etc.)
- TTY → prompt with default in brackets; validate; retry on error

Use `golang.org/x/term.IsTerminal(int(os.Stdin.Fd()))`.

## Phase 5 — Tests (required for Go templates)

Add `create_<snake_case>_test.go` mirroring `create_go_cli_test.go`:

| Test | Assert |
|------|--------|
| Scaffold into empty dir | key files exist |
| No leftover placeholders | no `__PROJECT_NAME__`, custom `__KEY__` tags, etc. |
| `go build ./...` or `go test ./...` | passes in scaffolded project |
| Reject non-empty dir | error contains `not empty` |
| Allow `.git`-only dir | scaffold succeeds |

Run:

```bash
go test ./tools/create -count=1 -v -run '<YourTemplate>'
go test ./tools/create -count=1
```

Manual smoke (user's expected command):

```bash
go run . create <kebab-name> /tmp/test-app
```

## Phase 6 — Verify end-to-end

1. All `tools/create` tests pass
2. Smoke scaffold to `/tmp/test-app`
3. Build commands in scaffolded project work (`go build`, `swift build`, `bun install`, etc.)
4. Scripts are executable (`chmod 0755` on `script/*.sh`)

## Canonical example

`macos-app-go-daemon` — extracted from `macos-agent-sessions`:

| Artifact | Path |
|----------|------|
| Template tree | `tools/create/macos_app_go_daemon_template/` |
| Handler | `tools/create/create_macos_app_go_daemon.go` |
| Tests | `tools/create/create_macos_app_go_daemon_test.go` |

Read these files before starting a new extraction.

## Checklist (copy for each extraction)

```
[ ] Brainstorm approved ("go ahead")
[ ] Template name chosen (<kebab-case>)
[ ] Scope table: keep / strip
[ ] Placeholder list defined
[ ] tools/create/<snake_case>_template/ created
[ ] create_<snake_case>.go handler
[ ] create.go wired (help + dispatch)
[ ] create_<snake_case>_test.go
[ ] go test ./tools/create passes
[ ] Manual smoke: go run . create <name> /tmp/test-app
```

## Additional references

- Template anatomy and file conventions: [references/template-anatomy.md](references/template-anatomy.md)
- Existing templates: `frontend_template`, `server_template`, `go_cli_template`,
  `go_react/`, `electron_app_template`, `macos_app_go_daemon_template`