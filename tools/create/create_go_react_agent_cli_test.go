package create

import (
	"os"
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
		"run/run.go":           true, // dispatcher replaces the base single-command run.go
		"run/server.go":        true,
		"run/client.go":        true,
		"run/skill.go":         true,
		"run/dispatch_test.go": true,
		"run/client_test.go":   true,
		"run/skill_test.go":    true,
		"skill/skill.go":       true,
		"skill/skill_test.go":  true,
		"skill/SKILL.md":       true,
		"server/server.go":     true, // adds the JSON 404 fallback
		"README.md":            true, // CLI-focused docs
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
