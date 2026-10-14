package minimatch3

import (
	"strings"

	"github.com/dlclark/regexp2"
	"github.com/web-infra-dev/rslint/internal/utils/ecmascript"
	esregexp "github.com/web-infra-dev/rslint/internal/utils/ecmascript/regexp"
)

// RegexpMatcher matches the regexp produced by minimatch 3's makeRe API.
// Unlike Match, each globstar retains the separators on both sides: a/**/b
// does not match a/b, and repeated separators in the subject are significant.
type RegexpMatcher struct{ re *regexp2.Regexp }

// MakeRe reuses the compiled glob parts. A nil result corresponds to upstream
// returning false for an empty, commented or invalid pattern.
func (m *Matcher) MakeRe() *RegexpMatcher {
	if m.invalid || len(m.set) == 0 {
		return nil
	}
	globstar := `(?:(?!(?:/|^)\.).)*?`
	if m.options.Dot {
		globstar = `(?:(?!(?:/|^)\.{1,2}($|/)).)*?`
	}
	rows := make([]string, 0, len(m.set))
	for _, row := range m.set {
		parts := make([]string, 0, len(row))
		for _, part := range row {
			switch {
			case part.globstar:
				parts = append(parts, globstar)
			case part.re != nil:
				parts = append(parts, part.source)
			default:
				var literal strings.Builder
				for _, unit := range ecmascript.StringCodeUnitRunes(part.literal) {
					literal.WriteString(codeUnitEscape(unit))
				}
				parts = append(parts, literal.String())
			}
		}
		rows = append(rows, strings.Join(parts, `/`))
	}
	source := `^(?:` + strings.Join(rows, `|`) + `)$`
	if m.negate {
		source = `^(?!` + source + `).*$`
	}
	flags := ""
	if m.options.NoCase {
		flags = "i"
	}
	expression, err := esregexp.Compile(source, flags)
	if err != nil {
		return nil
	}
	return &RegexpMatcher{re: expression.Unwrap()}
}

// Test compares UTF-16 code units, as an upstream regexp without /u does.
// Invalid patterns safely match nothing rather than crashing the lint run.
func (m *RegexpMatcher) Test(path string) bool {
	if m == nil {
		return false
	}
	matched, err := m.re.MatchRunes(ecmascript.StringCodeUnitRunes(path))
	return err == nil && matched
}
