package arrow_parens

import (
	"strings"
	"unicode/utf8"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

const (
	msgExpectedParens         = "expectedParens"
	msgUnexpectedParens       = "unexpectedParens"
	msgExpectedParensBlock    = "expectedParensBlock"
	msgUnexpectedParensInline = "unexpectedParensInline"

	descExpectedParens         = "Expected parentheses around arrow function argument."
	descUnexpectedParens       = "Unexpected parentheses around single function argument."
	descExpectedParensBlock    = "Expected parentheses around arrow function argument having a body with curly braces."
	descUnexpectedParensInline = "Unexpected parentheses around single function argument having a body with no curly braces."
)

type options struct {
	asNeeded            bool
	requireForBlockBody bool
}

// parseOptions mirrors upstream's option layout.
//
//	rule: ['arrow-parens']                                            → asNeeded=false  (default 'always')
//	rule: ['arrow-parens', 'always']                                  → asNeeded=false
//	rule: ['arrow-parens', 'as-needed']                               → asNeeded=true
//	rule: ['arrow-parens', 'as-needed', { requireForBlockBody: bool }] → as above + flag
//
// rslint's config loader collapses single-element options arrays; GetOptionsMap
// is not used here because the rule's primary option is a string, not an
// object, so the standard helper would discard the string.
func parseOptions(raw any) options {
	opts := options{}

	var arr []interface{}
	switch v := raw.(type) {
	case []interface{}:
		arr = v
	case string:
		arr = []interface{}{v}
	}

	if len(arr) > 0 {
		if s, ok := arr[0].(string); ok && s == "as-needed" {
			opts.asNeeded = true
		}
	}
	if opts.asNeeded && len(arr) > 1 {
		if m, ok := arr[1].(map[string]interface{}); ok {
			if b, ok := m["requireForBlockBody"].(bool); ok {
				opts.requireForBlockBody = b
			}
		}
	}
	return opts
}

func hasBlockBody(arrowFn *ast.ArrowFunction) bool {
	return arrowFn.Body != nil && arrowFn.Body.Kind == ast.KindBlock
}

// findOpenParenPos returns the byte position of the opening `(` of the arrow
// function's parameters, or -1 if the arrow function has no parens (single
// identifier param like `a => a`).
//
// We rely on the parser's invariant that for parens-wrapped parameter lists,
// the ParameterList's Pos starts immediately after `(`. For the simple-arrow
// form (`a => b`, with or without `async`), the parser sets the list's Loc to
// the parameter's own Loc, which never crosses an outer call's `(`. Comparing
// `parameters.Pos() - 1 >= node.Pos()` rejects the outer-call false positive
// (e.g. `foo(a => b)` — the `(` at `node.Pos() - 1` belongs to `foo`, not the
// arrow).
func findOpenParenPos(node *ast.Node, arrowFn *ast.ArrowFunction, text string) int {
	if arrowFn == nil || arrowFn.Parameters == nil {
		return -1
	}
	parenPos := arrowFn.Parameters.Pos() - 1
	if parenPos < node.Pos() || parenPos < 0 || parenPos >= len(text) {
		return -1
	}
	if text[parenPos] != '(' {
		return -1
	}
	return parenPos
}

// findCloseParenPos returns the byte position of the closing `)` after the
// single parameter, accounting for an optional trailing comma. Returns -1 on
// a parser-recovery shape that doesn't actually contain a `)` (defensive).
//
// Uses the tsgo scanner instead of byte-walking so we get Unicode whitespace
// handling and proper tokenization for free — the same recipe `no_array_delete`
// uses to locate `]`/`[` siblings via `scanner.GetRangeOfTokenAtPosition`.
func findCloseParenPos(sf *ast.SourceFile, paramEnd int) int {
	tok := scanner.ScanTokenAtPosition(sf, paramEnd)
	rng := scanner.GetRangeOfTokenAtPosition(sf, paramEnd)
	if tok == ast.KindCommaToken {
		tok = scanner.ScanTokenAtPosition(sf, rng.End())
		rng = scanner.GetRangeOfTokenAtPosition(sf, rng.End())
	}
	if tok == ast.KindCloseParenToken {
		return rng.Pos()
	}
	return -1
}

// hasCommentsBetween reports whether the byte range [low, high) contains any
// `//` or `/*` sequence. Between an arrow function's parens the only legal
// content is identifier/comma trivia plus the param itself — no strings or
// regexes can appear, so a plain substring search is safe.
func hasCommentsBetween(text string, low, high int) bool {
	if low < 0 {
		low = 0
	}
	if high > len(text) {
		high = len(text)
	}
	if low >= high {
		return false
	}
	s := text[low:high]
	return strings.Contains(s, "//") || strings.Contains(s, "/*")
}

// needsSpaceBeforeOpenParen reports whether the rune immediately before
// `openParenPos` is part of an identifier or keyword token — in which case
// removing the `(` would fuse two adjacent tokens into a single identifier
// (`async(a)=>a` → without a space would become an undefined identifier).
// Equivalent to upstream's `canTokensBeAdjacent` guard for the realistic
// keyword-prefix shapes, but Unicode-aware via tsgo's scanner.
func needsSpaceBeforeOpenParen(text string, openParenPos int) bool {
	if openParenPos <= 0 {
		return false
	}
	r, size := utf8.DecodeLastRuneInString(text[:openParenPos])
	if size == 0 || r == utf8.RuneError {
		return false
	}
	return scanner.IsIdentifierPart(r)
}

// ArrowParensRule enforces consistent use of parentheses around arrow function
// parameters. Ported from @stylistic/eslint-plugin's arrow-parens.
//
// The rule only fires when an arrow function has exactly one parameter — both
// `() => {}` and `(a, b) => {}` are out of scope. For single-parameter
// arrows, the rule either:
//
//   - in `'always'` mode (default), requires `(a) => ...`
//   - in `'as-needed'` mode, drops parens for plain `Identifier` params with
//     no annotation, default, optional marker, rest, generics, return type,
//     or interior comments
//   - with `requireForBlockBody: true`, restores the requirement when the
//     arrow's body is a `{ ... }` block.
var ArrowParensRule = rule.Rule{
	Name: "@stylistic/arrow-parens",
	Run: func(ctx rule.RuleContext, rawOptions any) rule.RuleListeners {
		opts := parseOptions(rawOptions)
		text := ctx.SourceFile.Text()

		check := func(node *ast.Node) {
			arrowFn := node.AsArrowFunction()
			if arrowFn == nil || arrowFn.Parameters == nil {
				return
			}
			params := arrowFn.Parameters.Nodes
			if len(params) != 1 {
				return
			}
			param := params[0]
			if param == nil || param.Kind != ast.KindParameter {
				return
			}

			shouldHaveParens := !opts.asNeeded || (opts.requireForBlockBody && hasBlockBody(arrowFn))
			openParenPos := findOpenParenPos(node, arrowFn, text)
			hasParens := openParenPos >= 0

			paramRange := utils.TrimNodeTextRange(ctx.SourceFile, param)
			paramStart := paramRange.Pos()
			paramEnd := paramRange.End()

			if shouldHaveParens && !hasParens {
				msgId := msgExpectedParens
				desc := descExpectedParens
				if opts.requireForBlockBody {
					msgId = msgExpectedParensBlock
					desc = descExpectedParensBlock
				}
				ctx.ReportRangeWithFixes(
					paramRange,
					rule.RuleMessage{Id: msgId, Description: desc},
					rule.RuleFix{Text: "(", Range: core.NewTextRange(paramStart, paramStart)},
					rule.RuleFix{Text: ")", Range: core.NewTextRange(paramEnd, paramEnd)},
				)
				return
			}

			if shouldHaveParens || !hasParens {
				return
			}

			// As-needed branch: parens exist but may be removable.
			paramDecl := param.AsParameterDeclaration()
			if paramDecl == nil {
				return
			}
			name := paramDecl.Name()
			if name == nil || name.Kind != ast.KindIdentifier {
				return
			}
			if paramDecl.DotDotDotToken != nil ||
				paramDecl.Initializer != nil ||
				paramDecl.QuestionToken != nil ||
				paramDecl.Type != nil {
				return
			}
			if arrowFn.Type != nil {
				return
			}
			// Generics force parens: `<T>(a) => b`. In tsgo TypeParameters are a
			// dedicated field, so we don't need to walk tokens — the presence
			// of the list is sufficient (upstream relies on `getFirstToken`
			// after skipping `async`).
			if arrowFn.TypeParameters != nil && len(arrowFn.TypeParameters.Nodes) > 0 {
				return
			}

			closeParenPos := findCloseParenPos(ctx.SourceFile, paramEnd)
			if closeParenPos < 0 {
				return
			}
			if hasCommentsBetween(text, openParenPos+1, closeParenPos) {
				return
			}

			fixes := []rule.RuleFix{}
			// Guard against token fusion: if the rune immediately before `(`
			// is part of an identifier / keyword (no whitespace separates them),
			// removing the paren would glue the keyword to the param name —
			// e.g. removing `async(a)=>a`'s parens without inserting a space
			// would join the `async` and `a` tokens into a single identifier.
			if needsSpaceBeforeOpenParen(text, openParenPos) {
				fixes = append(fixes, rule.RuleFix{
					Text:  " ",
					Range: core.NewTextRange(openParenPos, openParenPos),
				})
			}
			fixes = append(fixes,
				rule.RuleFix{Text: "", Range: core.NewTextRange(openParenPos, paramStart)},
				rule.RuleFix{Text: "", Range: core.NewTextRange(paramEnd, closeParenPos+1)},
			)

			msgId := msgUnexpectedParens
			desc := descUnexpectedParens
			if opts.requireForBlockBody {
				msgId = msgUnexpectedParensInline
				desc = descUnexpectedParensInline
			}
			ctx.ReportRangeWithFixes(
				paramRange,
				rule.RuleMessage{Id: msgId, Description: desc},
				fixes...,
			)
		}

		return rule.RuleListeners{
			ast.KindArrowFunction: check,
		}
	},
}
