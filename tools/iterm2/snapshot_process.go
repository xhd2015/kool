package iterm2

import (
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// rawProc is an unenriched process row from ps.
type rawProc struct {
	PID     int
	PPID    int
	Stat    string
	Etime   string
	RSSKB   int64
	Lstart  string
	Command string
}

var shellBasenames = map[string]bool{
	"bash": true,
	"zsh":  true,
	"sh":   true,
	"fish": true,
	"ksh":  true,
	"tcsh": true,
	"csh":  true,
	"dash": true,
}

func isShellCommand(cmd string) bool {
	base := commandBase(cmd)
	return shellBasenames[base]
}

func isLoginCommand(cmd string) bool {
	base := commandBase(cmd)
	return base == "login"
}

func commandBase(cmd string) string {
	cmd = strings.TrimSpace(cmd)
	if cmd == "" {
		return ""
	}
	first := strings.Fields(cmd)[0]
	base := filepath.Base(first)
	return strings.TrimPrefix(base, "-")
}

// etimeToSeconds parses ps etime: [[dd-]hh:]mm:ss
func etimeToSeconds(et string) int64 {
	et = strings.TrimSpace(et)
	if et == "" {
		return 0
	}
	var days int64
	rest := et
	if i := strings.Index(et, "-"); i >= 0 {
		d, err := strconv.ParseInt(et[:i], 10, 64)
		if err == nil {
			days = d
		}
		rest = et[i+1:]
	}
	parts := strings.Split(rest, ":")
	var h, m, s int64
	switch len(parts) {
	case 3:
		h, _ = strconv.ParseInt(parts[0], 10, 64)
		m, _ = strconv.ParseInt(parts[1], 10, 64)
		s, _ = strconv.ParseInt(parts[2], 10, 64)
	case 2:
		m, _ = strconv.ParseInt(parts[0], 10, 64)
		s, _ = strconv.ParseInt(parts[1], 10, 64)
	case 1:
		s, _ = strconv.ParseInt(parts[0], 10, 64)
	}
	return days*86400 + h*3600 + m*60 + s
}

func humanDuration(sec int64) string {
	if sec < 0 {
		sec = 0
	}
	d := sec / 86400
	h := (sec % 86400) / 3600
	m := (sec % 3600) / 60
	s := sec % 60
	if d > 0 {
		return strconv.FormatInt(d, 10) + "d" + strconv.FormatInt(h, 10) + "h" + strconv.FormatInt(m, 10) + "m"
	}
	if h > 0 {
		return strconv.FormatInt(h, 10) + "h" + strconv.FormatInt(m, 10) + "m" + strconv.FormatInt(s, 10) + "s"
	}
	if m > 0 {
		return strconv.FormatInt(m, 10) + "m" + strconv.FormatInt(s, 10) + "s"
	}
	return strconv.FormatInt(s, 10) + "s"
}

// parseLstart parses macOS ps lstart "Sat Jul 25 13:59:06 2026" as local time.
func parseLstart(ls string, loc *time.Location) (time.Time, bool) {
	ls = strings.TrimSpace(ls)
	if ls == "" {
		return time.Time{}, false
	}
	if loc == nil {
		loc = time.Local
	}
	t, err := time.ParseInLocation("Mon Jan 2 15:04:05 2006", ls, loc)
	if err != nil {
		// try zero-padded day
		t, err = time.ParseInLocation("Mon Jan _2 15:04:05 2006", ls, loc)
		if err != nil {
			return time.Time{}, false
		}
	}
	return t, true
}

// enrichFromProcs classifies idle and picks the primary pid from raw processes.
// idle = interactive shell with no non-shell children (login ignored).
func enrichFromProcs(procs []rawProc, cwds map[int]string, now time.Time) (
	idle *bool,
	shellPID *int,
	chosen *rawProc,
	cwd *string,
	snapProcs []SnapshotProc,
) {
	if len(procs) == 0 {
		return nil, nil, nil, nil, nil
	}

	var shellIdx = -1
	var workIdxs []int
	var fgIdx = -1

	for i, p := range procs {
		cmd := p.Command
		if isLoginCommand(cmd) {
			continue
		}
		if isShellCommand(cmd) {
			if shellIdx < 0 {
				shellIdx = i
			}
			if strings.Contains(p.Stat, "+") {
				fgIdx = i
			}
			continue
		}
		workIdxs = append(workIdxs, i)
		if strings.Contains(p.Stat, "+") {
			fgIdx = i
		}
	}

	var chosenIdx int
	if len(workIdxs) > 0 {
		idle = boolPtr(false)
		if fgIdx >= 0 {
			// use fg if it is work
			isWork := false
			for _, wi := range workIdxs {
				if wi == fgIdx {
					isWork = true
					break
				}
			}
			if isWork {
				chosenIdx = fgIdx
			} else {
				chosenIdx = workIdxs[len(workIdxs)-1]
			}
		} else {
			chosenIdx = workIdxs[len(workIdxs)-1]
		}
	} else {
		idle = boolPtr(true)
		if fgIdx >= 0 {
			chosenIdx = fgIdx
		} else if shellIdx >= 0 {
			chosenIdx = shellIdx
		} else {
			chosenIdx = 0
		}
	}

	c := procs[chosenIdx]
	chosen = &c
	if shellIdx >= 0 {
		shellPID = intPtr(procs[shellIdx].PID)
	}

	if path, ok := cwds[c.PID]; ok && path != "" {
		cwd = strPtr(path)
	} else if shellPID != nil {
		if path, ok := cwds[*shellPID]; ok && path != "" {
			cwd = strPtr(path)
		}
	}

	snapProcs = make([]SnapshotProc, 0, len(procs))
	for _, p := range procs {
		sp := toSnapshotProc(p, now)
		snapProcs = append(snapProcs, sp)
	}
	return idle, shellPID, chosen, cwd, snapProcs
}

func toSnapshotProc(p rawProc, now time.Time) SnapshotProc {
	dur := etimeToSeconds(p.Etime)
	sp := SnapshotProc{
		PID:             p.PID,
		PPID:            p.PPID,
		Stat:            p.Stat,
		Etime:           p.Etime,
		DurationSeconds: dur,
		Duration:        humanDuration(dur),
		RSSKB:           p.RSSKB,
		Command:         redactCommandLine(p.Command),
	}
	if t, ok := parseLstart(p.Lstart, now.Location()); ok {
		iso := t.Format("2006-01-02T15:04:05") + zoneOffset(t)
		u := t.Unix()
		sp.StartTime = &iso
		sp.StartTimeUnix = &u
	} else if !now.IsZero() && dur >= 0 {
		start := now.Add(-time.Duration(dur) * time.Second)
		iso := start.Format("2006-01-02T15:04:05") + zoneOffset(start)
		u := start.Unix()
		sp.StartTime = &iso
		sp.StartTimeUnix = &u
	}
	return sp
}

func zoneOffset(t time.Time) string {
	_, off := t.Zone()
	sign := "+"
	if off < 0 {
		sign = "-"
		off = -off
	}
	h := off / 3600
	m := (off % 3600) / 60
	return sign + twoDigit(h) + twoDigit(m)
}

func twoDigit(n int) string {
	if n < 10 {
		return "0" + strconv.Itoa(n)
	}
	return strconv.Itoa(n)
}

func applyChosenToSession(s *SnapshotSession, idle *bool, shellPID *int, chosen *rawProc, cwd *string, procs []SnapshotProc, now time.Time) {
	s.Idle = idle
	s.ShellPID = shellPID
	s.Cwd = cwd
	s.Processes = procs
	if chosen == nil {
		return
	}
	sp := toSnapshotProc(*chosen, now)
	s.PID = intPtr(chosen.PID)
	s.PPID = intPtr(chosen.PPID)
	s.Stat = strPtr(chosen.Stat)
	// Prefer argv0 basename with login-shell leading '-' preserved (e.g. -bash).
	cmd := shortCommandName(chosen.Command)
	cline := redactCommandLine(chosen.Command)
	s.Command = &cmd
	s.CommandLine = &cline
	s.Etime = strPtr(chosen.Etime)
	s.RSSKB = int64Ptr(chosen.RSSKB)
	s.DurationSeconds = int64Ptr(sp.DurationSeconds)
	s.Duration = strPtr(sp.Duration)
	s.StartTime = sp.StartTime
	s.StartTimeUnix = sp.StartTimeUnix
}

// shortCommandName returns the first argv token basename (keeps leading '-').
func shortCommandName(cmdline string) string {
	cmdline = strings.TrimSpace(cmdline)
	if cmdline == "" {
		return ""
	}
	first := strings.Fields(cmdline)[0]
	return filepath.Base(first)
}
