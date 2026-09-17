package nodeutil

import (
	"cmp"
	"slices"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// MayHaveRequire uses the parser's decoded identifiers to skip files that cannot
// reference require, including through a computed property on a global object.
func MayHaveRequire(file *ast.SourceFile) bool {
	for _, name := range []string{"require", "global", "globalThis", "self", "window"} {
		if file.HasIdentifier(name) {
			return true
		}
	}
	return false
}

// CollectRequireCalls returns require() and require.resolve() calls, following
// aliases and destructuring. Separate alias paths can return duplicate calls.
func CollectRequireCalls(ctx rule.RuleContext) []*ast.Node {
	var calls []*ast.Node
	collect := func(node *ast.Node) { calls = append(calls, node) }
	NewReferenceTracker(ctx).TrackGlobals(map[string]*ReferenceTrace{
		"require": {Call: collect, Properties: map[string]*ReferenceTrace{
			"resolve": {Call: collect},
		}},
	})
	return calls
}

// RequireTarget pairs a constant module name with the argument to report on.
type RequireTarget struct {
	Node *ast.Node
	Name string
}

// CollectRequireTargets keeps calls in source order and removes loader parameters.
// Like upstream getStringIfConstant without a scope, it does not follow argument identifiers.
func CollectRequireTargets(ctx rule.RuleContext) []RequireTarget {
	calls := CollectRequireCalls(ctx)
	slices.SortStableFunc(calls, func(a, b *ast.Node) int { return cmp.Compare(a.Pos(), b.Pos()) })
	var targets []RequireTarget
	var evaluator *utils.StaticStringEvaluator
	for _, call := range calls {
		args := call.Arguments()
		if len(args) == 0 {
			continue
		}
		source := utils.ESTreeRuntimeExpression(args[0])
		name, ok := utils.GetStaticExpressionValue(utils.SkipAssertionsAndParens(source))
		if !ok {
			if evaluator == nil {
				evaluator = utils.NewStaticStringEvaluatorWithoutScope()
			}
			name, ok = evaluator.EvalToString(source)
		}
		if ok {
			name, _, _ = strings.Cut(name, "!")
			targets = append(targets, RequireTarget{Node: source, Name: name})
		}
	}
	return targets
}

// CollectModulePropertyReads follows CommonJS module-object aliases to reads
// of one property. ESM imports use upstream's strict CJS mode: default imports
// expose the object; namespace imports expose it as .default. Module names
// match exactly, and destructuring/property-value aliases are not read reports.
func CollectModulePropertyReads(ctx rule.RuleContext, moduleName, propertyName string) map[*ast.Node]bool {
	reads := make(map[*ast.Node]bool)
	moduleValue := &ReferenceTrace{Properties: map[string]*ReferenceTrace{
		propertyName: {Read: func(node *ast.Node) {
			if ast.IsAccessExpression(node) {
				reads[node] = true
			}
		}},
	}}
	tracker := NewReferenceTracker(ctx)
	tracker.TrackGlobals(map[string]*ReferenceTrace{"require": {Call: func(call *ast.Node) {
		args := call.Arguments()
		if len(args) > 0 {
			if name, ok := tracker.constantString(args[0]); ok && name == moduleName {
				// Recurse while require aliases remain on the cycle guard stack.
				tracker.TrackExpression(call, moduleValue)
			}
		}
	}}})
	namespace := &ReferenceTrace{Properties: map[string]*ReferenceTrace{"default": moduleValue}}
	for _, statement := range ctx.SourceFile.Statements.Nodes {
		if statement.Kind != ast.KindImportDeclaration {
			continue
		}
		source := ast.GetExternalModuleName(statement)
		if source == nil || source.Text() != moduleName {
			continue
		}
		for _, binding := range utils.GetImportBindingNodes(statement) {
			value := moduleValue
			switch binding.Parent.Kind {
			case ast.KindNamespaceImport:
				value = namespace
			case ast.KindImportSpecifier:
				name := binding.Parent.PropertyName()
				if name == nil {
					name = binding
				}
				if name.Text() != "default" {
					continue
				}
			}
			tracker.TrackBinding(binding, value)
		}
	}
	return reads
}
