package nodeutil

import (
	"maps"
	"slices"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// ReferenceTrace describes the reads, calls and constructors a Node rule needs.
// Callbacks own rule policy; traversal only follows static properties and bindings.
type ReferenceTrace struct {
	Properties            map[string]*ReferenceTrace
	Read, Call, Construct func(*ast.Node)
}

func (trace *ReferenceTrace) read(node *ast.Node) {
	if trace.Read != nil {
		trace.Read(node)
	}
}

// ReferenceTracker shares alias and destructuring traversal between Node rules.
type ReferenceTracker struct {
	ctx               rule.RuleContext
	names             *utils.ReferenceIndex
	propertyEvaluator *utils.StaticStringEvaluator
	variableStack     map[*ast.Symbol]bool
	globalStack       map[string]bool
}

type referenceNamesKey struct{}

// NewReferenceTracker follows Node API aliases using existing name and symbol indexes.
// The linter supplies ctx.Refs; the name index is shared across Node rules per file.
func NewReferenceTracker(ctx rule.RuleContext) *ReferenceTracker {
	names := rule.CachedByFile(ctx, referenceNamesKey{}, func() *utils.ReferenceIndex {
		return utils.NewReferenceIndex(ctx.SourceFile, nil)
	})
	return &ReferenceTracker{ctx: ctx, names: names,
		variableStack: make(map[*ast.Symbol]bool), globalStack: make(map[string]bool)}
}

// TrackGlobals follows unmodified configured globals and their global-object properties.
func (tracker *ReferenceTracker) TrackGlobals(globals map[string]*ReferenceTrace) {
	for _, name := range slices.Sorted(maps.Keys(globals)) {
		tracker.trackGlobalRoot(name, globals[name], true)
	}
	for _, name := range []string{"global", "globalThis", "self", "window"} {
		tracker.trackGlobalRoot(name, &ReferenceTrace{Properties: globals}, false)
	}
}

func (tracker *ReferenceTracker) trackGlobalRoot(name string, value *ReferenceTrace, report bool) {
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
		if report {
			value.read(reference)
		}
		tracker.trackExpression(reference, value)
	}
}

func (tracker *ReferenceTracker) globalReferences(name string) []*ast.Node {
	var references []*ast.Node
	tracker.names.ForEachReferenceByName(name, nil, func(identifier *ast.Node) bool {
		if tracker.ctx.Refs.IsGlobalReference(identifier) {
			references = append(references, identifier)
		}
		return false
	})
	return references
}

func (tracker *ReferenceTracker) trackExpression(node *ast.Node, value *ReferenceTrace) {
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
		if utils.IsInJsxTagName(parent) {
			return
		}
		name, ok := tracker.accessExpressionStaticName(parent)
		if next := value.Properties[name]; ok && next != nil {
			next.read(parent)
			tracker.trackExpression(parent, next)
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

func (tracker *ReferenceTracker) trackAssignmentTarget(node *ast.Node, value *ReferenceTrace) {
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
				tracker.trackAssignmentTarget(ast.GetTargetOfBindingOrAssignmentElement(element), value.Properties[name])
			}
		}
	case ast.KindBinaryExpression:
		binary := node.AsBinaryExpression()
		if binary != nil && binary.OperatorToken != nil && binary.OperatorToken.Kind == ast.KindEqualsToken {
			tracker.trackAssignmentTarget(binary.Left, value)
		}
	}
}

func (tracker *ReferenceTracker) trackIdentifier(identifier *ast.Node, value *ReferenceTrace) {
	if symbol := tracker.ctx.Refs.ResolveInFile(identifier); symbol != nil {
		tracker.trackVariable(symbol, value)
		return
	}
	// Declaration names are not reference positions; use their binder symbol.
	if symbol := utils.BindingNameSymbol(identifier); symbol != nil {
		tracker.trackVariable(symbol, value)
		return
	}
	name := identifier.AsIdentifier().Text
	if tracker.ctx.Globals.Access(name).IsDeclared() && tracker.ctx.Refs.IsGlobalReference(identifier) {
		tracker.trackGlobalVariable(name, value)
	}
}

func (tracker *ReferenceTracker) trackVariable(symbol *ast.Symbol, value *ReferenceTrace) {
	if symbol == nil || tracker.variableStack[symbol] {
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

func (tracker *ReferenceTracker) trackGlobalVariable(name string, value *ReferenceTrace) {
	if tracker.globalStack[name] {
		return
	}
	tracker.globalStack[name] = true
	defer delete(tracker.globalStack, name)
	tracker.names.ForEachReferenceByName(name, nil, func(reference *ast.Node) bool {
		if !ast.IsWriteOnlyAccess(reference) && tracker.ctx.Refs.IsGlobalReference(reference) {
			tracker.trackExpression(reference, value)
		}
		return false
	})
}

func (tracker *ReferenceTracker) accessExpressionStaticName(node *ast.Node) (string, bool) {
	if node.Kind == ast.KindElementAccessExpression {
		argument := node.AsElementAccessExpression().ArgumentExpression
		return tracker.constantString(argument)
	}
	return utils.AccessExpressionStaticName(node)
}

func (tracker *ReferenceTracker) staticPropertyName(node *ast.Node) (string, bool) {
	if node != nil && node.Kind == ast.KindComputedPropertyName {
		return tracker.constantString(node.AsComputedPropertyName().Expression)
	}
	// tsgo's binding-property helper unwraps computed template literal keys.
	if node != nil && node.Kind == ast.KindNoSubstitutionTemplateLiteral {
		return utils.GetStaticExpressionValue(node)
	}
	return utils.GetStaticPropertyName(node)
}

func (tracker *ReferenceTracker) constantString(node *ast.Node) (string, bool) {
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

// TrackModules follows CommonJS, process.getBuiltinModule and legacy ESM imports.
// Builtin node: aliases use the same trace, retaining canonical callback names.
func (tracker *ReferenceTracker) TrackModules(modules map[string]*ReferenceTrace) {
	lookup := func(name string) *ReferenceTrace {
		if strings.HasPrefix(name, "node:") {
			if !isNodeBuiltin(name) {
				return nil
			}
			name = strings.TrimPrefix(name, "node:")
		}
		return modules[name]
	}
	load := func(node *ast.Node) {
		args := node.AsCallExpression().Arguments
		if args == nil || len(args.Nodes) == 0 {
			return
		}
		name, ok := tracker.constantString(args.Nodes[0])
		if trace := lookup(name); ok && trace != nil {
			trace.read(node)
			tracker.trackExpression(node, trace)
		}
	}
	tracker.TrackGlobals(map[string]*ReferenceTrace{"require": {Call: load}})
	tracker.TrackGlobals(map[string]*ReferenceTrace{"process": {Properties: map[string]*ReferenceTrace{
		"getBuiltinModule": {Call: load},
	}}})
	for _, node := range tracker.ctx.SourceFile.Statements.Nodes {
		if node.Kind != ast.KindImportDeclaration && node.Kind != ast.KindExportDeclaration {
			continue
		}
		source := ast.GetExternalModuleName(node)
		if source == nil || source.Kind != ast.KindStringLiteral {
			continue
		}
		trace := lookup(source.Text())
		if trace == nil {
			continue
		}
		trace.read(node)
		// In legacy mode a namespace's default property aliases the CommonJS
		// export. Reading that alias must not report the module a second time.
		moduleValue := &ReferenceTrace{Properties: trace.Properties, Call: trace.Call, Construct: trace.Construct}
		properties := maps.Clone(trace.Properties)
		if properties == nil {
			properties = map[string]*ReferenceTrace{}
		}
		properties["default"] = moduleValue
		if node.Kind == ast.KindImportDeclaration {
			for _, binding := range utils.GetImportBindingNodes(node) {
				value := moduleValue
				switch binding.Parent.Kind {
				case ast.KindNamespaceImport:
					value = &ReferenceTrace{Properties: properties}
				case ast.KindImportSpecifier:
					imported := binding.Parent.PropertyName()
					if imported == nil {
						imported = binding
					}
					value = properties[imported.Text()]
					if value == nil {
						continue
					}
					value.read(binding.Parent)
				}
				tracker.trackIdentifier(binding, value)
			}
		} else {
			clause := node.AsExportDeclaration().ExportClause
			if clause == nil || clause.Kind == ast.KindNamespaceExport {
				for _, name := range slices.Sorted(maps.Keys(trace.Properties)) {
					trace.Properties[name].read(node)
				}
			} else {
				for _, specifier := range clause.AsNamedExports().Elements.Nodes {
					local := specifier.AsExportSpecifier().PropertyName
					if local == nil {
						local = specifier.Name()
					}
					if next := properties[local.Text()]; next != nil {
						next.read(specifier)
					}
				}
			}
		}
	}
}
