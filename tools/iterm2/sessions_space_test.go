package iterm2

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/xhd2015/dot-pkgs/go-pkgs/computer-use/macos/space"
)

func stubCurrentSpace(t *testing.T, idx int) {
	t.Helper()
	SetCurrentSpaceIndexForTest(func() (int, error) { return idx, nil })
	t.Cleanup(func() { SetCurrentSpaceIndexForTest(nil) })
}

func stubCurrentSpaceErr(t *testing.T, err error) {
	t.Helper()
	SetCurrentSpaceIndexForTest(func() (int, error) { return 0, err })
	t.Cleanup(func() { SetCurrentSpaceIndexForTest(nil) })
}

func TestEnsureSpacePlacement_SwitchOK(t *testing.T) {
	stubCurrentSpace(t, 0) // not on target space 1
	mock := &space.MockBackend{Desktops: []int{1, 2}}
	SetSpaceBackendForTest(mock)
	t.Cleanup(func() { SetSpaceBackendForTest(nil) })
	SetSpaceSwitchSettleForTest(0)
	t.Cleanup(func() { SetSpaceSwitchSettleForTest(-1) })

	warns, err := ensureSpacePlacement(1) // Desktop 2
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if len(warns) != 0 {
		t.Fatalf("warns: %v", warns)
	}
	if len(mock.Switched) != 1 || mock.Switched[0] != 2 {
		t.Fatalf("Switched=%v", mock.Switched)
	}
}

func TestEnsureSpacePlacement_AlreadyOnTargetSkipsSwitch(t *testing.T) {
	stubCurrentSpace(t, 1)
	mock := &space.MockBackend{Desktops: []int{1, 2}}
	SetSpaceBackendForTest(mock)
	t.Cleanup(func() { SetSpaceBackendForTest(nil) })
	SetSpaceSwitchSettleForTest(0)
	t.Cleanup(func() { SetSpaceSwitchSettleForTest(-1) })

	warns, err := ensureSpacePlacement(1)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if len(warns) != 0 {
		t.Fatalf("warns: %v", warns)
	}
	if len(mock.Switched) != 0 {
		t.Fatalf("expected no Switch when already on target, Switched=%v", mock.Switched)
	}
}

func TestEnsureSpacePlacement_AlreadyOnSpace0SkipsSwitch(t *testing.T) {
	stubCurrentSpace(t, 0)
	mock := &space.MockBackend{Desktops: []int{1}}
	SetSpaceBackendForTest(mock)
	t.Cleanup(func() { SetSpaceBackendForTest(nil) })

	warns, err := ensureSpacePlacement(0)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if len(warns) != 0 {
		t.Fatalf("warns: %v", warns)
	}
	if len(mock.Switched) != 0 {
		t.Fatalf("Switched=%v", mock.Switched)
	}
}

func TestEnsureSpacePlacement_CurrentLookupFailFallsThrough(t *testing.T) {
	stubCurrentSpaceErr(t, errors.New("cgs unavailable"))
	mock := &space.MockBackend{Desktops: []int{1, 2}}
	SetSpaceBackendForTest(mock)
	t.Cleanup(func() { SetSpaceBackendForTest(nil) })
	SetSpaceSwitchSettleForTest(0)
	t.Cleanup(func() { SetSpaceSwitchSettleForTest(-1) })

	warns, err := ensureSpacePlacement(1)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if len(warns) != 0 {
		t.Fatalf("warns: %v", warns)
	}
	if len(mock.Switched) != 1 || mock.Switched[0] != 2 {
		t.Fatalf("Switched=%v want [2]", mock.Switched)
	}
}

func TestEnsureSpacePlacement_RetryThenOK(t *testing.T) {
	stubCurrentSpace(t, 0)
	b := &flakySwitchBackend{
		MockBackend: space.MockBackend{Desktops: []int{1, 2}},
		failTimes:   2,
		failErr:     fmt.Errorf("FAIL: desktop not found: Desktop 2"),
	}
	SetSpaceBackendForTest(b)
	t.Cleanup(func() { SetSpaceBackendForTest(nil) })
	SetSpaceSwitchSettleForTest(0)
	t.Cleanup(func() { SetSpaceSwitchSettleForTest(-1) })

	warns, err := ensureSpacePlacement(1)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if len(warns) != 0 {
		t.Fatalf("warns: %v", warns)
	}
	if b.switchCalls != 3 {
		t.Fatalf("switchCalls=%d want 3", b.switchCalls)
	}
	if len(b.Switched) != 1 || b.Switched[0] != 2 {
		t.Fatalf("Switched=%v", b.Switched)
	}
}

func TestEnsureSpacePlacement_SoftWarnAfterRetries(t *testing.T) {
	stubCurrentSpace(t, 0)
	mock := &space.MockBackend{
		Desktops:   []int{1, 2},
		FailSwitch: fmt.Errorf("FAIL: desktop not found: Desktop 2"),
	}
	SetSpaceBackendForTest(mock)
	t.Cleanup(func() { SetSpaceBackendForTest(nil) })
	SetSpaceSwitchSettleForTest(0)
	t.Cleanup(func() { SetSpaceSwitchSettleForTest(-1) })

	warns, err := ensureSpacePlacement(1)
	if err != nil {
		t.Fatalf("expected soft continue, got err: %v", err)
	}
	if len(warns) != 1 {
		t.Fatalf("warns=%v", warns)
	}
	w := warns[0]
	if !strings.Contains(w, "could not switch to Desktop 2 after 3 attempts") {
		t.Fatalf("warn: %s", w)
	}
	if !strings.Contains(w, "using current Desktop") {
		t.Fatalf("warn: %s", w)
	}
	if !strings.Contains(w, "desktop not found") {
		t.Fatalf("warn: %s", w)
	}
}

func TestEnsureSpacePlacement_Space0SoftWarn(t *testing.T) {
	stubCurrentSpace(t, 1) // not on space 0
	mock := &space.MockBackend{
		Desktops:   []int{1},
		FailSwitch: fmt.Errorf("FAIL: desktop not found: Desktop 1"),
	}
	SetSpaceBackendForTest(mock)
	t.Cleanup(func() { SetSpaceBackendForTest(nil) })
	SetSpaceSwitchSettleForTest(0)
	t.Cleanup(func() { SetSpaceSwitchSettleForTest(-1) })

	warns, err := ensureSpacePlacement(0)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if len(warns) != 1 || !strings.Contains(warns[0], "Desktop 1") {
		t.Fatalf("warns=%v", warns)
	}
}

func TestEnsureSpacePlacement_HighestStillHardFails(t *testing.T) {
	stubCurrentSpace(t, 0)
	b := &failHighestBackend{err: errors.New("no Desktop buttons found")}
	SetSpaceBackendForTest(b)
	t.Cleanup(func() { SetSpaceBackendForTest(nil) })

	_, err := ensureSpacePlacement(1)
	if err == nil || !strings.Contains(err.Error(), "highest Desktop") {
		t.Fatalf("want highest hard fail, got %v", err)
	}
}

func TestIsTransientSpacePlacementError(t *testing.T) {
	cases := []struct {
		err  error
		want bool
	}{
		{nil, false},
		{fmt.Errorf("FAIL: desktop not found: Desktop 2"), true},
		{fmt.Errorf("Invalid index"), true},
		{fmt.Errorf("something -1719 here"), true},
		{fmt.Errorf("create Desktop: permission denied"), false},
	}
	for _, tc := range cases {
		if got := isTransientSpacePlacementError(tc.err); got != tc.want {
			t.Fatalf("%v: got %v want %v", tc.err, got, tc.want)
		}
	}
}

// flakySwitchBackend fails Switch failTimes times, then delegates to MockBackend.
type flakySwitchBackend struct {
	space.MockBackend
	failTimes   int
	failErr     error
	switchCalls int
}

func (f *flakySwitchBackend) Switch(n int) error {
	f.switchCalls++
	if f.switchCalls <= f.failTimes {
		return f.failErr
	}
	return f.MockBackend.Switch(n)
}

type failHighestBackend struct {
	err error
}

func (f *failHighestBackend) Create() error                  { return nil }
func (f *failHighestBackend) Switch(n int) error             { return nil }
func (f *failHighestBackend) List() ([]space.Desktop, error) { return nil, nil }
func (f *failHighestBackend) Highest() (int, error)          { return 0, f.err }
