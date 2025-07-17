package watch

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/xhd2015/less-gen/flags"
)

const help = `
kool watch rerun command on file changes

Usage: kool watch [OPTIONS] <command> [args...]

Options:
  --throttle DURATION     throttle duration, default is 1s
  --include PATTERN       include file pattern, default is all files
  --exclude PATTERN       exclude file pattern, default is none
  -d, --dir               watch directory, default is current directory
  -h, --help              show help message

Example:
kool watch cmd "go run main.go"
kool watch                  # just print changed files
`

// WatchOptions contains configuration for the file watcher
type WatchOptions struct {
	Dir      string
	Throttle time.Duration
	Include  []string
	Exclude  []string
}

// WatchCallback is called when files change
type WatchCallback func(changedFiles []string)

func Handle(args []string) error {
	var dir string
	var throttle time.Duration = 1 * time.Second
	var include []string
	var exclude []string
	args, err := flags.String("-d,--dir", &dir).
		Duration("--throttle", &throttle).
		StringSlice("--include", &include).
		StringSlice("--exclude", &exclude).
		Help("-h,--help", help).
		StopOnFirstArg().
		Parse(args)
	if err != nil {
		return err
	}

	if throttle <= 0 {
		return fmt.Errorf("throttle duration must be greater than 0")
	}

	options := WatchOptions{
		Dir:      dir,
		Throttle: throttle,
		Include:  include,
		Exclude:  exclude,
	}

	if len(args) == 0 {
		// No command provided, just print changed files
		callback := func(changedFiles []string) {
			fmt.Printf("[watch] Files changed: %s\n", strings.Join(changedFiles, ", "))
		}
		return watchAndRestart(options, callback)
	}

	command := args[0]
	cmdArgs := args[1:]

	// Create callback that executes the command
	callback := func(changedFiles []string) {
		executeCommand(command, cmdArgs)
	}

	return watchAndRestart(options, callback)
}

func executeCommand(command string, args []string) {
	fmt.Printf("[watch] Starting: %s %s\n", command, strings.Join(args, " "))

	cmd := exec.Command(command, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	err := cmd.Run()
	if err != nil {
		log.Printf("[watch] Command exited with error: %v", err)
	}
}

func watchAndRestart(options WatchOptions, callback WatchCallback) error {
	// Create file watcher
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("failed to create watcher: %v", err)
	}
	defer watcher.Close()

	watchDir := options.Dir
	if options.Dir == "" {
		// Add current directory to watcher
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current directory: %v", err)
		}
		watchDir = cwd
	}

	err = addDirectoryRecursively(watcher, watchDir)
	if err != nil {
		return fmt.Errorf("failed to add directory to watcher: %v", err)
	}

	// Create context for cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle interrupt signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Channel for file change events with throttling
	changeChan := make(chan []string, 1)

	// Execute initial callback
	callback([]string{})

	// File watcher goroutine
	go func() {
		var changedFiles []string
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}

				// Filter out irrelevant events
				if shouldIgnoreEvent(event, options) {
					continue
				}

				fmt.Printf("[watch] File changed: %s\n", event.Name)
				changedFiles = append(changedFiles, event.Name)

				// If new directories are created, add them to watcher
				if event.Op&fsnotify.Create == fsnotify.Create {
					if info, err := os.Stat(event.Name); err == nil && info.IsDir() {
						addDirectoryRecursively(watcher, event.Name)
					}
				}

				// Signal change with throttling
				select {
				case changeChan <- changedFiles:
					changedFiles = nil // Reset the slice
				default:
					// Channel is full, ignore this event
				}

			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Printf("[watch] Watcher error: %v", err)
			}
		}
	}()

	// Throttling mechanism
	var throttleTimer *time.Timer

	// Main loop
	for {
		select {
		case <-sigChan:
			fmt.Println("\n[watch] Received interrupt signal, stopping...")
			return nil

		case files := <-changeChan:
			// Reset or start throttle timer
			if throttleTimer != nil {
				throttleTimer.Stop()
			}

			throttleTimer = time.AfterFunc(options.Throttle, func() {
				fmt.Println("[watch] Executing callback due to file changes...")
				callback(files)
			})

		case <-ctx.Done():
			return nil
		}
	}
}

func addDirectoryRecursively(watcher *fsnotify.Watcher, dir string) error {
	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			// Skip hidden directories and common build/cache directories
			if shouldIgnoreDirectory(path) {
				return filepath.SkipDir
			}

			err = watcher.Add(path)
			if err != nil {
				log.Printf("[watch] Failed to watch directory %s: %v", path, err)
			}
		}

		return nil
	})
}

func shouldIgnoreDirectory(path string) bool {
	name := filepath.Base(path)

	// Skip hidden directories
	if strings.HasPrefix(name, ".") {
		return true
	}

	// Skip common build/cache directories
	ignoreDirs := []string{
		"node_modules",
		"vendor",
		"build",
		"dist",
		"target",
		"bin",
		"obj",
		"tmp",
		"temp",
		"cache",
		"__pycache__",
	}

	for _, ignore := range ignoreDirs {
		if name == ignore {
			return true
		}
	}

	return false
}

func shouldIgnoreEvent(event fsnotify.Event, options WatchOptions) bool {
	name := filepath.Base(event.Name)

	// Ignore temporary files that start with ~
	if strings.HasPrefix(name, "~") {
		return true
	}

	// Ignore common temporary file patterns and specific dotfiles
	ignorePatterns := []string{
		".swp",
		".tmp",
		".temp",
		".log",
		".pid",
		".lock",
		".DS_Store",
		".git",    // Git internal files
		".vscode", // VS Code settings (directory)
		".idea",   // IntelliJ IDEA settings (directory)
	}

	for _, pattern := range ignorePatterns {
		if strings.HasSuffix(name, pattern) {
			return true
		}
	}

	// Only watch for Write, Create, Remove, and Rename events
	if event.Op&fsnotify.Write == 0 &&
		event.Op&fsnotify.Create == 0 &&
		event.Op&fsnotify.Remove == 0 &&
		event.Op&fsnotify.Rename == 0 {
		return true
	}

	// Apply include/exclude patterns
	if !matchesIncludeExclude(event.Name, options.Include, options.Exclude) {
		return true
	}

	return false
}

func matchesIncludeExclude(filename string, include, exclude []string) bool {
	// If include patterns are specified, file must match at least one
	if len(include) > 0 {
		matched := false
		for _, pattern := range include {
			if m, _ := filepath.Match(pattern, filepath.Base(filename)); m {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}

	// If exclude patterns are specified, file must not match any
	for _, pattern := range exclude {
		if m, _ := filepath.Match(pattern, filepath.Base(filename)); m {
			return false
		}
	}

	return true
}
