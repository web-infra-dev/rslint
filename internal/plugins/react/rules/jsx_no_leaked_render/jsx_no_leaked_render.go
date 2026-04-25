package jsx_no_leaked_render

import (
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/react/reactutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

const noPotentialLeakedRenderMessage = "Potential leaked value that might cause unintentionally rendered values or rendering crashes"

const (
	strategyCoerce  = "coerce"
	strategyTernary = "ternary"
)

type ruleOptions struct {
	// validStrategies is ordered; the first element is the preferred fix
	// strategy (matches upstream's `find(from(set), () => true)`).
	validStrategies  []string
	ignoreAttributes bool
}

func defaultOptions() ruleOptions {
	return ruleOptions{
		validStrategies:  []string{strategyTernary, strategyCoerce},
		ignoreAttributes: false,
	}
}

func parseOptions(raw any) ruleOptions {
	opts := defaultOptions()
	m := utils.GetOptionsMap(raw)
	if m == nil {
		return opts
	}
	if v, ok := m["ignoreAttributes"].(bool); ok {
		opts.ignoreAttributes = v
	}
	if v, ok := m["validStrategies"].([]interface{}); ok {
		var strategies []string
		seen := map[string]bool{}
		for _, item := range v {
			s, ok := item.(string)
			if !ok {
				continue
			}
			if s != strategyCoerce && s != strategyTernary {
				continue
			}
			if seen[s] {
				continue
			}
			seen[s] = true
			strategies = append(strategies, s)
		}
		if len(strategies) > 0 {
			opts.validStrategies = strategies
		}
	}
	return opts
}

func hasStrategy(list []string, s string) bool {
	for _, x := range list {
		if x == s {
			return true
		}
	}
	return false
}

var JsxNoLeakedRenderRule = rule.Rule{
	Name: "react/jsx-no-leaked-render",
	Run: func(ctx rule.RuleContext, raw any) rule.RuleListeners {
		opts := parseOptions(raw)
		var fixStrategy string
		if len(opts.validStrategies) > 0 {
			fixStrategy = opts.validStrategies[0]
		}
		coerceAllowed := hasStrategy(opts.validStrategies, strategyCoerce)
		ternaryAllowed := hasStrategy(opts.validStrategies, strategyTernary)
		isReact18Plus := !reactutil.ReactVersionLessThan(ctx.Settings, 18, 0, 0)

		return rule.RuleListeners{
			ast.KindJsxExpression: func(node *ast.Node) {
				je := node.AsJsxExpression()
				if je == nil || je.DotDotDotToken != nil {
					return
				}
				expr := je.Expression
				if expr == nil {
					return
				}
				inner := ast.SkipParentheses(expr)

				// Logical AND: matches upstream
				// `JSXExpressionContainer > LogicalExpression[operator="&&"]`.
				if ast.IsBinaryExpression(inner) {
					bin := inner.AsBinaryExpression()
					if bin.OperatorToken != nil && bin.OperatorToken.Kind == ast.KindAmpersandAmpersandToken {
						handleLogicalAnd(ctx, node, inner, bin, opts, coerceAllowed, isReact18Plus, fixStrategy)
						return
					}
				}

				// Conditional: matches upstream
				// `JSXExpressionContainer > ConditionalExpression`.
				if ast.IsConditionalExpression(inner) {
					handleConditional(ctx, node, inner, opts, ternaryAllowed, fixStrategy)
					return
				}
			},
		}
	},
}

func handleLogicalAnd(
	ctx rule.RuleContext,
	jsxExpr *ast.Node,
	reportedNode *ast.Node,
	bin *ast.BinaryExpression,
	opts ruleOptions,
	coerceAllowed bool,
	isReact18Plus bool,
	fixStrategy string,
) {
	if opts.ignoreAttributes && isWithinJsxAttribute(jsxExpr) {
		return
	}
	leftSide := bin.Left
	leftInner := ast.SkipParentheses(leftSide)

	if coerceAllowed {
		if isCoerceValidLeftSide(leftInner) || isCoerceValidNestedLogical(leftInner) {
			return
		}
		// Identifier resolution: if the binding's initializer is a boolean
		// literal, treat as safe (mirrors upstream's variableUtil scope walk).
		if leftInner.Kind == ast.KindIdentifier && ctx.TypeChecker != nil {
			decl := utils.GetDeclaration(ctx.TypeChecker, leftInner)
			if decl != nil && decl.Kind == ast.KindVariableDeclaration {
				init := decl.AsVariableDeclaration().Initializer
				if init != nil {
					initInner := ast.SkipParentheses(init)
					if initInner.Kind == ast.KindTrueKeyword || initInner.Kind == ast.KindFalseKeyword {
						return
					}
				}
			}
		}
	}

	// React >= 18: empty string literal '' on the left is safe — React 18 no
	// longer renders empty string in JSX.
	if isReact18Plus && leftInner.Kind == ast.KindStringLiteral {
		if leftInner.AsStringLiteral().Text == "" {
			return
		}
	}

	msg := rule.RuleMessage{
		Id:          "noPotentialLeakedRender",
		Description: noPotentialLeakedRenderMessage,
	}
	fix := buildFix(ctx, reportedNode, leftSide, bin.Right, fixStrategy)
	if fix == nil {
		ctx.ReportNode(reportedNode, msg)
		return
	}
	ctx.ReportNodeWithFixes(reportedNode, msg, *fix)
}

func handleConditional(
	ctx rule.RuleContext,
	jsxExpr *ast.Node,
	reportedNode *ast.Node,
	opts ruleOptions,
	ternaryAllowed bool,
	fixStrategy string,
) {
	if ternaryAllowed {
		return
	}
	if opts.ignoreAttributes && isWithinJsxAttribute(jsxExpr) {
		return
	}
	cond := reportedNode.AsConditionalExpression()
	if cond == nil || cond.WhenFalse == nil {
		return
	}
	if !isInvalidTernaryAlternate(cond.WhenFalse) {
		return
	}

	msg := rule.RuleMessage{
		Id:          "noPotentialLeakedRender",
		Description: noPotentialLeakedRenderMessage,
	}
	fix := buildFix(ctx, reportedNode, cond.Condition, cond.WhenTrue, fixStrategy)
	if fix == nil {
		ctx.ReportNode(reportedNode, msg)
		return
	}
	ctx.ReportNodeWithFixes(reportedNode, msg, *fix)
}

// isInvalidTernaryAlternate mirrors upstream's
//
//	const isValidTernaryAlternate = TERNARY_INVALID_ALTERNATE_VALUES.indexOf(node.alternate.value) === -1;
//	const isJSXElementAlternate = node.alternate.type === 'JSXElement';
//	if (isValidTernaryAlternate || isJSXElementAlternate) return;
//
// Upstream skips when the alternate is a non-falsy Literal value or a
// JSXElement; everything else (`null`/`false`, identifiers, calls, …) is
// treated as `value === undefined` and falls through to the report path.
// JSXFragment is intentionally NOT in the JSXElement allowlist — upstream's
// `=== 'JSXElement'` is strict and a fragment alternate would still report.
//
// `KindTrueKeyword` is treated as a non-falsy Literal here even though the
// boolean keywords aren't covered by `ast.IsLiteralKind`; ESTree models them
// as Literal nodes with `value === true`.
func isInvalidTernaryAlternate(alt *ast.Node) bool {
	inner := ast.SkipParentheses(alt)
	switch inner.Kind {
	case ast.KindNullKeyword, ast.KindFalseKeyword:
		return true
	case ast.KindTrueKeyword:
		return false
	case ast.KindJsxElement, ast.KindJsxSelfClosingElement:
		return false
	}
	return !ast.IsLiteralKind(inner.Kind)
}

// isWithinJsxAttribute walks up parents until a JsxElement / JsxSelfClosingElement
// / JsxFragment boundary, returning true if a JsxAttribute is encountered first.
// Mirrors upstream's `isWithinAttribute` (stopTypes = JSXElement, JSXFragment).
// JsxOpeningElement is included as a stop type because in tsgo it owns the
// attributes list — past it, we're outside the current element's attribute
// scope.
func isWithinJsxAttribute(node *ast.Node) bool {
	p := node.Parent
	for p != nil {
		if ast.IsJsxElement(p) || ast.IsJsxSelfClosingElement(p) || ast.IsJsxFragment(p) || ast.IsJsxOpeningElement(p) {
			return false
		}
		if ast.IsJsxAttribute(p) {
			return true
		}
		p = p.Parent
	}
	return false
}

// isCoerceValidLeftSide reports whether the (paren-skipped) node is one of
// upstream's `COERCE_VALID_LEFT_SIDE_EXPRESSIONS`: UnaryExpression,
// BinaryExpression (non-logical, non-assignment, non-comma), or CallExpression.
// TS-only `as` / `satisfies` / `!` non-null wrappers are treated as transparent
// (they don't affect the truthiness of the wrapped expression and have no
// ESTree analog upstream considers).
//
// The BinaryExpression branch must EXCLUDE logical (`&&`/`||`/`??`), assignment,
// and comma operators, which ESTree models as separate node types
// (LogicalExpression / AssignmentExpression / SequenceExpression) but tsgo
// flattens into BinaryExpression.
func isCoerceValidLeftSide(node *ast.Node) bool {
	n := skipTypeAssertions(ast.SkipParentheses(node))
	if isUnaryExpressionLike(n) {
		return true
	}
	if n.Kind == ast.KindCallExpression {
		return true
	}
	if ast.IsBinaryExpression(n) {
		bin := n.AsBinaryExpression()
		if bin.OperatorToken == nil {
			return false
		}
		op := bin.OperatorToken.Kind
		return !ast.IsLogicalOrCoalescingBinaryOperator(op) && !ast.IsAssignmentOperator(op) && op != ast.KindCommaToken
	}
	return false
}

// isCoerceValidNestedLogical mirrors upstream's
// `getIsCoerceValidNestedLogicalExpression`: the node tree, when reduced
// across logical operators, must consist entirely of coerce-valid leaves.
func isCoerceValidNestedLogical(node *ast.Node) bool {
	n := skipTypeAssertions(ast.SkipParentheses(node))
	if ast.IsLogicalOrCoalescingBinaryExpression(n) {
		bin := n.AsBinaryExpression()
		return isCoerceValidNestedLogical(bin.Left) && isCoerceValidNestedLogical(bin.Right)
	}
	return isCoerceValidLeftSide(n)
}

// isUnaryExpressionLike reports whether a tsgo node corresponds to an ESTree
// UnaryExpression (`!`, `~`, `+`, `-`, `typeof`, `void`, `delete`).
// PostfixUnaryExpression (`x++`, `x--`) is excluded — ESTree models those as
// UpdateExpression, which is NOT in upstream's coerce-valid list.
func isUnaryExpressionLike(n *ast.Node) bool {
	switch n.Kind {
	case ast.KindTypeOfExpression, ast.KindVoidExpression, ast.KindDeleteExpression:
		return true
	case ast.KindPrefixUnaryExpression:
		op := n.AsPrefixUnaryExpression().Operator
		return op == ast.KindExclamationToken || op == ast.KindTildeToken ||
			op == ast.KindPlusToken || op == ast.KindMinusToken
	}
	return false
}

// skipTypeAssertions strips TS-only wrappers that don't change the runtime
// expression: `x as T`, `x satisfies T`, `<T>x`, and `x!` non-null assertion.
// Used inside coerce-validity classification so that `(arr.length as number)`
// or `obj!` on the left of `&&` classifies the same as the unwrapped form
// (matching upstream, which sees these as the inner expression because tsgo's
// AST has shapes that ESLint's parser-emitted ESTree treats as Identifier /
// MemberExpression directly).
func skipTypeAssertions(n *ast.Node) *ast.Node {
	for {
		switch n.Kind {
		case ast.KindAsExpression:
			n = n.AsAsExpression().Expression
		case ast.KindSatisfiesExpression:
			n = n.AsSatisfiesExpression().Expression
		case ast.KindTypeAssertionExpression:
			n = n.AsTypeAssertion().Expression
		case ast.KindNonNullExpression:
			n = n.AsNonNullExpression().Expression
		default:
			return n
		}
		n = ast.SkipParentheses(n)
	}
}

// extractAndOperands flattens a left-leaning `&&` chain into its leaf operands,
// preserving any ParenthesizedExpression wrappers on the leaves themselves.
// Parentheses around the chain are skipped for traversal (upstream's ESTree
// is paren-flattened) but the leaf wrapping is retained so the source text
// emitted in the fix matches the input.
func extractAndOperands(node *ast.Node) []*ast.Node {
	skipped := ast.SkipParentheses(node)
	if ast.IsBinaryExpression(skipped) {
		bin := skipped.AsBinaryExpression()
		if bin.OperatorToken != nil && bin.OperatorToken.Kind == ast.KindAmpersandAmpersandToken {
			return append(extractAndOperands(bin.Left), extractAndOperands(bin.Right)...)
		}
	}
	return []*ast.Node{node}
}

// trimDoubleNot strips a leading `!!` boolean coercion. Mirrors upstream's
// `trimLeftNode` — but additionally skips the ParenthesizedExpression that tsgo
// inserts and ESTree flattens, so `!!(a && b)` returns `a && b` (not the
// paren-wrapped form), making the ternary fix reproduce upstream's output.
func trimDoubleNot(node *ast.Node) *ast.Node {
	n := ast.SkipParentheses(node)
	if n.Kind != ast.KindPrefixUnaryExpression {
		return node
	}
	u := n.AsPrefixUnaryExpression()
	if u.Operator != ast.KindExclamationToken {
		return node
	}
	inner := ast.SkipParentheses(u.Operand)
	if inner.Kind != ast.KindPrefixUnaryExpression {
		return node
	}
	u2 := inner.AsPrefixUnaryExpression()
	if u2.Operator != ast.KindExclamationToken {
		return node
	}
	return trimDoubleNot(ast.SkipParentheses(u2.Operand))
}

// buildFix produces the autofix for a report. fixStrategy controls the output
// shape (`coerce` → `!! left && right`, `ternary` → `left ? right : null`).
// reportedNode is the BinaryExpression (`a && b`) or ConditionalExpression
// (`a ? b : c`) being replaced; leftNode and rightNode are its operands.
//
// Returns nil when no safe fix is available (e.g. coerce strategy on a
// rightNode that is itself a Literal — auto-coercing would not eliminate the
// leak) or when the fix would be a no-op (the rule still reports, e.g. the
// inverse-ternary `!!a && !!b ? false : alt` shape, but emitting a no-op fix
// would otherwise feed back into the autofix loop indefinitely).
func buildFix(
	ctx rule.RuleContext,
	reportedNode *ast.Node,
	leftNode *ast.Node,
	rightNode *ast.Node,
	fixStrategy string,
) *rule.RuleFix {
	if fixStrategy == "" {
		return nil
	}
	sf := ctx.SourceFile
	existing := utils.TrimmedNodeText(sf, reportedNode)

	if fixStrategy == strategyTernary {
		trimmed := trimDoubleNot(leftNode)
		leftSideText := utils.TrimmedNodeText(sf, trimmed)
		rightSideText := utils.TrimmedNodeText(sf, rightNode)
		newText := leftSideText + " ? " + rightSideText + " : null"
		if newText == existing {
			return nil
		}
		f := rule.RuleFixReplace(sf, reportedNode, newText)
		return &f
	}

	// COERCE strategy.
	// hasFalseConsequent: special inverse-ternary case
	// (`cond ? false : alt` with validStrategies: ['coerce']). When the
	// reported node is a ConditionalExpression whose consequent is `false`,
	// upstream emits `!cond && alt` instead of `!!cond && false`.
	hasFalseConsequent := false
	if reportedNode.Kind == ast.KindConditionalExpression {
		cond := reportedNode.AsConditionalExpression()
		if cond.WhenTrue != nil {
			consequent := ast.SkipParentheses(cond.WhenTrue)
			if consequent.Kind == ast.KindFalseKeyword {
				hasFalseConsequent = true
			}
		}
	}

	operands := extractAndOperands(leftNode)
	parts := make([]string, 0, len(operands))
	for _, op := range operands {
		opText := utils.TrimmedNodeText(sf, op)
		valid := isCoerceValidNestedLogical(op)
		var prefix string
		if valid {
			prefix = ""
		} else if hasFalseConsequent && operandHasConditionalFalseParent(op) {
			prefix = "!"
		} else {
			prefix = "!!"
		}
		parts = append(parts, prefix+opText)
	}
	newText := strings.Join(parts, " && ")

	if reportedNode.Kind == ast.KindConditionalExpression && hasFalseConsequent {
		cond := reportedNode.AsConditionalExpression()
		// Source text of the alternate (preserves any parens around it).
		alternateText := utils.TrimmedNodeText(sf, cond.WhenFalse)
		// If the test itself was a logical chain, upstream preserves the
		// `? false : alt` shape; otherwise it collapses to `&& alt`.
		isTestLogical := ast.IsLogicalOrCoalescingBinaryExpression(ast.SkipParentheses(cond.Condition))
		var out string
		if isTestLogical {
			consequentText := utils.TrimmedNodeText(sf, cond.WhenTrue)
			out = newText + " ? " + consequentText + " : " + alternateText
		} else {
			out = newText + " && " + alternateText
		}
		if out == existing {
			return nil
		}
		f := rule.RuleFixReplace(sf, reportedNode, out)
		return &f
	}

	rightInner := ast.SkipParentheses(rightNode)
	rightTextWithParens := utils.TrimmedNodeText(sf, rightNode)
	rightInnerText := utils.TrimmedNodeText(sf, rightInner)

	// Right side is a logical / conditional expression — wrap in parens for
	// precedence (upstream rule). When the source already wraps the right side
	// in parens (tsgo preserves them), use the source text as-is to avoid
	// double-wrapping.
	if ast.IsConditionalExpression(rightInner) || ast.IsLogicalOrCoalescingBinaryExpression(rightInner) {
		var out string
		if rightNode != rightInner {
			out = newText + " && " + rightTextWithParens
		} else {
			out = newText + " && (" + rightInnerText + ")"
		}
		if out == existing {
			return nil
		}
		f := rule.RuleFixReplace(sf, reportedNode, out)
		return &f
	}

	// Right side is a multi-line JSX element — preserve indentation by
	// wrapping in `(\n…\n)` with computed leading whitespace, mirroring
	// upstream. Single-line JSX falls through to the default emit.
	if ast.IsJsxElement(rightInner) || ast.IsJsxSelfClosingElement(rightInner) {
		lines := strings.Split(rightInnerText, "\n")
		if len(lines) > 1 {
			lastLine := lines[len(lines)-1]
			indent := leadingWhitespaceLen(lastLine)
			indentStart := strings.Repeat(" ", indent)
			closeIndent := indent - 2
			if closeIndent < 0 {
				closeIndent = 0
			}
			indentClose := strings.Repeat(" ", closeIndent)
			out := newText + " && (\n" + indentStart + rightInnerText + "\n" + indentClose + ")"
			if out == existing {
				return nil
			}
			f := rule.RuleFixReplace(sf, reportedNode, out)
			return &f
		}
	}

	// Right side is a Literal — auto-coercing the left side would still leave
	// the literal as the rendered value; skip the fix entirely (upstream
	// returns null).
	if isLiteralLike(rightInner) {
		return nil
	}

	out := newText + " && " + rightTextWithParens
	if out == existing {
		return nil
	}
	f := rule.RuleFixReplace(sf, reportedNode, out)
	return &f
}

// operandHasConditionalFalseParent reports whether the operand's structural
// (paren-skipped) parent is a ConditionalExpression with `false` consequent.
// This drives upstream's special-case `!` (vs `!!`) prefix in the inverse
// ternary fix.
func operandHasConditionalFalseParent(op *ast.Node) bool {
	p := op.Parent
	for p != nil && p.Kind == ast.KindParenthesizedExpression {
		p = p.Parent
	}
	if p == nil || p.Kind != ast.KindConditionalExpression {
		return false
	}
	cond := p.AsConditionalExpression()
	if cond.WhenTrue == nil {
		return false
	}
	consequent := ast.SkipParentheses(cond.WhenTrue)
	return consequent.Kind == ast.KindFalseKeyword
}

// isLiteralLike mirrors ESTree's `Literal` test: any tsgo literal kind plus
// `true` / `false` / `null` keywords (modeled as Literal in ESTree). Used to
// suppress the coerce fix when the right side is itself a literal — coercing
// the left side wouldn't change the leak.
func isLiteralLike(n *ast.Node) bool {
	if ast.IsLiteralKind(n.Kind) {
		return true
	}
	switch n.Kind {
	case ast.KindTrueKeyword, ast.KindFalseKeyword, ast.KindNullKeyword:
		return true
	}
	return false
}

func leadingWhitespaceLen(s string) int {
	for i, r := range s {
		if r != ' ' && r != '\t' {
			return i
		}
	}
	return len(s)
}
