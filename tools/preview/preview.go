package preview

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/xhd2015/kool/tools/preview/uml"
	"github.com/xhd2015/kool/tools/preview/viewer"
)

func Handle(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("requires a file or directory to preview")
	}

	path := args[0]

	// Check if path exists
	stat, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("path does not exist: %s", path)
	}

	// If it's a directory, use the viewer
	if stat.IsDir() {
		return viewer.Serve(path)
	}

	// If it's a file, handle based on extension
	ext := strings.ToLower(filepath.Ext(path))

	if ext == ".uml" || ext == ".puml" {
		return uml.Serve(path)
	}

	return fmt.Errorf("unsupported file type: %s", ext)
}
