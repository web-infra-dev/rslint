package utils

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	testFramework "github.com/web-infra-dev/rslint/internal/utils/test_framework"
)

type rstestConcurrentContextFileCacheKey struct{}

type rstestConcurrentModeState uint8

const (
	rstestConcurrentModeUnknown rstestConcurrentModeState = iota
	rstestConcurrentModeVisiting
	rstestConcurrentModeSequential
	rstestConcurrentModeConcurrent
)

// RstestConcurrentContext answers whether a node runs inside a concurrent test.
// The callback ownership index it reads is immutable; the resolved modes are
// memoized here because they are a conclusion of this query, not a property of
// the ownership itself.
type RstestConcurrentContext struct {
	analysis  *RstestCallAnalysis
	ownership map[*ast.Node][]rstestCallbackRegistration
	modes     map[*ast.Node]rstestConcurrentModeState
}

func GetRstestConcurrentContext(
	ctx rule.RuleContext,
	analysis *RstestCallAnalysis,
) *RstestConcurrentContext {
	return rule.CachedByFile(
		ctx,
		rstestConcurrentContextFileCacheKey{},
		func() *RstestConcurrentContext {
			return &RstestConcurrentContext{
				analysis: analysis,
				modes:    map[*ast.Node]rstestConcurrentModeState{},
			}
		},
	)
}

// IsInConcurrentTest reports whether node belongs to a test or describe
// callback with an effective concurrent execution mode. Explicit modes on the
// nearest registration override inherited describe modes. Callback ownership
// supports both inline and same-file named functions.
func (context *RstestConcurrentContext) IsInConcurrentTest(node *ast.Node) bool {
	concurrent, _ := context.callbackRunsConcurrently(context.nearestOwnedCallback(node))
	return concurrent
}

func (context *RstestConcurrentContext) ownershipIndex() map[*ast.Node][]rstestCallbackRegistration {
	if context.ownership == nil {
		context.ownership = collectRstestCallbackOwnership(context.analysis)
	}
	return context.ownership
}

// nearestOwnedCallback finds the registration callback that node runs inside.
// Functions that are not themselves callbacks are walked through rather than
// stopping the search: a helper, hook or inline closure declared inside a
// concurrent test body still runs as part of that test whenever it runs at
// all, which is exactly when a non-local expect shares snapshot state.
func (context *RstestConcurrentContext) nearestOwnedCallback(node *ast.Node) *ast.Node {
	if node == nil {
		return nil
	}
	ownership := context.ownershipIndex()
	for current := node.Parent; current != nil; current = current.Parent {
		if testFramework.IsFunction(current) && ownership[current] != nil {
			return current
		}
	}
	return nil
}

// callbackRunsConcurrently reports whether function runs concurrently, and
// whether that answer is final. An answer is not final while an enclosing
// callback is still being resolved, which happens when registrations reference
// each other in a cycle; such answers are not memoized.
func (context *RstestConcurrentContext) callbackRunsConcurrently(
	function *ast.Node,
) (concurrent bool, resolved bool) {
	if function == nil {
		return false, true
	}
	registrations := context.ownershipIndex()[function]
	if registrations == nil {
		return false, true
	}
	switch context.modes[function] {
	case rstestConcurrentModeVisiting:
		return false, false
	case rstestConcurrentModeSequential:
		return false, true
	case rstestConcurrentModeConcurrent:
		return true, true
	}

	context.modes[function] = rstestConcurrentModeVisiting
	resolved = true
	for _, registration := range registrations {
		switch registration.parsed.ExecutionMode {
		case RstestExecutionConcurrent:
			context.modes[function] = rstestConcurrentModeConcurrent
			return true, true
		case RstestExecutionSequential:
			continue
		}
		outer := context.nearestOwnedCallback(registration.call)
		if outer == nil {
			continue
		}
		outerConcurrent, outerResolved := context.callbackRunsConcurrently(outer)
		if outerConcurrent {
			context.modes[function] = rstestConcurrentModeConcurrent
			return true, true
		}
		if !outerResolved {
			resolved = false
		}
	}
	if resolved {
		context.modes[function] = rstestConcurrentModeSequential
	} else {
		context.modes[function] = rstestConcurrentModeUnknown
	}
	return false, resolved
}
