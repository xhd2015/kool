# Go Best Practice Review — `kool`

**Scope:** codebase structure, CLI design, flag handling, package layout  
**Baseline:** [go-best-practice](https://github.com/xhd2015) skill topics — `flags-parsing`, `flags-parsing/subcommand`, `flags-parsing/types`, `flags-parsing/cut`, `flags-parsing/collect`, `cli` / `cli/dry-run` / `cli/color`, `cmd-exec`, `go-embed-assets`, `kool-create`  
**Date:** 2026-08-06  
**Module:** `github.com/xhd2015/kool` (Go 1.25.10)

This is a **review only** — no functional fixes implemented. Findings are ordered by severity.

---

## Executive summary

`kool` is a mature multi-command Swiss-army CLI: a root `main` switch dispatches into many `tools/*` packages, with solid doctest coverage under `tests/`. Newer surfaces (`for-every`, `modules`, `create go-cli`, parts of `iterm2`) already follow **less-flags**, **subcommand help**, and **single-path dry-run** well.

The main gaps are **inconsistent adoption** of those patterns across older commands, a **dual flag-parsing stack**, **incomplete embed/hydrate story** for UI assets, **heavy use of raw `exec.Command`**, and **package-layout debt** (root-level handlers, versioned cmd copies, dead/orphan templates).

| Area | Verdict |
|------|---------|
| less-flags | Adopted widely (~57 files); still mixed with `pkgs/flag` + hand loops |
| Subcommand `--help` | Good on newer cmds; many older handlers still error-only on empty args |
| Dry-run | Strong in `go modules` / tasks / iterm2; uneven elsewhere |
| cmd-exec (`xgo/support/cmd`) | Used well for `go` / install / create; timeout/watch/port often raw |
| go:embed assets | Layer 1 placeholders present; Layer 4 hydrate largely missing |
| kool create | Templates work; skill docs lag code; orphan `server_go_db_template` |

---

## Architecture snapshot

```text
main.go, vscode.go, snippet.go     # root package main (dispatch + some handlers)
tools/<feature>/                   # command implementations (Handle entrypoints)
pkgs/                              # shared helpers (flag, duration, terminal, release, …)
cmd/kool-with-go1.{18,19,24,25}/   # versioned bootstrap binaries (duplicated trees)
script/                            # build-react, install, release, gen-kool
tests/                             # doctest trees (cloudflare, git, iterm2, …)
tools/create/*_template/           # embedded scaffolds for `kool create`
```

**Dispatch model:** root `handle()` in `main.go` is a large `switch` on `args[0]`, then each tool owns nested switches. This matches a multi-skill host shape more than a single cobra tree — fine for this product, but it amplifies the need for **help at every level** (`flags-parsing/subcommand`).

---

## Findings (by severity)

### Critical / high

#### H1. Dual flag stacks: `pkgs/flag.ParseFlag` vs `less-flags`

**Topic:** `flags-parsing`, `flags-parsing/types`

**Evidence:**
- Canonical less-flags usage: `tools/for-every`, `tools/timeout`, `tools/watch`, `tools/go/*`, many git subcommands
- Manual `pkgs/flag.ParseFlag` loops: `tools/cloudflare`, `tools/http`, `tools/ssh`, `tools/port/check_ready`
- Hand-written argv loops without either helper: `tools/git/worktree/merge_back.go`, `tools/git/worktree/reclaim.go`, `tools/vscode/tasks/handle.go` (`parseFlags`), `tools/iterm2/sessions_checkpoint.go` (`parseSaveRestoreFlags` with comment *“Manual parse to keep less-flags optional”*)

**Why it matters:** Users get inconsistent behavior for `--flag=value`, unknown flags, and `-h`/`--help`. Maintainers re-implement type conversion (`strconv`, duration) that less-flags already covers (`Duration`, `Int`, `Bool`, `String`).

**Recommended change:**
1. Treat `github.com/xhd2015/less-flags` as the **only** CLI flag parser for user-facing commands.
2. Migrate remaining `pkgs/flag.ParseFlag` call sites; then delete or demote `pkgs/flag` to an internal adapter only if still needed by non-CLI code.
3. Prefer `Help("-h,--help", …)` + `StopOnFirstArg()` for dispatchers (`flags-parsing/subcommand`).

---

#### H2. Subcommand help is incomplete across the tree

**Topic:** `flags-parsing/subcommand`, `cli`

**Policy from skill:** every level must answer `-h`/`--help` with **that level’s** usage; empty args at a dispatch node should print help (friendly default), not only a terse error.

**Evidence of good practice:**
- `tools/cloudflare`, `tools/ssh`: root + subcommand help constants
- `tools/timeout`, `tools/for-every`: less-flags `Help` + clear usage
- `tools/vscode/tasks`: empty args → print usage

**Evidence of gaps:**
| Surface | Empty / missing help behavior |
|---------|-------------------------------|
| Root `kool` with no args | Error: `requires command, try 'kool --help'` (acceptable) |
| `kool create` (no args) | Error listing templates — does **not** print full `help` |
| `kool git` (no args) | Error: partial command list; omits staged/grep/compare-branch/worktree details |
| `kool go` (no args) | Error string of commands; `-h` works, empty does not print help |
| `kool http` | No help text at all; no `--help` case on serve |
| `kool yaml2json` / `json2yaml` / `html2text` / `html2md` / `uuid` / `react` / `github` / `rules` | Handlers without Help/`--help` wiring |
| Root help constant | Large but incomplete vs real switch (e.g. limited coverage of bash/sandbox/service/encoding); typos (`facalitate`, `splited`) |

Root help also does not systematically say: *“Run `kool <cmd> --help` for command-specific options.”* (recommended in subcommand recipe).

**Recommended change:**
1. Checklist for every `Handle`:
   - `-h`/`--help` at this level
   - empty args at pure dispatch nodes → print that level’s help (exit 0)
   - help text lists subcommands + flags actually implemented
2. Align root `help` constant and `README.md` with the real command tree (or generate help from a registry).
3. Prefer less-flags `Help` over ad-hoc `if arg == "--help"`.

---

#### H3. Package layout: root `main` is overloaded; package names shadow stdlib

**Topic:** general Go layout (skill-adjacent: multi-command host structure)

**Evidence:**
- Root package files: `main.go` (~390 lines switch), `vscode.go` (~386 lines), `snippet.go` — VS Code handling lives in `main` while tasks live under `tools/vscode/tasks`
- Packages named after stdlib: `tools/http` (`package http`), `tools/fs` (`package fs`) — force import aliases and confuse readers
- `tools/go` is `package go_tools` (necessary because `go` is reserved) but directory vs package name diverge
- `tools/for-each-dir/main.go` is **not** `package main`; filename is misleading
- `tools/cmd` is a thin wrapper around `xgo/support/cmd` — easy to confuse with “command packages”
- Four near-duplicate trees: `cmd/kool-with-go1.18` … `1.25` (~2.5k lines total) plus `script/gen-kool/kool-template` — drift risk

**Recommended change:**
1. Move `handleVscode` / snippet handlers into `tools/vscode` and `tools/snippet` (root `main` = dispatch + exit-code only).
2. Rename shadowing packages: e.g. `tools/httpserve` or `tools/statichttp`, `tools/fsutil`.
3. Rename `for-each-dir/main.go` → `for_each_dir.go` (or similar).
4. Collapse versioned `kool-with-go*` to **one** generated tree from `script/gen-kool` (or embed version as build input) so bugfixes are not four-way cherry-picks.

---

#### H4. UI embed: Layer 1 OK, completeness / hydrate incomplete

**Topic:** `go-embed-assets`

**What is already good:**
- `tools/web/react/dist/placeholder.txt` and `tools/preview/viewer/react/dist/placeholder.txt` keep `//go:embed all:react/dist` compilable for bare `go install`
- `script/install` + `script/build-react` implement fat local bundle (Layer 3 path)
- README documents plain `go install` as “no web ability”

**Gaps vs skill checklist:**
| Layer | Status |
|-------|--------|
| 1 placeholder in git | Present |
| gitignore un-ignore placeholder | **Inconsistent**: preview has `!dist/placeholder.txt`; `tools/web/react/.gitignore` ignores `dist` without un-ignore |
| Completeness check (`EmbedComplete`) | Not found as a shared helper |
| Runtime hydrate (cache + version-pinned release archive) | Not present for preview/web |
| Version pin + release asset naming | Partial release tooling under `pkgs/release` / scripts; not wired as asset hydrate for UI |

**go-react scaffold concern (`kool-create` + embed):**
- Template embeds `__PROJECT_NAME__-react/dist` with no committed placeholder strategy documented in the scaffold for *generated* projects
- Bare build of a freshly created go-react project will fail embed or ship empty UI until frontend is built — skill expects placeholders + optional hydrate for products others `go install`

**Recommended change:**
1. Align both frontend gitignores with Layer 1 (`dist/**` + `!dist/placeholder.txt`).
2. Add `EmbedComplete` (index.html + assets, not only placeholder) and fail/clearly degrade in serve paths.
3. For published binaries that need UI after thin install: implement Layer 4 hydrate (version-pinned tar.gz under release assets) **or** keep the documented “no web on go install” contract and make runtime errors explicit when embed is incomplete.
4. Teach `kool create go-react` to leave a `dist/placeholder.txt` so generated modules always compile.

---

### Medium

#### M1. External commands: mixed `xgo/support/cmd` vs raw `exec.Command`

**Topic:** `cmd-exec`

**Good:**
- `main` upgrade, `tools/go` rebuild/run, `tools/create` go-cli tidy, `script/install` use `cmd.Debug().Run` / `cmd.Output`
- Thin helper `tools/cmd.Debug(bool)` exists

**Inconsistent / prefer cmd-exec:**
- `tools/timeout`, `tools/watch`, `tools/for-every`, `tools/for-each-dir`, `tools/port/*` — `os/exec` directly
- `tools/preview/viewer` docker probes via `exec.Command`
- Versioned `cmd/kool-with-go*` mix `cmd.Debug` and raw `exec.Command` for nm/objdump

**Why it matters:** Debug visibility (`[cmd] …`), consistent env/dir, and stderr inheritance differ by call site. timeout/watch need process groups/signals — raw exec is sometimes justified, but capture helpers (`cmd.Output`) should still be preferred for fire-and-forget probes.

**Recommended change:**
- Default to `cmd.Debug()` / `cmd.Output` for non-interactive, non-PTY invocations.
- Keep raw `exec` only where you need custom `SysProcAttr`, PTY, or signal semantics — document why.
- Consider expanding `tools/cmd` as the project wrapper (env defaults, debug toggle) so call sites do not import two styles.

---

#### M2. Dry-run quality varies; some side-effect commands lack it

**Topic:** `cli/dry-run`

**Aligned with “one path, gate side effects”:**
- `tools/go/modules/update_local_deps.go` — `if opts.DryRun { plan… }` then live apply (excellent; covered by tests)
- `tools/vscode/tasks` — expand plan then `--dry-run` prints without execute
- `tools/iterm2` sessions save/restore & tab-set — dry-run flags present

**Weaker / missing:**
- `tools/git/worktree` dry-run uses manual flag loops (works if library gates side effects; still parse inconsistency)
- `kool kill-port` always kills (`kill -9`) with no `--dry-run`
- `kool create` always writes + may `npm`/`bun`/`go mod tidy` — no dry-run / no “print plan”
- `kool go vendor link|unlink` — no dry-run for filesystem changes
- Output prefix is inconsistent (`dry-run:` vs `[dry-run]` skill convention)

**Recommended change:**
1. For destructive/mutating commands, add `--dry-run` on the **same** code path.
2. Standardize plan lines as `[dry-run] …` on stdout; warnings on stderr.
3. Do not add parallel `handleDryRun()` implementations.

---

#### M3. less-flags advanced APIs underused (`Cut`, `CollectParsedFlags`)

**Topic:** `flags-parsing/cut`, `flags-parsing/collect`

**Evidence:** No production uses of `lessflags.Cut` or `CollectParsedFlags` found. Command-forwarding is done with `StopOnFirstArg()` + remainder slices (`timeout`, `watch`, `for-every`) — often correct — or by re-parsing manually.

**Opportunities:**
- `kool watch` / `timeout` / `for-every` / `with-go` style “run this opaque command” → document as StopOnFirstArg pattern (already used) or Cut when a marker flag is desired
- Parent flags that must be stripped before child forward → `CollectParsedFlags` + `Remove` + `Reconstruct` (e.g. debug wrappers around `go run`)

**Recommended change:** When adding wrappers that both parse own flags and forward argv, prefer collect/reconstruct over ad-hoc filtering.

---

#### M4. `kool create` surface vs skill / dead template weight

**Topic:** `kool-create`

**Code supports:** `react`, `go-cli`, `macos-app-go-daemon`, `go-react`, `frontend`, `server`, `electron`.

**Skill topic `kool-create` documents:** `react`, `go-react`, `frontend`, `server`, `electron` — **omits** `go-cli` and `macos-app-go-daemon`.

**Orphan weight:** `tools/create/server_go_db_template/` is a full multi-package server (dao, session, metrics, …) with module path hard-coded to kool’s tree, **not** wired into `create.Handle`. It inflates the module and confuses “what can I scaffold?”.

**Other notes:**
- Create uses embed correctly for live templates; post-create hooks use `cmd.Debug` (good).
- Empty `kool create` should print the long help (see H2).
- Placeholder replacement (`__PROJECT_NAME__`, `__MODULE_NAME__`) is clear in `placeholders.go`.

**Recommended change:**
1. Wire `server-go-db` into `kool create` **or** move the tree out of this module.
2. Update skill/docs/help lists together when templates change.
3. Keep go-cli template’s less-flags skeleton as the gold standard for new CLIs.

---

#### M5. Color / streaming conventions are local, not project-wide

**Topic:** `cli/color`, `cli/streaming`

**Evidence:**
- `iterm2` sessions/snapshot implement `--color` / `--no-color` with `NO_COLOR` and TTY detection — matches skill
- Most other tools ignore color policy
- No shared `pkgs/color` or forced-color helper for the whole CLI

**Recommended change:** Extract iterm2’s color resolution into `pkgs/terminal` (or similar) and reuse for any ANSI output. Optional root `--color`/`--no-color` only if many commands need it; otherwise per-command is fine if the resolution helper is shared.

---

#### M6. Stale / incorrect help strings in active commands

**Topic:** `cli`, `flags-parsing/subcommand`

Examples:
- `tools/go/run/run.go` help still talks about `create <name>` project scaffolding — actual handler is debug/run wrapper for `go run`
- `tools/git` help omits real commands: `staged`, `grep`, `compare-branch`, `check-merged`, `check-merge`, …
- `tools/encoding/encode.go`: `lessflags.String("--verbose", &verbose)` with `verbose bool` — wrong less-flags method (`Bool` expected); help does not document `--verbose`
- Root help vs `README.md` categories drift (README still centers string/json/git/go/net)

**Recommended change:** Treat help strings as product surface: review in the same PR as flag changes; add a lightweight test or doctest that `--help` exits 0 for each registered command.

---

### Low / nice-to-have

#### L1. Root dispatch aliases and legacy entrypoints

Top-level aliases `go-replace`, `go-update`, `go-resolve` duplicate `kool go …`. Useful for muscle memory; document as deprecated aliases or keep and list under help “Legacy”.

#### L2. Knowledge `kool ?` is a tiny hard-coded topic tree

Fine as a easter egg; not a skill-cli packaging shape. If grown, consider `cli/skill-cli` patterns (embed TOPIC.md tree).

#### L3. `tools/go/example` embeds parse-flag template as sample

Good internal docs; ensure it demonstrates current less-flags idioms (it largely does).

#### L4. README install story is clearer than in-binary help

README correctly distinguishes curl install, script/install (with frontend), and thin `go install`. Consider pointing `kool help` at the same matrix (one paragraph).

#### L5. Doctest investment is a strength

`tests/` for cloudflare, for-every, git worktree, iterm2, sandbox, vscode is exemplary for a multi-tool CLI. Use the same pattern when migrating flag parsers (regression on `--help` and happy paths).

---

## What is already strong (keep doing)

1. **less-flags as default** in newer code (`for-every`, modules, create go-cli template, timeout, watch).
2. **Injectable handlers** for tests (`cloudflare.HandleWith`, `ssh.HandleWith`) — excellent for doctests.
3. **Dry-run single-path** in `go modules update-local-deps` with thorough tests.
4. **Embed placeholders** so thin `go install` still compiles.
5. **install path** that builds React then `go install` for fat binaries (`script/install` + `build-react`).
6. **Exit-code aware errors** (`SilenceExitCode` for timeout 124 / interrupt 130) — good CLI UX.
7. **Shared duration parsing** (`pkgs/duration`) between timeout and for-every.
8. **Create templates** use embed FS + placeholder rewrite + `go mod tidy` via cmd-exec.

---

## Recommended change backlog (grounded in topics)

Suggested implementation order (still review-only here):

| Priority | Work item | Topics |
|----------|-----------|--------|
| P0 | Migrate `http`, `ssh`, `cloudflare`, `port/check_ready` from `pkgs/flag` → less-flags; delete dead parser when unused | `flags-parsing` |
| P0 | Fix help gaps: `create` empty→help, `git`/`go`/`http` help completeness, root “`<cmd> --help`” line | `flags-parsing/subcommand` |
| P1 | Align embed gitignore + completeness checks; decide hydrate vs documented thin-install | `go-embed-assets` |
| P1 | Prefer `xgo/support/cmd` for non-special process runs; document exceptions | `cmd-exec` |
| P1 | Add `--dry-run` to destructive ops (`kill-port`, vendor link/unlink); standardize `[dry-run]` prefix | `cli/dry-run` |
| P2 | Move `vscode.go` / snippet out of package main; rename `http`/`fs` packages | layout |
| P2 | Generate or single-source `cmd/kool-with-go*` | layout / maintainability |
| P2 | Wire or remove `server_go_db_template`; sync kool-create skill with `go-cli` / `macos-app-go-daemon` | `kool-create` |
| P3 | Shared color helper; optional Cut/Collect where forwarding gets complex | `cli/color`, `flags-parsing/cut|collect` |
| P3 | Fix stale `go run` help; encode `--verbose` as `Bool` | `flags-parsing/types` |

---

## Per-topic scorecard

| Topic | Adoption | Notes |
|-------|----------|-------|
| `flags-parsing` | **Partial** | Majority less-flags; dual stack remains |
| `flags-parsing/subcommand` | **Partial** | Newer cmds good; many empty-arg errors |
| `flags-parsing/types` | **Good** | Duration/string slice/bool used; some manual int/duration |
| `flags-parsing/cut` | **Unused** | StopOnFirstArg covers most “run command” cases |
| `flags-parsing/collect` | **Unused** | Opportunity for debug/wrapper argv filtering |
| `cli/dry-run` | **Partial** | Excellent modules/tasks; gaps on destructive tools |
| `cli/color` | **Local** | Strong only in iterm2 |
| `cli/streaming` | **N/A / low** | Most tools stream by default via inherited stdio |
| `cmd-exec` | **Partial** | Good in go/create/install; timeout/watch/port raw |
| `go-embed-assets` | **Partial** | Layer 1+3 present; Layer 4 + completeness weak |
| `kool-create` | **Good** | Productive templates; docs/orphan template cleanup needed |

---

## Suggested gold-standard command shape (for new work)

Use this when adding or rewriting a `kool` subcommand:

```go
const help = `
Usage: kool example [OPTIONS] <args>

Options:
  --dry-run     print plan without applying
  -h,--help     show this help

Run kool example --help for this message.
`

func Handle(args []string) error {
    var dryRun bool
    remain, err := lessflags.
        Bool("--dry-run", &dryRun).
        Help("-h,--help", help).
        StopOnFirstArg(). // if nested dispatch
        Parse(args)
    if err != nil {
        return err
    }
    if len(remain) == 0 {
        fmt.Print(strings.TrimPrefix(help, "\n")) // or require args with clear error
        return nil
    }
    // one pipeline; gate side effects with dryRun
    // external tools: cmd.Debug().Dir(dir).Run(...)
    return nil
}
```

Nested dispatch: parse global flags with `StopOnFirstArg`, switch on `remain[0]`, each case runs its own less-flags parse + help.

---

## Out of scope / not evaluated deeply

- Full security review of sandbox / crypto paths
- Performance of gitops / iterm2 snapshot pipelines
- Whether every doctest tree still passes CI
- Upstream quality of dependencies (`less-flags`, `xgo`, `gitops`, `dot-pkgs`)

---

## Conclusion

`kool` already embodies many go-best-practice ideas in its **newer** surfaces (less-flags, injectable handlers, single-path dry-run, embed placeholders, create scaffolds). The review pressure is on **consistency**: one flag library, help at every level, cmd-exec by default, embed completeness/hydrate policy, and a leaner package layout without versioned copies and orphan templates.

Addressing **H1–H4** would bring the project in line with its own skill documentation and make the large command tree safer to extend.
`
