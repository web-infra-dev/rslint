package dynamic_import_chunkname

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// referencesUndefinedGlobal reports whether evaluating the wrapper would read
// an identifier that does not exist in a fresh `vm` context. Only the wrapper
// function runs, so the bodies of functions and classes created inside it are
// not searched, and code that short-circuiting skips is not searched either.
func referencesUndefinedGlobal(root *ast.Node) bool {
	checker := &referenceChecker{evaluator: utils.NewStaticStringEvaluator(nil)}
	root.ForEachChild(checker.visit)
	return checker.found
}

type referenceChecker struct {
	evaluator      *utils.StaticStringEvaluator
	wrapperVisited bool
	found          bool
	// strict is set while a class's evaluated parts run: class code is strict.
	strict bool
	// scope holds the names bound by the class expressions and static blocks
	// around the node being visited, innermost last.
	scope []scopeEntry
}

// scopeEntry is a name bound by a class expression or a static block. A class
// name is initialized only after its heritage clause and computed keys run, so
// it is unbound until then; it is immutable everywhere.
type scopeEntry struct {
	name  string
	bound bool
	class bool
}

// lookup returns the innermost entry binding name, or nil.
func (c *referenceChecker) lookup(name string) *scopeEntry {
	for i := len(c.scope) - 1; i >= 0; i-- {
		if c.scope[i].name == name {
			return &c.scope[i]
		}
	}
	return nil
}

// readsUnresolved reports whether reading name would throw a ReferenceError,
// or read a class name before it is initialized.
func (c *referenceChecker) readsUnresolved(name string) bool {
	if entry := c.lookup(name); entry != nil {
		return !entry.bound
	}
	_, known := vmContextGlobals[name]
	return !known
}

// writesUnresolved reports whether assigning to name throws in strict mode: an
// inner class name is immutable, an undeclared name is not created, and the
// non-writable global values are read-only.
func (c *referenceChecker) writesUnresolved(name string) bool {
	if entry := c.lookup(name); entry != nil {
		return entry.class
	}
	if !c.strict {
		return false
	}
	if _, known := vmContextGlobals[name]; !known {
		return true
	}
	switch name {
	case "undefined", "NaN", "Infinity":
		return true
	}
	return false
}

// invalid records that the comment is certainly rejected by upstream, and
// stops the walk.
func (c *referenceChecker) invalid() bool {
	c.found = true
	return true
}

// visit reports true to stop the walk. An operand is searched only when it is
// known to be evaluated: when a condition or the left operand cannot be folded
// to a constant, the code it may skip is left alone rather than reported.
func (c *referenceChecker) visit(node *ast.Node) bool {
	if node == nil {
		return false
	}
	switch node.Kind {
	case ast.KindFunctionExpression:
		if c.wrapperVisited {
			return false
		}
		c.wrapperVisited = true
	case ast.KindClassExpression:
		return c.visitClass(node)
	case ast.KindArrowFunction, ast.KindFunctionDeclaration,
		ast.KindClassDeclaration, ast.KindMethodDeclaration, ast.KindGetAccessor, ast.KindSetAccessor:
		return false
	case ast.KindIdentifier:
		if isEvaluatedReference(node) && c.readsUnresolved(node.Text()) {
			c.found = true
			return true
		}
	case ast.KindBinaryExpression:
		return c.visitBinary(node.AsBinaryExpression())
	case ast.KindConditionalExpression:
		conditional := node.AsConditionalExpression()
		if c.visit(conditional.Condition) {
			return true
		}
		if truthy, known := c.evaluator.EvalControlFlowTruthiness(conditional.Condition); known {
			if truthy {
				return c.visit(conditional.WhenTrue)
			}
			return c.visit(conditional.WhenFalse)
		}
		return false
	case ast.KindPropertyAccessExpression, ast.KindElementAccessExpression, ast.KindCallExpression:
		if node.Flags&ast.NodeFlagsOptionalChain != 0 {
			return c.visitOptionalChain(node)
		}
	}
	return node.ForEachChild(c.visit)
}

// visitClass searches what runs when a class expression is evaluated: the
// extends clause and computed member names, which run before the class name is
// initialized, and then static initializers and static blocks, which run after.
// Method bodies and instance initializers run later, when called or
// constructed, so they are not searched. Class code is strict.
func (c *referenceChecker) visitClass(class *ast.Node) bool {
	savedStrict, savedScope := c.strict, len(c.scope)
	c.strict = true
	defer func() {
		c.strict = savedStrict
		c.scope = c.scope[:savedScope]
	}()

	index := -1
	if name := class.Name(); name != nil {
		c.scope = append(c.scope, scopeEntry{name: name.Text(), class: true})
		index = len(c.scope) - 1
	}

	data := class.AsClassExpression()
	if data.HeritageClauses != nil {
		for _, clause := range data.HeritageClauses.Nodes {
			for _, heritage := range clause.AsHeritageClause().Types.Nodes {
				if c.visit(heritage.AsExpressionWithTypeArguments().Expression) {
					return true
				}
			}
		}
	}
	for _, member := range class.Members() {
		if name := member.Name(); name != nil && name.Kind == ast.KindComputedPropertyName {
			if c.visit(name.AsComputedPropertyName().Expression) {
				return true
			}
		}
	}

	if index >= 0 {
		c.scope[index].bound = true
	}
	for _, member := range class.Members() {
		switch member.Kind {
		case ast.KindPropertyDeclaration:
			if !ast.HasStaticModifier(member) {
				continue
			}
			if initializer := member.AsPropertyDeclaration().Initializer; initializer != nil && c.visit(initializer) {
				return true
			}
		case ast.KindClassStaticBlockDeclaration:
			if c.visitStaticBlock(member.AsClassStaticBlockDeclaration().Body) {
				return true
			}
		}
	}
	return false
}

// visitStaticBlock searches a static block with the names it declares in
// scope, so reads of those names are not reported.
func (c *referenceChecker) visitStaticBlock(block *ast.Node) bool {
	savedScope := len(c.scope)
	defer func() { c.scope = c.scope[:savedScope] }()
	for _, name := range declaredNames(block) {
		c.scope = append(c.scope, scopeEntry{name: name, bound: true})
	}
	return c.visit(block)
}

// declaredNames returns the simple names that a static block binds for its
// whole body: `var` and top-level declarations. A let, const, class or function
// declared inside a nested block is visible only in that block.
func declaredNames(block *ast.Node) []string {
	var names []string
	nested := 0
	var visit func(node *ast.Node) bool
	visit = func(node *ast.Node) bool {
		switch node.Kind {
		case ast.KindBlock, ast.KindForStatement, ast.KindForInStatement, ast.KindForOfStatement,
			ast.KindSwitchStatement, ast.KindCatchClause:
			nested++
			node.ForEachChild(visit)
			nested--
			return false
		case ast.KindFunctionExpression, ast.KindArrowFunction, ast.KindClassExpression,
			ast.KindMethodDeclaration, ast.KindGetAccessor, ast.KindSetAccessor, ast.KindConstructor:
			// Their declarations are local to them.
			return false
		case ast.KindVariableDeclaration, ast.KindBindingElement, ast.KindFunctionDeclaration, ast.KindClassDeclaration:
			if name := node.Name(); name != nil && name.Kind == ast.KindIdentifier && (nested == 0 || !isLexicalDeclaration(node)) {
				names = append(names, name.Text())
			}
			if node.Kind == ast.KindFunctionDeclaration || node.Kind == ast.KindClassDeclaration {
				return false
			}
		}
		return node.ForEachChild(visit)
	}
	block.ForEachChild(visit)
	return names
}

// isLexicalDeclaration reports whether a variable or class or function binding
// is block scoped: a let or const, or a class or function declaration.
func isLexicalDeclaration(node *ast.Node) bool {
	if node.Kind == ast.KindFunctionDeclaration || node.Kind == ast.KindClassDeclaration {
		return true
	}
	list := node
	for list != nil && list.Kind != ast.KindVariableDeclarationList {
		list = list.Parent
	}
	return list != nil && list.Flags&(ast.NodeFlagsLet|ast.NodeFlagsConst) != 0
}

func (c *referenceChecker) visitBinary(binary *ast.BinaryExpression) bool {
	switch binary.OperatorToken.Kind {
	case ast.KindBarBarToken, ast.KindAmpersandAmpersandToken, ast.KindQuestionQuestionToken:
		if c.visit(binary.Left) {
			return true
		}
		// The right operand runs only for a falsy (`||`), truthy (`&&`) or
		// nullish (`??`) left operand.
		var evaluatesRight, known bool
		switch binary.OperatorToken.Kind {
		case ast.KindBarBarToken:
			var truthy bool
			truthy, known = c.evaluator.EvalControlFlowTruthiness(binary.Left)
			evaluatesRight = !truthy
		case ast.KindAmpersandAmpersandToken:
			evaluatesRight, known = c.evaluator.EvalControlFlowTruthiness(binary.Left)
		default:
			evaluatesRight, known = c.evaluator.EvalControlFlowNullish(binary.Left)
		}
		return known && evaluatesRight && c.visit(binary.Right)
	case ast.KindEqualsToken:
		if c.visitTarget(binary.Left, valueSource{expression: binary.Right}) {
			return true
		}
		return c.visit(binary.Right)
	}
	return binary.Node.ForEachChild(c.visit)
}

// visitOptionalChain searches the operand of a `?.` chain, and the rest of the
// chain only when no `?.` in it can short-circuit.
func (c *referenceChecker) visitOptionalChain(node *ast.Node) bool {
	var expression *ast.Node
	switch node.Kind {
	case ast.KindPropertyAccessExpression:
		expression = node.AsPropertyAccessExpression().Expression
	case ast.KindElementAccessExpression:
		expression = node.AsElementAccessExpression().Expression
	default:
		expression = node.AsCallExpression().Expression
	}
	if c.visit(expression) {
		return true
	}
	if !c.chainEvaluated(node) {
		return false
	}
	return node.ForEachChild(func(child *ast.Node) bool {
		if child == expression {
			return false
		}
		return c.visit(child)
	})
}

// chainEvaluated reports whether every `?.` in the chain below node is known
// to continue, that is, its operand is a constant that is not null or undefined.
func (c *referenceChecker) chainEvaluated(node *ast.Node) bool {
	for node != nil && node.Flags&ast.NodeFlagsOptionalChain != 0 {
		var expression *ast.Node
		var questionDot *ast.Node
		switch node.Kind {
		case ast.KindPropertyAccessExpression:
			access := node.AsPropertyAccessExpression()
			expression, questionDot = access.Expression, access.QuestionDotToken
		case ast.KindElementAccessExpression:
			access := node.AsElementAccessExpression()
			expression, questionDot = access.Expression, access.QuestionDotToken
		case ast.KindCallExpression:
			call := node.AsCallExpression()
			expression, questionDot = call.Expression, call.QuestionDotToken
		default:
			return true
		}
		if questionDot != nil {
			if nullish, known := c.evaluator.EvalControlFlowNullish(expression); !known || nullish {
				return false
			}
		}
		node = expression
	}
	return true
}

// isEvaluatedReference reports whether the identifier is read as a variable
// when the expression is evaluated. Parentheses are transparent, so the parent
// is the first node around them.
func isEvaluatedReference(node *ast.Node) bool {
	child := node
	parent := node.Parent
	for parent != nil && parent.Kind == ast.KindParenthesizedExpression {
		child = parent
		parent = parent.Parent
	}
	if parent == nil {
		return false
	}
	switch parent.Kind {
	case ast.KindPropertyAssignment:
		return parent.AsPropertyAssignment().Initializer == child
	case ast.KindPropertyAccessExpression:
		return parent.AsPropertyAccessExpression().Expression == child
	case ast.KindTypeOfExpression, ast.KindDeleteExpression:
		// `typeof missing` and `delete missing` (sloppy mode) do not throw.
		return false
	case ast.KindPropertyDeclaration, ast.KindVariableDeclaration, ast.KindBindingElement, ast.KindParameter:
		// A declared name is a binding, not a read; its initializer is a read.
		return parent.Initializer() == child
	case ast.KindMethodDeclaration, ast.KindGetAccessor, ast.KindSetAccessor, ast.KindFunctionExpression,
		ast.KindLabeledStatement, ast.KindBreakStatement, ast.KindContinueStatement, ast.KindMetaProperty,
		ast.KindQualifiedName:
		return false
	}
	return true
}

// valueSource is what a destructuring target reads from: a statically known
// expression, or an absent position of a literal, which reads undefined. The
// zero value is an unknown value.
type valueSource struct {
	expression *ast.Node
	absent     bool
}

var unknownValue = valueSource{}

// valueUndefined reports whether the value is undefined, when a literal or a
// void/undefined expression says so. Anything else is unknown.
func (source valueSource) valueUndefined() (undefined bool, known bool) {
	if source.absent {
		return true, true
	}
	if source.expression == nil {
		return false, false
	}
	expression := ast.SkipParentheses(source.expression)
	switch expression.Kind {
	case ast.KindVoidExpression:
		return true, true
	case ast.KindIdentifier:
		if expression.Text() == "undefined" {
			return true, true
		}
	case ast.KindStringLiteral, ast.KindNumericLiteral, ast.KindBigIntLiteral, ast.KindTrueKeyword,
		ast.KindFalseKeyword, ast.KindNullKeyword, ast.KindNoSubstitutionTemplateLiteral,
		ast.KindTemplateExpression, ast.KindRegularExpressionLiteral, ast.KindArrayLiteralExpression,
		ast.KindObjectLiteralExpression, ast.KindFunctionExpression, ast.KindArrowFunction, ast.KindClassExpression:
		return false, true
	}
	return false, false
}

// cannotDestructure reports whether destructuring a known value throws a
// TypeError: null and undefined never destructure, and an array pattern also
// rejects a value that is not iterable.
func (source valueSource) cannotDestructure(object bool) bool {
	if source.absent {
		return true
	}
	if source.expression == nil {
		return false
	}
	expression := ast.SkipParentheses(source.expression)
	switch expression.Kind {
	case ast.KindVoidExpression, ast.KindNullKeyword:
		return true
	case ast.KindIdentifier:
		return expression.Text() == "undefined"
	case ast.KindNumericLiteral, ast.KindBigIntLiteral, ast.KindTrueKeyword, ast.KindFalseKeyword,
		ast.KindObjectLiteralExpression, ast.KindRegularExpressionLiteral, ast.KindFunctionExpression,
		ast.KindArrowFunction, ast.KindClassExpression:
		return !object
	}
	return false
}

// element is the source of the index-th element of an array destructuring.
func (source valueSource) element(index int) valueSource {
	if source.absent || source.expression == nil {
		return unknownValue
	}
	array := ast.SkipParentheses(source.expression)
	if array.Kind != ast.KindArrayLiteralExpression {
		return unknownValue
	}
	elements := array.AsArrayLiteralExpression().Elements.Nodes
	// A spread before or at index makes the position unknown.
	for i := 0; i <= index && i < len(elements); i++ {
		if elements[i].Kind == ast.KindSpreadElement {
			return unknownValue
		}
	}
	if index >= len(elements) || elements[index].Kind == ast.KindOmittedExpression {
		return valueSource{absent: true}
	}
	return valueSource{expression: elements[index]}
}

// property is the source of the named property of an object destructuring.
func (source valueSource) property(name string, nameKnown bool) valueSource {
	if !nameKnown || source.absent || source.expression == nil {
		return unknownValue
	}
	object := ast.SkipParentheses(source.expression)
	if object.Kind != ast.KindObjectLiteralExpression {
		return unknownValue
	}
	var found *ast.Node
	for _, property := range object.AsObjectLiteralExpression().Properties.Nodes {
		switch property.Kind {
		case ast.KindPropertyAssignment:
			assignment := property.AsPropertyAssignment()
			key, ok := utils.GetStaticPropertyName(assignment.Name())
			if !ok {
				return unknownValue
			}
			if key == name {
				found = assignment.Initializer
			}
		case ast.KindShorthandPropertyAssignment:
			shorthand := property.AsShorthandPropertyAssignment()
			if shorthand.Name().Text() == name {
				found = shorthand.Name()
			}
		default:
			// Spreads and methods may provide anything.
			return unknownValue
		}
	}
	if found == nil {
		return valueSource{absent: true}
	}
	return valueSource{expression: found}
}

// visitTarget walks the target of a destructuring or plain assignment. A name
// that is only written is not a read. Member objects, computed keys and
// default values are evaluated as the destructuring reaches them.
func (c *referenceChecker) visitTarget(target *ast.Node, source valueSource) bool {
	if target.Kind == ast.KindParenthesizedExpression {
		// `([a]) = []` is a SyntaxError: a parenthesized pattern is not a pattern.
		switch ast.SkipParentheses(target).Kind {
		case ast.KindArrayLiteralExpression, ast.KindObjectLiteralExpression:
			return c.invalid()
		}
	}
	target = ast.SkipParentheses(target)
	switch target.Kind {
	case ast.KindIdentifier:
		if c.writesUnresolved(target.Text()) {
			return c.invalid()
		}
		return false
	case ast.KindNumericLiteral, ast.KindStringLiteral, ast.KindBigIntLiteral, ast.KindTrueKeyword,
		ast.KindFalseKeyword, ast.KindNullKeyword:
		// A literal is not an assignment target: a SyntaxError.
		return c.invalid()
	case ast.KindPropertyAccessExpression:
		return c.visit(target.AsPropertyAccessExpression().Expression)
	case ast.KindElementAccessExpression:
		access := target.AsElementAccessExpression()
		return c.visit(access.Expression) || c.visit(access.ArgumentExpression)
	case ast.KindArrayLiteralExpression:
		if source.cannotDestructure(false) {
			return c.invalid()
		}
		for i, element := range target.AsArrayLiteralExpression().Elements.Nodes {
			switch element.Kind {
			case ast.KindOmittedExpression:
				continue
			case ast.KindSpreadElement:
				if c.visitTarget(element.AsSpreadElement().Expression, unknownValue) {
					return true
				}
			default:
				if c.visitElement(element, source.element(i)) {
					return true
				}
			}
		}
		return false
	case ast.KindObjectLiteralExpression:
		if source.cannotDestructure(true) {
			return c.invalid()
		}
		for _, property := range target.AsObjectLiteralExpression().Properties.Nodes {
			if c.visitProperty(property, source) {
				return true
			}
		}
		return false
	}
	return c.visit(target)
}

// visitElement handles one array element or property value, which may carry a
// default: `target = fallback`.
func (c *referenceChecker) visitElement(element *ast.Node, source valueSource) bool {
	if element.Kind == ast.KindBinaryExpression {
		binary := element.AsBinaryExpression()
		if binary.OperatorToken.Kind == ast.KindEqualsToken {
			return c.visitDefault(binary.Left, binary.Right, source)
		}
	}
	return c.visitTarget(element, source)
}

// visitDefault handles `target = fallback`. The fallback runs only when the
// value is undefined. When that is not known statically the fallback is not
// searched, so a default that may not run is not reported.
func (c *referenceChecker) visitDefault(target, fallback *ast.Node, source valueSource) bool {
	undefined, known := source.valueUndefined()
	switch {
	case !known:
		return c.visitTarget(target, unknownValue)
	case undefined:
		if c.visit(fallback) {
			return true
		}
		return c.visitTarget(target, valueSource{expression: fallback})
	default:
		return c.visitTarget(target, source)
	}
}

func (c *referenceChecker) visitProperty(property *ast.Node, source valueSource) bool {
	switch property.Kind {
	case ast.KindPropertyAssignment:
		assignment := property.AsPropertyAssignment()
		if name := assignment.Name(); name.Kind == ast.KindComputedPropertyName {
			// A computed key is evaluated whenever the pattern is reached.
			if c.visit(name.AsComputedPropertyName().Expression) {
				return true
			}
		}
		key, ok := utils.GetStaticPropertyName(assignment.Name())
		return c.visitElement(assignment.Initializer, source.property(key, ok))
	case ast.KindShorthandPropertyAssignment:
		shorthand := property.AsShorthandPropertyAssignment()
		if shorthand.ObjectAssignmentInitializer != nil {
			return c.visitDefault(shorthand.Name(), shorthand.ObjectAssignmentInitializer,
				source.property(shorthand.Name().Text(), true))
		}
		return false
	case ast.KindSpreadAssignment:
		return c.visitTarget(property.AsSpreadAssignment().Expression, unknownValue)
	case ast.KindMethodDeclaration, ast.KindGetAccessor, ast.KindSetAccessor:
		// A method or accessor is not a property target: a SyntaxError.
		return c.invalid()
	}
	return false
}
