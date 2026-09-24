package create

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// copyAgentCLIBackend renders the go-react-agent-cli backend template into a
// temp dir with the given project name.
func copyAgentCLIBackend(t *testing.T, project string) string {
	t.Helper()
	dir := t.TempDir()
	if err := copyTemplateDir(goReactAgentCLITemplateFS, "go_react_agent_cli/backend", dir, project, "example.com/"+project); err != nil {
		t.Fatal(err)
	}
	return dir
}

func TestGoReactAgentCLITemplateShipsAgentCLI(t *testing.T) {
	dir := copyAgentCLIBackend(t, "demo")

	// The agent CLI files exist.
	for _, name := range []string{
		filepath.Join("cmd", "demo", "main.go"),
		filepath.Join("run", "run.go"),
		filepath.Join("run", "server.go"),
		filepath.Join("run", "client.go"),
		filepath.Join("run", "skill.go"),
		filepath.Join("skill", "skill.go"),
		filepath.Join("skill", "SKILL.md"),
		filepath.Join("run", "dispatch_test.go"),
		filepath.Join("run", "client_test.go"),
	} {
		if _, err := os.Stat(filepath.Join(dir, name)); err != nil {
			t.Fatalf("missing generated %s: %v", name, err)
		}
	}
	if _, err := os.Stat(filepath.Join(dir, "main.go")); err == nil {
		t.Fatal("root main.go must not exist; product entry is cmd/demo/main.go")
	}

	// Dispatcher covers the agent verb set.
	dispatch := mustReadCreateTest(t, filepath.Join(dir, "run", "run.go"))
	for _, want := range []string{`case "server":`, `case "get":`, `case "put":`, `case "post":`, `case "delete":`, `case "skill":`} {
		if !strings.Contains(dispatch, want) {
			t.Fatalf("run/run.go missing dispatch %q:\n%s", want, dispatch)
		}
	}

	// The client defaults to port 8080.
	client := mustReadCreateTest(t, filepath.Join(dir, "run", "client.go"))
	if !strings.Contains(client, "const defaultPort = 8080") {
		t.Fatalf("run/client.go should default to port 8080:\n%s", client)
	}
	serverGo := mustReadCreateTest(t, filepath.Join(dir, "run", "server.go"))
	if !strings.Contains(serverGo, "FindAvailablePort(defaultPort, 100)") {
		t.Fatalf("run/server.go should auto-select from the default port:\n%s", serverGo)
	}

	// Skill content is project-scoped.
	skillGo := mustReadCreateTest(t, filepath.Join(dir, "skill", "skill.go"))
	if !strings.Contains(skillGo, `SkillDirName = "demo"`) {
		t.Fatalf("skill/skill.go should substitute the project name:\n%s", skillGo)
	}
	skillMD := mustReadCreateTest(t, filepath.Join(dir, "skill", "SKILL.md"))
	if !strings.Contains(skillMD, "name: demo") {
		t.Fatalf("skill/SKILL.md should substitute the project name:\n%s", skillMD)
	}
	if strings.Contains(skillMD, "__PROJECT_NAME__") {
		t.Fatalf("skill/SKILL.md has unresolved placeholders:\n%s", skillMD)
	}

	// Unknown API paths get a JSON 404 (not the SPA fallback).
	serverFile := mustReadCreateTest(t, filepath.Join(dir, "server", "server.go"))
	if !strings.Contains(serverFile, "handleAPINotFound") {
		t.Fatalf("server/server.go should register the JSON 404 fallback:\n%s", serverFile)
	}

	// No unresolved placeholders anywhere in the generated tree.
	err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(path, ".go") && !strings.HasSuffix(path, ".md") {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		for _, ph := range []string{"__PROJECT_NAME__", "__MODULE_NAME__"} {
			if strings.Contains(string(data), ph) {
				t.Errorf("%s has unresolved placeholder %s", path, ph)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestGoReactAgentCLISharesBaseTemplateFiles(t *testing.T) {
	// The agent-cli variant is go-react plus CLI files. Everything else must
	// stay byte-identical to the go-react template so a fix in one template
	// cannot silently drift from the other; intentional divergence means
	// updating this test.
	baseRoot := "go_react/backend"
	variantRoot := "go_react_agent_cli/backend"

	variantOnly := map[string]bool{
		"run/run.go":               true, // dispatcher replaces the base single-command run.go
		"run/server.go":            true,
		"run/client.go":            true,
		"run/skill.go":             true,
		"run/dispatch_test.go":     true,
		"run/client_test.go":       true,
		"run/skill_test.go":        true,
		"run/imageupload.go":       true, // post --file plus the image library renders
		"run/imageupload_test.go":  true,
		"skill/skill.go":           true,
		"skill/skill_test.go":      true,
		"skill/SKILL.md":           true,
		"server/server.go":         true, // adds the JSON 404 fallback and the image API
		"server/images.go":         true, // the unified image library
		"server/images_api.go":     true,
		"server/images_audit.go":   true,
		"server/images_guard.go":   true,
		"server/images_sniff.go":   true,
		"server/images_json.go":    true,
		"server/images_test.go":    true,
		"server/gallery.go":        true, // the demo container and the delete guard registry
		"README.md":                true, // CLI-focused docs
		"AGENTS.md":                true, // adds the CLI≈web alignment rules for the CLI verbs
	}

	baseFS := goReactTemplateFS
	var walk func(prefix string) error
	walk = func(prefix string) error {
		entries, err := baseFS.ReadDir(prefix)
		if err != nil {
			return err
		}
		for _, e := range entries {
			full := prefix + "/" + e.Name()
			rel := strings.TrimPrefix(full, baseRoot+"/")
			if e.IsDir() {
				if err := walk(full); err != nil {
					return err
				}
				continue
			}
			if variantOnly[rel] {
				continue
			}
			baseData, err := baseFS.ReadFile(full)
			if err != nil {
				return err
			}
			variantData, err := goReactAgentCLITemplateFS.ReadFile(variantRoot + "/" + rel)
			if err != nil {
				t.Fatalf("variant template missing shared file %s: %v", rel, err)
			}
			if string(baseData) != string(variantData) {
				t.Errorf("shared file %s drifted between go-react and go-react-agent-cli templates", rel)
			}
		}
		return nil
	}
	if err := walk(baseRoot); err != nil {
		t.Fatal(err)
	}

	// The variant may not carry files beyond the shared set plus the
	// variant-only list above.
	var walkVariant func(prefix string) error
	walkVariant = func(prefix string) error {
		entries, err := goReactAgentCLITemplateFS.ReadDir(prefix)
		if err != nil {
			return err
		}
		for _, e := range entries {
			full := prefix + "/" + e.Name()
			rel := strings.TrimPrefix(full, variantRoot+"/")
			if e.IsDir() {
				if err := walkVariant(full); err != nil {
					return err
				}
				continue
			}
			if variantOnly[rel] {
				continue
			}
			if _, err := goReactTemplateFS.ReadFile(baseRoot + "/" + rel); err != nil {
				t.Errorf("variant-only file %s is not declared in variantOnly", rel)
			}
		}
		return nil
	}
	if err := walkVariant(variantRoot); err != nil {
		t.Fatal(err)
	}

	// Frontend trees must stay byte-identical too: both variants share the
	// same demo shell (App.tsx with __APP_GEN_*__ placeholders).
	const baseFrontend = "go_react/frontend"
	const variantFrontend = "go_react_agent_cli/frontend"
	var walkFrontend func(prefix string) error
	walkFrontend = func(prefix string) error {
		entries, err := goReactTemplateFS.ReadDir(prefix)
		if err != nil {
			return err
		}
		for _, e := range entries {
			full := prefix + "/" + e.Name()
			rel := strings.TrimPrefix(full, baseFrontend+"/")
			if e.IsDir() {
				if err := walkFrontend(full); err != nil {
					return err
				}
				continue
			}
			baseData, err := goReactTemplateFS.ReadFile(full)
			if err != nil {
				return err
			}
			variantData, err := goReactAgentCLITemplateFS.ReadFile(variantFrontend + "/" + rel)
			if err != nil {
				t.Fatalf("variant template missing shared frontend file %s: %v", rel, err)
			}
			if string(baseData) != string(variantData) {
				t.Errorf("shared frontend file %s drifted between go-react and go-react-agent-cli templates", rel)
			}
		}
		return nil
	}
	if err := walkFrontend(baseFrontend); err != nil {
		t.Fatal(err)
	}
}

// TestGoReactAgentCLITemplateShipsUnifiedImageLibrary pins the image half of the
// scaffold: one content-addressed library in the data dir, ids from the shared
// sequence, an audit that judges bytes, and a delete guard driven by one
// container registry.
func TestGoReactAgentCLITemplateShipsUnifiedImageLibrary(t *testing.T) {
	dir := copyAgentCLIBackend(t, "demo")

	for _, name := range []string{
		filepath.Join("server", "images.go"),
		filepath.Join("server", "images_api.go"),
		filepath.Join("server", "images_audit.go"),
		filepath.Join("server", "images_guard.go"),
		filepath.Join("server", "images_sniff.go"),
		filepath.Join("server", "images_json.go"),
		filepath.Join("server", "gallery.go"),
		filepath.Join("server", "images_test.go"),
		filepath.Join("run", "imageupload.go"),
		filepath.Join("run", "imageupload_test.go"),
	} {
		if _, err := os.Stat(filepath.Join(dir, name)); err != nil {
			t.Fatalf("missing generated %s: %v", name, err)
		}
	}

	// The library: dedup by content, a commit marker written last, bytes that
	// decide the format, and an id from the shared sequence with a floor. The
	// library is split across focused files, so assert on the whole set.
	libraryFiles := []string{
		"images.go", "images_api.go", "images_audit.go",
		"images_guard.go", "images_sniff.go", "images_json.go", "gallery.go",
	}
	var library strings.Builder
	for _, name := range libraryFiles {
		library.WriteString(mustReadCreateTest(t, filepath.Join(dir, "server", name)))
	}
	libraryGo := library.String()
	for _, want := range []string{
		"func md5Hex(", "findByMD5Locked", "// meta.json LAST",
		"func sniffImageMagic(", "func sniffImageType(", "not an image: detected",
		"imageContainers", "func (s *imageStore) Delete(", "cannot rewrite",
		"idalloc.New(", "func (s *imageStore) maxImageID(", "validationError",
		"detachGalleryFile", "gallery.json", "retired",
		`"/api/ids"`, `"/api/images"`, `"/api/images/"`, `"/api/gallery"`, `"/api/data/"`,
		"func RegisterImagesAPI(", "maxAllocatedIDs", "func writeStoreError(", "immutable",
	} {
		if !strings.Contains(libraryGo, want) {
			t.Fatalf("the generated image library is missing %q", want)
		}
	}

	// Every file of the library stays inside the project's file-size rule.
	for _, name := range libraryFiles {
		content := mustReadCreateTest(t, filepath.Join(dir, "server", name))
		if lines := strings.Count(content, "\n"); lines > 500 {
			t.Fatalf("server/%s is %d lines, over the 500-line cap", name, lines)
		}
	}

	serverGo := mustReadCreateTest(t, filepath.Join(dir, "server", "server.go"))
	if !strings.Contains(serverGo, "RegisterImagesAPI(mux)") {
		t.Fatalf("server/server.go does not register the image API:\n%s", serverGo)
	}

	// The CLI: an upload verb, a forced delete, and a path-first read.
	upload := mustReadCreateTest(t, filepath.Join(dir, "run", "imageupload.go"))
	for _, want := range []string{"func uploadImage(", "multipart.NewWriter(", "func renderImage(", "func imageAddress(", "tabwriter"} {
		if !strings.Contains(upload, want) {
			t.Fatalf("run/imageupload.go missing %q:\n%s", want, upload)
		}
	}
	client := mustReadCreateTest(t, filepath.Join(dir, "run", "client.go"))
	for _, want := range []string{`"--file"`, `"--force"`, `"--no-verify"`, "func doRequestData(", "func errorDetail("} {
		if !strings.Contains(client, want) {
			t.Fatalf("run/client.go missing %q:\n%s", want, client)
		}
	}
	dispatch := mustReadCreateTest(t, filepath.Join(dir, "run", "run.go"))
	for _, want := range []string{"/api/images", "/api/ids", "/api/gallery"} {
		if !strings.Contains(dispatch, want) {
			t.Fatalf("run/run.go root help missing path %q:\n%s", want, dispatch)
		}
	}

	// The docs an agent reads.
	skillMD := mustReadCreateTest(t, filepath.Join(dir, "skill", "SKILL.md"))
	for _, want := range []string{"/api/images", "storage/unified-assets", "NOT AN IMAGE", "images/<id>/"} {
		if !strings.Contains(skillMD, want) {
			t.Fatalf("skill/SKILL.md missing %q:\n%s", want, skillMD)
		}
	}
	agents := mustReadCreateTest(t, filepath.Join(dir, "AGENTS.md"))
	for _, want := range []string{"One container registry", "images/<id>", "go test ./server/"} {
		if !strings.Contains(agents, want) {
			t.Fatalf("AGENTS.md missing %q:\n%s", want, agents)
		}
	}

	// No placeholder survives into the new sources.
	for _, name := range []string{
		filepath.Join("server", "images.go"),
		filepath.Join("server", "images_api.go"),
		filepath.Join("server", "images_audit.go"),
		filepath.Join("server", "images_guard.go"),
		filepath.Join("server", "images_sniff.go"),
		filepath.Join("server", "images_json.go"),
		filepath.Join("server", "gallery.go"),
		filepath.Join("server", "images_test.go"),
		filepath.Join("run", "imageupload.go"),
		filepath.Join("run", "imageupload_test.go"),
	} {
		content := mustReadCreateTest(t, filepath.Join(dir, name))
		if strings.Contains(content, "__PROJECT_NAME__") || strings.Contains(content, "__MODULE_NAME__") {
			t.Fatalf("generated %s has unresolved placeholders:\n%s", name, content)
		}
	}
}

// TestGoReactAgentCLITemplateRendersAndPassesItsOwnTests compiles the rendered
// backend and runs the scaffold's shipped tests. The template sources carry
// //go:build ignore, so nothing else in this repo compiles them: without this
// test a template that does not build after placeholder substitution would only
// be discovered by whoever runs `kool create` and then `go build`.
//
// The rendered module resolves its dependencies from the local module cache, so
// on a machine that has never built a generated project the test skips instead
// of failing on a missing download.
func TestGoReactAgentCLITemplateRendersAndPassesItsOwnTests(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping rendered build in -short mode")
	}
	moduleRoot := koolModuleRoot(t)
	dir := t.TempDir()
	if err := copyTemplateDir(goReactAgentCLITemplateFS, "go_react_agent_cli/backend", dir, "demo", "example.com/demo"); err != nil {
		t.Fatal(err)
	}
	// A filesystem replace for kool itself keeps the render portable: the
	// generated run package imports github.com/xhd2015/kool/pkgs/web.
	goMod := "module example.com/demo\n\ngo 1.25.10\n\n" +
		"require github.com/xhd2015/kool v0.0.0\n\n" +
		"replace github.com/xhd2015/kool => " + moduleRoot + "\n"
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte(goMod), 0644); err != nil {
		t.Fatal(err)
	}

	cacheProxy := "file://" + filepath.Join(goEnv(t, "GOMODCACHE"), "cache", "download")
	env := append(os.Environ(), "GOFLAGS=-mod=mod", "GOSUMDB=off", "GOWORK=off", "GOPROXY="+cacheProxy)

	if out, err := runGo(t, dir, env, "mod", "tidy"); err != nil {
		skipIfOffline(t, out)
		t.Fatalf("go mod tidy: %v\n%s", err, out)
	}
	if out, err := runGo(t, dir, env, "build", "./server/", "./run/"); err != nil {
		skipIfOffline(t, out)
		t.Fatalf("the rendered backend does not build: %v\n%s", err, out)
	}
	out, err := runGo(t, dir, env, "test", "./server/", "./run/")
	if err != nil {
		skipIfOffline(t, out)
		t.Fatalf("the scaffold's own tests fail: %v\n%s", err, out)
	}
	for _, pkg := range []string{"example.com/demo/server", "example.com/demo/run"} {
		if !strings.Contains(out, pkg) {
			t.Fatalf("go test did not report %s:\n%s", pkg, out)
		}
	}
}

// koolModuleRoot walks up from the package directory to the module root.
func koolModuleRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("cannot find the kool module root")
		}
		dir = parent
	}
}

func goEnv(t *testing.T, key string) string {
	t.Helper()
	out, err := exec.Command("go", "env", key).Output()
	if err != nil {
		t.Fatalf("go env %s: %v", key, err)
	}
	return strings.TrimSpace(string(out))
}

func runGo(t *testing.T, dir string, env []string, args ...string) (string, error) {
	t.Helper()
	cmd := exec.Command("go", args...)
	cmd.Dir = dir
	cmd.Env = env
	out, err := cmd.CombinedOutput()
	return string(out), err
}

// skipIfOffline turns "this machine cannot fetch the module" into a skip. A
// compile error or a failing test never matches these markers, so those still
// fail.
func skipIfOffline(t *testing.T, output string) {
	t.Helper()
	for _, marker := range []string{
		"cannot find module providing package",
		"no required module provides package",
		"module lookup disabled",
		"dial tcp",
		"connection refused",
		"i/o timeout",
		"missing go.sum entry",
	} {
		if strings.Contains(output, marker) {
			t.Skipf("module cache cannot resolve the rendered module (%s); skipping", marker)
		}
	}
}

// TestGoReactAgentCLITemplateShipsServerOwnedCardMeta pins the meta half of the
// scaffold: the server owns each card's title/hint/empty, serves them with the
// page content, and the page imports the same parts over the @pagemeta alias.
func TestGoReactAgentCLITemplateShipsServerOwnedCardMeta(t *testing.T) {
	dir := copyAgentCLIBackend(t, "demo")

	// The generated package drops the template build tag and embeds the parts.
	metaGo := mustReadCreateTest(t, filepath.Join(dir, "server", "pagemeta", "pagemeta.go"))
	for _, want := range []string{"//go:embed parts/*.json", "func Sections() []string", "func Catalog()", "func Validate() error"} {
		if !strings.Contains(metaGo, want) {
			t.Fatalf("server/pagemeta/pagemeta.go missing %q:\n%s", want, metaGo)
		}
	}
	if strings.Contains(metaGo, "//go:build ignore") {
		t.Fatalf("generated server/pagemeta/pagemeta.go still carries the template build tag:\n%s", metaGo)
	}

	// Every registered card's words live in a part, and the guard test ships.
	part := mustReadCreateTest(t, filepath.Join(dir, "server", "pagemeta", "parts", "CounterCard.json"))
	for _, want := range []string{`"title"`, `"hint"`, `"empty"`} {
		if !strings.Contains(part, want) {
			t.Fatalf("parts/CounterCard.json missing %s:\n%s", want, part)
		}
	}
	if _, err := os.Stat(filepath.Join(dir, "server", "pagemeta", "pagemeta_test.go")); err != nil {
		t.Fatalf("missing generated server/pagemeta/pagemeta_test.go: %v", err)
	}

	// The page document carries the meta next to the rows.
	api := mustReadCreateTest(t, filepath.Join(dir, "server", "pagemeta_api.go"))
	for _, want := range []string{`"/api/page-meta"`, `"/api/pages/home"`, "Meta  pagemeta.Meta", "Empty bool"} {
		if !strings.Contains(api, want) {
			t.Fatalf("server/pagemeta_api.go missing %q:\n%s", want, api)
		}
	}
	serverGo := mustReadCreateTest(t, filepath.Join(dir, "server", "server.go"))
	if !strings.Contains(serverGo, "RegisterPageMetaAPI(mux)") {
		t.Fatalf("server/server.go does not register the page-meta API:\n%s", serverGo)
	}

	// The rule is binding, and the root help enumerates the paths.
	agents := mustReadCreateTest(t, filepath.Join(dir, "AGENTS.md"))
	for _, want := range []string{"Section Meta Is Server-Owned", "@pagemeta/<Card>.json", "go test ./server/pagemeta/"} {
		if !strings.Contains(agents, want) {
			t.Fatalf("AGENTS.md missing %q:\n%s", want, agents)
		}
	}
	dispatch := mustReadCreateTest(t, filepath.Join(dir, "run", "run.go"))
	for _, want := range []string{"/api/page-meta", "/api/pages/home"} {
		if !strings.Contains(dispatch, want) {
			t.Fatalf("run/run.go root help missing path %q:\n%s", want, dispatch)
		}
	}

	// No placeholder survives into the generated Go sources.
	for _, name := range []string{
		filepath.Join("server", "pagemeta", "pagemeta.go"),
		filepath.Join("server", "pagemeta_api.go"),
	} {
		content := mustReadCreateTest(t, filepath.Join(dir, name))
		if strings.Contains(content, "__PROJECT_NAME__") || strings.Contains(content, "__MODULE_NAME__") {
			t.Fatalf("generated %s has unresolved placeholders:\n%s", name, content)
		}
	}

	// The web side imports the server-owned parts and reads the page document.
	readTemplate := func(name string) string {
		t.Helper()
		data, err := goReactAgentCLITemplateFS.ReadFile("go_react_agent_cli/frontend/" + name)
		if err != nil {
			t.Fatalf("read template frontend/%s: %v", name, err)
		}
		return string(data)
	}
	vite := readTemplate("vite.config.ts")
	if !strings.Contains(vite, "'@pagemeta'") || !strings.Contains(vite, "../server/pagemeta/parts") {
		t.Fatalf("vite.config.ts missing the @pagemeta alias:\n%s", vite)
	}
	tsconfig := readTemplate("tsconfig.app.json")
	for _, want := range []string{`"@pagemeta/*"`, `"resolveJsonModule": true`} {
		if !strings.Contains(tsconfig, want) {
			t.Fatalf("tsconfig.app.json missing %q:\n%s", want, tsconfig)
		}
	}
	card := readTemplate("src/components/CounterCard.tsx")
	if !strings.Contains(card, "@pagemeta/CounterCard.json") {
		t.Fatalf("CounterCard.tsx does not import the server-owned part:\n%s", card)
	}
	app := readTemplate("src/App.tsx")
	if !strings.Contains(app, "<CounterCard />") {
		t.Fatalf("App.tsx does not render the meta-driven card:\n%s", app)
	}
	pageAPI := readTemplate("src/api/page.ts")
	if !strings.Contains(pageAPI, "/api/pages/home") {
		t.Fatalf("src/api/page.ts does not read the page document:\n%s", pageAPI)
	}
}
