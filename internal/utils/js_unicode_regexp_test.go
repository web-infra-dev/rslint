package utils

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"

	"github.com/iceisfun/gojs/jsregexp"
)

func TestJSUnicodeRegexpMatchesECMAScriptUnicodeSemantics(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
		input   string
		want    bool
	}{
		{name: "general category", pattern: `^\p{Letter}+$`, input: "日本語", want: true},
		{name: "script", pattern: `^\p{Script=Latin}+$`, input: "abcé", want: true},
		{name: "script extensions", pattern: `^\p{Script_Extensions=Hiragana}+$`, input: "ー", want: true},
		{name: "script is not script extensions", pattern: `^\p{Script=Hiragana}+$`, input: "ー", want: false},
		{name: "binary property", pattern: `^\p{Emoji}$`, input: "😀", want: true},
		{name: "negated property", pattern: `^\P{Letter}+$`, input: "123", want: true},
		{name: "unicode dot matches astral", pattern: `^.$`, input: "😀", want: true},
		{name: "dot excludes line separator", pattern: `^.$`, input: "\u2028", want: false},
		{name: "dot includes next line", pattern: `^.$`, input: "\u0085", want: true},
		{name: "word remains ASCII", pattern: `^\w+$`, input: "é", want: false},
		{name: "lookbehind", pattern: `(?<=a)b`, input: "ab", want: true},
		{name: "named backreference", pattern: `^(?<word>a+)\k<word>$`, input: "aaaa", want: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			re, err := CompileJSUnicodeRegexp(test.pattern)
			if err != nil {
				t.Fatalf("CompileJSUnicodeRegexp(%q): %v", test.pattern, err)
			}
			got, err := re.MatchString(test.input)
			if err != nil {
				t.Fatalf("MatchString: %v", err)
			}
			if got != test.want {
				t.Fatalf("match = %v, want %v", got, test.want)
			}
		})
	}
}

func TestJSUnicodeRegexpRejectsInvalidUnicodePatterns(t *testing.T) {
	patterns := []string{
		`(?i)a`,
		`\x{20}`,
		`\p{letter}`,
		`\p{Greek}`,
		`\p{Script-Extensions=Hiragana}`,
		`\p{Emoji=Yes}`,
		`\pL`,
		`[a-\d]`,
	}
	for _, pattern := range patterns {
		t.Run(pattern, func(t *testing.T) {
			if _, err := CompileJSUnicodeRegexp(pattern); err == nil {
				t.Fatalf("CompileJSUnicodeRegexp(%q) succeeded, want syntax error", pattern)
			}
		})
	}
}

func TestJSUnicodeRegexpNormalizesAllValidGroupNameSpellings(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
		input   string
		want    bool
	}{
		{name: "other id start", pattern: `^(?<℘>a)\k<\u2118>$`, input: "aa", want: true},
		{name: "other id continue", pattern: `^(?<a·>b)\k<a\u00B7>$`, input: "bb", want: true},
		{name: "Unicode 16 start", pattern: `^(?<𑎀>a)\k<\u{11380}>$`, input: "aa", want: true},
		{name: "Unicode 17 start", pattern: `^(?<𖄀>a)\k<\u{16100}>$`, input: "aa", want: true},
		{name: "braced escape leading zeroes in capture", pattern: `^(?<\u{00000061}>x)\k<a>$`, input: "xx", want: true},
		{name: "braced escape leading zeroes in backreference", pattern: `^(?<a>x)\k<\u{0000000000000061}>$`, input: "xx", want: true},
		{name: "surrogate escape pair", pattern: `^(?<\uD804\uDF80>a)\k<𑎀>$`, input: "aa", want: true},
		{name: "forward reference", pattern: `^\k<℘>(?<\u2118>a)$`, input: "a", want: true},
		{name: "duplicate name first alternative", pattern: `^(?:(?<℘>a)|(?<\u2118>b))\k<\u{2118}>$`, input: "aa", want: true},
		{name: "duplicate name second alternative", pattern: `^(?:(?<℘>a)|(?<\u2118>b))\k<\u{2118}>$`, input: "bb", want: true},
		{name: "dollar and joiner", pattern: "^(?<$a\u200C>x)\\k<\\u0024a\\u200C>$", input: "xx", want: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			re, err := CompileJSUnicodeRegexp(test.pattern)
			if err != nil {
				t.Fatalf("CompileJSUnicodeRegexp(%q): %v", test.pattern, err)
			}
			got, err := re.MatchString(test.input)
			if err != nil {
				t.Fatalf("MatchString: %v", err)
			}
			if got != test.want {
				t.Fatalf("match = %v, want %v", got, test.want)
			}
		})
	}
}

func TestNormalizeJSRegexpGroupNamesLexicalBoundaries(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
		want    string
	}{
		{
			name:    "lookbehind is untouched",
			pattern: `(?<=a)(?<!b)(?<℘>c)\k<℘>`,
			want:    `(?<=a)(?<!b)(?<rslintGroup0>c)\k<rslintGroup0>`,
		},
		{
			name:    "character class is untouched",
			pattern: `[\k<℘>](?<℘>a)\k<℘>`,
			want:    `[\k<℘>](?<rslintGroup0>a)\k<rslintGroup0>`,
		},
		{
			name:    "escaped pseudo syntax is untouched",
			pattern: `\\k<℘>\(?<℘>(?<℘>a)\k<℘>`,
			want:    `\\k<℘>\(?<℘>(?<rslintGroup0>a)\k<rslintGroup0>`,
		},
		{
			name:    "two source spellings share one generated name",
			pattern: `(?<℘>a)|(?<\u2118>b)|\k<\u{2118}>`,
			want:    `(?<rslintGroup0>a)|(?<rslintGroup0>b)|\k<rslintGroup0>`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := normalizeJSRegexpGroupNames(test.pattern)
			if err != nil {
				t.Fatalf("normalizeJSRegexpGroupNames: %v", err)
			}
			if got != test.want {
				t.Fatalf("normalized = %q, want %q", got, test.want)
			}
		})
	}
}

func TestJSUnicodeRegexpRejectsInvalidGroupNames(t *testing.T) {
	patterns := []string{
		`(?<>a)`,
		`(?<1a>a)`,
		`(?<a-b>a)`,
		`(?<\x61>a)`,
		`(?<\uD800>a)`,
		`(?<\uDC00>a)`,
		`(?<\uD800\u0041>a)`,
		`(?<\u{D800}>a)`,
		`(?<\u{110000}>a)`,
		`(?<\u{000000110000}>a)`,
		`(?<\u{}>a)`,
		`\k<missing>`,
	}
	for _, pattern := range patterns {
		t.Run(pattern, func(t *testing.T) {
			if _, err := CompileJSUnicodeRegexp(pattern); err == nil {
				t.Fatalf("CompileJSUnicodeRegexp(%q) succeeded, want syntax error", pattern)
			}
		})
	}
}

func TestJSUnicodeRegexpBudgetDoesNotMutateCompiledMatcher(t *testing.T) {
	re, err := CompileJSUnicodeRegexp(`^(a+)+$|!$`)
	if err != nil {
		t.Fatal(err)
	}
	_, err = re.MatchString(strings.Repeat("a", 24) + "!")
	if !errors.Is(err, jsregexp.ErrBudget) {
		t.Fatalf("pathological match error = %v, want ErrBudget", err)
	}
	matched, err := re.MatchString("!")
	if err != nil || !matched {
		t.Fatalf("ordinary match after ErrBudget = (%v, %v), want (true, nil)", matched, err)
	}
}

func TestJSUnicodeRegexpBoundsCaptureResetWork(t *testing.T) {
	re, err := CompileJSUnicodeRegexp("z" + strings.Repeat("()", JSUnicodeRegexpMaxGroups) + "|!$")
	if err != nil {
		t.Fatal(err)
	}
	captureSlots := 2 * (JSUnicodeRegexpMaxGroups + 1)
	attackLength := JSUnicodeRegexpCaptureSlotWorkBudget/captureSlots + 1
	_, err = re.MatchString(strings.Repeat("a", attackLength))
	if !errors.Is(err, jsregexp.ErrBudget) {
		t.Fatalf("capture-reset-heavy match error = %v, want ErrBudget", err)
	}
	matched, err := re.MatchString("!")
	if err != nil || !matched {
		t.Fatalf("ordinary match after capture reset budget = (%v, %v), want (true, nil)", matched, err)
	}
}

func TestJSUnicodeRegexpBoundsCaptureWorkPerStep(t *testing.T) {
	pattern := "^(?:(?=" + strings.Repeat("()", JSUnicodeRegexpMaxGroups-4) + ")(a+))+$"
	re, err := CompileJSUnicodeRegexp(pattern)
	if err != nil {
		t.Fatal(err)
	}
	_, err = re.MatchString(strings.Repeat("a", 24) + "!")
	if !errors.Is(err, jsregexp.ErrBudget) {
		t.Fatalf("capture-copy-heavy match error = %v, want ErrBudget", err)
	}
}

func TestJSUnicodeRegexpRejectsUnmeteredLookaroundCaptureWork(t *testing.T) {
	pattern := "^(?:z" + strings.Repeat("()", 2047) + "|" + strings.Repeat("(?=)", 2048) + "b)$"
	_, err := CompileJSUnicodeRegexp(pattern)
	if !errors.Is(err, ErrJSUnicodeRegexpResourceLimit) {
		t.Fatalf("lookaround capture-copy pattern error = %v, want resource limit", err)
	}
}

func TestJSUnicodeRegexpLookaroundCaptureWorkBoundary(t *testing.T) {
	const captures = 1999 // 4,000 capture slots including the whole match.
	atLimit := strings.Repeat("()", captures) + strings.Repeat("(?=)", 125)
	if _, err := CompileJSUnicodeRegexp(atLimit); err != nil {
		t.Fatalf("pattern at lookaround capture-work limit: %v", err)
	}
	aboveLimit := strings.Repeat("()", captures) + strings.Repeat("(?=)", 126)
	if _, err := CompileJSUnicodeRegexp(aboveLimit); !errors.Is(err, ErrJSUnicodeRegexpResourceLimit) {
		t.Fatalf("pattern above lookaround capture-work limit error = %v, want resource limit", err)
	}
}

func TestJSUnicodeRegexpRejectsCombinatorialZeroStepPaths(t *testing.T) {
	branch := "(?:|||||||||)" // Ten empty alternatives.
	pattern := "^" + strings.Repeat(branch, 10) + "(?!)$"
	_, err := CompileJSUnicodeRegexp(pattern)
	if !errors.Is(err, ErrJSUnicodeRegexpResourceLimit) {
		t.Fatalf("combinatorial zero-step pattern error = %v, want resource limit", err)
	}
}

func TestJSUnicodeRegexpRejectsZeroMaximumQuantifierPaths(t *testing.T) {
	branch := "(?:a{0}|b{0}|c{0}|d{0}|e{0}|f{0}|g{0}|h{0}|i{0}|j{0})"
	pattern := "^" + strings.Repeat(branch, 10) + "(?!)$"
	_, err := CompileJSUnicodeRegexp(pattern)
	if !errors.Is(err, ErrJSUnicodeRegexpResourceLimit) {
		t.Fatalf("zero-maximum quantifier path error = %v, want resource limit", err)
	}
}

func TestJSUnicodeRegexpAccumulatesZeroStepFailingAlternatives(t *testing.T) {
	unit := strings.Repeat("(?:|||)", 8) + "(?!)"
	pattern := "^(?:" + strings.Repeat(unit+"|", 15) + unit + ")$"
	_, err := CompileJSUnicodeRegexp(pattern)
	if !errors.Is(err, ErrJSUnicodeRegexpResourceLimit) {
		t.Fatalf("zero-step failing alternatives error = %v, want resource limit", err)
	}
}

func TestJSUnicodeRegexpWeightsZeroStepSuffixWork(t *testing.T) {
	prefix := strings.Repeat("(?:|||)", 8)
	for name, suffix := range map[string]string{
		"groups":   strings.Repeat("(?:)", 4000),
		"captures": strings.Repeat("()", 4000),
	} {
		t.Run(name, func(t *testing.T) {
			pattern := "^" + prefix + suffix + "(?!)$"
			_, err := CompileJSUnicodeRegexp(pattern)
			if !errors.Is(err, ErrJSUnicodeRegexpResourceLimit) {
				t.Fatalf("weighted zero-step suffix error = %v, want resource limit", err)
			}
		})
	}
}

func TestJSUnicodeRegexpWeightsRepeatedLookaroundCaptureWork(t *testing.T) {
	pattern := "^" + strings.Repeat("()", 100) + strings.Repeat("(?:|||)", 7) + "(?!)$"
	_, err := CompileJSUnicodeRegexp(pattern)
	if !errors.Is(err, ErrJSUnicodeRegexpResourceLimit) {
		t.Fatalf("weighted lookaround capture-work error = %v, want resource limit", err)
	}
}

func TestJSUnicodeRegexpWeightsCaptureFanoutAfterStep(t *testing.T) {
	// The character charges one engine step, but the nullable suffix can then
	// invoke every enclosing capture continuation and the failing lookaround
	// 16,384 times without charging another step.
	pattern := "^(((a" + strings.Repeat("(?:|||)", 7) + ")))(?!)$"
	_, err := CompileJSUnicodeRegexp(pattern)
	if !errors.Is(err, ErrJSUnicodeRegexpResourceLimit) {
		t.Fatalf("step-separated capture fanout error = %v, want resource limit", err)
	}
}

func TestJSUnicodeRegexpRejectsPostStepFanoutAcrossConcatBoundaries(t *testing.T) {
	branch := "(?:|||||||||)" // Ten empty alternatives.
	pattern := "(?:a" + strings.Repeat(branch, 4) + ")" + strings.Repeat(branch, 4) + "(?!)"
	_, err := CompileJSUnicodeRegexp(pattern)
	if !errors.Is(err, ErrJSUnicodeRegexpResourceLimit) {
		t.Fatalf("cross-boundary post-step fanout error = %v, want resource limit", err)
	}
}

func TestJSUnicodeRegexpTracksPostStepTailAcrossConcatBoundaries(t *testing.T) {
	re, err := jsregexp.Compile(`(?:a(?:|||))(?:|||)`, "u")
	if err != nil {
		t.Fatal(err)
	}
	got := jsUnicodeRegexpZeroStepComplexityFor(re.AST().Body)
	if got.maxPaths != 16 {
		t.Fatalf("cross-boundary post-step max paths = %d, want 16", got.maxPaths)
	}
}

func TestJSUnicodeRegexpTracksPostStepTailAcrossLaterAlternatives(t *testing.T) {
	re, err := jsregexp.Compile(`(?:a(?:|||)|)(?:|||)`, "u")
	if err != nil {
		t.Fatal(err)
	}
	got := jsUnicodeRegexpZeroStepComplexityFor(re.AST().Body)
	if got.maxPaths != 20 {
		t.Fatalf("post-step alternative max paths = %d, want 20", got.maxPaths)
	}
}

func TestJSUnicodeRegexpRejectsPostStepWorkAcrossLaterAlternatives(t *testing.T) {
	branch := "(?:|||)" // Four empty alternatives.
	fanout := strings.Repeat(branch, 7)
	pattern := "(?:a" + fanout + "|" + fanout + ")(?!)"
	_, err := CompileJSUnicodeRegexp(pattern)
	if !errors.Is(err, ErrJSUnicodeRegexpResourceLimit) {
		t.Fatalf("post-step alternative work error = %v, want resource limit", err)
	}
}

func TestJSUnicodeRegexpRejectsZeroStepPrefixAcrossLaterAlternatives(t *testing.T) {
	// The first alternative stays just below the zero-step limit. After it
	// reaches and fails the outer continuation, the second alternative traverses
	// another capture-heavy zero-step prefix before its first metered character.
	// Both runs belong to the same segment between matcher steps.
	zero := "(?:" + strings.Repeat("|", 48) + ")" // 49 empty alternatives.
	first := strings.Repeat("(", 1000) + zero + strings.Repeat(")", 1000)
	second := strings.Repeat("()", 3093) + "a"
	pattern := "(?:" + first + "|" + second + ")(?!)"
	_, err := CompileJSUnicodeRegexp(pattern)
	if !errors.Is(err, ErrJSUnicodeRegexpResourceLimit) {
		t.Fatalf("cross-alternative zero-step prefix error = %v, want resource limit", err)
	}
}

func TestJSUnicodeRegexpRejectsCaptureEntryPrefixAcrossLaterAlternatives(t *testing.T) {
	// Entering a capture saves two slots before its body reaches the first
	// metered consumer. Those wrappers follow the first alternative's nearly
	// maximal zero-step failure in the same segment.
	zero := "(?:" + strings.Repeat("|", 48) + ")" // 49 empty alternatives.
	first := strings.Repeat("(", 1000) + zero + strings.Repeat(")", 1000)
	second := strings.Repeat("(", 1000) + "ab" + strings.Repeat(")", 1000)
	pattern := "(?:" + first + "|" + second + ")(?!)"
	_, err := CompileJSUnicodeRegexp(pattern)
	if !errors.Is(err, ErrJSUnicodeRegexpResourceLimit) {
		t.Fatalf("cross-alternative capture-entry prefix error = %v, want resource limit", err)
	}
}

func TestJSUnicodeRegexpUsesLookbehindExecutionOrder(t *testing.T) {
	// gojs executes a lookbehind concat from right to left. The empty fanout is
	// therefore explored before the capture-heavy failing assertion, multiplying
	// work that a source-order walk would only add once.
	failing := strings.Repeat("(", 1000) + "(?!)" + strings.Repeat(")", 1000)
	fanout := "(?:" + strings.Repeat("|", 50) + ")" // 51 empty alternatives.
	_, err := CompileJSUnicodeRegexp("(?<=" + failing + fanout + ")")
	if !errors.Is(err, ErrJSUnicodeRegexpResourceLimit) {
		t.Fatalf("lookbehind execution-order error = %v, want resource limit", err)
	}
}

func TestJSUnicodeRegexpRejectsContinuationFailureResumeWork(t *testing.T) {
	// The right continuation can charge a step, exhaust a large zero-step tail,
	// and fail back into a later stepful path of the left matcher. The right tail
	// and the left matcher's resumed prefix are one segment between two steps.
	fanout := strings.Repeat("(?:|||)", 7)
	failing := "(" + fanout + ")(?!)"
	left := "(?:|(?:(?:" + failing + "|)c))"
	right := "(?:a" + failing + ")"
	_, err := CompileJSUnicodeRegexp(left + right)
	if !errors.Is(err, ErrJSUnicodeRegexpResourceLimit) {
		t.Fatalf("continuation-failure resume error = %v, want resource limit", err)
	}
}

func TestJSUnicodeRegexpPropagatesContinuationFailureResumeWork(t *testing.T) {
	// A leading consumer hides the left matcher entry prefix, but after the
	// right continuation fails the already-entered left matcher still resumes
	// its later branch. That retry cost must propagate through the nested concat.
	fanout := strings.Repeat("(?:|||)", 7)
	failing := "(" + fanout + ")(?!)"
	left := "(?:s(?:|(?:(?:" + failing + "|)c)))"
	right := "(?:a" + failing + ")"
	_, err := CompileJSUnicodeRegexp(left + right)
	if !errors.Is(err, ErrJSUnicodeRegexpResourceLimit) {
		t.Fatalf("nested continuation-failure resume error = %v, want resource limit", err)
	}
}

func TestJSUnicodeRegexpRejectsRepeatedZeroStepContinuationWork(t *testing.T) {
	// The left matcher can invoke the right continuation, resume after that
	// continuation fails, and invoke it again without charging an internal step.
	// The right tail, the left retry, and the right prefix before its next step
	// therefore form one zero-step segment.
	fanout := strings.Repeat("(?:|||)", 7)
	failing := "(" + fanout + ")(?!)"
	left := "(?:|" + failing + "|)"
	right := "(?:a" + failing + ")"
	_, err := CompileJSUnicodeRegexp(left + right)
	if !errors.Is(err, ErrJSUnicodeRegexpResourceLimit) {
		t.Fatalf("repeated zero-step continuation error = %v, want resource limit", err)
	}
}

func TestJSUnicodeRegexpTracksMixedCaptureFanout(t *testing.T) {
	re, err := jsregexp.Compile(`(((?:)|a(?:|||)(?:|||)))`, "u")
	if err != nil {
		t.Fatal(err)
	}
	got := jsUnicodeRegexpZeroStepComplexityFor(re.AST().Body)
	if got.paths != 1 || got.maxPaths != 16 || got.work != 6 || got.maxWork != 89 {
		t.Fatalf("mixed capture complexity = {paths:%d maxPaths:%d work:%d maxWork:%d}, want {1 16 6 89}", got.paths, got.maxPaths, got.work, got.maxWork)
	}
}

func TestJSUnicodeRegexpZeroStepWorkBoundary(t *testing.T) {
	branch := "(?:|||)" // Four empty alternatives.
	if _, err := CompileJSUnicodeRegexp(strings.Repeat(branch, 7)); err != nil {
		t.Fatalf("weighted zero-step work below the limit should be accepted: %v", err)
	}
	if _, err := CompileJSUnicodeRegexp(strings.Repeat(branch, 8)); !errors.Is(err, ErrJSUnicodeRegexpResourceLimit) {
		t.Fatalf("weighted zero-step work above the limit error = %v, want resource limit", err)
	}
}

func TestJSUnicodeRegexpConcurrentMatch(t *testing.T) {
	re, err := CompileJSUnicodeRegexp(`^(?<letter>\p{Letter}+)\k<letter>$`)
	if err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	for range 16 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range 100 {
				matched, matchErr := re.MatchString("日日")
				if matchErr != nil || !matched {
					t.Errorf("MatchString = (%v, %v), want (true, nil)", matched, matchErr)
					return
				}
			}
		}()
	}
	wg.Wait()
}

func TestJSUnicodeRegexpCompileAdversarialShapes(t *testing.T) {
	var named strings.Builder
	for i := range 256 {
		fmt.Fprintf(&named, "(?<group%d>a?)", i)
	}
	valid := map[string]string{
		"deep nesting":          strings.Repeat("(", JSUnicodeRegexpMaxNestingDepth) + "a" + strings.Repeat(")", JSUnicodeRegexpMaxNestingDepth),
		"many captures":         strings.Repeat("(a?)", JSUnicodeRegexpMaxGroups),
		"many named groups":     named.String(),
		"long ASCII group name": "(?<" + strings.Repeat("a", 64<<10) + ">x)",
		"property dense":        strings.Repeat(`\p{Letter}`, JSUnicodeRegexpMaxPropertyEscapes),
		"long literal":          strings.Repeat("a", JSUnicodeRegexpMaxPatternBytes),
	}
	for name, pattern := range valid {
		t.Run(name, func(t *testing.T) {
			if _, err := CompileJSUnicodeRegexp(pattern); err != nil {
				t.Fatalf("valid adversarial pattern failed to compile: %v", err)
			}
		})
	}

	invalidSyntax := map[string]string{
		"deep unterminated groups": strings.Repeat("(", 128),
		"long invalid escape":      strings.Repeat(`\u{`, 2_048),
		"long invalid group name":  "(?<" + strings.Repeat("a", 4_096),
	}
	for name, pattern := range invalidSyntax {
		t.Run(name, func(t *testing.T) {
			if _, err := CompileJSUnicodeRegexp(pattern); err == nil {
				t.Fatal("invalid adversarial pattern compiled successfully")
			}
		})
	}

	resourceLimited := map[string]string{
		"pattern bytes":     strings.Repeat("a", JSUnicodeRegexpMaxPatternBytes+1),
		"nesting depth":     strings.Repeat("(", JSUnicodeRegexpMaxNestingDepth+1) + "a" + strings.Repeat(")", JSUnicodeRegexpMaxNestingDepth+1),
		"groups":            strings.Repeat("(?:)", JSUnicodeRegexpMaxGroups+1),
		"property escapes":  strings.Repeat(`\p{Letter}`, JSUnicodeRegexpMaxPropertyEscapes+1),
		"fatal stack shape": strings.Repeat("(", 1_000_000) + "a" + strings.Repeat(")", 1_000_000),
	}
	for name, pattern := range resourceLimited {
		t.Run(name, func(t *testing.T) {
			if _, err := CompileJSUnicodeRegexp(pattern); !errors.Is(err, ErrJSUnicodeRegexpResourceLimit) {
				t.Fatalf("resource-limited pattern error = %v, want ErrJSUnicodeRegexpResourceLimit", err)
			}
		})
	}
}

func TestCountJSUnicodeRegexpPropertyEscapes(t *testing.T) {
	tests := map[string]int{
		`\p{Letter}`:                1,
		`[\P{Letter}]`:              1,
		`\\p{Letter}`:               0,
		`\\\p{Letter}`:              1,
		`\pLetter`:                  0,
		`\\P{Letter}\p{Letter}`:     1,
		`[()]\(\)(?:\p{Number})`:    1,
		`plain text with p{Letter}`: 0,
	}
	for pattern, want := range tests {
		if got := CountJSUnicodeRegexpPropertyEscapes(pattern); got != want {
			t.Errorf("CountJSUnicodeRegexpPropertyEscapes(%q) = %d, want %d", pattern, got, want)
		}
	}
}
