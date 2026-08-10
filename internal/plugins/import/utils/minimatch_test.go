package utils_test

import (
	"testing"

	import_utils "github.com/web-infra-dev/rslint/internal/plugins/import/utils"
)

func TestMatchMinimatch(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		value   string
		pattern string
		options import_utils.MinimatchOptions
		want    bool
	}{
		{name: "globstar", value: "@app/a/b", pattern: "@app/**", options: import_utils.MinimatchOptions{NoComment: true}, want: true},
		{name: "globstar zero directory", value: "Internal/lib", pattern: "Internal/**/*", options: import_utils.MinimatchOptions{NoComment: true}, want: true},
		{name: "globstar nested directory", value: "Internal/a/lib", pattern: "Internal/**/*", options: import_utils.MinimatchOptions{NoComment: true}, want: true},
		{name: "globstar leading zero directory", value: "b", pattern: "**/b", options: import_utils.MinimatchOptions{NoComment: true}, want: true},
		{name: "trailing globstar requires slash", value: "foo", pattern: "foo/**", options: import_utils.MinimatchOptions{NoComment: true}, want: false},
		{name: "trailing globstar accepts slash", value: "foo/", pattern: "foo/**", options: import_utils.MinimatchOptions{NoComment: true}, want: true},
		{name: "consecutive trailing globstars require slash", value: "foo", pattern: "foo/**/**", options: import_utils.MinimatchOptions{NoComment: true}, want: false},
		{name: "consecutive trailing globstars accept slash", value: "foo/", pattern: "foo/**/**", options: import_utils.MinimatchOptions{NoComment: true}, want: true},
		{name: "only globstars accept empty path", value: "", pattern: "**/**", options: import_utils.MinimatchOptions{NoComment: true}, want: true},
		{name: "only globstars accept one segment", value: "a", pattern: "**/**", options: import_utils.MinimatchOptions{NoComment: true}, want: true},
		{name: "dot hidden by default", value: ".hidden", pattern: "*", options: import_utils.MinimatchOptions{NoComment: true}, want: false},
		{name: "dot enabled", value: ".hidden", pattern: "*", options: import_utils.MinimatchOptions{NoComment: true, Dot: true}, want: true},
		{name: "globstar hidden by default", value: "a/.hidden", pattern: "a/**", options: import_utils.MinimatchOptions{NoComment: true}, want: false},
		{name: "globstar dot enabled", value: "a/.hidden", pattern: "a/**", options: import_utils.MinimatchOptions{NoComment: true, Dot: true}, want: true},
		{name: "globstar never consumes dot segment", value: "./foo", pattern: "**", options: import_utils.MinimatchOptions{NoComment: true, Dot: true}, want: false},
		{name: "literal dot segment", value: "./foo", pattern: "./**", options: import_utils.MinimatchOptions{NoComment: true}, want: true},
		{name: "brace expansion", value: "multiverse/a/b", pattern: "multiverse{*,*/**}", options: import_utils.MinimatchOptions{NoComment: true}, want: true},
		{name: "nonegate literal", value: "!/module", pattern: "!/**", options: import_utils.MinimatchOptions{NoNegate: true, NoComment: true}, want: true},
		{name: "negation", value: "allowed", pattern: "!blocked", options: import_utils.MinimatchOptions{NoComment: true}, want: true},
		{name: "comment pattern", value: "#/module", pattern: "#/**", want: false},
		{name: "nocomment literal", value: "#/module", pattern: "#/**", options: import_utils.MinimatchOptions{NoComment: true}, want: true},
		{name: "match base extglob", value: "styles/main.css", pattern: "*.+(css|svg)", options: import_utils.MinimatchOptions{MatchBase: true}, want: true},
		{name: "match base extglob other", value: "styles/main.js", pattern: "*.+(css|svg)", options: import_utils.MinimatchOptions{MatchBase: true}, want: false},
		{name: "negative extglob", value: "pkg/allowed", pattern: "pkg/!(blocked|private)", options: import_utils.MinimatchOptions{NoComment: true}, want: true},
		{name: "negative extglob excluded", value: "pkg/private", pattern: "pkg/!(blocked|private)", options: import_utils.MinimatchOptions{NoComment: true}, want: false},
		{name: "nocase", value: "React", pattern: "react", options: import_utils.MinimatchOptions{NoCase: true}, want: true},
		{name: "noglobstar", value: "a/b/c", pattern: "a/**", options: import_utils.MinimatchOptions{NoGlobStar: true}, want: false},
		{name: "noext", value: "css", pattern: "+(css|svg)", options: import_utils.MinimatchOptions{NoExt: true}, want: false},
		{name: "nobrace", value: "a", pattern: "{a,b}", options: import_utils.MinimatchOptions{NoBrace: true}, want: false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			if got := import_utils.MatchMinimatch(test.value, test.pattern, test.options); got != test.want {
				t.Fatalf("MatchMinimatch(%q, %q) = %v, want %v", test.value, test.pattern, got, test.want)
			}
		})
	}
}
