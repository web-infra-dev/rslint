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

func (c *referenceChecker) visitBinary(binary *ast.BinaryExpression) bool {
	switch binary.OperatorToken.Kind {
	case ast.KindBarBarToken, ast.KindAmpersandAmpersandToken, ast.KindQuestionQuestionToken,
		ast.KindBarBarEqualsToken, ast.KindAmpersandAmpersandEqualsToken, ast.KindQuestionQuestionEqualsToken:
		// The logical assignments `||=`, `&&=` and `??=` short-circuit too: the
		// left side is read, and the right side is evaluated only for the same
		// left operands as `||`, `&&` and `??`.
		if c.visit(binary.Left) {
			return true
		}
		var evaluatesRight, known bool
		switch binary.OperatorToken.Kind {
		case ast.KindBarBarToken, ast.KindBarBarEqualsToken:
			var truthy bool
			truthy, known = c.evaluator.EvalControlFlowTruthiness(binary.Left)
			evaluatesRight = !truthy
		case ast.KindAmpersandAmpersandToken, ast.KindAmpersandAmpersandEqualsToken:
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
// when the expression is evaluated: it is not a declared name, a property key,
// a label or a meta property, and it is not the operand of typeof or delete,
// which do not throw for a name that does not exist. Parentheses are
// transparent, so the operand may be wrapped in them.
func isEvaluatedReference(node *ast.Node) bool {
	if utils.IsNonReferenceIdentifier(node) {
		return false
	}
	parent := node.Parent
	for parent != nil && parent.Kind == ast.KindParenthesizedExpression {
		parent = parent.Parent
	}
	if parent == nil {
		return false
	}
	switch parent.Kind {
	case ast.KindTypeOfExpression, ast.KindDeleteExpression:
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
// rejects a value that is not iterable. A value counts as not iterable only when
// the literal cannot carry a Symbol.iterator, which needs a computed key, a
// spread or an inherited prototype, so objects and classes that may have one are
// left unknown.
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
		ast.KindRegularExpressionLiteral, ast.KindFunctionExpression, ast.KindArrowFunction:
		return !object
	case ast.KindObjectLiteralExpression:
		return !object && hasOnlyNamedMembers(expression)
	case ast.KindClassExpression:
		return !object && isPlainClass(expression)
	}
	return false
}

// hasOnlyNamedMembers reports an object literal whose keys are all plain names:
// no spread, computed key or __proto__ that could add a Symbol.iterator or a
// prototype that has one.
func hasOnlyNamedMembers(object *ast.Node) bool {
	for _, property := range object.AsObjectLiteralExpression().Properties.Nodes {
		switch property.Kind {
		case ast.KindPropertyAssignment, ast.KindMethodDeclaration, ast.KindGetAccessor, ast.KindSetAccessor:
			key, ok := utils.GetStaticPropertyName(property.Name())
			if !ok || key == "__proto__" {
				return false
			}
		case ast.KindShorthandPropertyAssignment:
		default:
			return false
		}
	}
	return true
}

// isPlainClass reports a class without a heritage clause or computed member,
// which cannot inherit or define a static Symbol.iterator.
func isPlainClass(class *ast.Node) bool {
	if class.AsClassExpression().HeritageClauses != nil {
		return false
	}
	for _, member := range class.Members() {
		if name := member.Name(); name != nil && name.Kind == ast.KindComputedPropertyName {
			return false
		}
	}
	return true
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

// objectPrototypeNames are the properties every plain object inherits from
// Object.prototype. One of them that an object literal does not define itself
// is still defined, so it is never an absent property.
var objectPrototypeNames = map[string]struct{}{
	"constructor": {}, "hasOwnProperty": {}, "isPrototypeOf": {}, "propertyIsEnumerable": {},
	"toLocaleString": {}, "toString": {}, "valueOf": {}, "__proto__": {},
	"__defineGetter__": {}, "__defineSetter__": {}, "__lookupGetter__": {}, "__lookupSetter__": {},
}

// property is the source of the named property of an object destructuring. The
// property is absent only when no own property and no inherited one can supply
// it: an object literal with a __proto__ key has an unknown prototype, and every
// object inherits the names of Object.prototype.
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
			if !ok || key == "__proto__" {
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
	if found != nil {
		return valueSource{expression: found}
	}
	if _, inherited := objectPrototypeNames[name]; inherited {
		return unknownValue
	}
	return valueSource{absent: true}
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
		return c.visitTarget(shorthand.Name(), unknownValue)
	case ast.KindSpreadAssignment:
		return c.visitTarget(property.AsSpreadAssignment().Expression, unknownValue)
	case ast.KindMethodDeclaration, ast.KindGetAccessor, ast.KindSetAccessor:
		// A method or accessor is not a property target: a SyntaxError.
		return c.invalid()
	}
	return false
}
