//go:build ignore

package run

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"syscall"

	lessflags "github.com/xhd2015/less-flags"
)

// defaultPort is the port the server listens on when auto-selecting and the
// port bare URIs (e.g. /api/counter) talk to.
const defaultPort = 8080

// progName is the generated binary's name, used in hints and warnings.
const progName = "__PROJECT_NAME__"

const clientFlagsHelpSuffix = `Connection options:
  --port <n>       talk to the __PROJECT_NAME__ server on this port (default: 8080)
  --url <url>      talk to this server origin, e.g. http://127.0.0.1:9000
  -h, --help       show help
`

const getHelp = `Usage: __PROJECT_NAME__ get <URI> [--json] [--no-verify] [--port <n> | --url <url>]

GET a URI from the running __PROJECT_NAME__ server and print the response body.

<URI> is a path ("/api/counter") or a full URL ("http://localhost:8080/api/counter");
a bare path talks to http://localhost:8080. A query string is kept
("/plans?stage=design"). --json is accepted for symmetry and prints the body
verbatim.

The image library is rendered for a human instead of dumped as JSON:
  get /api/images         a table of every image, its size, where it is used
                          and its state, plus a summary line
  get /api/images/<id>    the image's state, and its absolute file path on the
                          first line so a local step can open it

--no-verify skips the byte check (the audit then reports no state), which is
the only reason to turn it off; references are still listed.

` + clientFlagsHelpSuffix + `
Examples:
  __PROJECT_NAME__ get /ping
  __PROJECT_NAME__ get /api/counter
  __PROJECT_NAME__ get /api/images
  __PROJECT_NAME__ get /api/images/3
  __PROJECT_NAME__ get http://localhost:9000/api/counter
`

const putHelp = `Usage: __PROJECT_NAME__ put <URI> [<json|@file|->] [--json] [--dry-run] [--port <n> | --url <url>]

Replace a resource on the running __PROJECT_NAME__ server. The body is a JSON
literal, @file, or - for stdin. On success prints nothing unless --json is
given.

` + clientFlagsHelpSuffix + `
Examples:
  __PROJECT_NAME__ put /api/plan/1 '{"title":"Cubic roots"}'
  __PROJECT_NAME__ put /api/gallery '{"image_ids":["1","2"]}'
  cat plan.json | __PROJECT_NAME__ put /api/plan/1 -
`

const postHelp = `Usage: __PROJECT_NAME__ post <URI> [<json|@file|->] [--json] [--dry-run] [--port <n> | --url <url>]
       __PROJECT_NAME__ post /api/images --file <path> [--name <text>] [--caption <text>]

Create a resource on the running __PROJECT_NAME__ server. The body is optional
(the counter takes none), otherwise a JSON literal, @file, or - for stdin. On
success prints nothing unless --json is given.

--file uploads an image into the shared library instead of sending a JSON body:
the bytes are stored once (identical bytes reuse the existing record, so no new
id is allocated) and the reply names the new record, its absolute path and its
URL. Attach the printed id to whatever shows the picture.

` + clientFlagsHelpSuffix + `
Examples:
  __PROJECT_NAME__ post /api/counter
  __PROJECT_NAME__ post /api/images --file ~/Pictures/chengdu.jpg --name 成都
  __PROJECT_NAME__ post /api/images --file shot.png --json
`

const deleteHelp = `Usage: __PROJECT_NAME__ delete <URI> [--force] [--json] [--dry-run] [--port <n> | --url <url>]

Delete a resource on the running __PROJECT_NAME__ server. On success prints
nothing unless --json is given.

An image that a page shows is refused, and the refusal names the page: deleting
it would leave a broken picture behind. --force deletes it anyway and detaches
every reference first, so no page is left pointing at an id that is gone.

` + clientFlagsHelpSuffix + `
Examples:
  __PROJECT_NAME__ delete /api/images/3
  __PROJECT_NAME__ delete /api/images/3 --force
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
	var asJSON, noVerify bool
	var flags clientFlags
	remain, err := lessflags.
		Bool("--json", &asJSON).
		Bool("--no-verify", &noVerify).
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
	if noVerify {
		target, err = withQuery(target, "verify", "0")
		if err != nil {
			return err
		}
	}
	data, err := doRequestData(http.MethodGet, target, nil, requestOptions{json: asJSON})
	if err != nil {
		return err
	}
	if asJSON {
		// --json always prints the machine-readable body, whatever the path.
		printBody(data)
		return nil
	}
	// The image library reads for a human: a table for the collection, the
	// address an agent can open for one image. Every other path prints the
	// body verbatim.
	if handled, err := renderImageGet(target, data, os.Stdout, os.Stderr); handled {
		return err
	}
	printBody(data)
	return nil
}

// withQuery adds one query parameter to a target URL.
func withQuery(target, key, value string) (string, error) {
	parsed, err := url.Parse(target)
	if err != nil {
		return "", err
	}
	query := parsed.Query()
	query.Set(key, value)
	parsed.RawQuery = query.Encode()
	return parsed.String(), nil
}

func handlePut(args []string) error {
	return handleWrite(http.MethodPut, putHelp, args)
}

func handlePost(args []string) error {
	return handleWrite(http.MethodPost, postHelp, args)
}

func handleDelete(args []string) error {
	var asJSON, dryRun, force bool
	var flags clientFlags
	remain, err := lessflags.
		Bool("--dry-run", &dryRun).
		Bool("--json", &asJSON).
		Bool("--force", &force).
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
	if force {
		if target, err = withQuery(target, "force", "1"); err != nil {
			return err
		}
	}
	return doRequest(http.MethodDelete, target, nil, requestOptions{json: asJSON, dryRun: dryRun})
}

// handleWrite runs put/post: one URI plus an optional body. A post may upload
// an image with --file instead of sending a JSON body.
func handleWrite(method string, helpText string, args []string) error {
	var asJSON, dryRun bool
	var file, name, caption string
	var flags clientFlags
	remain, err := lessflags.
		Bool("--dry-run", &dryRun).
		Bool("--json", &asJSON).
		String("--file", &file).
		String("--name", &name).
		String("--caption", &caption).
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
	if file != "" {
		return handleImageUpload(method, target, file, name, caption, len(remain) > 1, requestOptions{json: asJSON, dryRun: dryRun})
	}
	if name != "" || caption != "" {
		return fmt.Errorf("--name and --caption describe an upload; pass --file <path>\nRun '__PROJECT_NAME__ %s --help' for usage", strings.ToLower(method))
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

// handleImageUpload runs `post <library-uri> --file <path>`: upload the bytes
// into the shared image library and report where they landed.
func handleImageUpload(method, target, file, name, caption string, hasBody bool, opts requestOptions) error {
	if method != http.MethodPost {
		return fmt.Errorf("--file uploads to the image library with post, not %s\nRun '__PROJECT_NAME__ post /api/images --file <path>'", strings.ToLower(method))
	}
	if collection, _, ok := imageLibraryPath(target); !ok || !collection {
		return fmt.Errorf("--file uploads to /api/images, not %s", target)
	}
	if hasBody {
		return fmt.Errorf("--file uploads the bytes itself; drop the JSON body")
	}
	if opts.dryRun {
		fmt.Printf("POST %s (multipart file=%s)\n", target, file)
		return nil
	}
	if _, err := os.Stat(file); err != nil {
		return fmt.Errorf("open image: %w", err)
	}
	image, err := uploadImage(target, file, name, caption)
	if err != nil {
		return err
	}
	if opts.json {
		data, err := json.Marshal(image)
		if err != nil {
			return err
		}
		printBody(data)
		return nil
	}
	reportUpload(image, os.Stdout)
	return nil
}

// doRequest sends one HTTP request and prints the result: the body verbatim
// for get; silent on 2xx for writes unless --json; a non-2xx status is an
// error carrying the status and a body snippet.
func doRequest(method string, target string, body []byte, opts requestOptions) error {
	data, err := doRequestData(method, target, body, opts)
	if err != nil {
		return err
	}
	if method == http.MethodGet || opts.json {
		printBody(data)
	}
	return nil
}

// doRequestData sends one HTTP request and returns the response body. It is
// silent: the caller decides whether to print the body, render it, or parse
// it.
func doRequestData(method string, target string, body []byte, opts requestOptions) ([]byte, error) {
	if opts.dryRun {
		fmt.Printf("%s %s\n", method, target)
		if len(body) > 0 {
			fmt.Println(string(body))
		}
		return nil, nil
	}
	var reader io.Reader
	if body != nil {
		reader = bytes.NewReader(body)
	}
	req, err := http.NewRequest(method, target, reader)
	if err != nil {
		return nil, err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	return send(req)
}

// send performs one request and returns the body, mapping a transport failure
// and a non-2xx status to an error the caller can print.
func send(req *http.Request) ([]byte, error) {
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		if errors.Is(err, syscall.ECONNREFUSED) {
			return nil, fmt.Errorf("%s %s: %w\n(is the server running? try '__PROJECT_NAME__ server')", req.Method, req.URL, err)
		}
		return nil, fmt.Errorf("%s %s: %w", req.Method, req.URL, err)
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("%s %s returned %d: %s", req.Method, req.URL, resp.StatusCode, errorDetail(data))
	}
	return data, nil
}

// errorDetail is what the server said went wrong. The API answers with
// {"error": "..."}, so the sentence is shown instead of raw JSON; anything
// else falls back to a trimmed body snippet.
func errorDetail(data []byte) string {
	var payload struct {
		Error string `json:"error"`
	}
	if json.Unmarshal(data, &payload) == nil && strings.TrimSpace(payload.Error) != "" {
		return strings.TrimSpace(payload.Error)
	}
	return bodySnippet(data)
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
