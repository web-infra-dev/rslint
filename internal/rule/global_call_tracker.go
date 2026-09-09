package rule

import (
	"maps"
	"slices"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// GlobalCallTrace selects calls, constructor calls and named properties of a
// global value. TrackGlobalCalls follows ESLint ReferenceTracker's aliases,
// destructuring and value-preserving expressions using the file's RefStore.
type GlobalCallTrace struct {
	Call, Construct bool
	Members         map[string]*GlobalCallTrace
}

type GlobalCallOptions struct {
	// Deduplicate visits each binding/trace pair once. Use this for rules that
	// report once per call, avoiding repeated traversal of converging aliases.
	Deduplicate bool
}

type globalCallTracker struct {
	ctx               RuleContext
	identifiersByName map[string][]*ast.Node
	propertyEvaluator *utils.StaticStringEvaluator
	variableStack     map[*ast.Symbol]bool
	globalStack       map[string]bool
	calls             []*ast.Node
	options           GlobalCallOptions
	variablesSeen     map[globalCallVariable]bool
	globalsSeen       map[globalCallGlobal]bool
}

type globalCallVariable struct {
	symbol *ast.Symbol
	trace  *GlobalCallTrace
}

type globalCallGlobal struct {
	name  string
	trace *GlobalCallTrace
}

// TrackGlobalCalls excludes undeclared, shadowed or written global roots.
// Reaching a call through multiple aliases preserves upstream's duplicates;
// consumers that report once per call can deduplicate the returned nodes.
func TrackGlobalCalls(ctx RuleContext, roots map[string]*GlobalCallTrace, options GlobalCallOptions) []*ast.Node {
	tracker := &globalCallTracker{
		ctx:               ctx,
		identifiersByName: make(map[string][]*ast.Node),
		options:           options,
	}
	if options.Deduplicate {
		tracker.variablesSeen = make(map[globalCallVariable]bool)
		tracker.globalsSeen = make(map[globalCallGlobal]bool)
	} else {
		tracker.variableStack = make(map[*ast.Symbol]bool)
		tracker.globalStack = make(map[string]bool)
	}
	var visit func(*ast.Node) bool
	visit = func(node *ast.Node) bool {
		if node.Kind == ast.KindIdentifier && !utils.IsNonReferenceIdentifier(node) {
			name := node.Text()
			tracker.identifiersByName[name] = append(tracker.identifiersByName[name], node)
		}
		node.ForEachChild(visit)
		return false
	}
	ctx.SourceFile.AsNode().ForEachChild(visit)
	for _, name := range slices.Sorted(maps.Keys(roots)) {
		tracker.trackGlobalRoot(name, roots[name])
	}
	globalObject := &GlobalCallTrace{Members: roots}
	for _, name := range []string{"global", "globalThis", "self", "window"} {
		tracker.trackGlobalRoot(name, globalObject)
	}
	return tracker.calls
}

func (tracker *globalCallTracker) trackGlobalRoot(name string, value *GlobalCallTrace) {
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

func (tracker *globalCallTracker) globalReferences(name string) []*ast.Node {
	var references []*ast.Node
	for _, identifier := range tracker.identifiersByName[name] {
		if tracker.isGlobalReference(identifier, name) {
			references = append(references, identifier)
		}
	}
	return references
}

func (tracker *globalCallTracker) isGlobalReference(identifier *ast.Node, name string) bool {
	if identifier == nil || identifier.Kind != ast.KindIdentifier || utils.IsNonReferenceIdentifier(identifier) {
		return false
	}
	if tracker.ctx.Refs != nil {
		return tracker.ctx.Refs.IsGlobalReference(identifier)
	}
	return !utils.IsShadowed(identifier, name)
}

func (tracker *globalCallTracker) trackExpression(node *ast.Node, value *GlobalCallTrace) {
	if node == nil {
		return
	}
	for node.Parent != nil && callValuePassesThrough(node, node.Parent) {
		node = node.Parent
	}
	parent := node.Parent
	if parent == nil {
		return
	}

	if ast.IsAccessExpression(parent) && utils.AccessExpressionObject(parent) == node {
		name, ok := tracker.accessExpressionStaticName(parent)
		if next := value.Members[name]; ok && next != nil {
			tracker.trackExpression(parent, next)
		}
		return
	}

	switch parent.Kind {
	case ast.KindCallExpression:
		if parent.AsCallExpression().Expression == node && value.Call {
			tracker.addCall(parent)
		}
	case ast.KindNewExpression:
		if parent.AsNewExpression().Expression == node && value.Construct {
			tracker.addCall(parent)
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

func (tracker *globalCallTracker) trackAssignmentTarget(node *ast.Node, value *GlobalCallTrace) {
	node = ast.SkipParentheses(node)
	if node == nil {
		return
	}
	switch node.Kind {
	case ast.KindIdentifier:
		tracker.trackIdentifier(node, value)
	case ast.KindObjectBindingPattern:
		if len(value.Members) == 0 {
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
			if name, ok := tracker.staticPropertyName(propertyName); ok && value.Members[name] != nil {
				tracker.trackAssignmentTarget(element.Name(), value.Members[name])
			}
		}
	case ast.KindObjectLiteralExpression:
		if len(value.Members) == 0 {
			return
		}
		for _, propertyNode := range node.AsObjectLiteralExpression().Properties.Nodes {
			switch propertyNode.Kind {
			case ast.KindPropertyAssignment:
				property := propertyNode.AsPropertyAssignment()
				if name, ok := tracker.staticPropertyName(property.Name()); ok && value.Members[name] != nil {
					tracker.trackAssignmentTarget(property.Initializer, value.Members[name])
				}
			case ast.KindShorthandPropertyAssignment:
				property := propertyNode.AsShorthandPropertyAssignment()
				if name, ok := tracker.staticPropertyName(property.Name()); ok && value.Members[name] != nil {
					tracker.trackAssignmentTarget(property.Name(), value.Members[name])
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

func (tracker *globalCallTracker) trackIdentifier(identifier *ast.Node, value *GlobalCallTrace) {
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

func (tracker *globalCallTracker) trackVariable(symbol *ast.Symbol, value *GlobalCallTrace) {
	if tracker.ctx.Refs == nil || symbol == nil {
		return
	}
	if tracker.options.Deduplicate {
		key := globalCallVariable{symbol, value}
		if tracker.variablesSeen[key] {
			return
		}
		tracker.variablesSeen[key] = true
	} else {
		if tracker.variableStack[symbol] {
			return
		}
		tracker.variableStack[symbol] = true
		defer delete(tracker.variableStack, symbol)
	}
	for _, reference := range tracker.ctx.Refs.References(symbol) {
		if !ast.IsWriteOnlyAccess(reference) {
			tracker.trackExpression(reference, value)
		}
	}
}

func (tracker *globalCallTracker) trackGlobalVariable(name string, value *GlobalCallTrace) {
	if tracker.options.Deduplicate {
		key := globalCallGlobal{name, value}
		if tracker.globalsSeen[key] {
			return
		}
		tracker.globalsSeen[key] = true
	} else {
		if tracker.globalStack[name] {
			return
		}
		tracker.globalStack[name] = true
		defer delete(tracker.globalStack, name)
	}
	for _, reference := range tracker.identifiersByName[name] {
		if !ast.IsWriteOnlyAccess(reference) && tracker.isGlobalReference(reference, name) {
			tracker.trackExpression(reference, value)
		}
	}
}

func (tracker *globalCallTracker) addCall(node *ast.Node) {
	tracker.calls = append(tracker.calls, node)
}

func (tracker *globalCallTracker) accessExpressionStaticName(node *ast.Node) (string, bool) {
	if node.Kind == ast.KindElementAccessExpression {
		argument := node.AsElementAccessExpression().ArgumentExpression
		if tracker.propertyEvaluator == nil {
			tracker.propertyEvaluator = utils.NewStaticStringEvaluatorWithoutScope()
		}
		return tracker.propertyEvaluator.EvalToString(argument)
	}
	return utils.AccessExpressionStaticName(node)
}

func (tracker *globalCallTracker) staticPropertyName(node *ast.Node) (string, bool) {
	if node != nil && node.Kind == ast.KindComputedPropertyName {
		if tracker.propertyEvaluator == nil {
			tracker.propertyEvaluator = utils.NewStaticStringEvaluatorWithoutScope()
		}
		return tracker.propertyEvaluator.EvalToString(node.AsComputedPropertyName().Expression)
	}
	return utils.GetStaticPropertyName(node)
}

func callValuePassesThrough(node *ast.Node, parent *ast.Node) bool {
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
