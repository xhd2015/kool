package preview

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestDOTPreview(t *testing.T) {
	if _, err := exec.LookPath("playwright-debug"); err != nil {
		t.Skip("playwright-debug not available; install with: npm install -g playwright-debug")
	}

	_, thisFile, _, _ := runtime.Caller(0)
	projectRoot := filepath.Join(filepath.Dir(thisFile), "..", "..")

	tmpDir := t.TempDir()

	dotContent := `digraph test {
    "Hello" [shape=box];
    "World" [shape=ellipse];
    "Hello" -> "World";
}
`

	mdContent := `# Test DOT in Markdown

Some text before.

` + "```dot\n" + `digraph inline {
    A [shape=box];
    B [shape=diamond];
    A -> B [label="test"];
}
` + "```\n\n" + `Some text after.
`

	if err := os.WriteFile(filepath.Join(tmpDir, "test_dot.dot"), []byte(dotContent), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "test_dot.md"), []byte(mdContent), 0644); err != nil {
		t.Fatal(err)
	}

	koolBin := filepath.Join(tmpDir, "kool")
	buildCmd := exec.Command("go", "build", "-o", koolBin)
	buildCmd.Dir = projectRoot
	if out, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to build kool: %v\n%s", err, out)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	previewCmd := exec.CommandContext(ctx, koolBin, "preview", "--no-watch", tmpDir)
	previewCmd.Dir = tmpDir

	stdoutPipe, err := previewCmd.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	stderrPipe, err := previewCmd.StderrPipe()
	if err != nil {
		t.Fatal(err)
	}

	if err := previewCmd.Start(); err != nil {
		t.Fatalf("failed to start kool preview: %v", err)
	}
	defer func() {
		previewCmd.Process.Kill()
		previewCmd.Wait()
	}()

	urlCh := make(chan string, 1)
	go func() {
		buf := make([]byte, 4096)
		remains := ""
		for {
			n, err := stdoutPipe.Read(buf)
			remains += string(buf[:n])
			lines := strings.Split(remains, "\n")
			remains = lines[len(lines)-1]
			for _, line := range lines[:len(lines)-1] {
				if strings.HasPrefix(line, "Serving directory preview at ") {
					urlCh <- strings.TrimPrefix(line, "Serving directory preview at ")
					return
				}
			}
			if err != nil {
				break
			}
		}
	}()
	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := stderrPipe.Read(buf)
			if n > 0 {
				t.Logf("[kool stderr] %s", string(buf[:n]))
			}
			if err != nil {
				break
			}
		}
	}()

	var serverURL string
	select {
	case url := <-urlCh:
		serverURL = strings.TrimSpace(url)
	case <-time.After(15 * time.Second):
		t.Fatal("timed out waiting for server URL")
	}

	// Poll until server is ready
	ready := false
	for i := 0; i < 30; i++ {
		resp, err := http.Get(serverURL + "/api/tree")
		if err == nil && resp.StatusCode == 200 {
			resp.Body.Close()
			ready = true
			break
		}
		if resp != nil {
			resp.Body.Close()
		}
		time.Sleep(500 * time.Millisecond)
	}
	if !ready {
		t.Fatal("server did not become ready")
	}

	playwrightScript := `await page.goto("` + serverURL + `");

// Wait for file tree to load
await page.waitForSelector('[data-path="test_dot.dot"]', { timeout: 10000 });

// Test standalone .dot file
await page.click('[data-path="test_dot.dot"]');
await page.waitForSelector('.dot-container svg', { timeout: 15000 });
const standaloneSvgCount = await page.$$eval('.dot-container svg', els => els.length);
if (standaloneSvgCount === 0) {
    throw new Error('standalone .dot did not render SVG in .dot-container');
}
console.log('PASS: standalone .dot rendered as SVG');

// Test inline dot in markdown
await page.click('[data-path="test_dot.md"]');
// V1 uses .dot-container, V2 uses .dot-container-v2
await page.waitForSelector('.dot-container svg, .dot-container-v2 svg', { timeout: 15000 });
console.log('PASS: inline ` + "```dot" + ` in markdown rendered as SVG');

console.log('ALL TESTS PASSED');`

	fmt.Printf("serverURL: %s\n", serverURL)
	fmt.Printf("playwrightScript:\n%s\n", playwrightScript)

	testCmd := exec.CommandContext(ctx, "playwright-debug", "run", playwrightScript)
	testOut, err := testCmd.CombinedOutput()
	t.Logf("playwright output:\n%s", string(testOut))
	if err != nil {
		t.Fatalf("playwright test failed: %v", err)
	}
}
