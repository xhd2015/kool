package modcache

import (
	"fmt"
	"io"
	"strings"
)

const (
	stageKindWidth     = 12
	heartbeatEveryVers = 25
)

type stageProgress struct {
	w     io.Writer
	n     int
	total int
}

func newStageProgress(w io.Writer, total int) *stageProgress {
	if w == nil {
		w = io.Discard
	}
	if total < 1 {
		total = 1
	}
	return &stageProgress{w: w, total: total}
}

func (p *stageProgress) start(kind, msg string) {
	if p == nil {
		return
	}
	p.n++
	p.line(kind, msg)
}

func (p *stageProgress) line(kind, msg string) {
	if p == nil || p.w == nil {
		return
	}
	fmt.Fprintf(p.w, "[%d/%d] %-*s %s\n", p.n, p.total, stageKindWidth, kind, msg)
	flushWriter(p.w)
}

func (p *stageProgress) ok(kind, msg string) {
	if msg == "" {
		p.line(kind, "ok")
		return
	}
	p.line(kind, "ok  "+msg)
}

func (p *stageProgress) detail(format string, args ...interface{}) {
	if p == nil || p.w == nil {
		return
	}
	prefix := fmt.Sprintf("[%d/%d] ", p.n, p.total)
	indent := strings.Repeat(" ", len(prefix))
	fmt.Fprintf(p.w, indent+format+"\n", args...)
	flushWriter(p.w)
}

func shouldHeartbeat(i, n int) bool {
	if i <= 0 || n <= 0 || i > n {
		return false
	}
	if i == 1 || i == n {
		return true
	}
	return i%heartbeatEveryVers == 0
}

func flushWriter(w io.Writer) {
	type flusher interface{ Flush() error }
	if f, ok := w.(flusher); ok {
		_ = f.Flush()
	}
}
