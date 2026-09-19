package referencetracker

import (
	"maps"
	"slices"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// Trace describes the reads, calls and constructors a rule needs.
// Read mirrors eslint-utils' READ event, including member assignment targets
// and destructuring elements. Callbacks own filtering and diagnostics.
type Trace struct {
	Properties            map[string]*Trace
	Read, Call, Construct func(*ast.Node)
}

func (trace *Trace) read(node *ast.Node) {
	if trace.Read != nil {
		trace.Read(node)
	}
}

// Tracker follows flow-insensitive aliases and static properties. Active
// variable stacks prevent cycles without suppressing reports from independent
// paths. Only its read-only name index is shared with other rules in the file.
type Tracker struct {
	ctx               rule.RuleContext
	names             *utils.ReferenceIndex
	propertyEvaluator *utils.StaticStringEvaluator
	variableStack     map[*ast.Symbol]bool
	globalStack       map[string]bool
	stableOnly        bool
}

type referenceNamesKey struct{}

// New follows API aliases using existing name and symbol indexes.
// The linter supplies ctx.Refs; the name index is shared across consumers per file.
func New(ctx rule.RuleContext) *Tracker {
	names := rule.CachedByFile(ctx, referenceNamesKey{}, func() *utils.ReferenceIndex {
		return utils.NewReferenceIndex(ctx.SourceFile, nil)
	})
	return &Tracker{ctx: ctx, names: names,
		variableStack: make(map[*ast.Symbol]bool), globalStack: make(map[string]bool)}
}

// NewForReplacement follows declaration-initialized, unwritten aliases only.
// Replacing a whole receiver requires a definite value, so conditional/default
// values and side-effecting pass-through expressions stop tracking.
func NewForReplacement(ctx rule.RuleContext) *Tracker {
	tracker := New(ctx)
	tracker.stableOnly = true
	return tracker
}

// TrackGlobals follows unmodified configured globals and their global-object properties.
func (tracker *Tracker) TrackGlobals(globals map[string]*Trace) {
	for _, name := range slices.Sorted(maps.Keys(globals)) {
		tracker.TrackGlobal(name, globals[name])
	}
	for _, name := range []string{"global", "globalThis", "self", "window"} {
		tracker.TrackGlobal(name, &Trace{Properties: globals})
	}
}

// TrackGlobal follows one unmodified configured global. Callers may select
// their own global-object roots and traversal order.
func (tracker *Tracker) TrackGlobal(name string, value *Trace) {
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
		value.read(reference)
		tracker.TrackExpression(reference, value)
	}
}

func (tracker *Tracker) globalReferences(name string) []*ast.Node {
	var references []*ast.Node
	tracker.names.ForEachReferenceByName(name, nil, func(identifier *ast.Node) bool {
		if tracker.isGlobalReference(identifier) {
			references = append(references, identifier)
		}
		return false
	})
	return references
}

// TrackExpression follows a known value from an expression. It does not emit
// a Read for the seed itself; module adapters own those import/load events.
func (tracker *Tracker) TrackExpression(node *ast.Node, value *Trace) {
	if node == nil {
		return
	}
	for node.Parent != nil && referenceValuePassesThrough(node, node.Parent) {
		if tracker.stableOnly && !ast.IsOuterExpression(node.Parent, ast.OEKParentheses|ast.OEKAssertions|ast.OEKExpressionsWithTypeArguments) {
			return
		}
		node = node.Parent
	}
	parent := node.Parent
	if parent == nil {
		return
	}

	if ast.IsAccessExpression(parent) && utils.AccessExpressionObject(parent) == node {
		if utils.IsInJsxTagName(parent) {
			return
		}
		name, ok := tracker.accessExpressionStaticName(parent)
		if next := value.Properties[name]; ok && next != nil {
			next.read(parent)
			tracker.TrackExpression(parent, next)
		}
		return
	}

	switch parent.Kind {
	case ast.KindCallExpression:
		if parent.AsCallExpression().Expression == node && value.Call != nil {
			value.Call(parent)
		}
	case ast.KindNewExpression:
		if parent.AsNewExpression().Expression == node && value.Construct != nil {
			value.Construct(parent)
		}
	case ast.KindBinaryExpression:
		if tracker.stableOnly {
			return
		}
		binary := parent.AsBinaryExpression()
		if binary != nil && binary.Right == node && binary.OperatorToken != nil && ast.IsAssignmentOperator(binary.OperatorToken.Kind) {
			tracker.TrackBinding(binary.Left, value)
			if !utils.IsDefaultValueInDestructuringAssignment(parent) {
				tracker.TrackExpression(parent, value)
			}
		}
	case ast.KindVariableDeclaration, ast.KindParameter, ast.KindBindingElement:
		if tracker.stableOnly && parent.Kind != ast.KindVariableDeclaration {
			return
		}
		if parent.Initializer() == node {
			tracker.TrackBinding(parent.Name(), value)
		}
	case ast.KindShorthandPropertyAssignment:
		if tracker.stableOnly {
			return
		}
		property := parent.AsShorthandPropertyAssignment()
		if property != nil && property.ObjectAssignmentInitializer == node {
			tracker.TrackBinding(property.Name(), value)
		}
	}
}

// TrackBinding follows a known value assigned to an identifier or pattern.
// Import adapters may seed a declaration name through this same entry point.
func (tracker *Tracker) TrackBinding(node *ast.Node, value *Trace) {
	node = ast.SkipParentheses(node)
	if node == nil {
		return
	}
	switch node.Kind {
	case ast.KindIdentifier:
		tracker.trackIdentifier(node, value)
	case ast.KindObjectBindingPattern, ast.KindObjectLiteralExpression:
		for _, element := range ast.GetElementsOfBindingOrAssignmentPattern(node) {
			if ast.GetRestIndicatorOfBindingOrAssignmentElement(element) != nil {
				continue
			}
			propertyName := ast.TryGetPropertyNameOfBindingOrAssignmentElement(element)
			if propertyName == nil {
				continue
			}
			if name, ok := tracker.staticPropertyName(propertyName); ok && value.Properties[name] != nil {
				value.Properties[name].read(element)
				tracker.TrackBinding(ast.GetTargetOfBindingOrAssignmentElement(element), value.Properties[name])
			}
		}
	case ast.KindBinaryExpression:
		binary := node.AsBinaryExpression()
		if binary != nil && binary.OperatorToken != nil && binary.OperatorToken.Kind == ast.KindEqualsToken {
			tracker.TrackBinding(binary.Left, value)
		}
	}
}

func (tracker *Tracker) trackIdentifier(identifier *ast.Node, value *Trace) {
	symbol := tracker.ctx.Refs.ResolveInFile(identifier)
	if symbol == nil && utils.IsDeclarationIdentifier(identifier) {
		// Binding names have binder symbols, but shorthand assignment properties
		// do not declare variables. Their property symbol must not hide a global.
		symbol = utils.BindingNameSymbol(identifier)
	}
	if symbol != nil {
		if tracker.stableOnly && !tracker.isStableBinding(identifier, symbol) {
			return
		}
		tracker.trackVariable(symbol, value)
		return
	}
	if tracker.stableOnly {
		return
	}
	name := identifier.AsIdentifier().Text
	if tracker.ctx.Globals.Access(name).IsDeclared() && tracker.isGlobalReference(identifier) {
		tracker.trackGlobalVariable(name, value)
	}
}

func (tracker *Tracker) isStableBinding(identifier *ast.Node, symbol *ast.Symbol) bool {
	if len(symbol.Declarations) != 1 {
		return false
	}
	declaration := symbol.Declarations[0]
	if declaration.Name() != identifier {
		return false
	}
	for declaration.Kind == ast.KindBindingElement {
		if declaration.Initializer() != nil || declaration.Parent == nil || declaration.Parent.Parent == nil {
			return false
		}
		declaration = declaration.Parent.Parent
	}
	if declaration.Kind != ast.KindVariableDeclaration || declaration.Initializer() == nil {
		return false
	}
	for _, reference := range tracker.ctx.Refs.References(symbol) {
		if utils.IsWriteReference(reference) || reference.Pos() < declaration.End() {
			return false
		}
	}
	return true
}

func (tracker *Tracker) trackVariable(symbol *ast.Symbol, value *Trace) {
	if symbol == nil || tracker.variableStack[symbol] {
		return
	}
	tracker.variableStack[symbol] = true
	defer delete(tracker.variableStack, symbol)
	for _, reference := range tracker.ctx.Refs.References(symbol) {
		if !ast.IsWriteOnlyAccess(reference) {
			tracker.TrackExpression(reference, value)
		}
	}
}

func (tracker *Tracker) trackGlobalVariable(name string, value *Trace) {
	if tracker.globalStack[name] {
		return
	}
	tracker.globalStack[name] = true
	defer delete(tracker.globalStack, name)
	tracker.names.ForEachReferenceByName(name, nil, func(reference *ast.Node) bool {
		if !ast.IsWriteOnlyAccess(reference) && tracker.isGlobalReference(reference) {
			tracker.TrackExpression(reference, value)
		}
		return false
	})
}

func (tracker *Tracker) accessExpressionStaticName(node *ast.Node) (string, bool) {
	if node.Kind == ast.KindElementAccessExpression {
		return tracker.propertyNames().EvalAccessExpressionName(node)
	}
	return utils.AccessExpressionStaticName(node)
}

func (tracker *Tracker) staticPropertyName(node *ast.Node) (string, bool) {
	if node != nil && (node.Kind == ast.KindComputedPropertyName || node.Kind == ast.KindNoSubstitutionTemplateLiteral) {
		return tracker.propertyNames().EvalPropertyName(node)
	}
	return utils.GetStaticPropertyName(node)
}

func (tracker *Tracker) propertyNames() *utils.StaticStringEvaluator {
	if tracker.propertyEvaluator == nil {
		tracker.propertyEvaluator = utils.NewStaticStringEvaluatorWithoutScope()
	}
	return tracker.propertyEvaluator
}

func (tracker *Tracker) isGlobalReference(identifier *ast.Node) bool {
	if tracker.ctx.Refs != nil {
		return tracker.ctx.Refs.IsGlobalReference(identifier)
	}
	return !utils.IsShadowed(identifier, identifier.Text())
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
