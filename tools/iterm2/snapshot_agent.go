package iterm2

import (
	"strings"

	"github.com/xhd2015/agent-pro/pkgs/procresolve"
)

// CaptureOpts controls optional phases of snapshot capture.
type CaptureOpts struct {
	// NoEnrich skips procresolve agent attach (CLI --no-enrich).
	NoEnrich bool
}

// attachAgent resolves grok/codex session info for a busy pane and sets s.Agent.
// Uses AgentResolveByTTY fixtures when present; otherwise live ResolveFromPID.
func (c *SnapshotCollector) attachAgent(s *SnapshotSession, noEnrich bool) {
	if noEnrich || s == nil {
		return
	}
	// Only enrich busy panes (idle / unknown skip).
	if s.Idle == nil || *s.Idle {
		return
	}

	short := strings.TrimPrefix(s.TTY, "/dev/")
	if short != "" && c.agentResolveByTTY != nil {
		if fx, ok := c.agentResolveByTTY[short]; ok {
			s.Agent = sessionAgentFromFixture(fx)
			return
		}
	}

	pid := 0
	if s.PID != nil {
		pid = *s.PID
	} else if s.ShellPID != nil {
		pid = *s.ShellPID
	}
	if pid <= 0 {
		return
	}

	resolve := c.ResolveFromPID
	if resolve == nil {
		resolve = defaultResolveFromPID
	}
	res, err := resolve(pid)
	if err != nil || res == nil {
		return
	}
	s.Agent = sessionAgentFromResult(res)
}

func sessionAgentFromFixture(fx AgentResolveFixture) *SessionAgent {
	if fx.Kind == "" || fx.Kind == "none" || fx.SessionID == "" {
		return nil
	}
	agent := &SessionAgent{
		Kind:      fx.Kind,
		SessionID: fx.SessionID,
		Title:     fx.Title,
	}
	if len(fx.Tree) > 0 {
		agent.Tree = append([]AgentTreeNode(nil), fx.Tree...)
	}
	return agent
}

func sessionAgentFromResult(res *procresolve.Result) *SessionAgent {
	if res == nil || res.Kind == "" || res.Kind == "none" || res.SessionID == "" {
		return nil
	}
	agent := &SessionAgent{
		Kind:      res.Kind,
		SessionID: res.SessionID,
		Title:     res.GrokTitle,
	}
	if len(res.Tree) > 0 {
		agent.Tree = make([]AgentTreeNode, len(res.Tree))
		for i, n := range res.Tree {
			agent.Tree[i] = AgentTreeNode{
				PID:  n.PID,
				PPID: n.PPID,
				Role: n.Role,
				Cmd:  n.Cmd,
			}
		}
	}
	return agent
}

func defaultResolveFromPID(pid int) (*procresolve.Result, error) {
	return procresolve.ResolveFromPID(pid, procresolve.Options{
		ListProcs:  procresolve.ListLiveProcs,
		Lsof:       procresolve.LiveLsof,
		EnrichInfo: true,
	})
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
