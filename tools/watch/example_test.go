package watch

import (
	"fmt"
	"testing"
	"time"
)

// ExampleWatchOptions demonstrates how to create WatchOptions
func ExampleWatchOptions() {
	// Create default options
	defaultOptions := WatchOptions{
		Dir:      "",
		Throttle: 1 * time.Second,
		Include:  []string{},
		Exclude:  []string{},
	}

	// Create custom options
	customOptions := WatchOptions{
		Dir:      "/tmp",
		Throttle: 500 * time.Millisecond,
		Include:  []string{"*.txt", "*.go"},
		Exclude:  []string{"*.log", "*.tmp"},
	}

	fmt.Printf("Default options: %+v\n", defaultOptions)
	fmt.Printf("Custom options: %+v\n", customOptions)

	// Output:
	// Default options: {Dir: Throttle:1s Include:[] Exclude:[]}
	// Custom options: {Dir:/tmp Throttle:500ms Include:[*.txt *.go] Exclude:[*.log *.tmp]}
}

// Example_matchesIncludeExclude demonstrates pattern matching
func Example_matchesIncludeExclude() {
	// Test various file patterns
	files := []string{
		"main.go",
		"test.txt",
		"debug.log",
		".gitignore",
		"temp.tmp",
	}

	include := []string{"*.go", "*.txt"}
	exclude := []string{"*.log", "*.tmp"}

	fmt.Println("File matching results:")
	for _, file := range files {
		result := matchesIncludeExclude(file, include, exclude)
		fmt.Printf("  %s: %v\n", file, result)
	}

	// Output:
	// File matching results:
	//   main.go: true
	//   test.txt: true
	//   debug.log: false
	//   .gitignore: false
	//   temp.tmp: false
}

// Example_shouldIgnoreDirectory demonstrates directory filtering
func Example_shouldIgnoreDirectory() {
	directories := []string{
		"/path/to/src",
		"/path/to/.git",
		"/path/to/node_modules",
		"/path/to/build",
		"/path/to/normal",
	}

	fmt.Println("Directory filtering results:")
	for _, dir := range directories {
		result := shouldIgnoreDirectory(dir)
		fmt.Printf("  %s: ignored=%v\n", dir, result)
	}

	// Output:
	// Directory filtering results:
	//   /path/to/src: ignored=false
	//   /path/to/.git: ignored=true
	//   /path/to/node_modules: ignored=true
	//   /path/to/build: ignored=true
	//   /path/to/normal: ignored=false
}

// TestExample_WatchCallback demonstrates how callbacks work
func TestExample_WatchCallback(t *testing.T) {
	// Example callback that just prints changed files
	printCallback := func(changedFiles []string) {
		if len(changedFiles) == 0 {
			fmt.Println("Initial callback - no files changed")
		} else {
			fmt.Printf("Files changed: %v\n", changedFiles)
		}
	}

	// Example callback that counts changes
	var changeCount int
	countCallback := func(changedFiles []string) {
		changeCount++
		fmt.Printf("Change #%d: %d files changed\n", changeCount, len(changedFiles))
	}

	// Test the callbacks
	printCallback([]string{})
	printCallback([]string{"file1.txt", "file2.go"})

	countCallback([]string{"file1.txt"})
	countCallback([]string{"file2.go", "file3.js"})

	// This test just demonstrates the callback concept
	// In actual usage, these would be called by the file watcher
}
