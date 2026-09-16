package nodeutil

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// CollectUnmodifiedGlobalReferences returns direct identifier references to
// configured globals with no local definition or write, as ReferenceTracker
// does. Reads through aliases and global-object properties are not identifiers
// referring directly to these globals.
func CollectUnmodifiedGlobalReferences(ctx rule.RuleContext, names ...string) []*ast.Node {
	tracker := newReferenceTracker(ctx)
	var result []*ast.Node
	for _, name := range names {
		result = append(result, tracker.unmodifiedGlobalReferences(name)...)
	}
	return result
}

// CollectModulePropertyReads follows CommonJS module-object aliases to reads
// of one property. ESM imports use ReferenceTracker's default strict CJS mode:
// default imports expose the object; namespace imports expose it as .default.
// Module names match exactly, and property-value aliases are not read reports.
func CollectModulePropertyReads(ctx rule.RuleContext, moduleName, propertyName string) map[*ast.Node]bool {
	reads := map[*ast.Node]bool{}
	value := &referenceTrace{properties: map[string]*referenceTrace{
		propertyName: {read: func(node *ast.Node) { reads[node] = true }},
	}}
	tracker := newReferenceTracker(ctx)
	for _, call := range tracker.collectRequireCalls(false) {
		args := call.Arguments()
		if len(args) == 0 {
			continue
		}
		if name, ok := tracker.constantString(args[0]); ok && name == moduleName {
			tracker.trackExpression(call, value)
		}
	}
	for _, statement := range ctx.SourceFile.Statements.Nodes {
		if statement.Kind != ast.KindImportDeclaration {
			continue
		}
		declaration := statement.AsImportDeclaration()
		if declaration.ModuleSpecifier == nil || declaration.ModuleSpecifier.Text() != moduleName {
			continue
		}
		for _, binding := range utils.GetImportBindingNodes(statement) {
			imported := value
			switch binding.Parent.Kind {
			case ast.KindNamespaceImport:
				imported = &referenceTrace{properties: map[string]*referenceTrace{"default": value}}
			case ast.KindImportSpecifier:
				name := binding.Parent.PropertyName()
				if name == nil {
					name = binding
				}
				if name.Text() != "default" {
					continue
				}
			}
			tracker.trackIdentifier(binding, imported)
		}
	}
	return reads
}
