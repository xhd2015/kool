package cloudflare

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"regexp"
	"strings"
	"syscall"
	"time"

	dotcf "github.com/xhd2015/dot-pkgs/go-pkgs/cloudflare"
	"github.com/xhd2015/dot-pkgs/go-pkgs/terminal/color"
	"github.com/xhd2015/kool/pkgs/duration"
	"github.com/xhd2015/kool/pkgs/flag"
)

const rootHelp = `kool cloudflare - expose a local HTTP origin on a public Cloudflare hostname

Usage:
  kool cloudflare serve --domain HOST --url URL [--tunnel NAME] [options]
  kool cloudflare -h|--help
  kool cloudflare serve -h|--help

Subcommands:
  serve                            start a named tunnel for a local origin

Serve options:
  --domain HOST                    public hostname (required)
  --url URL                        local origin URL, e.g. http://127.0.0.1:8080 (required)
  --tunnel NAME                    tunnel name (default: kool-lb-<leftmost-domain-label>)
  --ready-timeout DUR              max wait for public readiness (default 90s)
  --no-wait-ready                  skip blocking public readiness wait
  --ready-path PATH                public path to probe (default /)
  --no-teardown                    on Ctrl+C, keep DNS/tunnel/creds (default: full teardown)
  --color                          force ANSI color on (even when stdout is not a TTY)
  --no-color                       force ANSI color off
  -h,--help                        show help message

Examples:
  kool cloudflare serve --domain a.example.com --url http://127.0.0.1:8080
  kool cloudflare serve --domain a.example.com --url http://127.0.0.1:8080 --tunnel my-tun
`

const serveHelp = `kool cloudflare serve - start a Cloudflare tunnel for a local origin

Usage:
  kool cloudflare serve --domain HOST --url URL [--tunnel NAME] [options]

Options:
  --domain HOST                    public hostname (required)
  --url URL                        local origin URL, e.g. http://127.0.0.1:8080 (required)
  --tunnel NAME                    tunnel name (default: kool-lb-<leftmost-domain-label>)
  --ready-timeout DUR              max wait for public readiness (default 90s)
  --no-wait-ready                  skip blocking public readiness wait
  --ready-path PATH                public path to probe (default /)
  --no-teardown                    on Ctrl+C, keep DNS/tunnel/creds (default: full teardown)
  --color                          force ANSI color on (even when stdout is not a TTY)
  --no-color                       force ANSI color off
  -h,--help                        show help message

On Ctrl+C (default): stop connector, remove DNS for the hostname, delete the
Cloudflare tunnel and local credentials/managed dir. Re-run serve to recreate.
Use --no-teardown to only stop the local connector (hostname stays routed).

Examples:
  kool cloudflare serve --domain a.example.com --url http://127.0.0.1:8080
  kool cloudflare serve --domain life-backup.xhd2015.xyz --url http://127.0.0.1:6321
`

// Handle is production entry: HandleWith(args, HandleOpts{}).
func Handle(args []string) error {
	return HandleWith(args, HandleOpts{})
}

// HandleOpts injects dependencies for tests and production defaults.
type HandleOpts struct {
	// StartSession nil → real cloudflare.StartSession adapter.
	StartSession func(SessionStartOptions) (Session, error)
	// WaitSignal nil → block until SIGINT/SIGTERM.
	WaitSignal func() error
	// WaitReady nil → WaitPublicReady against the public probe URL.
	WaitReady func(WaitReadyOptions) (time.Duration, error)
	// HTTPClient nil → &http.Client{Timeout: 10s} for readiness probes.
	HTTPClient HTTPDoer
	// Stdout/Stderr nil → os.Stdout / os.Stderr (help + user messages).
	Stdout io.Writer
	Stderr io.Writer
	// NoColorEnv nil → os.Getenv("NO_COLOR"); set in tests to control auto mode.
	NoColorEnv *string
}

// SessionStartOptions is the subset of session config used by the CLI.
type SessionStartOptions struct {
	Domain     string
	LocalURL   string
	TunnelName string
	// Teardown true → Stop deletes DNS + tunnel + local managed state (kool default).
	Teardown bool
}

// Session is the stoppable surface returned by StartSession.
type Session interface {
	Stop() error
	PublicBaseURL() string
}

// HandleWith is the injectable entry used by doctests.
func HandleWith(args []string, opts HandleOpts) error {
	stdout := opts.Stdout
	if stdout == nil {
		stdout = os.Stdout
	}
	stderr := opts.Stderr
	if stderr == nil {
		stderr = os.Stderr
	}

	if len(args) == 0 {
		return fmt.Errorf("requires subcommand, try 'kool cloudflare --help'")
	}

	// Root help
	if args[0] == "-h" || args[0] == "--help" || args[0] == "help" {
		fmt.Fprint(stdout, rootHelp)
		if !strings.HasSuffix(rootHelp, "\n") {
			fmt.Fprintln(stdout)
		}
		return nil
	}

	cmd := args[0]
	rest := args[1:]

	switch cmd {
	case "serve":
		return handleServe(rest, opts, stdout, stderr)
	default:
		return fmt.Errorf("unrecognized command: %s", cmd)
	}
}

func handleServe(args []string, opts HandleOpts, stdout, stderr io.Writer) error {
	var domain, localURL, tunnel, readyTimeoutStr, readyPath string
	var domainSet, urlSet, tunnelSet, readyTimeoutSet, readyPathSet bool
	var noWaitReady, noTeardown, colorFlag, noColorFlag bool

	n := len(args)
	for i := 0; i < n; i++ {
		f, value := flag.ParseFlag(args, &i)
		if f == "" {
			return fmt.Errorf("unexpected argument: %s", args[i])
		}
		switch f {
		case "-h", "--help":
			fmt.Fprint(stdout, serveHelp)
			if !strings.HasSuffix(serveHelp, "\n") {
				fmt.Fprintln(stdout)
			}
			return nil
		case "--domain":
			v, ok := value()
			if !ok {
				return fmt.Errorf("--domain requires a value")
			}
			domain = v
			domainSet = true
		case "--url":
			v, ok := value()
			if !ok {
				return fmt.Errorf("--url requires a value")
			}
			localURL = v
			urlSet = true
		case "--tunnel":
			v, ok := value()
			if !ok {
				return fmt.Errorf("--tunnel requires a value")
			}
			tunnel = v
			tunnelSet = true
		case "--ready-timeout":
			v, ok := value()
			if !ok {
				return fmt.Errorf("--ready-timeout requires a value")
			}
			readyTimeoutStr = v
			readyTimeoutSet = true
		case "--ready-path":
			v, ok := value()
			if !ok {
				return fmt.Errorf("--ready-path requires a value")
			}
			readyPath = v
			readyPathSet = true
		case "--no-wait-ready":
			noWaitReady = true
		case "--no-teardown":
			noTeardown = true
		case "--color":
			colorFlag = true
		case "--no-color":
			noColorFlag = true
		default:
			return fmt.Errorf("unrecognized: %s", f)
		}
	}

	if !domainSet || strings.TrimSpace(domain) == "" {
		return fmt.Errorf("requires --domain")
	}
	if !urlSet || strings.TrimSpace(localURL) == "" {
		return fmt.Errorf("requires --url")
	}

	colorMode, err := color.ModeFromFlags(colorFlag, noColorFlag)
	if err != nil {
		return err
	}
	noColorEnv := ""
	if opts.NoColorEnv != nil {
		noColorEnv = *opts.NoColorEnv
	} else {
		noColorEnv = os.Getenv("NO_COLOR")
	}
	style := color.Style{Enabled: color.Resolve(colorMode, color.WriterIsTTY(stdout), noColorEnv)}

	readyTimeout := DefaultReadyTimeout
	if readyTimeoutSet {
		d, err := duration.Parse(readyTimeoutStr)
		if err != nil {
			return fmt.Errorf("--ready-timeout: %w", err)
		}
		readyTimeout = d
	}
	if !readyPathSet || strings.TrimSpace(readyPath) == "" {
		readyPath = DefaultReadyPath
	}

	tunnelName := tunnel
	if !tunnelSet || strings.TrimSpace(tunnel) == "" {
		tunnelName = deriveTunnelName(domain)
	}

	teardown := !noTeardown

	start := opts.StartSession
	if start == nil {
		start = func(o SessionStartOptions) (Session, error) {
			var dnsDeleter dotcf.DNSDeleter
			if o.Teardown {
				dnsDeleter = dotcf.NewOriginCertDNSDeleter()
			}
			sess, err := dotcf.StartSession(dotcf.SessionOptions{
				Domain:     o.Domain,
				LocalURL:   o.LocalURL,
				TunnelName: o.TunnelName,
				Log:        stdout,
				DNSDeleter: dnsDeleter,
				Teardown:   o.Teardown,
			})
			if err != nil {
				return nil, err
			}
			return sess, nil
		}
	}

	wait := opts.WaitSignal
	if wait == nil {
		wait = defaultWaitSignal
	}

	sess, err := start(SessionStartOptions{
		Domain:     domain,
		LocalURL:   localURL,
		TunnelName: tunnelName,
		Teardown:   teardown,
	})
	if err != nil {
		return err
	}

	publicURL := ""
	if sess != nil {
		publicURL = sess.PublicBaseURL()
	}
	if publicURL == "" {
		publicURL = "https://" + normalizeDomain(domain)
	}

	fmt.Fprintf(stdout, "Public URL: %s\n", publicURL)
	fmt.Fprintf(stdout, "Tunnel: %s\n", tunnelName)

	if !noWaitReady {
		probeURL := PublicProbeURL(domain, readyPath)
		fmt.Fprintf(stdout, "Checking public readiness (timeout %s)…\n", formatShortDuration(readyTimeout))

		client := opts.HTTPClient
		if client == nil {
			client = &http.Client{Timeout: 10 * time.Second}
		}
		waitOpts := WaitReadyOptions{
			Client:       client,
			URL:          probeURL,
			Timeout:      readyTimeout,
			PollInterval: DefaultReadyPollInterval,
			Log:          stdout,
			Interactive:  color.WriterIsTTY(stdout),
			Style:        style,
		}
		waitReady := opts.WaitReady
		if waitReady == nil {
			waitReady = func(o WaitReadyOptions) (time.Duration, error) {
				return WaitPublicReady(context.Background(), o)
			}
		}
		_, waitErr := waitReady(waitOpts)
		handleWaitReadyResult(waitErr, style, probeURL, readyTimeout, stderr)
	}

	fmt.Fprintln(stdout, "Press Ctrl+C to stop")

	if err := wait(); err != nil {
		if sess != nil {
			_ = sess.Stop()
		}
		return err
	}
	if sess != nil {
		return sess.Stop()
	}
	return nil
}

func handleWaitReadyResult(waitErr error, style color.Style, probeURL string, readyTimeout time.Duration, stderr io.Writer) {
	if waitErr == nil {
		return
	}
	if errors.Is(waitErr, ErrReadyTimeout) {
		fmt.Fprint(stderr, FormatReadyTimeoutWarning(style, probeURL, readyTimeout))
		return
	}
	// Unexpected wait errors: warn and keep serving (same spirit as timeout).
	fmt.Fprintln(stderr, style.Yellow("warning: wait ready: "+waitErr.Error()))
}

func defaultWaitSignal() error {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(sigChan)
	<-sigChan
	return nil
}

const tunnelNamePrefix = "kool-lb-"

// nonSlug collapses runs of characters outside [a-z0-9-] to a single '-'.
var nonSlug = regexp.MustCompile(`[^a-z0-9-]+`)

// deriveTunnelName: kool-lb- + slug of leftmost label of normalized domain.
// Empty slug after slugify → kool-lb- + short hash.
func deriveTunnelName(domain string) string {
	d := normalizeDomain(domain)
	label := d
	if i := strings.IndexByte(d, '.'); i >= 0 {
		label = d[:i]
	}
	slug := strings.Trim(nonSlug.ReplaceAllString(strings.ToLower(label), "-"), "-")
	if slug == "" {
		sum := sha256.Sum256([]byte(d))
		return tunnelNamePrefix + hex.EncodeToString(sum[:])[:8]
	}
	return tunnelNamePrefix + slug
}

func normalizeDomain(domain string) string {
	d := strings.TrimSpace(domain)
	d = strings.TrimPrefix(d, "https://")
	d = strings.TrimPrefix(d, "http://")
	d = strings.TrimSuffix(d, "/")
	if i := strings.IndexByte(d, '/'); i >= 0 {
		d = d[:i]
	}
	return strings.ToLower(d)
}
