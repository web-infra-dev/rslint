package prefer_object_spread

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func buildUseSpreadMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "useSpreadMessage",
		Description: "Use an object spread instead of `Object.assign` eg: `{ ...foo }`.",
	}
}

func buildUseLiteralMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "useLiteralMessage",
		Description: "Use an object literal instead of `Object.assign`. eg: `{ foo: bar }`.",
	}
}

// isGlobalIdentifier reports whether node, after unwrapping parentheses and TS
// assertion wrappers, is an unshadowed reference to the global `name` — i.e.
// no local declaration shadows it and no `/* global name: off */` /
// languageOptions.globals entry un-declares it. Mirrors the "is this the
// global variable, not a local rebinding" half of ESLint's ReferenceTracker.
func isGlobalIdentifier(node *ast.Node, name string, globals map[string]bool) bool {
	node = utils.SkipAssertionsAndParens(node)
	if node == nil || !ast.IsIdentifier(node) || node.AsIdentifier().Text != name {
		return false
	}
	if utils.IsShadowed(node, name) {
		return false
	}
	if declared, ok := globals[name]; ok && !declared {
		return false
	}
	return true
}

// unwrapValue strips parentheses, TS assertions, and — for a trailing
// comma-operator sequence such as `(0, Object.assign)` — descends into the
// sequence's last operand (mirrors ESLint's
// `SequenceExpression.expressions.at(-1)` handling). Repeats so a value
// reached through nested wrappers of either kind is fully peeled. Always
// terminates: each step either strips a wrapper or descends into a strictly
// smaller child subtree.
func unwrapValue(node *ast.Node) *ast.Node {
	node = utils.SkipAssertionsAndParens(node)
	for node != nil && node.Kind == ast.KindBinaryExpression {
		bin := node.AsBinaryExpression()
		if bin == nil || bin.OperatorToken == nil || bin.OperatorToken.Kind != ast.KindCommaToken {
			break
		}
		node = utils.SkipAssertionsAndParens(bin.Right)
	}
	return node
}

// isObjectReference reports whether node, once resolved, denotes the global
// `Object` — either directly or through a single stable local alias (e.g.
// `let o = Object; o.assign(...)`). Aliasing is resolved via
// evaluator.ResolveIdentifierInitializer, which only follows `const`
// bindings, or `let`/`var` bindings that are never reassigned in the file.
func isObjectReference(node *ast.Node, ctx rule.RuleContext, evaluator *utils.StaticStringEvaluator) bool {
	if isGlobalIdentifier(node, "Object", ctx.Globals) {
		return true
	}
	if node == nil || node.Kind != ast.KindIdentifier {
		return false
	}
	initializer, ok := evaluator.ResolveIdentifierInitializer(node)
	if !ok {
		return false
	}
	return isGlobalIdentifier(unwrapValue(initializer), "Object", ctx.Globals)
}

// isAssignDestructuredFromObject reports whether node is an identifier bound
// by destructuring the `assign` property off the global `Object` in a
// `const` declaration, e.g. `const {assign} = Object` or
// `const {assign: a} = Object` — regardless of the declaration's var/let/const
// kind. Unlike isObjectReference's alias resolution, this does not check for
// later reassignment of the bound identifier (there is no cheap
// write-tracking available for a BindingElement's symbol here); this mirrors
// ESLint's own ReferenceTracker, which likewise treats every read of a traced
// binding as the traced value regardless of intervening writes.
func isAssignDestructuredFromObject(node *ast.Node, ctx rule.RuleContext) bool {
	if node == nil || node.Kind != ast.KindIdentifier || ctx.TypeChecker == nil {
		return false
	}
	symbol := utils.GetReferenceSymbol(node, ctx.TypeChecker)
	if symbol == nil || len(symbol.Declarations) != 1 {
		return false
	}
	decl := symbol.Declarations[0]
	if decl == nil || decl.Kind != ast.KindBindingElement {
		return false
	}
	be := decl.AsBindingElement()
	if be == nil || be.DotDotDotToken != nil {
		return false
	}
	propertyName := be.PropertyName
	if propertyName == nil {
		propertyName = be.Name()
	}
	if propertyName == nil || !ast.IsIdentifier(propertyName) || propertyName.AsIdentifier().Text != "assign" {
		return false
	}

	pattern := decl.Parent
	if pattern == nil || pattern.Kind != ast.KindObjectBindingPattern {
		return false
	}
	declarationNode := pattern.Parent
	if declarationNode == nil || declarationNode.Kind != ast.KindVariableDeclaration {
		return false
	}
	declarationList := declarationNode.Parent
	if declarationList == nil || declarationList.Kind != ast.KindVariableDeclarationList {
		return false
	}
	initializer := declarationNode.AsVariableDeclaration().Initializer
	return isGlobalIdentifier(unwrapValue(initializer), "Object", ctx.Globals)
}

// isObjectAssignCallee reports whether callee, once resolved, denotes the
// `assign` method of the global `Object` — reached directly
// (`Object.assign`), through `globalThis.Object.assign`, through a stable
// local alias of the `Object` receiver (`let o = Object; o.assign(...)`), a
// stable local alias of `Object.assign` itself
// (`const assign = Object.assign; assign(...)`), a `const` destructure of
// `assign` off `Object` (`const {assign} = Object; assign(...)`), or a
// trailing comma-operator sequence around any of the above
// (`(0, Object.assign)(...)`, `(0, Object).assign(...)`). Dot and
// computed-static-string member access forms are both accepted, and
// parentheses / TS assertions are transparent throughout. Mirrors ESLint's
// ReferenceTracker tracking of the global `Object.assign` API through simple
// aliasing.
func isObjectAssignCallee(callee *ast.Node, ctx rule.RuleContext, evaluator *utils.StaticStringEvaluator) bool {
	return isObjectAssignCalleeDepth(callee, ctx, evaluator, 0)
}

// maxAliasDepth bounds how many identifier-alias hops isObjectAssignCallee
// follows. Real code never chains more than one or two; the cap just keeps a
// pathological forward-reference cycle (e.g. `const a = b; const b = a;`,
// which ResolveIdentifierInitializer cannot itself detect since each hop
// targets a distinct declaration) from recursing unboundedly.
const maxAliasDepth = 8

func isObjectAssignCalleeDepth(callee *ast.Node, ctx rule.RuleContext, evaluator *utils.StaticStringEvaluator, depth int) bool {
	if depth > maxAliasDepth {
		return false
	}
	callee = unwrapValue(callee)
	if callee == nil {
		return false
	}

	if callee.Kind == ast.KindIdentifier {
		if isAssignDestructuredFromObject(callee, ctx) {
			return true
		}
		initializer, ok := evaluator.ResolveIdentifierInitializer(callee)
		if !ok {
			return false
		}
		return isObjectAssignCalleeDepth(initializer, ctx, evaluator, depth+1)
	}

	if !utils.IsSpecificMemberAccess(callee, "", "assign") {
		return false
	}
	object := unwrapValue(utils.AccessExpressionObject(callee))
	if isObjectReference(object, ctx, evaluator) {
		return true
	}
	object = utils.SkipAssertionsAndParens(object)
	if object == nil || !utils.IsSpecificMemberAccess(object, "", "Object") {
		return false
	}
	return isGlobalIdentifier(utils.AccessExpressionObject(object), "globalThis", ctx.Globals)
}

// hasArraySpread reports whether any of the call's own arguments is a spread
// element, e.g. `Object.assign({}, ...objects)`. A spread argument makes the
// number/shape of merged sources unknowable statically, so the rule stays
// silent — mirrors upstream's hasArraySpread.
func hasArraySpread(args []*ast.Node) bool {
	for _, arg := range args {
		if arg.Kind == ast.KindSpreadElement {
			return true
		}
	}
	return false
}

// isAccessorProperty reports whether node is a getter/setter member of an
// object literal.
func isAccessorProperty(node *ast.Node) bool {
	return node.Kind == ast.KindGetAccessor || node.Kind == ast.KindSetAccessor
}

// hasAccessors reports whether the object literal node has at least one
// getter/setter property.
func hasAccessors(node *ast.Node) bool {
	obj := node.AsObjectLiteralExpression()
	if obj == nil || obj.Properties == nil {
		return false
	}
	for _, prop := range obj.Properties.Nodes {
		if isAccessorProperty(prop) {
			return true
		}
	}
	return false
}

// hasArgumentsWithAccessors reports whether any of the call's arguments is an
// object literal (parentheses-transparent) with a getter/setter property.
// Spreading such a literal would invoke the accessor eagerly at merge time
// instead of lazily on property access, changing observable behavior — so
// the rule stays silent whenever this holds and there is more than one
// argument (mirrors upstream's hasArgumentsWithAccessors).
func hasArgumentsWithAccessors(args []*ast.Node) bool {
	for _, arg := range args {
		inner := ast.SkipParentheses(arg)
		if inner != nil && inner.Kind == ast.KindObjectLiteralExpression && hasAccessors(inner) {
			return true
		}
	}
	return false
}

// needsWrappingParens determines whether the fixed object literal must be
// wrapped in parentheses to remain valid at its use site (an object literal
// at the start of a statement would otherwise parse as a block). Mirrors
// upstream's needsParens, adapted for tsgo's explicit ParenthesizedExpression
// nodes and BinaryExpression-collapsed AssignmentExpression.
func needsWrappingParens(node *ast.Node) bool {
	alreadyParenthesized := node.Parent != nil && node.Parent.Kind == ast.KindParenthesizedExpression

	current := node
	for current.Parent != nil && current.Parent.Kind == ast.KindParenthesizedExpression {
		current = current.Parent
	}
	parent := current.Parent
	if parent == nil {
		return !alreadyParenthesized
	}

	switch parent.Kind {
	case ast.KindVariableDeclaration,
		ast.KindArrayLiteralExpression,
		ast.KindReturnStatement,
		ast.KindCallExpression,
		ast.KindPropertyAssignment,
		ast.KindComputedPropertyName:
		return false
	case ast.KindBinaryExpression:
		bin := parent.AsBinaryExpression()
		if bin.OperatorToken != nil && ast.IsAssignmentOperator(bin.OperatorToken.Kind) {
			if bin.Left == current {
				return !alreadyParenthesized
			}
			return false
		}
		return !alreadyParenthesized
	default:
		return !alreadyParenthesized
	}
}

// needsSpreadParens reports whether a non-object argument needs to be
// wrapped in parentheses when prefixed with `...` — AssignmentExpression,
// ArrowFunctionExpression, and ConditionalExpression all bind looser than
// spread and would otherwise change meaning. Mirrors upstream's
// argNeedsParens.
func needsSpreadParens(argNode *ast.Node) bool {
	if argNode.Kind == ast.KindParenthesizedExpression {
		return false
	}
	inner := ast.SkipParentheses(argNode)
	if inner == nil {
		return false
	}
	switch inner.Kind {
	case ast.KindConditionalExpression, ast.KindArrowFunction:
		return true
	case ast.KindBinaryExpression:
		bin := inner.AsBinaryExpression()
		return bin.OperatorToken != nil && ast.IsAssignmentOperator(bin.OperatorToken.Kind)
	default:
		return false
	}
}

// extendForwardOverSpace advances pos over a run of whitespace characters.
func extendForwardOverSpace(text string, pos int) int {
	return utils.SkipLeadingWhitespace(text, pos, len(text))
}

// extendBackwardOverSpace mirrors extendForwardOverSpace but walks backward,
// unless a single-line comment ends exactly at the whitespace boundary — in
// that case the whitespace is left untouched so the comment keeps its own
// line and doesn't swallow the following token. Mirrors upstream's
// getStartWithSpaces, which special-cases a preceding Line comment token.
func extendBackwardOverSpace(text string, comments []*ast.CommentRange, pos int) int {
	boundary := utils.SkipTrailingWhitespace(text, 0, pos)
	if boundary == pos {
		return pos
	}
	for _, c := range comments {
		if c.End() == boundary && c.Kind == ast.KindSingleLineCommentTrivia {
			return pos
		}
	}
	return boundary
}

// findArgsOpenParenPos returns the position of the `(` that opens the call's
// argument list, skipping past any type arguments (`Object.assign<T, U>(`).
func findArgsOpenParenPos(sf *ast.SourceFile, call *ast.CallExpression) int {
	text := sf.Text()
	pos := call.Expression.End()
	if call.TypeArguments != nil {
		pos = call.TypeArguments.End()
	}
	for {
		pos = scanner.SkipTrivia(text, pos)
		if pos >= len(text) || text[pos] == '(' {
			return pos
		}
		pos++
	}
}

// parenChain returns the sequence of nodes from argNode (outermost,
// inclusive of any wrapping ParenthesizedExpression layers) down to inner
// (the real ObjectLiteralExpression), outermost first.
func parenChain(argNode *ast.Node, inner *ast.Node) []*ast.Node {
	chain := []*ast.Node{argNode}
	cur := argNode
	for cur != inner {
		cur = cur.AsParenthesizedExpression().Expression
		chain = append(chain, cur)
	}
	return chain
}

// buildObjectArgumentFixes strips the outer braces (and any redundant
// wrapping parentheses) from an ObjectExpression argument, baring its
// property list for the merged object literal, and removes a now-redundant
// argument-list comma when the stripped object was empty or already ended in
// its own trailing comma. Mirrors the ObjectExpression branch of upstream's
// defineFixer: only the outermost wrapping token absorbs adjacent
// whitespace; every nested paren/brace layer is removed as a bare single
// character, so whitespace strictly between two nested layers (or between
// the innermost paren and the object's own brace) is left untouched.
func buildObjectArgumentFixes(sf *ast.SourceFile, comments []*ast.CommentRange, argNode *ast.Node, inner *ast.Node) []rule.RuleFix {
	text := sf.Text()
	chain := parenChain(argNode, inner)
	outerRange := utils.TrimNodeTextRange(sf, chain[0])

	leftBoundary := extendForwardOverSpace(text, outerRange.Pos()+1)
	rightBoundary := extendBackwardOverSpace(text, comments, outerRange.End()-1)
	if len(chain) == 1 && rightBoundary < leftBoundary {
		rightBoundary = leftBoundary
	}

	fixes := []rule.RuleFix{
		rule.RuleFixRemoveRange(core.NewTextRange(outerRange.Pos(), leftBoundary)),
		rule.RuleFixRemoveRange(core.NewTextRange(rightBoundary, outerRange.End())),
	}
	for i := 1; i < len(chain); i++ {
		layerRange := utils.TrimNodeTextRange(sf, chain[i])
		fixes = append(fixes,
			rule.RuleFixRemoveRange(core.NewTextRange(layerRange.Pos(), layerRange.Pos()+1)),
			rule.RuleFixRemoveRange(core.NewTextRange(layerRange.End()-1, layerRange.End())),
		)
	}

	obj := inner.AsObjectLiteralExpression()
	var props []*ast.Node
	if obj.Properties != nil {
		props = obj.Properties.Nodes
	}
	hasTrailingComma := len(props) == 0
	if !hasTrailingComma {
		afterLastProp := scanner.SkipTrivia(text, props[len(props)-1].End())
		hasTrailingComma = afterLastProp < len(text) && text[afterLastProp] == ','
	}
	if hasTrailingComma {
		afterArg := scanner.SkipTrivia(text, outerRange.End())
		if afterArg < len(text) && text[afterArg] == ',' {
			fixes = append(fixes, rule.RuleFixRemoveRange(core.NewTextRange(afterArg, afterArg+1)))
		}
	}

	return fixes
}

// buildSpreadArgumentFixes prefixes a non-object argument with `...`,
// wrapping it in parentheses first when needed. Mirrors the else branch of
// upstream's defineFixer.
func buildSpreadArgumentFixes(sf *ast.SourceFile, argNode *ast.Node) []rule.RuleFix {
	outerRange := utils.TrimNodeTextRange(sf, argNode)
	if needsSpreadParens(argNode) {
		return []rule.RuleFix{
			rule.RuleFixReplaceRange(core.NewTextRange(outerRange.Pos(), outerRange.Pos()), "...("),
			rule.RuleFixReplaceRange(core.NewTextRange(outerRange.End(), outerRange.End()), ")"),
		}
	}
	return []rule.RuleFix{
		rule.RuleFixReplaceRange(core.NewTextRange(outerRange.Pos(), outerRange.Pos()), "..."),
	}
}

// buildFixes autofixes the Object.assign call to an object spread, mirroring
// upstream's defineFixer: the callee (and any type arguments) is removed, the
// argument list's parentheses become braces (wrapped in parens of their own
// when required by the use site, with a leading `;` when needed for ASI
// safety), and each argument is rewritten in place — object-literal
// arguments are unwrapped to a bare property list, everything else is
// prefixed with `...`.
func buildFixes(ctx rule.RuleContext, node *ast.Node, args []*ast.Node) []rule.RuleFix {
	sf := ctx.SourceFile
	call := node.AsCallExpression()
	comments := ctx.Comments.All()

	nodeRange := utils.TrimNodeTextRange(sf, node)
	leftParenPos := findArgsOpenParenPos(sf, call)
	rightParenPos := nodeRange.End() - 1

	wrap := needsWrappingParens(node)
	leftReplacement := "{"
	rightReplacement := "}"
	if wrap {
		leftReplacement = "("
		if utils.IsStartOfExpressionStatement(sf, node) && utils.NeedsPrecedingSemicolon(sf, node) {
			leftReplacement = ";("
		}
		leftReplacement += "{"
		rightReplacement = "})"
	}

	fixes := []rule.RuleFix{
		rule.RuleFixRemoveRange(core.NewTextRange(nodeRange.Pos(), leftParenPos)),
		rule.RuleFixReplaceRange(core.NewTextRange(leftParenPos, leftParenPos+1), leftReplacement),
		rule.RuleFixReplaceRange(core.NewTextRange(rightParenPos, rightParenPos+1), rightReplacement),
	}

	for _, argNode := range args {
		inner := ast.SkipParentheses(argNode)
		if inner != nil && inner.Kind == ast.KindObjectLiteralExpression {
			fixes = append(fixes, buildObjectArgumentFixes(sf, comments, argNode, inner)...)
		} else {
			fixes = append(fixes, buildSpreadArgumentFixes(sf, argNode)...)
		}
	}

	return fixes
}

// https://eslint.org/docs/latest/rules/prefer-object-spread
var PreferObjectSpreadRule = rule.Rule{
	Name: "prefer-object-spread",
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		evaluator := utils.NewStaticStringEvaluatorWithSourceFile(ctx.TypeChecker, ctx.SourceFile)
		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				call := node.AsCallExpression()
				if call.Arguments == nil || len(call.Arguments.Nodes) < 1 {
					return
				}
				if !isObjectAssignCallee(call.Expression, ctx, evaluator) {
					return
				}

				args := call.Arguments.Nodes
				firstArg := ast.SkipParentheses(args[0])
				if firstArg == nil || firstArg.Kind != ast.KindObjectLiteralExpression {
					return
				}
				if hasArraySpread(args) {
					return
				}
				if len(args) > 1 && hasArgumentsWithAccessors(args) {
					return
				}

				msg := buildUseSpreadMessage()
				if len(args) == 1 {
					msg = buildUseLiteralMessage()
				}

				ctx.ReportNodeWithFixes(node, msg, buildFixes(ctx, node, args)...)
			},
		}
	},
}
