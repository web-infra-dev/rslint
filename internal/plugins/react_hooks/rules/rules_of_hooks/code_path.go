package rules_of_hooks

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"

	"github.com/web-infra-dev/rslint/internal/utils/cfg"
)

// hookCodePath is the thin rules-of-hooks view over the shared CFG. The graph
// is built once per code path root and records only hook calls; all graph-shape
// reasoning remains in internal/utils/cfg.
type hookCodePath struct {
	root      *ast.Node
	paths     *cfg.PathAnalysis[struct{}]
	locations map[*ast.Node]*cfg.Block[struct{}]
}

type hookPathState struct {
	reachable           bool
	cyclic              bool
	onEveryFinalPath    bool
	possiblyEarlyReturn bool
}

func buildHookCodePath(root *ast.Node) *hookCodePath {
	locations := make(map[*ast.Node]*cfg.Block[struct{}])
	graph := cfg.Build(root, cfg.Hooks[struct{}]{
		Expression: func(builder *cfg.Builder[struct{}], node *ast.Node) {
			if node.Kind != ast.KindCallExpression {
				return
			}
			call := node.AsCallExpression()
			if call != nil && isHookCallee(call.Expression) {
				// A finally block is laid out once for normal flow and again for
				// abrupt flow. ESLint visits its AST once while the latter segment
				// is active, so the last layout is the observable location.
				locations[node] = builder.Current()
			}
		},
	})
	return &hookCodePath{
		root:      root,
		paths:     cfg.AnalyzePaths(graph),
		locations: locations,
	}
}

func (codePath *hookCodePath) state(hook *ast.Node) hookPathState {
	if codePath == nil || codePath.paths == nil {
		return hookPathState{}
	}
	location := codePath.locations[hook]
	state := hookPathState{}
	if location == nil {
		return state
	}
	shortestHookPath, reachesFromStart := codePath.paths.ShortestPathFromStart(location)
	if !location.Reachable {
		// ESLint ends a function segment at an abrupt statement before its
		// AST traversal reaches later unreachable nodes. Those nodes are then
		// observed on the still-active outer segment. Reproduce that shared
		// CodePathAnalyzer behavior without pretending the CFG block itself is
		// executable. A Program has no outer segment (and upstream can even
		// throw there), so retain the ordinary unreachable result at the root.
		if codePath.root == nil || codePath.root.Kind == ast.KindSourceFile {
			return state
		}
		state.reachable = true
		state.onEveryFinalPath = codePath.paths.HasSingleFinalPath()
		if shortestExit, ok := codePath.paths.ShortestExitPathFromStart(); ok {
			state.possiblyEarlyReturn = shortestExit <= 1
		}
		return state
	}
	// A for-loop incrementor is laid out before its body so the shared CFG can
	// preserve all of its events. If every body path exits abruptly, the block
	// remains provisionally reachable but has no path from the graph entry.
	// ESLint settles that segment as unreachable before this rule processes it.
	if !reachesFromStart {
		return state
	}
	state.reachable = true
	state.cyclic = codePath.paths.IsCyclic(location)
	state.onEveryFinalPath = codePath.paths.IsOnEveryFinalPath(location)
	shortestFinalPath, hasFinal := codePath.paths.ShortestExitPathFromStart()
	if hasFinal && shortestHookPath != 0 {
		if codePath.paths.IsExit(location) {
			state.possiblyEarlyReturn = shortestFinalPath <= shortestHookPath
		} else {
			state.possiblyEarlyReturn = shortestFinalPath < shortestHookPath
		}
	}
	return state
}

func isInsideDoWhileLoop(node *ast.Node, root *ast.Node) bool {
	for current := node; current != nil && current != root; current = current.Parent {
		if current.Kind == ast.KindDoStatement {
			return true
		}
	}
	return false
}
