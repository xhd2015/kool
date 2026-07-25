package iterm2

import (
	"fmt"
	"path/filepath"
	"strings"
)

// OutputFormat is how a snapshot is rendered.
type OutputFormat string

const (
	FormatCLI      OutputFormat = "cli"
	FormatJSON     OutputFormat = "json"
	FormatMarkdown OutputFormat = "markdown"
	FormatHTML     OutputFormat = "html"
)

// FormatFlags are mutually exclusive explicit format options.
type FormatFlags struct {
	JSON     bool
	Markdown bool
	HTML     bool
}

// ResolveFormat picks the output format from explicit flags and optional -o path.
// Explicit format flags win over file suffix. Multiple format flags → error.
// Unknown suffix with no flag → FormatCLI (plain text when writing to a file).
func ResolveFormat(flags FormatFlags, outputPath string) (OutputFormat, error) {
	n := 0
	var explicit OutputFormat
	if flags.JSON {
		n++
		explicit = FormatJSON
	}
	if flags.Markdown {
		n++
		explicit = FormatMarkdown
	}
	if flags.HTML {
		n++
		explicit = FormatHTML
	}
	if n > 1 {
		return "", fmt.Errorf("Error: --json, --markdown, and --html are mutually exclusive")
	}
	if n == 1 {
		return explicit, nil
	}
	if outputPath == "" {
		return FormatCLI, nil
	}
	return formatFromPath(outputPath), nil
}

func formatFromPath(path string) OutputFormat {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".json":
		return FormatJSON
	case ".md", ".markdown":
		return FormatMarkdown
	case ".html", ".htm":
		return FormatHTML
	default:
		return FormatCLI
	}
}
