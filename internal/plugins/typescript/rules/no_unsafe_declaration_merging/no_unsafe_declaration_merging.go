package no_unsafe_declaration_merging

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func buildUnsafeMergingMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "unsafeMerging",
		Description: "Unsafe declaration merging between classes and interfaces.",
	}
}

// nearestLocalsContainer climbs from the given node's parent to the nearest
// ancestor that owns a Locals SymbolTable. Used only for the `export default
// class Foo {}` fallback to look up the module-scope binding via GetLocals —
// GetLocals requires a LocalsContainer, whereas utils.FindEnclosingScope
// (used for the same-scope merge check) returns ModuleBlock / SourceFile /
// function-like nodes which is the right granularity for scope comparison
// but not always a LocalsContainer.
func nearestLocalsContainer(node *ast.Node) *ast.Node {
	if node == nil {
		return nil
	}
	p := node.Parent
	for p != nil && !ast.IsLocalsContainer(p) {
		p = p.Parent
	}
	return p
}

var NoUnsafeDeclarationMergingRule = rule.CreateRule(rule.Rule{
	Name:             "no-unsafe-declaration-merging",
	RequiresTypeInfo: true,
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		// reportIfUnsafeMerge mirrors upstream's scope-manager probe in
		// behavior while using tsgo's TypeChecker as the symbol source.
		//
		// Upstream does two things that look like implementation details
		// but are part of the rule's observable contract; we reproduce
		// both:
		//
		//   1. The interface listener calls sourceCode.getScope(node),
		//      which for a generic interface returns the type-parameter
		//      scope. That scope does not contain the interface's own
		//      name, so the lookup fails and the listener returns
		//      without reporting. ESLint thus skips the interface side
		//      whenever the interface has type parameters; the class
		//      listener uses .upper and is not affected.
		//
		//   2. Both listeners look up the name in a single lexical scope.
		//      A class and interface that share a name only count as
		//      "merging" if they live in the same enclosing scope (same
		//      SourceFile, same namespace block, same `declare module`
		//      block, …). Cross-`declare module` blocks and external-
		//      module class augmentation therefore never fire upstream,
		//      even though TypeScript really does merge the symbols.
		//
		// We keep TypeChecker.GetSymbolAtLocation as the source of truth
		// for "is there a merge at all", then filter symbol.Declarations
		// down to those sharing the same enclosing LocalsContainer as
		// the current declaration. The filter is what implements (2);
		// the early return on type parameters implements (1).
		reportIfUnsafeMerge := func(declNode *ast.Node, name *ast.Node, unsafeKind ast.Kind) {
			if name == nil {
				return
			}

			if declNode.Kind == ast.KindInterfaceDeclaration {
				iface := declNode.AsInterfaceDeclaration()
				if iface.TypeParameters != nil && len(iface.TypeParameters.Nodes) > 0 {
					return
				}
			}

			symbol := ctx.TypeChecker.GetSymbolAtLocation(name)
			// `export default class Foo {}` binds the class to the synthetic
			// __default symbol; the module-scope `Foo` local (which is what
			// merges with `interface Foo {}`) is a separate symbol in the
			// enclosing LocalsContainer. Recover it before the merge check.
			if symbol != nil && symbol.Name == ast.InternalSymbolNameDefault {
				if container := nearestLocalsContainer(declNode); container != nil {
					if local := ast.GetLocals(container)[name.Text()]; local != nil {
						symbol = local
					}
				}
			}
			if symbol == nil || len(symbol.Declarations) <= 1 {
				return
			}

			scope := utils.FindEnclosingScope(declNode)
			for _, decl := range symbol.Declarations {
				if decl == declNode || decl.Kind != unsafeKind {
					continue
				}
				if utils.FindEnclosingScope(decl) == scope {
					ctx.ReportNode(name, buildUnsafeMergingMessage())
					return
				}
			}
		}

		return rule.RuleListeners{
			ast.KindClassDeclaration: func(node *ast.Node) {
				reportIfUnsafeMerge(node, node.AsClassDeclaration().Name(), ast.KindInterfaceDeclaration)
			},
			ast.KindInterfaceDeclaration: func(node *ast.Node) {
				reportIfUnsafeMerge(node, node.AsInterfaceDeclaration().Name(), ast.KindClassDeclaration)
			},
		}
	},
})
