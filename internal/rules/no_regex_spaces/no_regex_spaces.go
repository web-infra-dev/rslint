package no_regex_spaces

import (
	"fmt"
	"regexp"
	"sort"

	"github.com/dlclark/regexp2"
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

var (
	doubleSpaceRe   = regexp.MustCompile(` {2}`)
	spacesPatternRe = regexp.MustCompile(`( {2,})(?: [+*{?]|[^+*{?]|$)`)
)

// https://eslint.org/docs/latest/rules/no-regex-spaces
var NoRegexSpacesRule = rule.Rule{
	Name: "no-regex-spaces",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		sf := ctx.SourceFile

		reportAndMaybeFix := func(reportNode *ast.Node, pattern, rawPattern string, rawPatternStart int, flags utils.RegexFlags) {
			// Fast path: skip if raw source has no consecutive spaces. This mirrors
			// ESLint's early-out — avoids false positives like RegExp(' \ ') where
			// the parsed pattern has two adjacent spaces but the source doesn't.
			if !doubleSpaceRe.MatchString(rawPattern) {
				return
			}

			// Match ESLint's `try { regExpParser.parsePattern(...) } catch { return }`:
			// skip any pattern that fails to parse under its flags.
			if !isValidRegexPattern(pattern, flags) {
				return
			}

			classRanges := collectClassRanges(pattern, flags)

			matches := spacesPatternRe.FindAllStringSubmatchIndex(pattern, -1)
			for _, m := range matches {
				index := m[2]
				length := m[3] - m[2]

				if indexInAnyRange(index, classRanges) {
					continue
				}

				msg := rule.RuleMessage{
					Id:          "multipleSpaces",
					Description: fmt.Sprintf("Spaces are hard to count. Use {%d}.", length),
				}
				if pattern == rawPattern {
					fix := rule.RuleFixReplaceRange(
						core.NewTextRange(rawPatternStart+index, rawPatternStart+index+length),
						fmt.Sprintf(" {%d}", length),
					)
					ctx.ReportNodeWithFixes(reportNode, msg, fix)
				} else {
					ctx.ReportNode(reportNode, msg)
				}
				// Report only the first occurrence of consecutive spaces.
				return
			}
		}

		checkRegExpConstructor := func(callee *ast.Node, args *ast.NodeList, reportNode *ast.Node) {
			callee = ast.SkipParentheses(callee)
			if callee == nil || callee.Kind != ast.KindIdentifier || callee.AsIdentifier().Text != "RegExp" {
				return
			}
			if utils.IsShadowed(callee, "RegExp") {
				return
			}
			if args == nil || len(args.Nodes) == 0 {
				return
			}

			patternNode := args.Nodes[0]
			if patternNode.Kind != ast.KindStringLiteral {
				return
			}
			pattern := patternNode.AsStringLiteral().Text

			patternRange := utils.TrimNodeTextRange(sf, patternNode)
			rawFull := sf.Text()[patternRange.Pos():patternRange.End()]
			if len(rawFull) < 2 {
				return
			}
			rawPattern := rawFull[1 : len(rawFull)-1]
			rawPatternStart := patternRange.Pos() + 1

			flagsStr := ""
			if len(args.Nodes) >= 2 {
				flagsNode := args.Nodes[1]
				if flagsNode.Kind != ast.KindStringLiteral {
					// Flags cannot be determined — ESLint skips in this case.
					return
				}
				flagsStr = flagsNode.AsStringLiteral().Text
			}

			reportAndMaybeFix(reportNode, pattern, rawPattern, rawPatternStart, utils.ParseRegexFlags(flagsStr))
		}

		return rule.RuleListeners{
			ast.KindRegularExpressionLiteral: func(node *ast.Node) {
				text := node.Text()
				pattern, flagsStr := utils.ExtractRegexPatternAndFlags(text)
				if pattern == "" {
					return
				}
				nodeRange := utils.TrimNodeTextRange(sf, node)
				rawPatternStart := nodeRange.Pos() + 1
				reportAndMaybeFix(node, pattern, pattern, rawPatternStart, utils.ParseRegexFlags(flagsStr))
			},
			ast.KindCallExpression: func(node *ast.Node) {
				call := node.AsCallExpression()
				checkRegExpConstructor(call.Expression, call.Arguments, node)
			},
			ast.KindNewExpression: func(node *ast.Node) {
				newExpr := node.AsNewExpression()
				checkRegExpConstructor(newExpr.Expression, newExpr.Arguments, node)
			},
		}
	},
}

// collectClassRanges returns byte ranges [start, end) for every character class
// in the pattern (including v-flag nested classes). Ranges are sorted by start
// and may overlap (outer wraps inner).
func collectClassRanges(pattern string, flags utils.RegexFlags) []classRange {
	var ranges []classRange
	utils.IterateRegexCharacterClasses(pattern, flags, func(start, end int) {
		ranges = append(ranges, classRange{start: start, end: end})
	})
	// IterateRegexCharacterClasses under v-flag visits nested classes first,
	// so sort ascending to let binary search short-circuit.
	sort.Slice(ranges, func(i, j int) bool { return ranges[i].start < ranges[j].start })
	return ranges
}

type classRange struct {
	start int
	end   int
}

func indexInAnyRange(idx int, ranges []classRange) bool {
	for _, cr := range ranges {
		if idx < cr.start {
			return false
		}
		if idx < cr.end {
			return true
		}
	}
	return false
}

// isValidRegexPattern reports whether the pattern parses cleanly under its
// flags, mirroring ESLint's `regexpp.parsePattern` try/catch. We combine two
// checks:
//
//   - regexp2 compile catches structural errors (unclosed `[`, unmatched
//     quantifier under Unicode mode, bad hex escapes, etc.).
//   - A narrow u-flag identity-escape check for the handful of escapes
//     regexp2 accepts but ES-u-mode rejects (`\a`, `\9`, …). ESLint rejects
//     these under u/v, and silently skipping them keeps us aligned.
//
// If ANY check fails the pattern is treated as unparsable and the rule
// skips reporting — matching ESLint.
func isValidRegexPattern(pattern string, flags utils.RegexFlags) bool {
	var opt regexp2.RegexOptions = regexp2.ECMAScript
	if flags.UV() {
		opt |= regexp2.Unicode
	}
	if _, err := regexp2.Compile(pattern, opt); err != nil {
		return false
	}
	if flags.UV() {
		if hasInvalidIdentityEscapeForUFlag(pattern) {
			return false
		}
		if hasUnmatchedBraceForUFlag(pattern) {
			return false
		}
	}
	return true
}

// hasUnmatchedBraceForUFlag reports whether the pattern contains a literal `{`
// that is not part of a valid `{n}` / `{n,}` / `{n,m}` quantifier or a
// recognized `\u{...}` / `\p{...}` / `\q{...}` escape. Under the u/v flag this
// is a SyntaxError per ECMAScript, but regexp2 accepts it (its .NET lineage
// treats a bare `{` as literal). The scan is outside character classes only —
// inside a class, `{` is always literal.
func hasUnmatchedBraceForUFlag(pattern string) bool {
	inClass := false
	i := 0
	for i < len(pattern) {
		c := pattern[i]
		if c == '\\' {
			if i+1 >= len(pattern) {
				return false
			}
			i += 2
			continue
		}
		if c == '[' && !inClass {
			inClass = true
			i++
			continue
		}
		if c == ']' && inClass {
			inClass = false
			i++
			continue
		}
		if inClass {
			i++
			continue
		}
		if c == '{' && !looksLikeQuantifier(pattern, i) {
			return true
		}
		i++
	}
	return false
}

// looksLikeQuantifier reports whether pattern[start] opens a valid
// `{n}` / `{n,}` / `{n,m}` quantifier.
func looksLikeQuantifier(pattern string, start int) bool {
	i := start + 1
	digits := 0
	for i < len(pattern) && pattern[i] >= '0' && pattern[i] <= '9' {
		i++
		digits++
	}
	if digits == 0 {
		return false
	}
	if i < len(pattern) && pattern[i] == ',' {
		i++
		for i < len(pattern) && pattern[i] >= '0' && pattern[i] <= '9' {
			i++
		}
	}
	return i < len(pattern) && pattern[i] == '}'
}

// hasInvalidIdentityEscapeForUFlag scans for escapes that ECMAScript u/v mode
// rejects but regexp2 accepts. Conservative: on malformed input it returns
// false (i.e. defers to regexp2's verdict) rather than inventing an error.
func hasInvalidIdentityEscapeForUFlag(pattern string) bool {
	i := 0
	for i < len(pattern) {
		c := pattern[i]
		if c != '\\' || i+1 >= len(pattern) {
			i++
			continue
		}
		next := pattern[i+1]
		switch next {
		// Recognized single-letter escapes.
		case 'd', 'D', 'w', 'W', 's', 'S', 'b', 'B', 'n', 't', 'r', 'v', 'f', '0',
			'x', 'u', 'c', 'p', 'P', 'k', 'q',
			// SyntaxCharacter / `/` — legal identity escapes under u.
			'^', '$', '.', '*', '+', '?', '(', ')', '[', ']', '{', '}', '|', '\\', '/':
			i += 2
			continue
		}
		// Decimal backreference (\1..\9) is legal.
		if next >= '1' && next <= '9' {
			i += 2
			continue
		}
		// Any letter/digit identity escape not recognized above is illegal under u.
		if (next >= 'a' && next <= 'z') || (next >= 'A' && next <= 'Z') {
			return true
		}
		i += 2
	}
	return false
}
