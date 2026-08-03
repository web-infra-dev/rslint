package prefer_array_some

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/scanner"
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

	if isKnownNonIndexedCollection(ctx, call.Object) {
		return
	}

	isCompare := isCheckingUndefined(node)
	if !isCompare &&
		!isBooleanExpression(ctx, node) &&
		!isControlFlowTest(node) &&
		!isFindResultVariableUsedOnlyAsBoolean(ctx, node, call) {
		return
	}

	methodNode := call.Property
	method := methodNode.AsIdentifier().Text

	comparison := comparisonParent(node) // non-nil only when isCompare

	// Removing the comparison would drop comments between the call and the end
	// of the comparison; upstream withholds the suggestion in that case.
	wouldDropComments := isCompare && utils.HasCommentInSpan(
		ctx.Comments.All(),
		parenthesizedEnd(ctx, node),
		comparison.End(),
	)

	if wouldDropComments {
		ctx.ReportNode(methodNode, messageArraySome(method))
		return
	}

	ctx.ReportNodeWithDeferredSuggestions(methodNode, messageArraySome(method), func() []rule.RuleSuggestion {
		fixes := []rule.RuleFix{rule.RuleFixReplace(ctx.SourceFile, methodNode, "some")}
		if isCompare {
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

// comparisonParent returns the BinaryExpression comparing `node` against
// undefined / null (per isCheckingUndefined), skipping parentheses around the
// call. Returns nil when node is not the left operand of such a comparison.
func comparisonParent(node *ast.Node) *ast.Node {
	parent := effectiveParent(node)
	if parent == nil || !ast.IsBinaryExpression(parent) {
		return nil
	}
	bin := parent.AsBinaryExpression()
	if ast.SkipParentheses(bin.Left) != node {
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

// isCheckingUndefined mirrors upstream: the call is the left operand of a
// comparison against `undefined` (`==`/`!=`/`===`/`!==`) or `null` (`==`/`!=`).
// Yoda comparisons (`undefined !== foo.find()`) are intentionally not matched.
func isCheckingUndefined(node *ast.Node) bool {
	return comparisonParent(node) != nil
}

// isFindResultVariableUsedOnlyAsBoolean mirrors upstream's helper: a
// `const x = arr.find(fn)` whose every read is used as a boolean. Requires the
// receiver to be a known array (upstream `isArray`), a plain single-name
// `const` binding that isn't exported, and ≥1 reference — all of which must be
// boolean-position reads.
func isFindResultVariableUsedOnlyAsBoolean(ctx rule.RuleContext, callExpression *ast.Node, call unicornutil.DotMethodCall) bool {
	if !isArray(ctx, call.Object) {
		return false
	}
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

	argsLen := 1
	call, ok := unicornutil.MatchDotMethodCall(bin.Left, unicornutil.DotMethodCallOptions{
		Methods:             []string{"findIndex", "findLastIndex"},
		ArgumentsLength:     &argsLen,
		RejectSpreadElement: true,
	})
	if !ok {
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

	if isKnownNonIndexedCollection(ctx, call.Object) {
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
	node = ast.SkipParentheses(node)
	if node == nil || node.Kind != ast.KindPrefixUnaryExpression {
		return false
	}
	unary := node.AsPrefixUnaryExpression()
	if unary.Operator != ast.KindMinusToken {
		return false
	}
	operand := ast.SkipParentheses(unary.Operand)
	return operand != nil && operand.Kind == ast.KindNumericLiteral &&
		utils.NormalizeNumericLiteral(operand.AsNumericLiteral().Text) == "1"
}

func isLiteralZero(node *ast.Node) bool {
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
	if isKnownNonIndexedCollection(ctx, filterObject) {
		return
	}

	filterArgs := filterCall.Call.Arguments()
	if len(filterArgs) == 0 || isNodeValueNotFunction(filterArgs[0]) {
		return
	}

	filterProperty := filterCall.Property
	ctx.ReportNodeWithDeferredFixes(filterProperty, messageArrayFilter(), func() []rule.RuleFix {
		lengthObjectEnd := parenthesizedEndOf(ctx, filterCall.Call)
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
			rule.RuleFixRemoveRange(core.NewTextRange(parenthesizedEndOf(ctx, lengthMember), end)),
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
		ast.KindPrefixUnaryExpression,
		ast.KindPostfixUnaryExpression,
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

func parenthesizedEndOf(ctx rule.RuleContext, node *ast.Node) int {
	return parenthesizedEnd(ctx, node)
}

// insertBeforeRange returns a zero-width fix that inserts text at the start of
// the given range.
func insertBeforeRange(textRange core.TextRange, text string) rule.RuleFix {
	return rule.RuleFixReplaceRange(textRange.WithEnd(textRange.Pos()), text)
}

// ---- type classification (upstream utils/is-array.js) ----

// isKnownNonIndexedCollection reports whether the receiver is definitely a
// keyed collection (`Map`/`Set`/`WeakMap`/`WeakSet`/`ReadonlyMap`/`ReadonlySet`)
// — types that carry a `.find` / `.filter` / `.findIndex` surface but are not
// arrays, so the rewrite would be wrong. Returns false when the type is unknown
// (including when no TypeChecker is available), matching upstream.
func isKnownNonIndexedCollection(ctx rule.RuleContext, node *ast.Node) bool {
	return classifyReceiver(ctx, node, indexedCollectionTargets, keyedCollectionNames) == classNonTarget
}

// isArray reports whether the receiver is definitely an Array / ReadonlyArray /
// tuple, using syntactic shape first and the type checker as a fallback.
func isArray(ctx rule.RuleContext, node *ast.Node) bool {
	return classifyReceiver(ctx, node, arrayTargets, knownNonArrayNames) == classTarget
}

type typeClass int

const (
	classUnknown typeClass = iota
	classTarget
	classNonTarget
)

var arrayTargets = utils.NewSetFromItems("Array", "ReadonlyArray")

var indexedCollectionTargets = func() *utils.Set[string] {
	s := utils.NewSetFromItems("Array", "ReadonlyArray")
	for _, name := range typedArrayNames {
		s.Add(name)
	}
	return s
}()

var keyedCollectionNames = utils.NewSetFromItems(
	"Map", "ReadonlyMap", "WeakMap", "Set", "ReadonlySet", "WeakSet",
)

var knownNonArrayNames = func() *utils.Set[string] {
	s := utils.NewSetFromItems(
		"Map", "ReadonlyMap", "WeakMap", "Set", "ReadonlySet", "WeakSet",
	)
	for _, name := range typedArrayNames {
		s.Add(name)
	}
	return s
}()

var typedArrayNames = []string{
	"Int8Array", "Uint8Array", "Uint8ClampedArray", "Int16Array", "Uint16Array",
	"Int32Array", "Uint32Array", "Float16Array", "Float32Array", "Float64Array",
	"BigInt64Array", "BigUint64Array",
}

// classifyReceiver mirrors the syntactic short-circuits of unicorn's `getType`
// (array literals, `Array()` / `new Array()` / `Array.from` / `Array.of`
// constructors) and then falls back to a type-checker classification analogous
// to `getTypeScriptType`. `targetNames` and `nonTargetNames` are the type-name
// sets that resolve to target / non-target respectively.
func classifyReceiver(ctx rule.RuleContext, node *ast.Node, targetNames, nonTargetNames *utils.Set[string]) typeClass {
	node = ast.SkipParentheses(node)
	if node == nil {
		return classUnknown
	}

	// Syntactic target shapes shared by both classifications: an array literal
	// or an `Array` constructor / factory call is always an array (target for
	// arrays; not a keyed collection either, so also non-target there).
	if isSyntacticArrayNode(node) {
		if targetNames.Has("Array") {
			return classTarget
		}
		return classNonTarget
	}

	// Syntactic non-array shapes (upstream isNonArrayNode): `new X()` and
	// object / function / arrow / class / template literals are never arrays,
	// so they resolve to non-target for both classifications without needing
	// the type system.
	if isSyntacticNonArrayNode(node) {
		return classNonTarget
	}

	// A plain `const x = <initializer>` reference resolves to the syntactic
	// class of its initializer (upstream getTypeFromVariable). This catches
	// shapes the checker alone can miss on error types, e.g.
	// `const store = new Store()` where `Store` is undeclared.
	if ast.IsIdentifier(node) {
		if init := constInitializer(ctx, node); init != nil {
			switch {
			case isSyntacticArrayNode(init):
				if targetNames.Has("Array") {
					return classTarget
				}
				return classNonTarget
			case isSyntacticNonArrayNode(init):
				return classNonTarget
			}
		}
	}

	if ctx.TypeChecker == nil {
		return classUnknown
	}
	t := utils.GetConstrainedTypeAtLocation(ctx.TypeChecker, node)
	return classifyType(ctx, t, targetNames, nonTargetNames)
}

// constInitializer returns the initializer expression of a `const` variable
// that idNode references, or nil. Only single-declaration `const` bindings with
// a plain identifier name are resolved (matching upstream's getTypeFromVariable
// preconditions).
func constInitializer(ctx rule.RuleContext, idNode *ast.Node) *ast.Node {
	if ctx.TypeChecker == nil {
		return nil
	}
	sym := ctx.TypeChecker.GetSymbolAtLocation(idNode)
	if sym == nil || len(sym.Declarations) != 1 {
		return nil
	}
	decl := sym.Declarations[0]
	if decl == nil || !ast.IsVariableDeclaration(decl) {
		return nil
	}
	parent := decl.Parent
	if parent == nil || !ast.IsVariableDeclarationList(parent) || parent.Flags&ast.NodeFlagsConst == 0 {
		return nil
	}
	varDecl := decl.AsVariableDeclaration()
	if varDecl.Name() == nil || !ast.IsIdentifier(varDecl.Name()) || varDecl.Type != nil {
		return nil
	}
	return ast.SkipParentheses(varDecl.Initializer)
}

// isSyntacticArrayNode matches `[]`, `Array(...)`, `new Array(...)`,
// `Array.from(...)`, and `Array.of(...)` — the shapes upstream's isArrayNode
// treats as a definite array without consulting the type system.
func isSyntacticArrayNode(node *ast.Node) bool {
	if node.Kind == ast.KindArrayLiteralExpression {
		return true
	}
	if ast.IsCallExpression(node) {
		callee := ast.SkipParentheses(node.AsCallExpression().Expression)
		if callee == nil {
			return false
		}
		// `Array.from(...)` / `Array.of(...)`
		if ast.IsPropertyAccessExpression(callee) && !ast.IsOptionalChain(callee) {
			pae := callee.AsPropertyAccessExpression()
			obj := ast.SkipParentheses(pae.Expression)
			name := pae.Name()
			if obj != nil && ast.IsIdentifier(obj) && obj.AsIdentifier().Text == "Array" &&
				name != nil && ast.IsIdentifier(name) {
				m := name.AsIdentifier().Text
				return m == "from" || m == "of"
			}
			return false
		}
		// `Array(...)`
		if ast.IsIdentifier(callee) {
			return callee.AsIdentifier().Text == "Array"
		}
		return false
	}
	// `new Array(...)`
	if ast.IsNewExpression(node) {
		callee := ast.SkipParentheses(node.AsNewExpression().Expression)
		return callee != nil && ast.IsIdentifier(callee) && callee.AsIdentifier().Text == "Array"
	}
	return false
}

// isSyntacticNonArrayNode mirrors upstream's isNonArrayNode: `new X()` and
// object / function / arrow / class / template literals are definitely not
// arrays. (An `Array` constructor / factory is handled by isSyntacticArrayNode
// before this is consulted, so `new Array()` never reaches here as non-array.)
func isSyntacticNonArrayNode(node *ast.Node) bool {
	switch node.Kind {
	case ast.KindNewExpression,
		ast.KindObjectLiteralExpression,
		ast.KindFunctionExpression,
		ast.KindArrowFunction,
		ast.KindClassExpression,
		ast.KindTemplateExpression,
		ast.KindNoSubstitutionTemplateLiteral:
		return true
	}
	return false
}

// classifyType mirrors unicorn's getTypeScriptType: recurse through
// unions/intersections/type-parameter constraints, then decide by the resolved
// symbol name.
func classifyType(ctx rule.RuleContext, t *checker.Type, targetNames, nonTargetNames *utils.Set[string]) typeClass {
	if t == nil {
		return classUnknown
	}
	if utils.IsTypeAnyType(t) || utils.IsTypeUnknownType(t) || utils.IsIntrinsicErrorType(t) {
		return classUnknown
	}
	// null / undefined are nullish → treated as non-target by both callers.
	if utils.IsTypeFlagSet(t, checker.TypeFlagsNull|checker.TypeFlagsUndefined) {
		return classNonTarget
	}

	if utils.IsTypeParameter(t) {
		constraint := checker.Checker_getBaseConstraintOfType(ctx.TypeChecker, t)
		if constraint == nil {
			return classUnknown
		}
		return classifyType(ctx, constraint, targetNames, nonTargetNames)
	}

	if utils.IsUnionType(t) {
		return combineUnion(ctx, utils.UnionTypeParts(t), targetNames, nonTargetNames)
	}
	if utils.IsIntersectionType(t) {
		return combineIntersection(ctx, utils.IntersectionTypeParts(t), targetNames, nonTargetNames)
	}

	// Array / tuple structural check (mirrors upstream isTargetType).
	if targetNames.Has("Array") && checker.Checker_isArrayOrTupleType(ctx.TypeChecker, t) {
		return classTarget
	}

	// Base constraint fallback (e.g. `T extends string[]` reached via apparent
	// constraint rather than the type-parameter branch above).
	constraint := checker.Checker_getBaseConstraintOfType(ctx.TypeChecker, t)
	if constraint != nil && constraint != t {
		return classifyType(ctx, constraint, targetNames, nonTargetNames)
	}

	name, ok := typeSymbolName(t)
	if !ok {
		// A primitive (string / number / boolean / bigint / symbol, literal or
		// not) or any other intrinsic is never an Array or keyed collection →
		// non-target. Mirrors upstream, where such a value resolves to
		// non-target via its intrinsicName or static value.
		if utils.IsTypeFlagSet(t, checker.TypeFlagsPrimitive|checker.TypeFlagsIntrinsic) {
			return classNonTarget
		}
		return classUnknown
	}
	if targetNames.Has(name) {
		return classTarget
	}
	// Interface / class heritage: `interface Items extends Array<string>` (and
	// `class X extends Array`) resolve to the array target through their base
	// types. Upstream reaches the same conclusion via its interface-heritage
	// walk. Only consult heritage when the receiver could be an array target;
	// keyed-collection classification does not widen through heritage.
	if targetNames.Has("Array") && classifyHeritage(ctx, t, targetNames, nonTargetNames) == classTarget {
		return classTarget
	}
	return classNonTarget
}

// classifyHeritage inspects the base types of a class / interface symbol,
// returning classTarget when any base resolves to the target classification.
func classifyHeritage(ctx rule.RuleContext, t *checker.Type, targetNames, nonTargetNames *utils.Set[string]) typeClass {
	symbol := checker.Type_symbol(t)
	if symbol == nil || symbol.Flags&(ast.SymbolFlagsClass|ast.SymbolFlagsInterface) == 0 {
		return classUnknown
	}
	declared := checker.Checker_getDeclaredTypeOfSymbol(ctx.TypeChecker, symbol)
	if declared == nil {
		return classUnknown
	}
	for _, base := range checker.Checker_getBaseTypes(ctx.TypeChecker, declared) {
		if classifyType(ctx, base, targetNames, nonTargetNames) == classTarget {
			return classTarget
		}
	}
	return classUnknown
}

func combineUnion(ctx rule.RuleContext, parts []*checker.Type, targetNames, nonTargetNames *utils.Set[string]) typeClass {
	// Filter nullish parts (normalizeType folds nullish → nonTarget, and a
	// union is target only when every non-nullish part is target).
	filtered := make([]typeClass, 0, len(parts))
	for _, part := range parts {
		if utils.IsTypeFlagSet(part, checker.TypeFlagsNull|checker.TypeFlagsUndefined) {
			continue
		}
		filtered = append(filtered, classifyType(ctx, part, targetNames, nonTargetNames))
	}
	if len(filtered) == 0 {
		return classNonTarget
	}
	allTarget := true
	allNonTarget := true
	for _, c := range filtered {
		if c != classTarget {
			allTarget = false
		}
		if c != classNonTarget {
			allNonTarget = false
		}
	}
	if allTarget {
		return classTarget
	}
	if allNonTarget {
		return classNonTarget
	}
	return classUnknown
}

func combineIntersection(ctx rule.RuleContext, parts []*checker.Type, targetNames, nonTargetNames *utils.Set[string]) typeClass {
	classes := make([]typeClass, 0, len(parts))
	for _, part := range parts {
		classes = append(classes, classifyType(ctx, part, targetNames, nonTargetNames))
	}
	for _, c := range classes {
		if c == classTarget {
			return classTarget
		}
	}
	for _, c := range classes {
		if c != classNonTarget {
			return classUnknown
		}
	}
	return classNonTarget
}

// typeSymbolName returns the symbol name of a type (its own symbol, falling
// back to the alias symbol), mirroring upstream's `type.getSymbol() ??
// type.aliasSymbol` + `symbol.getName()`.
func typeSymbolName(t *checker.Type) (string, bool) {
	if sym := checker.Type_symbol(t); sym != nil && sym.Name != "" {
		return sym.Name, true
	}
	if alias := checker.Type_alias(t); alias != nil {
		if sym := alias.Symbol(); sym != nil && sym.Name != "" {
			return sym.Name, true
		}
	}
	return "", false
}
