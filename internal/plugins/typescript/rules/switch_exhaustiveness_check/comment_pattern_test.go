package switch_exhaustiveness_check

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"

	"github.com/web-infra-dev/rslint/internal/utils"
)

func TestDefaultCaseCommentPatternSchemaUsesECMAScriptUnicode(t *testing.T) {
	if bytes.Contains(SwitchExhaustivenessCheckRule.Schema.RawJSON(), []byte(`"format"`)) {
		t.Fatal("raw upstream schema must not add a regex format")
	}
	for _, pattern := range []string{
		`^\p{Letter}+$`,
		`^\p{Script=Latin}+$`,
		`^\p{Script_Extensions=Hiragana}+$`,
		`^\p{Emoji}$`,
		`^(?<℘>a)\k<\u2118>$`,
		`^(?<\u{00000061}>x)\k<a>$`,
	} {
		options := []any{map[string]any{"defaultCaseCommentPattern": pattern}}
		if err := SwitchExhaustivenessCheckRule.Schema.Validate(options); err != nil {
			t.Errorf("valid ECMAScript Unicode pattern %q was rejected: %v", pattern, err)
		}
	}
	for _, pattern := range []string{`(?i)a`, `\x{20}`, `\p{letter}`, `\p{Script-Extensions=Latin}`} {
		options := []any{map[string]any{"defaultCaseCommentPattern": pattern}}
		if err := SwitchExhaustivenessCheckRule.Schema.Validate(options); err == nil {
			t.Errorf("invalid ECMAScript Unicode pattern %q was accepted", pattern)
		}
	}

	resourceLimited := map[string]string{
		"pattern bytes":       strings.Repeat("a", utils.JSUnicodeRegexpMaxPatternBytes+1),
		"nesting depth":       strings.Repeat("(", utils.JSUnicodeRegexpMaxNestingDepth+1) + "a" + strings.Repeat(")", utils.JSUnicodeRegexpMaxNestingDepth+1),
		"groups":              strings.Repeat("(?:)", utils.JSUnicodeRegexpMaxGroups+1),
		"property escapes":    strings.Repeat(`\p{Letter}`, utils.JSUnicodeRegexpMaxPropertyEscapes+1),
		"lookaround captures": strings.Repeat("()", 2047) + strings.Repeat("(?=)", 2048),
		"zero-step paths":     strings.Repeat("(?:|||||||||)", 10) + "(?!)",
	}
	for name, pattern := range resourceLimited {
		t.Run("resource limit/"+name, func(t *testing.T) {
			options := []any{map[string]any{"defaultCaseCommentPattern": pattern}}
			err := SwitchExhaustivenessCheckRule.Schema.Validate(options)
			if !errors.Is(err, utils.ErrJSUnicodeRegexpResourceLimit) {
				t.Fatalf("schema validation error = %v, want ErrJSUnicodeRegexpResourceLimit", err)
			}
		})
	}
}

func TestDefaultCommentPatternASCIICaseInsensitive(t *testing.T) {
	const lower = "no default"
	letterPositions := []int{0, 1, 3, 4, 5, 6, 7, 8, 9}
	for mask := range 1 << len(letterPositions) {
		candidate := []byte(lower)
		for bit, position := range letterPositions {
			if mask&(1<<bit) != 0 {
				candidate[position] -= 'a' - 'A'
			}
		}
		if !matchesDefaultCommentPattern(string(candidate)) {
			t.Fatalf("default pattern did not match ASCII case variant %q", candidate)
		}
	}

	for _, input := range []string{"", "no  default", "no defaults", " no default", "no default\n", "nø default", "NO DEFAULT!"} {
		if matchesDefaultCommentPattern(input) {
			t.Fatalf("default pattern unexpectedly matched %q", input)
		}
	}
}

func TestCommentPatternMatcherTripsOnlyCurrentFileState(t *testing.T) {
	pattern := `^(a+)+$|!$`
	firstFile := newCommentPatternMatcher(&pattern)
	attack := strings.Repeat("a", 24) + "!"
	if firstFile.matches(attack) {
		t.Fatal("pathological input unexpectedly matched")
	}
	if !firstFile.tripped {
		t.Fatal("budget exhaustion did not trip the file-local circuit breaker")
	}
	if firstFile.matches("!") {
		t.Fatal("tripped matcher should short-circuit later comments in the same file")
	}

	secondFile := newCommentPatternMatcher(&pattern)
	if !secondFile.matches("!") {
		t.Fatal("one file's budget exhaustion polluted the shared matcher")
	}
}

func TestCommentPatternMatcherBoundsCumulativeUnmeteredWork(t *testing.T) {
	t.Run("lookaround capture slots", func(t *testing.T) {
		pattern := strings.Repeat("()", 1999) + strings.Repeat("(?=)", 125)
		firstFile := newCommentPatternMatcher(&pattern)
		if !firstFile.matches("") {
			t.Fatal("first match at the per-file lookaround work limit failed")
		}
		if firstFile.matches("") || !firstFile.tripped {
			t.Fatal("second match did not trip the cumulative lookaround work limit")
		}
		secondFile := newCommentPatternMatcher(&pattern)
		if !secondFile.matches("") {
			t.Fatal("one file's lookaround work charge polluted another file")
		}
	})

	t.Run("zero-step paths", func(t *testing.T) {
		pattern := strings.Repeat("(?:|||)", 7)
		firstFile := newCommentPatternMatcher(&pattern)
		perMatch := firstFile.custom.ZeroStepWork()
		allowed := utils.JSUnicodeRegexpZeroStepWorkBudget / perMatch
		for range allowed {
			if !firstFile.matches("") {
				t.Fatal("match within the per-file zero-step work limit failed")
			}
		}
		if firstFile.matches("") || !firstFile.tripped {
			t.Fatal("match above the cumulative zero-step work limit did not trip")
		}
	})

	t.Run("ordinary lookaround", func(t *testing.T) {
		pattern := `(?=a)a`
		matcher := newCommentPatternMatcher(&pattern)
		for range 100 {
			if !matcher.matches("a") {
				t.Fatal("low-work lookaround unexpectedly tripped the cumulative limit")
			}
		}
	})
}

func TestCommentPatternCacheHasThreeIndependentBounds(t *testing.T) {
	t.Run("entry FIFO", func(t *testing.T) {
		cache := commentPatternCache{entries: make(map[string]cachedCommentPattern)}
		for i := range maxCachedCommentPatterns + 1 {
			pattern := fmt.Sprintf("^entry-%d$", i)
			if _, err := cache.compile(pattern); err != nil {
				t.Fatal(err)
			}
		}
		assertCommentPatternCacheInvariants(t, &cache)
		if len(cache.entries) != maxCachedCommentPatterns {
			t.Fatalf("entries = %d, want %d", len(cache.entries), maxCachedCommentPatterns)
		}
		if _, ok := cache.entries["^entry-0$"]; ok {
			t.Fatal("oldest entry was not evicted")
		}
	})

	t.Run("raw bytes", func(t *testing.T) {
		cache := commentPatternCache{entries: make(map[string]cachedCommentPattern)}
		first := "^" + strings.Repeat("a", maxCachedCommentPatternBytes/2) + "$"
		second := "^" + strings.Repeat("b", maxCachedCommentPatternBytes/2) + "$"
		if _, err := cache.compile(first); err != nil {
			t.Fatal(err)
		}
		if _, err := cache.compile(second); err != nil {
			t.Fatal(err)
		}
		assertCommentPatternCacheInvariants(t, &cache)
		if _, ok := cache.entries[first]; ok {
			t.Fatal("byte budget did not evict the oldest pattern")
		}
	})

	t.Run("Unicode property tokens", func(t *testing.T) {
		cache := commentPatternCache{entries: make(map[string]cachedCommentPattern)}
		first := strings.Repeat(`\p{Letter}`, maxCachedCommentPropertyTokens)
		if _, err := cache.compile(first); err != nil {
			t.Fatal(err)
		}
		if _, err := cache.compile(`\P{Letter}`); err != nil {
			t.Fatal(err)
		}
		assertCommentPatternCacheInvariants(t, &cache)
		if _, ok := cache.entries[first]; ok {
			t.Fatal("property-token budget did not evict the oldest pattern")
		}
	})

	t.Run("resource-limited patterns are rejected and not retained", func(t *testing.T) {
		cache := commentPatternCache{entries: make(map[string]cachedCommentPattern)}
		patterns := []string{
			strings.Repeat("a", maxCachedCommentPatternBytes+1),
			strings.Repeat(`\p{Letter}`, maxCachedCommentPropertyTokens+1),
		}
		for _, pattern := range patterns {
			if _, err := cache.compile(pattern); !errors.Is(err, utils.ErrJSUnicodeRegexpResourceLimit) {
				t.Fatalf("compile error = %v, want ErrJSUnicodeRegexpResourceLimit", err)
			}
		}
		assertCommentPatternCacheInvariants(t, &cache)
		if len(cache.entries) != 0 || len(cache.order) != 0 {
			t.Fatalf("resource-limited patterns were retained: %d entries", len(cache.entries))
		}
	})

	t.Run("compile errors are bounded and cached", func(t *testing.T) {
		cache := commentPatternCache{entries: make(map[string]cachedCommentPattern)}
		if _, err := cache.compile("("); err == nil {
			t.Fatal("invalid pattern compiled")
		}
		if _, err := cache.compile("("); err == nil {
			t.Fatal("cached invalid pattern lost its error")
		}
		assertCommentPatternCacheInvariants(t, &cache)
		if len(cache.entries) != 1 {
			t.Fatalf("entries = %d, want 1", len(cache.entries))
		}
	})

	t.Run("concurrent misses publish one matcher", func(t *testing.T) {
		cache := commentPatternCache{entries: make(map[string]cachedCommentPattern)}
		pattern := strings.Repeat(`\p{Letter}`, 65)
		matchers := make([]*utils.JSUnicodeRegexp, 16)
		errs := make([]error, len(matchers))
		var wg sync.WaitGroup
		for i := range matchers {
			wg.Add(1)
			go func() {
				defer wg.Done()
				matchers[i], errs[i] = cache.compile(pattern)
			}()
		}
		wg.Wait()
		for i := range matchers {
			if errs[i] != nil {
				t.Fatalf("compile %d: %v", i, errs[i])
			}
			if matchers[i] != matchers[0] {
				t.Fatalf("compile %d returned a different matcher", i)
			}
		}
		assertCommentPatternCacheInvariants(t, &cache)
		if len(cache.entries) != 1 {
			t.Fatalf("entries = %d, want 1", len(cache.entries))
		}
	})
}

func assertCommentPatternCacheInvariants(t *testing.T, cache *commentPatternCache) {
	t.Helper()
	cache.Lock()
	defer cache.Unlock()
	if len(cache.entries) != len(cache.order) {
		t.Fatalf("entries/order length mismatch: %d/%d", len(cache.entries), len(cache.order))
	}
	if len(cache.entries) > maxCachedCommentPatterns {
		t.Fatalf("entry count = %d, limit %d", len(cache.entries), maxCachedCommentPatterns)
	}
	if cache.bytes > maxCachedCommentPatternBytes {
		t.Fatalf("cached bytes = %d, limit %d", cache.bytes, maxCachedCommentPatternBytes)
	}
	if cache.propertyTokens > maxCachedCommentPropertyTokens {
		t.Fatalf("cached property tokens = %d, limit %d", cache.propertyTokens, maxCachedCommentPropertyTokens)
	}
	bytes := 0
	tokens := 0
	seen := make(map[string]bool, len(cache.order))
	for _, pattern := range cache.order {
		if seen[pattern] {
			t.Fatalf("duplicate FIFO key %q", pattern)
		}
		seen[pattern] = true
		entry, ok := cache.entries[pattern]
		if !ok {
			t.Fatalf("FIFO key %q is missing from entries", pattern)
		}
		bytes += entry.bytes
		tokens += entry.propertyTokens
	}
	if cache.bytes != bytes || cache.propertyTokens != tokens {
		t.Fatalf("counter mismatch: got (%d bytes, %d properties), want (%d, %d)", cache.bytes, cache.propertyTokens, bytes, tokens)
	}
}
