package utils_test

import (
	"math"
	"strconv"
	"testing"

	"github.com/microsoft/typescript-go/shim/stringutil"
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
		{name: "magic segment does not match empty", value: "", pattern: "*", want: false},
		{name: "empty literal matches empty", value: "", pattern: "", want: true},
		{name: "empty literal does not match root", value: "/", pattern: "", want: false},
		{name: "pattern is ECMAScript-trimmed", value: "pkg", pattern: "\uFEFF pkg \uFEFF", want: true},
		{name: "trailing slash may remain after pattern", value: "foo/", pattern: "*", want: true},
		{name: "adjacent slashes collapse", value: "foo//bar", pattern: "foo/bar", want: true},
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
		{name: "dot character class is not an explicit dot prefix", value: ".hidden", pattern: "[.]hidden", want: false},
		{name: "globstar never consumes dot segment", value: "./foo", pattern: "**", options: import_utils.MinimatchOptions{NoComment: true, Dot: true}, want: false},
		{name: "literal dot segment", value: "./foo", pattern: "./**", options: import_utils.MinimatchOptions{NoComment: true}, want: true},
		{name: "brace expansion", value: "multiverse/a/b", pattern: "multiverse{*,*/**}", options: import_utils.MinimatchOptions{NoComment: true}, want: true},
		{name: "numeric brace range", value: "pkg/2", pattern: "pkg/{1..3}", want: true},
		{name: "descending brace range", value: "pkg/2", pattern: "pkg/{3..1}", want: true},
		{name: "padded brace range", value: "pkg/02", pattern: "pkg/{01..03}", want: true},
		{name: "stepped letter brace range", value: "pkg/c", pattern: "pkg/{a..e..2}", want: true},
		{name: "brace range step excludes value", value: "pkg/d", pattern: "pkg/{a..e..2}", want: false},
		{name: "minimum integer brace step cannot overflow", value: "pkg/1", pattern: "pkg/{1..3.." + strconv.Itoa(math.MinInt) + "}", want: true},
		{name: "escaped brace remains literal", value: "{a,b}", pattern: `\{a,b\}`, options: import_utils.MinimatchOptions{AllowWindowsEscape: true}, want: true},
		{name: "nonegate literal", value: "!/module", pattern: "!/**", options: import_utils.MinimatchOptions{NoNegate: true, NoComment: true}, want: true},
		{name: "negation", value: "allowed", pattern: "!blocked", options: import_utils.MinimatchOptions{NoComment: true}, want: true},
		{name: "comment pattern", value: "#/module", pattern: "#/**", want: false},
		{name: "nocomment literal", value: "#/module", pattern: "#/**", options: import_utils.MinimatchOptions{NoComment: true}, want: true},
		{name: "match base extglob", value: "styles/main.css", pattern: "*.+(css|svg)", options: import_utils.MinimatchOptions{MatchBase: true}, want: true},
		{name: "match base extglob other", value: "styles/main.js", pattern: "*.+(css|svg)", options: import_utils.MinimatchOptions{MatchBase: true}, want: false},
		{name: "match base ignores trailing slash", value: "foo/bar/", pattern: "bar", options: import_utils.MinimatchOptions{MatchBase: true}, want: true},
		{name: "negative extglob", value: "pkg/allowed", pattern: "pkg/!(blocked|private)", options: import_utils.MinimatchOptions{NoComment: true}, want: true},
		{name: "negative extglob excluded", value: "pkg/private", pattern: "pkg/!(blocked|private)", options: import_utils.MinimatchOptions{NoComment: true}, want: false},
		{name: "negative extglob with suffix", value: "pkg/foo.js", pattern: "pkg/!(*.test).js", want: true},
		{name: "negative extglob suffix excluded", value: "pkg/foo.test.js", pattern: "pkg/!(*.test).js", want: false},
		{name: "nested extglob", value: "pkg/ab", pattern: "pkg/@(a|a@(b|c))", want: true},
		{name: "nocase", value: "React", pattern: "react", options: import_utils.MinimatchOptions{NoCase: true}, want: true},
		{name: "nocase Greek simple fold", value: "ς", pattern: "σ", options: import_utils.MinimatchOptions{NoCase: true}, want: true},
		{name: "nocase does not fold Kelvin sign to ASCII", value: "K", pattern: "k", options: import_utils.MinimatchOptions{NoCase: true}, want: false},
		{name: "nocase does not fold long s to ASCII", value: "ſ", pattern: "s", options: import_utils.MinimatchOptions{NoCase: true}, want: false},
		{name: "nocase does not apply multi-character sharp-s mapping", value: "ẞ", pattern: "ß", options: import_utils.MinimatchOptions{NoCase: true}, want: false},
		{name: "nocase character class uses JS folding", value: "ς", pattern: "[σ]", options: import_utils.MinimatchOptions{NoCase: true}, want: true},
		{name: "nocase character class preserves original broad range", value: "_", pattern: "[A-z]", options: import_utils.MinimatchOptions{NoCase: true}, want: true},
		{name: "question matches one UTF-16 unit", value: "😀", pattern: "?", want: false},
		{name: "two questions match an astral pair", value: "😀", pattern: "??", want: true},
		{name: "astral character class contains separate surrogate units", value: "😀", pattern: "[😀]", want: false},
		{name: "astral literal matches its surrogate pair", value: "😀", pattern: "😀", want: true},
		{name: "question matches one lone surrogate", value: stringutil.EncodeJSStringRune(0xD800), pattern: "?", want: true},
		{name: "noglobstar", value: "a/b/c", pattern: "a/**", options: import_utils.MinimatchOptions{NoGlobStar: true}, want: false},
		{name: "noglobstar does not consume empty trailing segment", value: "foo/", pattern: "foo/**", options: import_utils.MinimatchOptions{NoGlobStar: true}, want: false},
		{name: "noext", value: "css", pattern: "+(css|svg)", options: import_utils.MinimatchOptions{NoExt: true}, want: false},
		{name: "nobrace", value: "a", pattern: "{a,b}", options: import_utils.MinimatchOptions{NoBrace: true}, want: false},
		{name: "partial", value: "a", pattern: "a/b", options: import_utils.MinimatchOptions{Partial: true}, want: true},
		{name: "partial filesystem root", value: "/", pattern: "a/b", options: import_utils.MinimatchOptions{Partial: true}, want: true},
		{name: "partial still rejects mismatched prefix", value: "x", pattern: "a/b", options: import_utils.MinimatchOptions{Partial: true}, want: false},
		{name: "flip negate matching value", value: "a", pattern: "!a", options: import_utils.MinimatchOptions{FlipNegate: true}, want: true},
		{name: "flip negate nonmatching value", value: "b", pattern: "!a", options: import_utils.MinimatchOptions{FlipNegate: true}, want: false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			if got := import_utils.MatchMinimatch(test.value, test.pattern, test.options); got != test.want {
				t.Fatalf("MatchMinimatch(%q, %q) = %v, want %v", test.value, test.pattern, got, test.want)
			}
			if got := import_utils.CompileMinimatch(test.pattern, test.options).Match(test.value); got != test.want {
				t.Fatalf("CompileMinimatch(%q).Match(%q) = %v, want %v", test.pattern, test.value, got, test.want)
			}
		})
	}
}
