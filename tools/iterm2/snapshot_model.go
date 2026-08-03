package iterm2

// Snapshot is a full inventory of iTerm2 windows, tabs, and sessions (panes).
type Snapshot struct {
	CapturedAt string           `json:"captured_at"`
	Host       string           `json:"host"`
	Source     string           `json:"source"`
	Summary    SnapshotSummary  `json:"summary"`
	Windows    []SnapshotWindow `json:"windows"`
}

// SnapshotSummary counts windows/tabs/sessions and idle/busy.
type SnapshotSummary struct {
	Windows  int `json:"windows"`
	Tabs     int `json:"tabs"`
	Sessions int `json:"sessions"`
	Idle     int `json:"idle"`
	Busy     int `json:"busy"`
	Unknown  int `json:"unknown"`
}

// SnapshotWindow is one iTerm2 window.
type SnapshotWindow struct {
	Index int    `json:"index"`
	Name  string `json:"name"`
	// WindowID is the iTerm/AppleScript window id (CG window number when available).
	// Used to resolve macOS Space on sessions save; zero means unknown.
	WindowID uint64        `json:"window_id,omitempty"`
	Tabs     []SnapshotTab `json:"tabs"`
}

// SnapshotTab is one tab; sessions are panes within the tab.
type SnapshotTab struct {
	Index    int               `json:"index"`
	Name     string            `json:"name"`
	Sessions []SnapshotSession `json:"sessions"`
}

// SnapshotSession is one pane/session. Id is the iTerm2 session unique ID (UUID).
type SnapshotSession struct {
	Index             int              `json:"index"`
	ID                string           `json:"id"`
	Name              string           `json:"name"`
	TTY               string           `json:"tty"`
	Profile           string           `json:"profile"`
	ItermIsProcessing bool             `json:"iterm_is_processing"`
	Idle              *bool            `json:"idle"` // nil = unknown
	Cwd               *string          `json:"cwd"`
	ShellPID          *int             `json:"shell_pid"`
	PID               *int             `json:"pid"`
	PPID              *int             `json:"ppid"`
	Stat              *string          `json:"stat"`
	Command           *string          `json:"command"`
	CommandLine       *string          `json:"command_line"`
	StartTime         *string          `json:"start_time"`
	StartTimeUnix     *int64           `json:"start_time_unix"`
	DurationSeconds   *int64           `json:"duration_seconds"`
	Duration          *string          `json:"duration"`
	Etime             *string          `json:"etime"`
	RSSKB             *int64           `json:"rss_kb"`
	Processes         []SnapshotProc   `json:"processes"`
	// Agent is set when procresolve finds a grok/codex session on a busy pane.
	Agent *SessionAgent `json:"agent,omitempty"`
	// Layout hints (not required for id resolution).
	WindowIndex int `json:"window_index,omitempty"`
	TabIndex    int `json:"tab_index,omitempty"`
}

// SessionAgent is the procresolve result attached to a busy SnapshotSession.
type SessionAgent struct {
	Kind      string          `json:"kind"`
	SessionID string          `json:"session_id"`
	Title     string          `json:"title,omitempty"`
	Tree      []AgentTreeNode `json:"tree,omitempty"`
}

// AgentTreeNode is one process in the agent process tree (JSON + FormatTree).
type AgentTreeNode struct {
	PID  int    `json:"pid"`
	PPID int    `json:"ppid"`
	Role string `json:"role,omitempty"` // input | agent-run | … | grok | codex | other
	Cmd  string `json:"cmd"`
}

// AgentResolveFixture is an injectable procresolve result keyed by tty (tests).
type AgentResolveFixture struct {
	Kind      string // grok | codex | none
	SessionID string
	Title     string // optional GrokTitle
	Tree      []AgentTreeNode
}

// SnapshotProc is one process observed on a session tty.
type SnapshotProc struct {
	PID             int     `json:"pid"`
	PPID            int     `json:"ppid"`
	Stat            string  `json:"stat"`
	Etime           string  `json:"etime"`
	DurationSeconds int64   `json:"duration_seconds"`
	Duration        string  `json:"duration"`
	StartTime       *string `json:"start_time"`
	StartTimeUnix   *int64  `json:"start_time_unix"`
	RSSKB           int64   `json:"rss_kb"`
	Command         string  `json:"command"`
}

func boolPtr(v bool) *bool       { return &v }
func intPtr(v int) *int          { return &v }
func int64Ptr(v int64) *int64    { return &v }
func strPtr(v string) *string    { return &v }
