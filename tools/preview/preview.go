package preview

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/xhd2015/kool/tools/preview/uml"
)

func Handle(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("requires a file to preview")
	}
	file := args[0]
	ext := strings.ToLower(filepath.Ext(file))

	if ext == ".uml" || ext == ".puml" {
		return uml.Serve(file)
	}

	return fmt.Errorf("unsupported file type: %s", ext)
}
