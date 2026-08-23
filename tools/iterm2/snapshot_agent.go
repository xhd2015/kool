package iterm2

import (
	"strings"

	"github.com/xhd2015/agent-pro/pkgs/itermsnapshot"
	"github.com/xhd2015/agent-pro/pkgs/procresolve"
	"github.com/xhd2015/dot-pkgs/go-pkgs/shell/iterm2/snapshot"
)

// CaptureOpts controls optional phases of snapshot capture.
type CaptureOpts struct {
	// NoEnrich skips procresolve agent attach (CLI --no-enrich).
	NoEnrich bool
	// SpaceAllow, when non-empty, enables space-first filtering in the lib.
	SpaceAllow []int
	// SpaceSkipped, when non-nil, receives skipped window-header count.
	SpaceSkipped *int
}

// formatAgentTree returns Unicode FormatTree lines for an agent tree (no trailing blank).
func formatAgentTree(tree []AgentTreeNode) string {
	if len(tree) == 0 {
		return ""
	}
	nodes := make([]procresolve.ProcNode, len(tree))
	for i, n := range tree {
		nodes[i] = procresolve.ProcNode{
			PID:  n.PID,
			PPID: n.PPID,
			Role: n.Role,
			Cmd:  n.Cmd,
		}
	}
	return strings.TrimRight(procresolve.FormatTree(nodes, procresolve.TreeFormatOptions{}), "\n")
}

// makeResolve builds ResolveFromPID for itermsnapshot, honoring AgentResolveByTTY
// fixtures (keyed by short tty) then SnapshotCollector.ResolveFromPID / live default.
func (c *SnapshotCollector) makeResolve(libSnap *snapshot.Snapshot) func(pid int) (*procresolve.Result, error) {
	pidTTY := map[int]string{}
	if libSnap != nil {
		for wi := range libSnap.Windows {
			for ti := range libSnap.Windows[wi].Tabs {
				for si := range libSnap.Windows[wi].Tabs[ti].Sessions {
					s := &libSnap.Windows[wi].Tabs[ti].Sessions[si]
					short := strings.TrimPrefix(s.TTY, "/dev/")
					if short == "" {
						continue
					}
					if s.PID != nil {
						pidTTY[*s.PID] = short
					}
					if s.ShellPID != nil {
						pidTTY[*s.ShellPID] = short
					}
				}
			}
		}
	}
	agentByTTY := c.agentResolveByTTY
	resolve := c.ResolveFromPID
	return func(pid int) (*procresolve.Result, error) {
		if agentByTTY != nil {
			if short, ok := pidTTY[pid]; ok {
				if fx, ok := agentByTTY[short]; ok {
					return fixtureToProcResult(fx), nil
				}
			}
		}
		if resolve != nil {
			return resolve(pid)
		}
		return procresolve.ResolveFromPID(pid, procresolve.Options{
			ListProcs:  procresolve.ListLiveProcs,
			Lsof:       procresolve.LiveLsof,
			EnrichInfo: true,
		})
	}
}

func fixtureToProcResult(fx AgentResolveFixture) *procresolve.Result {
	if fx.Kind == "" || fx.Kind == "none" || fx.SessionID == "" {
		return &procresolve.Result{Kind: "none"}
	}
	tree := make([]procresolve.ProcNode, len(fx.Tree))
	for i, n := range fx.Tree {
		tree[i] = procresolve.ProcNode{
			PID:  n.PID,
			PPID: n.PPID,
			Role: n.Role,
			Cmd:  n.Cmd,
		}
	}
	return &procresolve.Result{
		Kind:      fx.Kind,
		SessionID: fx.SessionID,
		GrokTitle: fx.Title,
		Tree:      tree,
	}
}

// attachAgents maps itermsnapshot agents onto a kool Snapshot (by session ID).
func attachAgents(kool *Snapshot, agents map[string]*itermsnapshot.SessionAgent) {
	if kool == nil || len(agents) == 0 {
		return
	}
	for wi := range kool.Windows {
		for ti := range kool.Windows[wi].Tabs {
			for si := range kool.Windows[wi].Tabs[ti].Sessions {
				s := &kool.Windows[wi].Tabs[ti].Sessions[si]
				ag, ok := agents[s.ID]
				if !ok || ag == nil {
					continue
				}
				s.Agent = sessionAgentFromIterm(ag)
			}
		}
	}
}

func sessionAgentFromIterm(ag *itermsnapshot.SessionAgent) *SessionAgent {
	if ag == nil || ag.Kind == "" || ag.Kind == "none" || ag.SessionID == "" {
		return nil
	}
	out := &SessionAgent{
		Kind:      ag.Kind,
		SessionID: ag.SessionID,
		Title:     ag.Title,
	}
	if len(ag.Tree) > 0 {
		out.Tree = make([]AgentTreeNode, len(ag.Tree))
		for i, n := range ag.Tree {
			out.Tree[i] = AgentTreeNode{
				PID:  n.PID,
				PPID: n.PPID,
				Role: n.Role,
				Cmd:  n.Cmd,
			}
		}
	}
	return out
}

// enrichWindowAgents attaches agents for sessions in one lib window and returns
// the kool window view (used by progressive stream callbacks).
func (c *SnapshotCollector) enrichWindowAgents(libWin snapshot.SnapshotWindow, noEnrich bool) SnapshotWindow {
	kWin := toKoolWindow(libWin)
	if noEnrich {
		return kWin
	}
	partial := &snapshot.Snapshot{Windows: []snapshot.SnapshotWindow{libWin}}
	res, _, err := itermsnapshot.Capture(itermsnapshot.CaptureOpts{
		Snapshot:       partial,
		NoEnrich:       false,
		ResolveFromPID: c.makeResolve(partial),
	})
	if err != nil || res == nil {
		return kWin
	}
	attachAgentsToWindow(&kWin, res.Agents)
	return kWin
}

func attachAgentsToWindow(win *SnapshotWindow, agents map[string]*itermsnapshot.SessionAgent) {
	if win == nil || len(agents) == 0 {
		return
	}
	for ti := range win.Tabs {
		for si := range win.Tabs[ti].Sessions {
			s := &win.Tabs[ti].Sessions[si]
			if ag, ok := agents[s.ID]; ok {
				s.Agent = sessionAgentFromIterm(ag)
			}
		}
	}
}
