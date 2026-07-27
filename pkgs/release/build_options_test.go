package release

import "testing"

func TestApplyOptions_DefaultPackagePath(t *testing.T) {
	cfg := applyOptions(nil)
	if cfg.packagePath != "./" {
		t.Fatalf("default packagePath: got %q, want %q", cfg.packagePath, "./")
	}
}

func TestApplyOptions_WithPackagePath(t *testing.T) {
	cfg := applyOptions([]Option{WithPackagePath("./cmd/doctest")})
	if cfg.packagePath != "./cmd/doctest" {
		t.Fatalf("packagePath: got %q, want %q", cfg.packagePath, "./cmd/doctest")
	}
}

func TestApplyOptions_EmptyPackagePathIgnored(t *testing.T) {
	cfg := applyOptions([]Option{WithPackagePath("")})
	if cfg.packagePath != "./" {
		t.Fatalf("empty WithPackagePath should keep default: got %q, want %q", cfg.packagePath, "./")
	}
}

func TestApplyOptions_NilOptionSkipped(t *testing.T) {
	cfg := applyOptions([]Option{nil, WithPackagePath("./cmd/foo"), nil})
	if cfg.packagePath != "./cmd/foo" {
		t.Fatalf("packagePath: got %q, want %q", cfg.packagePath, "./cmd/foo")
	}
}
