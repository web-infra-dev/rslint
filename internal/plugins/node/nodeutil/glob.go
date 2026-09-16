// cspell:ignore globrex
package nodeutil

import (
	"fmt"
	"strings"

	"github.com/web-infra-dev/rslint/internal/utils/ecmascript"
	esregexp "github.com/web-infra-dev/rslint/internal/utils/ecmascript/regexp"
)

// GlobMatcher keeps literal names out of the regexp engine.
type GlobMatcher struct {
	literal    string
	expression *esregexp.RegExp
}

// CompileGlob implements the non-extended, globstar-enabled patterns used by
// Node restrictions and path conversion (globrex 0.1.2). Only * and ** are
// wildcards; ?, brackets, braces and backslashes are literal. Unlike minimatch,
// stars also match empty segments, . and .., without normalizing the input.
func CompileGlob(pattern string) *GlobMatcher {
	if !strings.Contains(pattern, "*") && !strings.Contains(pattern, "//") {
		return &GlobMatcher{literal: pattern}
	}
	units := ecmascript.StringCodeUnitRunes(pattern)
	var source strings.Builder
	source.WriteByte('^')
	for i := 0; i < len(units); i++ {
		switch c := units[i]; c {
		case '*':
			start := i
			for i+1 < len(units) && units[i+1] == '*' {
				i++
			}
			if i > start && (start == 0 || units[start-1] == '/') && (i+1 == len(units) || units[i+1] == '/') {
				source.WriteString(`(?:[^/]*(?:/|$))*`)
				i++ // A globstar includes its following separator.
			} else {
				source.WriteString(`[^/]*`)
			}
		case '/':
			source.WriteByte('/')
			if i+1 < len(units) && units[i+1] == '/' {
				source.WriteByte('?')
			}
		default:
			if c > 0x7f {
				fmt.Fprintf(&source, `\u%04x`, c)
				continue
			}
			if strings.ContainsRune(`\^$.+?()[]{}|`, c) {
				source.WriteByte('\\')
			}
			source.WriteByte(byte(c))
		}
	}
	source.WriteByte('$')
	return &GlobMatcher{expression: esregexp.MustCompile(source.String(), "")}
}

// Match compares a module name or path without normalizing its separators.
func (matcher *GlobMatcher) Match(name string) bool {
	if matcher.expression == nil {
		return ecmascript.CompareStrings(matcher.literal, name) == 0
	}
	// JavaScript glob expressions have no Unicode flag: wildcards can split
	// surrogate pairs. Keep that encoding choice inside the shared matcher.
	matched, err := matcher.expression.Unwrap().MatchRunes(ecmascript.StringCodeUnitRunes(name))
	return err == nil && matched
}
