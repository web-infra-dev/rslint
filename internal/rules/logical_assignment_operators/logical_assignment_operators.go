package logical_assignment_operators

import (
	_ "embed"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

//go:embed logical_assignment_operators.schema.json
var schemaJSON []byte

// globalValueMeaning is the set of declaration spaces an ESLint value
// reference can resolve to, used when asking whether `undefined` / `Boolean`
// still name the environment global rather than a declaration in this file.
const globalValueMeaning = ast.SymbolFlagsValue | ast.SymbolFlagsNamespace | ast.SymbolFlagsAlias

// assignmentPrecedence is the precedence ESLint's astUtils.getPrecedence
// assigns to an AssignmentExpression. The logical fixer compares it against
// the enclosing expression's precedence to decide whether the rewritten
// assignment needs parentheses of its own.
const assignmentPrecedence = 1

func assignmentMessage(operator string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "assignment",
		Description: "Assignment (=) can be replaced with operator assignment (" + operator + ").",
		Data:        map[string]string{"operator": operator},
	}
}

func useLogicalOperatorMessage(operator string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "useLogicalOperator",
		Description: "Convert this assignment to use the operator " + operator + ".",
		Data:        map[string]string{"operator": operator},
	}
}

func logicalMessage(operator string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "logical",
		Description: "Logical expression can be replaced with an assignment (" + operator + ").",
		Data:        map[string]string{"operator": operator},
	}
}

func convertLogicalMessage(operator string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "convertLogical",
		Description: "Replace this logical expression with an assignment with the operator " + operator + ".",
		Data:        map[string]string{"operator": operator},
	}
}

func ifMessage(operator string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "if",
		Description: "'if' statement can be replaced with a logical operator assignment with operator " + operator + ".",
		Data:        map[string]string{"operator": operator},
	}
}

func convertIfMessage(operator string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "convertIf",
		Description: "Replace this 'if' statement with a logical assignment with operator " + operator + ".",
		Data:        map[string]string{"operator": operator},
	}
}

func unexpectedMessage(operator string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "unexpected",
		Description: "Unexpected logical operator assignment (" + operator + ") shorthand.",
		Data:        map[string]string{"operator": operator},
	}
}

func separateMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "separate",
		Description: "Separate the logical assignment into an assignment with a logical operator.",
	}
}

type options struct {
	never                  bool
	enforceForIfStatements bool
}

func parseOptions(configured []any) options {
	opts := options{}
	if len(configured) == 0 {
		return opts
	}
	if mode, ok := configured[0].(string); ok && mode == "never" {
		opts.never = true
		return opts
	}
	if len(configured) > 1 {
		if settings, ok := configured[1].(map[string]any); ok {
			if value, ok := settings["enforceForIfStatements"].(bool); ok {
				opts.enforceForIfStatements = value
			}
		}
	}
	return opts
}

// skipParens mirrors ESTree, which has no node for a parenthesized
// expression: every child access upstream makes lands on the expression
// inside the parentheses.
func skipParens(node *ast.Node) *ast.Node {
	if node == nil {
		return nil
	}
	return ast.SkipParentheses(node)
}

// logicalOperatorText returns the source spelling of a logical operator
// token, and whether the token is one at all. These three are exactly the
// operators ESTree models as a LogicalExpression.
func logicalOperatorText(kind ast.Kind) (string, bool) {
	switch kind {
	case ast.KindBarBarToken:
		return "||", true
	case ast.KindAmpersandAmpersandToken:
		return "&&", true
	case ast.KindQuestionQuestionToken:
		return "??", true
	}
	return "", false
}

// logicalAssignmentOperator returns the plain logical operator behind a
// logical assignment token (`||=` -> `||`) as both a token kind and its
// source spelling, and whether the token is one.
func logicalAssignmentOperator(kind ast.Kind) (ast.Kind, string, bool) {
	switch kind {
	case ast.KindBarBarEqualsToken:
		return ast.KindBarBarToken, "||", true
	case ast.KindAmpersandAmpersandEqualsToken:
		return ast.KindAmpersandAmpersandToken, "&&", true
	case ast.KindQuestionQuestionEqualsToken:
		return ast.KindQuestionQuestionToken, "??", true
	}
	return ast.KindUnknown, "", false
}

// isLogicalExpression reports whether node is what ESTree calls a
// LogicalExpression, i.e. a binary expression joined by `&&`, `||`, or `??`.
func isLogicalExpression(node *ast.Node) bool {
	if node == nil || node.Kind != ast.KindBinaryExpression {
		return false
	}
	_, ok := logicalOperatorText(node.AsBinaryExpression().OperatorToken.Kind)
	return ok
}

// isReference reports whether node is an identifier other than `undefined`,
// or a member access. An access that is part of an optional chain is
// excluded: ESTree wraps the outermost link of a chain in a ChainExpression,
// which is neither of the two types upstream accepts.
func isReference(node *ast.Node) bool {
	if node == nil {
		return false
	}
	switch node.Kind {
	case ast.KindIdentifier:
		return node.AsIdentifier().Text != "undefined"
	case ast.KindPropertyAccessExpression, ast.KindElementAccessExpression:
		return !ast.IsOptionalChain(node)
	}
	return false
}

// spellsTypeAssertion reports whether node writes a TypeScript assertion
// anywhere along the reference chain that [sameReference] inspects.
//
// ESLint's isSameReference switches on the ESTree node type and has no case
// for `TSNonNullExpression`, `TSAsExpression`, `TSSatisfiesExpression`, or
// `TSTypeAssertion`, so it answers "not the same reference" as soon as one
// appears — even for two identical ones. Every position this walks is a
// position that comparison recurses into: when the static property name of an
// element access resolves at all, the key is a plain string or number
// literal, so a key that spells an assertion always reaches the recursive
// comparison too.
func spellsTypeAssertion(node *ast.Node) bool {
	node = skipParens(node)
	if node == nil {
		return false
	}
	switch node.Kind {
	case ast.KindNonNullExpression, ast.KindAsExpression,
		ast.KindSatisfiesExpression, ast.KindTypeAssertionExpression:
		return true
	case ast.KindPropertyAccessExpression:
		return spellsTypeAssertion(node.AsPropertyAccessExpression().Expression)
	case ast.KindElementAccessExpression:
		access := node.AsElementAccessExpression()
		return spellsTypeAssertion(access.Expression) ||
			spellsTypeAssertion(access.ArgumentExpression)
	}
	return false
}

// sameReference reports whether two expressions name the same value, the way
// ESLint's astUtils.isSameReference does. The shared helper treats TypeScript
// assertions as transparent wrappers; upstream treats them as unknown node
// types, so they are screened out first.
func sameReference(left, right *ast.Node) bool {
	if spellsTypeAssertion(left) || spellsTypeAssertion(right) {
		return false
	}
	return utils.IsSameReference(left, right, false)
}

// isNullLiteral reports whether node is the `null` literal.
func isNullLiteral(node *ast.Node) bool {
	node = skipParens(node)
	return node != nil && node.Kind == ast.KindNullKeyword
}

// isUndefined reports whether node is the global `undefined` or a void
// expression on a zero-valued numeric literal (`void 0`).
func isUndefined(ctx *rule.RuleContext, node *ast.Node) bool {
	node = skipParens(node)
	if node == nil {
		return false
	}
	if node.Kind == ast.KindIdentifier && node.AsIdentifier().Text == "undefined" {
		return ctx.Globals.Access("undefined").IsDeclared() &&
			ctx.Refs.IsGlobalNameReference(node, "undefined", globalValueMeaning)
	}
	if node.Kind != ast.KindVoidExpression {
		return false
	}
	argument := skipParens(node.AsVoidExpression().Expression)
	if argument == nil || argument.Kind != ast.KindNumericLiteral {
		return false
	}
	// Upstream compares the literal's evaluated value against 0, so every
	// spelling of zero counts — `void 0x0` and `void 0.0` as much as `void 0`.
	// A BigInt zero does not: `0n === 0` is false.
	return utils.NormalizeNumericLiteral(argument.AsNumericLiteral().Text) == "0"
}

// implicitNullishReference returns the reference operand of a loose nullish
// check such as `value == null` or `value == void 0`, in either operand
// order, or nil when expression is not one.
func implicitNullishReference(ctx *rule.RuleContext, expression *ast.Node) *ast.Node {
	if expression == nil || expression.Kind != ast.KindBinaryExpression {
		return nil
	}
	comparison := expression.AsBinaryExpression()
	if comparison.OperatorToken.Kind != ast.KindEqualsEqualsToken {
		return nil
	}

	left := skipParens(comparison.Left)
	right := skipParens(comparison.Right)
	reference, nullish := left, right
	if !isReference(left) {
		reference, nullish = right, left
	}
	if !isReference(reference) {
		return nil
	}
	if isNullLiteral(nullish) || isUndefined(ctx, nullish) {
		return reference
	}
	return nil
}

// doubleComparisonParts returns the two `===` comparisons of a
// `? === ? || ? === ?` condition, and whether expression has that shape.
func doubleComparisonParts(expression *ast.Node) (*ast.BinaryExpression, *ast.BinaryExpression, bool) {
	if expression == nil || expression.Kind != ast.KindBinaryExpression {
		return nil, nil, false
	}
	disjunction := expression.AsBinaryExpression()
	if disjunction.OperatorToken.Kind != ast.KindBarBarToken {
		return nil, nil, false
	}

	left := skipParens(disjunction.Left)
	right := skipParens(disjunction.Right)
	if left == nil || left.Kind != ast.KindBinaryExpression ||
		right == nil || right.Kind != ast.KindBinaryExpression {
		return nil, nil, false
	}
	leftComparison := left.AsBinaryExpression()
	rightComparison := right.AsBinaryExpression()
	if leftComparison.OperatorToken.Kind != ast.KindEqualsEqualsEqualsToken ||
		rightComparison.OperatorToken.Kind != ast.KindEqualsEqualsEqualsToken {
		return nil, nil, false
	}
	return leftComparison, rightComparison, true
}

// explicitNullishReference returns the reference operand of a condition that
// compares one value against both `null` and `undefined`, such as
// `value === null || value === undefined`, or nil when expression is not one.
func explicitNullishReference(ctx *rule.RuleContext, expression *ast.Node) *ast.Node {
	leftComparison, rightComparison, ok := doubleComparisonParts(expression)
	if !ok {
		return nil
	}

	leftLeft := skipParens(leftComparison.Left)
	leftRight := skipParens(leftComparison.Right)
	rightLeft := skipParens(rightComparison.Left)
	rightRight := skipParens(rightComparison.Right)

	leftReference, leftNullish := leftLeft, leftRight
	if !isReference(leftLeft) {
		leftReference, leftNullish = leftRight, leftLeft
	}
	rightReference, rightNullish := rightLeft, rightRight
	if !isReference(rightLeft) {
		rightReference, rightNullish = rightRight, rightLeft
	}

	if !sameReference(leftReference, rightReference) {
		return nil
	}
	if (isNullLiteral(leftNullish) && isUndefined(ctx, rightNullish)) ||
		(isUndefined(ctx, leftNullish) && isNullLiteral(rightNullish)) {
		return leftReference
	}
	return nil
}

// booleanCastArgument returns the single argument of a `Boolean(value)` call
// on the global `Boolean`, and whether node is such a call.
func booleanCastArgument(ctx *rule.RuleContext, node *ast.Node) (*ast.Node, bool) {
	if node == nil || node.Kind != ast.KindCallExpression || ast.IsOptionalChain(node) {
		return nil, false
	}
	call := node.AsCallExpression()
	callee := skipParens(call.Expression)
	if callee == nil || callee.Kind != ast.KindIdentifier || callee.AsIdentifier().Text != "Boolean" {
		return nil, false
	}
	if call.Arguments == nil || len(call.Arguments.Nodes) != 1 {
		return nil, false
	}
	if !ctx.Globals.Access("Boolean").IsDeclared() ||
		!ctx.Refs.IsGlobalNameReference(callee, "Boolean", globalValueMeaning) {
		return nil, false
	}
	return skipParens(call.Arguments.Nodes[0]), true
}

// existence describes a condition that tests whether a reference holds a
// value, together with the logical assignment operator that expresses the
// same test.
type existence struct {
	reference *ast.Node
	operator  string
}

// getExistence classifies an `if` test as one of the known existence checks:
// truthiness (`value`, `Boolean(value)`, `!!value`), its negation (`!value`,
// `!Boolean(value)`), or nullishness (`value == null`,
// `value === undefined || value === null`). It returns nil for anything else.
func getExistence(ctx *rule.RuleContext, test *ast.Node) *existence {
	expression := skipParens(test)
	if expression == nil {
		return nil
	}

	isNegated := expression.Kind == ast.KindPrefixUnaryExpression &&
		expression.AsPrefixUnaryExpression().Operator == ast.KindExclamationToken
	base := expression
	if isNegated {
		base = skipParens(expression.AsPrefixUnaryExpression().Operand)
	}

	if isReference(base) {
		if isNegated {
			return &existence{reference: base, operator: "||"}
		}
		return &existence{reference: base, operator: "&&"}
	}
	if base != nil && base.Kind == ast.KindPrefixUnaryExpression &&
		base.AsPrefixUnaryExpression().Operator == ast.KindExclamationToken {
		if inner := skipParens(base.AsPrefixUnaryExpression().Operand); isReference(inner) {
			return &existence{reference: inner, operator: "&&"}
		}
	}
	if argument, ok := booleanCastArgument(ctx, base); ok && isReference(argument) {
		if isNegated {
			return &existence{reference: argument, operator: "||"}
		}
		return &existence{reference: argument, operator: "&&"}
	}
	if reference := implicitNullishReference(ctx, expression); reference != nil {
		return &existence{reference: reference, operator: "??"}
	}
	if reference := explicitNullishReference(ctx, expression); reference != nil {
		return &existence{reference: reference, operator: "??"}
	}
	return nil
}

// isInsideWithBlock reports whether node sits in the body of a `with`
// statement, where a bare identifier may resolve to a property of the object
// and therefore run a getter.
func isInsideWithBlock(node *ast.Node) bool {
	for current := node; current != nil; current = current.Parent {
		parent := current.Parent
		if parent == nil || parent.Kind == ast.KindSourceFile {
			return false
		}
		if parent.Kind == ast.KindWithStatement && parent.AsWithStatement().Statement == current {
			return true
		}
	}
	return false
}

// estreeParent returns the innermost ancestor ESTree would report as the
// node's parent, skipping the parenthesized-expression nodes tsgo keeps and
// ESTree discards.
func estreeParent(node *ast.Node) *ast.Node {
	parent := node.Parent
	for parent != nil && parent.Kind == ast.KindParenthesizedExpression {
		parent = parent.Parent
	}
	return parent
}

// requiresOuterParenthesis reports whether an assignment written in place of
// the logical expression at node would bind too loosely for its surroundings.
func requiresOuterParenthesis(node *ast.Node) bool {
	parent := estreeParent(node)
	if parent == nil || parent.Kind == ast.KindExpressionStatement {
		return false
	}
	// tsgo represents a destructuring default in an assignment target as a
	// binary assignment, while ESTree gives the default expression an
	// AssignmentPattern parent. Preserve the parentheses ESLint adds for that
	// lower-precedence parent instead of treating it like an ordinary `=`.
	if utils.IsDefaultValueInDestructuringAssignment(parent) {
		return true
	}
	parentPrecedence := utils.EslintLikePrecedence(parent)
	return parentPrecedence == -1 || assignmentPrecedence < parentPrecedence
}

// previousTokenLastCharacter returns the last character of the token before
// node, or 0 when node starts the file. A node's Pos() is where the token
// before it ended, so the answer is one index back from there — no
// re-tokenizing, which is what makes this safe in a file that already contains
// a template literal or a regular expression. Re-scanning such a file from the
// start drifts out of step with the parser at the first `}` or `/` whose
// meaning depends on parser context.
func previousTokenLastCharacter(sourceFile *ast.SourceFile, node *ast.Node) byte {
	start := node.Pos()
	if start <= 0 {
		return 0
	}
	return sourceFile.Text()[start-1]
}

// nextTokenFirstCharacter returns the first character of the token after node,
// or 0 when node ends the file.
func nextTokenFirstCharacter(sourceFile *ast.SourceFile, node *ast.Node) byte {
	text := sourceFile.Text()
	position := scanner.SkipTrivia(text, node.End())
	if position >= len(text) {
		return 0
	}
	return text[position]
}

// isParenthesised mirrors ESLint's token-level astUtils.isParenthesised: the
// parentheses do not have to be the node's own. An argument list, an `if`
// head, or a `while` head all supply a `(` before and a `)` after the
// expression, and upstream counts those.
//
// Comparing single characters is enough to identify the two tokens: `(` is the
// only token that ends with `(`, and `)` the only one that starts with `)`.
func isParenthesised(sourceFile *ast.SourceFile, node *ast.Node) bool {
	return previousTokenLastCharacter(sourceFile, node) == '(' &&
		nextTokenFirstCharacter(sourceFile, node) == ')'
}

// isIdentifierOrKeywordToken reports whether token is what espree types as an
// "Identifier" or a "Keyword". The three literal keywords are excluded
// because espree gives them their own "Boolean" and "Null" token types.
func isIdentifierOrKeywordToken(token utils.SourceToken) bool {
	switch token.Kind {
	case ast.KindIdentifier:
		return true
	case ast.KindTrueKeyword, ast.KindFalseKeyword, ast.KindNullKeyword:
		return false
	}
	return ast.IsKeywordKind(token.Kind)
}

// insertAt builds a zero-width fix that writes text at position.
func insertAt(position int, text string) rule.RuleFix {
	return rule.RuleFixReplaceRange(core.NewTextRange(position, position), text)
}

// leftmostOperand walks down the left spine of a chain of same-operator
// logical expressions and returns the leftmost operand together with the
// logical expression that owns it. An owner of nil means the walk stopped at
// a parenthesized nesting of the same operator: the parentheses make the
// evaluation order explicit, so upstream leaves the expression alone.
func leftmostOperand(logical *ast.Node) (*ast.Node, *ast.Node) {
	operator := logical.AsBinaryExpression().OperatorToken.Kind
	owner := logical
	current := logical.AsBinaryExpression().Left

	for {
		parenthesized := current.Kind == ast.KindParenthesizedExpression
		inner := skipParens(current)
		if inner == nil {
			return nil, nil
		}
		if inner.Kind == ast.KindBinaryExpression &&
			inner.AsBinaryExpression().OperatorToken.Kind == operator {
			if parenthesized {
				return inner, nil
			}
			owner = inner
			current = inner.AsBinaryExpression().Left
			continue
		}
		return inner, owner
	}
}

type ruleState struct {
	ctx      *rule.RuleContext
	isStrict bool
}

// cannotBeGetter reports whether writing to node is guaranteed not to run a
// setter, which is what makes the rewrite safe enough to apply automatically.
func (s *ruleState) cannotBeGetter(node *ast.Node) bool {
	return node != nil && node.Kind == ast.KindIdentifier &&
		(s.isStrict || !isInsideWithBlock(node))
}

// accessesSingleProperty reports whether node reads exactly one property off
// a plain base, so rewriting it as a logical assignment cannot change how
// many times an intermediate getter runs.
func (s *ruleState) accessesSingleProperty(node *ast.Node) bool {
	if node == nil {
		return false
	}
	if !s.isStrict && isInsideWithBlock(node) {
		return node.Kind == ast.KindIdentifier
	}
	if !ast.IsAccessExpression(node) {
		return false
	}

	object := skipParens(utils.AccessExpressionObject(node))
	if object == nil {
		return false
	}
	switch object.Kind {
	case ast.KindIdentifier, ast.KindThisKeyword, ast.KindSuperKeyword:
	default:
		return false
	}

	if node.Kind == ast.KindElementAccessExpression {
		property := skipParens(node.AsElementAccessExpression().ArgumentExpression)
		if property == nil {
			return false
		}
		// Upstream rejects a computed key that is itself a member access or an
		// optional chain, since evaluating it can run another getter.
		if ast.IsAccessExpression(property) || ast.IsOptionalChain(property) {
			return false
		}
	}
	return true
}

// hasCommentsInside reports whether any comment sits within the given range.
// Every fixer in this rule bails out when one does: each of them deletes a
// span of source, and a comment inside it would go with it.
func (s *ruleState) hasCommentsInside(textRange core.TextRange) bool {
	return utils.HasCommentInSpan(s.ctx.Comments.All(), textRange.Pos(), textRange.End())
}

// report emits one diagnostic that carries either an autofix or a suggestion,
// never both, matching upstream's createConditionalFixer.
func (s *ruleState) report(
	node *ast.Node,
	message rule.RuleMessage,
	suggestionMessage rule.RuleMessage,
	shouldBeFixed bool,
	build func() []rule.RuleFix,
) {
	if shouldBeFixed {
		s.ctx.ReportNodeWithDeferredFixes(node, message, build)
		return
	}
	s.ctx.ReportNodeWithDeferredSuggestions(node, message, func() []rule.RuleSuggestion {
		fixes := build()
		if len(fixes) == 0 {
			return nil
		}
		return []rule.RuleSuggestion{{Message: suggestionMessage, FixesArr: fixes}}
	})
}

// checkAssignment reports `foo = foo || bar`, which the shorthand
// `foo ||= bar` expresses directly.
func (s *ruleState) checkAssignment(node *ast.Node, assignment *ast.BinaryExpression) {
	right := skipParens(assignment.Right)
	if right == nil || right.Kind != ast.KindBinaryExpression {
		return
	}
	operator, ok := logicalOperatorText(right.AsBinaryExpression().OperatorToken.Kind)
	if !ok {
		return
	}

	operand, owner := leftmostOperand(right)
	if owner == nil {
		return
	}
	left := skipParens(assignment.Left)
	if !sameReference(left, operand) {
		return
	}

	compound := operator + "="
	s.report(
		node,
		assignmentMessage(compound),
		useLogicalOperatorMessage(compound),
		s.cannotBeGetter(left),
		func() []rule.RuleFix {
			sourceFile := s.ctx.SourceFile
			nodeRange := utils.TrimNodeTextRange(sourceFile, node)
			if s.hasCommentsInside(nodeRange) {
				return nil
			}

			// No parentheses are needed around the assignment: swapping `=` for
			// `||=` leaves its precedence unchanged.
			operatorRange := utils.TrimNodeTextRange(sourceFile, assignment.OperatorToken)
			ownerRange := utils.TrimNodeTextRange(sourceFile, owner)
			rightOperandRange := utils.TrimNodeTextRange(sourceFile, owner.AsBinaryExpression().Right)

			return []rule.RuleFix{
				// -> foo ||= foo || bar
				insertAt(operatorRange.Pos(), operator),
				// -> foo ||= bar
				rule.RuleFixRemoveRange(core.NewTextRange(ownerRange.Pos(), rightOperandRange.Pos())),
			}
		},
	)
}

// checkLogical reports `foo || (foo = bar)`. The right side has to be
// parenthesized in the source, since `foo || foo = bar` would parse as
// `(foo || foo) = bar`.
func (s *ruleState) checkLogical(node *ast.Node, logical *ast.BinaryExpression, operator string) {
	right := skipParens(logical.Right)
	if right == nil || right.Kind != ast.KindBinaryExpression {
		return
	}
	assignment := right.AsBinaryExpression()
	if assignment.OperatorToken.Kind != ast.KindEqualsToken {
		return
	}

	left := skipParens(logical.Left)
	if !isReference(left) || !sameReference(left, assignment.Left) {
		return
	}

	compound := operator + "="
	shouldBeFixed := s.cannotBeGetter(left) || s.accessesSingleProperty(left)
	s.report(
		node,
		logicalMessage(compound),
		convertLogicalMessage(compound),
		shouldBeFixed,
		func() []rule.RuleFix {
			sourceFile := s.ctx.SourceFile
			nodeRange := utils.TrimNodeTextRange(sourceFile, node)
			if s.hasCommentsInside(nodeRange) {
				return nil
			}

			assignmentRange := utils.TrimNodeTextRange(sourceFile, right)
			operatorRange := utils.TrimNodeTextRange(sourceFile, assignment.OperatorToken)

			wrap := !isParenthesised(sourceFile, node) && requiresOuterParenthesis(node)
			fixes := make([]rule.RuleFix, 0, 5)
			if wrap {
				fixes = append(fixes, insertAt(nodeRange.Pos(), "("))
			}
			fixes = append(fixes,
				// Everything the logical expression adds around the assignment —
				// the left operand, the operator, and any parentheses — is
				// dropped. -> foo = bar
				rule.RuleFixRemoveRange(core.NewTextRange(nodeRange.Pos(), assignmentRange.Pos())),
				// -> foo ||= bar
				insertAt(operatorRange.Pos(), operator),
				rule.RuleFixRemoveRange(core.NewTextRange(assignmentRange.End(), nodeRange.End())),
			)
			if wrap {
				fixes = append(fixes, insertAt(nodeRange.End(), ")"))
			}
			return fixes
		},
	)
}

// checkIfStatement reports `if (foo) foo = bar` and its nullish-check
// variants, enabled by the enforceForIfStatements option.
func (s *ruleState) checkIfStatement(node *ast.Node) {
	ifStatement := node.AsIfStatement()
	if ifStatement.ElseStatement != nil {
		return
	}

	consequent := ifStatement.ThenStatement
	hasBody := consequent.Kind == ast.KindBlock
	body := consequent
	if hasBody {
		statements := consequent.AsBlock().Statements
		if statements == nil || len(statements.Nodes) != 1 {
			return
		}
		body = statements.Nodes[0]
	}
	if body.Kind != ast.KindExpressionStatement {
		return
	}
	expression := skipParens(body.AsExpressionStatement().Expression)
	if expression == nil || expression.Kind != ast.KindBinaryExpression {
		return
	}
	assignment := expression.AsBinaryExpression()
	if assignment.OperatorToken.Kind != ast.KindEqualsToken {
		return
	}

	found := getExistence(s.ctx, ifStatement.Expression)
	if found == nil || !sameReference(found.reference, assignment.Left) {
		return
	}

	compound := found.operator + "="
	test := skipParens(ifStatement.Expression)
	shouldBeFixed := s.cannotBeGetter(found.reference) ||
		(!isLogicalExpression(test) && s.accessesSingleProperty(found.reference))
	s.report(
		node,
		ifMessage(compound),
		convertIfMessage(compound),
		shouldBeFixed,
		func() []rule.RuleFix {
			sourceFile := s.ctx.SourceFile
			ifRange := utils.TrimNodeTextRange(sourceFile, node)
			if s.hasCommentsInside(ifRange) {
				return nil
			}

			bodyRange := utils.TrimNodeTextRange(sourceFile, body)
			firstBodyToken, hasFirstBodyToken := utils.TokenAtOrAfter(sourceFile, bodyRange.Pos())
			// Only `;` ends a `;` token, and only `{` ends a `{` token in a
			// position a statement can follow.
			previousCharacter := previousTokenLastCharacter(sourceFile, node)
			if previousCharacter != 0 && previousCharacter != ';' && previousCharacter != '{' &&
				(!hasFirstBodyToken || !isIdentifierOrKeywordToken(firstBodyToken)) {
				// The statement left behind could otherwise be read as a
				// continuation of the previous one, turning `fn() if (a == null)
				// (a) = b` into `fn()(a) ??= b`.
				return nil
			}

			operatorRange := utils.TrimNodeTextRange(sourceFile, assignment.OperatorToken)
			fixes := []rule.RuleFix{
				// -> if (foo) foo ||= bar
				insertAt(operatorRange.Pos(), found.operator),
				// -> foo ||= bar
				rule.RuleFixRemoveRange(core.NewTextRange(ifRange.Pos(), bodyRange.Pos())),
				rule.RuleFixRemoveRange(core.NewTextRange(bodyRange.End(), ifRange.End())),
			}

			// A braced body may have ended on `}` rather than `;`, and the
			// braces are about to go away.
			if hasBody {
				if next := nextTokenFirstCharacter(sourceFile, expression); next != 0 && next != ';' {
					fixes = append(fixes, insertAt(ifRange.End(), ";"))
				}
			}
			return fixes
		},
	)
}

// checkNever implements the "never" mode: every logical assignment shorthand
// is expanded back into an assignment with a logical operator.
func (s *ruleState) checkNever(node *ast.Node, assignment *ast.BinaryExpression) {
	plainOperatorKind, operator, ok := logicalAssignmentOperator(assignment.OperatorToken.Kind)
	if !ok {
		return
	}

	left := skipParens(assignment.Left)
	compound := operator + "="
	s.report(
		node,
		unexpectedMessage(compound),
		separateMessage(),
		s.cannotBeGetter(left),
		func() []rule.RuleFix {
			sourceFile := s.ctx.SourceFile
			nodeRange := utils.TrimNodeTextRange(sourceFile, node)
			if s.hasCommentsInside(nodeRange) {
				return nil
			}

			text := sourceFile.Text()
			operatorRange := utils.TrimNodeTextRange(sourceFile, assignment.OperatorToken)
			leftRange := utils.TrimNodeTextRange(sourceFile, left)

			// -> foo = foo || bar
			fixes := []rule.RuleFix{rule.RuleFixReplaceRange(
				operatorRange,
				"= "+text[leftRange.Pos():leftRange.End()]+" "+operator,
			)}

			right := assignment.Right
			// An already-parenthesized right side keeps its own parentheses and
			// needs no others.
			needsParenthesis := right.Kind != ast.KindParenthesizedExpression &&
				(utils.EslintLikePrecedence(right) <=
					utils.EslintLikeBinaryOperatorPrecedence(plainOperatorKind) ||
					// `??` cannot be mixed with `&&` or `||` even though they
					// share a precedence level.
					(assignment.OperatorToken.Kind == ast.KindQuestionQuestionEqualsToken &&
						isLogicalExpression(right)))
			if needsParenthesis {
				// -> foo = foo || (bar)
				rightRange := utils.TrimNodeTextRange(sourceFile, right)
				fixes = append(fixes,
					insertAt(rightRange.Pos(), "("),
					insertAt(rightRange.End(), ")"),
				)
			}
			return fixes
		},
	)
}

// https://eslint.org/docs/latest/rules/logical-assignment-operators
var LogicalAssignmentOperatorsRule = rule.Rule{
	Name:   "logical-assignment-operators",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		opts := parseOptions(options)
		state := &ruleState{
			ctx: &ctx,
			// Upstream reads the global scope's strictness, which a `"use
			// strict"` directive inside a `with` body does not reach.
			isStrict: utils.HasUseStrictDirective(ctx.SourceFile.AsNode()),
		}

		if opts.never {
			return rule.RuleListeners{
				ast.KindBinaryExpression: func(node *ast.Node) {
					state.checkNever(node, node.AsBinaryExpression())
				},
			}
		}

		listeners := rule.RuleListeners{
			ast.KindBinaryExpression: func(node *ast.Node) {
				binary := node.AsBinaryExpression()
				if binary.OperatorToken.Kind == ast.KindEqualsToken {
					state.checkAssignment(node, binary)
					return
				}
				if operator, ok := logicalOperatorText(binary.OperatorToken.Kind); ok {
					state.checkLogical(node, binary, operator)
				}
			},
		}
		if opts.enforceForIfStatements {
			listeners[ast.KindIfStatement] = func(node *ast.Node) {
				state.checkIfStatement(node)
			}
		}
		return listeners
	},
}
