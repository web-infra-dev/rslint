package prefer_nullish_coalescing

import (
	"cmp"
	_ "embed"
	"fmt"
	"slices"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	"github.com/web-infra-dev/rslint/internal/utils/ecmascript"
)

//go:embed prefer_nullish_coalescing.schema.json
var schemaJSON []byte

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

func parseOptions(options []any) Options {
	opts := defaultOptions()
	if len(options) == 0 {
		return opts
	}
	optsMap, _ := options[0].(map[string]interface{})
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

// nullishCheckOperator: '!', '!=', '!==', ”, '==', '==='
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
	return isBuiltInBooleanCall(parent)
}

// isBuiltInBooleanCall reports whether `callNode` is a call to the global
// Boolean(...) with at least one argument. Like upstream, the expression may be
// any of its arguments.
func isBuiltInBooleanCall(callNode *ast.Node) bool {
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
	if call.Arguments == nil || len(call.Arguments.Nodes) == 0 {
		return false
	}
	return !utils.IsShadowed(callNode, "Boolean")
}

// logicalExpression returns the BinaryExpression when `node` is what ESTree
// models as a `LogicalExpression` (`||`, `&&`, `??`).
func logicalExpression(node *ast.Node) *ast.BinaryExpression {
	if node == nil || node.Kind != ast.KindBinaryExpression {
		return nil
	}
	bin := node.AsBinaryExpression()
	if !isLogicalLikeOperator(bin.OperatorToken.Kind) {
		return nil
	}
	return bin
}

// isLogicalOrExpression mirrors upstream's isLogicalOrOperator: a
// `LogicalExpression` whose operator is `||`. `||=` is an
// AssignmentExpression in ESTree and does not match.
func isLogicalOrExpression(node *ast.Node) bool {
	bin := logicalExpression(node)
	return bin != nil && bin.OperatorToken.Kind == ast.KindBarBarToken
}

// isMixedLogicalExpression mirrors upstream's BFS over parent + left + right
// looking for an `&&` operator. Parentheses are transparent, and only
// `LogicalExpression` nodes are expanded; `||=` is not one.
func isMixedLogicalExpression(node *ast.Node) bool {
	seen := map[*ast.Node]bool{}
	bin := node.AsBinaryExpression()
	queue := []*ast.Node{
		utils.ESTreeParent(node),
		utils.ESTreeRuntimeExpression(bin.Left),
		utils.ESTreeRuntimeExpression(bin.Right),
	}
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		if current == nil || seen[current] {
			continue
		}
		seen[current] = true
		cb := logicalExpression(current)
		if cb == nil {
			continue
		}
		switch cb.OperatorToken.Kind {
		case ast.KindAmpersandAmpersandToken:
			return true
		case ast.KindBarBarToken:
			queue = append(queue,
				utils.ESTreeParent(current),
				utils.ESTreeRuntimeExpression(cb.Left),
				utils.ESTreeRuntimeExpression(cb.Right),
			)
		}
	}
	return false
}

// areNodesSimilarMemberAccess mirrors typescript-eslint's
// areNodesSimilarMemberAccess: two member-access sequences are similar when
// they access the same members in the same order, regardless of optional
// chaining. Parentheses are transparent as in ESTree; tsgo has no
// ChainExpression wrapper to skip. Other wrappers such as non-null
// assertions are compared by isNodeEqual, which rejects them.
func areNodesSimilarMemberAccess(a, b *ast.Node) bool {
	a = ast.SkipParentheses(a)
	b = ast.SkipParentheses(b)
	if a == nil || b == nil {
		return false
	}
	if isMemberExpressionLike(a) && isMemberExpressionLike(b) {
		if !areNodesSimilarMemberAccess(utils.AccessExpressionObject(a), utils.AccessExpressionObject(b)) {
			return false
		}
		aProp, aComputed := memberProperty(a)
		bProp, bComputed := memberProperty(b)
		if aComputed == bComputed {
			return isNodeEqual(aProp, bProp)
		}
		if aProp.Kind == ast.KindStringLiteral && bProp.Kind == ast.KindIdentifier {
			return aProp.AsStringLiteral().Text == bProp.AsIdentifier().Text
		}
		if aProp.Kind == ast.KindIdentifier && bProp.Kind == ast.KindStringLiteral {
			return aProp.AsIdentifier().Text == bProp.AsStringLiteral().Text
		}
		return false
	}
	return isNodeEqual(a, b)
}

// isMemberExpressionLike reports whether `node` is a MemberExpression-shaped
// access (PropertyAccessExpression / ElementAccessExpression in tsgo terms).
func isMemberExpressionLike(node *ast.Node) bool {
	return node.Kind == ast.KindPropertyAccessExpression ||
		node.Kind == ast.KindElementAccessExpression
}

// memberProperty returns the ESTree `property` node and whether the access is
// computed.
func memberProperty(node *ast.Node) (*ast.Node, bool) {
	switch node.Kind {
	case ast.KindPropertyAccessExpression:
		return node.AsPropertyAccessExpression().Name(), false
	case ast.KindElementAccessExpression:
		return ast.SkipParentheses(node.AsElementAccessExpression().ArgumentExpression), true
	}
	return nil, false
}

// isNodeEqual mirrors typescript-eslint's isNodeEqual, which only compares
// ThisExpression, Literal, Identifier and MemberExpression nodes. Template
// literals are not ESTree `Literal`s, and member expressions compare their
// object and property without regard to `computed`. Every other kind is
// unequal.
func isNodeEqual(a, b *ast.Node) bool {
	a = ast.SkipParentheses(a)
	b = ast.SkipParentheses(b)
	if a == nil || b == nil {
		return false
	}
	if isMemberExpressionLike(a) && isMemberExpressionLike(b) {
		aProp, _ := memberProperty(a)
		bProp, _ := memberProperty(b)
		return isNodeEqual(aProp, bProp) &&
			isNodeEqual(utils.AccessExpressionObject(a), utils.AccessExpressionObject(b))
	}
	if a.Kind != b.Kind {
		return false
	}
	switch a.Kind {
	case ast.KindThisKeyword, ast.KindTrueKeyword, ast.KindFalseKeyword, ast.KindNullKeyword:
		return true
	case ast.KindIdentifier:
		return a.AsIdentifier().Text == b.AsIdentifier().Text
	case ast.KindStringLiteral:
		return a.AsStringLiteral().Text == b.AsStringLiteral().Text
	case ast.KindNumericLiteral:
		av, aok := ecmascript.StringToNumber(a.AsNumericLiteral().Text)
		bv, bok := ecmascript.StringToNumber(b.AsNumericLiteral().Text)
		return aok && bok && av == bv
	case ast.KindBigIntLiteral:
		return a.AsBigIntLiteral().Text == b.AsBigIntLiteral().Text
	}
	return false
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
	t := ctx.TypeChecker.GetTypeAtLocation(ast.SkipParentheses(testNode))
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

// pendingReports buffers the rule's reports so they are emitted in source
// order, as ESLint orders them. No single traversal order does this: a `||`
// report covers only its operator, so a left-nested chain such as
// `(a || b) || c` reports the inner expression first while a right-nested one
// such as `a ||= b || c` reports the outer one first.
type pendingReports struct {
	reports []pendingReport
}

type pendingReport struct {
	pos  int
	emit func()
}

func (p *pendingReports) add(pos int, emit func()) {
	p.reports = append(p.reports, pendingReport{pos: pos, emit: emit})
}

func (p *pendingReports) flush() {
	slices.SortStableFunc(p.reports, func(a, b pendingReport) int {
		return cmp.Compare(a.pos, b.pos)
	})
	for _, report := range p.reports {
		report.emit()
	}
	p.reports = nil
}

// reportPreferNullishOverOr handles `x || y` and `x ||= y`. They share one
// BinaryExpression listener in tsgo.
func reportPreferNullishOverOr(ctx rule.RuleContext, opts *Options, pending *pendingReports, node *ast.Node, description, equals string) {
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
	pending.add(opPos, func() {
		ctx.ReportRangeWithDeferredSuggestions(opRange, buildPreferNullishOverOrMessage(description, equals), func() []rule.RuleSuggestion {
			newOp := strings.Replace(src[opPos:opEnd], "||", "??", 1)
			fixes := []rule.RuleFix{rule.RuleFixReplaceRange(opRange, newOp)}
			addOrParenFixes(ctx.SourceFile, node, &fixes)
			return []rule.RuleSuggestion{{
				Message:  buildSuggestNullishMessage(equals),
				FixesArr: fixes,
			}}
		})
	})
}

// addOrParenFixes mirrors upstream's paren-adding logic when the ESTree parent
// is a `||` LogicalExpression: `&&` and `??` cannot be mixed without
// parentheses. Parentheses are transparent as in ESTree, so `node.left`,
// `node.right` and `node.parent` refer to the nodes inside them; `||=` is an
// AssignmentExpression and is not a `||` parent.
func addOrParenFixes(sf *ast.SourceFile, node *ast.Node, fixes *[]rule.RuleFix) {
	if !isLogicalOrExpression(utils.ESTreeParent(node)) {
		return
	}
	bin := node.AsBinaryExpression()
	left := utils.ESTreeRuntimeExpression(bin.Left)
	right := utils.ESTreeRuntimeExpression(bin.Right)
	if leftLogical := logicalExpression(left); leftLogical != nil &&
		!isLogicalOrExpression(utils.ESTreeRuntimeExpression(leftLogical.Left)) {
		*fixes = append(*fixes, rule.RuleFixInsertBefore(sf, utils.ESTreeRuntimeExpression(leftLogical.Right), "("))
	} else {
		*fixes = append(*fixes, rule.RuleFixInsertBefore(sf, left, "("))
	}
	*fixes = append(*fixes, rule.RuleFixInsertAfter(right, ")"))
}

// reportPreferNullishOverTernary reports on a ConditionalExpression whose test
// can be replaced by ?? on the corresponding branches.
func reportPreferNullishOverTernary(ctx rule.RuleContext, opts *Options, pending *pendingReports, node *ast.Node) {
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

	pending.add(utils.TrimNodeTextRange(ctx.SourceFile, node).Pos(), func() {
		ctx.ReportNodeWithDeferredSuggestions(node, buildPreferNullishOverTernaryMessage(""), func() []rule.RuleSuggestion {
			replacement := getTextWithParentheses(ctx.SourceFile, leftNode) + " ?? " + getRightTextForTernary(ctx.SourceFile, nullishBranch)
			return []rule.RuleSuggestion{{
				Message:  buildSuggestNullishMessage(""),
				FixesArr: []rule.RuleFix{rule.RuleFixReplace(ctx.SourceFile, node, replacement)},
			}}
		})
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

// getRightTextForTernary mirrors upstream: a parenthesized nullish branch keeps
// one layer of its parentheses; otherwise it is wrapped when its precedence is
// not higher than `??`.
func getRightTextForTernary(sf *ast.SourceFile, node *ast.Node) string {
	if node.Kind == ast.KindParenthesizedExpression {
		return getTextWithParentheses(sf, node)
	}
	text := utils.TrimmedNodeText(sf, node)
	if needsParensAsCoalesceRight(node) {
		return "(" + text + ")"
	}
	return text
}

// needsParensAsCoalesceRight mirrors upstream's
// getWrappedCode(text, getOperatorPrecedenceForNode(node), Coalesce), which
// wraps every node whose precedence is at most Coalesce (= Conditional) or is
// unclassified (Invalid):
//   - yield, arrow functions, conditionals, sequences and `??`;
//   - assignments, including `=` and `/=`, which upstream does not classify;
//   - `==`, which upstream's table places with the assignment operators;
//   - `satisfies`, `import(...)` and instantiation expressions, which upstream
//     does not classify.
//
// Upstream leaves `||` and `&&` unwrapped, which yields a suggestion such as
// `a ?? b || c`, a syntax error. These are wrapped here instead.
func needsParensAsCoalesceRight(node *ast.Node) bool {
	switch node.Kind {
	case ast.KindYieldExpression, ast.KindArrowFunction, ast.KindConditionalExpression,
		ast.KindSatisfiesExpression, ast.KindExpressionWithTypeArguments:
		return true
	case ast.KindCallExpression:
		return node.AsCallExpression().Expression.Kind == ast.KindImportKeyword
	case ast.KindBinaryExpression:
		op := node.AsBinaryExpression().OperatorToken.Kind
		if ast.IsAssignmentOperator(op) {
			return true
		}
		switch op {
		case ast.KindCommaToken, ast.KindQuestionQuestionToken, ast.KindEqualsEqualsToken,
			ast.KindBarBarToken, ast.KindAmpersandAmpersandToken:
			return true
		}
	}
	return false
}

func reportPreferNullishOverIf(ctx rule.RuleContext, opts *Options, pending *pendingReports, node *ast.Node) {
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

	pending.add(utils.TrimNodeTextRange(ctx.SourceFile, node).Pos(), func() {
		ctx.ReportNodeWithDeferredSuggestions(node, buildPreferNullishOverAssignmentMessage("="), func() []rule.RuleSuggestion {
			// Mirrors upstream's getCommentsBefore(assignmentExpression) and, for
			// a block, getCommentsAfter(expressionStatement). The ESTree
			// assignment excludes its parentheses, so the comments before it start
			// after the token that precedes the innermost one.
			comments := ctx.Comments.All()
			separator := " "
			if isConsequentBlock {
				separator = "\n"
			}
			commentsBefore := formatComments(ctx.SourceFile, comments, assignmentExprStripped.Pos(), scanner.SkipTrivia(ctx.SourceFile.Text(), assignmentExprStripped.Pos()), separator)
			commentsAfter := ""
			if isConsequentBlock {
				exprStmt := is.ThenStatement.AsBlock().Statements.Nodes[0]
				// The closing `}` is the token after the statement.
				commentsAfter = formatComments(ctx.SourceFile, comments, exprStmt.End(), is.ThenStatement.End()-1, "\n")
			}

			replacement := getTextWithParentheses(ctx.SourceFile, leftAssignment) + " ??= " + getTextWithParentheses(ctx.SourceFile, rightAssignment) + ";"
			fixes := []rule.RuleFix{}
			if commentsBefore != "" {
				fixes = append(fixes, rule.RuleFixInsertBefore(ctx.SourceFile, node, commentsBefore))
			}
			fixes = append(fixes, rule.RuleFixReplace(ctx.SourceFile, node, replacement))
			if commentsAfter != "" {
				// Upstream drops the final separator: `commentsAfter.slice(0, -1)`.
				fixes = append(fixes, rule.RuleFixInsertAfter(node, " "+strings.TrimSuffix(commentsAfter, "\n")))
			}
			return []rule.RuleSuggestion{{
				Message:  buildSuggestNullishMessage("="),
				FixesArr: fixes,
			}}
		})
	})
}

// formatComments mirrors upstream's formatComments for the comments lying
// within [start, end): each comment's source text followed by `separator`.
func formatComments(sf *ast.SourceFile, comments []*ast.CommentRange, start, end int, separator string) string {
	var out strings.Builder
	text := sf.Text()
	for _, comment := range comments {
		if comment.Pos() < start {
			continue
		}
		if comment.End() > end {
			break
		}
		out.WriteString(text[comment.Pos():comment.End()])
		out.WriteString(separator)
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
		// Keep the parentheses around the tested expression: upstream's
		// getTextWithParentheses preserves one layer of them in the suggestion.
		if unwrapped := ast.SkipParentheses(test); unwrapped.Kind == ast.KindPrefixUnaryExpression {
			nullishCoalescingLeftNode = unwrapped.AsPrefixUnaryExpression().Operand
		} else {
			nullishCoalescingLeftNode = test
		}
		if !areNodesSimilarMemberAccess(nullishCoalescingLeftNode, nonNullishNode) {
			return nil, false
		}
	} else {
		for _, tn := range nodesInsideTestExpression {
			tnSkipped := ast.SkipParentheses(tn)
			if utils.IsNullLiteral(tnSkipped) {
				hasNullCheck = true
			} else if utils.IsUndefinedIdentifier(tnSkipped) {
				hasUndefinedCheck = true
			} else if areNodesSimilarMemberAccess(tn, nonNullishNode) {
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

	t := ctx.TypeChecker.GetTypeAtLocation(ast.SkipParentheses(nullishCoalescingLeftNode))
	flags := getTypeFlags(t)
	for _, part := range utils.UnionTypeParts(t) {
		// A generic type (`T`, `NoInfer<T>`, `T[K]`, ...) may be `null` or
		// `undefined` unless its constraint rules that out.
		if checker.Type_flags(part)&checker.TypeFlagsInstantiable == 0 {
			continue
		}
		constraint := checker.Checker_getBaseConstraintOfType(ctx.TypeChecker, part)
		if constraint == nil {
			return nil, false
		}
		flags |= getTypeFlags(constraint)
	}
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

// getTypeFlags mirrors typescript-eslint's getTypeFlags: the flags of every
// union constituent. Intersections are not decomposed.
func getTypeFlags(t *checker.Type) checker.TypeFlags {
	if t == nil {
		return 0
	}
	var flags checker.TypeFlags
	for _, p := range utils.UnionTypeParts(t) {
		flags |= checker.Type_flags(p)
	}
	return flags
}

var PreferNullishCoalescingRule = rule.CreateRule(rule.Rule{
	Name:             "prefer-nullish-coalescing",
	Schema:           rule.NewSchema(schemaJSON),
	RequiresTypeInfo: true,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		opts := parseOptions(options)
		compilerOptions := ctx.Program().Options()
		isStrictNullChecks := utils.IsStrictCompilerOptionEnabled(compilerOptions, compilerOptions.StrictNullChecks)
		if !isStrictNullChecks && !opts.allowRuleToRunWithoutStrictNullChecksIKnowWhatIAmDoing {
			ctx.ReportRange(core.NewTextRange(0, 0), buildNoStrictNullCheckMessage())
		}

		pending := &pendingReports{}
		return rule.RuleListeners{
			rule.ListenerOnExit(ast.KindBinaryExpression): func(node *ast.Node) {
				bin := node.AsBinaryExpression()
				op := bin.OperatorToken.Kind
				if !isLogicalOrLikeOperator(op) {
					return
				}
				if op == ast.KindBarBarToken {
					reportPreferNullishOverOr(ctx, &opts, pending, node, "or", "")
				} else {
					reportPreferNullishOverOr(ctx, &opts, pending, node, "assignment", "=")
				}
			},
			ast.KindConditionalExpression: func(node *ast.Node) {
				reportPreferNullishOverTernary(ctx, &opts, pending, node)
			},
			ast.KindIfStatement: func(node *ast.Node) {
				reportPreferNullishOverIf(ctx, &opts, pending, node)
			},
			rule.ListenerOnExit(ast.KindEndOfFile): func(*ast.Node) {
				pending.flush()
			},
		}
	},
})
