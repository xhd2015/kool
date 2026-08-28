package cloudflare

import (
	"strings"
	"testing"
	"time"

	"github.com/xhd2015/dot-pkgs/go-pkgs/terminal/color"
)

func TestIsTunnelReady(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name  string
		code  int
		body  string
		ready bool
	}{
		{name: "zero code", code: 0, body: "", ready: false},
		{name: "cf 530 1033", code: 530, body: "error code: 1033", ready: false},
		{name: "cf 530 html 1033", code: 530, body: "<html>error code: 1033</html>", ready: false},
		{name: "200 ok", code: 200, body: "ok", ready: true},
		{name: "404 origin", code: 404, body: "missing", ready: true},
		{name: "502 origin", code: 502, body: "bad gateway", ready: true},
		{name: "530 without 1033", code: 530, body: "other", ready: true},
		{name: "301 redirect", code: 301, body: "", ready: true},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := IsTunnelReady(tc.code, tc.body)
			if got != tc.ready {
				t.Fatalf("IsTunnelReady(%d, %q)=%v want %v", tc.code, tc.body, got, tc.ready)
			}
		})
	}
}

func TestPublicProbeURL(t *testing.T) {
	t.Parallel()
	cases := []struct {
		domain, path, want string
	}{
		{"a.example.com", "", "https://a.example.com/"},
		{"a.example.com", "/", "https://a.example.com/"},
		{"https://A.Example.COM/", "health", "https://a.example.com/health"},
		{"a.example.com", "/ping", "https://a.example.com/ping"},
	}
	for _, tc := range cases {
		got := PublicProbeURL(tc.domain, tc.path)
		if got != tc.want {
			t.Fatalf("PublicProbeURL(%q,%q)=%q want %q", tc.domain, tc.path, got, tc.want)
		}
	}
}

func TestFormatWaitingProgress(t *testing.T) {
	t.Parallel()
	got := FormatWaitingProgress(530, 12*time.Second, "error code: 1033")
	want := "waiting  12s  HTTP 530  error code: 1033"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestFormatReadyTimeoutWarning_color(t *testing.T) {
	t.Parallel()
	on := FormatReadyTimeoutWarning(color.Style{Enabled: true}, "https://a.example.com/", 5*time.Second)
	if !strings.Contains(on, "warning: public ready timeout") {
		t.Fatalf("missing warning text: %q", on)
	}
	if !strings.Contains(on, "\x1b[33m") {
		t.Fatalf("expected yellow SGR when color on: %q", on)
	}
	off := FormatReadyTimeoutWarning(color.Style{Enabled: false}, "https://a.example.com/", 5*time.Second)
	if strings.Contains(off, "\x1b") {
		t.Fatalf("expected no ANSI when color off: %q", off)
	}
}

func TestFormatPublicReadyLine_color(t *testing.T) {
	t.Parallel()
	on := FormatPublicReadyLine(color.Style{Enabled: true}, "https://a.example.com/", 200, 6*time.Second)
	if !strings.Contains(on, "Public ready: https://a.example.com/ (200) after 6s") {
		t.Fatalf("missing ready text: %q", on)
	}
	if !strings.Contains(on, "\x1b[32m") {
		t.Fatalf("expected green SGR when color on: %q", on)
	}
}
