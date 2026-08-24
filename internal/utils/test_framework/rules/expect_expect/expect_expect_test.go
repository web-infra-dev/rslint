package expect_expect

import "testing"

func TestAssertMatcherLiteralFastPath(t *testing.T) {
	matcher := compileAssertPatterns([]string{"expect", "assert.equal"})
	tests := []struct {
		name string
		want bool
	}{
		{name: "expect", want: true},
		{name: "EXPECT.toBe", want: true},
		{name: "expectSaga", want: false},
		{name: "assert.equal", want: true},
		{name: "assert.equal.deep", want: true},
		{name: "assert.notEqual", want: false},
	}

	for _, test := range tests {
		if got := matcher.matches(test.name); got != test.want {
			t.Errorf("matches(%q) = %t, want %t", test.name, got, test.want)
		}
	}
}

func TestAssertMatcherLiteralFastPathMatchesRegexp(t *testing.T) {
	patterns := []string{
		"expect",
		"assert.equal",
		"expectSaga",
		"foo_1.bar2",
		"A.B.C",
	}
	names := []string{
		"",
		"expect",
		"EXPECT",
		"expect.toBe",
		"expected",
		"assert",
		"assert.equal",
		"ASSERT.EQUAL.deep",
		"assert.equality",
		"expectSaga",
		"expectSaga.returns",
		"foo_1.bar2",
		"foo_1.bar2.baz",
		"a.b.c",
		"a.b.cd",
		"K.matcher",
		"ſ.matcher",
	}

	for _, pattern := range patterns {
		matcher := compileAssertPatterns([]string{pattern})
		if matcher.hasFallback {
			t.Fatalf("literal pattern %q unexpectedly disabled the fast path", pattern)
		}
		re := compileAssertPattern(pattern)
		for _, name := range names {
			want := name != "" && re.TestOrTimeout(name)
			if got := matcher.matches(name); got != want {
				t.Errorf("pattern %q, name %q: fast path = %t, regexp = %t", pattern, name, got, want)
			}
		}
	}
}

func TestAssertMatcherKeepsJavaScriptUnicodeCaseFolding(t *testing.T) {
	matcher := compileAssertPatterns([]string{"k"})
	if !matcher.matches("K.matcher") {
		t.Error("literal fast path did not fall back to /iu matching for a non-ASCII callee")
	}
}

func TestAssertMatcherWildcardFallback(t *testing.T) {
	matcher := compileAssertPatterns([]string{"request.**.expect"})
	if !matcher.matches("request.get.foo.expect") {
		t.Error("wildcard pattern did not use the regexp fallback")
	}
}

func TestAssertMatcherCandidateGate(t *testing.T) {
	matcher := compileAssertPatterns([]string{"expect", "assert.equal"})
	if matcher.hasFallback {
		t.Fatal("literal patterns unexpectedly disabled the candidate gate")
	}
	for _, pattern := range matcher.patterns {
		if pattern.asciiRoot == "" {
			t.Fatalf("literal pattern %#v has no candidate root", pattern)
		}
	}

	wildcard := compileAssertPatterns([]string{"expect*"})
	if !wildcard.hasFallback {
		t.Fatal("wildcard pattern must keep the conservative fallback")
	}
}
