package space

import (
	"sync"

	lib "github.com/xhd2015/dot-pkgs/go-pkgs/computer-use/macos/space"
)

// Test hooks: inject mock Backend or settle; production uses real library.

var (
	hookMu      sync.Mutex
	testBackend lib.Backend
	testRunner  Runner
	testSettle  *int // milliseconds; nil = default 500; 0 or negative = no sleep via cfg
)

// SetBackendForTest installs a Backend used instead of live Mission Control.
func SetBackendForTest(b lib.Backend) {
	hookMu.Lock()
	defer hookMu.Unlock()
	testBackend = b
}

// SetRunnerForTest installs a Runner for --run follow-ups.
func SetRunnerForTest(r Runner) {
	hookMu.Lock()
	defer hookMu.Unlock()
	testRunner = r
}

// SetSettleMSForTest sets settle delay in ms (0 = no sleep).
func SetSettleMSForTest(ms int) {
	hookMu.Lock()
	defer hookMu.Unlock()
	testSettle = &ms
}

// SetGOOSForTest overrides library platform check.
func SetGOOSForTest(goos string) {
	lib.SetGOOSForTest(goos)
}

// ResetTestHooks clears all test overrides.
func ResetTestHooks() {
	hookMu.Lock()
	defer hookMu.Unlock()
	testBackend = nil
	testRunner = nil
	testSettle = nil
	lib.SetGOOSForTest("")
}

func settleMS() int {
	hookMu.Lock()
	defer hookMu.Unlock()
	if testSettle != nil {
		return *testSettle
	}
	return 500
}

func getRunner() Runner {
	hookMu.Lock()
	r := testRunner
	hookMu.Unlock()
	if r != nil {
		return r
	}
	return &execRunner{}
}

func doCreate(cfg *lib.Config) error {
	if b := backend(); b != nil {
		return b.Create()
	}
	return lib.Create(cfg)
}

func doSwitch(n int, cfg *lib.Config) error {
	if b := backend(); b != nil {
		if err := b.Switch(n); err != nil {
			return err
		}
		// mock path: apply settle from cfg
		return nil
	}
	return lib.Switch(n, cfg)
}

func doList(cfg *lib.Config) ([]lib.Desktop, error) {
	if b := backend(); b != nil {
		return b.List()
	}
	return lib.List(cfg)
}

func doHighest(cfg *lib.Config) (int, error) {
	if b := backend(); b != nil {
		return b.Highest()
	}
	return lib.Highest(cfg)
}

func createAndActivate(cfg *lib.Config) (int, error) {
	if b := backend(); b != nil {
		if err := b.Create(); err != nil {
			return 0, err
		}
		n, err := b.Highest()
		if err != nil {
			return 0, err
		}
		if err := b.Switch(n); err != nil {
			return 0, err
		}
		return n, nil
	}
	return lib.CreateAndActivate(cfg)
}

func backend() lib.Backend {
	hookMu.Lock()
	defer hookMu.Unlock()
	return testBackend
}

// testOsascript is unused when Backend mock is set; kept for Config inject path.
func testOsascript() func(script string, args ...string) (string, error) {
	return nil
}
