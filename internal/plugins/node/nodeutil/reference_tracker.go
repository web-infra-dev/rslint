package nodeutil

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// referenceTrace describes only the member reads and calls needed by Node rules.
// Alias traversal is shared; callers retain module and diagnostic policy.
type referenceTrace struct {
	properties map[string]*referenceTrace
	read       func(*ast.Node)
	call       func(*ast.Node)
}

type referenceTracker struct {
	ctx               rule.RuleContext
	names             *utils.ReferenceIndex
	propertyEvaluator *utils.StaticStringEvaluator
	variableStack     map[*ast.Symbol]bool
	globalStack       map[string]bool
}

func newReferenceTracker(ctx rule.RuleContext) *referenceTracker {
	return &referenceTracker{
		ctx:           ctx,
		names:         utils.NewReferenceIndex(ctx.SourceFile, nil),
		variableStack: make(map[*ast.Symbol]bool),
		globalStack:   make(map[string]bool),
	}
}

func (tracker *referenceTracker) trackGlobals(properties map[string]*referenceTrace) {
	for name, value := range properties {
		tracker.trackGlobalRoot(name, value)
	}
	root := &referenceTrace{properties: properties}
	for _, name := range []string{"global", "globalThis", "self", "window"} {
		tracker.trackGlobalRoot(name, root)
	}
}

func (tracker *referenceTracker) trackGlobalRoot(name string, value *referenceTrace) {
	for _, reference := range tracker.unmodifiedGlobalReferences(name) {
		tracker.trackExpression(reference, value)
	}
}

func (tracker *referenceTracker) unmodifiedGlobalReferences(name string) []*ast.Node {
	if !tracker.ctx.Globals.Access(name).IsDeclared() {
		return nil
	}
	references := tracker.globalReferences(name)
	for _, reference := range references {
		if utils.IsWriteReference(reference) {
			return nil
		}
	}
	return references
}

func (tracker *referenceTracker) globalReferences(name string) []*ast.Node {
	var references []*ast.Node
	tracker.names.ForEachReferenceByName(name, nil, func(identifier *ast.Node) bool {
		if tracker.isGlobalReference(identifier, name) {
			references = append(references, identifier)
		}
		return false
	})
	return references
}

func (tracker *referenceTracker) isGlobalReference(identifier *ast.Node, name string) bool {
	if identifier == nil || identifier.Kind != ast.KindIdentifier || utils.IsNonReferenceIdentifier(identifier) {
		return false
	}
	if tracker.ctx.Refs != nil {
		return tracker.ctx.Refs.IsGlobalReference(identifier)
	}
	return !utils.IsShadowed(identifier, name)
}

func (tracker *referenceTracker) trackExpression(node *ast.Node, value *referenceTrace) {
	if node == nil {
		return
	}
	for node.Parent != nil && referenceValuePassesThrough(node, node.Parent) {
		node = node.Parent
	}
	parent := node.Parent
	if parent == nil {
		return
	}

	if ast.IsAccessExpression(parent) && utils.AccessExpressionObject(parent) == node {
		name, ok := tracker.accessExpressionStaticName(parent)
		if next := value.properties[name]; ok && next != nil {
			if next.read != nil {
				next.read(parent)
			}
			tracker.trackExpression(parent, next)
		}
		return
	}

	switch parent.Kind {
	case ast.KindCallExpression:
		if parent.AsCallExpression().Expression == node && value.call != nil {
			value.call(parent)
		}
	case ast.KindBinaryExpression:
		binary := parent.AsBinaryExpression()
		if binary != nil && binary.Right == node && binary.OperatorToken != nil && ast.IsAssignmentOperator(binary.OperatorToken.Kind) {
			tracker.trackAssignmentTarget(binary.Left, value)
			if !utils.IsDefaultValueInDestructuringAssignment(parent) {
				tracker.trackExpression(parent, value)
			}
		}
	case ast.KindVariableDeclaration, ast.KindParameter, ast.KindBindingElement:
		if parent.Initializer() == node {
			tracker.trackAssignmentTarget(parent.Name(), value)
		}
	case ast.KindShorthandPropertyAssignment:
		property := parent.AsShorthandPropertyAssignment()
		if property != nil && property.ObjectAssignmentInitializer == node {
			tracker.trackAssignmentTarget(property.Name(), value)
		}
	}
}

func (tracker *referenceTracker) trackAssignmentTarget(node *ast.Node, value *referenceTrace) {
	node = ast.SkipParentheses(node)
	if node == nil {
		return
	}
	switch node.Kind {
	case ast.KindIdentifier:
		tracker.trackIdentifier(node, value)
	case ast.KindObjectBindingPattern, ast.KindObjectLiteralExpression:
		if len(value.properties) == 0 {
			return
		}
		for _, element := range ast.GetElementsOfBindingOrAssignmentPattern(node) {
			if ast.GetRestIndicatorOfBindingOrAssignmentElement(element) != nil {
				continue
			}
			propertyName := ast.TryGetPropertyNameOfBindingOrAssignmentElement(element)
			if propertyName == nil {
				continue
			}
			if name, ok := tracker.staticPropertyName(propertyName); ok && value.properties[name] != nil {
				tracker.trackAssignmentTarget(ast.GetTargetOfBindingOrAssignmentElement(element), value.properties[name])
			}
		}
	case ast.KindBinaryExpression:
		binary := node.AsBinaryExpression()
		if binary != nil && binary.OperatorToken != nil && binary.OperatorToken.Kind == ast.KindEqualsToken {
			tracker.trackAssignmentTarget(binary.Left, value)
		}
	}
}

func (tracker *referenceTracker) trackIdentifier(identifier *ast.Node, value *referenceTrace) {
	if tracker.ctx.Refs != nil {
		if symbol := tracker.ctx.Refs.ResolveInFile(identifier); utils.IsValueSymbolDeclaredInFile(symbol, tracker.ctx.SourceFile) {
			tracker.trackVariable(symbol, value)
			return
		}
	}
	if symbol := utils.BindingNameSymbol(identifier); symbol != nil {
		tracker.trackVariable(symbol, value)
		return
	}
	name := identifier.AsIdentifier().Text
	if tracker.ctx.Globals.Access(name).IsDeclared() && tracker.isGlobalReference(identifier, name) {
		tracker.trackGlobalVariable(name, value)
	}
}

func (tracker *referenceTracker) trackVariable(symbol *ast.Symbol, value *referenceTrace) {
	if tracker.ctx.Refs == nil || symbol == nil || tracker.variableStack[symbol] {
		return
	}
	tracker.variableStack[symbol] = true
	defer delete(tracker.variableStack, symbol)
	for _, reference := range tracker.ctx.Refs.References(symbol) {
		if !ast.IsWriteOnlyAccess(reference) {
			tracker.trackExpression(reference, value)
		}
	}
}

func (tracker *referenceTracker) trackGlobalVariable(name string, value *referenceTrace) {
	if tracker.globalStack[name] {
		return
	}
	tracker.globalStack[name] = true
	defer delete(tracker.globalStack, name)
	tracker.names.ForEachReferenceByName(name, nil, func(reference *ast.Node) bool {
		if !ast.IsWriteOnlyAccess(reference) && tracker.isGlobalReference(reference, name) {
			tracker.trackExpression(reference, value)
		}
		return false
	})
}

func (tracker *referenceTracker) accessExpressionStaticName(node *ast.Node) (string, bool) {
	if node.Kind == ast.KindElementAccessExpression {
		argument := node.AsElementAccessExpression().ArgumentExpression
		return tracker.constantString(argument)
	}
	return utils.AccessExpressionStaticName(node)
}

func (tracker *referenceTracker) staticPropertyName(node *ast.Node) (string, bool) {
	if node != nil && node.Kind == ast.KindComputedPropertyName {
		return tracker.constantString(node.AsComputedPropertyName().Expression)
	}
	// tsgo's binding-property helper unwraps computed template literal keys.
	if node != nil && node.Kind == ast.KindNoSubstitutionTemplateLiteral {
		return tracker.constantString(node)
	}
	return utils.GetStaticPropertyName(node)
}

func (tracker *referenceTracker) constantString(node *ast.Node) (string, bool) {
	if tracker.propertyEvaluator == nil {
		tracker.propertyEvaluator = utils.NewStaticStringEvaluatorWithoutScope()
	}
	return tracker.propertyEvaluator.EvalToString(node)
}

func referenceValuePassesThrough(node *ast.Node, parent *ast.Node) bool {
	if ast.IsOuterExpression(parent, ast.OEKParentheses|ast.OEKAssertions|ast.OEKExpressionsWithTypeArguments) {
		return parent.Expression() == node
	}
	if parent.Kind == ast.KindConditionalExpression {
		conditional := parent.AsConditionalExpression()
		return conditional.WhenTrue == node || conditional.WhenFalse == node
	}
	if parent.Kind == ast.KindBinaryExpression {
		binary := parent.AsBinaryExpression()
		if binary == nil || binary.OperatorToken == nil {
			return false
		}
		if ast.IsLogicalOrCoalescingBinaryOperator(binary.OperatorToken.Kind) {
			return binary.Left == node || binary.Right == node
		}
		return binary.OperatorToken.Kind == ast.KindCommaToken && binary.Right == node
	}
	return false
}
