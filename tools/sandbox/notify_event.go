package sandbox

import (
	"fmt"
	"path/filepath"
	"strings"

	lessflags "github.com/xhd2015/less-flags"
)

const notifyEventHelp = `
kool sandbox notify-event - Publish events to live sandbox session sockets

Usage:
  kool sandbox notify-event --type TYPE --path ABS [--root DIR] [--dry-run]
  kool sandbox notify-event -h|--help

Options:
  --type TYPE     event type (e.g. devbox.updated)
  --path ABS      absolute path associated with the event (load seal for devbox.updated)
  --root DIR      sandbox root parent containing events/ (default: $KOOL_SANDBOX_ROOT or platform default)
  --dry-run       list target sock basenames without dialing
  -h,--help       show help message

Topology B: dials each $ROOT/events/*.sock and writes JSON
  {"v":1,"type":"…","path":"…","ts":"<RFC3339>"}
Stale sockets are skipped. No subscribers → warning, exit 0.
`

func handleNotifyEvent(args []string) error {
	// Manual help so stdout always ends with newline and exit is success (nil).
	for _, a := range args {
		if a == "-h" || a == "--help" || a == "help" {
			txt := strings.TrimPrefix(notifyEventHelp, "\n")
			fmt.Print(txt)
			if !strings.HasSuffix(txt, "\n") {
				fmt.Println()
			}
			return nil
		}
	}

	var eventType string
	var path string
	var root string
	var dryRun bool

	remain, err := lessflags.
		String("--type", &eventType).
		String("--path", &path).
		String("--root", &root).
		Bool("--dry-run", &dryRun).
		Help("-h,--help", notifyEventHelp).
		Parse(args)
	if err != nil {
		return err
	}
	if len(remain) > 0 {
		return fmt.Errorf("unexpected arguments: %s", strings.Join(remain, " "))
	}
	if strings.TrimSpace(eventType) == "" {
		return fmt.Errorf("requires --type")
	}
	if strings.TrimSpace(path) == "" {
		return fmt.Errorf("requires --path")
	}
	// Clean path for absolute contract; relative paths are Abs'd in NotifyEvent.
	path = filepath.Clean(path)

	return NotifyEvent(root, eventType, path, dryRun)
}
