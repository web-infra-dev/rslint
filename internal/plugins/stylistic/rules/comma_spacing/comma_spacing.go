package comma_spacing

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/plugins/stylistic/stylisticutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

type commaSpacingOptions struct {
	before bool
	after  bool
}

// parseOptions accepts upstream's option shapes. The default is
// `{ before: false, after: true }`. The CLI / rule_tester may deliver:
//
//	rule: ['comma-spacing']                                  → defaults
//	rule: ['comma-spacing', { before: true, after: false }]  → as written
//
// rslint's config loader collapses a single-element option array into the
// option itself, so we accept either a bare map or a wrapping array. The two
// flags are independent — partial-config payloads (only one of `before`/
// `after`) keep the unspecified flag at its default.
func parseOptions(raw any) commaSpacingOptions {
	opts := commaSpacingOptions{before: false, after: true}
	optsMap := utils.GetOptionsMap(raw)
	if optsMap == nil {
		return opts
	}
	if b, ok := optsMap["before"].(bool); ok {
		opts.before = b
	}
	if b, ok := optsMap["after"].(bool); ok {
		opts.after = b
	}
	return opts
}

// tokenInfo is a minimal token record used by the rule's token walker. The
// scanner can emit ordinary tokens AND trivia (whitespace, line breaks,
// comments) when SkipTrivia is off. We retain comments because upstream's
// `tokensAndComments` stream feeds comments into the prev/next-token slots —
// e.g. `, /* */ foo` puts the comment as the "next token" of `,`.
type tokenInfo struct {
	pos  int
	end  int
	kind ast.Kind
}

// isCloseBracketKind tests the three "right-side punctuation" tokens that
// upstream defers to their own dedicated rules (space-in-parens,
// array-bracket-spacing, object-curly-spacing). When the next token after a
// comma is one of these, the after-check is skipped entirely.
func isCloseBracketKind(k ast.Kind) bool {
	return k == ast.KindCloseParenToken ||
		k == ast.KindCloseBracketToken ||
		k == ast.KindCloseBraceToken
}

// isTriviaSkip reports whether a token kind should be invisible when walking
// for the "prev / next" of a comma. Whitespace and line breaks are skipped;
// comments are NOT — they participate in spacing checks (upstream merges
// `tokens` and `comments` into one stream).
func isTriviaSkip(k ast.Kind) bool {
	return k == ast.KindWhitespaceTrivia ||
		k == ast.KindNewLineTrivia ||
		k == ast.KindConflictMarkerTrivia ||
		k == ast.KindNonTextFileMarkerTrivia
}

// nextRealTokenKind reads the very next non-trivia, non-comment token at or
// after `from`, using the supplied (already-configured) scanner. Mirrors
// ESLint's `sourceCode.getTokenAfter(prev)` with default options — comments
// are skipped. Used by both the null-array-element walk (where the previous-
// token chain in upstream's `addNullElementsToIgnoreList` uses default
// getTokenAfter, so comments don't count) and the type-parameter-trailing-
// comma walk (same).
//
// The scanner is passed in (and reused across calls) rather than freshly
// allocated via `scanner.GetScannerForSourceFile`: in a single file each
// array literal / pattern + each type-parameter list issues 1–N calls, so
// allocating a Scanner per call wastes work proportional to AST size.
// Callers configure `SetText` / `SetLanguageVariant` once and call
// `ResetTokenState` per query (this function does the reset and scan).
func nextRealTokenKind(s *scanner.Scanner, from int) (ast.Kind, int) {
	s.ResetTokenState(from)
	for {
		k := s.Scan()
		if k == ast.KindEndOfFile {
			return k, s.TokenStart()
		}
		if isTriviaSkip(k) ||
			k == ast.KindSingleLineCommentTrivia ||
			k == ast.KindMultiLineCommentTrivia {
			continue
		}
		return k, s.TokenStart()
	}
}

// collectIgnoredCommas walks the AST and records the source positions of
// commas whose spacing must NOT be checked. Upstream's
// `ignoredTokens` set captures two distinct populations; both feed the same
// suppression code path (`isCommaToken(prev) || ignoredTokens.has(comma)`
// → both prev and next nulled out, so the validator emits nothing).
//
//   - **Null-element commas inside ArrayLiteralExpression /
//     ArrayBindingPattern**. For each `KindOmittedExpression` element the
//     ignored comma is the one immediately AFTER the previous token (the
//     `[` or the prior `,`). We model this by tracking `prevEnd` (start
//     position to scan from) and, when the current element is omitted,
//     scanning forward to the next non-trivia token, asserting it's a
//     comma, and recording it.
//   - **Trailing commas in TS type-parameter lists**. After the last type
//     parameter, if the next non-trivia token is a comma, that comma is
//     ignored. This covers the `<T,>`, `<T, P,>`, `<T, T1,>` shapes — and
//     deliberately does NOT cover the no-trailing-comma `<T, P>` form,
//     where the next token is `>`.
//
// Both populations are recorded into a single position-keyed map; downstream
// just asks `ignored[pos]` once per comma.
func collectIgnoredCommas(sf *ast.SourceFile) map[int]struct{} {
	ignored := map[int]struct{}{}

	// Single reusable scanner for all nextRealTokenKind lookups in this
	// pass. Configured with SkipTrivia=false so we can loop manually past
	// whitespace, line breaks, and comments — matching ESLint's default
	// getTokenAfter (comments excluded).
	s := scanner.NewScanner()
	s.SetText(sf.Text())
	s.SetLanguageVariant(sf.LanguageVariant)
	s.SetSkipTrivia(false)

	// isHoleElement reports whether an ArrayLiteral / ArrayBindingPattern
	// element represents an elision (a "missing" slot like `[ , x]`).
	//
	//   - In ArrayLiteralExpression tsgo materializes the hole as a
	//     `KindOmittedExpression` node.
	//   - In ArrayBindingPattern tsgo materializes the hole as a
	//     `KindBindingElement` whose entire span is zero-width (Pos ==
	//     End). The element has no name and no initializer; it's
	//     parser-emitted to keep the elements slice index-aligned with
	//     the comma-separated positions.
	//
	// Detecting both shapes via the same predicate lets one walker
	// handle ArrayLiteral and ArrayBindingPattern uniformly.
	isHoleElement := func(el *ast.Node) bool {
		if el == nil {
			return false
		}
		if el.Kind == ast.KindOmittedExpression {
			return true
		}
		if el.Kind == ast.KindBindingElement && el.Pos() == el.End() {
			return true
		}
		return false
	}

	addArrayNullCommas := func(node *ast.Node, elements []*ast.Node) {
		if len(elements) == 0 {
			return
		}
		// Mirror upstream's `addNullElementsToIgnoreList`:
		//   previousToken := getFirstToken(node)  // `[`
		//   for el in elements:
		//     if el is hole:
		//       token := getTokenAfter(previousToken)  // expect comma; ignore it
		//     else:
		//       token := getTokenAfter(el)             // step over the separator comma
		//     previousToken := token
		//
		// `scanFrom` corresponds to "scan starting position to find the
		// next real token after `previousToken`" — i.e. it's one past
		// `previousToken.End()`. We seed it with the `[`'s position
		// (trimmed.Pos()+1 = just past `[`) so the first `getTokenAfter`
		// reads the token that follows `[`.
		trimmed := utils.TrimNodeTextRange(sf, node)
		scanFrom := trimmed.Pos() + 1 // just past `[`
		for _, el := range elements {
			if isHoleElement(el) {
				k, pos := nextRealTokenKind(s, scanFrom)
				if k != ast.KindCommaToken {
					// Defensive: malformed input — abandon this list.
					return
				}
				ignored[pos] = struct{}{}
				scanFrom = pos + 1
				continue
			}
			if el == nil {
				continue
			}
			// Step previousToken over the comma that follows the
			// element (if any) so the next iteration sees the right
			// "next token". For the last element this lands on `]`,
			// which we ignore.
			k, pos := nextRealTokenKind(s, el.End())
			if k == ast.KindCommaToken {
				scanFrom = pos + 1
			} else {
				scanFrom = el.End()
			}
		}
	}

	addTypeParamTrailing := func(list *ast.NodeList) {
		if list == nil || len(list.Nodes) == 0 {
			return
		}
		last := list.Nodes[len(list.Nodes)-1]
		k, pos := nextRealTokenKind(s, last.End())
		if k == ast.KindCommaToken {
			ignored[pos] = struct{}{}
		}
	}

	// Any future FunctionLike kind picks up TypeParameterList support
	// automatically via the shim's IsFunctionLike predicate; the explicit
	// switch only covers the non-function-like type-bearing kinds.
	// `KindJSTypeAliasDeclaration` covers JSDoc `@typedef`-style aliases in
	// `.js` files (it routes through the same `AsTypeAliasDeclaration` getter
	// as `KindTypeAliasDeclaration`, so we treat it the same).
	// `KindJSDocTemplateTag` is intentionally excluded — its commas live
	// inside `/* */` block comments and our gap-walker never enters comment
	// content, so they cannot reach this check.
	hasTypeParamList := func(n *ast.Node) bool {
		if ast.IsFunctionLike(n) {
			return true
		}
		switch n.Kind {
		case ast.KindClassDeclaration, ast.KindClassExpression,
			ast.KindInterfaceDeclaration,
			ast.KindTypeAliasDeclaration, ast.KindJSTypeAliasDeclaration:
			return true
		}
		return false
	}

	var walk func(n *ast.Node)
	walk = func(n *ast.Node) {
		if n == nil {
			return
		}
		switch n.Kind {
		case ast.KindArrayLiteralExpression:
			if arr := n.AsArrayLiteralExpression(); arr != nil && arr.Elements != nil {
				addArrayNullCommas(n, arr.Elements.Nodes)
			}
		case ast.KindArrayBindingPattern:
			if bp := n.AsBindingPattern(); bp != nil && bp.Elements != nil {
				addArrayNullCommas(n, bp.Elements.Nodes)
			}
		}
		if hasTypeParamList(n) {
			addTypeParamTrailing(n.TypeParameterList())
		}
		n.ForEachChild(func(c *ast.Node) bool {
			walk(c)
			return false
		})
	}
	walk(sf.AsNode())

	return ignored
}

// scanAllTokens enumerates every visible token (real tokens AND comments,
// but not whitespace / line breaks) in source order, in a way that is
// robust to template literals with substitutions, regex literals, JSX
// text, and string literals.
//
// **Why the gap-scan recursion instead of one flat full-file scan?**
//
// tsgo's raw scanner does not maintain template-mode state across calls.
// Inside “ `a${e}\\\`(${typeof e})\` “, after scanning the `}` that
// closes `${e}`, the scanner re-enters default mode; the next “ ` “
// (whether escaped-literal or real-closing) is treated as the START of a
// brand-new template literal, and a long stretch of subsequent code is
// silently swallowed as that fake template's content. Commas in that
// stretch never reach our walker, producing false negatives. (See the
// rspack p-map.mjs differential — under a full-file raw scan rslint
// missed 53 commas that ESLint reports.)
//
// `internal/utils/ts_eslint.go HasSameTokens` solves the same problem
// with a recursive AST walk: scan only the **gaps between AST children**.
// AST nodes for templates (TemplateExpression / TemplateSpan /
// TemplateHead / TemplateMiddle / TemplateTail / NoSubstitutionTemplate-
// Literal) cover the literal text contiguously, so a gap between two
// children is **always pure trivia + punctuation** — never literal-text
// content. The scanner therefore cannot enter the confusable region.
//
// Within each gap we open a scanner with `SkipTrivia=false` (via fresh
// `NewScanner` rather than `GetScannerForSourceFile`, which auto-calls
// `Scan()` with `SkipTrivia=true` and would eat the first comment) so
// block / line comments surface as tokens — upstream's
// `tokensAndComments` stream includes comments as legitimate
// prev / next neighbors of a comma.
//
// **Leaf nodes** (any node with no `ForEachChild` children — identifiers,
// literals, keyword tokens, TemplateHead/Middle/Tail, OmittedExpression
// with zero width, BinaryExpression's OperatorToken when it is a
// `KindCommaToken`, etc.) are not gap-scanned; their entire range is the
// token, captured directly via `utils.TrimNodeTextRange`. This is how a
// CommaToken that lives as an operator child of a BinaryExpression (the
// sequence-expression case `a, b, c`) joins the stream — gaps around it
// are empty, but the leaf itself emits as `KindCommaToken`.
func scanAllTokens(sf *ast.SourceFile) []tokenInfo {
	var tokens []tokenInfo
	text := sf.Text()
	textLen := len(text)

	// Single reusable scanner, reset per gap via ResetTokenState.
	// SkipTrivia=false so block / line comments surface as tokens.
	s := scanner.NewScanner()
	s.SetText(text)
	s.SetLanguageVariant(sf.LanguageVariant)
	s.SetSkipTrivia(false)

	scanGap := func(start, end int) {
		if start >= end || start < 0 || end > textLen {
			return
		}
		s.ResetTokenState(start)
		for {
			k := s.Scan()
			if k == ast.KindEndOfFile {
				return
			}
			// Bound to the gap on both sides: tokens that start at /
			// past `end` belong to the next sibling, and tokens whose
			// end straddles `end` (defensive — shouldn't happen for
			// AST-sibling-aligned gaps, but matches HasSameTokens'
			// `liveA` invariant for safety) are also excluded.
			if s.TokenStart() >= end || s.TokenEnd() > end {
				return
			}
			if isTriviaSkip(k) {
				continue
			}
			tokens = append(tokens, tokenInfo{
				pos:  s.TokenStart(),
				end:  s.TokenEnd(),
				kind: k,
			})
		}
	}

	// `walk` recurses via `ForEachChild` directly — no per-node `[]*ast.Node`
	// allocation. `prevEnd` and `hasKids` live in the closure variables so
	// each non-leaf node uses O(1) extra memory regardless of arity. Matches
	// the pattern other recursive walkers in this repo follow
	// (e.g. internal/linter/linter.go's childVisitor,
	// internal/plugins/react_hooks/.../exhaustive_deps.go's visit).
	var walk func(n *ast.Node)
	walk = func(n *ast.Node) {
		if n == nil {
			return
		}
		prevEnd := n.Pos()
		hasKids := false
		n.ForEachChild(func(kid *ast.Node) bool {
			hasKids = true
			// Synthetic / parser-recovered children (e.g. an implicit
			// QuestionDotToken slot on a non-optional CallExpression)
			// have negative Pos. Skip them — they aren't backed by
			// source bytes and would crash TrimNodeTextRange. Keep
			// `hasKids` true so the parent is still treated as
			// composite (we already entered the iteration).
			if kid.Pos() < 0 || kid.End() < 0 {
				return false
			}
			// Use the trimmed Pos of each kid (Pos() includes leading
			// trivia, so content like comments that visually sit
			// *between* two AST siblings is technically "inside" the
			// next sibling's Pos range). Scanning up to the trimmed
			// Pos pulls those comments back into the gap so the
			// scanner sees them.
			kidTrimmedPos := utils.TrimNodeTextRange(sf, kid).Pos()
			scanGap(prevEnd, kidTrimmedPos)
			walk(kid)
			prevEnd = kid.End()
			return false
		})
		if !hasKids {
			// Leaf node — emit as a single token at its trimmed range
			// (Pos() includes leading trivia we don't want).
			// Zero-width leaves (OmittedExpression, empty BindingElement
			// holes) are skipped — they don't represent visible tokens
			// in the stream, and including them as zero-width entries
			// would confuse "is comma adjacent to comma" prev/next
			// logic. Synthetic nodes with negative Pos (parser
			// recovery slots, missing-optional placeholders) are also
			// skipped — they have no source-text backing.
			if n.Pos() < 0 || n.End() < 0 {
				return
			}
			trimmed := utils.TrimNodeTextRange(sf, n)
			if trimmed.Pos() < trimmed.End() {
				tokens = append(tokens, tokenInfo{
					pos:  trimmed.Pos(),
					end:  trimmed.End(),
					kind: n.Kind,
				})
			}
			return
		}
		scanGap(prevEnd, n.End())
	}
	walk(sf.AsNode())
	return tokens
}

// CommaSpacingRule enforces consistent spacing before and after commas.
// Ported from @stylistic/eslint-plugin's comma-spacing.
//
// Strategy: build a single global token stream (with comments retained as
// tokens, matching ESLint's `sourceCode.tokensAndComments`); for each
// `KindCommaToken` in the stream determine its prev/next visible token,
// apply upstream's null-prev / null-next rules (commas next to commas and
// commas in the AST-derived ignore set both null out their neighbors), and
// then run `validateCommaSpacing`. The work happens eagerly inside `Run`
// because the rule's scope is the whole file — there's no per-node listener
// that naturally fits this shape, and the linter never fires a
// `KindSourceFile` listener.
var CommaSpacingRule = rule.Rule{
	Name: "@stylistic/comma-spacing",
	Run: func(ctx rule.RuleContext, rawOptions any) rule.RuleListeners {
		opts := parseOptions(rawOptions)
		sf := ctx.SourceFile
		if sf == nil {
			return rule.RuleListeners{}
		}

		ignoredCommas := collectIgnoredCommas(sf)
		tokens := scanAllTokens(sf)

		for i := range tokens {
			t := tokens[i]
			if t.kind != ast.KindCommaToken {
				continue
			}

			var prev, next *tokenInfo
			if i > 0 {
				prev = &tokens[i-1]
			}
			if i+1 < len(tokens) {
				next = &tokens[i+1]
			}

			// Upstream:
			//
			//   isCommaToken(prevToken) || ignoredTokens.has(token) ? null : prevToken
			//   isCommaToken(nextToken) || ignoredTokens.has(token) ? null : nextToken
			//
			// Comma-next-to-comma neutralizes EACH side independently;
			// the ignore set neutralizes BOTH at once.
			if prev != nil && prev.kind == ast.KindCommaToken {
				prev = nil
			}
			if next != nil && next.kind == ast.KindCommaToken {
				next = nil
			}
			if _, ok := ignoredCommas[t.pos]; ok {
				prev = nil
				next = nil
			}

			validateCommaSpacing(ctx, sf, opts, t, prev, next)
		}

		return rule.RuleListeners{}
	},
}

// validateCommaSpacing emits up to two diagnostics for one comma, one for
// the before-side and one for the after-side. Mirrors upstream's
// validateCommaSpacing function byte-for-byte — the two if-blocks below
// translate the two upstream blocks one for one. See PORT_RULE.md "Testing
// Philosophy" for why we mirror logic shape, not just outputs.
func validateCommaSpacing(
	ctx rule.RuleContext,
	sf *ast.SourceFile,
	opts commaSpacingOptions,
	comma tokenInfo,
	prev, next *tokenInfo,
) {
	// --- before-side ---
	//
	// Upstream:
	//   if (prevToken && isTokenOnSameLine(prevToken, commaToken)
	//       && spaceBefore !== sourceCode.isSpaceBetween(prevToken, commaToken)) { ... }
	if prev != nil && stylisticutil.SameLineByPos(sf, prev.end, comma.pos) {
		hasSpace := prev.end < comma.pos
		if opts.before != hasSpace {
			var msg rule.RuleMessage
			var fix rule.RuleFix
			if opts.before {
				msg = rule.RuleMessage{
					Id:          "missing",
					Description: "A space is required before ','.",
					Data:        map[string]string{"loc": "before"},
				}
				fix = rule.RuleFix{
					Text:  " ",
					Range: core.NewTextRange(comma.pos, comma.pos),
				}
			} else {
				msg = rule.RuleMessage{
					Id:          "unexpected",
					Description: "There should be no space before ','.",
					Data:        map[string]string{"loc": "before"},
				}
				fix = rule.RuleFix{
					Text:  "",
					Range: core.NewTextRange(prev.end, comma.pos),
				}
			}
			ctx.ReportRangeWithFixes(
				core.NewTextRange(comma.pos, comma.end),
				msg,
				fix,
			)
		}
	}

	// --- after-side ---
	//
	// Upstream:
	//   if (nextToken && isTokenOnSameLine(commaToken, nextToken)
	//       && !isClosingParenToken(nextToken)
	//       && !isClosingBracketToken(nextToken)
	//       && !isClosingBraceToken(nextToken)
	//       && !(!spaceAfter && nextToken.type === AST_TOKEN_TYPES.Line)
	//       && spaceAfter !== sourceCode.isSpaceBetween(commaToken, nextToken)) { ... }
	if next == nil {
		return
	}
	if !stylisticutil.SameLineByPos(sf, comma.end, next.pos) {
		return
	}
	if isCloseBracketKind(next.kind) {
		return
	}
	// `!spaceAfter && nextToken is a single-line comment` exemption.
	// Deleting the space would fuse `,` into the `//` and silently re-shape
	// the source.
	if !opts.after && next.kind == ast.KindSingleLineCommentTrivia {
		return
	}
	hasSpace := comma.end < next.pos
	if opts.after == hasSpace {
		return
	}
	var msg rule.RuleMessage
	var fix rule.RuleFix
	if opts.after {
		msg = rule.RuleMessage{
			Id:          "missing",
			Description: "A space is required after ','.",
			Data:        map[string]string{"loc": "after"},
		}
		fix = rule.RuleFix{
			Text:  " ",
			Range: core.NewTextRange(comma.end, comma.end),
		}
	} else {
		msg = rule.RuleMessage{
			Id:          "unexpected",
			Description: "There should be no space after ','.",
			Data:        map[string]string{"loc": "after"},
		}
		fix = rule.RuleFix{
			Text:  "",
			Range: core.NewTextRange(comma.end, next.pos),
		}
	}
	ctx.ReportRangeWithFixes(
		core.NewTextRange(comma.pos, comma.end),
		msg,
		fix,
	)
}
