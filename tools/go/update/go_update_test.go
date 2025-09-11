package update

import "testing"

func TestStripSubDirFromTag(t *testing.T) {
	testCases := []struct {
		name       string
		tag        string
		modulePath string
		expected   string
	}{
		{
			name:       "sub-directory tag with matching module path",
			tag:        "sub/module-a/v1.20.1",
			modulePath: "github.com/example/repo/sub/module-a",
			expected:   "v1.20.1",
		},
		{
			name:       "sub-directory tag with matching module path (exact match)",
			tag:        "module-a/v2.0.0",
			modulePath: "github.com/example/repo/module-a",
			expected:   "v2.0.0",
		},
		{
			name:       "regular tag without sub-directory",
			tag:        "v1.20.1",
			modulePath: "github.com/example/repo",
			expected:   "v1.20.1",
		},
		{
			name:       "sub-directory tag with non-matching module path",
			tag:        "sub/module-a/v1.20.1",
			modulePath: "github.com/example/repo/other/module",
			expected:   "sub/module-a/v1.20.1",
		},
		{
			name:       "empty tag",
			tag:        "",
			modulePath: "github.com/example/repo",
			expected:   "",
		},
		{
			name:       "deeper sub-directory",
			tag:        "path/to/module/v2.0.0",
			modulePath: "github.com/example/repo/path/to/module",
			expected:   "v2.0.0",
		},
		{
			name:       "tag with single part (no slash)",
			tag:        "v1.0.0",
			modulePath: "github.com/example/repo/some/module",
			expected:   "v1.0.0",
		},
		{
			name:       "partial match should not strip",
			tag:        "sub/module-a/v1.20.1",
			modulePath: "github.com/example/repo/sub/module-b",
			expected:   "sub/module-a/v1.20.1",
		},
		{
			name:       "module path without sub-directory",
			tag:        "sub/module/v1.0.0",
			modulePath: "github.com/example/repo",
			expected:   "sub/module/v1.0.0",
		},
		{
			name:       "complex nested sub-directory",
			tag:        "internal/pkg/utils/v3.1.4",
			modulePath: "github.com/company/project/internal/pkg/utils",
			expected:   "v3.1.4",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := stripSubDirFromTag(tc.tag, tc.modulePath)
			if result != tc.expected {
				t.Errorf("stripSubDirFromTag(%q, %q) = %q, expected %q",
					tc.tag, tc.modulePath, result, tc.expected)
			}
		})
	}
}
