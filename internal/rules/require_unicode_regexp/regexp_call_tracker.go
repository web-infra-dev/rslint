package require_unicode_regexp

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// regexpCallTracker follows the built-in RegExp value through the same
// flow-insensitive value propagation forms used by ESLint's ReferenceTracker.
// It starts from effective global bindings, drops a modified root, and records
// calls reached through aliases, destructuring, and value-preserving wrappers.
type regexpCallTracker struct {
	ctx               rule.RuleContext
	identifiersByName map[string][]*ast.Node
	propertyEvaluator *utils.StaticStringEvaluator
	disabledRoots     map[string]bool
	tracedVariables   map[regexpTraceVariable]bool
	tracedGlobals     map[regexpTraceGlobal]bool
	calls             map[*ast.Node]bool
}

type regexpTraceValue uint8

const (
	regexpTraceConstructor regexpTraceValue = iota
	regexpTraceGlobalObject
)

type regexpTraceVariable struct {
	symbol *ast.Symbol
	value  regexpTraceValue
}

type regexpTraceGlobal struct {
	name  string
	value regexpTraceValue
}

var regexpGlobalObjectNames = [...]string{"globalThis", "window", "self", "global"}

func newRegexpCallTracker(ctx rule.RuleContext) *regexpCallTracker {
	tracker := &regexpCallTracker{
		ctx:               ctx,
		identifiersByName: make(map[string][]*ast.Node),
	}
	if ctx.SourceFile != nil {
		var visit func(*ast.Node) bool
		visit = func(node *ast.Node) bool {
			if node.Kind == ast.KindIdentifier && !utils.IsNonReferenceIdentifier(node) {
				name := node.AsIdentifier().Text
				tracker.identifiersByName[name] = append(tracker.identifiersByName[name], node)
			}
			node.ForEachChild(visit)
			return false
		}
		ctx.SourceFile.AsNode().ForEachChild(visit)
	}

	tracker.trackGlobalRoot("RegExp", regexpTraceConstructor)
	for _, name := range regexpGlobalObjectNames {
		tracker.trackGlobalRoot(name, regexpTraceGlobalObject)
	}
	return tracker
}

func (tracker *regexpCallTracker) isCall(node *ast.Node) bool {
	return tracker != nil && tracker.calls[node]
}

func (tracker *regexpCallTracker) trackGlobalRoot(name string, value regexpTraceValue) {
	if !tracker.ctx.Globals.Access(name).IsDeclared() {
		return
	}
	references := tracker.globalReferences(name)
	for _, reference := range references {
		if utils.IsWriteReference(reference) {
			if tracker.disabledRoots == nil {
				tracker.disabledRoots = make(map[string]bool)
			}
			tracker.disabledRoots[name] = true
			return
		}
	}
	for _, reference := range references {
		tracker.trackExpression(reference, value)
	}
}

func (tracker *regexpCallTracker) globalReferences(name string) []*ast.Node {
	var references []*ast.Node
	for _, identifier := range tracker.identifiersByName[name] {
		if tracker.isGlobalReference(identifier, name) {
			references = append(references, identifier)
		}
	}
	return references
}

func (tracker *regexpCallTracker) isGlobalReference(identifier *ast.Node, name string) bool {
	if identifier == nil || identifier.Kind != ast.KindIdentifier || utils.IsNonReferenceIdentifier(identifier) {
		return false
	}
	if tracker.ctx.Refs != nil {
		if symbol := tracker.ctx.Refs.Resolve(identifier); symbol != nil {
			return !utils.IsValueSymbolDeclaredInFile(symbol, tracker.ctx.SourceFile)
		}
	}
	return !utils.IsShadowed(identifier, name)
}

func (tracker *regexpCallTracker) trackExpression(node *ast.Node, value regexpTraceValue) {
	if node == nil {
		return
	}
	for node.Parent != nil && regexpValuePassesThrough(node, node.Parent) {
		node = node.Parent
	}
	parent := node.Parent
	if parent == nil {
		return
	}

	if ast.IsAccessExpression(parent) && utils.AccessExpressionObject(parent) == node {
		if value != regexpTraceGlobalObject {
			return
		}
		name, ok := tracker.accessExpressionStaticName(parent)
		if ok && name == "RegExp" {
			tracker.trackExpression(parent, regexpTraceConstructor)
		}
		return
	}

	switch parent.Kind {
	case ast.KindCallExpression:
		if parent.AsCallExpression().Expression == node && value == regexpTraceConstructor {
			tracker.addCall(parent)
		}
	case ast.KindNewExpression:
		if parent.AsNewExpression().Expression == node && value == regexpTraceConstructor {
			tracker.addCall(parent)
		}
	case ast.KindBinaryExpression:
		binary := parent.AsBinaryExpression()
		if binary != nil && binary.Right == node && binary.OperatorToken != nil && ast.IsAssignmentOperator(binary.OperatorToken.Kind) {
			tracker.trackAssignmentTarget(binary.Left, value)
			tracker.trackExpression(parent, value)
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

func (tracker *regexpCallTracker) trackAssignmentTarget(node *ast.Node, value regexpTraceValue) {
	node = ast.SkipParentheses(node)
	if node == nil {
		return
	}
	switch node.Kind {
	case ast.KindIdentifier:
		tracker.trackIdentifier(node, value)
	case ast.KindObjectBindingPattern:
		if value != regexpTraceGlobalObject {
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
			if name, ok := tracker.staticPropertyName(propertyName); ok && name == "RegExp" {
				tracker.trackAssignmentTarget(element.Name(), regexpTraceConstructor)
			}
		}
	case ast.KindObjectLiteralExpression:
		if value != regexpTraceGlobalObject {
			return
		}
		for _, propertyNode := range node.AsObjectLiteralExpression().Properties.Nodes {
			switch propertyNode.Kind {
			case ast.KindPropertyAssignment:
				property := propertyNode.AsPropertyAssignment()
				if name, ok := tracker.staticPropertyName(property.Name()); ok && name == "RegExp" {
					tracker.trackAssignmentTarget(property.Initializer, regexpTraceConstructor)
				}
			case ast.KindShorthandPropertyAssignment:
				property := propertyNode.AsShorthandPropertyAssignment()
				if name, ok := tracker.staticPropertyName(property.Name()); ok && name == "RegExp" {
					tracker.trackAssignmentTarget(property.Name(), regexpTraceConstructor)
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

func (tracker *regexpCallTracker) trackIdentifier(identifier *ast.Node, value regexpTraceValue) {
	if tracker.ctx.Refs != nil {
		if symbol := tracker.ctx.Refs.Resolve(identifier); utils.IsValueSymbolDeclaredInFile(symbol, tracker.ctx.SourceFile) {
			tracker.trackVariable(symbol, value)
			return
		}
	}
	if symbol := regexpBindingSymbol(identifier); symbol != nil {
		tracker.trackVariable(symbol, value)
		return
	}
	name := identifier.AsIdentifier().Text
	if tracker.ctx.Globals.Access(name).IsDeclared() && tracker.isGlobalReference(identifier, name) {
		tracker.trackGlobalVariable(name, value)
	}
}

func regexpBindingSymbol(identifier *ast.Node) *ast.Symbol {
	if identifier == nil || identifier.Kind != ast.KindIdentifier || identifier.Parent == nil {
		return nil
	}
	declaration := identifier.Parent
	if declaration.Name() != identifier {
		return nil
	}
	return declaration.Symbol()
}

func (tracker *regexpCallTracker) trackVariable(symbol *ast.Symbol, value regexpTraceValue) {
	if tracker.ctx.Refs == nil || symbol == nil {
		return
	}
	key := regexpTraceVariable{symbol: symbol, value: value}
	if tracker.tracedVariables != nil && tracker.tracedVariables[key] {
		return
	}
	if tracker.tracedVariables == nil {
		tracker.tracedVariables = make(map[regexpTraceVariable]bool)
	}
	tracker.tracedVariables[key] = true
	for _, reference := range tracker.ctx.Refs.References(symbol) {
		if !ast.IsWriteOnlyAccess(reference) {
			tracker.trackExpression(reference, value)
		}
	}
}

func (tracker *regexpCallTracker) trackGlobalVariable(name string, value regexpTraceValue) {
	key := regexpTraceGlobal{name: name, value: value}
	if tracker.tracedGlobals != nil && tracker.tracedGlobals[key] {
		return
	}
	if tracker.tracedGlobals == nil {
		tracker.tracedGlobals = make(map[regexpTraceGlobal]bool)
	}
	tracker.tracedGlobals[key] = true
	for _, reference := range tracker.identifiersByName[name] {
		if !ast.IsWriteOnlyAccess(reference) && tracker.isGlobalReference(reference, name) {
			tracker.trackExpression(reference, value)
		}
	}
}

func (tracker *regexpCallTracker) addCall(node *ast.Node) {
	if tracker.calls == nil {
		tracker.calls = make(map[*ast.Node]bool)
	}
	tracker.calls[node] = true
}

func (tracker *regexpCallTracker) accessExpressionStaticName(node *ast.Node) (string, bool) {
	if node.Kind == ast.KindElementAccessExpression {
		argument := node.AsElementAccessExpression().ArgumentExpression
		if tracker.propertyEvaluator == nil {
			tracker.propertyEvaluator = utils.NewStaticStringEvaluatorWithoutScope()
		}
		return tracker.propertyEvaluator.Eval(argument)
	}
	return utils.AccessExpressionStaticName(node)
}

func (tracker *regexpCallTracker) staticPropertyName(node *ast.Node) (string, bool) {
	if node != nil && node.Kind == ast.KindComputedPropertyName {
		if tracker.propertyEvaluator == nil {
			tracker.propertyEvaluator = utils.NewStaticStringEvaluatorWithoutScope()
		}
		return tracker.propertyEvaluator.Eval(node.AsComputedPropertyName().Expression)
	}
	return utils.GetStaticPropertyName(node)
}

func regexpValuePassesThrough(node *ast.Node, parent *ast.Node) bool {
	if ast.IsOuterExpression(parent, ast.OEKParentheses|ast.OEKAssertions) {
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
