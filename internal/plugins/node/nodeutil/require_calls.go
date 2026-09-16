package nodeutil

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

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
				tracker.trackExpression(call, moduleValue)
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
			tracker.trackIdentifier(binding, value)
		}
	}
	return reads
}
