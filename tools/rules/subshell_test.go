package rules

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestEnterSubshellBasics(t *testing.T) {
	// This test only checks if the function creates the temporary init script correctly
	// We can't fully test the interactive shell in an automated test

	// Create a mock exec.Command function to capture what would be executed
	origExecCommand := execCommand
	defer func() { execCommand = origExecCommand }()

	var cmdArgs []string
	var cmdPath string
	var scriptContent string

	execCommand = func(name string, arg ...string) *exec.Cmd {
		cmdPath = name
		cmdArgs = arg

		// Read the content of the temp file that would be used as --rcfile
		if len(arg) >= 2 && arg[0] == "--rcfile" {
			content, err := os.ReadFile(arg[1])
			if err == nil {
				scriptContent = string(content)
			}
		}

		// Return a dummy command that does nothing
		cmd := exec.Command("echo", "mock")
		return cmd
	}

	// Call the function
	err := EnterSubshell()
	if err != nil {
		t.Fatalf("EnterSubshell() failed: %v", err)
	}

	// Check if bash was called with the expected flags
	if cmdPath != "bash" {
		t.Errorf("Expected bash command, got %s", cmdPath)
	}

	// Check if --rcfile flag was provided
	rcFileFound := false
	for i, arg := range cmdArgs {
		if arg == "--rcfile" && i+1 < len(cmdArgs) {
			rcFileFound = true
			break
		}
	}
	if !rcFileFound {
		t.Errorf("Expected --rcfile flag, got args: %v", cmdArgs)
	}

	// Check if the script contains the 'use' function definition
	if scriptContent == "" {
		t.Errorf("Failed to read script content")
	} else {
		// Basic checks on script content
		checkScriptContent(t, scriptContent)
	}
}

// checkScriptContent verifies the generated bash script has the expected content
func checkScriptContent(t *testing.T, content string) {
	// Check for key elements in the script
	expected := []string{
		"PS1='(kool rules)",
		"function use",
		"alias list='kool rule list'",
		"alias dir='kool rule dir'",
		"alias add='kool rule add'",
		"alias rm='kool rule rm'",
		"Welcome to kool rules shell",
	}

	for _, exp := range expected {
		if !strings.Contains(content, exp) {
			t.Errorf("Script missing expected content: %s", exp)
		}
	}
}
