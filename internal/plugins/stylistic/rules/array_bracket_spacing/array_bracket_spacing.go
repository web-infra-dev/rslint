package array_bracket_spacing

import (
	"unicode/utf8"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

const (
	spacingNever  = "never"
	spacingAlways = "always"
)

type options struct {
	spaced                   bool
	singleElementException   bool
	objectsInArraysException bool
	arraysInArraysException  bool
}

// parseOptions mirrors upstream's option-shape juggling.
//
//	rule: ['array-bracket-spacing']                        → defaults (spaced=false)
//	rule: ['array-bracket-spacing', 'always']              → spaced=true, no exceptions
//	rule: ['array-bracket-spacing', 'never', { ... }]      → spaced=false + exceptions
//
// rslint's config loader collapses a single trailing option element into the
// option directly, so we accept several shapes for the secondary object as
// well — see utils.GetOptionsMap for the canonical extraction. The string
// element ('always'/'never') still travels as the first array slot.
//
// An option key is an "exception" iff its boolean value equals !spaced
// (upstream `context.options[1][option] === !spaced`). E.g. with spacing
// 'never', `singleValue: true` flips singleElement to spaced behavior; with
// 'always', `singleValue: false` flips it to never-spaced.
func parseOptions(raw any) options {
	opts := options{spaced: false}

	var arr []interface{}
	switch v := raw.(type) {
	case []interface{}:
		arr = v
	case string:
		arr = []interface{}{v}
	case map[string]interface{}:
		// Defensive: spacing-mode string is required as the first element by
		// the upstream schema; if a caller passes just an option object, the
		// spacing mode defaults to 'never' and we still honor the exceptions.
		arr = []interface{}{spacingNever, v}
	}

	if len(arr) > 0 {
		if s, ok := arr[0].(string); ok && s == spacingAlways {
			opts.spaced = true
		}
	}

	flipTo := !opts.spaced
	if len(arr) > 1 {
		if m, ok := arr[1].(map[string]interface{}); ok {
			if b, ok := m["singleValue"].(bool); ok && b == flipTo {
				opts.singleElementException = true
			}
			if b, ok := m["objectsInArrays"].(bool); ok && b == flipTo {
				opts.objectsInArraysException = true
			}
			if b, ok := m["arraysInArrays"].(bool); ok && b == flipTo {
				opts.arraysInArraysException = true
			}
		}
	}
	return opts
}

func isObjectType(node *ast.Node) bool {
	if node == nil {
		return false
	}
	return node.Kind == ast.KindObjectLiteralExpression || node.Kind == ast.KindObjectBindingPattern
}

func isArrayType(node *ast.Node) bool {
	if node == nil {
		return false
	}
	return node.Kind == ast.KindArrayLiteralExpression || node.Kind == ast.KindArrayBindingPattern
}

// nextRealStart returns the position of the first non-trivia character at or
// after `low`. Delegates to tsgo's scanner.SkipTrivia so we get UTF-8-safe
// handling of every trivia form the parser itself recognizes (whitespace,
// line comments, block comments, conflict markers, shebang). `high` exists
// only as a safety clamp — `]` is not trivia, so scanner.SkipTrivia naturally
// stops before crossing it.
func nextRealStart(text string, low, high int) int {
	p := scanner.SkipTrivia(text, low)
	if p > high {
		return high
	}
	return p
}

// prevRealEnd is the reverse-direction counterpart that tsgo doesn't expose:
// from `high` (exclusive) it walks back past trailing whitespace and
// trailing-edge block comments, returning the position one past the last
// real-token character (or `low` if none).
//
// **`low` MUST be the byte position immediately after the last array
// element's end.** That guarantees the bytes in `[low, high)` are pure
// trivia (whitespace, comments, and the optional trailing `,`) — there is
// no string or regex literal content to misinterpret. The reverse scan is
// not token-aware (no string / regex state machine), so without that
// invariant a regex like `[/abc*/]` would have its closing `*/` byte
// sequence falsely classified as a block-comment terminator and the scan
// would over-run into the regex body.
//
// We deliberately don't reverse-scan `//` line comments — that would require
// locating the line start and verifying no string literal straddles it. In
// practice the only effect is that with a trailing `// … \n ]` shape we
// stop at the comment's body text rather than at the token before the
// comment; the caller's `containsNewline` check then sees the intervening
// newline and skips the closing report, matching upstream's
// `isTokenOnSameLine` short-circuit for this shape. See
// jsx-curly-spacing.scanBraceBody for the analogous rationale.
func prevRealEnd(text string, low, high int) int {
	p := high
	for p > low {
		// Fast path: ASCII trivia whitespace.
		if text[p-1] < 0x80 {
			if utils.IsTriviaWhitespaceByte(text[p-1]) {
				p--
				continue
			}
			// Block comment terminator `*/` — reverse-scan to the matching
			// `/*` and continue. See function docstring for why this is
			// safe here (callers guarantee the scan range is pure trivia).
			if p-2 >= low && text[p-2] == '*' && text[p-1] == '/' {
				p -= 2
				for p-2 >= low && (text[p-2] != '/' || text[p-1] != '*') {
					p--
				}
				if p-2 >= low {
					p -= 2
				} else {
					p = low
				}
				continue
			}
			break
		}
		// Slow path: non-ASCII rune — decode and dispatch to the
		// Unicode-aware whitespace table (NBSP, IDEO, LS, PS, BOM, …).
		r, size := utf8.DecodeLastRuneInString(text[:p])
		if size == 0 || r == utf8.RuneError || !utils.IsTriviaWhitespaceRune(r) {
			break
		}
		p -= size
	}
	return p
}

func elementsOf(node *ast.Node) []*ast.Node {
	switch node.Kind {
	case ast.KindArrayLiteralExpression:
		if arr := node.AsArrayLiteralExpression(); arr != nil && arr.Elements != nil {
			return arr.Elements.Nodes
		}
	case ast.KindArrayBindingPattern:
		if bp := node.AsBindingPattern(); bp != nil && bp.Elements != nil {
			return bp.Elements.Nodes
		}
	}
	return nil
}

// bindingElementBound returns the kind that should be tested against the
// objectsInArrays / arraysInArrays exceptions for one element of the array.
// In addition to BindingElement→name unwrapping (for ArrayBindingPattern
// children), it also calls ast.SkipParentheses for ArrayLiteralExpression
// elements: tsgo preserves `( ... )` as ParenthesizedExpression while
// ESLint's TS parser flattens them, so `[({foo: 1})]` reads as
// `ObjectExpression` upstream but as `ParenthesizedExpression` here. Without
// the unwrap, the objectsInArrays / arraysInArrays exceptions would silently
// miss every paren-wrapped first/last element.
func bindingElementBound(node *ast.Node) *ast.Node {
	if node == nil {
		return nil
	}
	if node.Kind == ast.KindBindingElement {
		be := node.AsBindingElement()
		if be == nil {
			return node
		}
		if be.DotDotDotToken != nil {
			// `[ ...x ]` — upstream's RestElement isn't ArrayPattern /
			// ObjectPattern even when the rest target is one, so don't
			// unwrap.
			return node
		}
		name := be.Name()
		if name == nil {
			return node
		}
		return name
	}
	return ast.SkipParentheses(node)
}

// resolveBrackets returns the byte positions of `[` and `]` for `node` and the
// list of its element/binding nodes. Returns ok=false when the trimmed source
// range doesn't actually start/end with brackets (e.g. parser recovery on
// malformed input — defensive, should not happen for valid programs).
func resolveBrackets(text string, sf *ast.SourceFile, node *ast.Node) (openPos, closePos int, elements []*ast.Node, ok bool) {
	trimmed := utils.TrimNodeTextRange(sf, node)
	start := trimmed.Pos()
	end := trimmed.End()
	if start >= end || end > len(text) {
		return 0, 0, nil, false
	}
	if text[start] != '[' || text[end-1] != ']' {
		return 0, 0, nil, false
	}
	return start, end - 1, elementsOf(node), true
}

// exceptionFor returns whether the bracket on `side` (firstElement or
// lastElement) should flip the spaced default. Mirrors upstream's
// `openingBracketMustBeSpaced` / `closingBracketMustBeSpaced` derivation.
func exceptionFor(opts options, element *ast.Node, length int) bool {
	if opts.singleElementException && length == 1 {
		return true
	}
	if element == nil {
		return false
	}
	if opts.objectsInArraysException && isObjectType(element) {
		return true
	}
	if opts.arraysInArraysException && isArrayType(element) {
		return true
	}
	return false
}

// ArrayBracketSpacingRule enforces consistent spacing inside array brackets,
// covering both array literals and array destructuring patterns. Ported from
// @stylistic/eslint-plugin's array-bracket-spacing.
//
// The opening-bracket check runs on the listener's enter call and the
// closing-bracket check runs on exit. For nested arrays this interleaves
// enter/exit calls into source order
// (outer-open → inner-open → inner-close → outer-close), matching how
// ESLint sorts its reports by location.
var ArrayBracketSpacingRule = rule.Rule{
	Name: "@stylistic/array-bracket-spacing",
	Run: func(ctx rule.RuleContext, rawOptions any) rule.RuleListeners {
		opts := parseOptions(rawOptions)
		text := ctx.SourceFile.Text()

		enter := func(node *ast.Node) {
			openPos, closePos, elements, ok := resolveBrackets(text, ctx.SourceFile, node)
			if !ok {
				return
			}
			if opts.spaced && len(elements) == 0 {
				return
			}

			var firstElement *ast.Node
			if len(elements) > 0 {
				firstElement = bindingElementBound(elements[0])
			}
			openingException := exceptionFor(opts, firstElement, len(elements))
			openingBracketMustBeSpaced := opts.spaced
			if openingException {
				openingBracketMustBeSpaced = !opts.spaced
			}

			innerLow := openPos + 1
			innerHigh := closePos

			var secondStart int
			if len(elements) == 0 {
				secondStart = innerHigh
			} else {
				secondStart = nextRealStart(text, innerLow, innerHigh)
			}
			hasSpaceAfterOpen := secondStart > innerLow
			if utils.ContainsLineTerminator(text, innerLow, secondStart) {
				return
			}

			if openingBracketMustBeSpaced && !hasSpaceAfterOpen {
				ctx.ReportRangeWithFixes(
					core.NewTextRange(openPos, openPos+1),
					rule.RuleMessage{
						Id:          "missingSpaceAfter",
						Description: "A space is required after '['.",
						Data:        map[string]string{"tokenValue": "["},
					},
					rule.RuleFix{
						Text:  " ",
						Range: core.NewTextRange(openPos+1, openPos+1),
					},
				)
				return
			}
			if !openingBracketMustBeSpaced && hasSpaceAfterOpen {
				ctx.ReportRangeWithFixes(
					core.NewTextRange(innerLow, secondStart),
					rule.RuleMessage{
						Id:          "unexpectedSpaceAfter",
						Description: "There should be no space after '['.",
						Data:        map[string]string{"tokenValue": "["},
					},
					rule.RuleFix{
						Text:  "",
						Range: core.NewTextRange(innerLow, secondStart),
					},
				)
			}
		}

		exit := func(node *ast.Node) {
			openPos, closePos, elements, ok := resolveBrackets(text, ctx.SourceFile, node)
			if !ok {
				return
			}
			// Empty arrays: upstream's `first === penultimate` short-circuit
			// skips the closing check entirely. (This subsumes the
			// `opts.spaced && len(elements) == 0` early-return in `enter` —
			// closing checks never fire for empty arrays regardless of mode.)
			if len(elements) == 0 {
				return
			}

			rawLast := elements[len(elements)-1]
			lastElement := bindingElementBound(rawLast)
			closingException := exceptionFor(opts, lastElement, len(elements))
			closingBracketMustBeSpaced := opts.spaced
			if closingException {
				closingBracketMustBeSpaced = !opts.spaced
			}

			innerHigh := closePos
			// Lower-bound the reverse trivia scan at the last element's end
			// position. Everything in `[rawLast.End(), closePos)` is pure
			// trivia (optionally with a trailing `,`), so the non-token-aware
			// reverse byte scan can't be tricked by regex / string literal
			// suffixes like `*/`. Use `rawLast` (not `lastElement`) — the
			// latter is the post-`SkipParentheses` / `BindingElement.Name()`
			// unwrap used purely for the Array/Object classification, and
			// its End() would be earlier than the actual element's End().
			scanLow := rawLast.End()
			if scanLow < openPos+1 {
				scanLow = openPos + 1
			}

			penultimateEnd := prevRealEnd(text, scanLow, innerHigh)
			hasSpaceBeforeClose := penultimateEnd < innerHigh
			if utils.ContainsLineTerminator(text, penultimateEnd, innerHigh) {
				return
			}

			if closingBracketMustBeSpaced && !hasSpaceBeforeClose {
				ctx.ReportRangeWithFixes(
					core.NewTextRange(closePos, closePos+1),
					rule.RuleMessage{
						Id:          "missingSpaceBefore",
						Description: "A space is required before ']'.",
						Data:        map[string]string{"tokenValue": "]"},
					},
					rule.RuleFix{
						Text:  " ",
						Range: core.NewTextRange(closePos, closePos),
					},
				)
				return
			}
			if !closingBracketMustBeSpaced && hasSpaceBeforeClose {
				ctx.ReportRangeWithFixes(
					core.NewTextRange(penultimateEnd, closePos),
					rule.RuleMessage{
						Id:          "unexpectedSpaceBefore",
						Description: "There should be no space before ']'.",
						Data:        map[string]string{"tokenValue": "]"},
					},
					rule.RuleFix{
						Text:  "",
						Range: core.NewTextRange(penultimateEnd, closePos),
					},
				)
			}
		}

		return rule.RuleListeners{
			ast.KindArrayLiteralExpression:                      enter,
			ast.KindArrayBindingPattern:                         enter,
			rule.ListenerOnExit(ast.KindArrayLiteralExpression): exit,
			rule.ListenerOnExit(ast.KindArrayBindingPattern):    exit,
		}
	},
}
