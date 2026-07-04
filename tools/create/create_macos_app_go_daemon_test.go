package create

import (
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

func TestCreateMacOSAppGoDaemonIntoEmptyDir(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "test-app")
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}

	if err := HandleCreateMacOSAppGoDaemon([]string{dir}); err != nil {
		t.Fatal(err)
	}

	daemonName := "test-app-daemon"
	for _, file := range []string{
		filepath.Join("test-app-swift", "Package.swift"),
		".gitignore",
		"README.md",
		filepath.Join("test-app-swift", "App.swift"),
		filepath.Join("test-app-swift", "SettingsView.swift"),
		filepath.Join("test-app-swift", "BrowserPreference.swift"),
		filepath.Join("test-app-swift", "BrowserOpener.swift"),
		filepath.Join("test-app-swift", "OpenInBrowserLabelFormatter.swift"),
		filepath.Join("go-pkgs", "go.mod"),
		filepath.Join("go-pkgs", "go.sum"),
		filepath.Join("go-pkgs", "server", "daemon.go"),
		filepath.Join("go-pkgs", "cmd", daemonName, "main.go"),
		filepath.Join("script", "dev.sh"),
		filepath.Join("script", "bundle.sh"),
		filepath.Join("script", "install.sh"),
		filepath.Join("script", "install-debug.sh"),
	} {
		if _, err := os.Stat(filepath.Join(dir, file)); err != nil {
			t.Fatalf("expected %s: %v", file, err)
		}
	}

	if _, err := os.Stat(filepath.Join(dir, ".git")); err != nil {
		t.Fatalf("expected git repository to be initialized: %v", err)
	}

	assertNoPlaceholdersLeft(t, dir)

	gitignore := mustReadCreateTest(t, filepath.Join(dir, ".gitignore"))
	for _, want := range []string{
		"test-app-swift/.build/",
		"test-app-swift/.swiftpm/",
		"test-app.app/",
		"test-app-debug.app/",
		"test-app.dmg",
		"*.dmg",
	} {
		if !strings.Contains(gitignore, want) {
			t.Fatalf(".gitignore missing %q:\n%s", want, gitignore)
		}
	}

	daemonGo := mustReadCreateTest(t, filepath.Join(dir, "go-pkgs", "server", "daemon.go"))
	port := extractPortConstant(t, daemonGo)
	if port < 1 || port > 65535 {
		t.Fatalf("invalid DEFAULT_PORT in daemon.go: %d", port)
	}
	if isPortInUse(port) {
		t.Fatalf("chosen port %d should not be in use in test", port)
	}

	appSwift := mustReadCreateTest(t, filepath.Join(dir, "test-app-swift", "App.swift"))
	for _, want := range []string{
		`Window("Settings", id: "settings")`,
		"settings-menu-button",
		"OpenInBrowserLabelFormatter.format(browser: defaultBrowser)",
		"BrowserOpener.open",
	} {
		if !strings.Contains(appSwift, want) {
			t.Fatalf("App.swift missing %q:\n%s", want, appSwift)
		}
	}

	daemonConfigSwift := mustReadCreateTest(t, filepath.Join(dir, "test-app-swift", "DaemonConfig.swift"))
	if !strings.Contains(daemonConfigSwift, "static let productionPort = "+strconv.Itoa(port)) {
		t.Fatalf("DaemonConfig.swift missing productionPort %d:\n%s", port, daemonConfigSwift)
	}

	cmd := exec.Command("go", "build", "./...")
	cmd.Dir = filepath.Join(dir, "go-pkgs")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("go build ./...: %v\n%s", err, output)
	}
}

func TestCreateMacOSAppGoDaemonRejectsNonEmptyDir(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "test-app")
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "README.md"), []byte("hello"), 0644); err != nil {
		t.Fatal(err)
	}

	err := HandleCreateMacOSAppGoDaemon([]string{dir})
	if err == nil {
		t.Fatal("expected non-empty directory error")
	}
	if !strings.Contains(err.Error(), "not empty") {
		t.Fatalf("expected not empty error, got: %v", err)
	}
}

func TestCreateMacOSAppGoDaemonAllowsDirWithOnlyGit(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "test-app")
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, ".git"), 0755); err != nil {
		t.Fatal(err)
	}

	if err := HandleCreateMacOSAppGoDaemon([]string{dir}); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(filepath.Join(dir, "go-pkgs", "go.mod")); err != nil {
		t.Fatalf("expected go.mod to be created: %v", err)
	}
}

func TestDeriveBundleID(t *testing.T) {
	if got := deriveBundleID("my-cool-app"); got != "com.mycoolapp" {
		t.Fatalf("expected com.mycoolapp, got %s", got)
	}
}

func TestResolveDaemonPortNonTTY(t *testing.T) {
	port, err := resolveDaemonPort()
	if err != nil {
		t.Fatal(err)
	}
	if port < 1 || port > 65535 {
		t.Fatalf("invalid port: %d", port)
	}
	if isPortInUse(port) {
		t.Fatalf("port %d should be available", port)
	}
}

func assertNoPlaceholdersLeft(t *testing.T, dir string) {
	t.Helper()
	placeholders := []string{
		"__PROJECT_NAME__",
		"__MODULE_NAME__",
		"__DAEMON_NAME__",
		"__BUNDLE_ID__",
		"__STATE_SUBPATH__",
		"__DEFAULT_PORT__",
	}
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if strings.HasSuffix(path, ".sum") {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		content := string(data)
		for _, ph := range placeholders {
			if strings.Contains(content, ph) {
				t.Errorf("placeholder %q still present in %s", ph, path)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func extractPortConstant(t *testing.T, daemonGo string) int {
	t.Helper()
	re := regexp.MustCompile(`const defaultPort = (\d+)`)
	m := re.FindStringSubmatch(daemonGo)
	if len(m) != 2 {
		t.Fatalf("could not find defaultPort in daemon.go:\n%s", daemonGo)
	}
	port, err := strconv.Atoi(m[1])
	if err != nil {
		t.Fatal(err)
	}
	return port
}