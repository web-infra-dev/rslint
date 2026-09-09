package nodeutil

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

type requireValue uint8

const (
	noRequireValue requireValue = iota
	requireFunction
	requireResolve
	requireGlobalObject
)

func (value requireValue) member(name string) requireValue {
	if value == requireGlobalObject && name == "require" {
		return requireFunction
	}
	if value == requireFunction && name == "resolve" {
		return requireResolve
	}
	return noRequireValue
}

type requireCallTracker struct {
	ctx               rule.RuleContext
	names             *utils.ReferenceIndex
	propertyEvaluator *utils.StaticStringEvaluator
	variableStack     map[*ast.Symbol]bool
	globalStack       map[string]bool
	calls             []*ast.Node
}

// CollectRequireCalls returns require() and require.resolve() calls, following
// aliases and destructuring as upstream's visitRequire does. It respects
// effective globals, shadowing and writes to global roots. Calls are returned
// in traversal order; separate alias paths can return the same call more than once.
func CollectRequireCalls(ctx rule.RuleContext) []*ast.Node {
	tracker := requireCallTracker{
		ctx:           ctx,
		names:         utils.NewReferenceIndex(ctx.SourceFile, nil),
		variableStack: make(map[*ast.Symbol]bool),
		globalStack:   make(map[string]bool),
	}
	tracker.trackGlobalRoot("require", requireFunction)
	for _, name := range []string{"global", "globalThis", "self", "window"} {
		tracker.trackGlobalRoot(name, requireGlobalObject)
	}
	return tracker.calls
}

func (tracker *requireCallTracker) trackGlobalRoot(name string, value requireValue) {
	if !tracker.ctx.Globals.Access(name).IsDeclared() {
		return
	}
	references := tracker.globalReferences(name)
	for _, reference := range references {
		if utils.IsWriteReference(reference) {
			return
		}
	}
	for _, reference := range references {
		tracker.trackExpression(reference, value)
	}
}

func (tracker *requireCallTracker) globalReferences(name string) []*ast.Node {
	var references []*ast.Node
	tracker.names.ForEachReferenceByName(name, nil, func(identifier *ast.Node) bool {
		if tracker.isGlobalReference(identifier, name) {
			references = append(references, identifier)
		}
		return false
	})
	return references
}

func (tracker *requireCallTracker) isGlobalReference(identifier *ast.Node, name string) bool {
	if identifier == nil || identifier.Kind != ast.KindIdentifier || utils.IsNonReferenceIdentifier(identifier) {
		return false
	}
	if tracker.ctx.Refs != nil {
		return tracker.ctx.Refs.IsGlobalReference(identifier)
	}
	return !utils.IsShadowed(identifier, name)
}

func (tracker *requireCallTracker) trackExpression(node *ast.Node, value requireValue) {
	if node == nil {
		return
	}
	for node.Parent != nil && requireValuePassesThrough(node, node.Parent) {
		node = node.Parent
	}
	parent := node.Parent
	if parent == nil {
		return
	}

	if ast.IsAccessExpression(parent) && utils.AccessExpressionObject(parent) == node {
		name, ok := tracker.accessExpressionStaticName(parent)
		if next := value.member(name); ok && next != noRequireValue {
			tracker.trackExpression(parent, next)
		}
		return
	}

	switch parent.Kind {
	case ast.KindCallExpression:
		if parent.AsCallExpression().Expression == node && (value == requireFunction || value == requireResolve) {
			tracker.calls = append(tracker.calls, parent)
		}
	case ast.KindBinaryExpression:
		binary := parent.AsBinaryExpression()
		if binary != nil && binary.Right == node && binary.OperatorToken != nil && ast.IsAssignmentOperator(binary.OperatorToken.Kind) {
			tracker.trackAssignmentTarget(binary.Left, value)
			if !utils.IsDefaultValueInDestructuringAssignment(parent) {
				tracker.trackExpression(parent, value)
			}
		}
	case ast.KindVariableDeclaration:
		declaration := parent.AsVariableDeclaration()
		if declaration != nil && declaration.Initializer == node {
			tracker.trackAssignmentTarget(declaration.Name(), value)
		}
	case ast.KindParameter:
		parameter := parent.AsParameterDeclaration()
		if parameter != nil && parameter.Initializer == node {
			tracker.trackAssignmentTarget(parameter.Name(), value)
		}
	case ast.KindBindingElement:
		element := parent.AsBindingElement()
		if element != nil && element.Initializer == node {
			tracker.trackAssignmentTarget(element.Name(), value)
		}
	case ast.KindShorthandPropertyAssignment:
		property := parent.AsShorthandPropertyAssignment()
		if property != nil && property.ObjectAssignmentInitializer == node {
			tracker.trackAssignmentTarget(property.Name(), value)
		}
	}
}

func (tracker *requireCallTracker) trackAssignmentTarget(node *ast.Node, value requireValue) {
	node = ast.SkipParentheses(node)
	if node == nil {
		return
	}
	switch node.Kind {
	case ast.KindIdentifier:
		tracker.trackIdentifier(node, value)
	case ast.KindObjectBindingPattern:
		if value == requireResolve {
			return
		}
		pattern := node.AsBindingPattern()
		if pattern == nil || pattern.Elements == nil {
			return
		}
		for _, elementNode := range pattern.Elements.Nodes {
			element := elementNode.AsBindingElement()
			if element == nil || element.DotDotDotToken != nil || element.Name() == nil {
				continue
			}
			propertyName := element.PropertyName
			if propertyName == nil {
				propertyName = element.Name()
			}
			if name, ok := tracker.staticPropertyName(propertyName); ok && value.member(name) != noRequireValue {
				tracker.trackAssignmentTarget(element.Name(), value.member(name))
			}
		}
	case ast.KindObjectLiteralExpression:
		if value == requireResolve {
			return
		}
		for _, propertyNode := range node.AsObjectLiteralExpression().Properties.Nodes {
			switch propertyNode.Kind {
			case ast.KindPropertyAssignment:
				property := propertyNode.AsPropertyAssignment()
				if name, ok := tracker.staticPropertyName(property.Name()); ok && value.member(name) != noRequireValue {
					tracker.trackAssignmentTarget(property.Initializer, value.member(name))
				}
			case ast.KindShorthandPropertyAssignment:
				property := propertyNode.AsShorthandPropertyAssignment()
				if name, ok := tracker.staticPropertyName(property.Name()); ok && value.member(name) != noRequireValue {
					tracker.trackAssignmentTarget(property.Name(), value.member(name))
				}
			}
		}
	case ast.KindBinaryExpression:
		binary := node.AsBinaryExpression()
		if binary != nil && binary.OperatorToken != nil && binary.OperatorToken.Kind == ast.KindEqualsToken {
			tracker.trackAssignmentTarget(binary.Left, value)
		}
	}
}

func (tracker *requireCallTracker) trackIdentifier(identifier *ast.Node, value requireValue) {
	if tracker.ctx.Refs != nil {
		if symbol := tracker.ctx.Refs.Resolve(identifier); utils.IsValueSymbolDeclaredInFile(symbol, tracker.ctx.SourceFile) {
			tracker.trackVariable(symbol, value)
			return
		}
	}
	if symbol := callBindingSymbol(identifier); symbol != nil {
		tracker.trackVariable(symbol, value)
		return
	}
	name := identifier.AsIdentifier().Text
	if tracker.ctx.Globals.Access(name).IsDeclared() && tracker.isGlobalReference(identifier, name) {
		tracker.trackGlobalVariable(name, value)
	}
}

func callBindingSymbol(identifier *ast.Node) *ast.Symbol {
	if identifier == nil || identifier.Kind != ast.KindIdentifier || identifier.Parent == nil {
		return nil
	}
	declaration := identifier.Parent
	if declaration.Name() != identifier {
		return nil
	}
	return declaration.Symbol()
}

func (tracker *requireCallTracker) trackVariable(symbol *ast.Symbol, value requireValue) {
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

func (tracker *requireCallTracker) trackGlobalVariable(name string, value requireValue) {
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

func (tracker *requireCallTracker) accessExpressionStaticName(node *ast.Node) (string, bool) {
	if node.Kind == ast.KindElementAccessExpression {
		argument := node.AsElementAccessExpression().ArgumentExpression
		if tracker.propertyEvaluator == nil {
			tracker.propertyEvaluator = utils.NewStaticStringEvaluatorWithoutScope()
		}
		return tracker.propertyEvaluator.EvalToString(argument)
	}
	return utils.AccessExpressionStaticName(node)
}

func (tracker *requireCallTracker) staticPropertyName(node *ast.Node) (string, bool) {
	if node != nil && node.Kind == ast.KindComputedPropertyName {
		if tracker.propertyEvaluator == nil {
			tracker.propertyEvaluator = utils.NewStaticStringEvaluatorWithoutScope()
		}
		return tracker.propertyEvaluator.EvalToString(node.AsComputedPropertyName().Expression)
	}
	return utils.GetStaticPropertyName(node)
}

func requireValuePassesThrough(node *ast.Node, parent *ast.Node) bool {
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
		switch binary.OperatorToken.Kind {
		case ast.KindBarBarToken, ast.KindAmpersandAmpersandToken, ast.KindQuestionQuestionToken:
			return binary.Left == node || binary.Right == node
		case ast.KindCommaToken:
			return binary.Right == node
		}
	}
	return false
}
