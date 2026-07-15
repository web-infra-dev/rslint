package discovery

// cspell:ignore filetest filetests filetesttest srcsrc srctest xfoo

import (
	"testing"

	"gotest.tools/v3/assert"
)

func TestGlobParent(t *testing.T) {
	t.Parallel()

	tests := []struct {
		pattern string
		want    string
	}{
		{pattern: "/repo/src/*.js", want: "/repo/src"},
		{pattern: "/repo/src/**/index.js", want: "/repo/src"},
		{pattern: "/repo/{src,test}/*.js", want: "/repo"},
		{pattern: "/repo/src/{a/b,c/d}.js", want: "/repo/src"},
		{pattern: "/repo/src/+(foo|bar)/*.js", want: "/repo/src"},
		{pattern: "/repo/src/!(foo)/*.js", want: "/repo/src"},
		{pattern: "/repo/src/file.js", want: "/repo/src"},
		{pattern: "/repo/src/", want: "/repo/src"},
		{pattern: "src/*.js", want: "src"},
		{pattern: "*.js", want: "."},
		{pattern: "!src/*.js", want: "."},
		{pattern: `/repo/a\*b/file.js`, want: "/repo/a*b"},
		{pattern: "/repo/[abc]/file.js", want: "/repo"},
		// glob-parent uses is-glob's strict mode, where a bare question
		// mark is not considered magic.
		{pattern: "/repo/pkg?/file.js", want: "/repo/pkg?"},
	}

	for _, test := range tests {
		t.Run(test.pattern, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, GlobParent(test.pattern), test.want)
		})
	}
}

func TestIsGlobPattern(t *testing.T) {
	t.Parallel()

	tests := []struct {
		pattern string
		want    bool
	}{
		{pattern: "src/file.js", want: false},
		{pattern: "src/*.js", want: true},
		{pattern: "src/**/file.js", want: true},
		{pattern: "src/file?.js", want: false},
		{pattern: "src/[ab].js", want: true},
		{pattern: "{src,test}/*.js", want: true},
		{pattern: "+(src|test)/*.js", want: true},
		{pattern: "src/(unit|e2e)/*.js", want: true},
		{pattern: "src/(?!generated)/*.js", want: true},
		{pattern: "src/file+?.js", want: true},
		{pattern: "!src/*.js", want: true},
		{pattern: `src/\*.js`, want: false},
	}

	for _, test := range tests {
		t.Run(test.pattern, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, isGlobPattern(test.pattern), test.want)
		})
	}
}

func TestSearchPatternExactMatch(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		pattern    string
		basePath   string
		matches    []string
		nonMatches []string
	}{
		{
			name:       "direct file",
			pattern:    "/repo/src/file.js",
			basePath:   "/repo/src",
			matches:    []string{"file.js"},
			nonMatches: []string{"other.js", "nested/file.js"},
		},
		{
			name:       "direct directory expansion",
			pattern:    "/repo/src/**",
			basePath:   "/repo/src",
			matches:    []string{"", "file.js", ".hidden.js", "deep/file.ts"},
			nonMatches: []string{"../file.js"},
		},
		{
			name:       "globstar dot true",
			pattern:    "/repo/**/*.js",
			basePath:   "/repo",
			matches:    []string{"index.js", ".hidden.js", "src/index.js", "src/.hidden.js"},
			nonMatches: []string{"index.ts", "src/index.ts"},
		},
		{
			name:       "question and character class",
			pattern:    "/repo/src/?oo.[jt]s",
			basePath:   "/repo",
			matches:    []string{"src/foo.js", "src/zoo.ts"},
			nonMatches: []string{"src/fooo.js", "src/foo.css", "src/deep/foo.js"},
		},
		{
			name:       "braces including slash",
			pattern:    "/repo/{src,test}/{unit,e2e}/*.{js,ts}",
			basePath:   "/repo",
			matches:    []string{"src/unit/a.js", "src/e2e/a.ts", "test/unit/b.ts"},
			nonMatches: []string{"lib/unit/a.js", "src/other/a.js", "src/unit/a.css"},
		},
		{
			name:       "numeric brace range",
			pattern:    "/repo/file{01..03}.js",
			basePath:   "/repo",
			matches:    []string{"file01.js", "file02.js", "file03.js"},
			nonMatches: []string{"file00.js", "file1.js", "file04.js"},
		},
		{
			name:       "nested and non-expanding braces",
			pattern:    "/repo/{x}a{b,c{d,e}}.js",
			basePath:   "/repo",
			matches:    []string{"{x}ab.js", "{x}acd.js", "{x}ace.js"},
			nonMatches: []string{"xab.js", "{x}ac.js"},
		},
		{
			name:       "relative brace can produce leading slash",
			pattern:    "{,src}/*.js",
			basePath:   "/repo",
			matches:    []string{"src/a.js"},
			nonMatches: []string{"a.js", "deep/a.js", "src/a.ts"},
		},
		{
			name:       "at extglob",
			pattern:    "/repo/@(src|test)/*.js",
			basePath:   "/repo",
			matches:    []string{"src/a.js", "test/a.js"},
			nonMatches: []string{"srcsrc/a.js", "lib/a.js", "src/a.ts"},
		},
		{
			name:       "plus extglob",
			pattern:    "/repo/+(src|test)/*.js",
			basePath:   "/repo",
			matches:    []string{"src/a.js", "test/a.js", "srctest/a.js", "srcsrc/a.js"},
			nonMatches: []string{"lib/a.js", "a.js"},
		},
		{
			name:       "question extglob",
			pattern:    "/repo/file?(test).js",
			basePath:   "/repo",
			matches:    []string{"file.js", "filetest.js"},
			nonMatches: []string{"filetesttest.js", "file.ts"},
		},
		{
			name:       "optional extglob can match empty path component",
			pattern:    "?(src)",
			basePath:   "/repo",
			matches:    []string{"", "src"},
			nonMatches: []string{"srcsrc", "test"},
		},
		{
			name:       "star extglob",
			pattern:    "/repo/file*(test).js",
			basePath:   "/repo",
			matches:    []string{"file.js", "filetest.js", "filetesttest.js"},
			nonMatches: []string{"filetests.js", "file.ts"},
		},
		{
			name:       "negative extglob",
			pattern:    "/repo/pkg/!(fixtures|generated)/*.js",
			basePath:   "/repo",
			matches:    []string{"pkg/src/a.js", "pkg/test/a.js"},
			nonMatches: []string{"pkg/fixtures/a.js", "pkg/generated/a.js", "pkg/src/a.ts"},
		},
		{
			name:       "embedded negative extglob",
			pattern:    "/repo/a!(b|c)d.js",
			basePath:   "/repo",
			matches:    []string{"ad.js", "axd.js", "abcd.js"},
			nonMatches: []string{"abd.js", "acd.js"},
		},
		{
			name:       "negative extglob observes suffix boundary",
			pattern:    "/repo/*.!(js)",
			basePath:   "/repo",
			matches:    []string{"file.ts", "file.jsx"},
			nonMatches: []string{"file.js"},
		},
		{
			name:       "brace output keeps comment and negation literals",
			pattern:    "/repo/{#,x}*.js",
			basePath:   "/repo",
			matches:    []string{"#foo.js", "xfoo.js"},
			nonMatches: []string{"foo.js"},
		},
		{
			name:       "brace output exclamation is literal",
			pattern:    "/repo/{!,x}foo.js",
			basePath:   "/repo",
			matches:    []string{"!foo.js", "xfoo.js"},
			nonMatches: []string{"foo.js"},
		},
		{
			name:       "leading minimatch negation",
			pattern:    "!src/**/*.js",
			basePath:   "/repo",
			matches:    []string{"test/a.js", "src/a.ts", "README.md"},
			nonMatches: []string{"src/a.js", "src/deep/a.js"},
		},
		{
			name:       "double leading negation",
			pattern:    "!!src/**/*.js",
			basePath:   "/repo",
			matches:    []string{"src/a.js", "src/deep/a.js"},
			nonMatches: []string{"test/a.js", "src/a.ts"},
		},
		{
			name:       "absolute pattern relativizes before minimatch negation",
			pattern:    "/repo/!src/**/*.js",
			basePath:   "/repo",
			matches:    []string{"test/a.js", "src/a.ts", "README.md"},
			nonMatches: []string{"src/a.js", "src/deep/a.js"},
		},
		{
			name:       "escaped magic is literal",
			pattern:    `/repo/src/\*.js`,
			basePath:   "/repo",
			matches:    []string{"src/*.js"},
			nonMatches: []string{"src/a.js"},
		},
		{
			name:       "unicode question matches one rune",
			pattern:    "/repo/?.js",
			basePath:   "/repo",
			matches:    []string{"界.js", "a.js"},
			nonMatches: []string{"世界.js"},
		},
		{
			name:       "malformed glob constructs are literals",
			pattern:    `/repo/[abc.js`,
			basePath:   "/repo",
			matches:    []string{"[abc.js"},
			nonMatches: []string{"a.js"},
		},
		{
			name:       "unclosed extglob is literal",
			pattern:    `/repo/+(src|test.js`,
			basePath:   "/repo",
			matches:    []string{"+(src|test.js"},
			nonMatches: []string{"src.js", "test.js"},
		},
		{
			name:       "trailing escape is literal",
			pattern:    "/repo/src/\\",
			basePath:   "/repo",
			matches:    []string{"src/\\"},
			nonMatches: []string{"src"},
		},
		{
			name:       "invalid character range matches nothing",
			pattern:    "/repo/[z-a].js",
			basePath:   "/repo",
			nonMatches: []string{"z.js", "a.js", "-.js"},
		},
		{
			name:       "unicode posix character classes",
			pattern:    "/repo/[[:alpha:]][[:digit:]][[:word:]].js",
			matches:    []string{"界٣_.js", "a1b.js"},
			nonMatches: []string{"_1a.js", "aa_.js", "a1-.js"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			pattern, err := CompileSearchPattern(test.pattern, test.basePath)
			assert.NilError(t, err)
			for _, relativePath := range test.matches {
				assert.Assert(t, pattern.Match(relativePath), "%q should match %q", test.pattern, relativePath)
			}
			for _, relativePath := range test.nonMatches {
				assert.Assert(t, !pattern.Match(relativePath), "%q should not match %q", test.pattern, relativePath)
			}
		})
	}
}

func TestSearchPatternPartialMatch(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		pattern    string
		matches    []string
		nonMatches []string
	}{
		{
			name:       "static directory prefix",
			pattern:    "/repo/src/*.js",
			matches:    []string{"src"},
			nonMatches: []string{"lib", "src/nested", "src/file.ts"},
		},
		{
			name:       "dynamic directory prefix",
			pattern:    "/repo/packages/*/src/**/*.js",
			matches:    []string{"packages", "packages/app", "packages/app/src", "packages/app/src/deep"},
			nonMatches: []string{"src", "packages/app/test", "other/app/src"},
		},
		{
			name:       "brace alternatives",
			pattern:    "/repo/{src,test}/**/*.js",
			matches:    []string{"src", "src/deep", "test", "test/fixtures"},
			nonMatches: []string{"lib", "examples"},
		},
		{
			name:       "extglob alternatives",
			pattern:    "/repo/@(src|test)/**/*.js",
			matches:    []string{"src", "src/deep", "test"},
			nonMatches: []string{"lib", "srcsrc"},
		},
		{
			name:       "negated pattern follows minimatch partial inversion",
			pattern:    "!src/*.js",
			matches:    []string{"test", "src/nested"},
			nonMatches: []string{"src"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			pattern, err := CompileSearchPattern(test.pattern, "/repo")
			assert.NilError(t, err)
			for _, relativeDir := range test.matches {
				assert.Assert(t, pattern.PartialMatch(relativeDir), "%q should partially match %q", test.pattern, relativeDir)
			}
			for _, relativeDir := range test.nonMatches {
				assert.Assert(t, !pattern.PartialMatch(relativeDir), "%q should not partially match %q", test.pattern, relativeDir)
			}
		})
	}
}

func TestCompileSearchPatternDerivesBasePath(t *testing.T) {
	t.Parallel()

	pattern, err := CompileSearchPattern("/repo/packages/*/src/**/*.js", "")
	assert.NilError(t, err)
	assert.Equal(t, pattern.BasePath(), "/repo/packages")
	assert.Equal(t, pattern.RawPattern(), "/repo/packages/*/src/**/*.js")
	assert.Assert(t, pattern.Match("app/src/index.js"))
	assert.Assert(t, !pattern.Match("packages/app/src/index.js"))
}

func TestCompileSearchPatternDerivesUnescapedLiteralBasePath(t *testing.T) {
	t.Parallel()

	pattern, err := CompileSearchPattern(`/repo/packages/a\*b/*.js`, "")
	assert.NilError(t, err)
	assert.Equal(t, pattern.BasePath(), "/repo/packages/a*b")
	assert.Assert(t, pattern.Match("index.js"))
}

func TestCompileSearchPatternAcceptsNormalizedWindowsAbsolutePath(t *testing.T) {
	t.Parallel()

	pattern, err := CompileSearchPattern("C:/repo/packages/*/src/**/*.js", "C:/repo/packages")
	assert.NilError(t, err)
	assert.Assert(t, pattern.Match("app/src/index.js"))
	assert.Assert(t, pattern.PartialMatch("app/src/deep"))
}

func TestCompileSearchPatternAcceptsPatternOutsideBaseLikeNodePathRelative(t *testing.T) {
	t.Parallel()

	pattern, err := CompileSearchPattern("/outside/*.js", "/repo")
	assert.NilError(t, err)
	assert.Assert(t, !pattern.Match("index.js"))
	assert.Assert(t, !pattern.PartialMatch("src"))
}

func TestSearchPatternCommentMatchesNothing(t *testing.T) {
	t.Parallel()

	pattern, err := CompileSearchPattern("# generated files", "/repo")
	assert.NilError(t, err)
	assert.Assert(t, !pattern.Match("# generated files"))
	assert.Assert(t, !pattern.PartialMatch("src"))
}

func TestSearchPatternRetainsESLintUnmatchedKeysForLeadingNegation(t *testing.T) {
	t.Parallel()

	pattern, err := CompileSearchPattern("/repo/!!src/*.js", "/repo")
	assert.NilError(t, err)
	assert.Equal(t, pattern.UnmatchedKey(), "!!src/*.js")
	assert.Equal(t, pattern.MatchRemovalKey(), "src/*.js")
}

func TestSearchPatternEmptyExtglobMatchesMinimatch(t *testing.T) {
	t.Parallel()

	for _, rawPattern := range []string{"?()", "*()", "+()", "@()"} {
		pattern, err := CompileSearchPattern(rawPattern, "/repo")
		assert.NilError(t, err)
		assert.Assert(t, !pattern.Match(""), "%q should not match an empty path component", rawPattern)
		assert.Assert(t, !pattern.Match("a"), "%q should not match a non-empty path component", rawPattern)
	}

	negative, err := CompileSearchPattern("x/!()", "/repo")
	assert.NilError(t, err)
	assert.Assert(t, !negative.Match("x/"))
	assert.Assert(t, negative.Match("x/a"))
}

func TestSearchPatternTraversalComponentsMatchMinimatch(t *testing.T) {
	t.Parallel()

	tests := []struct {
		pattern string
		path    string
		want    bool
	}{
		{pattern: "x/*", path: "x/.", want: false},
		{pattern: "x/.*", path: "x/..", want: false},
		{pattern: `x/\.`, path: "x/.", want: true},
		{pattern: "x/[.]", path: "x/.", want: true},
		{pattern: "x/@(.)", path: "x/.", want: true},
		{pattern: "x/+(.)", path: "x/..", want: true},
		{pattern: "x/!(generated)", path: "x/..", want: true},
	}

	for _, test := range tests {
		pattern, err := CompileSearchPattern(test.pattern, "/repo")
		assert.NilError(t, err)
		assert.Equal(t, pattern.Match(test.path), test.want, "%q against %q", test.pattern, test.path)
	}
}
