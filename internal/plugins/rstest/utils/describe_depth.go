package utils

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

type rstestDescribeDepthFileCacheKey struct{}

type RstestDescribeDepth struct {
	analysis  *RstestCallAnalysis
	ownership map[*ast.Node][]rstestCallbackRegistration
	depths    map[*ast.Node]int
	visiting  map[*ast.Node]bool
}

func GetRstestDescribeDepth(
	ctx rule.RuleContext,
	analysis *RstestCallAnalysis,
) *RstestDescribeDepth {
	return rule.CachedByFile(
		ctx,
		rstestDescribeDepthFileCacheKey{},
		func() *RstestDescribeDepth {
			return &RstestDescribeDepth{
				analysis: analysis,
				depths:   map[*ast.Node]int{},
				visiting: map[*ast.Node]bool{},
			}
		},
	)
}

// Depth returns the deepest suite nesting level at which a describe
// registration can execute. Callback ownership accounts for both inline
// callbacks and same-file callback references.
func (context *RstestDescribeDepth) Depth(call *ast.Node) int {
	parsed := context.analysis.ParseFnCall(call)
	if parsed == nil || parsed.Kind != RstestFnTypeDescribe {
		return 0
	}
	return context.describeDepth(call)
}

func (context *RstestDescribeDepth) describeDepth(call *ast.Node) int {
	if depth := context.depths[call]; depth != 0 {
		return depth
	}
	if context.visiting[call] {
		return 0
	}
	context.visiting[call] = true
	defer delete(context.visiting, call)

	depth := 1
	callback := context.nearestOwnedCallback(call)
	for _, registration := range context.ownershipIndex()[callback] {
		if registration.parsed.Kind != RstestFnTypeDescribe {
			continue
		}
		if parentDepth := context.describeDepth(registration.call); parentDepth+1 > depth {
			depth = parentDepth + 1
		}
	}
	context.depths[call] = depth
	return depth
}

func (context *RstestDescribeDepth) ownershipIndex() map[*ast.Node][]rstestCallbackRegistration {
	if context.ownership == nil {
		context.ownership = context.analysis.callbackOwnership()
	}
	return context.ownership
}

func (context *RstestDescribeDepth) nearestOwnedCallback(node *ast.Node) *ast.Node {
	return nearestRstestOwnedCallback(node, context.ownershipIndex())
}
