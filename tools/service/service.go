package service

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/xhd2015/less-gen/flags"
)

// TODO:
// - [ ] abstract service management to a common interface so that brew and systemd details are hidden
// - [ ] fix missing logs after tasks stopped

const help = `
kool service wraps systemd services and provides easier management with little knowledge of the systemd description file.

Usage: kool service <cmd> [OPTIONS]

Available commands:
  list                             list all services and status
  add --name NAME --pwd PWD "cmd args..."  add a background service that runs "cmd args..."
  status <name>                    check status
  logs <name>                      show stdout and stderr logs
  stop <name>                      stop task
  restart <name>                   restart task
  rm <name>                        remove task
  help                             show help message

Options for add:
  --name <name>                    service name (optional, auto-generated from command if not provided)
  --pwd <dir>                      working directory (optional, defaults to current directory)

Examples:
  kool service list                           show all services
  kool service add "python -m http.server"   add a simple HTTP server
  kool service add --name web --pwd /var/www "python -m http.server 8080"
  kool service status web                     check status of 'web' service
  kool service logs web                       show logs for 'web' service
  kool service stop web                       stop 'web' service
  kool service restart web                    restart 'web' service
  kool service rm web                         remove 'web' service
`

type ServiceTask struct {
	Name        string `json:"name"`
	Command     string `json:"command"`
	WorkingDir  string `json:"working_dir"`
	ServiceName string `json:"service_name"` // actual service name used by system
}

func Handle(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("requires command, try 'kool service --help'")
	}
	cmd := args[0]
	args = args[1:]

	if cmd == "help" || cmd == "--help" {
		fmt.Print(strings.TrimPrefix(help, "\n"))
		return nil
	}

	switch cmd {
	case "list":
		return handleList(args)
	case "add":
		return handleAdd(args)
	case "status":
		return handleStatus(args)
	case "logs":
		return handleLogs(args)
	case "stop":
		return handleStop(args)
	case "restart":
		return handleRestart(args)
	case "rm":
		return handleRemove(args)
	default:
		return fmt.Errorf("unrecognized command: %s", cmd)
	}
}

// getUserConfigDir returns the user configuration directory for kool services
func getUserConfigDir() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user config directory: %w", err)
	}

	dir := filepath.Join(configDir, "kool", "services")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("failed to create config directory: %w", err)
	}

	return dir, nil
}

// getUserCacheDir returns the user cache directory for kool service logs
func getUserCacheDir() (string, error) {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user cache directory: %w", err)
	}

	dir := filepath.Join(cacheDir, "kool", "services")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("failed to create cache directory: %w", err)
	}

	return dir, nil
}

// getTasksFile returns the path to tasks.json
func getTasksFile() (string, error) {
	configDir, err := getUserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, "tasks.json"), nil
}

// loadTasks loads all tasks from tasks.json
func loadTasks() ([]ServiceTask, error) {
	tasksFile, err := getTasksFile()
	if err != nil {
		return nil, err
	}

	if _, err := os.Stat(tasksFile); os.IsNotExist(err) {
		return []ServiceTask{}, nil
	}

	data, err := os.ReadFile(tasksFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read tasks file: %w", err)
	}

	var tasks []ServiceTask
	if err := json.Unmarshal(data, &tasks); err != nil {
		return nil, fmt.Errorf("failed to parse tasks file: %w", err)
	}

	return tasks, nil
}

// saveTasks saves tasks to tasks.json
func saveTasks(tasks []ServiceTask) error {
	tasksFile, err := getTasksFile()
	if err != nil {
		return err
	}

	// Ensure we always have a valid JSON array, even if empty
	if tasks == nil {
		tasks = []ServiceTask{}
	}

	data, err := json.MarshalIndent(tasks, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal tasks: %w", err)
	}

	if err := os.WriteFile(tasksFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write tasks file: %w", err)
	}

	return nil
}

// generateServiceName generates a safe service name from command
func generateServiceName(command string) string {
	// Replace spaces with underscores and remove special characters
	reg := regexp.MustCompile(`[^a-zA-Z0-9_-]`)
	name := reg.ReplaceAllString(strings.ReplaceAll(command, " ", "_"), "")

	// Ensure it starts with a letter
	if len(name) > 0 && (name[0] >= '0' && name[0] <= '9') {
		name = "service_" + name
	}

	// Limit length
	if len(name) > 50 {
		name = name[:50]
	}

	return name
}

// findTaskByName finds a task by name
func findTaskByName(tasks []ServiceTask, name string) *ServiceTask {
	for i := range tasks {
		if tasks[i].Name == name {
			return &tasks[i]
		}
	}
	return nil
}

// removeTaskByName removes a task by name
func removeTaskByName(tasks []ServiceTask, name string) []ServiceTask {
	var result []ServiceTask
	for _, task := range tasks {
		if task.Name != name {
			result = append(result, task)
		}
	}
	return result
}

// Platform-specific service management
func isMacOS() bool {
	return runtime.GOOS == "darwin"
}

func isLinux() bool {
	return runtime.GOOS == "linux"
}

// runCommand executes a command and returns output and error
func runCommand(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	output, err := cmd.CombinedOutput()
	return string(output), err
}

// ServiceInfo contains detailed information about a service
type ServiceInfo struct {
	Name      string
	Status    string
	StartTime string
	EndTime   string
	ExitCode  string
	PID       string
}

// getServiceInfo gets detailed information about a service
func getServiceInfo(task ServiceTask) (*ServiceInfo, error) {
	info := &ServiceInfo{
		Name:   task.Name,
		Status: "unknown",
	}

	if isMacOS() {
		return getBrewServiceInfo(task, info)
	} else if isLinux() {
		return getSystemdServiceInfo(task, info)
	}

	return info, fmt.Errorf("unsupported platform: %s", runtime.GOOS)
}

// getBrewServiceInfo gets service info for macOS launchctl
func getBrewServiceInfo(task ServiceTask, info *ServiceInfo) (*ServiceInfo, error) {
	fullServiceName := fmt.Sprintf("kool.%s", task.Name)

	// Get basic status
	output, err := runCommand("launchctl", "list", fullServiceName)
	if err != nil {
		if strings.Contains(output, "Could not find service") {
			info.Status = "stopped"
			return info, nil
		}
		return info, err
	}

	// Parse launchctl list output (property list format)
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Look for LastExitStatus
		if strings.Contains(line, "LastExitStatus") {
			if strings.Contains(line, "=") {
				parts := strings.Split(line, "=")
				if len(parts) == 2 {
					exitCode := strings.TrimSpace(strings.Trim(parts[1], " ;"))
					if exitCode != "0" {
						info.ExitCode = exitCode
						info.Status = "stopped"
					}
				}
			}
		}

		// Look for PID (if running)
		if strings.Contains(line, "PID") && strings.Contains(line, "=") {
			parts := strings.Split(line, "=")
			if len(parts) == 2 {
				pid := strings.TrimSpace(strings.Trim(parts[1], " ;"))
				if pid != "" && pid != "0" {
					info.PID = pid
					info.Status = "running"
				}
			}
		}
	}

	// If we found the service but no clear status, it's loaded but not running
	if info.Status == "unknown" {
		if strings.Contains(output, fullServiceName) {
			info.Status = "loaded"
		} else {
			info.Status = "stopped"
		}
	}

	return info, nil
}

// getSystemdServiceInfo gets service info for Linux systemd
func getSystemdServiceInfo(task ServiceTask, info *ServiceInfo) (*ServiceInfo, error) {
	fullServiceName := fmt.Sprintf("kool-%s.service", task.Name)

	// Get detailed status with show command
	output, err := runCommand("systemctl", "--user", "show", fullServiceName, "--property=ActiveState,SubState,MainPID,ExecMainStartTimestamp,ExecMainExitTimestamp,ExecMainCode,ExecMainStatus")
	if err != nil {
		info.Status = "stopped"
		return info, nil
	}

	// Parse systemctl show output
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(line, "=") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				key, value := parts[0], parts[1]
				switch key {
				case "ActiveState":
					switch value {
					case "active":
						info.Status = "running"
					case "inactive":
						info.Status = "stopped"
					case "failed":
						info.Status = "failed"
					default:
						info.Status = value
					}
				case "MainPID":
					if value != "0" && value != "" {
						info.PID = value
					}
				case "ExecMainStartTimestamp":
					if value != "" && value != "n/a" {
						if t, err := time.Parse("Mon 2006-01-02 15:04:05 MST", value); err == nil {
							info.StartTime = t.Format("2006-01-02 15:04:05")
						} else {
							info.StartTime = value
						}
					}
				case "ExecMainExitTimestamp":
					if value != "" && value != "n/a" {
						if t, err := time.Parse("Mon 2006-01-02 15:04:05 MST", value); err == nil {
							info.EndTime = t.Format("2006-01-02 15:04:05")
						} else {
							info.EndTime = value
						}
					}
				case "ExecMainCode":
					if value != "" && value != "0" {
						info.ExitCode = value
					}
				case "ExecMainStatus":
					if value != "" && value != "0" && info.ExitCode == "" {
						info.ExitCode = value
					}
				}
			}
		}
	}

	return info, nil
}

// macOS brew services integration
func createBrewService(task ServiceTask) error {
	configDir, err := getUserConfigDir()
	if err != nil {
		return err
	}

	cacheDir, err := getUserCacheDir()
	if err != nil {
		return err
	}

	// Create plist file for brew services
	plistDir := filepath.Join(configDir, "plists")
	if err := os.MkdirAll(plistDir, 0755); err != nil {
		return fmt.Errorf("failed to create plist directory: %w", err)
	}

	plistFile := filepath.Join(plistDir, fmt.Sprintf("kool.%s.plist", task.Name))

	plistContent := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>kool.%s</string>
    <key>ProgramArguments</key>
    <array>
        <string>/bin/bash</string>
        <string>-c</string>
        <string>%s</string>
    </array>
    <key>WorkingDirectory</key>
    <string>%s</string>
    <key>RunAtLoad</key>
    <false/>
    <key>KeepAlive</key>
    <true/>
    <key>StandardOutPath</key>
    <string>%s/kool.%s.out.log</string>
    <key>StandardErrorPath</key>
    <string>%s/kool.%s.err.log</string>
</dict>
</plist>`, task.Name, task.Command, task.WorkingDir, cacheDir, task.Name, cacheDir, task.Name)

	if err := os.WriteFile(plistFile, []byte(plistContent), 0644); err != nil {
		return fmt.Errorf("failed to write plist file: %w", err)
	}

	// Link the plist to homebrew services directory
	homebrewServices := filepath.Join(os.Getenv("HOME"), "Library", "LaunchAgents")
	if err := os.MkdirAll(homebrewServices, 0755); err != nil {
		return fmt.Errorf("failed to create LaunchAgents directory: %w", err)
	}

	linkPath := filepath.Join(homebrewServices, fmt.Sprintf("kool.%s.plist", task.Name))

	// Remove existing link if it exists
	os.Remove(linkPath)

	if err := os.Symlink(plistFile, linkPath); err != nil {
		return fmt.Errorf("failed to create symlink: %w", err)
	}

	return nil
}

func removeBrewService(task ServiceTask) error {
	// Remove symlink from LaunchAgents
	linkPath := filepath.Join(os.Getenv("HOME"), "Library", "LaunchAgents", fmt.Sprintf("kool.%s.plist", task.Name))
	os.Remove(linkPath)

	// Remove plist file
	configDir, err := getUserConfigDir()
	if err != nil {
		return err
	}
	plistFile := filepath.Join(configDir, "plists", fmt.Sprintf("kool.%s.plist", task.Name))
	os.Remove(plistFile)

	return nil
}

func getBrewServiceStatus(serviceName string) (string, error) {
	output, err := runCommand("launchctl", "list", fmt.Sprintf("kool.%s", serviceName))
	if err != nil {
		if strings.Contains(output, "Could not find service") {
			return "stopped", nil
		}
		return "unknown", err
	}

	if strings.Contains(output, fmt.Sprintf("kool.%s", serviceName)) {
		return "running", nil
	}
	return "stopped", nil
}

func controlBrewService(serviceName, action string) error {
	fullServiceName := fmt.Sprintf("kool.%s", serviceName)

	switch action {
	case "start":
		_, err := runCommand("launchctl", "load", "-w", filepath.Join(os.Getenv("HOME"), "Library", "LaunchAgents", fmt.Sprintf("%s.plist", fullServiceName)))
		return err
	case "stop":
		_, err := runCommand("launchctl", "unload", "-w", filepath.Join(os.Getenv("HOME"), "Library", "LaunchAgents", fmt.Sprintf("%s.plist", fullServiceName)))
		return err
	case "restart":
		if err := controlBrewService(serviceName, "stop"); err != nil {
			return err
		}
		return controlBrewService(serviceName, "start")
	default:
		return fmt.Errorf("unknown action: %s", action)
	}
}

// Linux systemd integration
func createSystemdService(task ServiceTask) error {
	cacheDir, err := getUserCacheDir()
	if err != nil {
		return err
	}

	// Create systemd user service directory
	systemdDir := filepath.Join(os.Getenv("HOME"), ".config", "systemd", "user")
	if err := os.MkdirAll(systemdDir, 0755); err != nil {
		return fmt.Errorf("failed to create systemd directory: %w", err)
	}

	serviceFile := filepath.Join(systemdDir, fmt.Sprintf("kool-%s.service", task.Name))

	serviceContent := fmt.Sprintf(`[Unit]
Description=Kool Service: %s
After=default.target

[Service]
Type=simple
ExecStart=/bin/bash -c '%s'
WorkingDirectory=%s
Restart=always
RestartSec=5
StandardOutput=append:%s/kool.%s.out.log
StandardError=append:%s/kool.%s.err.log

[Install]
WantedBy=default.target
`, task.Name, task.Command, task.WorkingDir, cacheDir, task.Name, cacheDir, task.Name)

	if err := os.WriteFile(serviceFile, []byte(serviceContent), 0644); err != nil {
		return fmt.Errorf("failed to write service file: %w", err)
	}

	// Reload systemd daemon
	_, err = runCommand("systemctl", "--user", "daemon-reload")
	return err
}

func removeSystemdService(task ServiceTask) error {
	serviceName := fmt.Sprintf("kool-%s.service", task.Name)

	// Stop and disable service first
	runCommand("systemctl", "--user", "stop", serviceName)
	runCommand("systemctl", "--user", "disable", serviceName)

	// Remove service file
	serviceFile := filepath.Join(os.Getenv("HOME"), ".config", "systemd", "user", serviceName)
	os.Remove(serviceFile)

	// Reload daemon
	_, err := runCommand("systemctl", "--user", "daemon-reload")
	return err
}

func getSystemdServiceStatus(serviceName string) (string, error) {
	fullServiceName := fmt.Sprintf("kool-%s.service", serviceName)
	output, err := runCommand("systemctl", "--user", "is-active", fullServiceName)
	if err != nil {
		return "stopped", nil
	}

	status := strings.TrimSpace(output)
	switch status {
	case "active":
		return "running", nil
	case "inactive":
		return "stopped", nil
	case "failed":
		return "failed", nil
	default:
		return status, nil
	}
}

func controlSystemdService(serviceName, action string) error {
	fullServiceName := fmt.Sprintf("kool-%s.service", serviceName)

	switch action {
	case "start":
		_, err := runCommand("systemctl", "--user", "start", fullServiceName)
		return err
	case "stop":
		_, err := runCommand("systemctl", "--user", "stop", fullServiceName)
		return err
	case "restart":
		_, err := runCommand("systemctl", "--user", "restart", fullServiceName)
		return err
	default:
		return fmt.Errorf("unknown action: %s", action)
	}
}

// Command handlers
func handleList(args []string) error {
	tasks, err := loadTasks()
	if err != nil {
		return err
	}

	if len(tasks) == 0 {
		fmt.Println("No services found.")
		return nil
	}

	fmt.Printf("%-20s %-10s %-30s %s\n", "NAME", "STATUS", "COMMAND", "WORKING_DIR")
	fmt.Println(strings.Repeat("-", 80))

	for _, task := range tasks {
		var status string
		var err error

		if isMacOS() {
			status, err = getBrewServiceStatus(task.Name)
		} else if isLinux() {
			status, err = getSystemdServiceStatus(task.Name)
		} else {
			status = "unsupported"
		}

		if err != nil {
			status = "error"
		}

		// Truncate long commands for display
		displayCmd := task.Command
		if len(displayCmd) > 30 {
			displayCmd = displayCmd[:27] + "..."
		}

		fmt.Printf("%-20s %-10s %-30s %s\n", task.Name, status, displayCmd, task.WorkingDir)
	}

	return nil
}

func handleAdd(args []string) error {
	if !isMacOS() && !isLinux() {
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}

	var name string
	var pwd string

	args, err := flags.String("--name", &name).
		String("--pwd", &pwd).
		Parse(args)
	if err != nil {
		return err
	}

	if len(args) == 0 {
		return fmt.Errorf("requires command to run")
	}

	command := strings.Join(args, " ")

	// Generate name if not provided
	if name == "" {
		name = generateServiceName(command)
	}

	// Use current directory if pwd not provided
	if pwd == "" {
		pwd, err = os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current directory: %w", err)
		}
	}

	// Validate working directory
	if _, err := os.Stat(pwd); os.IsNotExist(err) {
		return fmt.Errorf("working directory does not exist: %s", pwd)
	}

	// Load existing tasks
	tasks, err := loadTasks()
	if err != nil {
		return err
	}

	// Check if name already exists
	if findTaskByName(tasks, name) != nil {
		return fmt.Errorf("service with name '%s' already exists", name)
	}

	// Create new task
	task := ServiceTask{
		Name:        name,
		Command:     command,
		WorkingDir:  pwd,
		ServiceName: fmt.Sprintf("kool.%s", name),
	}

	// Create system service
	if isMacOS() {
		if err := createBrewService(task); err != nil {
			return fmt.Errorf("failed to create macOS service: %w", err)
		}
	} else if isLinux() {
		if err := createSystemdService(task); err != nil {
			return fmt.Errorf("failed to create systemd service: %w", err)
		}
	}

	// Add to tasks and save
	tasks = append(tasks, task)
	if err := saveTasks(tasks); err != nil {
		return err
	}

	fmt.Printf("Service '%s' added successfully.\n", name)
	fmt.Printf("Command: %s\n", command)
	fmt.Printf("Working directory: %s\n", pwd)
	fmt.Printf("Use 'kool service status %s' to check status.\n", name)

	return nil
}

func handleStatus(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("requires service name")
	}

	serviceName := args[0]

	tasks, err := loadTasks()
	if err != nil {
		return err
	}

	task := findTaskByName(tasks, serviceName)
	if task == nil {
		return fmt.Errorf("service '%s' not found", serviceName)
	}

	// Get detailed service information
	info, err := getServiceInfo(*task)
	if err != nil {
		return fmt.Errorf("failed to get service info: %w", err)
	}

	fmt.Printf("Service: %s\n", task.Name)
	fmt.Printf("Status: %s\n", info.Status)
	fmt.Printf("Command: %s\n", task.Command)
	fmt.Printf("Working directory: %s\n", task.WorkingDir)

	if info.PID != "" {
		fmt.Printf("PID: %s\n", info.PID)
	}

	if info.StartTime != "" {
		fmt.Printf("Start time: %s\n", info.StartTime)
	}

	if info.EndTime != "" {
		fmt.Printf("End time: %s\n", info.EndTime)
	}

	if info.ExitCode != "" {
		fmt.Printf("Exit code: %s\n", info.ExitCode)
	}

	// Show log file paths
	cacheDir, err := getUserCacheDir()
	if err == nil {
		outLog := filepath.Join(cacheDir, fmt.Sprintf("kool.%s.out.log", task.Name))
		errLog := filepath.Join(cacheDir, fmt.Sprintf("kool.%s.err.log", task.Name))

		if _, err := os.Stat(outLog); err == nil {
			fmt.Printf("Output log: %s\n", outLog)
		}
		if _, err := os.Stat(errLog); err == nil {
			fmt.Printf("Error log: %s\n", errLog)
		}
	}

	return nil
}

func handleLogs(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("requires service name")
	}

	serviceName := args[0]

	tasks, err := loadTasks()
	if err != nil {
		return err
	}

	task := findTaskByName(tasks, serviceName)
	if task == nil {
		return fmt.Errorf("service '%s' not found", serviceName)
	}

	cacheDir, err := getUserCacheDir()
	if err != nil {
		return err
	}

	outLog := filepath.Join(cacheDir, fmt.Sprintf("kool.%s.out.log", task.Name))
	errLog := filepath.Join(cacheDir, fmt.Sprintf("kool.%s.err.log", task.Name))

	// Check if log files exist
	outExists := false
	errExists := false

	if _, err := os.Stat(outLog); err == nil {
		outExists = true
	}
	if _, err := os.Stat(errLog); err == nil {
		errExists = true
	}

	if !outExists && !errExists {
		fmt.Printf("No log files found for service '%s'\n", serviceName)
		return nil
	}

	// Display stdout logs
	if outExists {
		fmt.Printf("=== STDOUT (%s) ===\n", outLog)
		if err := displayLogFile(outLog); err != nil {
			fmt.Printf("Error reading stdout log: %v\n", err)
		}
		fmt.Println()
	}

	// Display stderr logs
	if errExists {
		fmt.Printf("=== STDERR (%s) ===\n", errLog)
		if err := displayLogFile(errLog); err != nil {
			fmt.Printf("Error reading stderr log: %v\n", err)
		}
		fmt.Println()
	}

	return nil
}

// displayLogFile displays the contents of a log file, limiting output
func displayLogFile(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Read and display file contents, limiting to last 100 lines and 4KB
	var lines []string
	scanner := bufio.NewScanner(file)
	totalBytes := 0
	maxBytes := 4096
	maxLines := 100

	for scanner.Scan() {
		line := scanner.Text()
		lineBytes := len(line) + 1 // +1 for newline

		if totalBytes+lineBytes > maxBytes {
			break
		}

		lines = append(lines, line)
		totalBytes += lineBytes

		if len(lines) > maxLines {
			// Keep only the last maxLines
			lines = lines[1:]
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	// Display the lines
	for _, line := range lines {
		fmt.Println(line)
	}

	// Show truncation warning if needed
	file.Seek(0, io.SeekEnd)
	fileSize, _ := file.Seek(0, io.SeekCurrent)
	if fileSize > int64(maxBytes) || len(lines) == maxLines {
		fmt.Printf("\n[Output truncated - showing last %d lines / %d bytes. Full log: %s]\n", len(lines), totalBytes, filePath)
	}

	return nil
}

func handleStop(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("requires service name")
	}

	serviceName := args[0]

	tasks, err := loadTasks()
	if err != nil {
		return err
	}

	task := findTaskByName(tasks, serviceName)
	if task == nil {
		return fmt.Errorf("service '%s' not found", serviceName)
	}

	if isMacOS() {
		err = controlBrewService(task.Name, "stop")
	} else if isLinux() {
		err = controlSystemdService(task.Name, "stop")
	} else {
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}

	if err != nil {
		return fmt.Errorf("failed to stop service: %w", err)
	}

	fmt.Printf("Service '%s' stopped.\n", serviceName)
	return nil
}

func handleRestart(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("requires service name")
	}

	serviceName := args[0]

	tasks, err := loadTasks()
	if err != nil {
		return err
	}

	task := findTaskByName(tasks, serviceName)
	if task == nil {
		return fmt.Errorf("service '%s' not found", serviceName)
	}

	if isMacOS() {
		err = controlBrewService(task.Name, "restart")
	} else if isLinux() {
		err = controlSystemdService(task.Name, "restart")
	} else {
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}

	if err != nil {
		return fmt.Errorf("failed to restart service: %w", err)
	}

	fmt.Printf("Service '%s' restarted.\n", serviceName)
	return nil
}

func handleRemove(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("requires service name")
	}

	serviceName := args[0]

	tasks, err := loadTasks()
	if err != nil {
		return err
	}

	task := findTaskByName(tasks, serviceName)
	if task == nil {
		return fmt.Errorf("service '%s' not found", serviceName)
	}

	// Stop service first
	if isMacOS() {
		controlBrewService(task.Name, "stop") // Ignore errors
		err = removeBrewService(*task)
	} else if isLinux() {
		controlSystemdService(task.Name, "stop") // Ignore errors
		err = removeSystemdService(*task)
	} else {
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}

	if err != nil {
		return fmt.Errorf("failed to remove system service: %w", err)
	}

	// Remove from tasks list
	tasks = removeTaskByName(tasks, serviceName)
	if err := saveTasks(tasks); err != nil {
		return err
	}

	fmt.Printf("Service '%s' removed.\n", serviceName)
	return nil
}
