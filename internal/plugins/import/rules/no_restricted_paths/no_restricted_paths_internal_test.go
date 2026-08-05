// These tests cover the two path helpers directly, including the shapes no rule
// test can reach: the embedded fixture tree leaves out dot-prefixed names, and
// the case-insensitive comparison only runs on Windows.
package no_restricted_paths

import "testing"

// TestContainsPath pins the restriction predicate, which reads any leftover
// starting with two dots as escaping the target.
func TestContainsPath(t *testing.T) {
	tests := []struct {
		name          string
		filePath      string
		target        string
		caseSensitive bool
		want          bool
	}{
		{
			name:          "the target itself",
			filePath:      "/project/server",
			target:        "/project/server",
			caseSensitive: true,
			want:          true,
		},
		{
			name:          "a descendant of the target",
			filePath:      "/project/server/one/a.ts",
			target:        "/project/server",
			caseSensitive: true,
			want:          true,
		},
		{
			name:          "a sibling sharing a name prefix",
			filePath:      "/project/serverless/a.ts",
			target:        "/project/server",
			caseSensitive: true,
			want:          false,
		},
		{
			// `path.relative('/project/server', '/project/server/..secret.ts')`
			// is `'..secret.ts'`, which upstream reads as escaping the target.
			name:          "a child whose name starts with two dots",
			filePath:      "/project/server/..secret.ts",
			target:        "/project/server",
			caseSensitive: true,
			want:          false,
		},
		{
			name:          "a two-dot directory named by the target itself",
			filePath:      "/project/server/..secret/a.ts",
			target:        "/project/server/..secret",
			caseSensitive: true,
			want:          true,
		},
		{
			name:          "differing case on a case-sensitive platform",
			filePath:      "c:/PROJECT/client/a.ts",
			target:        "C:/Project/client",
			caseSensitive: true,
			want:          false,
		},
		{
			name:          "differing case on a case-insensitive platform",
			filePath:      "c:/PROJECT/client/a.ts",
			target:        "C:/Project/client",
			caseSensitive: false,
			want:          true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := containsPath(test.filePath, test.target, test.caseSensitive)
			if got != test.want {
				t.Errorf("containsPath(%q, %q, %v) = %v, want %v", test.filePath, test.target, test.caseSensitive, got, test.want)
			}
		})
	}
}

// TestIsPathWithin pins the exception-validity predicate, which accepts a
// leftover starting with two dots because upstream's importType only classifies
// a whole `..` component as "parent".
func TestIsPathWithin(t *testing.T) {
	tests := []struct {
		name          string
		filePath      string
		target        string
		caseSensitive bool
		want          bool
	}{
		{
			name:          "the target itself",
			filePath:      "/project/server",
			target:        "/project/server",
			caseSensitive: true,
			want:          true,
		},
		{
			name:          "a descendant of the target",
			filePath:      "/project/server/one/a.ts",
			target:        "/project/server",
			caseSensitive: true,
			want:          true,
		},
		{
			// `path.relative('/project/server', '/project/server/..secret')`
			// is `'..secret'`, which importType classifies as "external", not
			// "parent", so the exception stays valid.
			name:          "a child whose name starts with two dots",
			filePath:      "/project/server/..secret",
			target:        "/project/server",
			caseSensitive: true,
			want:          true,
		},
		{
			name:          "a path outside the target",
			filePath:      "/project/client/a.ts",
			target:        "/project/server",
			caseSensitive: true,
			want:          false,
		},
		{
			name:          "a sibling sharing a name prefix",
			filePath:      "/project/serverless/a.ts",
			target:        "/project/server",
			caseSensitive: true,
			want:          false,
		},
		{
			name:          "differing case on a case-insensitive platform",
			filePath:      "c:/PROJECT/server/..secret",
			target:        "C:/Project/server",
			caseSensitive: false,
			want:          true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := isPathWithin(test.filePath, test.target, test.caseSensitive)
			if got != test.want {
				t.Errorf("isPathWithin(%q, %q, %v) = %v, want %v", test.filePath, test.target, test.caseSensitive, got, test.want)
			}
		})
	}
}
