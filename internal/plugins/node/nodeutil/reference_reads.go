package nodeutil

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
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
	for _, call := range collectRequireCalls(ctx, false) {
		args := call.AsCallExpression().Arguments
		if args == nil || len(args.Nodes) == 0 {
			continue
		}
		if name, ok := tracker.constantString(args.Nodes[0]); ok && name == moduleName {
			tracker.trackExpression(call, value)
		}
	}
	for _, statement := range ctx.SourceFile.Statements.Nodes {
		if statement.Kind != ast.KindImportDeclaration {
			continue
		}
		declaration := statement.AsImportDeclaration()
		if declaration.ModuleSpecifier == nil || declaration.ModuleSpecifier.Text() != moduleName || declaration.ImportClause == nil {
			continue
		}
		clause := declaration.ImportClause.AsImportClause()
		if clause.Name() != nil {
			tracker.trackIdentifier(clause.Name(), value)
		}
		bindings := clause.NamedBindings
		if bindings == nil {
			continue
		}
		if bindings.Kind == ast.KindNamespaceImport {
			tracker.trackIdentifier(bindings.Name(), &referenceTrace{properties: map[string]*referenceTrace{"default": value}})
			continue
		}
		for _, node := range bindings.AsNamedImports().Elements.Nodes {
			specifier := node.AsImportSpecifier()
			name := specifier.PropertyName
			if name == nil {
				name = specifier.Name()
			}
			if name.Text() == "default" {
				tracker.trackIdentifier(specifier.Name(), value)
			}
		}
	}
	return reads
}
