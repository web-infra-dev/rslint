package prefer_nullish_coalescing

import (
	"fmt"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func buildNoStrictNullCheckMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "noStrictNullCheck",
		Description: "This rule requires the `strictNullChecks` compiler option to be turned on to function correctly.",
	}
}

func buildPreferNullishOverAssignmentMessage(equals string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "preferNullishOverAssignment",
		Description: fmt.Sprintf("Prefer using nullish coalescing operator (`??%s`) instead of an assignment expression, as it is simpler to read.", equals),
	}
}

func buildPreferNullishOverOrMessage(description, equals string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "preferNullishOverOr",
		Description: fmt.Sprintf("Prefer using nullish coalescing operator (`??%s`) instead of a logical %s (`||%s`), as it is a safer operator.", equals, description, equals),
	}
}

func buildPreferNullishOverTernaryMessage(equals string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "preferNullishOverTernary",
		Description: fmt.Sprintf("Prefer using nullish coalescing operator (`??%s`) instead of a ternary expression, as it is simpler to read.", equals),
	}
}

func buildSuggestNullishMessage(equals string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "suggestNullish",
		Description: fmt.Sprintf("Fix to nullish coalescing operator (`??%s`).", equals),
	}
}

type ignorePrimitivesObj struct {
	bigint  bool
	boolean bool
	number  bool
	string_ bool
}

type Options struct {
	allowRuleToRunWithoutStrictNullChecksIKnowWhatIAmDoing bool
	ignoreBooleanCoercion                                  bool
	ignoreConditionalTests                                 bool
	ignoreIfStatements                                     bool
	ignoreMixedLogicalExpressions                          bool
	ignorePrimitivesAll                                    bool
	ignorePrimitives                                       ignorePrimitivesObj
	ignoreTernaryTests                                     bool
}

func defaultOptions() Options {
	return Options{
		ignoreConditionalTests: true,
	}
}

func parseOptions(options any) Options {
	opts := defaultOptions()
	optsMap := utils.GetOptionsMap(options)
	if optsMap == nil {
		return opts
	}
	if v, ok := optsMap["allowRuleToRunWithoutStrictNullChecksIKnowWhatIAmDoing"].(bool); ok {
		opts.allowRuleToRunWithoutStrictNullChecksIKnowWhatIAmDoing = v
	}
	if v, ok := optsMap["ignoreBooleanCoercion"].(bool); ok {
		opts.ignoreBooleanCoercion = v
	}
	if v, ok := optsMap["ignoreConditionalTests"].(bool); ok {
		opts.ignoreConditionalTests = v
	}
	if v, ok := optsMap["ignoreIfStatements"].(bool); ok {
		opts.ignoreIfStatements = v
	}
	if v, ok := optsMap["ignoreMixedLogicalExpressions"].(bool); ok {
		opts.ignoreMixedLogicalExpressions = v
	}
	if v, ok := optsMap["ignoreTernaryTests"].(bool); ok {
		opts.ignoreTernaryTests = v
	}
	if v, ok := optsMap["ignorePrimitives"]; ok {
		switch vv := v.(type) {
		case bool:
			if vv {
				opts.ignorePrimitivesAll = true
			}
		case map[string]interface{}:
			if b, ok := vv["bigint"].(bool); ok {
				opts.ignorePrimitives.bigint = b
			}
			if b, ok := vv["boolean"].(bool); ok {
				opts.ignorePrimitives.boolean = b
			}
			if b, ok := vv["number"].(bool); ok {
				opts.ignorePrimitives.number = b
			}
			if b, ok := vv["string"].(bool); ok {
				opts.ignorePrimitives.string_ = b
			}
		}
	}
	return opts
}

// nullishCheckOperator: '!', '!=', '!==', '', '==', '==='
type nullishCheckOperator string

const (
	opTruthy   nullishCheckOperator = ""
	opNotTruth nullishCheckOperator = "!"
	opEqEq     nullishCheckOperator = "=="
	opEqEqEq   nullishCheckOperator = "==="
	opNotEq    nullishCheckOperator = "!="
	opNotEqEq  nullishCheckOperator = "!=="
)

func isLogicalOrLikeOperator(op ast.Kind) bool {
	return op == ast.KindBarBarToken || op == ast.KindBarBarEqualsToken
}

// isLogicalLikeOperator covers exactly what ESTree models as a
// `LogicalExpression`: `||`, `&&`, `??`. The compound-assignment forms
// (`||=`, `&&=`, `??=`) are `AssignmentExpression` in ESTree and upstream's
// parent walks (`isConditionalTest` / `isBooleanConstructorContext`) don't
// recurse through them, so neither do we.
func isLogicalLikeOperator(op ast.Kind) bool {
	switch op {
	case ast.KindBarBarToken,
		ast.KindAmpersandAmpersandToken,
		ast.KindQuestionQuestionToken:
		return true
	}
	return false
}

// isMemberAccessLike mirrors upstream's isNodeOfTypes([ChainExpression,
// Identifier, MemberExpression]) exactly. In tsgo there's no
// `ChainExpression` wrapper — optional chains are flagged on the access
// expression itself — so PropertyAccessExpression / ElementAccessExpression
// cover both the `MemberExpression` AND `ChainExpression` ESTree shapes.
// `this` (ThisExpression) and `super` (Super) are NOT in upstream's set and
// must be excluded here for parity.
func isMemberAccessLike(node *ast.Node) bool {
	if node == nil {
		return false
	}
	switch node.Kind {
	case ast.KindIdentifier,
		ast.KindPropertyAccessExpression,
		ast.KindElementAccessExpression:
		return true
	}
	return false
}

func isNullLiteralOrUndefinedIdentifier(node *ast.Node) bool {
	return utils.IsNullLiteral(node) || utils.IsUndefinedIdentifier(node)
}

func isNodeNullishComparisonBinary(bin *ast.BinaryExpression) bool {
	return isNullLiteralOrUndefinedIdentifier(bin.Left) && isNullLiteralOrUndefinedIdentifier(bin.Right)
}

// isConditionalTest walks parents to see if the expression's value is used as
// the test of a conditional/loop construct.
func isConditionalTest(node *ast.Node) bool {
	parent := node.Parent
	if parent == nil {
		return false
	}
	if parent.Kind == ast.KindParenthesizedExpression {
		return isConditionalTest(parent)
	}
	if parent.Kind == ast.KindBinaryExpression {
		bin := parent.AsBinaryExpression()
		if isLogicalLikeOperator(bin.OperatorToken.Kind) {
			return isConditionalTest(parent)
		}
		// SequenceExpression (`,`): only the trailing element flows up.
		if bin.OperatorToken.Kind == ast.KindCommaToken {
			if bin.Right == node {
				return isConditionalTest(parent)
			}
			return false
		}
	}
	if parent.Kind == ast.KindConditionalExpression {
		ce := parent.AsConditionalExpression()
		if ce.WhenTrue == node || ce.WhenFalse == node {
			return isConditionalTest(parent)
		}
		if ce.Condition == node {
			return true
		}
	}
	if parent.Kind == ast.KindPrefixUnaryExpression {
		if parent.AsPrefixUnaryExpression().Operator == ast.KindExclamationToken {
			return isConditionalTest(parent)
		}
	}
	switch parent.Kind {
	case ast.KindIfStatement:
		return parent.AsIfStatement().Expression == node
	case ast.KindWhileStatement:
		return parent.AsWhileStatement().Expression == node
	case ast.KindDoStatement:
		return parent.AsDoStatement().Expression == node
	case ast.KindForStatement:
		return parent.AsForStatement().Condition == node
	}
	return false
}

// isBooleanConstructorContext walks parents to see if this expression is used
// as the first argument of a global Boolean(...) call.
func isBooleanConstructorContext(node *ast.Node) bool {
	parent := node.Parent
	if parent == nil {
		return false
	}
	if parent.Kind == ast.KindParenthesizedExpression {
		return isBooleanConstructorContext(parent)
	}
	if parent.Kind == ast.KindBinaryExpression {
		bin := parent.AsBinaryExpression()
		if isLogicalLikeOperator(bin.OperatorToken.Kind) {
			return isBooleanConstructorContext(parent)
		}
		if bin.OperatorToken.Kind == ast.KindCommaToken {
			if bin.Right == node {
				return isBooleanConstructorContext(parent)
			}
			return false
		}
	}
	if parent.Kind == ast.KindConditionalExpression {
		ce := parent.AsConditionalExpression()
		if ce.WhenTrue == node || ce.WhenFalse == node {
			return isBooleanConstructorContext(parent)
		}
	}
	return isBuiltInBooleanCall(parent, node)
}

// isBuiltInBooleanCall reports whether `callNode` is a CallExpression to the
// global Boolean(...) and `child` (paren-skipped) is its first argument.
func isBuiltInBooleanCall(callNode *ast.Node, child *ast.Node) bool {
	if callNode.Kind != ast.KindCallExpression {
		return false
	}
	call := callNode.AsCallExpression()
	callee := ast.SkipParentheses(call.Expression)
	if callee == nil || callee.Kind != ast.KindIdentifier {
		return false
	}
	if callee.AsIdentifier().Text != "Boolean" {
		return false
	}
	args := call.Arguments
	if args == nil || len(args.Nodes) == 0 {
		return false
	}
	// match either bare or paren-wrapped child
	first := args.Nodes[0]
	if first != child && ast.SkipParentheses(first) != child {
		return false
	}
	return !utils.IsShadowed(callNode, "Boolean")
}

// isMixedLogicalExpression mirrors upstream's BFS over parent + left + right
// looking for an `&&` operator.
func isMixedLogicalExpression(node *ast.Node) bool {
	seen := map[*ast.Node]bool{}
	bin := node.AsBinaryExpression()
	queue := []*ast.Node{node.Parent, bin.Left, bin.Right}
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		if current == nil || seen[current] {
			continue
		}
		seen[current] = true
		if current.Kind == ast.KindParenthesizedExpression {
			queue = append(queue, current.AsParenthesizedExpression().Expression)
			continue
		}
		if current.Kind == ast.KindBinaryExpression {
			cb := current.AsBinaryExpression()
			op := cb.OperatorToken.Kind
			if op == ast.KindAmpersandAmpersandToken {
				return true
			}
			if op == ast.KindBarBarToken || op == ast.KindBarBarEqualsToken {
				queue = append(queue, current.Parent, cb.Left, cb.Right)
			}
		}
	}
	return false
}

// areNodesSimilarMemberAccess mirrors typescript-eslint's
// areNodesSimilarMemberAccess + isNodeEqual byte-for-byte — including the
// upstream limitation that `isNodeEqual` only handles a small set of node
// kinds (ThisExpression / Literal / Identifier / MemberExpression). Anything
// else (PrivateIdentifier, TSAsExpression-wrapped computed keys, …) falls
// through to `return false`, which means rules don't pair the two branches.
//
// We deliberately NOT generalize this with `AreNodesStructurallyEqual` —
// upstream's narrower compare is part of the rule's contract; widening it
// would cause spurious reports relative to ESLint.
func areNodesSimilarMemberAccess(sf *ast.SourceFile, a, b *ast.Node) bool {
	a = ast.SkipParentheses(a)
	b = ast.SkipParentheses(b)
	if a == nil || b == nil {
		return false
	}
	if a.Kind == ast.KindNonNullExpression {
		return areNodesSimilarMemberAccess(sf, a.AsNonNullExpression().Expression, b)
	}
	if b.Kind == ast.KindNonNullExpression {
		return areNodesSimilarMemberAccess(sf, a, b.AsNonNullExpression().Expression)
	}
	aIsMember := isMemberExpressionLike(a)
	bIsMember := isMemberExpressionLike(b)
	if aIsMember && bIsMember {
		// Recurse on objects.
		if !areNodesSimilarMemberAccess(sf, utils.AccessExpressionObject(a), utils.AccessExpressionObject(b)) {
			return false
		}
		// Property comparison — mirrors upstream's two-branch logic:
		//   if (a.computed === b.computed) return isNodeEqual(a.property, b.property)
		//   else if (a.property is Literal && b.property is Identifier) name-compare
		//   else if vice versa, name-compare
		//   else false
		aProp, aComputed := memberProperty(a)
		bProp, bComputed := memberProperty(b)
		if aComputed == bComputed {
			return isNodeEqual(aProp, bProp)
		}
		if isStringLikeLiteral(aProp) && bProp.Kind == ast.KindIdentifier {
			return literalString(aProp) == bProp.AsIdentifier().Text
		}
		if aProp.Kind == ast.KindIdentifier && isStringLikeLiteral(bProp) {
			return aProp.AsIdentifier().Text == literalString(bProp)
		}
		return false
	}
	if aIsMember != bIsMember {
		return false
	}
	return isNodeEqual(a, b)
}

// isMemberExpressionLike reports whether `node` is a MemberExpression-shaped
// access (PropertyAccessExpression / ElementAccessExpression in tsgo terms).
// Mirrors upstream's "MemberExpression" kind.
func isMemberExpressionLike(node *ast.Node) bool {
	return node.Kind == ast.KindPropertyAccessExpression ||
		node.Kind == ast.KindElementAccessExpression
}

// memberProperty returns the property node and whether it's a computed access.
//   - PropertyAccessExpression: property = `.name` (Identifier / PrivateIdentifier), computed=false
//   - ElementAccessExpression : property = `[arg]` (any Expression), computed=true
func memberProperty(node *ast.Node) (*ast.Node, bool) {
	switch node.Kind {
	case ast.KindPropertyAccessExpression:
		return node.AsPropertyAccessExpression().Name(), false
	case ast.KindElementAccessExpression:
		return node.AsElementAccessExpression().ArgumentExpression, true
	}
	return nil, false
}

// isNodeEqual mirrors upstream's restricted-set isNodeEqual exactly.
// Only handles ThisExpression / Literal / Identifier / MemberExpression;
// returns false for any other kind (PrivateIdentifier, TSAsExpression, etc.).
func isNodeEqual(a, b *ast.Node) bool {
	if a == nil || b == nil {
		return false
	}
	a = ast.SkipParentheses(a)
	b = ast.SkipParentheses(b)
	if a == nil || b == nil {
		return false
	}
	switch a.Kind {
	case ast.KindThisKeyword:
		return b.Kind == ast.KindThisKeyword
	case ast.KindIdentifier:
		if b.Kind != ast.KindIdentifier {
			return false
		}
		return a.AsIdentifier().Text == b.AsIdentifier().Text
	case ast.KindStringLiteral, ast.KindNoSubstitutionTemplateLiteral:
		if !isStringLikeLiteral(b) {
			return false
		}
		return literalString(a) == literalString(b)
	case ast.KindNumericLiteral:
		if b.Kind != ast.KindNumericLiteral {
			return false
		}
		return a.AsNumericLiteral().Text == b.AsNumericLiteral().Text
	case ast.KindBigIntLiteral:
		if b.Kind != ast.KindBigIntLiteral {
			return false
		}
		return a.AsBigIntLiteral().Text == b.AsBigIntLiteral().Text
	case ast.KindTrueKeyword, ast.KindFalseKeyword, ast.KindNullKeyword:
		return a.Kind == b.Kind
	case ast.KindPropertyAccessExpression, ast.KindElementAccessExpression:
		if !isMemberExpressionLike(b) {
			return false
		}
		aProp, aComputed := memberProperty(a)
		bProp, bComputed := memberProperty(b)
		if aComputed != bComputed {
			return false
		}
		return isNodeEqual(aProp, bProp) &&
			isNodeEqual(utils.AccessExpressionObject(a), utils.AccessExpressionObject(b))
	}
	// Any other kind (PrivateIdentifier, TSAsExpression, ComputedPropertyName,
	// SpreadElement, …) — upstream's isNodeEqual returns false here, and so
	// do we. Don't widen.
	return false
}

// isStringLikeLiteral reports whether the node is a StringLiteral or a
// no-substitution template literal — both map to ESTree's `Literal` with a
// string value.
func isStringLikeLiteral(node *ast.Node) bool {
	if node == nil {
		return false
	}
	return node.Kind == ast.KindStringLiteral ||
		node.Kind == ast.KindNoSubstitutionTemplateLiteral
}

// literalString returns the cooked string value of a string-like literal.
func literalString(node *ast.Node) string {
	switch node.Kind {
	case ast.KindStringLiteral:
		return node.AsStringLiteral().Text
	case ast.KindNoSubstitutionTemplateLiteral:
		return node.AsNoSubstitutionTemplateLiteral().Text
	}
	return ""
}

func getOperatorAndNodesInsideTestExpression(test *ast.Node) (operator nullishCheckOperator, nodes []*ast.Node, ok bool) {
	test = ast.SkipParentheses(test)
	if test == nil {
		return "", nil, false
	}
	// Member-access-like (truthy check).
	if isMemberAccessLike(test) {
		return opTruthy, nil, true
	}
	if test.Kind == ast.KindPrefixUnaryExpression {
		un := test.AsPrefixUnaryExpression()
		if un.Operator == ast.KindExclamationToken {
			arg := ast.SkipParentheses(un.Operand)
			if isMemberAccessLike(arg) {
				return opNotTruth, nil, true
			}
			return "", nil, false
		}
		return "", nil, false
	}
	if test.Kind != ast.KindBinaryExpression {
		return "", nil, false
	}
	bin := test.AsBinaryExpression()
	op := bin.OperatorToken.Kind
	switch op {
	case ast.KindEqualsEqualsToken:
		return opEqEq, []*ast.Node{bin.Left, bin.Right}, true
	case ast.KindEqualsEqualsEqualsToken:
		return opEqEqEq, []*ast.Node{bin.Left, bin.Right}, true
	case ast.KindExclamationEqualsToken:
		return opNotEq, []*ast.Node{bin.Left, bin.Right}, true
	case ast.KindExclamationEqualsEqualsToken:
		return opNotEqEq, []*ast.Node{bin.Left, bin.Right}, true
	}
	if op != ast.KindBarBarToken && op != ast.KindBarBarEqualsToken && op != ast.KindAmpersandAmpersandToken {
		return "", nil, false
	}
	leftRaw := ast.SkipParentheses(bin.Left)
	rightRaw := ast.SkipParentheses(bin.Right)
	if leftRaw == nil || rightRaw == nil ||
		leftRaw.Kind != ast.KindBinaryExpression ||
		rightRaw.Kind != ast.KindBinaryExpression {
		return "", nil, false
	}
	leftBin := leftRaw.AsBinaryExpression()
	rightBin := rightRaw.AsBinaryExpression()
	if isNodeNullishComparisonBinary(leftBin) || isNodeNullishComparisonBinary(rightBin) {
		return "", nil, false
	}
	lo := leftBin.OperatorToken.Kind
	ro := rightBin.OperatorToken.Kind
	nodes = []*ast.Node{leftBin.Left, leftBin.Right, rightBin.Left, rightBin.Right}
	switch op {
	case ast.KindBarBarToken, ast.KindBarBarEqualsToken:
		if lo == ast.KindEqualsEqualsEqualsToken && ro == ast.KindEqualsEqualsEqualsToken {
			return opEqEqEq, nodes, true
		}
		both := lo == ast.KindEqualsEqualsToken && ro == ast.KindEqualsEqualsToken
		mixed := (lo == ast.KindEqualsEqualsEqualsToken || ro == ast.KindEqualsEqualsEqualsToken) &&
			(lo == ast.KindEqualsEqualsToken || ro == ast.KindEqualsEqualsToken)
		if both || mixed {
			return opEqEq, nodes, true
		}
	case ast.KindAmpersandAmpersandToken:
		if lo == ast.KindExclamationEqualsEqualsToken && ro == ast.KindExclamationEqualsEqualsToken {
			return opNotEqEq, nodes, true
		}
		both := lo == ast.KindExclamationEqualsToken && ro == ast.KindExclamationEqualsToken
		mixed := (lo == ast.KindExclamationEqualsEqualsToken || ro == ast.KindExclamationEqualsEqualsToken) &&
			(lo == ast.KindExclamationEqualsToken || ro == ast.KindExclamationEqualsToken)
		if both || mixed {
			return opNotEq, nodes, true
		}
	}
	return "", nil, false
}

func getBranchNodes(ce *ast.ConditionalExpression, op nullishCheckOperator) (nonNullish, nullish *ast.Node) {
	switch op {
	case opTruthy, opNotEq, opNotEqEq:
		return ce.WhenTrue, ce.WhenFalse
	}
	return ce.WhenFalse, ce.WhenTrue
}

func (opts *Options) typeFlagsToIgnore() checker.TypeFlags {
	var flags checker.TypeFlags
	if opts.ignorePrimitivesAll || opts.ignorePrimitives.bigint {
		flags |= checker.TypeFlagsBigIntLike
	}
	if opts.ignorePrimitivesAll || opts.ignorePrimitives.boolean {
		flags |= checker.TypeFlagsBooleanLike
	}
	if opts.ignorePrimitivesAll || opts.ignorePrimitives.number {
		flags |= checker.TypeFlagsNumberLike
	}
	if opts.ignorePrimitivesAll || opts.ignorePrimitives.string_ {
		flags |= checker.TypeFlagsStringLike
	}
	return flags
}

// isTypeEligibleForPreferNullish mirrors upstream's
// isTypeEligibleForPreferNullish: nullable AND (no ignored-primitive overlap).
func isTypeEligibleForPreferNullish(t *checker.Type, opts *Options) bool {
	if !utils.IsNullableType(t) {
		return false
	}
	ignorable := opts.typeFlagsToIgnore()
	if ignorable == 0 {
		return true
	}
	// any/unknown — could be anything, including the ignorable primitives.
	if utils.IsTypeFlagSet(t, checker.TypeFlagsAny|checker.TypeFlagsUnknown) {
		return false
	}
	for _, c := range utils.UnionTypeParts(t) {
		for _, ic := range utils.IntersectionTypeParts(c) {
			if utils.IsTypeFlagSet(ic, ignorable) {
				return false
			}
		}
	}
	return true
}

// isTruthinessCheckEligible mirrors upstream — note the carve-out: when the
// node is a ConditionalExpression that is the direct argument of a
// CallExpression, the ignoreBooleanCoercion early-return does NOT fire.
func isTruthinessCheckEligible(ctx rule.RuleContext, opts *Options, node *ast.Node, testNode *ast.Node) bool {
	t := ctx.TypeChecker.GetTypeAtLocation(testNode)
	if !isTypeEligibleForPreferNullish(t, opts) {
		return false
	}
	if opts.ignoreConditionalTests && isConditionalTest(node) {
		return false
	}
	if opts.ignoreBooleanCoercion && isBooleanConstructorContext(node) {
		// Carve-out: a ConditionalExpression whose direct parent is a
		// CallExpression should still report.
		//
		// In ESTree, ParenthesizedExpression is transparent — `node.parent` of
		// `(a ? b : c)` inside `Boolean(...)` is the CallExpression directly.
		// tsgo keeps an explicit ParenthesizedExpression node, so walk through
		// any paren layers when matching upstream's `node.parent.type ===
		// CallExpression` check.
		p := node.Parent
		for p != nil && p.Kind == ast.KindParenthesizedExpression {
			p = p.Parent
		}
		if node.Kind != ast.KindConditionalExpression ||
			p == nil ||
			p.Kind != ast.KindCallExpression {
			return false
		}
	}
	return true
}

// reportPreferNullishOverOr handles `x || y` and `x ||= y`. They share one
// BinaryExpression listener in tsgo.
func reportPreferNullishOverOr(ctx rule.RuleContext, opts *Options, node *ast.Node, description, equals string) {
	bin := node.AsBinaryExpression()
	if !isTruthinessCheckEligible(ctx, opts, node, bin.Left) {
		return
	}
	if opts.ignoreMixedLogicalExpressions && isMixedLogicalExpression(node) {
		return
	}

	// Range covering only the operator token (`||` or `||=`), trivia stripped.
	src := ctx.SourceFile.Text()
	opPos := scanner.SkipTrivia(src, bin.OperatorToken.Pos())
	opEnd := bin.OperatorToken.End()
	opRange := core.NewTextRange(opPos, opEnd)
	rawOp := src[opPos:opEnd]
	newOp := strings.Replace(rawOp, "||", "??", 1)

	fixes := []rule.RuleFix{rule.RuleFixReplaceRange(opRange, newOp)}
	addOrParenFixes(ctx.SourceFile, node, &fixes)

	ctx.ReportRangeWithSuggestions(opRange, buildPreferNullishOverOrMessage(description, equals), rule.RuleSuggestion{
		Message:  buildSuggestNullishMessage(equals),
		FixesArr: fixes,
	})
}

// addOrParenFixes mirrors upstream's paren-adding logic when our parent is `||`
// / `||=`. Mixing `&&` and `??` requires parens.
//
// In ESTree, ParenthesizedExpression is transparent — `node.parent` of an
// expression inside `(...)` is the enclosing expression, not the paren. tsgo
// keeps explicit ParenthesizedExpression nodes; walk through them so the
// paren-add decision matches upstream byte-for-byte.
func addOrParenFixes(sf *ast.SourceFile, node *ast.Node, fixes *[]rule.RuleFix) {
	parent := node.Parent
	for parent != nil && parent.Kind == ast.KindParenthesizedExpression {
		parent = parent.Parent
	}
	if parent == nil || parent.Kind != ast.KindBinaryExpression {
		return
	}
	pop := parent.AsBinaryExpression().OperatorToken.Kind
	if pop != ast.KindBarBarToken && pop != ast.KindBarBarEqualsToken {
		return
	}
	bin := node.AsBinaryExpression()
	leftSkipped := ast.SkipParentheses(bin.Left)
	if leftSkipped != nil && leftSkipped.Kind == ast.KindBinaryExpression {
		innerLeftBin := leftSkipped.AsBinaryExpression()
		innerLeftLeftSkipped := ast.SkipParentheses(innerLeftBin.Left)
		innerLeftLeftIsLogicalOr := innerLeftLeftSkipped != nil &&
			innerLeftLeftSkipped.Kind == ast.KindBinaryExpression &&
			(innerLeftLeftSkipped.AsBinaryExpression().OperatorToken.Kind == ast.KindBarBarToken ||
				innerLeftLeftSkipped.AsBinaryExpression().OperatorToken.Kind == ast.KindBarBarEqualsToken)
		if !innerLeftLeftIsLogicalOr {
			*fixes = append(*fixes,
				rule.RuleFixInsertBefore(sf, innerLeftBin.Right, "("),
				rule.RuleFixInsertAfter(bin.Right, ")"),
			)
			return
		}
	}
	*fixes = append(*fixes,
		rule.RuleFixInsertBefore(sf, bin.Left, "("),
		rule.RuleFixInsertAfter(bin.Right, ")"),
	)
}

// reportPreferNullishOverTernary reports on a ConditionalExpression whose test
// can be replaced by ?? on the corresponding branches.
func reportPreferNullishOverTernary(ctx rule.RuleContext, opts *Options, node *ast.Node) {
	if opts.ignoreTernaryTests {
		return
	}
	ce := node.AsConditionalExpression()
	op, nodesInsideTest, ok := getOperatorAndNodesInsideTestExpression(ce.Condition)
	if !ok {
		return
	}
	nonNullishBranch, nullishBranch := getBranchNodes(ce, op)
	leftNode, fixable := getNullishCoalescingParams(ctx, opts, node, nonNullishBranch, nodesInsideTest, op)
	if !fixable {
		return
	}

	leftText := getTextWithParentheses(ctx.SourceFile, leftNode)
	rightText := getRightTextForTernary(ctx.SourceFile, nullishBranch)
	replacement := leftText + " ?? " + rightText
	ctx.ReportNodeWithSuggestions(node, buildPreferNullishOverTernaryMessage(""), rule.RuleSuggestion{
		Message:  buildSuggestNullishMessage(""),
		FixesArr: []rule.RuleFix{rule.RuleFixReplace(ctx.SourceFile, node, replacement)},
	})
}

// getTextWithParentheses mirrors upstream's `getTextWithParentheses` —
// returns the source text including one immediate layer of parens iff the
// node is parenthesized in source.
//
// In tsgo, parens are explicit `ParenthesizedExpression` nodes. We walk down
// from the outermost wrapper to the innermost paren layer and emit its text.
// That is: `(((foo).a))` becomes `((foo).a)` — exactly one paren preserved.
func getTextWithParentheses(sf *ast.SourceFile, node *ast.Node) string {
	current := node
	var lastParen *ast.Node
	for current != nil && current.Kind == ast.KindParenthesizedExpression {
		lastParen = current
		current = current.AsParenthesizedExpression().Expression
	}
	if lastParen != nil {
		return utils.TrimmedNodeText(sf, lastParen)
	}
	return utils.TrimmedNodeText(sf, node)
}

// getRightTextForTernary mirrors upstream: if the nullish branch is already
// parenthesized in source, keep it as-is; otherwise wrap with parens iff its
// precedence is below `??`.
func getRightTextForTernary(sf *ast.SourceFile, node *ast.Node) string {
	// If the SOURCE-LEVEL form of `node` is wrapped in parens, keep the parens
	// in the output by reading the outer node's text.
	if node.Kind == ast.KindParenthesizedExpression {
		return utils.TrimmedNodeText(sf, node)
	}
	text := utils.TrimmedNodeText(sf, node)
	if isLowPrecedenceForCoalesce(node) {
		return "(" + text + ")"
	}
	return text
}

// isLowPrecedenceForCoalesce reports whether `node`'s precedence is below the
// coalesce operator (`??`), so it must be parenthesized when placed on the
// right-hand side of a `??` substitution.
func isLowPrecedenceForCoalesce(node *ast.Node) bool {
	switch node.Kind {
	case ast.KindBinaryExpression:
		op := node.AsBinaryExpression().OperatorToken.Kind
		switch op {
		case ast.KindAmpersandAmpersandToken, ast.KindBarBarToken, ast.KindCommaToken:
			return true
		}
		if ast.IsAssignmentOperator(op) {
			return true
		}
	case ast.KindConditionalExpression, ast.KindYieldExpression, ast.KindArrowFunction:
		return true
	}
	return false
}

func reportPreferNullishOverIf(ctx rule.RuleContext, opts *Options, node *ast.Node) {
	if opts.ignoreIfStatements {
		return
	}
	is := node.AsIfStatement()
	if is.ElseStatement != nil {
		return
	}
	var assignmentExpr *ast.Node
	isConsequentBlock := is.ThenStatement.Kind == ast.KindBlock
	if isConsequentBlock {
		stmts := is.ThenStatement.AsBlock().Statements.Nodes
		if len(stmts) != 1 || stmts[0].Kind != ast.KindExpressionStatement {
			return
		}
		assignmentExpr = stmts[0].AsExpressionStatement().Expression
	} else if is.ThenStatement.Kind == ast.KindExpressionStatement {
		assignmentExpr = is.ThenStatement.AsExpressionStatement().Expression
	} else {
		return
	}

	// tsgo preserves outer parens around the assignment as
	// ParenthesizedExpression nodes; ESTree flattens them.
	assignmentExprStripped := ast.SkipParentheses(assignmentExpr)
	if assignmentExprStripped == nil || assignmentExprStripped.Kind != ast.KindBinaryExpression {
		return
	}
	bin := assignmentExprStripped.AsBinaryExpression()
	// Accept any assignment-shaped operator (`=`, `||=`, `??=`, `+=`, …) — in
	// ESTree this is one AssignmentExpression node.
	if !ast.IsAssignmentOperator(bin.OperatorToken.Kind) {
		return
	}
	if !isMemberAccessLike(ast.SkipParentheses(bin.Left)) {
		return
	}
	leftAssignment := bin.Left
	rightAssignment := bin.Right

	op, nodesInsideTest, ok := getOperatorAndNodesInsideTestExpression(is.Expression)
	if !ok {
		return
	}
	if op != opNotTruth && op != opEqEq && op != opEqEqEq {
		return
	}
	_, fixable := getNullishCoalescingParams(ctx, opts, node, leftAssignment, nodesInsideTest, op)
	if !fixable {
		return
	}

	src := ctx.SourceFile.Text()
	leftText := getTextWithParentheses(ctx.SourceFile, leftAssignment)
	rightText := getTextWithParentheses(ctx.SourceFile, rightAssignment)

	// Comment handling: mirror upstream's commentsBefore / commentsAfter
	// via forward-walking the trivia chunks around the assignment.
	//
	//   block:    `if (cond) { <leading-trivia> stmt; <trailing-trivia> }`
	//   single:   `if (cond) <leading-trivia> stmt;`
	commentsBefore := ""
	commentsAfter := ""
	if isConsequentBlock {
		exprStmt := is.ThenStatement.AsBlock().Statements.Nodes[0]
		leadingStart := exprStmt.Pos()
		leadingEnd := scanner.SkipTrivia(src, leadingStart)
		if leadingStart < leadingEnd {
			commentsBefore = formatComments(src[leadingStart:leadingEnd], "\n")
		}
		// Trailing trivia: between the end of the (only) statement and the
		// closing `}` of the block.
		trailingStart := exprStmt.End()
		trailingEnd := is.ThenStatement.End() - 1 // points at `}`
		if trailingStart < trailingEnd {
			commentsAfter = formatComments(src[trailingStart:trailingEnd], "\n")
		}
	} else {
		// `if (cond) /* c */ foo = makeFoo();` form — preserve comments
		// that sit in the trivia before the expression statement.
		leadingStart := is.ThenStatement.Pos()
		leadingEnd := scanner.SkipTrivia(src, leadingStart)
		if leadingStart < leadingEnd {
			commentsBefore = formatComments(src[leadingStart:leadingEnd], " ")
		}
	}

	replacement := leftText + " ??= " + rightText + ";"
	fixes := []rule.RuleFix{}
	if commentsBefore != "" {
		fixes = append(fixes, rule.RuleFixInsertBefore(ctx.SourceFile, node, commentsBefore))
	}
	fixes = append(fixes, rule.RuleFixReplace(ctx.SourceFile, node, replacement))
	if commentsAfter != "" {
		// Strip the trailing separator added by formatComments and emit with
		// a leading space — matches upstream's `slice(0, -1)` + space.
		trimmed := strings.TrimRight(commentsAfter, "\n ")
		fixes = append(fixes, rule.RuleFixInsertAfter(node, " "+trimmed))
	}

	ctx.ReportNodeWithSuggestions(node, buildPreferNullishOverAssignmentMessage("="), rule.RuleSuggestion{
		Message:  buildSuggestNullishMessage("="),
		FixesArr: fixes,
	})
}

// formatComments extracts comments out of a trivia chunk and returns them with
// `separator` between each.
func formatComments(triviaChunk string, separator string) string {
	var out strings.Builder
	i := 0
	for i < len(triviaChunk) {
		c := triviaChunk[i]
		if c == ' ' || c == '\t' || c == '\n' || c == '\r' {
			i++
			continue
		}
		if c == '/' && i+1 < len(triviaChunk) {
			if triviaChunk[i+1] == '/' {
				j := i + 2
				for j < len(triviaChunk) && triviaChunk[j] != '\n' && triviaChunk[j] != '\r' {
					j++
				}
				out.WriteString(triviaChunk[i:j])
				out.WriteString(separator)
				i = j
				continue
			}
			if triviaChunk[i+1] == '*' {
				j := i + 2
				for j+1 < len(triviaChunk) && (triviaChunk[j] != '*' || triviaChunk[j+1] != '/') {
					j++
				}
				if j+1 < len(triviaChunk) {
					out.WriteString(triviaChunk[i : j+2])
					out.WriteString(separator)
					i = j + 2
					continue
				}
				return out.String()
			}
		}
		// Anything else — give up; this isn't pure trivia.
		return out.String()
	}
	return out.String()
}

// getNullishCoalescingParams matches typescript-eslint's getNullishCoalescingParams
// for ConditionalExpression / IfStatement: figures out whether the test is
// equivalent to a nullish check, and returns the LHS for the suggested ??.
func getNullishCoalescingParams(
	ctx rule.RuleContext,
	opts *Options,
	node *ast.Node,
	nonNullishNode *ast.Node,
	nodesInsideTestExpression []*ast.Node,
	op nullishCheckOperator,
) (*ast.Node, bool) {
	var nullishCoalescingLeftNode *ast.Node
	hasTruthinessCheck := false
	hasNullCheck := false
	hasUndefinedCheck := false

	if len(nodesInsideTestExpression) == 0 {
		hasTruthinessCheck = true
		var test *ast.Node
		switch node.Kind {
		case ast.KindConditionalExpression:
			test = node.AsConditionalExpression().Condition
		case ast.KindIfStatement:
			test = node.AsIfStatement().Expression
		}
		if test == nil {
			return nil, false
		}
		test = ast.SkipParentheses(test)
		if test.Kind == ast.KindPrefixUnaryExpression {
			nullishCoalescingLeftNode = ast.SkipParentheses(test.AsPrefixUnaryExpression().Operand)
		} else {
			nullishCoalescingLeftNode = test
		}
		if !areNodesSimilarMemberAccess(ctx.SourceFile, nullishCoalescingLeftNode, nonNullishNode) {
			return nil, false
		}
	} else {
		for _, tn := range nodesInsideTestExpression {
			tnSkipped := ast.SkipParentheses(tn)
			if utils.IsNullLiteral(tnSkipped) {
				hasNullCheck = true
			} else if utils.IsUndefinedIdentifier(tnSkipped) {
				hasUndefinedCheck = true
			} else if areNodesSimilarMemberAccess(ctx.SourceFile, tn, nonNullishNode) {
				if nullishCoalescingLeftNode == nil {
					nullishCoalescingLeftNode = tn
				}
			} else {
				return nil, false
			}
		}
	}

	if nullishCoalescingLeftNode == nil {
		return nil, false
	}

	if hasTruthinessCheck {
		if !isTruthinessCheckEligible(ctx, opts, node, nullishCoalescingLeftNode) {
			return nil, false
		}
		return nullishCoalescingLeftNode, true
	}

	// Non-truthiness path: equivalence between null/undefined branches.
	if hasUndefinedCheck == hasNullCheck {
		if hasUndefinedCheck {
			return nullishCoalescingLeftNode, true
		}
		return nil, false
	}
	if op == opEqEq || op == opNotEq {
		return nullishCoalescingLeftNode, true
	}

	t := ctx.TypeChecker.GetTypeAtLocation(nullishCoalescingLeftNode)
	flags := getTypeFlags(t)
	if flags&(checker.TypeFlagsAny|checker.TypeFlagsUnknown) != 0 {
		return nil, false
	}
	hasNullType := flags&checker.TypeFlagsNull != 0
	if hasUndefinedCheck && !hasNullType {
		return nullishCoalescingLeftNode, true
	}
	hasUndefinedType := flags&checker.TypeFlagsUndefined != 0
	if hasNullCheck && !hasUndefinedType {
		return nullishCoalescingLeftNode, true
	}
	return nil, false
}

func getTypeFlags(t *checker.Type) checker.TypeFlags {
	if t == nil {
		return 0
	}
	var flags checker.TypeFlags
	for _, p := range utils.UnionTypeParts(t) {
		for _, ip := range utils.IntersectionTypeParts(p) {
			flags |= checker.Type_flags(ip)
		}
	}
	return flags
}

var PreferNullishCoalescingRule = rule.CreateRule(rule.Rule{
	Name:             "prefer-nullish-coalescing",
	RequiresTypeInfo: true,
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := parseOptions(options)
		compilerOptions := ctx.Program.Options()
		isStrictNullChecks := utils.IsStrictCompilerOptionEnabled(compilerOptions, compilerOptions.StrictNullChecks)
		if !isStrictNullChecks && !opts.allowRuleToRunWithoutStrictNullChecksIKnowWhatIAmDoing {
			ctx.ReportRange(core.NewTextRange(0, 0), buildNoStrictNullCheckMessage())
		}

		// Use exit listeners (post-order / leaf-first) so the diagnostic
		// emission order matches ESLint's `LogicalExpression` listener
		// semantics. ESLint reports inner `||` before outer `||` for chains
		// like `(a && b) || c || d`; pre-order would reverse that.
		return rule.RuleListeners{
			rule.ListenerOnExit(ast.KindBinaryExpression): func(node *ast.Node) {
				bin := node.AsBinaryExpression()
				op := bin.OperatorToken.Kind
				if !isLogicalOrLikeOperator(op) {
					return
				}
				if op == ast.KindBarBarToken {
					reportPreferNullishOverOr(ctx, &opts, node, "or", "")
				} else {
					reportPreferNullishOverOr(ctx, &opts, node, "assignment", "=")
				}
			},
			rule.ListenerOnExit(ast.KindConditionalExpression): func(node *ast.Node) {
				reportPreferNullishOverTernary(ctx, &opts, node)
			},
			rule.ListenerOnExit(ast.KindIfStatement): func(node *ast.Node) {
				reportPreferNullishOverIf(ctx, &opts, node)
			},
		}
	},
})
