package jsx_indent

import (
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/plugins/react/reactutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// JsxIndentRule enforces JSX indentation.
//
// Ported from eslint-plugin-react's `jsx-indent` rule. Each opening / closing
// tag, JSXExpressionContainer, JSXText and JSX-returning ReturnStatement is
// checked against a parent-derived "expected indent"; violations carry an
// autofix that rewrites the leading whitespace of the offending line.
//
// tsgo↔ESTree shape adjustments (vs the upstream JS rule):
//   - Self-closing `<Foo />` is `JsxSelfClosingElement` with no wrapping
//     `JsxElement`, so the listener resolves the "operand position" via
//     `jsxOperandPosition(node)` before walking up the tree.
//   - Parens / `as` / `satisfies` / `!` are explicit nodes; we use
//     `reactutil.SkipExpressionWrappersUp` to flatten them when matching
//     LogicalExpression / ConditionalExpression ancestors.
//   - Token positions for source-level scans (e.g. "is the previous char a
//     colon") are computed from `SourceFile.Text()`; line numbers from
//     `ECMALineMap()` + `scanner.ComputeLineOfPosition`.
var JsxIndentRule = rule.Rule{
	Name: "react/jsx-indent",
	Run: func(ctx rule.RuleContext, _options []any) rule.RuleListeners {
		options := rule.UnwrapOptions(_options)
		indentType, indentSize, indentChar := parseIndentOption(options)
		checkAttributes, indentLogicalExpressions := parseSecondOption(options)

		text := ctx.SourceFile.Text()
		lineMap := ctx.SourceFile.ECMALineMap()

		// lineOfPos is the only line-number helper kept local; the indent
		// / first-in-line / message helpers live in `reactutil`.
		lineOfPos := func(pos int) int {
			return scanner.ComputeLineOfPosition(lineMap, pos)
		}

		// reportLiteral emits one diagnostic per mis-indented line of a
		// JsxText. The report and fix BOTH use the JsxText's RAW range —
		// tsgo treats the leading newline / whitespace as trivia, but
		// upstream's `report(node, …)` anchors at `node.loc.start` which
		// sits at the JsxText's raw start (right after the preceding `>`
		// of the parent's opening tag).
		reportLiteral := func(node *ast.Node, needed, gotten int, fixed string) {
			rawRange := core.NewTextRange(node.Pos(), node.End())
			ctx.ReportRangeWithFixes(rawRange, reactutil.WrongIndentMessage(needed, gotten, indentType), rule.RuleFix{
				Text:  fixed,
				Range: rawRange,
			})
		}

		// reportReturn emits a wrongIndent diagnostic for a ReturnStatement
		// whose opening / closing column don't match. Fix replaces just
		// the last line of the return; mirrors upstream's
		// `replaceTextRange([lastNL, end])`.
		reportReturn := func(node *ast.Node, needed, gotten int) {
			indent := strings.Repeat(string(indentChar), needed)
			trimmed := utils.TrimNodeTextRange(ctx.SourceFile, node)
			start := trimmed.Pos()
			end := trimmed.End()
			raw := text[start:end]
			msg := reactutil.WrongIndentMessage(needed, gotten, indentType)
			if !strings.Contains(raw, "\n") {
				ctx.ReportNode(node, msg)
				return
			}
			lastNL := strings.LastIndex(raw, "\n")
			lastLine := raw[lastNL:]
			fixedLast := replaceFirstLeadingIndent(lastLine, indent)
			ctx.ReportNodeWithFixes(node, msg, rule.RuleFix{
				Text:  fixedLast,
				Range: core.NewTextRange(start+lastNL, end),
			})
		}

		// reportAttribute emits a wrongIndent diagnostic anchored at the
		// `)` punctuator (or whatever first-in-line anchor we found
		// inside the attribute value). Mirrors upstream's
		// `report(firstInLine, …)` — the diagnostic line/col reports on
		// the anchor token, NOT on the JsxAttribute.
		reportAttribute := func(anchorPos int, needed, gotten int, fixRange core.TextRange, fixText string) {
			anchorRange := core.NewTextRange(anchorPos, anchorPos+1)
			ctx.ReportRangeWithFixes(anchorRange, reactutil.WrongIndentMessage(needed, gotten, indentType), rule.RuleFix{
				Text:  fixText,
				Range: fixRange,
			})
		}

		// jsxOperandPosition returns the AST node that occupies the
		// "operand position" upstream's ESTree-walking checks reach via
		// `node.parent`. For a wrapping JsxElement / JsxFragment the
		// answer is that element; for a self-closing element the answer
		// is the node itself.
		jsxOperandPosition := func(node *ast.Node) *ast.Node {
			switch node.Kind {
			case ast.KindJsxOpeningElement, ast.KindJsxOpeningFragment:
				return node.Parent
			default:
				return node
			}
		}

		// containerAfterWrappers returns the first non-paren / non-TS-cast
		// ancestor of `operand`. Equivalent to `reactutil.SkipExpressionWrappersUp`
		// rooted at `operand` rather than at `operand.Parent`. We need the
		// child position too (which wrapper is the immediate operand of
		// the resulting container) so the caller can compare with `.Right`,
		// `.WhenFalse`, etc.
		containerAfterWrappers := func(operand *ast.Node) (cur *ast.Node, parent *ast.Node) {
			cur = operand
			parent = cur.Parent
			for parent != nil {
				switch parent.Kind {
				case ast.KindParenthesizedExpression,
					ast.KindAsExpression,
					ast.KindSatisfiesExpression,
					ast.KindNonNullExpression,
					ast.KindTypeAssertionExpression:
					cur = parent
					parent = cur.Parent
					continue
				}
				break
			}
			return cur, parent
		}

		// isRightInLogicalExp — node sits in the right operand of a
		// logical (&&, ||, ??) expression, AND `indentLogicalExpressions`
		// is OFF.
		isRightInLogicalExp := func(node *ast.Node) bool {
			if indentLogicalExpressions {
				return false
			}
			operand := jsxOperandPosition(node)
			if operand == nil {
				return false
			}
			cur, parent := containerAfterWrappers(operand)
			if parent == nil || parent.Kind != ast.KindBinaryExpression {
				return false
			}
			be := parent.AsBinaryExpression()
			op := be.OperatorToken.Kind
			if op != ast.KindAmpersandAmpersandToken &&
				op != ast.KindBarBarToken &&
				op != ast.KindQuestionQuestionToken {
				return false
			}
			return be.Right == cur
		}

		// isAlternateInConditionalExp — node sits in the alternate
		// (whenFalse) branch of a `?:`, AND the previous non-whitespace
		// char before node start is not `(` (upstream skips alternates
		// wrapped in their own parens).
		isAlternateInConditionalExp := func(node *ast.Node) bool {
			operand := jsxOperandPosition(node)
			if operand == nil {
				return false
			}
			cur, parent := containerAfterWrappers(operand)
			if parent == nil || parent.Kind != ast.KindConditionalExpression {
				return false
			}
			ce := parent.AsConditionalExpression()
			if ce.WhenFalse != cur {
				return false
			}
			trimmed := utils.TrimNodeTextRange(ctx.SourceFile, node)
			i := trimmed.Pos() - 1
			for i >= 0 {
				c := text[i]
				if c == ' ' || c == '\t' || c == '\r' || c == '\n' {
					i--
					continue
				}
				return c != '('
			}
			return true
		}

		checkNodesIndent := func(node *ast.Node, indent int) {
			nodeIndent := reactutil.NodeStartIndent(ctx.SourceFile, node, indentChar)
			isCorrectRightInLogicalExp := isRightInLogicalExp(node) && (nodeIndent-indent) == indentSize
			isCorrectAlternateInCondExp := isAlternateInConditionalExp(node) && (nodeIndent-indent) == 0
			if nodeIndent != indent &&
				reactutil.IsNodeFirstInLine(ctx.SourceFile, node) &&
				!isCorrectRightInLogicalExp &&
				!isCorrectAlternateInCondExp {
				reactutil.ReportIndentReplaceLeading(ctx, node, indent, nodeIndent, indentChar, indentType)
			}
		}

		// checkLiteralNodeIndent — for each line of a JsxText that has
		// content, compare its leading indent to the expected value;
		// emit one diagnostic per mismatched line.
		checkLiteralNodeIndent := func(node *ast.Node, expected int) {
			rawStart := node.Pos()
			rawEnd := node.End()
			value := text[rawStart:rawEnd]
			indents := scanLiteralIndents(value, indentChar)
			if len(indents) == 0 {
				return
			}
			allMatch := true
			for _, ind := range indents {
				if ind != expected {
					allMatch = false
					break
				}
			}
			if allMatch {
				return
			}
			indentStr := strings.Repeat(string(indentChar), expected)
			fixedText := replaceLeadingIndentInText(value, indentStr)
			for _, actualIndent := range indents {
				reportLiteral(node, expected, actualIndent, fixedText)
			}
		}

		// commaContainer walks up from the operand position through any
		// expression wrappers and returns the nearest list-y container —
		// the AST node that ESLint's `getNodeByRangeIndex(commaPos)` would
		// resolve to for a comma between siblings.
		commaContainer := func(node *ast.Node) *ast.Node {
			operand := jsxOperandPosition(node)
			if operand == nil {
				return nil
			}
			_, parent := containerAfterWrappers(operand)
			if parent == nil {
				return nil
			}
			switch parent.Kind {
			case ast.KindArrayLiteralExpression,
				ast.KindCallExpression,
				ast.KindNewExpression,
				ast.KindObjectLiteralExpression:
				return parent
			}
			return nil
		}

		// colonAnchor — when the previous source char before `<` is a `:`,
		// upstream rewinds past punctuators (excluding `/`) until it hits
		// a non-punctuator, then climbs the AST until the parent is a
		// ConditionalExpression. The result is whichever
		// ConditionalExpression child contains that token — almost always
		// the consequent. We replicate the same outcome by locating the
		// nearest ConditionalExpression ancestor (skipping paren wrappers)
		// and returning its `WhenTrue` after stripping wrappers.
		colonAnchor := func(node *ast.Node) *ast.Node {
			operand := jsxOperandPosition(node)
			if operand == nil {
				return nil
			}
			cur := operand
			parent := cur.Parent
			for parent != nil {
				if parent.Kind == ast.KindConditionalExpression {
					ce := parent.AsConditionalExpression()
					if ce.WhenFalse == cur {
						return reactutil.SkipExpressionWrappers(ce.WhenTrue)
					}
					return nil
				}
				switch parent.Kind {
				case ast.KindParenthesizedExpression,
					ast.KindAsExpression,
					ast.KindSatisfiesExpression,
					ast.KindNonNullExpression,
					ast.KindTypeAssertionExpression:
					cur = parent
					parent = cur.Parent
					continue
				}
				return nil
			}
			return nil
		}

		// jsxParentOpening returns the opening tag of the surrounding
		// JsxElement / JsxFragment when `node` (or the JsxElement that
		// wraps it) is positioned as one of that container's children. In
		// upstream this is the JSXText/parent re-mapping branch:
		//
		//   if (prevToken.type === 'JSXText' || ...) {
		//     prevToken = sourceCode.getNodeByRangeIndex(prevToken.range[0]);
		//     prevToken = prevToken.type === 'Literal' || prevToken.type === 'JSXText'
		//       ? prevToken.parent : prevToken;
		//   }
		//
		// Same idea — when a JSX child is preceded by JsxText (whitespace
		// or content), the anchor is the surrounding element's opening
		// tag, not the previous sibling's tail.
		jsxParentOpening := func(node *ast.Node) *ast.Node {
			operand := jsxOperandPosition(node)
			if operand == nil {
				return nil
			}
			parent := operand.Parent
			if parent == nil {
				return nil
			}
			switch parent.Kind {
			case ast.KindJsxElement:
				return parent.AsJsxElement().OpeningElement
			case ast.KindJsxFragment:
				return parent.AsJsxFragment().OpeningFragment
			}
			return nil
		}

		// previousAnchorIndent returns (parentIndent, sameLine) for the
		// "previous token" before node's opening `<`. Order of dispatch
		// mirrors upstream:
		//
		//  1. If we sit inside a JsxElement/JsxFragment as a child, the
		//     anchor is the parent's opening tag (upstream JSXText
		//     remap).
		//  2. Otherwise scan back to the previous non-whitespace source
		//     char: comma → list-container, `:` → conditional's
		//     consequent, anything else → that char's line.
		previousAnchorIndent := func(node *ast.Node) (int, bool, bool) {
			trimmed := utils.TrimNodeTextRange(ctx.SourceFile, node)
			startPos := trimmed.Pos()

			if opening := jsxParentOpening(node); opening != nil {
				openingTrimmed := utils.TrimNodeTextRange(ctx.SourceFile, opening)
				sameLine := lineOfPos(openingTrimmed.Pos()) == lineOfPos(startPos)
				return reactutil.IndentLeading(text, lineMap, openingTrimmed.Pos(), indentChar), sameLine, true
			}

			i := startPos - 1
			for i >= 0 && (text[i] == ' ' || text[i] == '\t' || text[i] == '\r' || text[i] == '\n') {
				i--
			}
			if i < 0 {
				return 0, false, false
			}
			if text[i] == ',' {
				container := commaContainer(node)
				if container != nil {
					containerTrimmed := utils.TrimNodeTextRange(ctx.SourceFile, container)
					sameLine := lineOfPos(containerTrimmed.Pos()) == lineOfPos(startPos)
					return reactutil.IndentLeading(text, lineMap, containerTrimmed.Pos(), indentChar), sameLine, true
				}
				i--
				for i >= 0 && (text[i] == ' ' || text[i] == '\t' || text[i] == '\r' || text[i] == '\n') {
					i--
				}
				if i < 0 {
					return 0, false, true
				}
				return reactutil.IndentLeading(text, lineMap, i, indentChar), false, true
			}
			if text[i] == ':' {
				anchor := colonAnchor(node)
				if anchor != nil {
					anchorTrimmed := utils.TrimNodeTextRange(ctx.SourceFile, anchor)
					sameLine := lineOfPos(anchorTrimmed.Pos()) == lineOfPos(startPos)
					return reactutil.IndentLeading(text, lineMap, anchorTrimmed.Pos(), indentChar), sameLine, true
				}
				// Fall through: no enclosing conditional — use the `:`
				// position.
			}
			sameLine := lineOfPos(i) == lineOfPos(startPos)
			return reactutil.IndentLeading(text, lineMap, i, indentChar), sameLine, true
		}

		handleOpeningElement := func(node *ast.Node) {
			parentIndent, sameLine, ok := previousAnchorIndent(node)
			if !ok {
				return
			}
			additional := indentSize
			if sameLine ||
				isRightInLogicalExp(node) ||
				isAlternateInConditionalExp(node) {
				additional = 0
			}
			checkNodesIndent(node, parentIndent+additional)
		}

		handleClosingElement := func(node *ast.Node) {
			parent := node.Parent
			if parent == nil {
				return
			}
			var openingNode *ast.Node
			switch parent.Kind {
			case ast.KindJsxElement:
				openingNode = parent.AsJsxElement().OpeningElement
			case ast.KindJsxFragment:
				openingNode = parent.AsJsxFragment().OpeningFragment
			default:
				return
			}
			peerIndent := reactutil.NodeStartIndent(ctx.SourceFile, openingNode, indentChar)
			checkNodesIndent(node, peerIndent)
		}

		handleAttribute := func(node *ast.Node) {
			if !checkAttributes {
				return
			}
			attr := node.AsJsxAttribute()
			if attr == nil || attr.Initializer == nil || attr.Initializer.Kind != ast.KindJsxExpression {
				return
			}
			nameNode := attr.Name()
			if nameNode == nil {
				return
			}
			value := attr.Initializer
			je := value.AsJsxExpression()
			if je == nil || je.Expression == nil {
				return
			}
			// Replicate upstream's `getFirstNodeInLine(lastToken)`: walk
			// back from `}` (the JsxExpression's last char) over all
			// trivia (whitespace, newlines) to land on the previous
			// non-trivia source byte. That byte is the anchor; we report
			// against the line containing it.
			closeBracePos := value.End() - 1
			i := closeBracePos - 1
			for i >= 0 && (text[i] == ' ' || text[i] == '\t' || text[i] == '\r' || text[i] == '\n') {
				i--
			}
			if i < 0 {
				return
			}
			anchorPos := i
			lineStart := reactutil.IndentLineStart(lineMap, anchorPos)
			actualIndent := 0
			for j := lineStart; j < anchorPos; j++ {
				if text[j] != indentChar {
					break
				}
				actualIndent++
			}
			// The anchor must itself be first-in-line (after WS / commas)
			// to be a meaningful indent target — mirrors checkNodesIndent's
			// `isNodeFirstInLine` gate.
			isFirst := true
			for j := anchorPos - 1; j >= lineStart; j-- {
				c := text[j]
				if c != ' ' && c != '\t' && c != ',' && c != '\r' {
					isFirst = false
					break
				}
			}
			if !isFirst {
				return
			}
			nameLine := lineOfPos(nameNode.Pos())
			anchorLine := lineOfPos(anchorPos)
			expectedIndent := reactutil.NodeStartIndent(ctx.SourceFile, nameNode, indentChar)
			if nameLine == anchorLine {
				expectedIndent = 0
			}
			if actualIndent == expectedIndent {
				return
			}
			indentStr := strings.Repeat(string(indentChar), expectedIndent)
			fixRange := core.NewTextRange(lineStart, anchorPos)
			reportAttribute(anchorPos, expectedIndent, actualIndent, fixRange, indentStr)
		}

		handleJsxExpression := func(node *ast.Node) {
			parent := node.Parent
			if parent == nil {
				return
			}
			parentIndent := reactutil.NodeStartIndent(ctx.SourceFile, parent, indentChar)
			checkNodesIndent(node, parentIndent+indentSize)
		}

		handleJsxText := func(node *ast.Node) {
			parent := node.Parent
			if parent == nil {
				return
			}
			if parent.Kind != ast.KindJsxElement && parent.Kind != ast.KindJsxFragment {
				return
			}
			parentIndent := reactutil.NodeStartIndent(ctx.SourceFile, parent, indentChar)
			checkLiteralNodeIndent(node, parentIndent+indentSize)
		}

		handleReturn := func(node *ast.Node) {
			ret := node.AsReturnStatement()
			if ret == nil || ret.Expression == nil {
				return
			}
			// Mirror upstream's `jsxUtil.isJSX(node.argument)` gate
			// exactly: only JsxElement / JsxFragment / JsxSelfClosingElement
			// qualifies; TS wrappers (`as`, `satisfies`, `!`,
			// `<T>x`) deliberately do NOT, since ESTree exposes them
			// as their own node types and `isJSX` returns false. We do
			// strip plain parens because ESTree flattens those.
			arg := ast.SkipParentheses(ret.Expression)
			if !reactutil.IsJsxLike(arg) {
				return
			}
			fn := node.Parent
			for fn != nil &&
				fn.Kind != ast.KindFunctionDeclaration &&
				fn.Kind != ast.KindFunctionExpression {
				fn = fn.Parent
			}
			if fn == nil {
				return
			}
			openingIndent := reactutil.NodeStartIndent(ctx.SourceFile, node, indentChar)
			closingIndent := reactutil.NodeEndIndent(ctx.SourceFile, node, indentChar)
			if openingIndent != closingIndent {
				reportReturn(node, openingIndent, closingIndent)
			}
		}

		return rule.RuleListeners{
			ast.KindJsxOpeningElement:     handleOpeningElement,
			ast.KindJsxSelfClosingElement: handleOpeningElement,
			ast.KindJsxOpeningFragment:    handleOpeningElement,
			ast.KindJsxClosingElement:     handleClosingElement,
			ast.KindJsxClosingFragment:    handleClosingElement,
			ast.KindJsxAttribute:          handleAttribute,
			ast.KindJsxExpression:         handleJsxExpression,
			ast.KindJsxText:               handleJsxText,
			ast.KindReturnStatement:       handleReturn,
		}
	},
}

// parseIndentOption parses the first option:
//   - "tab" → ("tab", 1, '\t')
//   - integer N → ("space", N, ' ')
//   - default → ("space", 4, ' ')
func parseIndentOption(options any) (string, int, byte) {
	indentType := "space"
	indentSize := 4
	indentChar := byte(' ')

	var first any
	if options == nil {
		return indentType, indentSize, indentChar
	}
	if arr, ok := options.([]interface{}); ok {
		if len(arr) > 0 {
			first = arr[0]
		}
	} else {
		first = options
	}

	switch v := first.(type) {
	case string:
		if v == "tab" {
			indentType = "tab"
			indentSize = 1
			indentChar = '\t'
		}
	case float64:
		indentType = "space"
		indentSize = int(v)
		indentChar = ' '
	case int:
		indentType = "space"
		indentSize = v
		indentChar = ' '
	}
	return indentType, indentSize, indentChar
}

// parseSecondOption parses the second-position object:
//
//	{ checkAttributes: bool, indentLogicalExpressions: bool }
func parseSecondOption(options any) (bool, bool) {
	checkAttributes := false
	indentLogicalExpressions := false

	arr, ok := options.([]interface{})
	if !ok || len(arr) < 2 {
		return checkAttributes, indentLogicalExpressions
	}
	m, ok := arr[1].(map[string]interface{})
	if !ok {
		return checkAttributes, indentLogicalExpressions
	}
	if v, ok := m["checkAttributes"].(bool); ok {
		checkAttributes = v
	}
	if v, ok := m["indentLogicalExpressions"].(bool); ok {
		indentLogicalExpressions = v
	}
	return checkAttributes, indentLogicalExpressions
}

// scanLiteralIndents extracts the leading-indent run of every "content line"
// of a JsxText literal. A content line is a line that has at least one
// non-whitespace character; the run is consecutive `indentChar` characters
// at line start (mirroring upstream's `\n( *)[\t ]*\S` regex).
//
// Upstream applies its regex against the JSXText *decoded* `value`, which
// turns `&nbsp;` / `&#160;` / `&#xA0;` and a literal NBSP byte sequence
// into U+00A0 — and U+00A0 matches JS `\s`, so a line whose only content
// is an NBSP-equivalent entity does NOT count as a content line. tsgo
// keeps `JsxText.Text` raw, so we explicitly skip those entities here to
// stay aligned.
func scanLiteralIndents(value string, indentChar byte) []int {
	var out []int
	for i := range len(value) {
		if value[i] != '\n' {
			continue
		}
		j := i + 1
		count := 0
		for j < len(value) && value[j] == indentChar {
			count++
			j++
		}
		j = skipWhitespaceAndNBSPEntities(value, j)
		if j >= len(value) {
			continue
		}
		if value[j] == '\n' || value[j] == '\r' {
			continue
		}
		out = append(out, count)
	}
	return out
}

// skipWhitespaceAndNBSPEntities advances past ASCII spaces, tabs, raw
// UTF-8 NBSP bytes (\xc2\xa0), and the case-insensitive HTML entities that
// decode to U+00A0: `&nbsp;`, `&#160;`, `&#xA0;`. Upstream's regex
// `[\t ]*` runs against the decoded value where these are already real
// whitespace chars; we replicate that by recognizing them in raw source.
func skipWhitespaceAndNBSPEntities(s string, i int) int {
	for i < len(s) {
		c := s[i]
		if c == ' ' || c == '\t' {
			i++
			continue
		}
		if i+1 < len(s) && c == '\xc2' && s[i+1] == '\xa0' {
			i += 2
			continue
		}
		if c == '&' {
			if n := matchNBSPEntity(s, i); n > 0 {
				i += n
				continue
			}
		}
		break
	}
	return i
}

// matchNBSPEntity returns the byte length of an NBSP-equivalent HTML
// entity reference starting at s[i] ('&'), or 0 if no match. Recognises
// `&nbsp;` (any case), `&#160;`, and `&#xA0;` / `&#xa0;` / `&#XA0;`.
func matchNBSPEntity(s string, i int) int {
	if i >= len(s) || s[i] != '&' {
		return 0
	}
	rest := s[i:]
	if len(rest) >= 6 {
		hd := rest[:6]
		if hd[0] == '&' && (hd[1] == 'n' || hd[1] == 'N') && (hd[2] == 'b' || hd[2] == 'B') &&
			(hd[3] == 's' || hd[3] == 'S') && (hd[4] == 'p' || hd[4] == 'P') && hd[5] == ';' {
			return 6
		}
		if hd == "&#160;" {
			return 6
		}
	}
	if len(rest) >= 6 {
		hd := rest[:6]
		if hd[0] == '&' && hd[1] == '#' && (hd[2] == 'x' || hd[2] == 'X') &&
			(hd[3] == 'A' || hd[3] == 'a') && (hd[4] == '0') && hd[5] == ';' {
			return 6
		}
	}
	return 0
}

// replaceLeadingIndentInText rewrites every `\n<ws>*<non-ws>` run so the
// post-newline whitespace is exactly `indent`. Mirrors upstream's
// `\n[\t ]*(\S)` → `\n<indent>$1` substitution: a trailing `\n<ws>+` with
// no following non-whitespace char (e.g. the JsxText that closes a
// container, ending right before `</tag>`) is preserved verbatim — the
// upstream regex doesn't match, so the indent stays as-is. Lines whose
// only post-indent content is an NBSP-equivalent entity (`&nbsp;` etc.)
// are also left alone, mirroring upstream where `&nbsp;` decodes to
// U+00A0 (whitespace) and the regex doesn't match.
func replaceLeadingIndentInText(s, indent string) string {
	var b strings.Builder
	i := 0
	for i < len(s) {
		c := s[i]
		b.WriteByte(c)
		if c != '\n' {
			i++
			continue
		}
		i++
		wsStart := i
		for i < len(s) && (s[i] == ' ' || s[i] == '\t') {
			i++
		}
		afterWs := skipWhitespaceAndNBSPEntities(s, i)
		if afterWs >= len(s) || s[afterWs] == '\n' || s[afterWs] == '\r' {
			b.WriteString(s[wsStart:i])
			continue
		}
		b.WriteString(indent)
	}
	return b.String()
}

// replaceFirstLeadingIndent replaces only the first `\n<ws>*<non-ws>` run.
// Used for the ReturnStatement last-line fix.
func replaceFirstLeadingIndent(s, indent string) string {
	if len(s) == 0 || s[0] != '\n' {
		return s
	}
	i := 1
	for i < len(s) && (s[i] == ' ' || s[i] == '\t') {
		i++
	}
	if i >= len(s) || s[i] == '\n' || s[i] == '\r' {
		return s
	}
	return "\n" + indent + s[i:]
}
