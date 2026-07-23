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

// trackedGlobalNames lists the globals this rule ever uses as a tracking
// entry point: `Object` itself plus the global-object aliases.
var trackedGlobalNames = map[string]bool{
	"Object":     true,
	"globalThis": true,
	"window":     true,
	"self":       true,
	"global":     true,
}

// trackedValue is the abstract-value domain the rule resolves expressions
// into: the global object, the global `Object` constructor reached through
// it, and `Object.assign` itself. Everything else is valueUnknown.
type trackedValue int

const (
	valueUnknown trackedValue = iota
	valueGlobalObject
	valueObjectCtor
	valueObjectAssign
)

// globalWrites lazily records, for each tracked global name, the position of
// the first bare write reference to it in the file (`Object = {}`,
// `window ||= foo`, ...). A reference only denotes the pristine global when
// it appears before every such write; a later write does not retroactively
// untrack earlier references, and a value captured before the write keeps
// matching. Writes to a local binding that shadows the name do not count.
type globalWrites struct {
	sourceFile *ast.SourceFile
	computed   bool
	earliest   map[string]int
}

func (w *globalWrites) writtenBefore(name string, pos int) bool {
	if !w.computed {
		w.compute()
	}
	first, ok := w.earliest[name]
	return ok && first < pos
}

func (w *globalWrites) compute() {
	w.computed = true
	w.earliest = map[string]int{}
	var visit func(node *ast.Node)
	visit = func(node *ast.Node) {
		if node == nil {
			return
		}
		if node.Kind == ast.KindIdentifier {
			name := node.AsIdentifier().Text
			if trackedGlobalNames[name] && utils.IsWriteReference(node) && !utils.IsShadowed(node, name) {
				if first, ok := w.earliest[name]; !ok || node.Pos() < first {
					w.earliest[name] = node.Pos()
				}
			}
		}
		node.ForEachChild(func(child *ast.Node) bool {
			visit(child)
			return false
		})
	}
	visit(&w.sourceFile.Node)
}

// symbolWriteKind selects which payload field of a symbolWrite is meaningful.
type symbolWriteKind int

const (
	// writeOpaque is a write whose assigned value cannot be tracked (compound
	// assignment, ++/--, destructuring assignment, for-in/of target, ...).
	writeOpaque symbolWriteKind = iota
	// writeValue is a declaration initializer or simple `x = expr`
	// assignment; value holds the assigned expression.
	writeValue
	// writeDestructure is a declaration-destructuring binding; binding holds
	// the BindingElement that introduces the variable.
	writeDestructure
)

// symbolWrite records one write to a local variable: where it happens and
// what it assigns.
type symbolWrite struct {
	pos     int
	kind    symbolWriteKind
	value   *ast.Node
	binding *ast.Node
}

// objectTracker resolves expressions to the trackedValue domain. Local
// variables resolve flow-sensitively: a reference takes the value of the last
// write (declaration initializer, destructuring declaration, or simple
// assignment) that precedes it in source order, so an alias only matches
// while it actually holds the tracked value — `let o = Object;
// o.assign({}, a); o = foo; o.assign({}, b)` reports the first call only.
// The analysis is position-based, not path-based: a write that only executes
// on some branches still counts for everything after it.
type objectTracker struct {
	ctx            rule.RuleContext
	evaluator      *utils.StaticStringEvaluator
	globals        *globalWrites
	writesComputed bool
	symbolWrites   map[*ast.Symbol][]symbolWrite
}

// localBindingSymbol resolves an identifier to the local variable symbol it
// references by walking the binder's Locals tables outward — nearest scope
// that binds the name wins, mirroring lexical lookup. The binder is used
// instead of the checker because in a script (non-module) file a top-level
// binding whose name collides with a lib global (`const {Object} =
// globalThis`) makes GetSymbolAtLocation resolve to the merged global symbol;
// walking Locals keeps write sites and read sites resolving to the same
// symbol. Returns nil for globals and for names bound only by a named
// function/class expression, which can never hold a tracked value.
func localBindingSymbol(node *ast.Node) *ast.Symbol {
	name := node.AsIdentifier().Text
	for current := node.Parent; current != nil; current = current.Parent {
		if current.Kind == ast.KindFunctionExpression || current.Kind == ast.KindClassExpression {
			if exprName := current.Name(); exprName != nil && ast.IsIdentifier(exprName) &&
				exprName.AsIdentifier().Text == name {
				return nil
			}
		}
		if !ast.IsLocalsContainer(current) {
			continue
		}
		if local := ast.GetLocals(current)[name]; local != nil {
			return local
		}
	}
	return nil
}

func (t *objectTracker) computeWrites() {
	t.writesComputed = true
	t.symbolWrites = map[*ast.Symbol][]symbolWrite{}
	var visit func(node *ast.Node)
	visit = func(node *ast.Node) {
		if node == nil {
			return
		}
		switch node.Kind {
		case ast.KindVariableDeclaration:
			declaration := node.AsVariableDeclaration()
			if name := declaration.Name(); name != nil && ast.IsIdentifier(name) && declaration.Initializer != nil {
				t.recordWrite(name, symbolWrite{pos: name.Pos(), kind: writeValue, value: declaration.Initializer})
			}
		case ast.KindBindingElement:
			if name := node.AsBindingElement().Name(); name != nil && ast.IsIdentifier(name) {
				t.recordWrite(name, symbolWrite{pos: name.Pos(), kind: writeDestructure, binding: node})
			}
		case ast.KindIdentifier:
			if utils.IsWriteReference(node) {
				write := symbolWrite{pos: node.Pos(), kind: writeOpaque}
				if parent := node.Parent; parent != nil && parent.Kind == ast.KindBinaryExpression {
					binary := parent.AsBinaryExpression()
					if binary.OperatorToken != nil && binary.OperatorToken.Kind == ast.KindEqualsToken && binary.Left == node {
						write.kind = writeValue
						write.value = binary.Right
					}
				}
				t.recordWrite(node, write)
			}
		}
		node.ForEachChild(func(child *ast.Node) bool {
			visit(child)
			return false
		})
	}
	visit(&t.ctx.SourceFile.Node)
}

func (t *objectTracker) recordWrite(name *ast.Node, write symbolWrite) {
	if symbol := localBindingSymbol(name); symbol != nil {
		t.symbolWrites[symbol] = append(t.symbolWrites[symbol], write)
	}
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

// valueOf resolves node to its abstract value. visited guards against cycles
// through mutually-referential bindings (`var a = b; var b = a;`) — every
// expression node is resolved at most once per query, so arbitrarily long
// stable alias chains resolve without a depth cap.
func (t *objectTracker) valueOf(node *ast.Node, visited map[*ast.Node]bool) trackedValue {
	node = unwrapValue(node)
	if node == nil || visited[node] {
		return valueUnknown
	}
	visited[node] = true

	switch node.Kind {
	case ast.KindIdentifier:
		name := node.AsIdentifier().Text
		if trackedGlobalNames[name] && t.isPristineGlobal(node, name) {
			if name == "Object" {
				return valueObjectCtor
			}
			return valueGlobalObject
		}
		return t.valueOfLocal(node, visited)
	case ast.KindPropertyAccessExpression, ast.KindElementAccessExpression:
		name, ok := memberAccessName(node, t.evaluator)
		if !ok {
			return valueUnknown
		}
		return memberOf(t.valueOf(utils.AccessExpressionObject(node), visited), name)
	}
	return valueUnknown
}

// isPristineGlobal reports whether the identifier reference denotes the
// untouched global `name`: no local declaration shadows it, no
// languageOptions.globals entry un-declares it, and no bare write to the
// global appears before this reference in the file.
func (t *objectTracker) isPristineGlobal(node *ast.Node, name string) bool {
	if utils.IsShadowed(node, name) {
		return false
	}
	if declared, ok := t.ctx.Globals[name]; ok && !declared {
		return false
	}
	return !t.globals.writtenBefore(name, node.Pos())
}

// valueOfLocal resolves a local variable reference to the value assigned by
// the last write that precedes it in source order. No preceding write, or a
// preceding write whose value cannot be tracked, resolves to unknown.
func (t *objectTracker) valueOfLocal(node *ast.Node, visited map[*ast.Node]bool) trackedValue {
	symbol := localBindingSymbol(node)
	if symbol == nil {
		return valueUnknown
	}
	if !t.writesComputed {
		t.computeWrites()
	}
	pos := node.Pos()
	var last *symbolWrite
	writes := t.symbolWrites[symbol]
	for i := range writes {
		if writes[i].pos < pos && (last == nil || writes[i].pos > last.pos) {
			last = &writes[i]
		}
	}
	if last == nil {
		return valueUnknown
	}
	switch last.kind {
	case writeValue:
		return t.valueOf(last.value, visited)
	case writeDestructure:
		return t.destructuredValue(last.binding, visited)
	default:
		return valueUnknown
	}
}

// destructuredValue resolves the value bound by a (possibly nested)
// declaration destructuring, e.g. `assign` in `const { Object: { assign } } =
// globalThis`: it collects the property path from the binding element up to
// the declaration's initializer and applies it to the initializer's resolved
// value. Rest elements, default values, array patterns, and property names
// that cannot be statically evaluated all resolve to unknown.
func (t *objectTracker) destructuredValue(binding *ast.Node, visited map[*ast.Node]bool) trackedValue {
	var path []string
	for current := binding; ; {
		element := current.AsBindingElement()
		if element == nil || element.DotDotDotToken != nil || element.Initializer != nil {
			return valueUnknown
		}
		propertyName := element.PropertyName
		if propertyName == nil {
			propertyName = element.Name()
		}
		name, ok := bindingPropertyName(propertyName, t.evaluator)
		if !ok {
			return valueUnknown
		}
		path = append(path, name)
		pattern := current.Parent
		if pattern == nil || pattern.Kind != ast.KindObjectBindingPattern {
			return valueUnknown
		}
		container := pattern.Parent
		if container == nil {
			return valueUnknown
		}
		if container.Kind == ast.KindBindingElement {
			current = container
			continue
		}
		if container.Kind != ast.KindVariableDeclaration {
			return valueUnknown
		}
		value := t.valueOf(container.AsVariableDeclaration().Initializer, visited)
		for i := len(path) - 1; i >= 0; i-- {
			value = memberOf(value, path[i])
		}
		return value
	}
}

// memberOf applies one property hop to an abstract value: the global object's
// `Object` entry is the Object constructor, and the constructor's `assign`
// entry is Object.assign. Every other hop leaves the tracked domain.
func memberOf(value trackedValue, name string) trackedValue {
	switch {
	case value == valueGlobalObject && name == "Object":
		return valueObjectCtor
	case value == valueObjectCtor && name == "assign":
		return valueObjectAssign
	default:
		return valueUnknown
	}
}

// memberAccessName returns the statically-known member name of a dot or
// computed member access. Computed names go through the full static string
// evaluator, so folded forms like `Object["as" + "sign"]` resolve — mirroring
// ESLint's ReferenceTracker, which resolves member names via getStaticValue.
func memberAccessName(node *ast.Node, evaluator *utils.StaticStringEvaluator) (string, bool) {
	switch node.Kind {
	case ast.KindPropertyAccessExpression:
		name := node.AsPropertyAccessExpression().Name()
		if name != nil && ast.IsIdentifier(name) {
			return name.AsIdentifier().Text, true
		}
	case ast.KindElementAccessExpression:
		return evaluator.Eval(node.AsElementAccessExpression().ArgumentExpression)
	}
	return "", false
}

// bindingPropertyName returns the property name a BindingElement destructures
// — an identifier (`{assign}`), a string literal (`{"assign": a}`), or a
// computed name whose expression statically evaluates (`{["as" + "sign"]:
// a}`). Mirrors ESLint's getPropertyName over destructuring patterns.
func bindingPropertyName(propertyName *ast.Node, evaluator *utils.StaticStringEvaluator) (string, bool) {
	if propertyName == nil {
		return "", false
	}
	switch propertyName.Kind {
	case ast.KindIdentifier:
		return propertyName.AsIdentifier().Text, true
	case ast.KindStringLiteral:
		return propertyName.AsStringLiteral().Text, true
	case ast.KindComputedPropertyName:
		return evaluator.Eval(propertyName.AsComputedPropertyName().Expression)
	}
	return "", false
}

// isObjectAssignCallee reports whether callee resolves to the global
// `Object.assign` — directly, through global-object entries
// (`globalThis.Object.assign`, with the global object itself possibly
// aliased), through aliases of `Object` or of `Object.assign` established by
// declaration initializers or simple assignments (followed flow-sensitively),
// through (nested) declaration destructuring (`const { Object: { assign } } =
// globalThis`), including string-literal and statically-evaluable computed
// property-name forms, or through a trailing comma-operator sequence around
// any of the above; parentheses and TS assertions are transparent throughout.
func (t *objectTracker) isObjectAssignCallee(callee *ast.Node) bool {
	return t.valueOf(callee, map[*ast.Node]bool{}) == valueObjectAssign
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

// hasProtoSetter reports whether the object literal has a prototype-setting
// `__proto__` property — the plain `__proto__: value` / `"__proto__": value`
// PropertyAssignment form. Computed (`["__proto__"]: v`), shorthand, and
// method forms create an ordinary own property instead and do not count.
func hasProtoSetter(node *ast.Node) bool {
	obj := node.AsObjectLiteralExpression()
	if obj == nil || obj.Properties == nil {
		return false
	}
	for _, prop := range obj.Properties.Nodes {
		if prop.Kind != ast.KindPropertyAssignment {
			continue
		}
		name := prop.AsPropertyAssignment().Name()
		if name == nil {
			continue
		}
		switch name.Kind {
		case ast.KindIdentifier:
			if name.AsIdentifier().Text == "__proto__" {
				return true
			}
		case ast.KindStringLiteral:
			if name.AsStringLiteral().Text == "__proto__" {
				return true
			}
		}
	}
	return false
}

// needsWrappingParens determines whether the fixed object literal must be
// wrapped in parentheses to remain valid at its use site (an object literal
// at the start of a statement would otherwise parse as a block, and one in
// callee position would not parse at all). Mirrors upstream's needsParens,
// adapted for tsgo's explicit ParenthesizedExpression nodes and
// BinaryExpression-collapsed AssignmentExpression.
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
		ast.KindPropertyAssignment,
		ast.KindComputedPropertyName:
		return false
	case ast.KindCallExpression:
		// `Object.assign({}, foo)()` — an object literal cannot be a call's
		// callee; as an argument it needs no parentheses.
		if parent.AsCallExpression().Expression == current {
			return !alreadyParenthesized
		}
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
// prefixed with `...`. A source object literal with a prototype-setting
// `__proto__:` property must not be unwrapped (in the merged literal it would
// set the result's prototype, which `Object.assign` never does for a source);
// it is preserved whole behind a spread instead, which copies nothing, like
// the original call. The first argument keeps unwrapping: its `__proto__:`
// sets the target's prototype in the original call too.
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

	for i, argNode := range args {
		inner := ast.SkipParentheses(argNode)
		if inner != nil && inner.Kind == ast.KindObjectLiteralExpression && (i == 0 || !hasProtoSetter(inner)) {
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
		tracker := &objectTracker{
			ctx:       ctx,
			evaluator: utils.NewStaticStringEvaluatorWithSourceFile(ctx.TypeChecker, ctx.SourceFile),
			globals:   &globalWrites{sourceFile: ctx.SourceFile},
		}
		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				call := node.AsCallExpression()
				if call.Arguments == nil || len(call.Arguments.Nodes) < 1 {
					return
				}
				if !tracker.isObjectAssignCallee(call.Expression) {
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
