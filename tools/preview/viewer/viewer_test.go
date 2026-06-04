package viewer

import "testing"

func TestDetectFileType(t *testing.T) {
	tests := []struct {
		ext      string
		expected string
	}{
		{".uml", "uml"},
		{".puml", "uml"},
		{".mmd", "mermaid"},
		{".dot", "dot"},
		{".md", "markdown"},
		{".diff", "diff"},
		{".patch", "diff"},
		{".txt", "text"},
		{".go", "text"},
		{".json", "text"},
		{"", "text"},
		{".UNKNOWN", "text"},
	}

	for _, tt := range tests {
		t.Run(tt.ext+"->"+tt.expected, func(t *testing.T) {
			got := detectFileType(tt.ext)
			if got != tt.expected {
				t.Errorf("detectFileType(%q) = %q, want %q", tt.ext, got, tt.expected)
			}
		})
	}
}
