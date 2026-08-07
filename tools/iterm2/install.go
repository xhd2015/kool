package iterm2

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	iterm2install "github.com/xhd2015/dot-pkgs/go-pkgs/shell/iterm2/install"
	"github.com/xhd2015/kool/pkgs/errs"
	lessflags "github.com/xhd2015/less-flags"
)

const installHelp = `iterm2 install — download/install official iTerm2 (no Homebrew)

Usage:
  kool iterm2 install [flags]
  kool iterm2 install -h|--help

Installs the latest stable iTerm2 zip from iterm2.com into ~/Applications/iTerm.app
by default (via github.com/xhd2015/dot-pkgs/go-pkgs/shell/iterm2/install).

Flags:
  --dry-run              print plan only (resolve URL; no download/install)
  --download-dir <dir>   directory for the zip (and install cache workdir)
  --download-only        download zip only; do not install to Applications
  -h, --help             show this help

Examples:
  kool iterm2 install
  kool iterm2 install --dry-run
  kool iterm2 install --download-dir ~/Downloads
  kool iterm2 install --download-only --download-dir ~/Downloads
`

// installHTTPClient is overridable in tests (nil → http.DefaultClient / library default).
var installHTTPClient *http.Client

// installLatestURL overrides the library latest endpoint in tests when non-empty.
var installLatestURL string

// installHome overrides the install home directory in tests when non-empty.
var installHome string

// runInstall handles: kool iterm2 install [flags]
func runInstall(args []string, stdout, stderr io.Writer) error {
	var dryRun bool
	var downloadOnly bool
	var downloadDir string

	remain, err := lessflags.Bool("--dry-run", &dryRun).
		Bool("--download-only", &downloadOnly).
		String("--download-dir", &downloadDir).
		HelpFunc("-h,--help", func() {}).
		HelpNoExit().
		Parse(args)
	if err != nil {
		if err == lessflags.ErrHelp {
			fmt.Fprint(stdout, strings.TrimSpace(installHelp)+"\n")
			return nil
		}
		fmt.Fprint(stderr, err.Error())
		return errs.NewSilenceExitCode(1)
	}
	if len(remain) > 0 {
		fmt.Fprintf(stderr, "Error: unexpected arguments: %s\nRun 'kool iterm2 install --help' for usage.\n", strings.Join(remain, " "))
		return errs.NewSilenceExitCode(1)
	}

	ctx := context.Background()
	resolveOpts := iterm2install.ResolveOpts{
		LatestURL:  installLatestURL,
		HTTPClient: installHTTPClient,
	}

	url, version, err := iterm2install.ResolveLatestStableURL(ctx, resolveOpts)
	if err != nil {
		fmt.Fprintf(stderr, "Error: %v\n", err)
		return errs.NewSilenceExitCode(1)
	}

	targetApp, err := defaultInstallTarget()
	if err != nil {
		fmt.Fprintf(stderr, "Error: %v\n", err)
		return errs.NewSilenceExitCode(1)
	}

	cacheDir, err := resolveDownloadDir(downloadDir)
	if err != nil {
		fmt.Fprintf(stderr, "Error: %v\n", err)
		return errs.NewSilenceExitCode(1)
	}
	zipPath := plannedZipPath(cacheDir, url)

	if dryRun {
		fmt.Fprintf(stdout, "dry-run: would install iTerm2 %s\n", version)
		fmt.Fprintf(stdout, "  url:    %s\n", url)
		fmt.Fprintf(stdout, "  zip:    %s\n", zipPath)
		if downloadOnly {
			fmt.Fprintf(stdout, "  mode:   download-only (skip Applications install)\n")
			fmt.Fprintf(stdout, "  steps:  download\n")
		} else {
			fmt.Fprintf(stdout, "  target: %s\n", targetApp)
			fmt.Fprintf(stdout, "  steps:  download, extract, install, register, verify\n")
		}
		return nil
	}

	if downloadOnly {
		if err := os.MkdirAll(cacheDir, 0o755); err != nil {
			fmt.Fprintf(stderr, "Error: mkdir download dir: %v\n", err)
			return errs.NewSilenceExitCode(1)
		}
		fmt.Fprintf(stdout, "iTerm2  resolve  %s (%s)\n", url, version)
		if err := iterm2install.Download(ctx, url, zipPath, iterm2install.DownloadOpts{
			HTTPClient: installHTTPClient,
		}); err != nil {
			fmt.Fprintf(stderr, "Error: %v\n", err)
			return errs.NewSilenceExitCode(1)
		}
		fmt.Fprintf(stdout, "iTerm2  download %s\n", zipPath)
		return nil
	}

	// Full install via library.
	fmt.Fprintf(stdout, "iTerm2  resolve  %s (%s)\n", url, version)
	opts := iterm2install.InstallOpts{
		Home:           installHome,
		CacheDir:       cacheDir,
		LatestURL:      installLatestURL,
		HTTPClient:     installHTTPClient,
		SkipScriptable: false,
	}
	// InstallLatest re-resolves; that's fine. Ensure CacheDir is set so zip lands under --download-dir.
	result, err := iterm2install.InstallLatest(ctx, opts)
	if err != nil {
		fmt.Fprintf(stderr, "Error: %v\n", err)
		return errs.NewSilenceExitCode(1)
	}
	fmt.Fprintf(stdout, "iTerm2  download %s\n", result.ZipPath)
	fmt.Fprintf(stdout, "iTerm2  install  %s\n", result.AppPath)
	if result.BackupPath != "" {
		fmt.Fprintf(stdout, "iTerm2  backup   %s\n", result.BackupPath)
	}
	fmt.Fprintf(stdout, "iTerm2  verify   ok (%s)\n", result.Version)
	return nil
}

func defaultInstallTarget() (string, error) {
	if installHome != "" {
		return filepath.Join(installHome, "Applications", iterm2install.AppBundleName), nil
	}
	return iterm2install.HomeApplicationsITermApp()
}

// resolveDownloadDir returns the cache/download directory.
// Empty downloadDir → UserCacheDir/kool-iterm2 (or temp fallback).
func resolveDownloadDir(downloadDir string) (string, error) {
	if strings.TrimSpace(downloadDir) != "" {
		return filepath.Abs(downloadDir)
	}
	if cache, err := os.UserCacheDir(); err == nil && cache != "" {
		return filepath.Join(cache, "kool-iterm2"), nil
	}
	return filepath.Join(os.TempDir(), "kool-iterm2"), nil
}

func plannedZipPath(cacheDir, url string) string {
	zipName := filepath.Base(strings.Split(url, "?")[0])
	if zipName == "" || zipName == "." || zipName == "/" {
		zipName = fmt.Sprintf("iTerm2-%d.zip", time.Now().Unix())
	}
	return filepath.Join(cacheDir, zipName)
}
