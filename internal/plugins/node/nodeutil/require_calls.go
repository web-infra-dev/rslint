package nodeutil

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// requireValue represents the CommonJS values followed by this collector.
// Scope and symbol lookup remain owned by RuleContext.Refs.
type requireValue uint8

const (
	noRequireValue requireValue = iota
	requireFunction
	requireResolve
	requireGlobalObject
	requiredModule
	requiredModuleNamespace
)

func (value requireValue) member(name string) requireValue {
	switch {
	case value == requireGlobalObject && name == "require":
		return requireFunction
	case value == requireFunction && name == "resolve":
		return requireResolve
	case value == requiredModuleNamespace && name == "default":
		return requiredModule
	default:
		return noRequireValue
	}
}

type requireCallTracker struct {
	ctx               rule.RuleContext
	names             *utils.ReferenceIndex
	propertyEvaluator *utils.StaticStringEvaluator
	variableStack     map[*ast.Symbol]bool
	globalStack       map[string]bool
	onCall            func(*ast.Node)
	includeResolve    bool
	propertyName      string
	propertyReads     map[*ast.Node]bool
}

func newRequireCallTracker(ctx rule.RuleContext) *requireCallTracker {
	return &requireCallTracker{
		ctx:           ctx,
		names:         utils.NewReferenceIndex(ctx.SourceFile, nil),
		variableStack: make(map[*ast.Symbol]bool),
		globalStack:   make(map[string]bool),
	}
}

// CollectRequireCalls returns require() and require.resolve() calls, following
// aliases and destructuring as upstream's visitRequire does. It respects
// effective globals, shadowing and writes to global roots. Calls are returned
// in traversal order; separate alias paths can return the same call more than once.
func CollectRequireCalls(ctx rule.RuleContext) []*ast.Node {
	var calls []*ast.Node
	newRequireCallTracker(ctx).visitRequireCalls(true, func(call *ast.Node) {
		calls = append(calls, call)
	})
	return calls
}

func (tracker *requireCallTracker) visitRequireCalls(includeResolve bool, onCall func(*ast.Node)) {
	tracker.includeResolve = includeResolve
	tracker.onCall = onCall
	tracker.trackGlobalRoot("require", requireFunction)
	for _, name := range []string{"global", "globalThis", "self", "window"} {
		tracker.trackGlobalRoot(name, requireGlobalObject)
	}
}

// CollectModulePropertyReads follows CommonJS module-object aliases to reads
// of one property, reusing the require collector's alias traversal. ESM imports
// use upstream's strict CJS mode: default imports expose the object; namespace
// imports expose it as .default. Module names match exactly, and property-value
// aliases are not read reports.
func CollectModulePropertyReads(ctx rule.RuleContext, moduleName, propertyName string) map[*ast.Node]bool {
	tracker := newRequireCallTracker(ctx)
	tracker.propertyName = propertyName
	tracker.propertyReads = make(map[*ast.Node]bool)
	// Follow each result while its require aliases are still on the stack,
	// matching the cycle guard of upstream's lazy reference iterator.
	tracker.visitRequireCalls(false, func(call *ast.Node) {
		args := call.Arguments()
		if len(args) == 0 {
			return
		}
		if name, ok := tracker.constantString(args[0]); ok && name == moduleName {
			tracker.trackExpression(call, requiredModule)
		}
	})
	for _, statement := range ctx.SourceFile.Statements.Nodes {
		if statement.Kind != ast.KindImportDeclaration {
			continue
		}
		declaration := statement.AsImportDeclaration()
		if declaration.ModuleSpecifier == nil || declaration.ModuleSpecifier.Text() != moduleName {
			continue
		}
		for _, binding := range utils.GetImportBindingNodes(statement) {
			value := requiredModule
			switch binding.Parent.Kind {
			case ast.KindNamespaceImport:
				value = requiredModuleNamespace
			case ast.KindImportSpecifier:
				name := binding.Parent.PropertyName()
				if name == nil {
					name = binding
				}
				if name.Text() != "default" {
					continue
				}
			}
			tracker.trackIdentifier(binding, value)
		}
	}
	return tracker.propertyReads
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
		if ok && value == requiredModule && name == tracker.propertyName {
			tracker.propertyReads[parent] = true
		} else if next := value.member(name); ok && next != noRequireValue {
			tracker.trackExpression(parent, next)
		}
		return
	}

	switch parent.Kind {
	case ast.KindCallExpression:
		if parent.AsCallExpression().Expression == node && (value == requireFunction || tracker.includeResolve && value == requireResolve) {
			tracker.onCall(parent)
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

func (tracker *requireCallTracker) trackAssignmentTarget(node *ast.Node, value requireValue) {
	node = ast.SkipParentheses(node)
	if node == nil {
		return
	}
	switch node.Kind {
	case ast.KindIdentifier:
		tracker.trackIdentifier(node, value)
	case ast.KindObjectBindingPattern, ast.KindObjectLiteralExpression:
		if value == requireResolve || value == requiredModule {
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
			if name, ok := tracker.staticPropertyName(propertyName); ok && value.member(name) != noRequireValue {
				tracker.trackAssignmentTarget(ast.GetTargetOfBindingOrAssignmentElement(element), value.member(name))
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
		return tracker.constantString(argument)
	}
	return utils.AccessExpressionStaticName(node)
}

func (tracker *requireCallTracker) staticPropertyName(node *ast.Node) (string, bool) {
	if node != nil && node.Kind == ast.KindComputedPropertyName {
		return tracker.constantString(node.AsComputedPropertyName().Expression)
	}
	// tsgo's binding-property helper unwraps computed template literal keys.
	if node != nil && node.Kind == ast.KindNoSubstitutionTemplateLiteral {
		return tracker.constantString(node)
	}
	return utils.GetStaticPropertyName(node)
}

func (tracker *requireCallTracker) constantString(node *ast.Node) (string, bool) {
	if tracker.propertyEvaluator == nil {
		tracker.propertyEvaluator = utils.NewStaticStringEvaluatorWithoutScope()
	}
	return tracker.propertyEvaluator.EvalToString(node)
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
		if ast.IsLogicalOrCoalescingBinaryOperator(binary.OperatorToken.Kind) {
			return binary.Left == node || binary.Right == node
		}
		return binary.OperatorToken.Kind == ast.KindCommaToken && binary.Right == node
	}
	return false
}
