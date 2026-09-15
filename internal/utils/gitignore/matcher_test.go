// cspell:ignore ignorecase
package gitignore

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestMatcherExplicitPatterns(t *testing.T) {
	cases := []struct {
		patterns  []string
		path      string
		sensitive bool
		want      bool
	}{
		{[]string{"*.js\n!a.js"}, "b.js", false, false},
		{[]string{"*.js", "!a.js"}, "b.js", false, true},
		{[]string{"a\nb"}, "a\nb", false, true},
		{[]string{"[!b]oo.js"}, "foo.js", false, true},
		{[]string{"[^b]oo.js"}, "foo.js", false, true},
		{[]string{"[!b]oo.js"}, "!oo.js", false, true},
		{[]string{"[^b]oo.js"}, "^oo.js", false, true},
		{[]string{"[fb]oo.js"}, "foo.js", false, true},
		{[]string{"[!b]oo.js"}, "boo.js", false, false},
		{[]string{"[^b]oo.js"}, "boo.js", false, false},
		{[]string{"a\u00a0"}, "a\u00a0", false, true},
		{[]string{"a\t"}, "a\t", false, true},
		{[]string{"??.js"}, "😀.js", false, true},
		{[]string{"?.js"}, "😀.js", false, false},
		{[]string{"[😀].js"}, "😀.js", false, false},
		{[]string{"K.js", "[K].js"}, "K.js", false, false},
		{[]string{"K.js"}, "k.js", false, true},
		{[]string{" a"}, " a", false, true},
		{[]string{"a\u00a0"}, "a", false, false},
		{[]string{"a\n"}, "a", false, false},
		{[]string{"a\\\t"}, "a\t", false, true},
		{[]string{"a\\\u00a0"}, "a\u00a0", false, true},
		{[]string{"a\\  "}, "a ", false, true},
		{[]string{"a\\\\ "}, "a\\", false, true},
		{[]string{"\ufeff#a"}, "#a", false, false},
		{[]string{"\ufeff!a"}, "!a", false, false},
		{[]string{"a//b"}, "a/b", false, false},
		// Preserve glob meaning even where ignore 5 interprets regexp escapes.
		{[]string{`a\*`}, "a.js", false, false},
		{[]string{`a\*`}, "a*", false, true},
		{[]string{`a\?`}, "a?", false, true},
		{[]string{`a\b`}, "ab", false, true},
		{[]string{"[abc"}, "[abc", false, true},
		{[]string{"[z-a]"}, "[z-a]", false, true},
		{[]string{"[z-a0-9]"}, "0", false, false},
		{[]string{"**/a"}, "b\n/a", false, true},

		{[]string{"*.js"}, "nested/a.js", true, true},
		{[]string{"/*.js"}, "nested/a.js", true, false},
		{[]string{"build/", "!build/a.js"}, "build/a.js", true, true},
		{[]string{"build/*", "!build/a.js"}, "build/a.js", true, false},
		{[]string{"build/", "!build/", "build/drop.js"}, "build/drop.js", true, true},
		{[]string{"build/", "!build/", "build/drop.js"}, "build/keep.js", true, false},
		{[]string{"cache/"}, "cache", true, false},
		{[]string{"cache/"}, "cache/", true, true},
		{[]string{"cache/**"}, "cache", true, false},
		{[]string{"cache/**"}, "cache/file.js", true, true},
		{[]string{`\#literal`, `\!literal`, `trailing\ `, "# comment"}, "#literal", true, true},
		{[]string{`\#literal`, `\!literal`, `trailing\ `}, "!literal", true, true},
		{[]string{`trailing\ `}, "trailing ", true, true},
		{[]string{"*.JS"}, "a.js", false, true},
		{[]string{"*.JS"}, "a.js", true, false},
		{[]string{"**"}, "../a.js", true, false},
		{[]string{"**"}, "./a.js", true, false},
		{[]string{"**"}, "a//b.js", true, false},
		{[]string{"**"}, "/a.js", true, false},
		{[]string{"**"}, "", true, false},
	}
	for _, test := range cases {
		if got := NewMatcher(test.patterns, test.sensitive).Match(test.path); got != test.want {
			t.Errorf("patterns %q, path %q: got %v, want %v", test.patterns, test.path, got, test.want)
		}
	}
}

func TestMatcherFromText(t *testing.T) {
	for _, text := range []string{"*.js\n!a.js", "*.js\r\n!a.js"} {
		matcher := NewMatcherFromText(text, false)
		if !matcher.Match("b.js") || matcher.Match("a.js") {
			t.Errorf("text %q did not preserve separate rules", text)
		}
	}
}

// Validate intended glob meaning against Git, independently of ignore's
// JavaScript implementation. No path needs to exist for --no-index matching.
func TestMatcherGitPatternMeaning(t *testing.T) {
	git, err := exec.LookPath("git")
	if err != nil {
		t.Skip("git is not installed")
	}
	directory := t.TempDir()
	if out, err := exec.Command(git, "init", "-q", "--template=", directory).CombinedOutput(); err != nil {
		t.Fatalf("git init: %s: %v", out, err)
	}
	cases := []struct {
		pattern string
		names   []string
	}{
		{"[!b]oo.js", []string{"foo.js", "boo.js", "!oo.js"}},
		{"[^b]oo.js", []string{"foo.js", "boo.js", "^oo.js"}},
		{`cli\*`, []string{"cli*", "cli.js"}},
		{`cli\?`, []string{"cli?", "cli1"}},
		{`a\b.js`, []string{"ab.js", "a.js"}},
		{"a\u00a0", []string{"a", "a\u00a0"}},
		{"a\t", []string{"a", "a\t"}},
		{"a\\\t", []string{"a ", "a\t"}},
		{"**/cli.js", []string{"dir/cli.js", "dir\n/cli.js"}},
		{"build/\n!build/a.js", []string{"build/a.js", "build/b.js"}},
		{"build/*\n!build/a.js", []string{"build/a.js", "build/b.js"}},
	}
	for _, test := range cases {
		t.Run(test.pattern, func(t *testing.T) {
			if err := os.WriteFile(filepath.Join(directory, ".gitignore"), []byte(test.pattern+"\n"), 0600); err != nil {
				t.Fatal(err)
			}
			command := exec.Command(git, "-C", directory, "-c", "core.ignorecase=false", "-c", "core.excludesFile="+filepath.Join(directory, "no-global-excludes"), "check-ignore", "--no-index", "--stdin", "-z")
			command.Stdin = strings.NewReader(strings.Join(test.names, "\x00") + "\x00")
			output, err := command.Output()
			if err != nil {
				var exit *exec.ExitError
				if !errors.As(err, &exit) || exit.ExitCode() != 1 {
					t.Fatal(err)
				}
			}
			ignored := map[string]bool{}
			for _, name := range strings.Split(string(output), "\x00") {
				ignored[name] = true
			}
			matcher := NewMatcherFromText(test.pattern, true)
			for _, name := range test.names {
				if got := matcher.Match(name); got != ignored[name] {
					t.Errorf("%q: matcher=%v git=%v", name, got, ignored[name])
				}
			}
		})
	}
}
