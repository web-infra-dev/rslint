package utils

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

type rstestDescribeDepthFileCacheKey struct{}

type rstestDescribeContainmentState uint8

const (
	rstestDescribeContainmentUnknown rstestDescribeContainmentState = iota
	rstestDescribeContainmentVisiting
	rstestDescribeContainmentOutside
	rstestDescribeContainmentInside
)

type RstestDescribeDepth struct {
	analysis    *RstestCallAnalysis
	ownership   map[*ast.Node][]rstestCallbackRegistration
	depths      map[*ast.Node]int
	visiting    map[*ast.Node]bool
	containment map[*ast.Node]rstestDescribeContainmentState
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
				analysis:    analysis,
				depths:      map[*ast.Node]int{},
				visiting:    map[*ast.Node]bool{},
				containment: map[*ast.Node]rstestDescribeContainmentState{},
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

// InsideDescribe reports whether node executes inside a describe callback.
// Callback ownership covers both inline callbacks and same-file callbacks
// passed by reference, so a registration written in a named function that a
// describe runs counts as nested even though no describe encloses it
// lexically.
func (context *RstestDescribeDepth) InsideDescribe(node *ast.Node) bool {
	inside, _ := context.callbackInsideDescribe(context.nearestOwnedCallback(node))
	return inside
}

// callbackInsideDescribe reports whether function runs inside a suite, and
// whether that answer is final. An answer is not final while an enclosing
// callback is still being resolved, which happens when registrations reference
// each other in a cycle; such answers are not memoized.
func (context *RstestDescribeDepth) callbackInsideDescribe(
	function *ast.Node,
) (inside bool, resolved bool) {
	if function == nil {
		return false, true
	}
	registrations := context.ownershipIndex()[function]
	if registrations == nil {
		return false, true
	}
	switch context.containment[function] {
	case rstestDescribeContainmentVisiting:
		return false, false
	case rstestDescribeContainmentOutside:
		return false, true
	case rstestDescribeContainmentInside:
		return true, true
	}

	context.containment[function] = rstestDescribeContainmentVisiting
	resolved = true
	for _, registration := range registrations {
		if registration.parsed.Kind == RstestFnTypeDescribe {
			context.containment[function] = rstestDescribeContainmentInside
			return true, true
		}
		outerInside, outerResolved := context.callbackInsideDescribe(
			context.nearestOwnedCallback(registration.call),
		)
		if outerInside {
			context.containment[function] = rstestDescribeContainmentInside
			return true, true
		}
		if !outerResolved {
			resolved = false
		}
	}
	if resolved {
		context.containment[function] = rstestDescribeContainmentOutside
	} else {
		context.containment[function] = rstestDescribeContainmentUnknown
	}
	return false, resolved
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
