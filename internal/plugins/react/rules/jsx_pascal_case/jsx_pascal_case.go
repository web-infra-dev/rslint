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
				cfg.ignore = append(cfg.ignore, compiledPattern{raw: s, re: reactutil.GlobToRegex(s)})
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

var JsxPascalCaseRule = rule.Rule{
	Name: "react/jsx-pascal-case",
	Run: func(ctx rule.RuleContext, _options []any) rule.RuleListeners {
		options := rule.UnwrapOptions(_options)
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
