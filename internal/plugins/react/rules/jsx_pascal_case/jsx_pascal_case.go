package jsx_pascal_case

import (
	"regexp"
	"strings"
	"unicode"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/react/reactutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

const (
	msgUsePascalCase        = "Imported JSX component {{name}} must be in PascalCase"
	msgUsePascalOrSnakeCase = "Imported JSX component {{name}} must be in PascalCase or SCREAMING_SNAKE_CASE"
)

type pascalCaseOptions struct {
	allowAllCaps           bool
	allowLeadingUnderscore bool
	allowNamespace         bool
	ignore                 []compiledPattern
}

// compiledPattern pairs the raw ignore entry with its regex translation. The
// raw string handles the exact-match fast path and the regex covers
// minimatch-style globs (including extglob `+(a|b)`). `re` may be nil when
// the pattern failed to compile — in that case only the exact-match arm
// matches, which degrades gracefully on malformed input.
type compiledPattern struct {
	raw string
	re  *regexp.Regexp
}

func parseOptions(opts any) pascalCaseOptions {
	cfg := pascalCaseOptions{}
	optsMap := utils.GetOptionsMap(opts)
	if optsMap == nil {
		return cfg
	}
	if v, ok := optsMap["allowAllCaps"].(bool); ok {
		cfg.allowAllCaps = v
	}
	if v, ok := optsMap["allowLeadingUnderscore"].(bool); ok {
		cfg.allowLeadingUnderscore = v
	}
	if v, ok := optsMap["allowNamespace"].(bool); ok {
		cfg.allowNamespace = v
	}
	if raw, ok := optsMap["ignore"].([]interface{}); ok {
		for _, entry := range raw {
			if s, ok := entry.(string); ok {
				cfg.ignore = append(cfg.ignore, compiledPattern{raw: s, re: compileMinimatch(s)})
			}
		}
	}
	return cfg
}

// testDigit mirrors the upstream ASCII-only digit check (`charCode 48-57`).
func testDigit(r rune) bool {
	return r >= '0' && r <= '9'
}

// testUpperCase mirrors upstream `char === char.toUpperCase() && upperCase !== char.toLowerCase()`.
// True iff the rune is a cased letter in its upper form — excludes digits
// and non-letter symbols (where upper == lower).
func testUpperCase(r rune) bool {
	return unicode.ToUpper(r) == r && unicode.ToLower(r) != r
}

// testLowerCase mirrors upstream `char === char.toLowerCase() && lowerCase !== char.toUpperCase()`.
func testLowerCase(r rune) bool {
	return unicode.ToLower(r) == r && unicode.ToUpper(r) != r
}

// isNonAlphaNumeric mirrors upstream
// `char.toLowerCase() === char.toUpperCase() && !testDigit(char)` — true for
// symbols like `_`, `-`, `$`, where upper and lower coincide, excluding digits.
func isNonAlphaNumeric(r rune) bool {
	return unicode.ToLower(r) == unicode.ToUpper(r) && !testDigit(r)
}

// testPascalCase returns true when `name` begins with an upper-case letter,
// contains no non-alphanumeric characters after the first, and includes at
// least one lower-case letter or digit after the first position. A bare
// uppercase letter (e.g. `"T"`) fails the final "has lower/digit" check and
// returns false — callers short-circuit on single-character names before
// reaching here, so the bare-upper case never reaches testPascalCase.
func testPascalCase(name string) bool {
	runes := []rune(name)
	if len(runes) == 0 {
		return false
	}
	if !testUpperCase(runes[0]) {
		return false
	}
	rest := runes[1:]
	for _, r := range rest {
		if isNonAlphaNumeric(r) {
			return false
		}
	}
	for _, r := range rest {
		if testLowerCase(r) || testDigit(r) {
			return true
		}
	}
	return false
}

// testAllCaps mirrors upstream: first char upper or digit, every interior
// char upper / digit / underscore, last char upper or digit. Names of length
// 1 never reach this function because `splitName.length === 1` short-circuits.
func testAllCaps(name string) bool {
	runes := []rune(name)
	if len(runes) == 0 {
		return false
	}
	first := runes[0]
	if !testUpperCase(first) && !testDigit(first) {
		return false
	}
	// Upstream loops `i = 1; i < name.length - 1` — i.e. strictly interior
	// positions only. For a 2-char name the loop body never executes.
	for i := 1; i < len(runes)-1; i++ {
		c := runes[i]
		if !testUpperCase(c) && !testDigit(c) && c != '_' {
			return false
		}
	}
	last := runes[len(runes)-1]
	return testUpperCase(last) || testDigit(last)
}

func ignoreCheck(patterns []compiledPattern, name string) bool {
	for _, p := range patterns {
		if name == p.raw {
			return true
		}
		if p.re != nil && p.re.MatchString(name) {
			return true
		}
	}
	return false
}

// compileMinimatch translates a minimatch-style glob (with extglob support)
// into an anchored Go regex. Covers the syntax used by
// eslint-plugin-react/jsx-pascal-case `ignore` entries:
//
//   - `*`          — any run of characters
//   - `?`          — a single character
//   - `[...]`      — character class (leading `!` / `^` inverts)
//   - `?(a|b)`     — zero or one of alternatives
//   - `*(a|b)`     — zero or more of alternatives
//   - `+(a|b)`     — one or more of alternatives
//   - `@(a|b)`     — exactly one of alternatives
//
// Upstream passes `{ noglobstar: true }` — `**` is collapsed to `*`. Returns
// nil on malformed patterns; callers treat that as "exact-match only".
// NOTE: `!(...)` negation would require lookarounds which RE2 does not
// support. None of the upstream tests exercise it; we still emit a non-nil
// regex so the pattern isn't silently dropped, but it matches permissively
// (equivalent to `(?:...)?`). Document if a user-facing divergence appears.
func compileMinimatch(pattern string) *regexp.Regexp {
	body := minimatchBody([]rune(pattern))
	re, err := regexp.Compile("^" + body + "$")
	if err != nil {
		return nil
	}
	return re
}

// minimatchBody / findMatchingClose / splitTopLevelAlts operate on `[]rune`
// so that indices and slicing are codepoint-aligned — essential when
// `ignore` patterns contain non-ASCII characters (`Año_*`, CJK, emoji).
// Mixing rune indices with `string` byte offsets would misalign after the
// first multi-byte rune; keeping all three helpers on the same rune slice
// avoids that trap.
func minimatchBody(runes []rune) string {
	var sb strings.Builder
	i := 0
	for i < len(runes) {
		r := runes[i]
		if i+1 < len(runes) && runes[i+1] == '(' && strings.ContainsRune("?*+@!", r) {
			if end, ok := findMatchingClose(runes, i+2); ok {
				alts := splitTopLevelAlts(runes[i+2 : end])
				parts := make([]string, len(alts))
				for j, a := range alts {
					parts[j] = minimatchBody(a)
				}
				body := strings.Join(parts, "|")
				switch r {
				case '?':
					sb.WriteString("(?:" + body + ")?")
				case '*':
					sb.WriteString("(?:" + body + ")*")
				case '+':
					sb.WriteString("(?:" + body + ")+")
				case '@':
					sb.WriteString("(?:" + body + ")")
				case '!':
					// RE2 lacks lookarounds; approximate as "zero or one"
					// so the pattern still compiles. The upstream test
					// suite doesn't exercise `!(...)`.
					sb.WriteString("(?:" + body + ")?")
				}
				i = end + 1
				continue
			}
		}
		switch r {
		case '*':
			for i < len(runes) && runes[i] == '*' {
				i++
			}
			sb.WriteString(".*")
		case '?':
			sb.WriteString(".")
			i++
		case '[':
			closeIdx := -1
			for j := i + 1; j < len(runes); j++ {
				if runes[j] == ']' {
					closeIdx = j
					break
				}
			}
			// Need at least two runes between `[` and `]` so `[^x]` / `[!x]`
			// negation doesn't collapse to an empty inverted class (`[^]`),
			// which RE2 rejects as malformed.
			if closeIdx > i+1 {
				body := string(runes[i+1 : closeIdx])
				if len(body) > 1 && (body[0] == '!' || body[0] == '^') {
					body = "^" + body[1:]
				}
				sb.WriteString("[" + body + "]")
				i = closeIdx + 1
			} else {
				sb.WriteString("\\[")
				i++
			}
		case '\\':
			if i+1 < len(runes) {
				sb.WriteString(regexp.QuoteMeta(string(runes[i+1])))
				i += 2
			} else {
				sb.WriteString("\\\\")
				i++
			}
		default:
			sb.WriteString(regexp.QuoteMeta(string(r)))
			i++
		}
	}
	return sb.String()
}

func findMatchingClose(runes []rune, start int) (int, bool) {
	depth := 1
	for j := start; j < len(runes); j++ {
		switch runes[j] {
		case '(':
			depth++
		case ')':
			depth--
			if depth == 0 {
				return j, true
			}
		}
	}
	return -1, false
}

func splitTopLevelAlts(runes []rune) [][]rune {
	var parts [][]rune
	depth := 0
	start := 0
	for i := range runes {
		switch runes[i] {
		case '(':
			depth++
		case ')':
			depth--
		case '|':
			if depth == 0 {
				parts = append(parts, runes[start:i])
				start = i + 1
			}
		}
	}
	parts = append(parts, runes[start:])
	return parts
}

var JsxPascalCaseRule = rule.Rule{
	Name: "react/jsx-pascal-case",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		cfg := parseOptions(options)

		check := func(element *ast.Node) {
			if reactutil.IsDOMComponent(element) {
				return
			}
			name := reactutil.GetJsxElementTypeString(element)
			if name == "" {
				return
			}

			// Upstream splits on `:` first, else on `.` — never on both.
			var checkNames []string
			if strings.LastIndex(name, ":") > -1 {
				checkNames = strings.Split(name, ":")
			} else if strings.LastIndex(name, ".") > -1 {
				checkNames = strings.Split(name, ".")
			} else {
				checkNames = []string{name}
			}

			for index := range checkNames {
				splitName := checkNames[index]
				// Single-character parts are always allowed — matches upstream
				// `if (splitName.length === 1) return undefined;` which bails
				// out of the whole rule for this element.
				if len([]rune(splitName)) == 1 {
					return
				}
				isIgnored := ignoreCheck(cfg.ignore, splitName)

				checkName := splitName
				if cfg.allowLeadingUnderscore && strings.HasPrefix(splitName, "_") {
					checkName = splitName[1:]
				}
				isPascal := testPascalCase(checkName)
				isAllowedAllCaps := cfg.allowAllCaps && testAllCaps(checkName)

				if !isPascal && !isAllowedAllCaps && !isIgnored {
					msgId := "usePascalCase"
					msg := msgUsePascalCase
					if cfg.allowAllCaps {
						msgId = "usePascalOrSnakeCase"
						msg = msgUsePascalOrSnakeCase
					}
					ctx.ReportNode(element, rule.RuleMessage{
						Id:          msgId,
						Description: strings.ReplaceAll(msg, "{{name}}", splitName),
					})
					break
				}
				// Upstream's `do...while (index < checkNames.length && !allowNamespace)`:
				// with allowNamespace, the loop exits after the first successful
				// iteration; without, it continues through every split part.
				if cfg.allowNamespace {
					break
				}
			}
		}

		return rule.RuleListeners{
			ast.KindJsxOpeningElement:     check,
			ast.KindJsxSelfClosingElement: check,
		}
	},
}
