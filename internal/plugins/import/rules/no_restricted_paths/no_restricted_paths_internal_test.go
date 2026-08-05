// TestContainsPath covers the path containment helper directly, including the
// two shapes a plain component-prefix comparison gets wrong: a child name that
// starts with two dots, and a host that resolves names case-insensitively.
// Neither is reachable from a rule test — the embedded fixture tree leaves out
// dot-prefixed names, and the fixture host is always case-sensitive.
package no_restricted_paths

import "testing"

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
			name:          "differing case on a case-sensitive host",
			filePath:      "c:/PROJECT/client/a.ts",
			target:        "C:/Project/client",
			caseSensitive: true,
			want:          false,
		},
		{
			name:          "differing case on a case-insensitive host",
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
