//go:build ignore

package run

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"
)

// imageLibraryItem is the CLI's view of one library record. Path is the
// absolute path of the stored bytes, so `get` can hand an agent a file it can
// open instead of an <img> it cannot see.
type imageLibraryItem struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Source  string `json:"source"`
	Ext     string `json:"ext"`
	URL     string `json:"url"`
	Path    string `json:"path"`
	Bytes   int64  `json:"bytes"`
	Valid   bool   `json:"valid"`
	Problem string `json:"problem"`
	Reused  bool   `json:"reused"`
	Usages  []struct {
		Kind     string `json:"kind"`
		PagePath string `json:"page_path"`
	} `json:"usages"`
}

// imageLibraryPath reports whether a target addresses the image library, and
// returns the image id when it addresses one image.
func imageLibraryPath(target string) (collection bool, id string, ok bool) {
	parsed, err := url.Parse(target)
	if err != nil {
		return false, "", false
	}
	path := strings.TrimSuffix(parsed.Path, "/")
	if path == "/api/images" {
		return true, "", true
	}
	if rest, found := strings.CutPrefix(path, "/api/images/"); found && rest != "" && !strings.Contains(rest, "/") {
		return false, rest, true
	}
	return false, "", false
}

// renderImageGet renders a library read for a human and reports whether it
// handled the path. Anything else stays the caller's business, so the generic
// verb keeps printing the body verbatim.
func renderImageGet(target string, data []byte, stdout, stderr io.Writer) (bool, error) {
	collection, id, ok := imageLibraryPath(target)
	if !ok {
		return false, nil
	}
	if collection {
		return true, renderImageLibrary(data, stdout, stderr)
	}
	return true, renderImage(id, data, stdout, stderr)
}

// renderImageLibrary lists the whole library, which is how a file that is not
// a picture, an unused image, or a stale reference becomes visible at all.
func renderImageLibrary(data []byte, stdout, stderr io.Writer) error {
	var images []imageLibraryItem
	if err := json.Unmarshal(data, &images); err != nil {
		return fmt.Errorf("parse image library: %w", err)
	}
	problems, unused := 0, 0
	table := tabwriter.NewWriter(stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(table, "ID\tSize\tFormat\tUsed by\tState")
	for _, img := range images {
		if !img.Valid {
			problems++
		}
		if len(img.Usages) == 0 && img.Ext != "" {
			unused++
		}
		fmt.Fprintf(table, "%s\t%s\t%s\t%s\t%s\n",
			img.ID, sizeText(img.Bytes), dashIfEmpty(img.Ext), usageText(img), stateText(img))
	}
	if err := table.Flush(); err != nil {
		return err
	}
	summary := pluralize(len(images), "image")
	if problems > 0 {
		summary += fmt.Sprintf(" · %d not an image", problems)
	}
	if unused > 0 {
		summary += fmt.Sprintf(" · %d unused", unused)
	}
	fmt.Fprintf(stdout, "\n%s\n", summary)
	if problems > 0 {
		// A file that is not a picture is a warning, not a failed read: a
		// script watching stderr sees it without the command failing.
		fmt.Fprintf(stderr, "warning: %d of %d library images are not pictures; run '%s get /api/images' to list them\n",
			problems, len(images), progName)
	}
	return nil
}

// renderImage reports one library image. The first line is the address an
// agent can act on — the absolute path of the stored bytes, or the source URL
// when the record has no local file — so `get` output pipes into a file
// operation instead of an HTML parser.
func renderImage(id string, data []byte, stdout, stderr io.Writer) error {
	var img imageLibraryItem
	if err := json.Unmarshal(data, &img); err != nil {
		return fmt.Errorf("parse image %s: %w", id, err)
	}
	if img.ID == "" {
		return fmt.Errorf("image %s not found", id)
	}
	if address := imageAddress(img); address != "" {
		fmt.Fprintln(stdout, address)
	}
	if img.Name != "" {
		fmt.Fprintf(stdout, "name: %s\n", img.Name)
	}
	format := dashIfEmpty(img.Ext)
	if img.Bytes > 0 {
		format += " · " + sizeText(img.Bytes)
	}
	fmt.Fprintf(stdout, "format: %s\n", format)
	if img.Source != "" {
		fmt.Fprintf(stdout, "source: %s\n", img.Source)
	}
	fmt.Fprintf(stdout, "state: %s\n", stateText(img))
	if used := usageText(img); used != "-" {
		fmt.Fprintf(stdout, "used by: %s\n", used)
	}
	if !img.Valid {
		fmt.Fprintf(stderr, "warning: image %s is %s\n", img.ID, img.Problem)
	}
	return nil
}

// imageAddress is the local path of the stored bytes when the server reported
// one, else the record's URL. An absolute URL is printed as-is: joining an
// origin onto it would produce http://hosthttps://elsewhere/x.jpg.
func imageAddress(img imageLibraryItem) string {
	if strings.TrimSpace(img.Path) != "" {
		if _, err := os.Stat(img.Path); err == nil {
			return img.Path
		}
	}
	return strings.TrimSpace(img.URL)
}

func usageText(img imageLibraryItem) string {
	if len(img.Usages) == 0 {
		return "-"
	}
	first := img.Usages[0]
	text := first.PagePath
	if text == "" {
		text = first.Kind
	}
	if len(img.Usages) > 1 {
		text += fmt.Sprintf(" +%d", len(img.Usages)-1)
	}
	return text
}

func stateText(img imageLibraryItem) string {
	if !img.Valid {
		return "NOT AN IMAGE (" + img.Problem + ")"
	}
	if img.Ext == "" {
		// Lives at its Source URL; there are no local bytes to check.
		return "external"
	}
	if len(img.Usages) == 0 {
		return "unused"
	}
	return "ok"
}

// uploadImage posts a file to the image library as multipart form data and
// returns the created record. Uploading does not attach it to anything: the
// caller adds the returned id to whatever container shows it.
func uploadImage(target, file, name, caption string) (imageLibraryItem, error) {
	f, err := os.Open(file)
	if err != nil {
		return imageLibraryItem{}, fmt.Errorf("open image: %w", err)
	}
	defer f.Close()

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	for key, value := range map[string]string{"name": name, "caption": caption} {
		if strings.TrimSpace(value) == "" {
			continue
		}
		if err := writer.WriteField(key, value); err != nil {
			return imageLibraryItem{}, err
		}
	}
	part, err := writer.CreateFormFile("file", filepath.Base(file))
	if err != nil {
		return imageLibraryItem{}, err
	}
	if _, err := io.Copy(part, f); err != nil {
		return imageLibraryItem{}, fmt.Errorf("read image: %w", err)
	}
	if err := writer.Close(); err != nil {
		return imageLibraryItem{}, err
	}

	req, err := http.NewRequest(http.MethodPost, target, &buf)
	if err != nil {
		return imageLibraryItem{}, err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	data, err := send(req)
	if err != nil {
		return imageLibraryItem{}, err
	}
	var image imageLibraryItem
	if err := json.Unmarshal(data, &image); err != nil {
		return imageLibraryItem{}, fmt.Errorf("parse upload: %w", err)
	}
	return image, nil
}

// reportUpload prints where the bytes landed. Both addresses are printed: the
// path is what a local step opens, the URL is what a page shows.
func reportUpload(image imageLibraryItem, stdout io.Writer) {
	verb := "uploaded"
	if image.Reused {
		verb = "reused"
	}
	name := image.Name
	if name == "" {
		name = "image"
	}
	fmt.Fprintf(stdout, "%s image %s · %s", verb, image.ID, name)
	if image.Reused {
		fmt.Fprint(stdout, " (identical bytes already stored)")
	}
	fmt.Fprintln(stdout)
	if image.Path != "" {
		fmt.Fprintf(stdout, "path %s\n", image.Path)
	}
	if image.URL != "" {
		fmt.Fprintf(stdout, "url  %s\n", image.URL)
	}
}

func sizeText(n int64) string {
	if n <= 0 {
		return "-"
	}
	return humanBytes(n)
}

// humanBytes renders a byte count for a table column.
func humanBytes(n int64) string {
	const unit = 1024
	if n < unit {
		return fmt.Sprintf("%d B", n)
	}
	value := float64(n)
	for _, suffix := range []string{"KB", "MB", "GB"} {
		value /= unit
		if value < unit {
			return fmt.Sprintf("%.1f %s", value, suffix)
		}
	}
	return fmt.Sprintf("%.1f TB", value/unit)
}

func dashIfEmpty(s string) string {
	if strings.TrimSpace(s) == "" {
		return "-"
	}
	return s
}

func pluralize(n int, noun string) string {
	if n == 1 {
		return fmt.Sprintf("%d %s", n, noun)
	}
	return fmt.Sprintf("%d %ss", n, noun)
}
