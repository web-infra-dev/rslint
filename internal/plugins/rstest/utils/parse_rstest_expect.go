package utils

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	testFramework "github.com/web-infra-dev/rslint/internal/utils/test_framework"
)

func IsRstestExpectCall(
	node *ast.Node,
	ctx rule.RuleContext,
	callbacks RstestTestCallbacks,
) bool {
	if node == nil || node.Kind != ast.KindCallExpression || findTopMostCallExpression(node) != node {
		return false
	}

	if isImportMetaRstestExpectCall(node) {
		return true
	}

	entries := testFramework.GetMemberEntries(node)
	if len(entries) == 0 || entries[0].Node == nil || entries[0].Node.Kind != ast.KindIdentifier {
		return false
	}

	root := entries[0].Node
	if isContextExpectRoot(root, entries, ctx, callbacks) {
		return true
	}
	return isImportedOrGlobalExpectRoot(root, entries, ctx)
}

// findTopMostCallExpression walks up the call/member chain that node is the
// head of. Ascending only continues while node stays on the callee side of its
// parent: a call in argument or computed-key position heads its own chain.
func findTopMostCallExpression(node *ast.Node) *ast.Node {
	top := node
	current := node
	for parent := current.Parent; parent != nil; {
		switch parent.Kind {
		case ast.KindParenthesizedExpression:
		case ast.KindCallExpression:
			if parent.AsCallExpression().Expression != current {
				return top
			}
			top = parent
		case ast.KindPropertyAccessExpression:
			if parent.AsPropertyAccessExpression().Expression != current {
				return top
			}
		case ast.KindElementAccessExpression:
			if parent.AsElementAccessExpression().Expression != current {
				return top
			}
		default:
			return top
		}
		current, parent = parent, parent.Parent
	}
	return top
}

func isContextExpectRoot(
	root *ast.Node,
	entries []testFramework.MemberEntry,
	ctx rule.RuleContext,
	callbacks RstestTestCallbacks,
) bool {
	if ctx.TypeChecker == nil {
		return false
	}
	symbol := ctx.TypeChecker.GetSymbolAtLocation(root)
	if symbol == nil {
		return false
	}
	if callbacks.ContextExpectNames[symbol] {
		return true
	}
	return callbacks.ContextReceivers[symbol] &&
		len(entries) > 1 &&
		entries[1].Name == "expect"
}

func isImportedOrGlobalExpectRoot(
	root *ast.Node,
	entries []testFramework.MemberEntry,
	ctx rule.RuleContext,
) bool {
	localName := root.AsIdentifier().Text
	var symbol *ast.Symbol
	if ctx.TypeChecker != nil {
		symbol = ctx.TypeChecker.GetSymbolAtLocation(root)
	}

	if name, _, ok := resolveImportMetaRstestBinding(symbol); ok {
		return name == "expect"
	}
	if isImportMetaRstestObjectAlias(symbol) {
		return len(entries) > 1 && entries[1].Name == "expect"
	}

	for _, module := range []string{RstestImportModule, RstestPlaywrightImportModule} {
		name, _, _ := testFramework.ResolveFunctionIdentifierReferenceFromSymbol(
			localName,
			root,
			symbol,
			ctx.SourceFile,
			module,
		)
		if name == "expect" {
			return true
		}
		if testFramework.IsModuleNamespaceSymbol(symbol, module) &&
			len(entries) > 1 &&
			entries[1].Name == "expect" {
			return true
		}
	}
	return false
}

func isImportMetaRstestExpectCall(node *ast.Node) bool {
	call := node.AsCallExpression()
	if call == nil {
		return false
	}
	_, parts, _, ok := parseImportMetaRstestChain(call.Expression)
	return ok && len(parts) > 0 && parts[0].name == "expect"
}

func isImportMetaRstestObjectAlias(symbol *ast.Symbol) bool {
	if symbol == nil {
		return false
	}
	for _, declaration := range symbol.Declarations {
		if declaration == nil || declaration.Kind != ast.KindVariableDeclaration {
			continue
		}
		initializer := declaration.AsVariableDeclaration().Initializer
		if isImportMetaRstest(initializer) {
			return true
		}
	}
	return false
}
