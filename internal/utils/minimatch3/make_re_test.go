package minimatch3_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/utils/minimatch3"
)

// Expectations from minimatch 3.1.5 makeRe().test(), including differences
// from Match. The options that only affect Match do not affect MakeRe.
func TestMakeRe(t *testing.T) {
	tests := []struct {
		pattern, path string
		options       minimatch3.Options
		want          bool
	}{
		{"app/**/**", "app/a", minimatch3.Options{}, false},
		{"app/**/**", "app/a/b", minimatch3.Options{}, true},
		{"**/api/*", "api/service", minimatch3.Options{}, false},
		{"**/api/*", "/api/service", minimatch3.Options{}, true},
		{"a/**/b", "a/b", minimatch3.Options{}, false},
		{"a/**/b", "a/x/b", minimatch3.Options{}, true},
		{"a/**/b", "a//b", minimatch3.Options{}, true},
		{"a/*", "a//b", minimatch3.Options{}, false},
		{"a/*", "a/b/", minimatch3.Options{}, false},
		{"**", "a/.hidden/b", minimatch3.Options{}, false},
		{"**", ".hidden/b", minimatch3.Options{}, false},
		{"**", "a/b", minimatch3.Options{}, true},
		{"**", "a\nb", minimatch3.Options{}, false},
		{"*", "a\nb", minimatch3.Options{}, true},
		{"a/{b,c}", "a/b", minimatch3.Options{}, true},
		{"a/{b,c}", "a/c", minimatch3.Options{}, true},
		{"a/{b,c}", "a/d", minimatch3.Options{}, false},
		{"a/@(b|c)", "a/b", minimatch3.Options{}, true},
		{"a/!(b)", "a/c", minimatch3.Options{}, true},
		{"a/!(b)", "a/b", minimatch3.Options{}, false},
		{"!a/b", "a/c", minimatch3.Options{}, true},
		{"!a/b", "a/b", minimatch3.Options{}, false},
		{"!a/b", "a/c\n", minimatch3.Options{}, false},
		{"a/??", "a/😀", minimatch3.Options{}, true},
		{"a/?", "a/😀", minimatch3.Options{}, false},
		{"a/😀", "a/😀", minimatch3.Options{}, true},
		{"a/[😀]", "a/😀", minimatch3.Options{}, false},
		{"a/\\*", "a/*", minimatch3.Options{}, true},
		{"a/[", "a/[", minimatch3.Options{}, true},
		{"#a", "a", minimatch3.Options{}, false},
		{"", "", minimatch3.Options{}, false},
		{"a/**", "a/x\r", minimatch3.Options{}, false},
		{"a/*", "a/x\r", minimatch3.Options{}, true},
		{"a/**/b", "a/.hidden/b", minimatch3.Options{}, true},
		{"a/**/b", "a/.hidden/b", minimatch3.Options{Dot: true}, true},
		{"a/**/b", "a/../b", minimatch3.Options{Dot: true}, true},
		{"a/**", "a/x/y", minimatch3.Options{NoGlobStar: true}, false},
		{"a/**", "a/x", minimatch3.Options{NoGlobStar: true}, true},
		{"a/b", "A/B", minimatch3.Options{NoCase: true}, true},
		{"k", "K", minimatch3.Options{NoCase: true}, false},
		{"a/{b,c}", "a/{b,c}", minimatch3.Options{NoBrace: true}, true},
		{"!a/b", "a/c", minimatch3.Options{NoNegate: true}, false},
		{"#a", "#a", minimatch3.Options{NoComment: true}, true},
		{"**/x", "x", minimatch3.Options{Partial: true}, false},
		{"x", "a/x", minimatch3.Options{MatchBase: true}, false},
		{"!x", "x", minimatch3.Options{FlipNegate: true}, false},
		{"a/?(b)", "a/(b)", minimatch3.Options{NoExt: true}, false},
	}
	for _, tt := range tests {
		t.Run(tt.pattern+"/"+tt.path, func(t *testing.T) {
			if got := minimatch3.New(tt.pattern, tt.options).MakeRe().Test(tt.path); got != tt.want {
				t.Errorf("MakeRe(%q).Test(%q) = %v, want %v", tt.pattern, tt.path, got, tt.want)
			}
		})
	}
}
