package comma_style

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/plugins/stylistic/stylisticutil"
	"github.com/web-infra-dev/rslint/internal/rule"
)

const (
	styleFirst = "first"
	styleLast  = "last"

	msgUnexpectedLineBeforeAndAfterComma = "unexpectedLineBeforeAndAfterComma"
	msgExpectedCommaFirst                = "expectedCommaFirst"
	msgExpectedCommaLast                 = "expectedCommaLast"

	descUnexpectedLineBeforeAndAfterComma = "Bad line breaking before and after ','."
	descExpectedCommaFirst                = "',' should be placed first."
	descExpectedCommaLast                 = "',' should be placed last."

	fixStyleBetween = "between"
)

// Upstream exception keys (ESTree node-type names). Each key gates one of the
// listener fan-outs below. We match upstream verbatim so user-facing option
// shapes are byte-for-byte identical.
const (
	excVariableDeclaration             = "VariableDeclaration"
	excObjectExpression                = "ObjectExpression"
	excObjectPattern                   = "ObjectPattern"
	excArrayExpression                 = "ArrayExpression"
	excArrayPattern                    = "ArrayPattern"
	excFunctionDeclaration             = "FunctionDeclaration"
	excFunctionExpression              = "FunctionExpression"
	excArrowFunctionExpression         = "ArrowFunctionExpression"
	excCallExpression                  = "CallExpression"
	excImportDeclaration               = "ImportDeclaration"
	excNewExpression                   = "NewExpression"
	excExportAllDeclaration            = "ExportAllDeclaration"
	excExportNamedDeclaration          = "ExportNamedDeclaration"
	excImportExpression                = "ImportExpression"
	excSequenceExpression              = "SequenceExpression"
	excClassDeclaration                = "ClassDeclaration"
	excClassExpression                 = "ClassExpression"
	excTSDeclareFunction               = "TSDeclareFunction"
	excTSFunctionType                  = "TSFunctionType"
	excTSConstructorType               = "TSConstructorType"
	excTSEmptyBodyFunctionExpression   = "TSEmptyBodyFunctionExpression"
	excTSEnumBody                      = "TSEnumBody"
	excTSTypeLiteral                   = "TSTypeLiteral"
	excTSIndexSignature                = "TSIndexSignature"
	excTSMethodSignature               = "TSMethodSignature"
	excTSCallSignatureDeclaration      = "TSCallSignatureDeclaration"
	excTSConstructSignatureDeclaration = "TSConstructSignatureDeclaration"
	excTSInterfaceBody                 = "TSInterfaceBody"
	excTSInterfaceDeclaration          = "TSInterfaceDeclaration"
	excTSTupleType                     = "TSTupleType"
	excTSTypeParameterDeclaration      = "TSTypeParameterDeclaration"
	excTSTypeParameterInstantiation    = "TSTypeParameterInstantiation"
)

type options struct {
	style      string
	exceptions map[string]bool
}

// parseOptions mirrors upstream's schema:
//
//	rule: ['comma-style']                                    → 'last', no exceptions
//	rule: ['comma-style', 'first' | 'last']                  → style, no exceptions
//	rule: ['comma-style', <style>, { exceptions: { … } }]    → style + exception map
//
// rslint's config loader collapses a single trailing option element into the
// option directly, so the first slot might travel as a bare string. We accept
// both shapes here.
func parseOptions(raw any) options {
	opts := options{style: styleLast}

	var arr []any
	switch v := raw.(type) {
	case []any:
		arr = v
	case string:
		arr = []any{v}
	}

	if len(arr) > 0 {
		if s, ok := arr[0].(string); ok {
			switch s {
			case styleFirst, styleLast:
				opts.style = s
			}
		}
	}
	if len(arr) > 1 {
		if m, ok := arr[1].(map[string]any); ok {
			if e, ok := m["exceptions"].(map[string]any); ok {
				exc := make(map[string]bool, len(e))
				for k, val := range e {
					if b, ok := val.(bool); ok {
						exc[k] = b
					}
				}
				opts.exceptions = exc
			}
		}
	}
	return opts
}

// tk is a single token captured by the forward scan over a list range.
// Stored as a flat slice so per-comma neighbor lookups are O(1).
type tk struct {
	kind ast.Kind
	pos  int
	end  int
}

type scanCtx struct {
	ctx   rule.RuleContext
	sf    *ast.SourceFile
	text  string
	style string
}

// listScanSpec describes a single list validation.
//
//   - startPos / scanEnd: byte range to forward-scan. For bracketed lists
//     pass `list.Pos()` and a position past the close bracket (typically
//     `parent.End()` — the close-bracket token gets pulled into the token
//     slice so trailing-comma neighbors are well-defined).
//   - items: the AST elements of the list (NodeList.Nodes). Used to filter
//     out commas that fall *inside* an element's range (heritage-clause
//     types, sequence operands containing class expressions, …). Pass the
//     element list as-is — `KindOmittedExpression` entries are zero-width
//     and don't shadow any token.
//   - startsAfterArrayOpen: true for `ArrayLiteralExpression` /
//     `ArrayBindingPattern` lists, where upstream's array-hole carve-out
//     skips a comma whose preceding context (walking back through any
//     contiguous comma chain) reaches the `[` token. Our scan starts AFTER
//     `[`, so this flag stands in for "no earlier token in the slice =
//     the `[` is what's there in the source."
type listScanSpec struct {
	startPos             int
	scanEnd              int
	items                []*ast.Node
	startsAfterArrayOpen bool
}

// validateList forward-scans [spec.startPos, spec.scanEnd), collects every
// comma whose position falls *outside* every item's range (those are
// separator commas at this list's level), then dispatches each to
// `validateOneComma`.
//
// The "outside every item range" filter is what makes this AST-driven scan
// resilient to nested unbracketed comma-bearing constructs (heritage
// clauses, sequence expressions containing classes, etc.) — those commas
// physically live inside one of the list's items, so they're skipped here
// and picked up by the dedicated listener for the inner node.
func (sc *scanCtx) validateList(spec listScanSpec) {
	if spec.startPos < 0 || spec.scanEnd <= spec.startPos {
		return
	}

	itemRanges := make([][2]int, 0, len(spec.items))
	for _, it := range spec.items {
		if it == nil {
			continue
		}
		// OmittedExpression carries a zero-width range at the position of
		// the surrounding `,` — it can't contain any token, so skipping it
		// here keeps the filter from masking the separator comma whose
		// position it shares.
		if it.Kind == ast.KindOmittedExpression {
			continue
		}
		// Compute the item's "first significant token" position by
		// skipping leading trivia (whitespace + comments). Used as the
		// pseudo-token's `pos` so adjacent-comma sameLine checks
		// reference the real token, not the trivia run.
		startTok := scanner.SkipTrivia(sc.text, it.Pos())
		// tsgo's `parseTypeMemberSemicolon` consumes a trailing `,` or `;`
		// AS PART OF the TypeElement's parse, so member.End() lands past
		// the separator (often at the start of the next member). The
		// `insideItem` filter would then incorrectly classify the separator
		// comma as "inside this member". `itemContentEnd` walks the item's
		// own tokens and returns the position right after its LAST
		// significant (non-`,`, non-`;`) token, which gives us the real
		// content end. For items whose parser does NOT consume the
		// separator (most kinds — array elements, object literal
		// properties, parameters, …), this is a no-op.
		end := sc.itemContentEnd(it)
		itemRanges = append(itemRanges, [2]int{startTok, end})
	}

	// itemContaining returns the (pos, end) of the smallest item-range
	// that strictly contains `pos` (pos in [r[0], r[1])), or
	// (-1, -1, false) when `pos` is outside every item. Used to skip
	// the scanner past list items whose interior content could confuse
	// tokenization (notably templated strings — when scanned without
	// template-state context, a nested closing back-tick is
	// mis-tokenized as a fresh `NoSubstitutionTemplateLiteral`,
	// swallowing every byte until the next back-tick or EOF).
	itemContaining := func(pos int) (int, int, bool) {
		for _, r := range itemRanges {
			if pos >= r[0] && pos < r[1] {
				return r[0], r[1], true
			}
		}
		return -1, -1, false
	}

	s := scanner.GetScannerForSourceFile(sc.sf, spec.startPos)
	var tokens []tk
	var commaIdx []int

	for {
		kind := s.Token()
		if kind == ast.KindEndOfFile {
			break
		}
		start := s.TokenStart()
		if start >= spec.scanEnd {
			break
		}
		if itemPos, itemEnd, ok := itemContaining(start); ok {
			// Append a pseudo-token representing the item as a single
			// opaque unit. The `kind` is `Unknown` — deliberately not
			// `,` and not `[` so the validateOneComma skip-conditions
			// don't fire on it; only its (pos, end) is used, for the
			// sameLine check against an adjacent separator comma.
			tokens = append(tokens, tk{ast.KindUnknown, itemPos, itemEnd})
			s = scanner.GetScannerForSourceFile(sc.sf, itemEnd)
			continue
		}
		end := s.TokenEnd()
		tokens = append(tokens, tk{kind, start, end})
		if kind == ast.KindCommaToken {
			commaIdx = append(commaIdx, len(tokens)-1)
		}
		s.Scan()
	}

	for _, i := range commaIdx {
		sc.validateOneComma(tokens, i, spec.startsAfterArrayOpen)
	}
}

// validateOneComma applies upstream's three skip-conditions, then dispatches
// to validateCommaItemSpacing for the actual reporting decision.
//
// `startsAfterArrayOpen` extends upstream's second skip-condition into our
// scan model: upstream walks back through any preceding comma chain and
// skips when it lands on `[`. Our token slice starts *after* the array's
// `[`, so when the walk-back runs off the front of the slice we treat
// that as "would have been `[`" iff the list is an array.
func (sc *scanCtx) validateOneComma(tokens []tk, idx int, startsAfterArrayOpen bool) {
	if idx <= 0 || idx >= len(tokens)-1 {
		// Need both tokenBefore and tokenAfter to validate; bail otherwise.
		return
	}
	before := tokens[idx-1]
	comma := tokens[idx]
	after := tokens[idx+1]

	// Skip: token before the comma is `[` — this is the first comma of an
	// array, i.e. `[, …]`. Upstream uses `isOpeningBracketToken` for this.
	if before.kind == ast.KindOpenBracketToken {
		return
	}

	// Skip: token before is a comma AND walking back past consecutive commas
	// either lands on `[` (upstream) or runs off the front of an array
	// scan (we started after `[`).
	if before.kind == ast.KindCommaToken {
		j := idx - 2
		for j >= 0 && tokens[j].kind == ast.KindCommaToken {
			j--
		}
		if j < 0 {
			if startsAfterArrayOpen {
				return
			}
		} else if tokens[j].kind == ast.KindOpenBracketToken {
			return
		}
	}

	// Skip: token after is a comma AND on a different line from the current
	// comma. Two commas vertically stacked are "consecutive holes" upstream
	// chooses to leave alone (the next iteration's comma will be reported).
	if after.kind == ast.KindCommaToken && !sc.sameLine(comma.pos, after.pos) {
		return
	}

	sc.validateCommaItemSpacing(before, comma, after)
}

// validateCommaItemSpacing is the direct port of upstream's same-named
// function. The four branches map 1:1 to upstream:
//
//  1. before AND after on same line as comma → no diagnostic.
//  2. before AND after both on different lines (lone comma) → report
//     `unexpectedLineBeforeAndAfterComma`. Fix shape: "between" (strip the
//     first linebreak from the trivia, prepend the comma) UNLESS the first
//     comment after the comma is a block comment on the same line — in that
//     case use the configured style so the block comment stays adjacent.
//  3. style == 'first' && comma not on same line as `after` → report
//     `expectedCommaFirst` with a 'first' fix.
//  4. style == 'last' && comma on same line as `after` → report
//     `expectedCommaLast` with a 'last' fix.
func (sc *scanCtx) validateCommaItemSpacing(before, comma, after tk) {
	sameLineBefore := sc.sameLine(before.end, comma.pos)
	sameLineAfter := sc.sameLine(comma.end, after.pos)

	switch {
	case sameLineBefore && sameLineAfter:
		return
	case !sameLineBefore && !sameLineAfter:
		fixStyle := fixStyleBetween
		if sc.hasBlockCommentOnCommaLine(comma) {
			fixStyle = sc.style
		}
		sc.report(comma, msgUnexpectedLineBeforeAndAfterComma, descUnexpectedLineBeforeAndAfterComma,
			sc.buildFix(before, comma, after, fixStyle))
	case sc.style == styleFirst && !sameLineAfter:
		sc.report(comma, msgExpectedCommaFirst, descExpectedCommaFirst,
			sc.buildFix(before, comma, after, styleFirst))
	case sc.style == styleLast && sameLineAfter:
		sc.report(comma, msgExpectedCommaLast, descExpectedCommaLast,
			sc.buildFix(before, comma, after, styleLast))
	}
}

// hasBlockCommentOnCommaLine mirrors upstream's
// `getCommentsAfter(commaToken)[0]` + "Block && same line" guard. Walks
// trivia bytes after the comma; returns true iff the FIRST comment-like
// trivia we encounter is a `/* … */` block comment that starts before any
// linebreak (i.e. on the same line as the comma).
//
// Operates on byte-level trivia rather than asking the scanner because the
// scanner skips comments entirely — once it's moved past `comma`, comment
// tokens are gone. The trivia between the comma and the next significant
// position is by construction not inside a string / regex / template
// literal, so byte-level `//` / `/*` detection is unambiguous.
//
// Linebreak detection mirrors ECMAScript LineTerminator (§12.3): LF, CR,
// LS (U+2028, UTF-8 `E2 80 A8`), PS (U+2029, UTF-8 `E2 80 A9`).
func (sc *scanCtx) hasBlockCommentOnCommaLine(comma tk) bool {
	p := comma.end
	for p < len(sc.text) {
		b := sc.text[p]
		switch {
		case b == ' ' || b == '\t' || b == '\v' || b == '\f':
			p++
		case b == '\n' || b == '\r':
			return false
		case b == 0xE2 && p+2 < len(sc.text) && sc.text[p+1] == 0x80 &&
			(sc.text[p+2] == 0xA8 || sc.text[p+2] == 0xA9):
			// LS / PS — counts as a linebreak under ECMA §12.3.
			return false
		case b == '/' && p+1 < len(sc.text):
			next := sc.text[p+1]
			if next == '/' {
				return false
			}
			if next == '*' {
				return true
			}
			return false
		default:
			return false
		}
	}
	return false
}

// buildFix constructs the (replacement-text, range) pair for a comma move.
// The range covers [before.end, after.pos) — everything between the two
// neighbor tokens, including the comma itself and surrounding trivia. The
// replacement is derived from upstream's `getReplacedText`:
//
//	between → ',' + trivia-with-first-linebreak-removed
//	first   → trivia + ','
//	last    → ',' + trivia
//
// `trivia` is the original byte range with the comma carved out.
func (sc *scanCtx) buildFix(before, comma, after tk, fixStyle string) rule.RuleFix {
	trivia := sc.text[before.end:comma.pos] + sc.text[comma.end:after.pos]
	var replaced string
	switch fixStyle {
	case fixStyleBetween:
		replaced = "," + removeFirstLinebreak(trivia)
	case styleFirst:
		replaced = trivia + ","
	case styleLast:
		replaced = "," + trivia
	}
	return rule.RuleFix{
		Text:  replaced,
		Range: core.NewTextRange(before.end, after.pos),
	}
}

// removeFirstLinebreak strips the first ECMAScript LineTerminator sequence
// from `s`: CR LF (treated as one), LF, CR, U+2028 (LS), or U+2029 (PS).
// Mirrors `text.replace(LINEBREAK_MATCHER, '')` upstream — non-global,
// only the first match is removed.
func removeFirstLinebreak(s string) string {
	for i := range len(s) {
		c := s[i]
		switch c {
		case '\r':
			if i+1 < len(s) && s[i+1] == '\n' {
				return s[:i] + s[i+2:]
			}
			return s[:i] + s[i+1:]
		case '\n':
			return s[:i] + s[i+1:]
		case 0xE2:
			if i+2 < len(s) && s[i+1] == 0x80 && (s[i+2] == 0xA8 || s[i+2] == 0xA9) {
				return s[:i] + s[i+3:]
			}
		}
	}
	return s
}

// report emits the diagnostic anchored on the comma's single-char range,
// matching upstream's `loc: commaToken.loc`.
func (sc *scanCtx) report(comma tk, id, desc string, fix rule.RuleFix) {
	sc.ctx.ReportRangeWithFixes(
		core.NewTextRange(comma.pos, comma.end),
		rule.RuleMessage{Id: id, Description: desc},
		fix,
	)
}

func (sc *scanCtx) sameLine(a, b int) bool {
	return stylisticutil.SameLineByPos(sc.sf, a, b)
}

// itemContentEnd returns the position right after the LAST significant
// token (non-`,`, non-`;`) inside `item`'s source range. Necessary because
// tsgo's `parseTypeMemberSemicolon` folds a trailing separator into the
// TypeElement's range; without this, the separator-comma filter would
// flag those commas as "inside this member" and skip them.
func (sc *scanCtx) itemContentEnd(item *ast.Node) int {
	if item == nil {
		return 0
	}
	end := item.End()
	pos := item.Pos()
	if pos >= end {
		return end
	}
	s := scanner.GetScannerForSourceFile(sc.sf, pos)
	contentEnd := pos
	for {
		kind := s.Token()
		if kind == ast.KindEndOfFile {
			break
		}
		start := s.TokenStart()
		if start >= end {
			break
		}
		tokEnd := s.TokenEnd()
		if kind != ast.KindCommaToken && kind != ast.KindSemicolonToken {
			contentEnd = tokEnd
		}
		s.Scan()
	}
	if contentEnd > end {
		contentEnd = end
	}
	return contentEnd
}

// validateBracketedList wraps the common shape for "scan a NodeList that
// sits inside brackets". The scan goes from `list.Pos()` to `parent.End()`
// so the close bracket lands in the token slice as the trailing-comma
// neighbor.
func (sc *scanCtx) validateBracketedList(list *ast.NodeList, parentEnd int, startsAfterArrayOpen bool) {
	if list == nil {
		return
	}
	sc.validateList(listScanSpec{
		startPos:             list.Pos(),
		scanEnd:              parentEnd,
		items:                list.Nodes,
		startsAfterArrayOpen: startsAfterArrayOpen,
	})
}

// validateBracketedListAtCloseChar is like validateBracketedList but bounds
// `scanEnd` to one byte past the actual close character, located by
// `scanner.SkipTrivia` from `list.End()`. Required for containers whose
// "node" extends past the close bracket (e.g. function-like — `node.End()`
// includes the body, not just the parameter `)`).
//
// `closeChar` is the expected close byte (`)`, `]`, etc.). When the byte
// at the located position doesn't match, falls back to `list.End()` —
// defensive against parser recovery shapes.
func (sc *scanCtx) validateBracketedListAtCloseChar(list *ast.NodeList, closeChar byte, startsAfterArrayOpen bool) {
	if list == nil {
		return
	}
	scanEnd := list.End()
	closePos := scanner.SkipTrivia(sc.text, list.End())
	if closePos < len(sc.text) && sc.text[closePos] == closeChar {
		scanEnd = closePos + 1
	}
	sc.validateList(listScanSpec{
		startPos:             list.Pos(),
		scanEnd:              scanEnd,
		items:                list.Nodes,
		startsAfterArrayOpen: startsAfterArrayOpen,
	})
}

// findHeritageClause returns the first HeritageClause matching `tokenKind`
// (KindImplementsKeyword or KindExtendsKeyword) on a class / interface
// declaration, or nil. Used to surface the only clause comma-style cares
// about for each container — `implements` for classes, `extends` for
// interfaces.
func findHeritageClause(list *ast.NodeList, tokenKind ast.Kind) *ast.Node {
	if list == nil {
		return nil
	}
	for _, clause := range list.Nodes {
		hc := clause.AsHeritageClause()
		if hc != nil && hc.Token == tokenKind {
			return clause
		}
	}
	return nil
}

// CommaStyleRule enforces consistent comma placement (last / first) across
// every comma-separated list ESLint Stylistic's comma-style recognizes:
// arrays, objects, parameter lists, import / export specifiers, sequence
// expressions, TS enums / type literals / interfaces / tuples / type params,
// etc. Ported from `@stylistic/eslint-plugin`'s comma-style.
//
// Listener fan-out vs. upstream ESTree:
//
//   - tsgo collapses ESLint's `VariableDeclaration` (statement-level
//     `var a, b`) and the for-init form into a `VariableDeclarationList`. We
//     listen on that single kind to cover both.
//   - tsgo collapses ESTree's `MethodDefinition.value = FunctionExpression`
//     into standalone `MethodDeclaration` / `Constructor` / `Get|SetAccessor`
//     nodes. We fire them with the `FunctionExpression` exception bucket so
//     class methods retain upstream's coverage; body-less forms (abstract
//     methods, overload signatures) switch to the
//     `TSEmptyBodyFunctionExpression` bucket.
//   - tsgo has no separate `TSDeclareFunction` kind; `declare function f()`
//     parses as a body-less `FunctionDeclaration`. We split by `Body == nil`
//     to pick between the `FunctionDeclaration` and `TSDeclareFunction`
//     exception buckets.
//   - tsgo has no `ImportExpression` kind; dynamic `import(source, opts)`
//     parses as a `CallExpression` whose `Expression.Kind` is
//     `KindImportKeyword`. We branch on that inside the CallExpression
//     listener.
//   - tsgo has no flat `SequenceExpression`; `a, b, c` is left-associative
//     `BinaryExpression(BinaryExpression(a, ',', b), ',', c)`. We only fire
//     on the outermost comma-BinaryExpression (parent isn't itself a
//     comma-BinaryExpression), then scan its whole range for separator
//     commas. Nested parenthesized sequences fire independently because
//     their parent is a `ParenthesizedExpression`, not a comma-BinExpr.
//   - tsgo has no `TSTypeParameterDeclaration` / `TSTypeParameterInstantiation`
//     kinds; type-param / type-arg lists are fields on a wide set of carrier
//     nodes. We delegate to `node.TypeArgumentList()` /
//     `node.TypeParameterList()` (typescript-go/internal/ast/ast.go:483, 515)
//     so the carrier-kind set stays in lockstep with tsgo's own dispatch.
//   - tsgo has no separate `TSInterfaceBody` / `TSEnumBody` nodes; the
//     members live directly on `InterfaceDeclaration` / `EnumDeclaration`.
//     The InterfaceDeclaration listener handles both the `extends` clause
//     (TSInterfaceDeclaration exception) and the member list (TSInterfaceBody
//     exception) in one pass.
//   - For class `implements` / interface `extends`, tsgo wraps the type list
//     inside a `HeritageClause` keyed by token kind (Implements/Extends).
//     We pick the relevant clause and scan its `Types` NodeList.
//
// Separator detection is AST-driven: every comma found in a list's scan
// range is filtered against the children's source ranges, and only commas
// that fall *outside* every child's range are treated as list-level
// separators. This is what lets a `VariableDeclarationList` listener
// scan over `var a = class implements X, Y {}` without misclassifying the
// heritage-clause comma as a var-decl separator — the comma sits inside the
// `VariableDeclaration` child's range, so the filter rejects it.
var CommaStyleRule = rule.Rule{
	Name: "@stylistic/comma-style",
	Run: func(ctx rule.RuleContext, rawOptions any) rule.RuleListeners {
		opts := parseOptions(rawOptions)
		sc := &scanCtx{
			ctx:   ctx,
			sf:    ctx.SourceFile,
			text:  ctx.SourceFile.Text(),
			style: opts.style,
		}
		exc := opts.exceptions

		// checkTypeArgs / checkTypeParams handle the type-argument /
		// type-parameter lists exposed on a wide set of carrier kinds.
		// `node.TypeArgumentList()` / `.TypeParameterList()` keep the
		// carrier set in lockstep with tsgo's own dispatch.
		// Type-arg / type-param lists end at the `>` close token, whose
		// position is `list.End()`. We pass `list.End() + 1` so the scan
		// includes the close `>` itself (needed for the trailing-comma
		// `tokenAfter` check). Using `node.End()` would pull in tokens
		// from BEYOND the close `>` — e.g. `type Foo<A> = Bar<A>` would
		// have a TypeAlias-level scan that bleeds into Bar's contents.
		checkTypeArgs := func(node *ast.Node) {
			if exc[excTSTypeParameterInstantiation] {
				return
			}
			list := node.TypeArgumentList()
			if list == nil || len(list.Nodes) == 0 {
				return
			}
			sc.validateBracketedListAtCloseChar(list, '>', false)
		}
		checkTypeParams := func(node *ast.Node) {
			if exc[excTSTypeParameterDeclaration] {
				return
			}
			list := node.TypeParameterList()
			if list == nil || len(list.Nodes) == 0 {
				return
			}
			sc.validateBracketedListAtCloseChar(list, '>', false)
		}

		// checkFunctionLike validates a function-like node's parameters
		// against `paramsBucket` and (when applicable) its type parameters.
		// Upstream splits the parameter check across `FunctionDeclaration`,
		// `FunctionExpression`, `TSDeclareFunction`, `TSEmptyBodyFunctionExpression`
		// based on AST kind + body presence; we recover that split at the
		// call site by choosing `paramsBucket`.
		//
		// scanEnd MUST be `Parameters.End()+1` (just past the close `)`)
		// rather than `node.End()` — `node.End()` for a function-like
		// extends past the close paren into the body, and our
		// AST-driven separator filter only excludes commas inside the
		// listed `items` (= parameters). Body-level commas would
		// otherwise be misclassified as parameter-list separators.
		checkFunctionLike := func(node *ast.Node, paramsBucket string) {
			if !exc[paramsBucket] {
				fl := node.FunctionLikeData()
				if fl != nil && fl.Parameters != nil {
					sc.validateBracketedListAtCloseChar(fl.Parameters, ')', false)
				}
			}
			checkTypeParams(node)
		}

		return rule.RuleListeners{
			ast.KindVariableDeclarationList: func(node *ast.Node) {
				if exc[excVariableDeclaration] {
					return
				}
				vdl := node.AsVariableDeclarationList()
				if vdl == nil || vdl.Declarations == nil {
					return
				}
				sc.validateList(listScanSpec{
					startPos: vdl.Declarations.Pos(),
					scanEnd:  node.End(),
					items:    vdl.Declarations.Nodes,
				})
			},
			ast.KindObjectLiteralExpression: func(node *ast.Node) {
				if exc[excObjectExpression] {
					return
				}
				obj := node.AsObjectLiteralExpression()
				if obj == nil {
					return
				}
				sc.validateBracketedList(obj.Properties, node.End(), false)
			},
			ast.KindObjectBindingPattern: func(node *ast.Node) {
				if exc[excObjectPattern] {
					return
				}
				bp := node.AsBindingPattern()
				if bp == nil {
					return
				}
				sc.validateBracketedList(bp.Elements, node.End(), false)
			},
			ast.KindArrayLiteralExpression: func(node *ast.Node) {
				if exc[excArrayExpression] {
					return
				}
				arr := node.AsArrayLiteralExpression()
				if arr == nil {
					return
				}
				sc.validateBracketedList(arr.Elements, node.End(), true)
			},
			ast.KindArrayBindingPattern: func(node *ast.Node) {
				if exc[excArrayPattern] {
					return
				}
				bp := node.AsBindingPattern()
				if bp == nil {
					return
				}
				sc.validateBracketedList(bp.Elements, node.End(), true)
			},
			ast.KindFunctionDeclaration: func(node *ast.Node) {
				bucket := excFunctionDeclaration
				if node.Body() == nil {
					bucket = excTSDeclareFunction
				}
				checkFunctionLike(node, bucket)
			},
			ast.KindFunctionExpression: func(node *ast.Node) {
				checkFunctionLike(node, excFunctionExpression)
			},
			ast.KindArrowFunction: func(node *ast.Node) {
				checkFunctionLike(node, excArrowFunctionExpression)
			},
			// Class members map to the FunctionExpression bucket when they
			// have a body and to TSEmptyBodyFunctionExpression when they
			// don't — mirroring upstream's MethodDefinition.value =
			// FunctionExpression vs. TSEmptyBodyFunctionExpression split.
			ast.KindMethodDeclaration: func(node *ast.Node) {
				bucket := excFunctionExpression
				if node.Body() == nil {
					bucket = excTSEmptyBodyFunctionExpression
				}
				checkFunctionLike(node, bucket)
			},
			ast.KindConstructor: func(node *ast.Node) {
				bucket := excFunctionExpression
				if node.Body() == nil {
					bucket = excTSEmptyBodyFunctionExpression
				}
				if !exc[bucket] {
					fl := node.FunctionLikeData()
					if fl != nil && fl.Parameters != nil {
						sc.validateBracketedListAtCloseChar(fl.Parameters, ')', false)
					}
				}
				// Constructor has no type parameters in tsgo's model.
			},
			ast.KindGetAccessor: func(node *ast.Node) {
				bucket := excFunctionExpression
				if node.Body() == nil {
					bucket = excTSEmptyBodyFunctionExpression
				}
				checkFunctionLike(node, bucket)
			},
			ast.KindSetAccessor: func(node *ast.Node) {
				bucket := excFunctionExpression
				if node.Body() == nil {
					bucket = excTSEmptyBodyFunctionExpression
				}
				checkFunctionLike(node, bucket)
			},
			ast.KindFunctionType: func(node *ast.Node) {
				checkFunctionLike(node, excTSFunctionType)
			},
			ast.KindConstructorType: func(node *ast.Node) {
				checkFunctionLike(node, excTSConstructorType)
			},
			ast.KindMethodSignature: func(node *ast.Node) {
				checkFunctionLike(node, excTSMethodSignature)
			},
			ast.KindCallSignature: func(node *ast.Node) {
				checkFunctionLike(node, excTSCallSignatureDeclaration)
			},
			ast.KindConstructSignature: func(node *ast.Node) {
				checkFunctionLike(node, excTSConstructSignatureDeclaration)
			},
			ast.KindIndexSignature: func(node *ast.Node) {
				if exc[excTSIndexSignature] {
					return
				}
				fl := node.FunctionLikeData()
				if fl != nil && fl.Parameters != nil {
					// IndexSignature's params are inside `[ ]`; the type
					// (after `:`) follows and could contain its own
					// commas (object-type literals). Bound the scan
					// strictly to the `[ ]` so those don't leak in.
					sc.validateBracketedListAtCloseChar(fl.Parameters, ']', false)
				}
			},
			ast.KindCallExpression: func(node *ast.Node) {
				call := node.AsCallExpression()
				if call == nil {
					return
				}
				// Dynamic `import(...)` lives under the ImportExpression
				// exception bucket; everything else is a regular call.
				isImport := call.Expression != nil && call.Expression.Kind == ast.KindImportKeyword
				argBucket := excCallExpression
				if isImport {
					argBucket = excImportExpression
				}
				if !exc[argBucket] && call.Arguments != nil {
					sc.validateBracketedList(call.Arguments, node.End(), false)
				}
				if !isImport {
					checkTypeArgs(node)
				}
			},
			ast.KindNewExpression: func(node *ast.Node) {
				ne := node.AsNewExpression()
				if ne == nil {
					return
				}
				if !exc[excNewExpression] && ne.Arguments != nil {
					sc.validateBracketedList(ne.Arguments, node.End(), false)
				}
				checkTypeArgs(node)
			},
			ast.KindImportDeclaration: func(node *ast.Node) {
				// Upstream's whole `ImportDeclaration` visitor is gated on
				// `!exceptions.ImportDeclaration` — both the specifier list
				// and the with-attributes are skipped together.
				if exc[excImportDeclaration] {
					return
				}
				id := node.AsImportDeclaration()
				if id == nil {
					return
				}
				sc.checkImportDeclarationSpecifiers(id)
				if id.Attributes != nil {
					sc.checkImportAttributes(id.Attributes)
				}
			},
			ast.KindExportDeclaration: func(node *ast.Node) {
				ed := node.AsExportDeclaration()
				if ed == nil {
					return
				}
				if ed.ExportClause == nil {
					// `export * from 'x'` — gated by ExportAllDeclaration.
					if !exc[excExportAllDeclaration] && ed.Attributes != nil {
						sc.checkImportAttributes(ed.Attributes)
					}
					return
				}
				// `export { … } from 'x'` / `export { … }` — gated by
				// ExportNamedDeclaration. Both the specifier list and the
				// attributes are guarded by the same bucket upstream.
				if exc[excExportNamedDeclaration] {
					return
				}
				if ed.ExportClause.Kind == ast.KindNamedExports {
					ne := ed.ExportClause.AsNamedExports()
					if ne != nil {
						sc.validateBracketedList(ne.Elements, ed.ExportClause.End(), false)
					}
				}
				if ed.Attributes != nil {
					sc.checkImportAttributes(ed.Attributes)
				}
			},
			ast.KindBinaryExpression: func(node *ast.Node) {
				if exc[excSequenceExpression] {
					return
				}
				bin := node.AsBinaryExpression()
				if bin == nil || bin.OperatorToken == nil ||
					bin.OperatorToken.Kind != ast.KindCommaToken {
					return
				}
				// Only validate from the top-level comma-BinExpr — otherwise
				// each nested level would re-validate already-covered commas.
				if parent := node.Parent; parent != nil && parent.Kind == ast.KindBinaryExpression {
					if parentBin := parent.AsBinaryExpression(); parentBin != nil &&
						parentBin.OperatorToken != nil &&
						parentBin.OperatorToken.Kind == ast.KindCommaToken {
						return
					}
				}
				// Flatten the left-associative comma chain into a flat
				// `items` list so the filter excludes commas hidden inside
				// any operand (parens around a class with implements, etc.).
				items := flattenCommaChain(node)
				sc.validateList(listScanSpec{
					startPos: node.Pos(),
					scanEnd:  node.End(),
					items:    items,
				})
			},
			ast.KindClassDeclaration: func(node *ast.Node) {
				cd := node.AsClassDeclaration()
				if cd == nil {
					return
				}
				if !exc[excClassDeclaration] {
					if clause := findHeritageClause(cd.HeritageClauses, ast.KindImplementsKeyword); clause != nil {
						hc := clause.AsHeritageClause()
						if hc != nil && hc.Types != nil {
							sc.validateList(listScanSpec{
								startPos: hc.Types.Pos(),
								scanEnd:  clause.End(),
								items:    hc.Types.Nodes,
							})
						}
					}
				}
				checkTypeParams(node)
			},
			ast.KindClassExpression: func(node *ast.Node) {
				ce := node.AsClassExpression()
				if ce == nil {
					return
				}
				if !exc[excClassExpression] {
					if clause := findHeritageClause(ce.HeritageClauses, ast.KindImplementsKeyword); clause != nil {
						hc := clause.AsHeritageClause()
						if hc != nil && hc.Types != nil {
							sc.validateList(listScanSpec{
								startPos: hc.Types.Pos(),
								scanEnd:  clause.End(),
								items:    hc.Types.Nodes,
							})
						}
					}
				}
				checkTypeParams(node)
			},
			ast.KindInterfaceDeclaration: func(node *ast.Node) {
				id := node.AsInterfaceDeclaration()
				if id == nil {
					return
				}
				// extends clause — TSInterfaceDeclaration bucket
				if !exc[excTSInterfaceDeclaration] {
					if clause := findHeritageClause(id.HeritageClauses, ast.KindExtendsKeyword); clause != nil {
						hc := clause.AsHeritageClause()
						if hc != nil && hc.Types != nil {
							sc.validateList(listScanSpec{
								startPos: hc.Types.Pos(),
								scanEnd:  clause.End(),
								items:    hc.Types.Nodes,
							})
						}
					}
				}
				// member list — TSInterfaceBody bucket
				if !exc[excTSInterfaceBody] && id.Members != nil {
					sc.validateBracketedList(id.Members, node.End(), false)
				}
				checkTypeParams(node)
			},
			ast.KindTypeAliasDeclaration: func(node *ast.Node) {
				checkTypeParams(node)
			},
			ast.KindEnumDeclaration: func(node *ast.Node) {
				if exc[excTSEnumBody] {
					return
				}
				ed := node.AsEnumDeclaration()
				if ed == nil {
					return
				}
				sc.validateBracketedList(ed.Members, node.End(), false)
			},
			ast.KindTypeLiteral: func(node *ast.Node) {
				if exc[excTSTypeLiteral] {
					return
				}
				tl := node.AsTypeLiteralNode()
				if tl == nil {
					return
				}
				sc.validateBracketedList(tl.Members, node.End(), false)
			},
			ast.KindTupleType: func(node *ast.Node) {
				if exc[excTSTupleType] {
					return
				}
				tt := node.AsTupleTypeNode()
				if tt == nil {
					return
				}
				sc.validateBracketedList(tt.Elements, node.End(), false)
			},
			// Pure type-arg carriers — only `TSTypeParameterInstantiation`
			// in upstream terms.
			ast.KindTaggedTemplateExpression:    checkTypeArgs,
			ast.KindTypeReference:               checkTypeArgs,
			ast.KindExpressionWithTypeArguments: checkTypeArgs,
			ast.KindImportType:                  checkTypeArgs,
			ast.KindTypeQuery:                   checkTypeArgs,
			ast.KindJsxOpeningElement:           checkTypeArgs,
			ast.KindJsxSelfClosingElement:       checkTypeArgs,
		}
	},
}

// flattenCommaChain walks a left-associative `a, b, c` chain
// (`BinaryExpression(BinaryExpression(a, ',', b), ',', c)`) back into its
// flat operand list `[a, b, c]`. Required so the separator-comma filter
// can correctly exclude commas hidden inside any single operand (e.g. a
// class expression with multiple `implements` types).
func flattenCommaChain(node *ast.Node) []*ast.Node {
	var out []*ast.Node
	var walk func(n *ast.Node)
	walk = func(n *ast.Node) {
		if n == nil {
			return
		}
		if n.Kind == ast.KindBinaryExpression {
			bin := n.AsBinaryExpression()
			if bin != nil && bin.OperatorToken != nil &&
				bin.OperatorToken.Kind == ast.KindCommaToken {
				walk(bin.Left)
				walk(bin.Right)
				return
			}
		}
		out = append(out, n)
	}
	walk(node)
	return out
}

// checkImportDeclarationSpecifiers reproduces upstream's
// `validateComma(node, node.specifiers)` semantics on tsgo's split shape.
//
// ESTree's `specifiers` flattens default / namespace / named imports into
// one list; tsgo splits them across `ImportClause.name` (default),
// `ImportClause.NamedBindings` (namespace OR named), so the commas of
// interest are:
//
//  1. The comma between default and named-bindings, when both exist
//     (e.g. `import a, { b } from 'x'`).
//  2. The commas inside `{ a, b, c }` (NamedImports.Elements).
//
// We make two scans — one over a synthetic two-item list (default,
// named-bindings) bounded by [default.Pos(), namedBindings.End()] so the
// inside-item filter naturally hides the commas inside `{ … }` — and one
// over the bracketed elements.
func (sc *scanCtx) checkImportDeclarationSpecifiers(id *ast.ImportDeclaration) {
	if id == nil || id.ImportClause == nil {
		return
	}
	ic := id.ImportClause.AsImportClause()
	if ic == nil {
		return
	}
	def := ic.Name()
	nb := ic.NamedBindings
	if def != nil && nb != nil {
		sc.validateList(listScanSpec{
			startPos: def.Pos(),
			scanEnd:  nb.End(),
			items:    []*ast.Node{def, nb},
		})
	}
	if nb != nil && nb.Kind == ast.KindNamedImports {
		named := nb.AsNamedImports()
		if named != nil {
			sc.validateBracketedList(named.Elements, nb.End(), false)
		}
	}
}

// checkImportAttributes validates commas inside an import / export
// `with { a: 'v', b: 'v' }` clause. The clause's grammar is the same shape
// as a NamedImports `{ … }` list, so we reuse the bracketed-list helper.
func (sc *scanCtx) checkImportAttributes(attrs *ast.Node) {
	if attrs == nil {
		return
	}
	ia := attrs.AsImportAttributes()
	if ia == nil || ia.Attributes == nil {
		return
	}
	sc.validateBracketedList(ia.Attributes, attrs.End(), false)
}
