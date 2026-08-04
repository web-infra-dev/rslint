package no_unsafe_declaration_merging

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

type declarationKinds uint8

const (
	declarationKindClass declarationKinds = 1 << iota
	declarationKindInterface
	declarationScanCacheThreshold = 8
)

func declarationKind(kind ast.Kind) declarationKinds {
	switch kind {
	case ast.KindClassDeclaration:
		return declarationKindClass
	case ast.KindInterfaceDeclaration:
		return declarationKindInterface
	default:
		return 0
	}
}

var unsafeMergingMessage = rule.RuleMessage{
	Id:          "unsafeMerging",
	Description: "Unsafe declaration merging between classes and interfaces.",
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

type mergeRuleState struct {
	ctx                rule.RuleContext
	scopeKindsBySymbol map[*ast.Symbol]map[*ast.Node]declarationKinds
}

func (s *mergeRuleState) checkClass(node *ast.Node) {
	s.reportIfUnsafeMerge(node, node.AsClassDeclaration().Name(), ast.KindInterfaceDeclaration)
}

func (s *mergeRuleState) checkInterface(node *ast.Node) {
	s.reportIfUnsafeMerge(node, node.AsInterfaceDeclaration().Name(), ast.KindClassDeclaration)
}

// reportIfUnsafeMerge mirrors upstream's scope-manager probe in behavior
// while using tsgo's binder symbol as the binding source.
//
// Upstream does two things that look like implementation details but are part
// of the rule's observable contract; we reproduce both:
//
//  1. The interface listener calls sourceCode.getScope(node), which for a
//     generic interface returns the type-parameter scope. That scope does not
//     contain the interface's own name, so the lookup fails and the listener
//     returns without reporting. ESLint thus skips the interface side whenever
//     the interface has type parameters; the class listener uses .upper and is
//     not affected.
//  2. Both listeners look up the name in a single lexical scope. A class and
//     interface that share a name only count as "merging" if they live in the
//     same enclosing scope (same SourceFile, namespace block, `declare module`
//     block, etc.). Cross-`declare module` blocks and external-module class
//     augmentation therefore never fire upstream, even though TypeScript does
//     merge the symbols.
//
// The binder has already attached every same-binding declaration to the
// declaration node's symbol before lint traversal begins. Reading it directly
// avoids a checker lookup for every class and interface. We still filter
// symbol.Declarations by enclosing scope because export symbols can be shared
// by separate namespace or ambient-module blocks. The filter implements (2);
// the early return on type parameters implements (1).
func (s *mergeRuleState) reportIfUnsafeMerge(declNode *ast.Node, name *ast.Node, unsafeKind ast.Kind) {
	if name == nil {
		return
	}

	if declNode.Kind == ast.KindInterfaceDeclaration {
		iface := declNode.AsInterfaceDeclaration()
		if iface.TypeParameters != nil && len(iface.TypeParameters.Nodes) > 0 {
			return
		}
	}

	symbol := declNode.Symbol()
	// `export default class Foo {}` binds the class to the synthetic
	// __default symbol; the module-scope `Foo` local (which is what
	// merges with `interface Foo {}`) is a separate symbol in the
	// enclosing LocalsContainer. Prefer the binder's direct link to that
	// local, with the Locals lookup retained as a defensive fallback.
	if symbol != nil && symbol.Name == ast.InternalSymbolNameDefault {
		if local := declNode.LocalSymbol(); local != nil {
			symbol = local
		} else if container := nearestLocalsContainer(declNode); container != nil {
			if local := ast.GetLocals(container)[name.Text()]; local != nil {
				symbol = local
			}
		}
	}
	if symbol == nil || len(symbol.Declarations) <= 1 {
		return
	}

	scope := utils.FindEnclosingScope(declNode)
	hasUnsafeDeclaration := false
	if len(symbol.Declarations) <= declarationScanCacheThreshold {
		for _, declaration := range symbol.Declarations {
			if declaration.Kind == unsafeKind && utils.FindEnclosingScope(declaration) == scope {
				hasUnsafeDeclaration = true
				break
			}
		}
	} else {
		if s.scopeKindsBySymbol == nil {
			s.scopeKindsBySymbol = make(map[*ast.Symbol]map[*ast.Node]declarationKinds)
		}
		kindsByScope := s.scopeKindsBySymbol[symbol]
		if kindsByScope == nil {
			kindsByScope = make(map[*ast.Node]declarationKinds)
			for _, declaration := range symbol.Declarations {
				kind := declarationKind(declaration.Kind)
				if kind != 0 {
					declarationScope := utils.FindEnclosingScope(declaration)
					kindsByScope[declarationScope] |= kind
				}
			}
			s.scopeKindsBySymbol[symbol] = kindsByScope
		}
		hasUnsafeDeclaration = kindsByScope[scope]&declarationKind(unsafeKind) != 0
	}
	if hasUnsafeDeclaration {
		s.ctx.ReportNode(name, unsafeMergingMessage)
	}
}

var NoUnsafeDeclarationMergingRule = rule.CreateRule(rule.Rule{
	Name: "no-unsafe-declaration-merging",
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		// Declaration merges are almost always pairs. Keep that common path as
		// a tiny slice scan, and only build an index for unusually large merged
		// symbols so adversarial input cannot make every listener scan the same
		// declaration list repeatedly.
		state := mergeRuleState{ctx: ctx}
		return rule.RuleListeners{
			ast.KindClassDeclaration:     state.checkClass,
			ast.KindInterfaceDeclaration: state.checkInterface,
		}
	},
})
