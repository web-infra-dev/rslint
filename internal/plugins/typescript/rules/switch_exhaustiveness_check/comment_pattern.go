package switch_exhaustiveness_check

import (
	"fmt"
	"sync"

	"github.com/web-infra-dev/rslint/internal/utils"
)

const (
	maxCachedCommentPatterns       = 64
	maxCachedCommentPatternBytes   = utils.JSUnicodeRegexpMaxPatternBytes
	maxCachedCommentPropertyTokens = utils.JSUnicodeRegexpMaxPropertyEscapes
)

type cachedCommentPattern struct {
	regexp         *utils.JSUnicodeRegexp
	err            error
	bytes          int
	propertyTokens int
}

// commentPatternCache is bounded by entry count and by two estimates of
// retained compiler weight. In particular, jsregexp materializes Unicode
// property ranges per escape, so a small number of property-dense patterns can
// otherwise retain far more memory than an entry-count bound suggests.
type commentPatternCache struct {
	sync.RWMutex
	entries        map[string]cachedCommentPattern
	order          []string
	bytes          int
	propertyTokens int
}

var defaultCaseCommentPatternCache = commentPatternCache{
	entries: make(map[string]cachedCommentPattern),
}

func (c *commentPatternCache) compile(pattern string) (*utils.JSUnicodeRegexp, error) {
	// Check an existing accepted entry before rescanning its key. A long-pattern
	// hit pays only the map's hash/equality work, without a separate lexical
	// resource scan; an oversized key reaches the compiler's cheap byte-length
	// guard without a full scan.
	if len(pattern) <= maxCachedCommentPatternBytes {
		c.RLock()
		cached, ok := c.entries[pattern]
		c.RUnlock()
		if ok {
			return cached.regexp, cached.err
		}
	}

	weight := cachedCommentPattern{
		bytes: len(pattern),
	}
	if weight.bytes <= maxCachedCommentPatternBytes {
		weight.propertyTokens = utils.CountJSUnicodeRegexpPropertyEscapes(pattern)
	}
	cacheable := weight.bytes <= maxCachedCommentPatternBytes && weight.propertyTokens <= maxCachedCommentPropertyTokens

	if !cacheable {
		weight.regexp, weight.err = utils.CompileJSUnicodeRegexp(pattern)
		return weight.regexp, weight.err
	}

	c.Lock()
	defer c.Unlock()
	if cached, ok := c.entries[pattern]; ok {
		return cached.regexp, cached.err
	}
	// Config validation is cold and compilation can be memory-intensive near
	// the accepted resource limits. Keeping the lock through a cache miss makes
	// compilation single-flight, preventing concurrent validators for the same
	// option from multiplying peak memory.
	weight.regexp, weight.err = utils.CompileJSUnicodeRegexp(pattern)
	for len(c.order) >= maxCachedCommentPatterns ||
		c.bytes+weight.bytes > maxCachedCommentPatternBytes ||
		c.propertyTokens+weight.propertyTokens > maxCachedCommentPropertyTokens {
		c.evictOldest()
	}
	c.entries[pattern] = weight
	c.order = append(c.order, pattern)
	c.bytes += weight.bytes
	c.propertyTokens += weight.propertyTokens
	return weight.regexp, weight.err
}

func (c *commentPatternCache) evictOldest() {
	oldest := c.order[0]
	c.order[0] = ""
	c.order = c.order[1:]
	entry := c.entries[oldest]
	delete(c.entries, oldest)
	c.bytes -= entry.bytes
	c.propertyTokens -= entry.propertyTokens
}

func validateSwitchExhaustivenessCheckOptions(options []any) error {
	opts := parseOptions(options)
	if opts.DefaultCaseCommentPattern == nil {
		return nil
	}
	if _, err := defaultCaseCommentPatternCache.compile(*opts.DefaultCaseCommentPattern); err != nil {
		return fmt.Errorf("defaultCaseCommentPattern must be a valid ECMAScript Unicode regular expression: %w", err)
	}
	return nil
}

// A custom commentPatternMatcher owns file-local fuse state. Compiled regexps
// are immutable and shared, while tripped prevents one pathological
// configuration from spending the match budget repeatedly across switches in
// the same file. The default matcher has no state and is shared directly.
type commentPatternMatcher struct {
	custom                    *utils.JSUnicodeRegexp
	lookaroundCaptureSlotWork int
	zeroStepWork              int
	tripped                   bool
}

var defaultCommentPatternMatcher = &commentPatternMatcher{}

func newCommentPatternMatcher(pattern *string) *commentPatternMatcher {
	if pattern == nil {
		// The default ASCII matcher is stateless, so all files can share it.
		return defaultCommentPatternMatcher
	}
	compiled, err := defaultCaseCommentPatternCache.compile(*pattern)
	if err != nil {
		// Normal config loading calls the schema post-validator first. Reaching
		// this branch means an internal caller bypassed that contract; never
		// silently fall back to the default pattern.
		panic(fmt.Sprintf("validated defaultCaseCommentPattern failed to compile: %v", err))
	}
	return &commentPatternMatcher{custom: compiled}
}

func (m *commentPatternMatcher) matches(text string) bool {
	if m.custom == nil {
		return matchesDefaultCommentPattern(text)
	}
	if m.tripped {
		return false
	}
	lookaroundWork := m.custom.LookaroundCaptureSlotWork()
	zeroStepWork := m.custom.ZeroStepWork()
	if lookaroundWork > utils.JSUnicodeRegexpLookaroundCaptureSlotWorkBudget-m.lookaroundCaptureSlotWork ||
		zeroStepWork > utils.JSUnicodeRegexpZeroStepWorkBudget-m.zeroStepWork {
		m.tripped = true
		return false
	}
	m.lookaroundCaptureSlotWork += lookaroundWork
	m.zeroStepWork += zeroStepWork
	matched, err := m.custom.MatchString(text)
	if utils.IsJSUnicodeRegexpBudgetError(err) {
		m.tripped = true
	}
	return err == nil && matched
}

func matchesDefaultCommentPattern(text string) bool {
	const expected = "no default"
	if len(text) != len(expected) {
		return false
	}
	for i := range len(expected) {
		got := text[i]
		if got >= 'A' && got <= 'Z' {
			got += 'a' - 'A'
		}
		if got != expected[i] {
			return false
		}
	}
	return true
}
