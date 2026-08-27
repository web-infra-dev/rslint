// Package prefer_ternary mirrors eslint-plugin-unicorn's prefer-ternary rule:
// simple if/else statements that return or assign a value are rewritten to a
// ternary, and a let declaration immediately followed by an if (without else)
// that re-assigns that same variable is folded into a const initialized with
// the ternary. The rule is fully syntactic — it never needs a type checker —
// so it does not declare RequiresTypeInfo.
package prefer_ternary

import (
	_ "embed"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

const (
	messageID            = "prefer-ternary"
	suggestionMessageID  = "prefer-ternary/suggestion"
	optionAlways         = "always"
	optionOnlySingleLine = "only-single-line"
)

//go:embed prefer_ternary.schema.json
var schemaJSON []byte

var messagePreferTernary = rule.RuleMessage{
	Id:          messageID,
	Description: "This `if` statement can be replaced by a ternary expression.",
}

var messageSuggestion = rule.RuleMessage{
	Id:          suggestionMessageID,
	Description: "Use a ternary expression.",
}

// PreferTernaryRule mirrors eslint-plugin-unicorn's prefer-ternary.
var PreferTernaryRule = rule.Rule{
	Name:   "unicorn/prefer-ternary",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		isOnlySingleLine := false
		if len(options) > 0 {
			if opt, ok := options[0].(string); ok && opt == optionOnlySingleLine {
				isOnlySingleLine = true
			}
		}

		// The static evaluator is optional: a nil evaluator just means the
		// computed-key fast path falls back to literal-only static names.
		// The rule stays syntactic on JS/gap files where no TypeChecker is
		// available — literal concat (`x['b' + 'ar']`) still folds through
		// the evaluator's literal handling, but identifier resolution
		// degrades to "unknown".
		var staticStrings *utils.StaticStringEvaluator
		if ctx.TypeChecker != nil {
			staticStrings = utils.NewStaticStringEvaluatorWithReferenceResolver(ctx.TypeChecker, ctx.SourceFile, ctx.Refs)
		}

		return rule.RuleListeners{
			ast.KindIfStatement: func(node *ast.Node) {
				if node.AsIfStatement() == nil {
					return
				}
				checkIfStatement(ctx, node, isOnlySingleLine, staticStrings)
			},
		}
	},
}

// checkIfStatement dispatches to the simple-return/assignment branch when the
// if has an else, otherwise to the let-then-if branch.
func checkIfStatement(ctx rule.RuleContext, node *ast.Node, isOnlySingleLine bool, staticStrings *utils.StaticStringEvaluator) {
	ifStmt := node.AsIfStatement()

	if ifStmt.ElseStatement == nil {
		// Only the let-declaration branch handles an if without an else.
		if problem, ok := checkLetPlusIfProblem(ctx, node, isOnlySingleLine); ok && problem != nil {
			reportLetPlusIfProblem(ctx, problem)
		}
		return
	}

	// `else if (cond) { … }` is represented as an IfStatement whose parent is
	// another IfStatement and whose role is the `ElseStatement`. Skip those:
	// the outer (or, ultimately, the topmost) `if` will be reported if it's
	// mergeable.
	if node.Parent != nil && node.Parent.Kind == ast.KindIfStatement {
		if parentIf := node.Parent.AsIfStatement(); parentIf != nil && parentIf.ElseStatement == node {
			return
		}
	}

	// The test cannot already be a ternary. Parens are transparent in ESTree
	// and must be unwrapped before the check, otherwise `(cond ? a : b)` is
	// a ParenthesizedExpression, not a ConditionalExpression, and would slip
	// through.
	if isConditionalExpression(ast.SkipParentheses(ifStmt.Expression)) {
		return
	}

	consequent := getNodeBody(ifStmt.ThenStatement)
	alternate := getNodeBody(ifStmt.ElseStatement)

	if consequent == nil || alternate == nil {
		return
	}

	if isOnlySingleLine {
		lineStarts := ctx.SourceFile.ECMALineMap()
		// Upstream's getLoc is unaffected by expression-statement wrappers or
		// parens, so it reports `if (t) { x=a; } else { x=b; }` even when the
		// `if` block spans multiple lines (t, x=a, x=b are individually
		// single-line). Mirror that by checking the unwrapped consequent /
		// alternate, and walk parens off the test so a multi-line parenthesized
		// `t` doesn't reject a single-line inner expression.
		testNode := ast.SkipParentheses(ifStmt.Expression)
		if !isSingleLineNode(ctx, testNode, lineStarts) ||
			!isSingleLineNode(ctx, consequent, lineStarts) ||
			!isSingleLineNode(ctx, alternate, lineStarts) {
			return
		}
	}

	result, ok := buildMerge(ctx, consequent, alternate, staticStrings)
	if !ok {
		return
	}

	// Don't fix when there are comments anywhere inside the if.
	if hasCommentsInside(ctx, node) {
		ctx.ReportNode(node, messagePreferTernary)
		return
	}

	ctx.ReportNodeWithDeferredFixes(node, messagePreferTernary, func() []rule.RuleFix {
		return buildFix(ctx, node, result)
	})
}

// ---- getNodeBody / mergeability helpers (mirrors upstream getNodeBody) ----

// getNodeBody unwraps a statement-level node to the meaningful inner
// expression: ExpressionStatement → expression, ParenthesizedExpression →
// its inner expression, Block → the only non-EmptyStatement child
// (recursively). Other shapes — including Blocks with multiple statements —
// return the node itself, which is what we want (the mergeability check
// will then reject it).
func getNodeBody(node *ast.Node) *ast.Node {
	if node == nil {
		return nil
	}
	switch node.Kind {
	case ast.KindExpressionStatement:
		return getNodeBody(node.AsExpressionStatement().Expression)
	case ast.KindParenthesizedExpression:
		// Unwrap through THIS node's expression field. ast.WalkUpParenthesizedExpressions
		// climbs past parens toward a value's parent context and can return the
		// surrounding ExpressionStatement, which would panic when cast back to a
		// ParenthesizedExpression here. Mirroring upstream `getNodeBody` only
		// requires peeling parens off the current node, not the enclosing one.
		return getNodeBody(node.AsParenthesizedExpression().Expression)
	case ast.KindBlock:
		block := node.AsBlock()
		if block == nil || block.Statements == nil {
			return node
		}
		filtered := make([]*ast.Node, 0, len(block.Statements.Nodes))
		for _, stmt := range block.Statements.Nodes {
			if stmt.Kind != ast.KindEmptyStatement {
				filtered = append(filtered, stmt)
			}
		}
		if len(filtered) == 1 {
			return getNodeBody(filtered[0])
		}
	}
	return node
}

// isConditionalExpression mirrors upstream's `isTernary` helper.
func isConditionalExpression(node *ast.Node) bool {
	return node != nil && node.Kind == ast.KindConditionalExpression
}

// isBooleanLiteral reports whether node is the literal `true` or `false`.
func isBooleanLiteral(node *ast.Node) bool {
	return node != nil && (node.Kind == ast.KindTrueKeyword || node.Kind == ast.KindFalseKeyword)
}

// isSingleLineNode reports whether node sits on a single source line, matching
// upstream's `isSingleLineNode`.
func isSingleLineNode(ctx rule.RuleContext, node *ast.Node, lineStarts []core.TextPos) bool {
	if node == nil {
		return false
	}
	r := utils.TrimNodeTextRange(ctx.SourceFile, node)
	return scanner.ComputeLineOfPosition(lineStarts, r.Pos()) == scanner.ComputeLineOfPosition(lineStarts, r.End())
}

// isMergeableReturnStatement matches the `return <expr>` / `return;` shape on
// both sides, excluding cases upstream explicitly avoids: ternary arguments
// (would loop) and both sides being boolean literals.
func isMergeableReturnStatement(consequent, alternate *ast.Node) bool {
	if consequent.Kind != ast.KindReturnStatement || alternate.Kind != ast.KindReturnStatement {
		return false
	}
	c := consequent.AsReturnStatement()
	if c == nil {
		return false
	}
	a := alternate.AsReturnStatement()
	if a == nil {
		return false
	}
	if isConditionalExpression(c.Expression) || isConditionalExpression(a.Expression) {
		return false
	}
	if isBooleanLiteral(c.Expression) && isBooleanLiteral(a.Expression) {
		return false
	}
	return true
}

// isMergeableAssignmentExpression matches the `x = …` shape on both sides.
// Requires identical `left` and `operator`, and no ternary anywhere — those
// shape gates mirror upstream. The `left` comparison goes through
// isSameAssignmentTarget, which extends utils.IsSameReference with a static
// evaluator so that `(x)['b' + 'ar']` matches `x.bar` (the literal-only
// `AccessExpressionStaticName` used by IsSameReference cannot fold string
// concatenation).
func isMergeableAssignmentExpression(consequent, alternate *ast.Node, staticStrings *utils.StaticStringEvaluator) bool {
	if consequent.Kind != ast.KindBinaryExpression || alternate.Kind != ast.KindBinaryExpression {
		return false
	}
	c := consequent.AsBinaryExpression()
	if c == nil {
		return false
	}
	a := alternate.AsBinaryExpression()
	if a == nil {
		return false
	}
	if c.OperatorToken == nil || a.OperatorToken == nil {
		return false
	}
	if c.OperatorToken.Kind != a.OperatorToken.Kind {
		return false
	}
	if isConditionalExpression(c.Left) || isConditionalExpression(a.Left) {
		return false
	}
	if isConditionalExpression(c.Right) || isConditionalExpression(a.Right) {
		return false
	}
	return isSameAssignmentTarget(staticStrings, c.Left, a.Left)
}

// isSameAssignmentTarget mirrors utils.IsSameReference for the assignment LHS
// comparison, but uses a static evaluator for the computed-name fast path so
// that e.g. (x)['b' + 'ar'] and x.bar are considered the same reference. The
// fallback to utils.IsSameReference handles literal-only computed keys and
// structural comparison for non-access nodes.
func isSameAssignmentTarget(staticStrings *utils.StaticStringEvaluator, left, right *ast.Node) bool {
	if left == nil || right == nil {
		return false
	}
	left = ast.SkipOuterExpressions(left, ast.OEKParentheses|ast.OEKAssertions)
	right = ast.SkipOuterExpressions(right, ast.OEKParentheses|ast.OEKAssertions)
	if left == nil || right == nil {
		return false
	}

	if ast.IsAccessExpression(left) && ast.IsAccessExpression(right) && staticStrings != nil {
		// Try the static-name fast path first. This handles the case where one
		// side uses `x.bar` and the other uses `(x)['b' + 'ar']` — both fold to
		// the same property name through the static evaluator.
		leftName, leftOK := staticAccessName(staticStrings, left)
		if leftOK {
			rightName, rightOK := staticAccessName(staticStrings, right)
			if rightOK && leftName == rightName {
				return isSameAssignmentTarget(staticStrings, utils.AccessExpressionObject(left), utils.AccessExpressionObject(right))
			}
			return false
		}
	}

	return utils.IsSameReference(left, right, false)
}

// staticAccessName returns the static property name of an access expression,
// evaluating a computed argument through the static evaluator so that
// `x['b' + 'ar']` and `x.bar` produce the same name. Mirrors upstream's
// `astUtils.getStaticPropertyName` but with broader computed-key support.
func staticAccessName(staticStrings *utils.StaticStringEvaluator, node *ast.Node) (string, bool) {
	if staticStrings == nil || node == nil {
		return "", false
	}
	switch node.Kind {
	case ast.KindPropertyAccessExpression:
		name := node.AsPropertyAccessExpression().Name()
		if name != nil {
			return name.Text(), true
		}
	case ast.KindElementAccessExpression:
		arg := node.AsElementAccessExpression().ArgumentExpression
		if arg == nil {
			return "", false
		}
		arg = ast.SkipOuterExpressions(arg, ast.OEKParentheses|ast.OEKAssertions)
		if arg == nil {
			return "", false
		}
		// EvalToString mirrors JavaScript's String() coercion, which is what
		// computed property names see at runtime (e.g. `[0]` → "0", `[true]`
		// → "true"). Falls back to false when the argument can't be folded.
		if str, ok := staticStrings.EvalToString(arg); ok {
			return str, true
		}
	}
	return "", false
}

// hasCommentsInside reports whether any comment falls within node's source
// range, matching upstream's `getCommentsInside(node).length > 0`.
func hasCommentsInside(ctx rule.RuleContext, node *ast.Node) bool {
	if node == nil {
		return false
	}
	r := utils.TrimNodeTextRange(ctx.SourceFile, node)
	return utils.HasCommentInSpan(ctx.Comments.All(), r.Pos(), r.End())
}

// ---- `let x = a; if (test) { x = b; }` branch ----

// letPlusIfProblem carries everything the let-then-if branch needs to report:
// the if itself, the previous declaration statement, the declarator with init
// and the reassignment target. The check populates it; reportLetPlusIfProblem
// uses it to build the suggestion.
type letPlusIfProblem struct {
	ifNode         *ast.Node
	declStmt       *ast.Node
	declarator     *ast.Node
	left           *ast.Node // identifier (consequent.Left)
	right          *ast.Node // consequent.Right (the `b`)
	test           *ast.Node
	init           *ast.Node // declarator.Initializer (the `a`)
	hasOtherWrites bool
}

func checkLetPlusIfProblem(ctx rule.RuleContext, node *ast.Node, isOnlySingleLine bool) (*letPlusIfProblem, bool) {
	ifStmt := node.AsIfStatement()
	consequent := getNodeBody(ifStmt.ThenStatement)
	if consequent == nil || consequent.Kind != ast.KindBinaryExpression {
		return nil, false
	}
	consBin := consequent.AsBinaryExpression()
	if consBin == nil || consBin.OperatorToken == nil || consBin.OperatorToken.Kind != ast.KindEqualsToken {
		return nil, false
	}
	left := consBin.Left
	if left == nil || left.Kind != ast.KindIdentifier {
		return nil, false
	}
	if isConditionalExpression(ast.SkipParentheses(ifStmt.Expression)) || isConditionalExpression(consBin.Right) {
		return nil, false
	}

	lineStarts := ctx.SourceFile.ECMALineMap()
	if isOnlySingleLine {
		// Walk parens off the test for the same reason the if/else branch
		// does: a multi-line parenthesized test should be treated as the
		// inner expression for the single-line gate.
		testNode := ast.SkipParentheses(ifStmt.Expression)
		if !isSingleLineNode(ctx, testNode, lineStarts) ||
			!isSingleLineNode(ctx, consBin.Right, lineStarts) {
			return nil, false
		}
	}

	previousNode := getPreviousNode(node)
	if previousNode == nil || previousNode.Kind != ast.KindVariableStatement {
		return nil, false
	}
	declStmt := previousNode.AsVariableStatement()
	if declStmt == nil {
		return nil, false
	}
	// The declaration list must be `let`.
	declList := declStmt.DeclarationList
	if declList == nil || (declList.Flags&ast.NodeFlagsLet) == 0 {
		return nil, false
	}
	vdl := declList.AsVariableDeclarationList()
	if vdl == nil || vdl.Declarations == nil || len(vdl.Declarations.Nodes) != 1 {
		return nil, false
	}
	declarator := vdl.Declarations.Nodes[0]
	declName := declarator.Name()
	if declName == nil || declName.Kind != ast.KindIdentifier {
		return nil, false
	}
	if declName.AsIdentifier().Text != left.AsIdentifier().Text {
		return nil, false
	}
	if declarator.Initializer() == nil || isConditionalExpression(declarator.Initializer()) {
		return nil, false
	}
	if isOnlySingleLine && !isSingleLineNode(ctx, declarator.Initializer(), lineStarts) {
		return nil, false
	}

	// Init with side effects (function call / `new`) must not be re-evaluated
	// in the alternate position.
	if hasSideEffect(declarator.Initializer()) {
		return nil, false
	}

	if ctx.Refs == nil {
		return nil, false
	}
	sym := declarator.Symbol()
	if sym == nil {
		return nil, false
	}
	references := ctx.Refs.References(sym)

	// Variable cannot be referenced inside the test or the right side of the
	// assignment: that would create a different evaluation order.
	testRange := utils.TrimNodeTextRange(ctx.SourceFile, ifStmt.Expression)
	rightRange := utils.TrimNodeTextRange(ctx.SourceFile, consBin.Right)
	for _, ref := range references {
		refRange := utils.TrimNodeTextRange(ctx.SourceFile, ref)
		if rangeContains(testRange, refRange) || rangeContains(rightRange, refRange) {
			return nil, false
		}
	}

	// If there are any writes outside the if statement, the variable still
	// needs to be mutable — keep `let` instead of promoting to `const`.
	hasOtherWrites := false
	ifNodeRange := utils.TrimNodeTextRange(ctx.SourceFile, node)
	for _, ref := range references {
		if !utils.IsWriteReference(ref) {
			continue
		}
		refRange := utils.TrimNodeTextRange(ctx.SourceFile, ref)
		if !rangeContains(ifNodeRange, refRange) {
			hasOtherWrites = true
			break
		}
	}

	return &letPlusIfProblem{
		ifNode:         node,
		declStmt:       previousNode,
		declarator:     declarator,
		left:           left,
		right:          consBin.Right,
		test:           ifStmt.Expression,
		init:           declarator.Initializer(),
		hasOtherWrites: hasOtherWrites,
	}, true
}

func reportLetPlusIfProblem(ctx rule.RuleContext, p *letPlusIfProblem) {
	// No suggestion when comments are present anywhere that would be lost.
	if hasCommentsInside(ctx, p.ifNode) ||
		hasCommentsInside(ctx, p.declStmt) ||
		hasCommentsBetween(ctx, p.declStmt, p.ifNode) {
		ctx.ReportNode(p.ifNode, messagePreferTernary)
		return
	}

	keyword := "let"
	if !p.hasOtherWrites {
		keyword = "const"
	}

	ctx.ReportNodeWithDeferredSuggestions(p.ifNode, messagePreferTernary, func() []rule.RuleSuggestion {
		testText := getTextForConditionalChild(ctx, p.test)
		consequentText := getTextForConditionalChild(ctx, p.right)
		alternateText := getTextForConditionalChild(ctx, p.init)
		ternary := testText + " ? " + consequentText + " : " + alternateText

		// 1. Replace the `let` keyword with `const` (or keep `let`). The
		//    keyword is the first non-trivia token of the variable
		//    statement; replacing just it leaves the trailing whitespace
		//    and the rest of the declaration alone.
		kwRange := firstKeywordRange(ctx, p.declStmt)

		// 2. Replace the init expression with the ternary.
		initRange := utils.TrimNodeTextRange(ctx.SourceFile, p.init)
		// 3. Drop the if statement (and any whitespace / trailing semicolon
		//    that bridged the two). A `;` is added only when the original
		//    would have needed one for ASI — that is, when the next token
		//    begins with `[` / `(` / `` ` `` / `/` / `+` / `-` / `,` / `.`.
		declEnd := utils.TrimNodeTextRange(ctx.SourceFile, p.declStmt).End()
		ifEnd := utils.TrimNodeTextRange(ctx.SourceFile, p.ifNode).End()
		dropText := ";"
		if next, ok := utils.TokenAtOrAfter(ctx.SourceFile, ifEnd); ok {
			if !unicornNeedsSemicolonBefore(next.Text) {
				dropText = ""
			}
		} else {
			dropText = ""
		}
		fixes := []rule.RuleFix{
			rule.RuleFixReplaceRange(kwRange, keyword),
			rule.RuleFixReplaceRange(initRange, ternary),
			rule.RuleFixReplaceRange(core.NewTextRange(declEnd, ifEnd), dropText),
		}

		return []rule.RuleSuggestion{{
			Message:  messageSuggestion,
			FixesArr: fixes,
		}}
	})
}

// hasSideEffect reports whether evaluating node can produce a side effect.
// This is a port of @eslint-community/eslint-utils' hasSideEffect with
// default options ({considerGetters: false, considerImplicitTypeConversion:
// false}): only the leaves that are unconditionally side-effect-bearing —
// CallExpression, NewExpression, AwaitExpression, YieldExpression, ++/--, and
// assignments (plain `=`) — trip the predicate. Everything else recurses
// into its children so nested effects (a call inside an object literal, an
// assignment inside parentheses, a `new` inside a binary expression) are
// still caught. ArrowFunction and FunctionExpression short-circuit to false:
// their bodies do not run until invoked.
//
// The non-assignment BinaryExpression case intentionally visits children
// rather than returning true: upstream considers `a + b` and `a ? b : c`
// pure, and the only assignment-shaped binary with a side effect is the
// plain `=` (compound assignments like `+=` also visit children — the
// right-hand side carries the effect, and the LHS write is recursed into
// separately).
func hasSideEffect(node *ast.Node) bool {
	if node == nil {
		return false
	}
	switch node.Kind {
	case ast.KindArrowFunction, ast.KindFunctionExpression:
		// Bodies don't run until called.
		return false
	case ast.KindCallExpression,
		ast.KindNewExpression,
		ast.KindAwaitExpression,
		ast.KindYieldExpression,
		ast.KindPostfixUnaryExpression:
		// PostfixUnaryExpression covers `x++` / `x--`.
		return true
	case ast.KindPrefixUnaryExpression:
		prefix := node.AsPrefixUnaryExpression()
		if prefix == nil {
			return false
		}
		// `delete x`, `++x`, `--x` are always side-effect-bearing. Other
		// prefix operators (`!x`, `+x`, `-x`, `~x`, `typeof x`,
		// `void x`) visit children to surface any nested effect.
		switch prefix.Operator {
		case ast.KindDeleteKeyword, ast.KindPlusPlusToken, ast.KindMinusMinusToken:
			return true
		}
		return hasSideEffectInChildren(node)
	case ast.KindBinaryExpression:
		binary := node.AsBinaryExpression()
		if binary != nil && binary.OperatorToken != nil && binary.OperatorToken.Kind == ast.KindEqualsToken {
			// Plain `=` — the assignment itself is a side effect. Compound
			// assignments (`+=`, etc.) fall through to the child walk below,
			// so the LHS write and the RHS are both recursed into.
			return true
		}
		return hasSideEffectInChildren(node)
	}
	return hasSideEffectInChildren(node)
}

// hasSideEffectInChildren walks the immediate children of node and returns
// true as soon as any of them reports a side effect. Mirrors the visitor
// fallthrough in eslint-utils' hasSideEffect.
func hasSideEffectInChildren(node *ast.Node) bool {
	if node == nil {
		return false
	}
	found := false
	node.ForEachChild(func(child *ast.Node) bool {
		if hasSideEffect(child) {
			found = true
			return true
		}
		return false
	})
	return found
}

func rangeContains(outer, inner core.TextRange) bool {
	return outer.Pos() <= inner.Pos() && outer.End() >= inner.End()
}

// firstKeywordRange returns the range of the `let` / `var` / `const` keyword
// that opens a VariableStatement. The keyword is the first non-trivia token
// of the statement; the scanner gives its exact range, which covers just
// the keyword (no surrounding trivia), so replacing it leaves the trailing
// whitespace and the rest of the statement untouched.
func firstKeywordRange(ctx rule.RuleContext, stmt *ast.Node) core.TextRange {
	if stmt == nil {
		return core.TextRange{}
	}
	declRange := utils.TrimNodeTextRange(ctx.SourceFile, stmt)
	if declRange.End() <= declRange.Pos() {
		return declRange
	}
	return scanner.GetRangeOfTokenAtPosition(ctx.SourceFile, declRange.Pos())
}

// getPreviousNode mirrors upstream's `getPreviousNode`: the preceding sibling
// statement, or nil when `node` is the first in its parent block.
func getPreviousNode(node *ast.Node) *ast.Node {
	parent := node.Parent
	if parent == nil {
		return nil
	}
	// Walk up through Block and SourceFile to find the statement list.
	var list []*ast.Node
	switch parent.Kind {
	case ast.KindBlock:
		block := parent.AsBlock()
		if block == nil || block.Statements == nil {
			return nil
		}
		list = block.Statements.Nodes
	case ast.KindSourceFile:
		sf := parent.AsSourceFile()
		if sf == nil || sf.Statements == nil {
			return nil
		}
		list = sf.Statements.Nodes
	case ast.KindCaseClause, ast.KindDefaultClause:
		clause := parent.AsCaseOrDefaultClause()
		if clause == nil || clause.Statements == nil {
			return nil
		}
		list = clause.Statements.Nodes
	case ast.KindModuleBlock:
		mb := parent.AsModuleBlock()
		if mb == nil || mb.Statements == nil {
			return nil
		}
		list = mb.Statements.Nodes
	default:
		return nil
	}
	for i, n := range list {
		if n == node && i > 0 {
			return list[i-1]
		}
	}
	return nil
}

func hasCommentsBetween(ctx rule.RuleContext, a, b *ast.Node) bool {
	rA := utils.TrimNodeTextRange(ctx.SourceFile, a)
	rB := utils.TrimNodeTextRange(ctx.SourceFile, b)
	return utils.HasCommentInSpan(ctx.Comments.All(), rA.End(), rB.Pos())
}

// ---- Conditional child parenthesization ----

// shouldAddParenthesesToConditionalExpressionChild mirrors
// rules/utils/should-add-parentheses-to-conditional-expression-child.js: the
// listed node kinds all bind more loosely than `?:`, so a child of a
// ConditionalExpression must be wrapped in parens to preserve semantics.
func shouldAddParenthesesToConditionalExpressionChild(node *ast.Node) bool {
	if node == nil {
		return false
	}
	switch node.Kind {
	case ast.KindAwaitExpression,
		ast.KindYieldExpression,
		ast.KindBinaryExpression, // AssignmentExpression
		ast.KindAsExpression,
		ast.KindTypeAssertionExpression:
		return true
	}
	return false
}

// getTextForConditionalChild returns the source text for `node` as it should
// appear inside a conditional, wrapping the result in parens when the node
// isn't already parenthesized and would otherwise change meaning.
func getTextForConditionalChild(ctx rule.RuleContext, node *ast.Node) string {
	if node == nil {
		return ""
	}
	text := getParenthesizedText(ctx, node)
	if !isParenthesizedNode(ctx, node) && shouldAddParenthesesToConditionalExpressionChild(node) {
		text = "(" + text + ")"
	}
	return text
}

// getParenthesizedText returns the source text of the outermost parenthesized
// wrapper of `node` (the node itself when unparenthesized), matching
// upstream's `getParenthesizedText`.
func getParenthesizedText(ctx rule.RuleContext, node *ast.Node) string {
	outer := utils.OutermostParenthesizedExpression(node)
	if outer == nil {
		outer = node
	}
	return scannerGetText(ctx, outer)
}

// isParenthesizedNode reports whether `node` is wrapped in at least one
// ParenthesizedExpression.
func isParenthesizedNode(ctx rule.RuleContext, node *ast.Node) bool {
	return utils.OutermostParenthesizedExpression(node) != node
}

// scannerGetText pulls the raw source text for node, falling back to the
// trimmed range when the scanner can't address it.
func scannerGetText(ctx rule.RuleContext, node *ast.Node) string {
	if node == nil {
		return ""
	}
	r := utils.TrimNodeTextRange(ctx.SourceFile, node)
	if r.Pos() < 0 || r.End() > len(ctx.SourceFile.Text()) {
		return ""
	}
	return ctx.SourceFile.Text()[r.Pos():r.End()]
}

// ---- `unicornNeedsSemicolonBefore` (ASI safety) ----

// unicornNeedsSemicolonBefore reports whether replacing the if-statement with
// the const-declared ternary would need a leading `;` to keep ASI from
// continuing the previous statement. Mirrors upstream's `needsSemicolon` only
// for the slot between the `let x = …` declaration and the if body.
func unicornNeedsSemicolonBefore(nextTokenText string) bool {
	if nextTokenText == "" {
		return false
	}
	switch nextTokenText[0] {
	case '[', '(', '`', '/', '+', '-', ',', '.':
		return true
	}
	return false
}

// ---- buildMerge / buildFix (the simple if/else → ternary branch) ----

// mergeResult is the output of the merge walker: the leading and trailing
// text fragments and the two operand nodes (or string literals for the
// "bare return;" form that the upstream rewrite flattens to `undefined`).
type mergeResult struct {
	before     string
	after      string
	consequent any // *ast.Node or string
	alternate  any // *ast.Node or string
}

// buildMerge walks the two bodies, descending into ReturnStatement and
// AssignmentExpression one level, exactly as upstream's `merge` function does.
// The boolean returned is false when the pair is not mergeable.
func buildMerge(ctx rule.RuleContext, consequent, alternate *ast.Node, staticStrings *utils.StaticStringEvaluator) (mergeResult, bool) {
	if consequent == nil || alternate == nil || consequent.Kind != alternate.Kind {
		return mergeResult{}, false
	}

	if isMergeableReturnStatement(consequent, alternate) {
		c := consequent.AsReturnStatement()
		a := alternate.AsReturnStatement()
		// Bare `return;` is rewritten to the literal `undefined`.
		consequentArg := any(c.Expression)
		if c.Expression == nil {
			consequentArg = "undefined"
		}
		alternateArg := any(a.Expression)
		if a.Expression == nil {
			alternateArg = "undefined"
		}
		return mergeResult{
			before:     "return ",
			after:      ";",
			consequent: consequentArg,
			alternate:  alternateArg,
		}, true
	}

	if isMergeableAssignmentExpression(consequent, alternate, staticStrings) {
		c := consequent.AsBinaryExpression()
		a := alternate.AsBinaryExpression()
		// The merge's `before` and the consequent/alternate operands need to
		// be aligned with what upstream's `consequent.left` exposes. For a
		// chain like `$0 |= $1 ^= $2 &= … = _STOP_ = … = 1`, tsgo parses
		// the outermost operator as `|=`, but the upstream test expects the
		// LHS to be the entire chain and the operator to be the first `=`
		// that follows it. Walk down the right side of the chain to the
		// first `=` (or any plain assignment operator) and use its LHS
		// and right operand instead.
		leftText := getParenthesizedText(ctx, c.Left)
		rightNode := c.Right
		opNode := c.OperatorToken
		cur := c
		for cur.Right != nil && cur.Right.Kind == ast.KindBinaryExpression {
			rb := cur.Right.AsBinaryExpression()
			if rb.OperatorToken == nil {
				break
			}
			if !isAssignmentOperatorKind(rb.OperatorToken.Kind) {
				break
			}
			// If the operator is `=` (plain assignment), that's our target:
			// the chain above it becomes the new LHS and the right operand
			// becomes the new consequent. Compound operators (`|=`, `^=`,
			// `&=`, etc.) keep walking down because the chain continues.
			if rb.OperatorToken.Kind == ast.KindEqualsToken {
				// Use the source text from the start of the outermost
				// assignment to the position of this `=` operator, so
				// the chain's compound assignment operators (which are
				// right-nested under the `|=` in tsgo's parse) are
				// included in the LHS prefix. Trim any trailing
				// whitespace so the operator separator below doesn't
				// produce a double space.
				outerStart := utils.TrimNodeTextRange(ctx.SourceFile, c.AsNode()).Pos()
				equalsPos := utils.TrimNodeTextRange(ctx.SourceFile, rb.OperatorToken).Pos()
				leftText = strings.TrimRight(ctx.SourceFile.Text()[outerStart:equalsPos], " \t")
				rightNode = rb.Right
				opNode = rb.OperatorToken
				break
			}
			cur = rb
		}
		return mergeResult{
			before:     leftText + " " + operatorText(opNode) + " ",
			after:      ";",
			consequent: rightNode,
			alternate:  findMatchingRight(a, opNode),
		}, true
	}

	return mergeResult{}, false
}

// isAssignmentOperatorKind reports whether the given kind is one of the
// assignment operators that share the same precedence as `=` (used to walk
// down a chain like `a |= b ^= c = d` to find the outermost plain `=`).
func isAssignmentOperatorKind(kind ast.Kind) bool {
	switch kind {
	case ast.KindEqualsToken,
		ast.KindPlusEqualsToken,
		ast.KindMinusEqualsToken,
		ast.KindAsteriskEqualsToken,
		ast.KindAsteriskAsteriskEqualsToken,
		ast.KindSlashEqualsToken,
		ast.KindPercentEqualsToken,
		ast.KindLessThanLessThanEqualsToken,
		ast.KindGreaterThanGreaterThanEqualsToken,
		ast.KindGreaterThanGreaterThanGreaterThanEqualsToken,
		ast.KindAmpersandEqualsToken,
		ast.KindBarEqualsToken,
		ast.KindCaretEqualsToken,
		ast.KindBarBarEqualsToken,
		ast.KindAmpersandAmpersandEqualsToken,
		ast.KindQuestionQuestionEqualsToken:
		return true
	}
	return false
}

// findMatchingRight walks down the right side of `start` following the same
// chain shape used for the consequent, returning the right operand at the
// matching depth.
func findMatchingRight(start *ast.BinaryExpression, op *ast.Node) *ast.Node {
	cur := start
	for cur.Right != nil && cur.Right.Kind == ast.KindBinaryExpression {
		rb := cur.Right.AsBinaryExpression()
		if rb.OperatorToken == nil {
			break
		}
		if !isAssignmentOperatorKind(rb.OperatorToken.Kind) {
			break
		}
		if rb.OperatorToken.Kind == op.Kind {
			return rb.Right
		}
		cur = rb
	}
	return cur.Right
}

// what upstream reads off the `consequent.operator` field. tsgo keeps the
// operator in an internal Token wrapper that has no Text() method, so we map
// the kind to its canonical string for the assignment operators this rule
// actually rewrites.
func operatorText(tok *ast.Node) string {
	if tok == nil {
		return ""
	}
	switch tok.Kind {
	case ast.KindEqualsToken:
		return "="
	case ast.KindPlusEqualsToken:
		return "+="
	case ast.KindMinusEqualsToken:
		return "-="
	case ast.KindAsteriskEqualsToken:
		return "*="
	case ast.KindAsteriskAsteriskEqualsToken:
		return "**="
	case ast.KindSlashEqualsToken:
		return "/="
	case ast.KindPercentEqualsToken:
		return "%="
	case ast.KindLessThanLessThanEqualsToken:
		return "<<="
	case ast.KindGreaterThanGreaterThanEqualsToken:
		return ">>="
	case ast.KindGreaterThanGreaterThanGreaterThanEqualsToken:
		return ">>>="
	case ast.KindAmpersandEqualsToken:
		return "&="
	case ast.KindBarEqualsToken:
		return "|="
	case ast.KindCaretEqualsToken:
		return "^="
	case ast.KindBarBarEqualsToken:
		return "||="
	case ast.KindAmpersandAmpersandEqualsToken:
		return "&&="
	case ast.KindQuestionQuestionEqualsToken:
		return "??="
	}
	return ""
}

// buildFix assembles the autofix and yields a slice of `RuleFix`es. The caller
// already checked the no-comments-in-if body, so this is purely string work.
func buildFix(ctx rule.RuleContext, ifNode *ast.Node, result mergeResult) []rule.RuleFix {
	testText := getTextForConditionalChild(ctx, ifNode.AsIfStatement().Expression)

	consequentText := renderMergeOperand(ctx, result.consequent)
	alternateText := renderMergeOperand(ctx, result.alternate)

	fixed := result.before + testText + " ? " + consequentText + " : " + alternateText + result.after

	// ASI safety: prepend `;` when the previous token would absorb the
	// replacement into the prior statement.
	ifPrev := utils.TrimNodeTextRange(ctx.SourceFile, ifNode).Pos()
	if prev, ok := utils.TokenBeforePosition(ctx.SourceFile, ifPrev); ok {
		if needsSemicolonBeforeAfterText(prev, fixed) {
			fixed = ";" + fixed
		}
	}

	return []rule.RuleFix{rule.RuleFixReplace(ctx.SourceFile, ifNode, fixed)}
}

// renderMergeOperand returns the source text for a merge operand. A string
// operand is a literal placeholder produced by the bare-`return;` unwrap
// (`"undefined"`); a node operand flows through the conditional-child
// parenthesization.
func renderMergeOperand(ctx rule.RuleContext, operand any) string {
	if s, ok := operand.(string); ok {
		return s
	}
	if n, ok := operand.(*ast.Node); ok {
		return getTextForConditionalChild(ctx, n)
	}
	return ""
}

// needsSemicolonBeforeAfterText mirrors unicorn's `needsSemicolon` for the
// fixed-output case: does the result, when placed where the if was, need a
// leading `;` to prevent ASI from continuing the prior statement?
func needsSemicolonBeforeAfterText(prev utils.SourceToken, replacement string) bool {
	if replacement == "" {
		return false
	}
	switch replacement[0] {
	case '[', '(', '`', '/', '+', '-', ',', '.':
		// ok, fall through
	default:
		return false
	}
	switch prev.Kind {
	case ast.KindCloseBracketToken,
		ast.KindCloseParenToken,
		ast.KindIdentifier,
		ast.KindStringLiteral,
		ast.KindNumericLiteral,
		ast.KindBigIntLiteral,
		ast.KindRegularExpressionLiteral,
		ast.KindNoSubstitutionTemplateLiteral,
		ast.KindTemplateTail,
		ast.KindTrueKeyword,
		ast.KindFalseKeyword,
		ast.KindNullKeyword:
		return true
	}
	return false
}
