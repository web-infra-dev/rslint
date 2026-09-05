package prefer_array_some

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/unicornutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

const (
	messageIDArraySome           = "some"
	messageIDArraySomeSuggestion = "some-suggestion"
	messageIDArrayFilter         = "filter"
)

func messageArraySome(method string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          messageIDArraySome,
		Description: "Prefer `.some(…)` over `." + method + "(…)`.",
		Data:        map[string]string{"method": method},
	}
}

func messageArraySomeSuggestion(method string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          messageIDArraySomeSuggestion,
		Description: "Replace `." + method + "(…)` with `.some(…)`.",
		Data:        map[string]string{"method": method},
	}
}

func messageArrayFilter() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          messageIDArrayFilter,
		Description: "Prefer `.some(…)` over non-zero length check from `.filter(…)`.",
	}
}

// PreferArraySomeRule mirrors eslint-plugin-unicorn's prefer-array-some. It is
// syntactic-first: the type checker is only consulted (when available) to skip
// receivers that are known non-indexed collections (`Map`/`Set`/typed arrays'
// keyed cousins) and, for the `.find()` variable-usage branch, to confirm the
// receiver is an array. On JS / gap files (`ctx.TypeChecker == nil`) those
// checks degrade to "unknown", exactly like upstream when no parser services
// are present — so the rule stays useful and MUST NOT declare RequiresTypeInfo.
var PreferArraySomeRule = rule.Rule{
	Name:   "unicorn/prefer-array-some",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		return rule.RuleListeners{
			// `.find(…)` / `.findLast(…)` used in a boolean position.
			ast.KindCallExpression: func(node *ast.Node) {
				checkFindCall(ctx, node)
			},
			// `.{findIndex,findLastIndex}(…)` compared against -1 / 0, and
			// `.filter(…).length` compared against 0. Both collapse onto
			// tsgo's single BinaryExpression kind.
			ast.KindBinaryExpression: func(node *ast.Node) {
				if checkFindIndexComparison(ctx, node) {
					return
				}
				checkFilterLength(ctx, node)
			},
		}
	},
}

// ---- `.find()` / `.findLast()` ----

func checkFindCall(ctx rule.RuleContext, node *ast.Node) {
	if !isPotentialFindMethodCall(node) {
		return
	}

	comparison := comparisonParent(node)
	isBooleanUse := comparison != nil || isBooleanExpression(ctx, node) || isControlFlowTest(node)
	if !isBooleanUse && !isVariableDeclarationInitializer(node) {
		return
	}

	minArgs := 1
	maxArgs := 2
	call, ok := unicornutil.MatchDotMethodCall(node, unicornutil.DotMethodCallOptions{
		Methods:             []string{"find", "findLast"},
		MinimumArguments:    &minArgs,
		MaximumArguments:    &maxArgs,
		RejectSpreadElement: true,
	})
	if !ok {
		return
	}
	// `foo.find<Item>(fn)` — explicit type arguments mean the result is used
	// for its value, not as a boolean. tsgo exposes them on the call node.
	if node.AsCallExpression().TypeArguments != nil {
		return
	}

	if isBooleanUse {
		if unicornutil.IsKnownNonIndexedCollection(ctx, call.Object) {
			return
		}
	} else if !isFindResultVariableUsedOnlyAsBoolean(ctx, node, call) {
		return
	}

	methodNode := call.Property
	method := methodNode.AsIdentifier().Text

	ctx.ReportNodeWithDeferredSuggestions(methodNode, messageArraySome(method), func() []rule.RuleSuggestion {
		// Removing the comparison would drop comments between the call and the
		// end of the comparison; upstream withholds the suggestion in that
		// case. Checked inside the builder so lint-only runs never materialize
		// the comment list — returning no suggestions reports exactly what
		// ReportNode would.
		if comparison != nil && utils.HasCommentInSpan(
			ctx.Comments.All(),
			parenthesizedEnd(ctx, node),
			comparison.End(),
		) {
			return nil
		}

		fixes := []rule.RuleFix{rule.RuleFixReplace(ctx.SourceFile, methodNode, "some")}
		if comparison != nil {
			// Remove the `== undefined` / `!== null` tail.
			parenEnd := parenthesizedEnd(ctx, node)
			fixes = append(fixes, rule.RuleFixRemoveRange(core.NewTextRange(parenEnd, comparison.End())))

			// `!=` / `!==` already mean "found"; other operators (`==` / `===`)
			// mean "not found", so negate.
			operator := comparison.AsBinaryExpression().OperatorToken.Kind
			if operator != ast.KindExclamationEqualsToken && operator != ast.KindExclamationEqualsEqualsToken {
				fixes = append(fixes, insertBeforeRange(parenthesizedRange(ctx, node), "!"))
			}
		}
		return []rule.RuleSuggestion{{
			Message:  messageArraySomeSuggestion(method),
			FixesArr: fixes,
		}}
	})
}

// isPotentialFindMethodCall is a conservative, allocation-free initial filter for
// the CallExpression listener. MatchDotMethodCall still owns the optional-call,
// argument-count, spread, and parenthesized-callee checks.
func isPotentialFindMethodCall(node *ast.Node) bool {
	if node == nil || !ast.IsCallExpression(node) {
		return false
	}
	expression := node.AsCallExpression().Expression
	if expression == nil {
		return false
	}
	callee := ast.SkipParentheses(expression)
	if callee == nil || !ast.IsPropertyAccessExpression(callee) {
		return false
	}
	name := callee.AsPropertyAccessExpression().Name()
	if name == nil || !ast.IsIdentifier(name) {
		return false
	}
	method := name.AsIdentifier().Text
	return method == "find" || method == "findLast"
}

func isVariableDeclarationInitializer(node *ast.Node) bool {
	declarator := node.Parent
	return declarator != nil && ast.IsVariableDeclaration(declarator) &&
		declarator.AsVariableDeclaration().Initializer == node
}

// comparisonParent returns the BinaryExpression comparing `node` against
// undefined / null, skipping parentheses around the call. Returns nil when
// node is not the left operand of such a comparison.
func comparisonParent(node *ast.Node) *ast.Node {
	parent := effectiveParent(node)
	if parent == nil || !ast.IsBinaryExpression(parent) {
		return nil
	}
	bin := parent.AsBinaryExpression()
	if bin.Left == nil || bin.Right == nil || bin.OperatorToken == nil ||
		ast.SkipParentheses(bin.Left) != node {
		return nil
	}
	right := bin.Right
	switch bin.OperatorToken.Kind {
	case ast.KindEqualsEqualsToken, ast.KindExclamationEqualsToken:
		if utils.IsUndefinedIdentifier(right) || isNullLiteral(right) {
			return parent
		}
	case ast.KindEqualsEqualsEqualsToken, ast.KindExclamationEqualsEqualsToken:
		if utils.IsUndefinedIdentifier(right) {
			return parent
		}
	}
	return nil
}

// isFindResultVariableUsedOnlyAsBoolean mirrors upstream's helper: a
// `const x = arr.find(fn)` whose every read is used as a boolean. Requires the
// receiver to be a known array (upstream `isArray`), a plain single-name
// `const` binding that isn't exported, and ≥1 reference — all of which must be
// boolean-position reads.
func isFindResultVariableUsedOnlyAsBoolean(ctx rule.RuleContext, callExpression *ast.Node, call unicornutil.DotMethodCall) bool {
	if ctx.Refs == nil {
		return false
	}

	// callExpression.Parent must be `VariableDeclaration` whose Initializer is
	// the call, with a plain Identifier name and no type annotation.
	declarator := callExpression.Parent
	if declarator == nil || !ast.IsVariableDeclaration(declarator) {
		return false
	}
	varDecl := declarator.AsVariableDeclaration()
	if varDecl.Initializer != callExpression {
		return false
	}
	name := varDecl.Name()
	if name == nil || !ast.IsIdentifier(name) || varDecl.Type != nil {
		return false
	}

	// The declaration list must be `const` and not part of an `export`.
	declList := declarator.Parent
	if declList == nil || !ast.IsVariableDeclarationList(declList) ||
		declList.Flags&ast.NodeFlagsConst == 0 {
		return false
	}
	statement := declList.Parent
	if statement == nil || !ast.IsVariableStatement(statement) {
		return false
	}
	if statement.Parent != nil && ast.IsExportDeclaration(statement.Parent) {
		return false
	}
	if statement.ModifierFlags()&ast.ModifierFlagsExport != 0 {
		return false
	}
	if !unicornutil.IsArray(ctx, call.Object) {
		return false
	}

	sym := declarator.Symbol()
	if sym == nil {
		return false
	}
	references := ctx.Refs.References(sym)
	if len(references) == 0 {
		return false
	}

	for _, reference := range references {
		if !isBooleanExpression(ctx, reference) && !isControlFlowTest(reference) {
			return false
		}
	}
	return true
}

// ---- `.{findIndex,findLastIndex}(…)` compared against -1 / 0 ----

func checkFindIndexComparison(ctx rule.RuleContext, node *ast.Node) bool {
	bin := node.AsBinaryExpression()
	if bin.OperatorToken == nil {
		return false
	}

	operator := bin.OperatorToken.Kind
	rhsMatches := false
	switch operator {
	case ast.KindExclamationEqualsEqualsToken, ast.KindExclamationEqualsToken,
		ast.KindGreaterThanToken, ast.KindEqualsEqualsEqualsToken, ast.KindEqualsEqualsToken:
		rhsMatches = isNegativeOne(bin.Right)
	case ast.KindGreaterThanEqualsToken, ast.KindLessThanToken:
		rhsMatches = isLiteralZero(bin.Right)
	}
	if !rhsMatches {
		return false
	}

	argsLen := 1
	call, ok := unicornutil.MatchDotMethodCall(bin.Left, unicornutil.DotMethodCallOptions{
		Methods:             []string{"findIndex", "findLastIndex"},
		ArgumentsLength:     &argsLen,
		RejectSpreadElement: true,
	})
	if !ok {
		return false
	}

	if unicornutil.IsKnownNonIndexedCollection(ctx, call.Object) {
		return false
	}

	methodNode := call.Property
	ctx.ReportNodeWithDeferredFixes(methodNode, messageArraySome(methodNode.AsIdentifier().Text), func() []rule.RuleFix {
		// The operator token starts the `!== -1` / `>= 0` tail to remove. Use
		// its trimmed source range directly (robust across `>=` / `!==` forms
		// that a kind-matching token scan handles inconsistently).
		operatorStart := utils.TrimNodeTextRange(ctx.SourceFile, bin.OperatorToken).Pos()
		end := node.End()
		// Removing the comparison would drop comments in the removed range.
		if utils.HasCommentInSpan(ctx.Comments.All(), operatorStart, end) {
			return nil
		}

		fixes := make([]rule.RuleFix, 0, 3)
		// `=== -1` / `== -1` / `< 0` mean "not found", so negate.
		if operator == ast.KindEqualsEqualsEqualsToken || operator == ast.KindEqualsEqualsToken ||
			operator == ast.KindLessThanToken {
			fixes = append(fixes, insertBeforeRange(utils.TrimNodeTextRange(ctx.SourceFile, node), "!"))
		}
		fixes = append(fixes, rule.RuleFixReplace(ctx.SourceFile, methodNode, "some"))
		fixes = append(fixes, rule.RuleFixRemoveRange(core.NewTextRange(operatorStart, end)))
		return fixes
	})
	return true
}

// isNegativeOne matches `-1` (with any parenthesized / spaced form of the `1`).
// tsgo models `-1` as PrefixUnaryExpression(minus, NumericLiteral "1").
func isNegativeOne(node *ast.Node) bool {
	if node == nil {
		return false
	}
	node = ast.SkipParentheses(node)
	if node == nil || node.Kind != ast.KindPrefixUnaryExpression {
		return false
	}
	unary := node.AsPrefixUnaryExpression()
	if unary.Operator != ast.KindMinusToken {
		return false
	}
	if unary.Operand == nil {
		return false
	}
	operand := ast.SkipParentheses(unary.Operand)
	return operand != nil && operand.Kind == ast.KindNumericLiteral &&
		utils.NormalizeNumericLiteral(operand.AsNumericLiteral().Text) == "1"
}

func isLiteralZero(node *ast.Node) bool {
	if node == nil {
		return false
	}
	node = ast.SkipParentheses(node)
	return node != nil && node.Kind == ast.KindNumericLiteral &&
		utils.NormalizeNumericLiteral(node.AsNumericLiteral().Text) == "0"
}

// ---- `.filter(…).length > 0` / `!== 0` ----

func checkFilterLength(ctx rule.RuleContext, node *ast.Node) {
	bin := node.AsBinaryExpression()
	if bin.OperatorToken == nil {
		return
	}
	// We assume the user already follows `unicorn/explicit-length-check`; only
	// `> 0` and `!== 0` are allowed there.
	if bin.OperatorToken.Kind != ast.KindGreaterThanToken &&
		bin.OperatorToken.Kind != ast.KindExclamationEqualsEqualsToken {
		return
	}
	// The right side must be the RAW literal `0` — `0.` / `0x00` / `.0` are
	// distinct source forms upstream deliberately does not match. ESTree drops
	// parentheses, so upstream sees the literal through `(( 0 ))`; skip them
	// here to match, the inner literal's source text is still `0`.
	if !isRawZeroLiteral(ctx.SourceFile, ast.SkipParentheses(bin.Right)) {
		return
	}

	// The left side must be `<filterCall>.length` (non-optional dot access).
	lengthMember := ast.SkipParentheses(bin.Left)
	if lengthMember == nil || !ast.IsPropertyAccessExpression(lengthMember) ||
		ast.IsOptionalChain(lengthMember) {
		return
	}
	lengthAccess := lengthMember.AsPropertyAccessExpression()
	lengthName := lengthAccess.Name()
	if lengthName == nil || !ast.IsIdentifier(lengthName) || lengthName.AsIdentifier().Text != "length" {
		return
	}

	filterCall, ok := unicornutil.MatchDotMethodCall(
		ast.SkipParentheses(lengthAccess.Expression),
		unicornutil.DotMethodCallOptions{Method: "filter"},
	)
	if !ok {
		return
	}

	filterObject := filterCall.Object
	// `$`-prefixed receivers are treated as jQuery-like (`$element.filter(...)`).
	skippedObject := ast.SkipParentheses(filterObject)
	if skippedObject != nil && ast.IsIdentifier(skippedObject) &&
		len(skippedObject.AsIdentifier().Text) > 0 && skippedObject.AsIdentifier().Text[0] == '$' {
		return
	}
	if unicornutil.IsKnownNonIndexedCollection(ctx, filterObject) {
		return
	}

	filterArgs := filterCall.Call.Arguments()
	if len(filterArgs) == 0 || isNodeValueNotFunction(filterArgs[0]) {
		return
	}

	filterProperty := filterCall.Property
	ctx.ReportNodeWithDeferredFixes(filterProperty, messageArrayFilter(), func() []rule.RuleFix {
		lengthObjectEnd := parenthesizedEnd(ctx, filterCall.Call)
		end := node.End()
		// Removing `.length > 0` would drop comments in the removed range.
		if utils.HasCommentInSpan(ctx.Comments.All(), lengthObjectEnd, end) {
			return nil
		}

		// `.filter` → `.some`; remove `.length`; remove `> 0`.
		return []rule.RuleFix{
			rule.RuleFixReplace(ctx.SourceFile, filterProperty, "some"),
			// Remove `.length`: from the end of the filter call up to the end
			// of the member expression itself. Upstream's
			// removeMemberExpressionProperty stops at the member's own end, so
			// parentheses wrapping the member survive the removal.
			rule.RuleFixRemoveRange(core.NewTextRange(lengthObjectEnd, utils.TrimNodeTextRange(ctx.SourceFile, lengthMember).End())),
			// Remove `> 0`, starting after any parentheses around the member.
			rule.RuleFixRemoveRange(core.NewTextRange(parenthesizedEnd(ctx, lengthMember), end)),
		}
	})
}

// isRawZeroLiteral reports whether node is a numeric literal whose raw source
// text is exactly "0" (matching upstream's `right.raw === '0'`). Uses raw
// source text because NumericLiteral.Text is decimal-normalized at parse time,
// which would incorrectly fold `0.` / `0x00` into "0".
func isRawZeroLiteral(sourceFile *ast.SourceFile, node *ast.Node) bool {
	if node == nil || node.Kind != ast.KindNumericLiteral {
		return false
	}
	return scanner.GetSourceTextOfNodeFromSourceFile(sourceFile, node, false) == "0"
}

// isNodeValueNotFunction mirrors upstream's is-node-value-not-function helper.
// It rejects `.filter` arguments that cannot be a callback (literals, objects,
// arrays, etc.), while treating `.bind()` calls as possible functions.
func isNodeValueNotFunction(node *ast.Node) bool {
	switch node.Kind {
	case ast.KindArrayLiteralExpression,
		ast.KindObjectLiteralExpression,
		ast.KindClassExpression,
		ast.KindTemplateExpression,
		ast.KindNoSubstitutionTemplateLiteral,
		// ESTree's UnaryExpression covers `!x` / `-x` as well as `typeof x`,
		// `void x` and `delete x`; tsgo splits the last three into their own
		// kinds, so all five have to be listed here.
		ast.KindPrefixUnaryExpression,
		ast.KindPostfixUnaryExpression,
		ast.KindTypeOfExpression,
		ast.KindVoidExpression,
		ast.KindDeleteExpression,
		// Literals (ESTree collapses these into one `Literal` node type).
		ast.KindStringLiteral,
		ast.KindNumericLiteral,
		ast.KindBigIntLiteral,
		ast.KindRegularExpressionLiteral,
		ast.KindTrueKeyword,
		ast.KindFalseKeyword,
		ast.KindNullKeyword,
		// mostLikelyNotNodeTypes
		ast.KindAwaitExpression,
		ast.KindNewExpression,
		ast.KindTaggedTemplateExpression,
		ast.KindThisKeyword:
		return true
	case ast.KindBinaryExpression:
		// ESTree splits BinaryExpression / LogicalExpression / AssignmentExpression;
		// upstream lists BinaryExpression and AssignmentExpression as impossible,
		// LogicalExpression is not. Match by operator.
		operator := node.AsBinaryExpression().OperatorToken.Kind
		return isImpossibleBinaryOperator(operator)
	case ast.KindCallExpression:
		// A call could return a function only via `.bind()`.
		_, isBind := unicornutil.MatchDotMethodCall(node, unicornutil.DotMethodCallOptions{Method: "bind"})
		return !isBind
	}
	return utils.IsUndefinedIdentifier(node)
}

// isImpossibleBinaryOperator returns true for arithmetic / comparison / bitwise
// operators (ESTree BinaryExpression) and assignment operators (ESTree
// AssignmentExpression), but false for `&&` / `||` / `??` (ESTree
// LogicalExpression), matching upstream's impossible / most-likely-not sets.
func isImpossibleBinaryOperator(operator ast.Kind) bool {
	switch operator {
	case ast.KindAmpersandAmpersandToken, ast.KindBarBarToken, ast.KindQuestionQuestionToken,
		ast.KindCommaToken:
		return false
	}
	return true
}

// ---- boolean / control-flow position (upstream utils/boolean.js) ----

func isLogicalExpression(node *ast.Node) bool {
	if node == nil || !ast.IsBinaryExpression(node) {
		return false
	}
	operator := node.AsBinaryExpression().OperatorToken.Kind
	return operator == ast.KindAmpersandAmpersandToken || operator == ast.KindBarBarToken
}

func isLogicNot(node *ast.Node) bool {
	return node != nil && node.Kind == ast.KindPrefixUnaryExpression &&
		node.AsPrefixUnaryExpression().Operator == ast.KindExclamationToken
}

func isLogicNotArgument(node *ast.Node) bool {
	parent := effectiveParent(node)
	return isLogicNot(parent) &&
		ast.SkipParentheses(parent.AsPrefixUnaryExpression().Operand) == node
}

// isGlobalBooleanCall reports whether node is a `Boolean(...)` call to the
// global Boolean (single non-spread argument, not optional, not shadowed).
func isGlobalBooleanCall(ctx rule.RuleContext, node *ast.Node) bool {
	if node == nil || !ast.IsCallExpression(node) {
		return false
	}
	call := node.AsCallExpression()
	if call.QuestionDotToken != nil {
		return false
	}
	callee := ast.SkipParentheses(call.Expression)
	if callee == nil || !ast.IsIdentifier(callee) || callee.AsIdentifier().Text != "Boolean" {
		return false
	}
	args := node.Arguments()
	if len(args) != 1 || args[0].Kind == ast.KindSpreadElement {
		return false
	}
	return !utils.IsShadowed(callee, "Boolean")
}

func isBooleanCallArgument(ctx rule.RuleContext, node *ast.Node) bool {
	parent := effectiveParent(node)
	if !isGlobalBooleanCall(ctx, parent) {
		return false
	}
	args := parent.Arguments()
	return len(args) == 1 && ast.SkipParentheses(args[0]) == node
}

func isDirectBooleanExpression(ctx rule.RuleContext, node *ast.Node) bool {
	return isLogicNot(node) ||
		isLogicNotArgument(node) ||
		isGlobalBooleanCall(ctx, node) ||
		isBooleanCallArgument(ctx, node)
}

func isBooleanExpression(ctx rule.RuleContext, node *ast.Node) bool {
	if isDirectBooleanExpression(ctx, node) {
		return true
	}
	parent := effectiveParent(node)
	if isLogicalExpression(parent) {
		return isBooleanExpression(ctx, parent)
	}
	return false
}

func isDirectControlFlowTest(node *ast.Node) bool {
	parent := effectiveParent(node)
	if parent == nil {
		return false
	}
	switch parent.Kind {
	case ast.KindIfStatement:
		return ast.SkipParentheses(parent.AsIfStatement().Expression) == node
	case ast.KindConditionalExpression:
		return ast.SkipParentheses(parent.AsConditionalExpression().Condition) == node
	case ast.KindWhileStatement:
		return ast.SkipParentheses(parent.AsWhileStatement().Expression) == node
	case ast.KindDoStatement:
		return ast.SkipParentheses(parent.AsDoStatement().Expression) == node
	case ast.KindForStatement:
		return ast.SkipParentheses(parent.AsForStatement().Condition) == node
	}
	return false
}

func isControlFlowTest(node *ast.Node) bool {
	if isDirectControlFlowTest(node) {
		return true
	}
	parent := effectiveParent(node)
	if isLogicalExpression(parent) {
		return isControlFlowTest(parent)
	}
	return false
}

// ---- misc AST helpers ----

// effectiveParent returns node's nearest ancestor that is not a
// ParenthesizedExpression, mirroring how ESTree "flattens" parentheses so that
// a node's parent is the semantic construct wrapping it.
func effectiveParent(node *ast.Node) *ast.Node {
	if node == nil {
		return nil
	}
	return ast.WalkUpParenthesizedExpressions(node.Parent)
}

func isNullLiteral(node *ast.Node) bool {
	node = ast.SkipParentheses(node)
	return node != nil && node.Kind == ast.KindNullKeyword
}

// parenthesizedRange returns the outermost parenthesized range of node (its own
// range when unparenthesized), mirroring upstream's getParenthesizedRange.
func parenthesizedRange(ctx rule.RuleContext, node *ast.Node) core.TextRange {
	return utils.TrimNodeTextRange(ctx.SourceFile, outermostParenthesized(node))
}

// outermostParenthesized climbs from node through any enclosing
// ParenthesizedExpression parents and returns the outermost one (node itself
// when it isn't parenthesized).
func outermostParenthesized(node *ast.Node) *ast.Node {
	result := node
	for result.Parent != nil && result.Parent.Kind == ast.KindParenthesizedExpression {
		result = result.Parent
	}
	return result
}

func parenthesizedEnd(ctx rule.RuleContext, node *ast.Node) int {
	return parenthesizedRange(ctx, node).End()
}

// insertBeforeRange returns a zero-width fix that inserts text at the start of
// the given range.
func insertBeforeRange(textRange core.TextRange, text string) rule.RuleFix {
	return rule.RuleFixReplaceRange(textRange.WithEnd(textRange.Pos()), text)
}
