package utils

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	internalUtils "github.com/web-infra-dev/rslint/internal/utils"
	testFramework "github.com/web-infra-dev/rslint/internal/utils/test_framework"
)

// Rstest ships Chai's `assert` as a global alongside `expect`
// (packages/core/src/utils/constants.ts globalApiList), so a rule that treats
// `assert` as an assertion has to resolve it the way it resolves `expect`.
// Matching the callee text alone recognizes only the literal spelling, which
// misses `assert as check`, `rstest.assert` and `const { assert: check } =
// import.meta.rstest`.
//
// `assert` is resolved separately from `expect` rather than by generalizing the
// expect path over an API name. The expect path is shared by every Rstest rule
// and carries its own caches, matcher-chain parsing and test-context handling;
// `assert` needs none of that, because a Chai assertion is a single call with
// no chain to walk.
const rstestAssertAPIName = "assert"

// IsAssertCall reports whether node is a call on Rstest's `assert`, through any
// binding the analysis can resolve: the global, a named import, an import
// alias, a namespace member, or a destructured `import.meta.rstest` key.
//
// The test context is deliberately not consulted. Rstest hands a callback its
// own `expect`, not its own `assert`, so `ctx.assert` names something the
// framework did not provide.
func (analysis *RstestCallAnalysis) IsAssertCall(node *ast.Node) bool {
	if node == nil || node.Kind != ast.KindCallExpression {
		return false
	}
	if !analysis.isAssertCandidate(node) {
		return false
	}
	if result, ok := analysis.isAssert[node]; ok {
		return result
	}
	result := analysis.resolveAssertCall(node)
	analysis.isAssert[node] = result
	return result
}

func (analysis *RstestCallAnalysis) isAssertCandidate(node *ast.Node) bool {
	root := testFramework.ResolveFirstIdentifier(node.AsCallExpression().Expression)
	return root == nil ||
		root.Kind != ast.KindIdentifier ||
		analysis.candidates[root.AsIdentifier().Text]&rstestCandidateAssert != 0
}

func (analysis *RstestCallAnalysis) resolveAssertCall(node *ast.Node) bool {
	// `import.meta.rstest.assert.equal(1, 1)` has no identifier root, so it has
	// to be recognized before the identifier path, the way the expect resolver
	// recognizes `import.meta.rstest.expect`.
	if isImportMetaRstestAssertCall(node) {
		return true
	}
	entries := firstRstestExpectEntries(node.AsCallExpression().Expression)
	if entries.count == 0 || entries.first == nil || entries.first.Kind != ast.KindIdentifier {
		return false
	}
	root := entries.first
	localName := root.AsIdentifier().Text
	symbol := resolveRstestRootSymbol(analysis.ctx, root)
	if symbol == nil {
		// An unresolved root is the injected global. A file that assigns to it
		// has replaced the framework's assert with something the analysis
		// cannot follow.
		return localName == rstestAssertAPIName && !analysis.globalAssertWritten
	}
	rootKind, ok := analysis.assertRoots[symbol]
	if !ok {
		rootKind = classifyRstestAssertRoot(localName, root, symbol, analysis)
		analysis.assertRoots[symbol] = rootKind
	}
	switch rootKind {
	case rstestExpectRootDirect:
		return true
	case rstestExpectRootReceiver:
		return entries.count > 1 && entries.secondName == rstestAssertAPIName
	default:
		return false
	}
}

// classifyRstestAssertRoot answers what the root identifier of a call chain
// binds: Rstest's `assert` itself, an object that carries it, or neither.
//
// localName and root are the identifier as it is written at the call site, and
// both have to be the real ones. The module resolver returns the name it was
// given whenever it has no identifier to inspect, so handing it a placeholder
// makes every symbol answer "assert" and a local binding that shadows the
// import is then mistaken for the framework's own.
func classifyRstestAssertRoot(
	localName string,
	root *ast.Node,
	symbol *ast.Symbol,
	analysis *RstestCallAnalysis,
) rstestExpectRootKind {
	ctx := analysis.ctx
	if ctx.Refs != nil {
		for _, reference := range ctx.Refs.References(symbol) {
			if internalUtils.IsWriteReference(reference) {
				return rstestExpectRootNone
			}
		}
	}
	if name, _, ok := resolveImportMetaRstestBinding(symbol); ok {
		if name == rstestAssertAPIName {
			return rstestExpectRootDirect
		}
		return rstestExpectRootNone
	}
	if isImportMetaRstestObjectAlias(symbol) {
		return rstestExpectRootReceiver
	}
	if testFramework.IsModuleNamespaceSymbolModules(symbol, RstestAllImportModules) {
		return rstestExpectRootReceiver
	}
	name, _, _ := testFramework.ResolveFunctionIdentifierReferenceFromSymbolModules(
		localName,
		root,
		symbol,
		ctx.SourceFile,
		RstestAllImportModules,
	)
	if name == rstestAssertAPIName {
		return rstestExpectRootDirect
	}
	return rstestExpectRootNone
}

// isImportMetaRstestAssertCall reports whether node calls through
// `import.meta.rstest.assert`, the form that names the API without ever
// binding it to an identifier.
func isImportMetaRstestAssertCall(node *ast.Node) bool {
	call := node.AsCallExpression()
	if call == nil {
		return false
	}
	_, parts, _, ok := parseImportMetaRstestChain(call.Expression)
	return ok && len(parts) > 0 && parts[0].name == rstestAssertAPIName
}
