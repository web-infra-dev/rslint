// The table below is minimatch's own conformance corpus — the bash glob-test
// cases in test/patterns.js of minimatch v3.1.2 — expanded to one expectation
// per file. Every `matches` and `misses` entry is what the npm package answers
// on minimatch 3.1.5, so a divergence here is a divergence from the matcher
// ESLint itself uses.
//
// Regenerate by running the corpus through minimatch and grouping per pattern;
// do not hand-edit an expectation to make a change pass.
package minimatch_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/utils/minimatch"
)

type corpusCase struct {
	pattern string
	options minimatch.Options
	matches []string
	misses  []string
}

var corpus = []corpusCase{
	{
		pattern: `a*`,
		matches: []string{`a`, `abc`, `abd`, `abe`},
		misses:  []string{`b`, `c`, `d`, `bb`, `bcd`, `ca`, `cb`, `dd`, `de`, `bdir/`, `bdir/cfile`},
	},
	{
		pattern: `X*`,
		misses:  []string{`a`, `b`, `c`, `d`, `abc`, `abd`, `abe`, `bb`, `bcd`, `ca`, `cb`, `dd`, `de`, `bdir/`, `bdir/cfile`},
	},
	{
		pattern: `\*`,
		misses:  []string{`a`, `b`, `c`, `d`, `abc`, `abd`, `abe`, `bb`, `bcd`, `ca`, `cb`, `dd`, `de`, `bdir/`, `bdir/cfile`},
	},
	{
		pattern: `\**`,
		misses:  []string{`a`, `b`, `c`, `d`, `abc`, `abd`, `abe`, `bb`, `bcd`, `ca`, `cb`, `dd`, `de`, `bdir/`, `bdir/cfile`},
	},
	{
		pattern: `\*\*`,
		misses:  []string{`a`, `b`, `c`, `d`, `abc`, `abd`, `abe`, `bb`, `bcd`, `ca`, `cb`, `dd`, `de`, `bdir/`, `bdir/cfile`},
	},
	{
		pattern: `b*/`,
		matches: []string{`bdir/`},
		misses:  []string{`a`, `b`, `c`, `d`, `abc`, `abd`, `abe`, `bb`, `bcd`, `ca`, `cb`, `dd`, `de`, `bdir/cfile`},
	},
	{
		pattern: `c*`,
		matches: []string{`c`, `ca`, `cb`},
		misses:  []string{`a`, `b`, `d`, `abc`, `abd`, `abe`, `bb`, `bcd`, `dd`, `de`, `bdir/`, `bdir/cfile`},
	},
	{
		pattern: `**`,
		matches: []string{`a`, `b`, `c`, `d`, `abc`, `abd`, `abe`, `bb`, `bcd`, `ca`, `cb`, `dd`, `de`, `bdir/`, `bdir/cfile`},
	},
	{
		pattern: `\.\./*/`,
		misses:  []string{`a`, `b`, `c`, `d`, `abc`, `abd`, `abe`, `bb`, `bcd`, `ca`, `cb`, `dd`, `de`, `bdir/`, `bdir/cfile`},
	},
	{
		pattern: `s/\..*//`,
		misses:  []string{`a`, `b`, `c`, `d`, `abc`, `abd`, `abe`, `bb`, `bcd`, `ca`, `cb`, `dd`, `de`, `bdir/`, `bdir/cfile`},
	},
	{
		pattern: `/^root:/{s/^[^:]*:[^:]*:([^:]*).*$/\1/`,
		misses:  []string{`a`, `b`, `c`, `d`, `abc`, `abd`, `abe`, `bb`, `bcd`, `ca`, `cb`, `dd`, `de`, `bdir/`, `bdir/cfile`},
	},
	{
		pattern: "/^root:/{s/^[^:]*:[^:]*:([^:]*).*$/\u0001/",
		misses:  []string{`a`, `b`, `c`, `d`, `abc`, `abd`, `abe`, `bb`, `bcd`, `ca`, `cb`, `dd`, `de`, `bdir/`, `bdir/cfile`},
	},
	{
		pattern: `[a-c]b*`,
		matches: []string{`abc`, `abd`, `abe`, `bb`, `cb`},
		misses:  []string{`a`, `b`, `c`, `d`, `bcd`, `ca`, `dd`, `de`, `bdir/`, `bdir/cfile`},
	},
	{
		pattern: `[a-y]*[^c]`,
		matches: []string{`abd`, `abe`, `bb`, `bcd`, `ca`, `cb`, `dd`, `de`, `bdir/`},
		misses:  []string{`a`, `b`, `c`, `d`, `abc`, `bdir/cfile`},
	},
	{
		pattern: `a*[^c]`,
		matches: []string{`abd`, `abe`},
		misses:  []string{`a`, `b`, `c`, `d`, `abc`, `bb`, `bcd`, `ca`, `cb`, `dd`, `de`, `bdir/`, `bdir/cfile`},
	},
	{
		pattern: `a[X-]b`,
		matches: []string{`a-b`, `aXb`},
		misses:  []string{`a`, `b`, `c`, `d`, `abc`, `abd`, `abe`, `bb`, `bcd`, `ca`, `cb`, `dd`, `de`, `bdir/`, `bdir/cfile`},
	},
	{
		pattern: `[^a-c]*`,
		matches: []string{`d`, `dd`, `de`},
		misses:  []string{`a`, `b`, `c`, `abc`, `abd`, `abe`, `bb`, `bcd`, `ca`, `cb`, `bdir/`, `bdir/cfile`, `a-b`, `aXb`, `.x`, `.y`},
	},
	{
		pattern: `a\*b/*`,
		matches: []string{`a*b/ooo`},
		misses:  []string{`a`, `b`, `c`, `d`, `abc`, `abd`, `abe`, `bb`, `bcd`, `ca`, `cb`, `dd`, `de`, `bdir/`, `bdir/cfile`, `a-b`, `aXb`, `.x`, `.y`, `a*b/`},
	},
	{
		pattern: `a\*?/*`,
		matches: []string{`a*b/ooo`},
		misses:  []string{`a`, `b`, `c`, `d`, `abc`, `abd`, `abe`, `bb`, `bcd`, `ca`, `cb`, `dd`, `de`, `bdir/`, `bdir/cfile`, `a-b`, `aXb`, `.x`, `.y`, `a*b/`},
	},
	{
		pattern: `*\\!*`,
		misses:  []string{`echo !7`},
	},
	{
		pattern: `*\!*`,
		matches: []string{`echo !7`},
	},
	{
		pattern: `*.\*`,
		matches: []string{`r.*`},
	},
	{
		pattern: `a[b]c`,
		matches: []string{`abc`},
		misses:  []string{`a`, `b`, `c`, `d`, `abd`, `abe`, `bb`, `bcd`, `ca`, `cb`, `dd`, `de`, `bdir/`, `bdir/cfile`, `a-b`, `aXb`, `.x`, `.y`, `a*b/`, `a*b/ooo`},
	},
	{
		pattern: `a[\b]c`,
		matches: []string{`abc`},
		misses:  []string{`a`, `b`, `c`, `d`, `abd`, `abe`, `bb`, `bcd`, `ca`, `cb`, `dd`, `de`, `bdir/`, `bdir/cfile`, `a-b`, `aXb`, `.x`, `.y`, `a*b/`, `a*b/ooo`},
	},
	{
		pattern: `a?c`,
		matches: []string{`abc`},
		misses:  []string{`a`, `b`, `c`, `d`, `abd`, `abe`, `bb`, `bcd`, `ca`, `cb`, `dd`, `de`, `bdir/`, `bdir/cfile`, `a-b`, `aXb`, `.x`, `.y`, `a*b/`, `a*b/ooo`},
	},
	{
		pattern: `a\*c`,
		misses:  []string{`abc`},
	},
	{
		pattern: ``,
		matches: []string{``},
	},
	{
		pattern: `*/man*/bash.*`,
		matches: []string{`man/man1/bash.1`},
		misses:  []string{`a`, `b`, `c`, `d`, `abc`, `abd`, `abe`, `bb`, `bcd`, `ca`, `cb`, `dd`, `de`, `bdir/`, `bdir/cfile`, `a-b`, `aXb`, `.x`, `.y`, `a*b/`, `a*b/ooo`, `man/`, `man/man1/`},
	},
	{
		pattern: `man/man1/bash.1`,
		matches: []string{`man/man1/bash.1`},
		misses:  []string{`a`, `b`, `c`, `d`, `abc`, `abd`, `abe`, `bb`, `bcd`, `ca`, `cb`, `dd`, `de`, `bdir/`, `bdir/cfile`, `a-b`, `aXb`, `.x`, `.y`, `a*b/`, `a*b/ooo`, `man/`, `man/man1/`},
	},
	{
		pattern: `a***c`,
		matches: []string{`abc`},
	},
	{
		pattern: `a*****?c`,
		matches: []string{`abc`},
	},
	{
		pattern: `?*****??`,
		matches: []string{`abc`},
	},
	{
		pattern: `*****??`,
		matches: []string{`abc`},
	},
	{
		pattern: `?*****?c`,
		matches: []string{`abc`},
	},
	{
		pattern: `?***?****c`,
		matches: []string{`abc`},
	},
	{
		pattern: `?***?****?`,
		matches: []string{`abc`},
	},
	{
		pattern: `?***?****`,
		matches: []string{`abc`},
	},
	{
		pattern: `*******c`,
		matches: []string{`abc`},
	},
	{
		pattern: `*******?`,
		matches: []string{`abc`},
	},
	{
		pattern: `a*cd**?**??k`,
		matches: []string{`abcdecdhjk`},
	},
	{
		pattern: `a**?**cd**?**??k`,
		matches: []string{`abcdecdhjk`},
	},
	{
		pattern: `a**?**cd**?**??k***`,
		matches: []string{`abcdecdhjk`},
	},
	{
		pattern: `a**?**cd**?**??***k`,
		matches: []string{`abcdecdhjk`},
	},
	{
		pattern: `a**?**cd**?**??***k**`,
		matches: []string{`abcdecdhjk`},
	},
	{
		pattern: `a****c**?**??*****`,
		matches: []string{`abcdecdhjk`},
	},
	{
		pattern: `[-abc]`,
		matches: []string{`-`},
	},
	{
		pattern: `[abc-]`,
		matches: []string{`-`},
	},
	{
		pattern: `\`,
		matches: []string{`\`},
	},
	{
		pattern: `[\\]`,
		matches: []string{`\`},
	},
	{
		pattern: `[[]`,
		matches: []string{`[`},
	},
	{
		pattern: `[`,
		matches: []string{`[`},
	},
	{
		pattern: `[*`,
		matches: []string{`[abc`},
	},
	{
		pattern: `[]]`,
		matches: []string{`]`},
	},
	{
		pattern: `[]-]`,
		matches: []string{`]`},
	},
	{
		pattern: `[a-z]`,
		matches: []string{`p`},
	},
	{
		pattern: `??**********?****?`,
		misses:  []string{`abc`},
	},
	{
		pattern: `??**********?****c`,
		misses:  []string{`abc`},
	},
	{
		pattern: `?************c****?****`,
		misses:  []string{`abc`},
	},
	{
		pattern: `*c*?**`,
		misses:  []string{`abc`},
	},
	{
		pattern: `a*****c*?**`,
		misses:  []string{`abc`},
	},
	{
		pattern: `a********???*******`,
		misses:  []string{`abc`},
	},
	{
		pattern: `[]`,
		misses:  []string{`a`},
	},
	{
		pattern: `[abc`,
		misses:  []string{`[`},
	},
	{
		pattern: `XYZ`,
		options: minimatch.Options{NoCase: true},
		matches: []string{`xYz`},
		misses:  []string{`ABC`, `IjK`},
	},
	{
		pattern: `ab*`,
		options: minimatch.Options{NoCase: true},
		matches: []string{`ABC`},
		misses:  []string{`xYz`, `IjK`},
	},
	{
		pattern: `[ia]?[ck]`,
		options: minimatch.Options{NoCase: true},
		matches: []string{`ABC`, `IjK`},
		misses:  []string{`xYz`},
	},
	{
		pattern: `{/*,*}`,
		misses:  []string{`/asdf/asdf/asdf`},
	},
	{
		pattern: `{/?,*}`,
		matches: []string{`/a`, `bb`},
		misses:  []string{`/b/b`, `/a/b/c`},
	},
	{
		pattern: `**`,
		matches: []string{`a/b`},
		misses:  []string{`a/.d`, `.a/.d`},
	},
	{
		pattern: `a/*/b`,
		options: minimatch.Options{Dot: true},
		matches: []string{`a/c/b`, `a/.d/b`},
		misses:  []string{`a/./b`, `a/../b`},
	},
	{
		pattern: `a/.*/b`,
		options: minimatch.Options{Dot: true},
		matches: []string{`a/./b`, `a/../b`, `a/.d/b`},
		misses:  []string{`a/c/b`},
	},
	{
		pattern: `a/*/b`,
		matches: []string{`a/c/b`},
		misses:  []string{`a/./b`, `a/../b`, `a/.d/b`},
	},
	{
		pattern: `a/.*/b`,
		matches: []string{`a/./b`, `a/../b`, `a/.d/b`},
		misses:  []string{`a/c/b`},
	},
	{
		pattern: `**`,
		options: minimatch.Options{Dot: true},
		matches: []string{`.a/.d`, `a/.d`, `a/b`},
	},
	{
		pattern: `*(a/b)`,
		misses:  []string{`a/b`},
	},
	{
		pattern: `*(a|{b),c)}`,
		matches: []string{`a`, `ab`, `ac`},
		misses:  []string{`ad`},
	},
	{
		pattern: `[!a*`,
		matches: []string{`[!ab`},
		misses:  []string{`[ab`},
	},
	{
		pattern: `[#a*`,
		matches: []string{`[#ab`},
		misses:  []string{`[ab`},
	},
	{
		pattern: `+(a|*\|c\\|d\\\|e\\\\|f\\\\\|g`,
		matches: []string{`+(a|b\|c\\|d\\|e\\\\|f\\\\|g`},
		misses:  []string{`a`, `b\c`},
	},
	{
		pattern: `*(a|{b,c})`,
		matches: []string{`a`, `b`, `c`, `ab`, `ac`},
		misses:  []string{`d`, `ad`, `bc`, `cb`, `bc,d`, `c,db`, `c,d`, `d)`, `(b|c`, `*(b|c`, `b|c`, `b|cc`, `cb|c`, `x(a|b|c)`, `x(a|c)`, `(a|b|c)`, `(a|c)`},
	},
	{
		pattern: `{a,*(b|c,d)}`,
		matches: []string{`a`, `d)`, `(b|c`, `*(b|c`},
		misses:  []string{`b`, `c`, `d`, `ab`, `ac`, `ad`, `bc`, `cb`, `bc,d`, `c,db`, `c,d`, `b|c`, `b|cc`, `cb|c`, `x(a|b|c)`, `x(a|c)`, `(a|b|c)`, `(a|c)`},
	},
	{
		pattern: `{a,*(b|{c,d})}`,
		matches: []string{`a`, `b`, `c`, `d`, `bc`, `cb`},
		misses:  []string{`ab`, `ac`, `ad`, `bc,d`, `c,db`, `c,d`, `d)`, `(b|c`, `*(b|c`, `b|c`, `b|cc`, `cb|c`, `x(a|b|c)`, `x(a|c)`, `(a|b|c)`, `(a|c)`},
	},
	{
		pattern: `*(a|{b|c,c})`,
		matches: []string{`a`, `b`, `c`, `ab`, `ac`, `bc`, `cb`},
		misses:  []string{`d`, `ad`, `bc,d`, `c,db`, `c,d`, `d)`, `(b|c`, `*(b|c`, `b|c`, `b|cc`, `cb|c`, `x(a|b|c)`, `x(a|c)`, `(a|b|c)`, `(a|c)`},
	},
	{
		pattern: `*(a|{b|c,c})`,
		options: minimatch.Options{NoExt: true},
		matches: []string{`x(a|b|c)`, `x(a|c)`, `(a|b|c)`, `(a|c)`},
		misses:  []string{`a`, `b`, `c`, `d`, `ab`, `ac`, `ad`, `bc`, `cb`, `bc,d`, `c,db`, `c,d`, `d)`, `(b|c`, `*(b|c`, `b|c`, `b|cc`, `cb|c`},
	},
	{
		pattern: `a?b`,
		options: minimatch.Options{MatchBase: true},
		matches: []string{`x/y/acb`, `acb/`},
		misses:  []string{`acb/d/e`, `x/y/acb/d`},
	},
	{
		pattern: `#*`,
		options: minimatch.Options{NoComment: true},
		matches: []string{`#a`, `#b`},
		misses:  []string{`c#d`},
	},
	{
		pattern: `!a*`,
		matches: []string{`d`, `e`, `!ab`, `!abc`, `\!a`},
		misses:  []string{`a!b`},
	},
	{
		pattern: `!a*`,
		options: minimatch.Options{NoNegate: true},
		matches: []string{`!ab`, `!abc`},
		misses:  []string{`d`, `e`, `a!b`, `\!a`},
	},
	{
		pattern: `!!a*`,
		matches: []string{`a!b`},
		misses:  []string{`d`, `e`, `!ab`, `!abc`, `\!a`},
	},
	{
		pattern: `!\!a*`,
		matches: []string{`d`, `e`, `a!b`, `\!a`},
		misses:  []string{`!ab`, `!abc`},
	},
	{
		pattern: `*.!(js)`,
		matches: []string{`foo.bar`, `foo.js.js`, `foo.`, `boo.js.boo`},
		misses:  []string{`foo.js`, `blar.js`},
	},
	{
		pattern: `**/.x/**`,
		matches: []string{`a/b/.x/c`, `a/b/.x/c/d`, `a/b/.x/c/d/e`, `a/b/.x/`, `a/.x/b`, `.x/`, `.x/a`, `.x/a/b`},
		misses:  []string{`a/b/.x`, `.x`, `a/.x/b/.x/c`, `.x/.x`},
	},
	{
		pattern: `[z-a]`,
		misses:  []string{`a/b/.x/c`, `a/b/.x/c/d`, `a/b/.x/c/d/e`, `a/b/.x`, `a/b/.x/`, `a/.x/b`, `.x`, `.x/`, `.x/a`, `.x/a/b`, `a/.x/b/.x/c`, `.x/.x`},
	},
	{
		pattern: `a/[2015-03-10T00:23:08.647Z]/z`,
		misses:  []string{`a/b/.x/c`, `a/b/.x/c/d`, `a/b/.x/c/d/e`, `a/b/.x`, `a/b/.x/`, `a/.x/b`, `.x`, `.x/`, `.x/a`, `.x/a/b`, `a/.x/b/.x/c`, `.x/.x`},
	},
	{
		pattern: `[a-0][a-Ā]`,
		misses:  []string{`a/b/.x/c`, `a/b/.x/c/d`, `a/b/.x/c/d/e`, `a/b/.x`, `a/b/.x/`, `a/.x/b`, `.x`, `.x/`, `.x/a`, `.x/a/b`, `a/.x/b/.x/c`, `.x/.x`},
	},
	{
		pattern: `# ignore this`,
		misses:  []string{`a/b/.x/c`, `a/b/.x/c/d`, `a/b/.x/c/d/e`, `a/b/.x`, `a/b/.x/`, `a/.x/b`, `.x`, `.x/`, `.x/a`, `.x/a/b`, `a/.x/b/.x/c`, `.x/.x`},
	},
}

// TestCorpus runs every corpus case in both directions.
func TestCorpus(t *testing.T) {
	for _, test := range corpus {
		matcher := minimatch.New(test.pattern, test.options)
		for _, path := range test.matches {
			if !matcher.Match(path) {
				t.Errorf("Match(%q, %q, %+v) = false, want true", test.pattern, path, test.options)
			}
		}
		for _, path := range test.misses {
			if matcher.Match(path) {
				t.Errorf("Match(%q, %q, %+v) = true, want false", test.pattern, path, test.options)
			}
		}
	}
}
