package create

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGoReactDevTemplateUsesSharedSupervisor(t *testing.T) {
	const project = "demo"
	const module = "example.com/demo"
	dir := t.TempDir()
	if err := copyTemplateDir(goReactTemplateFS, "go_react/backend", dir, project, module); err != nil {
		t.Fatal(err)
	}

	for _, name := range []string{
		filepath.Join("cmd", "demo", "main.go"),
		filepath.Join("script", "dev", "main.go"),
		filepath.Join("internal", "dev", "dev.go"),
		filepath.Join("internal", "dev", "dev_test.go"),
		filepath.Join("server", "route_prefix_test.go"),
	} {
		if _, err := os.Stat(filepath.Join(dir, name)); err != nil {
			t.Fatalf("missing generated %s: %v", name, err)
		}
	}
	if _, err := os.Stat(filepath.Join(dir, "main.go")); err == nil {
		t.Fatal("root main.go must not exist; product entry is cmd/demo/main.go")
	}

	main := mustReadCreateTest(t, filepath.Join(dir, "script", "dev", "main.go"))
	for _, want := range []string{
		`"github.com/xhd2015/dot-pkgs/go-pkgs/dev/server"`,
		`"example.com/demo/internal/dev"`,
		"dev.Run(os.Args[1:])",
	} {
		if !strings.Contains(main, want) {
			t.Fatalf("script/dev/main.go missing %q:\n%s", want, main)
		}
	}
	for _, unwanted := range []string{"hot_reload/air", "exec.Command", "killListenersOnPort"} {
		if strings.Contains(main, unwanted) {
			t.Fatalf("script/dev/main.go still owns %q:\n%s", unwanted, main)
		}
	}

	adapter := mustReadCreateTest(t, filepath.Join(dir, "internal", "dev", "dev.go"))
	for _, want := range []string{
		`BuildPackage: "./script/dev"`,
		"routePrefixEnv",
		`"--base", routePrefix + "/"`,
		"server.DevHandler(backend.FrontendURL, routePrefix)",
	} {
		if !strings.Contains(adapter, want) {
			t.Fatalf("internal/dev/dev.go missing %q:\n%s", want, adapter)
		}
	}
	if strings.Contains(adapter, "__PROJECT_NAME__") || strings.Contains(adapter, "__MODULE_NAME__") {
		t.Fatalf("internal/dev/dev.go has unresolved placeholders:\n%s", adapter)
	}
}
