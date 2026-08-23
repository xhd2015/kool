package iterm2

import (
	"time"

	"github.com/xhd2015/dot-pkgs/go-pkgs/shell/iterm2/snapshot"
)

// rawProc / liveTTYProc aliases keep live_scan and older tests compiling against
// the lib process model.
type rawProc = snapshot.ProcRow

type liveTTYProc struct {
	rawProc
	TTY string
}

func enrichFromProcs(procs []rawProc, cwds map[int]string, now time.Time) (
	idle *bool,
	shellPID *int,
	chosen *rawProc,
	cwd *string,
	snapProcs []SnapshotProc,
) {
	idle, shellPID, chosenRow, cwd, libProcs := snapshot.EnrichFromProcs(procs, cwds, now)
	if libProcs != nil {
		snapProcs = toKoolProcs(libProcs)
	}
	return idle, shellPID, chosenRow, cwd, snapProcs
}

func defaultListTTYProcs(ttyShorts []string) ([]liveTTYProc, error) {
	c := snapshot.NewCollector()
	rows, err := c.ListTTYProcs(ttyShorts)
	if err != nil {
		return nil, err
	}
	out := make([]liveTTYProc, len(rows))
	for i, r := range rows {
		out[i] = liveTTYProc{rawProc: r.ProcRow, TTY: r.TTY}
	}
	return out, nil
}
