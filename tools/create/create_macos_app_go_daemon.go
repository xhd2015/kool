package create

import (
	"bufio"
	"embed"
	"fmt"
	"io/fs"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/xhd2015/xgo/support/cmd"
	"golang.org/x/term"
)

//go:embed all:macos_app_go_daemon_template
var macOSAppGoDaemonTemplateFS embed.FS

const macOSAppGoDaemonHelp = `
Usage: kool create macos-app-go-daemon <dir>

Create a new macOS menu bar app with a Go HTTP daemon.
`

func HandleCreateMacOSAppGoDaemon(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("requires dir")
	}
	if args[0] == "--help" || args[0] == "-h" || args[0] == "help" {
		fmt.Println(strings.TrimPrefix(macOSAppGoDaemonHelp, "\n"))
		return nil
	}

	projectDir := filepath.Clean(args[0])
	args = args[1:]
	if len(args) > 0 {
		return fmt.Errorf("unrecognized extra arguments: %v", strings.Join(args, ","))
	}

	projectName := filepath.Base(projectDir)
	if projectName == "" || projectName == "." || projectName == string(filepath.Separator) {
		return fmt.Errorf("invalid project directory: %s", projectDir)
	}

	_, err := prepareProjectDir(projectDir)
	if err != nil {
		return err
	}

	port, err := resolveDaemonPort()
	if err != nil {
		return err
	}

	modulePath, _ := suggestGoModPath(projectDir)
	if modulePath == "" {
		modulePath = filepath.Join(projectName, "go-pkgs")
	} else {
		modulePath = filepath.Join(modulePath, "go-pkgs")
	}

	daemonName := projectName + "-daemon"
	bundleID := deriveBundleID(projectName)
	stateSubpath := "." + projectName + "/daemon"

	err = copyMacOSAppGoDaemonTemplate(
		projectDir,
		projectName,
		modulePath,
		daemonName,
		bundleID,
		stateSubpath,
		port,
	)
	if err != nil {
		return err
	}

	err = os.Rename(
		filepath.Join(projectDir, "go-pkgs", "go.mod.template"),
		filepath.Join(projectDir, "go-pkgs", "go.mod"),
	)
	if err != nil {
		return fmt.Errorf("failed to rename go.mod.template to go.mod: %v", err)
	}

	err = cmd.Debug().Dir(filepath.Join(projectDir, "go-pkgs")).Run("go", "mod", "tidy")
	if err != nil {
		return err
	}

	if err := initGitRepo(projectDir); err != nil {
		return err
	}

	fmt.Printf("Successfully created macos-app-go-daemon project: %s\n", projectDir)
	fmt.Printf("Daemon port: %d\n", port)
	fmt.Printf("To get started:\n  cd %s\n  ./script/dev.sh\n", projectDir)
	return nil
}

func copyMacOSAppGoDaemonTemplate(
	targetDir, projectName, modulePath, daemonName, bundleID, stateSubpath string,
	port int,
) error {
	const srcRoot = "macos_app_go_daemon_template"
	portStr := strconv.Itoa(port)

	err := fs.WalkDir(macOSAppGoDaemonTemplateFS, srcRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(srcRoot, path)
		if err != nil {
			return err
		}
		if relPath == "." {
			return nil
		}

		placeholders := macOSAppGoDaemonPlaceholders(
			projectName, modulePath, daemonName, bundleID, stateSubpath, portStr,
		)
		targetRelPath := applyPlaceholders(relPath, placeholders)
		targetFilePath := filepath.Join(targetDir, targetRelPath)

		if d.IsDir() {
			return os.MkdirAll(targetFilePath, 0755)
		}

		contentBytes, err := macOSAppGoDaemonTemplateFS.ReadFile(path)
		if err != nil {
			return err
		}
		content := stripTemplateBuildIgnore(string(contentBytes))
		content = applyPlaceholders(content, placeholders)

		if err := os.MkdirAll(filepath.Dir(targetFilePath), 0755); err != nil {
			return err
		}
		return os.WriteFile(targetFilePath, []byte(content), 0644)
	})
	if err != nil {
		return err
	}

	for _, script := range []string{"dev.sh", "bundle.sh", "install.sh", "install-debug.sh"} {
		scriptPath := filepath.Join(targetDir, "script", script)
		if err := os.Chmod(scriptPath, 0755); err != nil {
			return fmt.Errorf("chmod %s: %w", scriptPath, err)
		}
	}

	return nil
}

func stripTemplateBuildIgnore(content string) string {
	const tag = "//go:build ignore\n"
	if strings.HasPrefix(content, tag) {
		content = strings.TrimPrefix(content, tag)
		content = strings.TrimPrefix(content, "\n")
	}
	return content
}

func macOSAppGoDaemonPlaceholders(
	projectName, modulePath, daemonName, bundleID, stateSubpath, port string,
) map[string]string {
	return map[string]string{
		"PROJECT_NAME":  projectName,
		"MODULE_NAME":   modulePath,
		"DAEMON_NAME":   daemonName,
		"BUNDLE_ID":     bundleID,
		"STATE_SUBPATH": stateSubpath,
		"DEFAULT_PORT":  port,
	}
}

func deriveBundleID(projectName string) string {
	var b strings.Builder
	b.WriteString("com.")
	for _, r := range strings.ToLower(projectName) {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
		}
	}
	id := b.String()
	if id == "com." {
		return "com.app"
	}
	return id
}

func resolveDaemonPort() (int, error) {
	defaultPort, err := pickRandomAvailablePort()
	if err != nil {
		return 0, fmt.Errorf("pick random port: %w", err)
	}
	if !term.IsTerminal(int(os.Stdin.Fd())) {
		return defaultPort, nil
	}

	reader := bufio.NewReader(os.Stdin)
	current := defaultPort
	for {
		fmt.Printf("Daemon port [%d]: ", current)
		line, err := reader.ReadString('\n')
		if err != nil {
			return 0, fmt.Errorf("read port: %w", err)
		}
		line = strings.TrimSpace(line)
		if line == "" {
			if isPortInUse(current) {
				fmt.Printf("Port %d is already in use, try another.\n", current)
				continue
			}
			return current, nil
		}

		port, err := strconv.Atoi(line)
		if err != nil {
			fmt.Println("Must be a number, try again.")
			continue
		}
		if port < 1 || port > 65535 {
			fmt.Println("Port must be between 1 and 65535, try again.")
			continue
		}
		if isPortInUse(port) {
			fmt.Printf("Port %d is already in use, try another.\n", port)
			current = port
			continue
		}
		return port, nil
	}
}

func pickRandomAvailablePort() (int, error) {
	for i := 0; i < 20; i++ {
		port, err := randomBoundPort()
		if err != nil {
			return 0, err
		}
		if !isPortInUse(port) {
			return port, nil
		}
	}
	return 0, fmt.Errorf("failed to find available port")
}

func randomBoundPort() (int, error) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, err
	}
	defer listener.Close()
	return listener.Addr().(*net.TCPAddr).Port, nil
}

func isPortInUse(port int) bool {
	listener, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		return true
	}
	listener.Close()
	return false
}