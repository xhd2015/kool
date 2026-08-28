package cloudflare

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/xhd2015/dot-pkgs/go-pkgs/terminal/color"
	"github.com/xhd2015/dot-pkgs/go-pkgs/terminal/cursor"
)

const (
	// DefaultReadyTimeout is the max block wait for public readiness.
	DefaultReadyTimeout = 90 * time.Second
	// DefaultReadyPollInterval is the wait-ready poll interval.
	DefaultReadyPollInterval = 2 * time.Second
	// DefaultHeartbeatInterval is how often non-interactive wait logs re-emit
	// the same status signature.
	DefaultHeartbeatInterval = 10 * time.Second
	// DefaultReadyPath is the public path probed for tunnel readiness.
	DefaultReadyPath = "/"
)

// ErrReadyTimeout is returned by WaitPublicReady when the public probe never becomes ready.
var ErrReadyTimeout = errors.New("public ready timeout")

// HTTPDoer is the injectable HTTP client surface used by readiness probes.
type HTTPDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

// WaitReadyOptions configures WaitPublicReady.
type WaitReadyOptions struct {
	Client       HTTPDoer
	URL          string
	Timeout      time.Duration // 0 → DefaultReadyTimeout
	PollInterval time.Duration // 0 → DefaultReadyPollInterval
	Log          io.Writer     // progress; may be nil
	// Interactive: TTY in-place progress (\r + clear). false = newline + heartbeat.
	Interactive bool
	// HeartbeatInterval: non-interactive re-emit interval for same signature (0 → 10s).
	HeartbeatInterval time.Duration
	Style             color.Style
	// Optional injectables for tests (nil → real time):
	Sleep func(d time.Duration)
	Now   func() time.Time
}

// ProbePublic GET-probes url once.
// ready is true when the response indicates the Cloudflare tunnel connector is up
// (any HTTP response other than CF 530 with body containing "1033").
func ProbePublic(ctx context.Context, client HTTPDoer, url string) (code int, body string, ready bool, err error) {
	if client == nil {
		client = http.DefaultClient
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return 0, "", false, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return 0, "", false, err
	}
	defer resp.Body.Close()

	const maxBody = 64 * 1024
	raw, readErr := io.ReadAll(io.LimitReader(resp.Body, maxBody))
	body = string(raw)
	code = resp.StatusCode
	if readErr != nil {
		return code, body, false, readErr
	}
	return code, body, IsTunnelReady(code, body), nil
}

// IsTunnelReady reports whether an HTTP probe result means the tunnel connector is up.
// Transport failures are not ready (caller passes err separately).
// Cloudflare connector-down is HTTP 530 with body containing "1033".
// Any other HTTP status (including origin 404/502) counts as tunnel-up.
func IsTunnelReady(code int, body string) bool {
	if code <= 0 {
		return false
	}
	if code == 530 && strings.Contains(body, "1033") {
		return false
	}
	return true
}

// PublicProbeURL builds https://<domain><path> for readiness probes.
func PublicProbeURL(domain, readyPath string) string {
	d := normalizeDomain(domain)
	path := strings.TrimSpace(readyPath)
	if path == "" {
		path = DefaultReadyPath
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	return "https://" + d + path
}

// WaitPublicReady polls ProbePublic until the tunnel looks ready or timeout.
// Progress is collapsed: Interactive uses in-place \r; non-Interactive emits
// wait lines only on signature change or heartbeat (~10s).
// On ready / timeout: final line ends with \n.
// err is nil when ready; err is ErrReadyTimeout (errors.Is) on timeout.
func WaitPublicReady(ctx context.Context, opts WaitReadyOptions) (elapsed time.Duration, err error) {
	timeout := opts.Timeout
	if timeout <= 0 {
		timeout = DefaultReadyTimeout
	}
	poll := opts.PollInterval
	if poll <= 0 {
		poll = DefaultReadyPollInterval
	}
	heartbeat := opts.HeartbeatInterval
	if heartbeat <= 0 {
		heartbeat = DefaultHeartbeatInterval
	}
	nowFn := opts.Now
	if nowFn == nil {
		nowFn = time.Now
	}
	sleepFn := opts.Sleep
	if sleepFn == nil {
		sleepFn = time.Sleep
	}
	logw := opts.Log
	if logw == nil {
		logw = io.Discard
	}
	style := opts.Style

	start := nowFn()
	deadline := start.Add(timeout)
	url := opts.URL

	var (
		lastSig         string
		lastEmitElapsed time.Duration
		inPlaceOpen     bool
	)

	writeInPlace := func(text string) {
		_, _ = io.WriteString(logw, cursor.Rewrite(text))
		inPlaceOpen = true
	}
	clearInPlace := func() {
		if inPlaceOpen {
			_, _ = io.WriteString(logw, cursor.Clear())
			inPlaceOpen = false
		}
	}
	commitInPlace := func() {
		if inPlaceOpen {
			_, _ = io.WriteString(logw, "\n")
			inPlaceOpen = false
		}
	}

	for {
		if err := ctx.Err(); err != nil {
			if opts.Interactive {
				clearInPlace()
			}
			return nowFn().Sub(start), err
		}

		code, body, ready, probeErr := ProbePublic(ctx, opts.Client, url)
		elapsed = nowFn().Sub(start)
		if probeErr == nil && ready {
			if opts.Interactive {
				clearInPlace()
			}
			line := FormatPublicReadyLine(style, url, code, elapsed)
			_, _ = io.WriteString(logw, line)
			return elapsed, nil
		}

		waitCode := code
		snippet := bodySnippet(body)
		if snippet == "" && probeErr != nil {
			snippet = probeErr.Error()
		}
		sig := waitStatusSignature(waitCode, snippet)
		progress := FormatWaitingProgress(waitCode, elapsed, snippet)

		if opts.Interactive {
			if lastSig != "" && sig != lastSig {
				commitInPlace()
			}
			writeInPlace(progress)
			lastSig = sig
		} else {
			shouldEmit := false
			if lastSig == "" || sig != lastSig {
				shouldEmit = true
			} else if elapsed-lastEmitElapsed >= heartbeat {
				shouldEmit = true
			}
			if shouldEmit {
				_, _ = io.WriteString(logw, progress+"\n")
				lastSig = sig
				lastEmitElapsed = elapsed
			}
		}

		if !nowFn().Before(deadline) {
			elapsed = nowFn().Sub(start)
			if opts.Interactive {
				clearInPlace()
			}
			// Caller writes FormatReadyTimeoutWarning to stderr (cli warning path).
			return elapsed, ErrReadyTimeout
		}

		sleepFn(poll)
	}
}

func bodySnippet(body string) string {
	s := strings.TrimSpace(body)
	if len(s) > 120 {
		s = s[:120]
	}
	return s
}

func waitStatusSignature(code int, snippet string) string {
	return fmt.Sprintf("%d\x00%s", code, strings.TrimSpace(snippet))
}

// FormatWaitingProgress builds a single-line wait progress string without trailing \n.
func FormatWaitingProgress(code int, elapsed time.Duration, snippet string) string {
	snippet = strings.TrimSpace(snippet)
	if snippet != "" {
		return fmt.Sprintf("waiting  %s  HTTP %d  %s", formatShortDuration(elapsed), code, snippet)
	}
	return fmt.Sprintf("waiting  %s  HTTP %d", formatShortDuration(elapsed), code)
}

// FormatPublicReadyLine formats a successful public readiness line (green when color on).
func FormatPublicReadyLine(style color.Style, url string, code int, elapsed time.Duration) string {
	msg := fmt.Sprintf("Public ready: %s (%d) after %s", url, code, formatShortDuration(elapsed))
	return style.Green(msg) + "\n"
}

// FormatReadyTimeoutWarning formats a yellow warning when wait-ready times out.
// Written by callers to stderr typically; returns a full line including trailing \n.
func FormatReadyTimeoutWarning(style color.Style, url string, timeout time.Duration) string {
	msg := fmt.Sprintf("warning: public ready timeout after %s for %s (tunnel keeps running)", formatShortDuration(timeout), url)
	return style.Yellow(msg) + "\n"
}

func formatShortDuration(d time.Duration) string {
	if d < 0 {
		d = 0
	}
	if d >= time.Second && d%time.Second == 0 {
		return fmt.Sprintf("%ds", int(d/time.Second))
	}
	if d >= time.Second {
		return fmt.Sprintf("%ds", int(d.Round(time.Second)/time.Second))
	}
	if d >= time.Millisecond {
		return fmt.Sprintf("%dms", int(d/time.Millisecond))
	}
	return d.String()
}
