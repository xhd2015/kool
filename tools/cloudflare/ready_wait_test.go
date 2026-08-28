package cloudflare

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/xhd2015/dot-pkgs/go-pkgs/terminal/color"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) Do(req *http.Request) (*http.Response, error) { return f(req) }

func TestWaitPublicReady_becomesReady(t *testing.T) {
	t.Parallel()
	var calls int
	client := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		calls++
		if calls < 2 {
			return &http.Response{
				StatusCode: 530,
				Body:       io.NopCloser(strings.NewReader("error code: 1033")),
			}, nil
		}
		return &http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(strings.NewReader("ok")),
		}, nil
	})

	now := time.Unix(0, 0)
	var log strings.Builder
	elapsed, err := WaitPublicReady(context.Background(), WaitReadyOptions{
		Client:       client,
		URL:          "https://a.example.com/",
		Timeout:      30 * time.Second,
		PollInterval: time.Millisecond,
		Log:          &log,
		Interactive:  false,
		Style:        color.Style{Enabled: false},
		Now: func() time.Time {
			// advance a little each call so elapsed moves
			n := now
			now = now.Add(time.Second)
			return n
		},
		Sleep: func(d time.Duration) {},
	})
	if err != nil {
		t.Fatalf("WaitPublicReady: %v", err)
	}
	if calls < 2 {
		t.Fatalf("expected at least 2 probes, got %d", calls)
	}
	if elapsed <= 0 {
		t.Fatalf("expected positive elapsed, got %v", elapsed)
	}
	out := log.String()
	if !strings.Contains(out, "waiting") {
		t.Fatalf("expected waiting progress: %q", out)
	}
	if !strings.Contains(out, "Public ready:") {
		t.Fatalf("expected ready line: %q", out)
	}
}

func TestWaitPublicReady_timeout(t *testing.T) {
	t.Parallel()
	client := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: 530,
			Body:       io.NopCloser(strings.NewReader("error code: 1033")),
		}, nil
	})
	start := time.Unix(100, 0)
	now := start
	_, err := WaitPublicReady(context.Background(), WaitReadyOptions{
		Client:       client,
		URL:          "https://a.example.com/",
		Timeout:      3 * time.Second,
		PollInterval: time.Millisecond,
		Log:          io.Discard,
		Now: func() time.Time {
			n := now
			now = now.Add(2 * time.Second)
			return n
		},
		Sleep: func(d time.Duration) {},
	})
	if !errors.Is(err, ErrReadyTimeout) {
		t.Fatalf("err=%v want ErrReadyTimeout", err)
	}
}
