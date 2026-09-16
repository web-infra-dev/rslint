package nodeutil

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// CollectRequireCalls returns require() and require.resolve() calls, following
// aliases and destructuring as upstream's visitRequire does. It respects
// effective globals, shadowing and writes to global roots. Calls are returned
// in traversal order; separate alias paths can return the same call more than once.
func CollectRequireCalls(ctx rule.RuleContext) []*ast.Node {
	return newReferenceTracker(ctx).collectRequireCalls(true)
}

func (tracker *referenceTracker) collectRequireCalls(includeResolve bool) []*ast.Node {
	var calls []*ast.Node
	value := &referenceTrace{call: func(node *ast.Node) { calls = append(calls, node) }}
	if includeResolve {
		value.properties = map[string]*referenceTrace{"resolve": {call: value.call}}
	}
	tracker.trackGlobals(map[string]*referenceTrace{"require": value})
	return calls
}
