package create

import "testing"

func TestPackageNameFromModuleBasename(t *testing.T) {
	cases := []struct {
		module string
		want   string
	}{
		{"github.com/xhd2015/interview-system", "interview_system"},
		{"github.com/xhd2015/my-app", "my_app"},
		{"cooltool", "cooltool"},
		{"github.com/xhd2015/cooltool", "cooltool"},
	}
	for _, tc := range cases {
		if got := packageNameFromModuleBasename(tc.module); got != tc.want {
			t.Errorf("packageNameFromModuleBasename(%q)=%q want %q", tc.module, got, tc.want)
		}
	}
}

func TestStandardPlaceholders_PackageName(t *testing.T) {
	m := standardPlaceholders("interview-system", "github.com/xhd2015/interview-system")
	if m["PACKAGE_NAME"] != "interview_system" {
		t.Fatalf("PACKAGE_NAME=%q", m["PACKAGE_NAME"])
	}
	if m["PROJECT_NAME"] != "interview-system" {
		t.Fatalf("PROJECT_NAME=%q", m["PROJECT_NAME"])
	}
}
