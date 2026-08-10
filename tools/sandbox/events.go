package sandbox

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

// SandboxEvent is the Topology B wire format written to session unix sockets.
type SandboxEvent struct {
	V    int    `json:"v"`
	Type string `json:"type"`
	Path string `json:"path"`
	Ts   string `json:"ts"`
}

// maxUnixSockPath is the practical AF_UNIX path length limit (bytes).
// macOS sockaddr_un.sun_path is 104 including trailing NUL → 103 usable.
// Linux is typically 108 including NUL → 107; use the stricter macOS bound.
const maxUnixSockPath = 103

// maxSessionSockBase is an upper bound on session sock basenames
// (e.g. "kool-sandbox-<random>.sock") used when probing path length.
const maxSessionSockBase = 48

// EventsDir returns the directory holding Topology B session sockets for a
// sandbox root parent (KOOL_SANDBOX_ROOT / --root).
//
// Mapping:
//  1. Prefer <root>/events when a worst-case sock path fits AF_UNIX limits.
//  2. Otherwise map to a short stable dir:
//     /tmp/kse/<16-hex-sha256-of-abs-root>/
//     so publisher (NotifyEvent) and runner (startEventListener) agree without
//     a broker. Long WorkingDir / TempDir trees on macOS often exceed sun_path.
func EventsDir(sandboxRoot string) string {
	root := strings.TrimSpace(sandboxRoot)
	if root == "" {
		root = defaultSandboxRootParent()
	}
	root = filepath.Clean(root)
	if abs, err := filepath.Abs(root); err == nil {
		root = abs
	}
	candidate := filepath.Join(root, "events")
	probe := filepath.Join(candidate, strings.Repeat("s", maxSessionSockBase)+".sock")
	if len(probe) <= maxUnixSockPath {
		return candidate
	}
	return shortEventsDir(root)
}

// shortEventsDir maps an absolute sandbox root parent to a short events dir.
// Always prefer "/tmp" on Unix: macOS os.TempDir() is under long
// /var/folders/.../T and would still blow AF_UNIX sun_path.
func shortEventsDir(absRoot string) string {
	sum := sha256.Sum256([]byte(absRoot))
	h := hex.EncodeToString(sum[:8]) // 16 hex chars
	base := "/tmp"
	if runtime.GOOS == "windows" {
		base = os.TempDir()
	}
	return filepath.Join(base, "kse", h)
}

// defaultSandboxRootParent mirrors createMaterializeRoot's parent selection
// without creating a session directory.
func defaultSandboxRootParent() string {
	if parent := strings.TrimSpace(os.Getenv("KOOL_SANDBOX_ROOT")); parent != "" {
		return parent
	}
	if runtime.GOOS == "linux" {
		if st, err := os.Stat("/dev/shm"); err == nil && st.IsDir() {
			return "/dev/shm/kool-sandbox"
		}
	}
	return os.TempDir()
}

// resolveSandboxRootParent returns --root if non-empty, else env/platform default.
func resolveSandboxRootParent(rootFlag string) string {
	if r := strings.TrimSpace(rootFlag); r != "" {
		return r
	}
	return defaultSandboxRootParent()
}

// listEventSocks returns basenames of *.sock under eventsDir (missing dir → empty).
func listEventSocks(eventsDir string) ([]string, error) {
	ents, err := os.ReadDir(eventsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var names []string
	for _, e := range ents {
		if e.IsDir() {
			continue
		}
		if strings.HasSuffix(e.Name(), ".sock") {
			names = append(names, e.Name())
		}
	}
	return names, nil
}

// NotifyEvent publishes a JSON event to every *.sock under EventsDir(root).
// dryRun lists targets without dialing. No socks / all dials fail → warning,
// exit-style nil error (caller prints warning; process exit 0).
func NotifyEvent(root, eventType, path string, dryRun bool) error {
	root = resolveSandboxRootParent(root)
	absPath := filepath.Clean(path)
	if path != "" && !filepath.IsAbs(absPath) {
		// Prefer absolute; if relative, still clean and use as given after Abs attempt.
		if a, err := filepath.Abs(absPath); err == nil {
			absPath = a
		}
	}

	events := EventsDir(root)
	names, err := listEventSocks(events)
	if err != nil {
		return err
	}

	if dryRun {
		if len(names) == 0 {
			fmt.Fprintln(os.Stderr, "Warning: no event sockets under "+events+" (empty or missing)")
			return nil
		}
		fmt.Printf("dry-run: would notify %d socket(s) under %s\n", len(names), events)
		for _, n := range names {
			fmt.Println(n)
		}
		return nil
	}

	if len(names) == 0 {
		fmt.Fprintln(os.Stderr, "Warning: no event sockets (no subscribers) under "+events)
		return nil
	}

	ev := SandboxEvent{
		V:    1,
		Type: eventType,
		Path: absPath,
		Ts:   time.Now().Format(time.RFC3339Nano),
	}
	payload, err := json.Marshal(ev)
	if err != nil {
		return err
	}
	payload = append(payload, '\n')

	ok := 0
	skipped := 0
	for _, n := range names {
		sock := filepath.Join(events, n)
		if err := dialAndSend(sock, payload); err != nil {
			// Stale sock: skip (optional unlink).
			_ = os.Remove(sock)
			skipped++
			continue
		}
		ok++
	}

	if ok == 0 {
		fmt.Fprintf(os.Stderr, "Warning: no sockets dialed successfully (%d stale/skipped) under %s\n", skipped, events)
		return nil
	}
	fmt.Printf("notified %d socket(s) (%d skipped) type=%s path=%s\n", ok, skipped, eventType, absPath)
	return nil
}

func dialAndSend(sockPath string, payload []byte) error {
	conn, err := net.DialTimeout("unix", sockPath, 2*time.Second)
	if err != nil {
		return err
	}
	defer conn.Close()
	_ = conn.SetWriteDeadline(time.Now().Add(2 * time.Second))
	_, err = conn.Write(payload)
	return err
}

// ReloadRuntimeLoadFiles re-unseals loadAbsPath and re-applies only that load's
// Files into the existing sessionRoot (no env / full rematerialize).
func ReloadRuntimeLoadFiles(sessionRoot, loadAbsPath string) error {
	blob, err := openAndUnsealDevbox(loadAbsPath)
	if err != nil {
		return err
	}
	// Write/overwrite only this load's packed file paths into the live session.
	return materializeFiles(sessionRoot, blob)
}

// eventListener binds EventsDir(parent)/<sessionID>.sock and reloads load Files
// on devbox.updated when path is in loadSet.
type eventListener struct {
	ln          net.Listener
	sockPath    string
	sessionRoot string
	loadSet     map[string]struct{}
	reloadMu    sync.Mutex
	stopOnce    sync.Once
	wg          sync.WaitGroup
}

// startEventListener creates EventsDir (0700), binds unix stream sock (0600),
// and starts an accept loop. stop closes the listener and unlinks the sock.
func startEventListener(parent, sessionID, sessionRoot string, loadAbsPaths []string) (*eventListener, error) {
	events := EventsDir(parent)
	if err := os.MkdirAll(events, 0o700); err != nil {
		return nil, fmt.Errorf("events dir: %w", err)
	}
	// Re-assert mode (MkdirAll may leave existing dir modes alone).
	_ = os.Chmod(events, 0o700)

	sockPath := filepath.Join(events, sessionID+".sock")
	// Final guard: if still over limit (very long sessionID), fail clearly.
	if len(sockPath) > maxUnixSockPath {
		return nil, fmt.Errorf("listen events sock: path too long for AF_UNIX (%d > %d): %s",
			len(sockPath), maxUnixSockPath, sockPath)
	}
	_ = os.Remove(sockPath)

	ln, err := net.Listen("unix", sockPath)
	if err != nil {
		return nil, fmt.Errorf("listen events sock: %w", err)
	}
	if err := os.Chmod(sockPath, 0o600); err != nil {
		_ = ln.Close()
		_ = os.Remove(sockPath)
		return nil, fmt.Errorf("chmod events sock: %w", err)
	}

	loadSet := make(map[string]struct{}, len(loadAbsPaths))
	for _, p := range loadAbsPaths {
		if p == "" {
			continue
		}
		loadSet[filepath.Clean(p)] = struct{}{}
	}

	el := &eventListener{
		ln:          ln,
		sockPath:    sockPath,
		sessionRoot: sessionRoot,
		loadSet:     loadSet,
	}
	el.wg.Add(1)
	go el.acceptLoop()
	return el, nil
}

func (el *eventListener) acceptLoop() {
	defer el.wg.Done()
	for {
		conn, err := el.ln.Accept()
		if err != nil {
			return
		}
		el.wg.Add(1)
		go func(c net.Conn) {
			defer el.wg.Done()
			el.handleConn(c)
		}(conn)
	}
}

func (el *eventListener) handleConn(c net.Conn) {
	defer c.Close()
	_ = c.SetReadDeadline(time.Now().Add(5 * time.Second))
	b, err := io.ReadAll(c)
	if err != nil && len(b) == 0 {
		return
	}
	raw := strings.TrimSpace(string(b))
	if raw == "" {
		return
	}
	var ev SandboxEvent
	if err := json.Unmarshal([]byte(raw), &ev); err != nil {
		return
	}
	if ev.Type != "devbox.updated" {
		return
	}
	path := filepath.Clean(ev.Path)
	if !filepath.IsAbs(path) {
		return
	}
	if _, ok := el.loadSet[path]; !ok {
		// Unknown path: discard.
		return
	}
	el.reloadMu.Lock()
	defer el.reloadMu.Unlock()
	_ = ReloadRuntimeLoadFiles(el.sessionRoot, path)
}

// Stop closes the listener, waits for handlers, and unlinks the sock.
func (el *eventListener) Stop() {
	if el == nil {
		return
	}
	el.stopOnce.Do(func() {
		_ = el.ln.Close()
		el.wg.Wait()
		_ = os.Remove(el.sockPath)
	})
}
