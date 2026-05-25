package no_unnecessary_qualifier

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// NoUnnecessaryQualifierRule ports @typescript-eslint/no-unnecessary-qualifier.
//
// The rule flags namespace or enum qualifiers that resolve to a symbol
// already in scope at the access site. The walk maintains a stack of
// enclosing enum / namespace declaration nodes; whenever a
// PropertyAccessExpression or QualifiedName whose qualifier resolves to a
// symbol on that stack is encountered, the rule re-resolves the right-hand
// name from the qualifier's location and reports if the same symbol is
// reachable there directly.
//
// Upstream uses a single `currentFailedNamespaceExpression` cursor to avoid
// nested reports inside a chain whose outer qualifier already fired
// (e.g. `A.B.C.D` should report once on `A.B.C`, not also on `A.B`); the
// cursor is cleared on the matching node's exit. We mirror that exactly.
//
// References:
//   - Rule:      https://typescript-eslint.io/rules/no-unnecessary-qualifier
//   - Upstream:  https://github.com/typescript-eslint/typescript-eslint/blob/main/packages/eslint-plugin/src/rules/no-unnecessary-qualifier.ts
var NoUnnecessaryQualifierRule = rule.CreateRule(rule.Rule{
	Name:             "no-unnecessary-qualifier",
	RequiresTypeInfo: true,
	Run:              run,
})

func run(ctx rule.RuleContext, options any) rule.RuleListeners {
	// namespacesInScope tracks the enum / namespace declaration nodes whose
	// bodies enclose the current traversal point. The qualifier check
	// resolves a symbol and verifies one of its declarations is on this
	// stack — same identity comparison upstream does.
	var namespacesInScope []*ast.Node
	// currentFailedNamespaceExpression suppresses nested chain reports.
	// When `A.B.C.D` fires on the outer access, the inner `A.B.C` would
	// otherwise also report; the cursor blocks that, then resets on the
	// outer node's exit so sibling chains still get visited.
	var currentFailedNamespaceExpression *ast.Node

	sf := ctx.SourceFile

	tryGetAliasedSymbol := func(symbol *ast.Symbol) *ast.Symbol {
		if symbol == nil || symbol.Flags&ast.SymbolFlagsAlias == 0 {
			return nil
		}
		return ctx.TypeChecker.GetAliasedSymbol(symbol)
	}

	var symbolIsNamespaceInScope func(symbol *ast.Symbol) bool
	symbolIsNamespaceInScope = func(symbol *ast.Symbol) bool {
		if symbol == nil {
			return false
		}
		for _, decl := range symbol.Declarations {
			for _, ns := range namespacesInScope {
				if decl == ns {
					return true
				}
			}
		}
		alias := tryGetAliasedSymbol(symbol)
		if alias == nil {
			return false
		}
		return symbolIsNamespaceInScope(alias)
	}

	getSymbolInScope := func(location *ast.Node, flags ast.SymbolFlags, name string) *ast.Symbol {
		for _, sym := range ctx.TypeChecker.GetSymbolsInScope(location, flags) {
			if sym.Name == name {
				return sym
			}
		}
		return nil
	}

	symbolsAreEqual := func(accessed *ast.Symbol, inScope *ast.Symbol) bool {
		if accessed == nil || inScope == nil {
			return false
		}
		return accessed == ctx.TypeChecker.GetExportSymbolOfSymbol(inScope)
	}

	qualifierIsUnnecessary := func(qualifier *ast.Node, name *ast.Node) bool {
		// Mirror upstream verbatim: resolve the qualifier and the accessed
		// name via the type checker, then ensure the accessed symbol is
		// reachable in scope at the qualifier's position and matches what
		// the qualifier resolves to. tsgo's GetSymbolAtLocation handles
		// Identifier / QualifiedName / PropertyAccessExpression natively
		// (see typescript-go/internal/checker/checker.go getSymbolAtLocation
		// switch on KindPropertyAccessExpression / KindQualifiedName), so we
		// pass the qualifier node as-is — same as upstream's
		// `services.getSymbolAtLocation(qualifier)`.
		namespaceSymbol := ctx.TypeChecker.GetSymbolAtLocation(qualifier)
		if namespaceSymbol == nil || !symbolIsNamespaceInScope(namespaceSymbol) {
			return false
		}
		accessedSymbol := ctx.TypeChecker.GetSymbolAtLocation(name)
		if accessedSymbol == nil {
			return false
		}
		fromScope := getSymbolInScope(qualifier, accessedSymbol.Flags, name.Text())
		if fromScope == nil {
			return false
		}
		return symbolsAreEqual(accessedSymbol, fromScope)
	}

	visitNamespaceAccess := func(node *ast.Node, qualifier *ast.Node, name *ast.Node) {
		if currentFailedNamespaceExpression != nil {
			return
		}
		if !qualifierIsUnnecessary(qualifier, name) {
			return
		}
		currentFailedNamespaceExpression = node
		qualifierStart := scanner.GetTokenPosOfNode(qualifier, sf, false)
		nameStart := scanner.GetTokenPosOfNode(name, sf, false)
		ctx.ReportNodeWithFixes(qualifier, rule.RuleMessage{
			Id:          "unnecessaryQualifier",
			Description: "Qualifier is unnecessary since '" + name.Text() + "' is in scope.",
			Data:        map[string]string{"name": name.Text()},
		}, rule.RuleFixRemoveRange(core.NewTextRange(qualifierStart, nameStart)))
	}

	resetIfMatchingCursor := func(node *ast.Node) {
		if node == currentFailedNamespaceExpression {
			currentFailedNamespaceExpression = nil
		}
	}

	pushNamespace := func(node *ast.Node) {
		namespacesInScope = append(namespacesInScope, node)
	}
	popNamespace := func(*ast.Node) {
		if n := len(namespacesInScope); n > 0 {
			namespacesInScope = namespacesInScope[:n-1]
		}
	}

	return rule.RuleListeners{
		// Enum declarations always have a body; they are always pushed.
		ast.KindEnumDeclaration:                      pushNamespace,
		rule.ListenerOnExit(ast.KindEnumDeclaration): popNamespace,

		// Namespace / module declarations: only the INNERMOST level of a
		// dotted name (`namespace A.B.C { ... }`) holds the actual block
		// body. tsgo nests outer levels as `ModuleDeclaration { Body:
		// ModuleDeclaration { ... } }`, so we push when (and only when) we
		// enter the ModuleBlock — whose parent is the innermost
		// ModuleDeclaration. This matches upstream's
		// `TSModuleDeclaration > TSModuleBlock` push, with a guarded
		// matching pop instead of the unconditional one upstream uses
		// (which would underflow on outer levels of a dotted name; the
		// JS `.pop()` happens to no-op on empty, ours would too but the
		// guard makes the intent explicit).
		ast.KindModuleBlock: func(node *ast.Node) {
			if node.Parent != nil && node.Parent.Kind == ast.KindModuleDeclaration {
				pushNamespace(node.Parent)
			}
		},
		rule.ListenerOnExit(ast.KindModuleBlock): func(node *ast.Node) {
			if node.Parent != nil && node.Parent.Kind == ast.KindModuleDeclaration {
				popNamespace(node)
			}
		},

		// Value-position chain: `A.B`, `A.B.C`, ...
		ast.KindPropertyAccessExpression: func(node *ast.Node) {
			pae := node.AsPropertyAccessExpression()
			if pae.Expression == nil {
				return
			}
			nameNode := pae.Name()
			// PrivateIdentifier (`#x`) is a member-name kind that never
			// participates in namespace / enum access — skip to match
			// upstream's `node.property as TSESTree.Identifier` narrowing.
			if nameNode == nil || nameNode.Kind != ast.KindIdentifier {
				return
			}
			// Mirror upstream's `isEntityNameExpression(node.object)`
			// gate — Identifier OR PAE-with-identifier-receiver chain.
			// tsgo's helper has the same semantics (Identifier ||
			// IsPropertyAccessExpression && IsIdentifier(Name) &&
			// IsEntityNameExpression(Expression)) including the same
			// "no parens skip" behavior we depend on for the documented
			// `(A).B` divergence.
			if !ast.IsEntityNameExpression(pae.Expression) {
				return
			}
			visitNamespaceAccess(node, pae.Expression, nameNode)
		},
		rule.ListenerOnExit(ast.KindPropertyAccessExpression): resetIfMatchingCursor,

		// Type-position chain: `A.B`, `A.B.C`, ...
		ast.KindQualifiedName: func(node *ast.Node) {
			qn := node.AsQualifiedName()
			if qn.Left == nil || qn.Right == nil {
				return
			}
			visitNamespaceAccess(node, qn.Left, qn.Right)
		},
		rule.ListenerOnExit(ast.KindQualifiedName): resetIfMatchingCursor,
	}
}

