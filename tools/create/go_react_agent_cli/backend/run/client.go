//go:build ignore

package run

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"syscall"

	lessflags "github.com/xhd2015/less-flags"
)

// defaultPort is the port the server listens on when auto-selecting and the
// port bare URIs (e.g. /api/counter) talk to.
const defaultPort = 8080

const clientFlagsHelpSuffix = `Connection options:
  --port <n>       talk to the __PROJECT_NAME__ server on this port (default: 8080)
  --url <url>      talk to this server origin, e.g. http://127.0.0.1:9000
  -h, --help       show help
`

const getHelp = `Usage: __PROJECT_NAME__ get <URI> [--json] [--port <n> | --url <url>]

GET a URI from the running __PROJECT_NAME__ server and print the response body.

<URI> is a path ("/api/counter") or a full URL ("http://localhost:8080/api/counter");
a bare path talks to http://localhost:8080. A query string is kept
("/plans?stage=design"). --json is accepted for symmetry and prints the body
verbatim.

` + clientFlagsHelpSuffix + `
Examples:
  __PROJECT_NAME__ get /ping
  __PROJECT_NAME__ get /api/counter
  __PROJECT_NAME__ get /api/plan/1 --json
  __PROJECT_NAME__ get http://localhost:9000/api/counter
`

const putHelp = `Usage: __PROJECT_NAME__ put <URI> [<json|@file|->] [--json] [--dry-run] [--port <n> | --url <url>]

Replace a resource on the running __PROJECT_NAME__ server. The body is a JSON
literal, @file, or - for stdin. On success prints nothing unless --json is
given.

` + clientFlagsHelpSuffix + `
Examples:
  __PROJECT_NAME__ put /api/plan/1 '{"title":"Cubic roots"}'
  __PROJECT_NAME__ put /api/plan/1 @plan.json --json
  cat plan.json | __PROJECT_NAME__ put /api/plan/1 -
`

const postHelp = `Usage: __PROJECT_NAME__ post <URI> [<json|@file|->] [--json] [--dry-run] [--port <n> | --url <url>]

Create a resource on the running __PROJECT_NAME__ server. The body is optional
(the counter takes none), otherwise a JSON literal, @file, or - for stdin. On
success prints nothing unless --json is given.

` + clientFlagsHelpSuffix + `
Examples:
  __PROJECT_NAME__ post /api/counter
  __PROJECT_NAME__ post /api/plans '{"title":"Cubic roots"}' --json
  __PROJECT_NAME__ post /api/plans @plan.json --dry-run
`

const deleteHelp = `Usage: __PROJECT_NAME__ delete <URI> [--json] [--dry-run] [--port <n> | --url <url>]

Delete a resource on the running __PROJECT_NAME__ server. On success prints
nothing unless --json is given.

` + clientFlagsHelpSuffix + `
Examples:
  __PROJECT_NAME__ delete /api/plan/1
  __PROJECT_NAME__ delete /api/plan/1 --json
`

// clientFlags are the connection options shared by get/put/post/delete.
type clientFlags struct {
	Port int
	URL  string
}

// requestOptions shape one request's output and side effects.
type requestOptions struct {
	json   bool // print the response body on writes; get always prints
	dryRun bool // print the request without sending it
}

// resolveTarget resolves the request URL: a full http(s) URL is used as-is;
// a path starting with "/" talks to the configured origin (--url, else
// http://localhost:<--port or 8080>).
func resolveTarget(uri string, flags clientFlags) (string, error) {
	if strings.HasPrefix(uri, "http://") || strings.HasPrefix(uri, "https://") {
		return uri, nil
	}
	if !strings.HasPrefix(uri, "/") {
		return "", fmt.Errorf("URI must start with / or be a full http(s) URL: %q", uri)
	}
	base := flags.URL
	if base == "" {
		port := flags.Port
		if port == 0 {
			port = defaultPort
		}
		base = fmt.Sprintf("http://localhost:%d", port)
	}
	return strings.TrimSuffix(base, "/") + uri, nil
}

// readPayload resolves a body argument: a JSON literal, @file, or - for stdin.
func readPayload(arg string) ([]byte, error) {
	switch {
	case arg == "-":
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			return nil, fmt.Errorf("read stdin: %w", err)
		}
		return data, nil
	case strings.HasPrefix(arg, "@"):
		path := strings.TrimPrefix(arg, "@")
		if path == "" {
			return nil, fmt.Errorf("empty @file path")
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", path, err)
		}
		return data, nil
	default:
		return []byte(arg), nil
	}
}

func handleGet(args []string) error {
	var asJSON bool
	var flags clientFlags
	remain, err := lessflags.
		Bool("--json", &asJSON).
		Int("--port", &flags.Port).
		String("--url", &flags.URL).
		Help("-h,--help", getHelp).
		HelpNoExit().
		Parse(args)
	if errors.Is(err, lessflags.ErrHelp) {
		return nil
	}
	if err != nil {
		return err
	}
	if len(remain) != 1 {
		return fmt.Errorf("expected exactly one URI, e.g. /api/counter (got %d)\nRun '__PROJECT_NAME__ get --help' for usage", len(remain))
	}
	target, err := resolveTarget(remain[0], flags)
	if err != nil {
		return err
	}
	// get always prints the body verbatim; --json is accepted so the same
	// call shape works across verbs (and for any future human-readable
	// render, where --json would switch back to raw JSON).
	_ = asJSON
	return doRequest(http.MethodGet, target, nil, requestOptions{})
}

func handlePut(args []string) error {
	return handleWrite(http.MethodPut, putHelp, args)
}

func handlePost(args []string) error {
	return handleWrite(http.MethodPost, postHelp, args)
}

func handleDelete(args []string) error {
	var asJSON, dryRun bool
	var flags clientFlags
	remain, err := lessflags.
		Bool("--dry-run", &dryRun).
		Bool("--json", &asJSON).
		Int("--port", &flags.Port).
		String("--url", &flags.URL).
		Help("-h,--help", deleteHelp).
		HelpNoExit().
		Parse(args)
	if errors.Is(err, lessflags.ErrHelp) {
		return nil
	}
	if err != nil {
		return err
	}
	if len(remain) != 1 {
		return fmt.Errorf("expected exactly one URI, e.g. /api/plan/1 (got %d)\nRun '__PROJECT_NAME__ delete --help' for usage", len(remain))
	}
	target, err := resolveTarget(remain[0], flags)
	if err != nil {
		return err
	}
	return doRequest(http.MethodDelete, target, nil, requestOptions{json: asJSON, dryRun: dryRun})
}

// handleWrite runs put/post: one URI plus an optional body.
func handleWrite(method string, helpText string, args []string) error {
	var asJSON, dryRun bool
	var flags clientFlags
	remain, err := lessflags.
		Bool("--dry-run", &dryRun).
		Bool("--json", &asJSON).
		Int("--port", &flags.Port).
		String("--url", &flags.URL).
		Help("-h,--help", helpText).
		HelpNoExit().
		Parse(args)
	if errors.Is(err, lessflags.ErrHelp) {
		return nil
	}
	if err != nil {
		return err
	}
	if len(remain) == 0 {
		return fmt.Errorf("expected a URI, e.g. /api/plan/1\nRun '__PROJECT_NAME__ %s --help' for usage", strings.ToLower(method))
	}
	if len(remain) > 2 {
		return fmt.Errorf("expected one URI and one optional JSON body, got %d arguments\nQuote the body so the shell keeps it whole, or pass @file / -", len(remain))
	}
	target, err := resolveTarget(remain[0], flags)
	if err != nil {
		return err
	}
	var body []byte
	if len(remain) == 2 {
		body, err = readPayload(remain[1])
		if err != nil {
			return err
		}
	}
	return doRequest(method, target, body, requestOptions{json: asJSON, dryRun: dryRun})
}

// doRequest sends one HTTP request and prints the result: the body verbatim
// for get; silent on 2xx for writes unless --json; a non-2xx status is an
// error carrying the status and a body snippet.
func doRequest(method string, target string, body []byte, opts requestOptions) error {
	if opts.dryRun {
		fmt.Printf("%s %s\n", method, target)
		if len(body) > 0 {
			fmt.Println(string(body))
		}
		return nil
	}
	var reader io.Reader
	if body != nil {
		reader = bytes.NewReader(body)
	}
	req, err := http.NewRequest(method, target, reader)
	if err != nil {
		return err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		if errors.Is(err, syscall.ECONNREFUSED) {
			return fmt.Errorf("%s %s: %w\n(is the server running? try '__PROJECT_NAME__ server')", method, target, err)
		}
		return fmt.Errorf("%s %s: %w", method, target, err)
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("%s %s returned %d: %s", method, target, resp.StatusCode, bodySnippet(data))
	}
	if method == http.MethodGet || opts.json {
		printBody(data)
	}
	return nil
}

// printBody writes the response body verbatim with a trailing newline.
func printBody(data []byte) {
	fmt.Print(string(data))
	if len(data) == 0 || data[len(data)-1] != '\n' {
		fmt.Println()
	}
}

// bodySnippet trims a response body for error messages.
func bodySnippet(data []byte) string {
	snippet := strings.TrimSpace(string(data))
	if len(snippet) > 300 {
		snippet = snippet[:300] + "…"
	}
	return snippet
}
