// These tests cover the path and glob helpers directly, including the shapes no
// rule test can reach: the win32 flavor of the path comparison only runs on
// Windows.
package no_restricted_paths

import "testing"

// TestContainsPath pins the restriction predicate, which reads any leftover
// starting with two dots as escaping the target.
func TestContainsPath(t *testing.T) {
	tests := []struct {
		name     string
		filePath string
		target   string
		windows  bool
		want     bool
	}{
		{
			name:     "the target itself",
			filePath: "/project/server",
			target:   "/project/server",
			want:     true,
		},
		{
			name:     "a descendant of the target",
			filePath: "/project/server/one/a.ts",
			target:   "/project/server",
			want:     true,
		},
		{
			name:     "a sibling sharing a name prefix",
			filePath: "/project/serverless/a.ts",
			target:   "/project/server",
			want:     false,
		},
		{
			// `path.relative('/project/server', '/project/server/..secret.ts')`
			// is `'..secret.ts'`, which upstream reads as escaping the target.
			name:     "a child whose name starts with two dots",
			filePath: "/project/server/..secret.ts",
			target:   "/project/server",
			want:     false,
		},
		{
			name:     "a two-dot directory named by the target itself",
			filePath: "/project/server/..secret/a.ts",
			target:   "/project/server/..secret",
			want:     true,
		},
		{
			name:     "differing case on a case-sensitive platform",
			filePath: "c:/PROJECT/client/a.ts",
			target:   "C:/Project/client",
			want:     false,
		},
		{
			name:     "differing case on a case-insensitive platform",
			filePath: "c:/PROJECT/client/a.ts",
			target:   "C:/Project/client",
			windows:  true,
			want:     true,
		},
		{
			// `path.win32.relative('C:\\project\\server', 'D:\\vendor\\a.ts')` is
			// `'D:\\vendor\\a.ts'`, which upstream reads as inside the target.
			name:     "a file on another drive",
			filePath: "D:/vendor/a.ts",
			target:   "C:/project/server",
			windows:  true,
			want:     true,
		},
		{
			name:     "a drive path against a UNC target",
			filePath: "C:/other/a.ts",
			target:   "//server/share/project",
			windows:  true,
			want:     true,
		},
		{
			name:     "a file on another UNC server",
			filePath: "//other/share/a.ts",
			target:   "//server/share/project",
			windows:  true,
			want:     true,
		},
		{
			// `path.win32.relative` still reaches the shared server segment here
			// and answers `'..\\..\\two\\a.ts'`, which escapes the target.
			name:     "another share on the same UNC server",
			filePath: "//server/two/a.ts",
			target:   "//server/one/project",
			windows:  true,
			want:     false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := containsPath(test.filePath, test.target, test.windows)
			if got != test.want {
				t.Errorf("containsPath(%q, %q, %v) = %v, want %v", test.filePath, test.target, test.windows, got, test.want)
			}
		})
	}
}

// TestIsPathWithin pins the exception-validity predicate, which accepts a
// leftover starting with two dots because upstream's importType only classifies
// a whole `..` component as "parent".
func TestIsPathWithin(t *testing.T) {
	tests := []struct {
		name     string
		filePath string
		target   string
		windows  bool
		want     bool
	}{
		{
			name:     "the target itself",
			filePath: "/project/server",
			target:   "/project/server",
			want:     true,
		},
		{
			name:     "a descendant of the target",
			filePath: "/project/server/one/a.ts",
			target:   "/project/server",
			want:     true,
		},
		{
			// `path.relative('/project/server', '/project/server/..secret')`
			// is `'..secret'`, which importType classifies as "external", not
			// "parent", so the exception stays valid.
			name:     "a child whose name starts with two dots",
			filePath: "/project/server/..secret",
			target:   "/project/server",
			want:     true,
		},
		{
			name:     "a path outside the target",
			filePath: "/project/client/a.ts",
			target:   "/project/server",
			want:     false,
		},
		{
			name:     "a sibling sharing a name prefix",
			filePath: "/project/serverless/a.ts",
			target:   "/project/server",
			want:     false,
		},
		{
			name:     "differing case on a case-insensitive platform",
			filePath: "c:/PROJECT/server/..secret",
			target:   "C:/Project/server",
			windows:  true,
			want:     true,
		},
		{
			// importType classifies the `'D:\\vendor\\allowed'` that
			// `path.win32.relative` hands back as "external", so the exception
			// stays valid rather than being reported as escaping `from`.
			name:     "an exception on another drive",
			filePath: "D:/vendor/allowed",
			target:   "C:/project/server",
			windows:  true,
			want:     true,
		},
		{
			name:     "an exception on another share of the same UNC server",
			filePath: "//server/two/allowed",
			target:   "//server/one/project",
			windows:  true,
			want:     false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := isPathWithin(test.filePath, test.target, test.windows)
			if got != test.want {
				t.Errorf("isPathWithin(%q, %q, %v) = %v, want %v", test.filePath, test.target, test.windows, got, test.want)
			}
		})
	}
}

// TestEscapeExtglob pins the rewrite that keeps the extended glob syntax
// doublestar does not implement out of the base wildcard syntax.
func TestEscapeExtglob(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
		want    string
	}{
		{
			name:    "a pattern without extended glob syntax",
			pattern: "/src/**/*.ts",
			want:    "/src/**/*.ts",
		},
		{
			name:    "a zero-or-one list, whose `?` would match any character",
			pattern: "/src/?(server)/*",
			want:    `/src/\?(server)/*`,
		},
		{
			name:    "a zero-or-more list, whose `*` would match any segment",
			pattern: "/src/*(server)/*",
			want:    `/src/\*(server)/*`,
		},
		{
			name:    "an alternation list, which carries no base wildcard",
			pattern: "/src/@(server|shared)/**/*",
			want:    "/src/@(server|shared)/**/*",
		},
		{
			name:    "wildcards nested inside a list body",
			pattern: "/src/+(a*|b?)/x",
			want:    `/src/+(a\*|b\?)/x`,
		},
		{
			name:    "a nested list",
			pattern: "/src/!(a|?(b))/x",
			want:    `/src/!(a|\?(b))/x`,
		},
		{
			name:    "an escaped character before a list",
			pattern: `/src/\@x/?(a)`,
			want:    `/src/\@x/\?(a)`,
		},
		{
			name:    "an unterminated list, which is not extended glob syntax",
			pattern: "/src/?(server/*",
			want:    "/src/?(server/*",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := escapeExtglob(test.pattern)
			if got != test.want {
				t.Errorf("escapeExtglob(%q) = %q, want %q", test.pattern, got, test.want)
			}
		})
	}
}
