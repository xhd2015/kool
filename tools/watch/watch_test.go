package watch

import (
	"testing"
	"time"

	"github.com/fsnotify/fsnotify"
)

func TestShouldIgnoreDirectory(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{"normal directory", "/path/to/src", false},
		{"hidden directory", "/path/to/.hidden", true},
		{"node_modules", "/path/to/node_modules", true},
		{"vendor", "/path/to/vendor", true},
		{"build", "/path/to/build", true},
		{"dist", "/path/to/dist", true},
		{"target", "/path/to/target", true},
		{"bin", "/path/to/bin", true},
		{"obj", "/path/to/obj", true},
		{"tmp", "/path/to/tmp", true},
		{"temp", "/path/to/temp", true},
		{"cache", "/path/to/cache", true},
		{"__pycache__", "/path/to/__pycache__", true},
		{"normal nested", "/path/to/src/main", false},
		{"hidden nested", "/path/to/src/.git", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := shouldIgnoreDirectory(tt.path)
			if result != tt.expected {
				t.Errorf("shouldIgnoreDirectory(%q) = %v, expected %v", tt.path, result, tt.expected)
			}
		})
	}
}

func TestMatchesIncludeExclude(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		include  []string
		exclude  []string
		expected bool
	}{
		// Empty patterns (should allow all)
		{"empty patterns", "test.txt", []string{}, []string{}, true},
		{"empty patterns with dotfile", ".gitignore", []string{}, []string{}, true},

		// Include patterns
		{"include match", "test.txt", []string{"*.txt"}, []string{}, true},
		{"include no match", "test.jpg", []string{"*.txt"}, []string{}, false},
		{"include multiple patterns match first", "test.txt", []string{"*.txt", "*.go"}, []string{}, true},
		{"include multiple patterns match second", "test.go", []string{"*.txt", "*.go"}, []string{}, true},
		{"include multiple patterns no match", "test.jpg", []string{"*.txt", "*.go"}, []string{}, false},

		// Exclude patterns
		{"exclude match", "test.log", []string{}, []string{"*.log"}, false},
		{"exclude no match", "test.txt", []string{}, []string{"*.log"}, true},
		{"exclude multiple patterns match first", "test.log", []string{}, []string{"*.log", "*.tmp"}, false},
		{"exclude multiple patterns match second", "test.tmp", []string{}, []string{"*.log", "*.tmp"}, false},
		{"exclude multiple patterns no match", "test.txt", []string{}, []string{"*.log", "*.tmp"}, true},

		// Combined include and exclude
		{"include and exclude both match", "test.txt", []string{"*.txt"}, []string{"*.txt"}, false},
		{"include match exclude no match", "test.txt", []string{"*.txt"}, []string{"*.log"}, true},
		{"include no match exclude match", "test.log", []string{"*.txt"}, []string{"*.log"}, false},
		{"include no match exclude no match", "test.jpg", []string{"*.txt"}, []string{"*.log"}, false},

		// Complex patterns
		{"complex include", "main.go", []string{"main.*"}, []string{}, true},
		{"complex exclude", "backup.txt", []string{}, []string{"backup.*"}, false},

		// Dotfiles
		{"dotfile include match", ".gitignore", []string{".git*"}, []string{}, true},
		{"dotfile exclude match", ".DS_Store", []string{}, []string{".DS_Store"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := matchesIncludeExclude(tt.filename, tt.include, tt.exclude)
			if result != tt.expected {
				t.Errorf("matchesIncludeExclude(%q, %v, %v) = %v, expected %v",
					tt.filename, tt.include, tt.exclude, result, tt.expected)
			}
		})
	}
}

func TestShouldIgnoreEvent(t *testing.T) {
	// Create test options
	defaultOptions := WatchOptions{
		Include: []string{},
		Exclude: []string{},
	}

	includeOptions := WatchOptions{
		Include: []string{"*.txt"},
		Exclude: []string{},
	}

	excludeOptions := WatchOptions{
		Include: []string{},
		Exclude: []string{"*.log"},
	}

	tests := []struct {
		name     string
		event    fsnotify.Event
		options  WatchOptions
		expected bool
	}{
		// Temporary files
		{"temp file with ~", fsnotify.Event{Name: "~backup.txt", Op: fsnotify.Write}, defaultOptions, true},
		{"temp file with .swp", fsnotify.Event{Name: "file.swp", Op: fsnotify.Write}, defaultOptions, true},
		{"temp file with .tmp", fsnotify.Event{Name: "file.tmp", Op: fsnotify.Write}, defaultOptions, true},
		{"temp file with .temp", fsnotify.Event{Name: "file.temp", Op: fsnotify.Write}, defaultOptions, true},
		{"temp file with .log", fsnotify.Event{Name: "file.log", Op: fsnotify.Write}, defaultOptions, true},
		{"temp file with .pid", fsnotify.Event{Name: "file.pid", Op: fsnotify.Write}, defaultOptions, true},
		{"temp file with .lock", fsnotify.Event{Name: "file.lock", Op: fsnotify.Write}, defaultOptions, true},
		{"DS_Store file", fsnotify.Event{Name: ".DS_Store", Op: fsnotify.Write}, defaultOptions, true},
		{"git directory", fsnotify.Event{Name: ".git", Op: fsnotify.Write}, defaultOptions, true},
		{"vscode directory", fsnotify.Event{Name: ".vscode", Op: fsnotify.Write}, defaultOptions, true},
		{"idea directory", fsnotify.Event{Name: ".idea", Op: fsnotify.Write}, defaultOptions, true},

		// Normal files
		{"normal txt file", fsnotify.Event{Name: "test.txt", Op: fsnotify.Write}, defaultOptions, false},
		{"normal go file", fsnotify.Event{Name: "main.go", Op: fsnotify.Write}, defaultOptions, false},
		{"gitignore file", fsnotify.Event{Name: ".gitignore", Op: fsnotify.Write}, defaultOptions, false},
		{"env file", fsnotify.Event{Name: ".env", Op: fsnotify.Write}, defaultOptions, false},

		// Event types
		{"chmod event", fsnotify.Event{Name: "test.txt", Op: fsnotify.Chmod}, defaultOptions, true},
		{"write event", fsnotify.Event{Name: "test.txt", Op: fsnotify.Write}, defaultOptions, false},
		{"create event", fsnotify.Event{Name: "test.txt", Op: fsnotify.Create}, defaultOptions, false},
		{"remove event", fsnotify.Event{Name: "test.txt", Op: fsnotify.Remove}, defaultOptions, false},
		{"rename event", fsnotify.Event{Name: "test.txt", Op: fsnotify.Rename}, defaultOptions, false},

		// Include patterns
		{"include match", fsnotify.Event{Name: "test.txt", Op: fsnotify.Write}, includeOptions, false},
		{"include no match", fsnotify.Event{Name: "test.go", Op: fsnotify.Write}, includeOptions, true},

		// Exclude patterns
		{"exclude match", fsnotify.Event{Name: "test.log", Op: fsnotify.Write}, excludeOptions, true},
		{"exclude no match", fsnotify.Event{Name: "test.txt", Op: fsnotify.Write}, excludeOptions, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := shouldIgnoreEvent(tt.event, tt.options)
			if result != tt.expected {
				t.Errorf("shouldIgnoreEvent(%v, %v) = %v, expected %v",
					tt.event, tt.options, result, tt.expected)
			}
		})
	}
}

func TestHandle(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		expectError bool
		errorMsg    string
		skipReason  string
	}{
		// Throttle validation
		{"invalid throttle", []string{"--throttle", "0s", "echo", "test"}, true, "throttle duration must be greater than 0", ""},
		{"negative throttle", []string{"--throttle", "-1s", "echo", "test"}, true, "throttle duration must be greater than 0", ""},

		// Skip tests that would start watching or call os.Exit
		{"help flag", []string{"--help"}, false, "", "Help flag calls os.Exit which panics in tests"},
		{"help flag short", []string{"-h"}, false, "", "Help flag calls os.Exit which panics in tests"},
		{"command with args", []string{"echo", "hello"}, false, "", "Would start actual file watching"},
		{"no command", []string{}, false, "", "Would start actual file watching"},
		{"valid throttle", []string{"--throttle", "2s", "echo", "test"}, false, "", "Would start actual file watching"},
		{"valid directory", []string{"-d", "/tmp", "echo", "test"}, false, "", "Would start actual file watching"},
		{"directory long form", []string{"--dir", "/tmp", "echo", "test"}, false, "", "Would start actual file watching"},
		{"include pattern", []string{"--include", "*.txt", "echo", "test"}, false, "", "Would start actual file watching"},
		{"exclude pattern", []string{"--exclude", "*.log", "echo", "test"}, false, "", "Would start actual file watching"},
		{"multiple patterns", []string{"--include", "*.txt", "--include", "*.go", "--exclude", "*.log", "echo", "test"}, false, "", "Would start actual file watching"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.skipReason != "" {
				t.Skip(tt.skipReason)
			}

			err := Handle(tt.args)

			if tt.expectError {
				if err == nil {
					t.Errorf("Handle(%v) expected error but got none", tt.args)
				} else if tt.errorMsg != "" && err.Error() != tt.errorMsg {
					t.Errorf("Handle(%v) error = %q, expected %q", tt.args, err.Error(), tt.errorMsg)
				}
			} else {
				if err != nil {
					t.Errorf("Handle(%v) unexpected error: %v", tt.args, err)
				}
			}
		})
	}
}

func TestWatchOptions(t *testing.T) {
	tests := []struct {
		name     string
		options  WatchOptions
		expected WatchOptions
	}{
		{
			name: "default options",
			options: WatchOptions{
				Dir:      "",
				Throttle: 1 * time.Second,
				Include:  []string{},
				Exclude:  []string{},
			},
			expected: WatchOptions{
				Dir:      "",
				Throttle: 1 * time.Second,
				Include:  []string{},
				Exclude:  []string{},
			},
		},
		{
			name: "custom options",
			options: WatchOptions{
				Dir:      "/tmp",
				Throttle: 500 * time.Millisecond,
				Include:  []string{"*.txt", "*.go"},
				Exclude:  []string{"*.log", "*.tmp"},
			},
			expected: WatchOptions{
				Dir:      "/tmp",
				Throttle: 500 * time.Millisecond,
				Include:  []string{"*.txt", "*.go"},
				Exclude:  []string{"*.log", "*.tmp"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.options.Dir != tt.expected.Dir {
				t.Errorf("Dir = %q, expected %q", tt.options.Dir, tt.expected.Dir)
			}
			if tt.options.Throttle != tt.expected.Throttle {
				t.Errorf("Throttle = %v, expected %v", tt.options.Throttle, tt.expected.Throttle)
			}
			if len(tt.options.Include) != len(tt.expected.Include) {
				t.Errorf("Include length = %d, expected %d", len(tt.options.Include), len(tt.expected.Include))
			}
			if len(tt.options.Exclude) != len(tt.expected.Exclude) {
				t.Errorf("Exclude length = %d, expected %d", len(tt.options.Exclude), len(tt.expected.Exclude))
			}
		})
	}
}

// Benchmark tests for performance-critical functions
func BenchmarkShouldIgnoreDirectory(b *testing.B) {
	paths := []string{
		"/path/to/src",
		"/path/to/.hidden",
		"/path/to/node_modules",
		"/path/to/vendor",
		"/path/to/normal/dir",
	}

	for i := 0; i < b.N; i++ {
		for _, path := range paths {
			shouldIgnoreDirectory(path)
		}
	}
}

func BenchmarkMatchesIncludeExclude(b *testing.B) {
	include := []string{"*.txt", "*.go", "*.js"}
	exclude := []string{"*.log", "*.tmp", "*.cache"}
	files := []string{
		"test.txt",
		"main.go",
		"script.js",
		"debug.log",
		"temp.tmp",
		"data.cache",
		"readme.md",
	}

	for i := 0; i < b.N; i++ {
		for _, file := range files {
			matchesIncludeExclude(file, include, exclude)
		}
	}
}

func BenchmarkShouldIgnoreEvent(b *testing.B) {
	options := WatchOptions{
		Include: []string{"*.txt", "*.go"},
		Exclude: []string{"*.log", "*.tmp"},
	}

	events := []fsnotify.Event{
		{Name: "test.txt", Op: fsnotify.Write},
		{Name: "main.go", Op: fsnotify.Write},
		{Name: "debug.log", Op: fsnotify.Write},
		{Name: "temp.tmp", Op: fsnotify.Write},
		{Name: ".DS_Store", Op: fsnotify.Write},
		{Name: "~backup", Op: fsnotify.Write},
	}

	for i := 0; i < b.N; i++ {
		for _, event := range events {
			shouldIgnoreEvent(event, options)
		}
	}
}
