package max_depth

import (
	"fmt"
	"math"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// MaxDepthRule enforces a maximum depth that blocks can be nested in a function.
// https://eslint.org/docs/latest/rules/max-depth
var MaxDepthRule = rule.Rule{
	Name: "max-depth",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		maxDepth := parseMaxDepth(options)

		// Stack of depth counters, one per scope-bearing container. The
		// linter does not fire a KindSourceFile listener, so we seed the
		// stack with one frame for the source file scope; function-likes
		// and class static blocks push their own frames on top via
		// startFunction.
		stack := []int{0}

		startFunction := func(node *ast.Node) {
			stack = append(stack, 0)
		}
		endFunction := func(node *ast.Node) {
			if len(stack) > 1 {
				stack = stack[:len(stack)-1]
			}
		}
		pushBlock := func(node *ast.Node) {
			n := len(stack)
			if n == 0 {
				return
			}
			stack[n-1]++
			depth := stack[n-1]
			if depth > maxDepth {
				ctx.ReportNode(node, buildTooDeeplyMessage(depth, maxDepth))
			}
		}
		popBlock := func(node *ast.Node) {
			n := len(stack)
			if n == 0 {
				return
			}
			stack[n-1]--
		}

		// `else if` chains: tsgo (like ESTree) represents `if (a) {} else if (b) {}`
		// as `IfStatement(a) { alternate: IfStatement(b) }`. ESLint suppresses the
		// inner push when the parent is itself an IfStatement so the chain counts
		// as a single nesting level.
		//
		// NOTE: The exit handler is intentionally `popBlock` (unconditional),
		// matching ESLint exactly. ESLint's `IfStatement:exit` pops on every
		// IfStatement regardless of whether the entry was suppressed, which
		// leaves a negative residual on the depth counter after each `else if`
		// chain. Sibling code following the chain therefore enjoys an
		// artificial allowance equal to the chain length. We mirror this
		// behavior to preserve 1:1 parity with ESLint's diagnostics.
		ifEnter := func(node *ast.Node) {
			if node.Parent != nil && node.Parent.Kind == ast.KindIfStatement {
				return
			}
			pushBlock(node)
		}

		return rule.RuleListeners{
			ast.KindFunctionDeclaration:                              startFunction,
			rule.ListenerOnExit(ast.KindFunctionDeclaration):         endFunction,
			ast.KindFunctionExpression:                               startFunction,
			rule.ListenerOnExit(ast.KindFunctionExpression):          endFunction,
			ast.KindArrowFunction:                                    startFunction,
			rule.ListenerOnExit(ast.KindArrowFunction):               endFunction,
			ast.KindClassStaticBlockDeclaration:                      startFunction,
			rule.ListenerOnExit(ast.KindClassStaticBlockDeclaration): endFunction,
			// ESLint listens FunctionExpression for class/object methods,
			// getters, setters, and constructors because ESTree wraps each in
			// a FunctionExpression value. tsgo represents them as distinct
			// kinds, so listen explicitly to keep depth tracking aligned.
			ast.KindMethodDeclaration:                      startFunction,
			rule.ListenerOnExit(ast.KindMethodDeclaration): endFunction,
			ast.KindGetAccessor:                            startFunction,
			rule.ListenerOnExit(ast.KindGetAccessor):       endFunction,
			ast.KindSetAccessor:                            startFunction,
			rule.ListenerOnExit(ast.KindSetAccessor):       endFunction,
			ast.KindConstructor:                            startFunction,
			rule.ListenerOnExit(ast.KindConstructor):       endFunction,

			ast.KindIfStatement:                          ifEnter,
			rule.ListenerOnExit(ast.KindIfStatement):     popBlock,
			ast.KindSwitchStatement:                      pushBlock,
			rule.ListenerOnExit(ast.KindSwitchStatement): popBlock,
			ast.KindTryStatement:                         pushBlock,
			rule.ListenerOnExit(ast.KindTryStatement):    popBlock,
			ast.KindDoStatement:                          pushBlock,
			rule.ListenerOnExit(ast.KindDoStatement):     popBlock,
			ast.KindWhileStatement:                       pushBlock,
			rule.ListenerOnExit(ast.KindWhileStatement):  popBlock,
			ast.KindWithStatement:                        pushBlock,
			rule.ListenerOnExit(ast.KindWithStatement):   popBlock,
			ast.KindForStatement:                         pushBlock,
			rule.ListenerOnExit(ast.KindForStatement):    popBlock,
			ast.KindForInStatement:                       pushBlock,
			rule.ListenerOnExit(ast.KindForInStatement):  popBlock,
			ast.KindForOfStatement:                       pushBlock,
			rule.ListenerOnExit(ast.KindForOfStatement):  popBlock,
		}
	},
}

func buildTooDeeplyMessage(depth, maxDepth int) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "tooDeeply",
		Description: fmt.Sprintf("Blocks are nested too deeply (%d). Maximum allowed is %d.", depth, maxDepth),
	}
}

// parseMaxDepth resolves the configured maximum depth, mirroring ESLint's
// `option.maximum || option.max` coercion. The legacy `maximum` key is honored
// only when truthy (matching JS coercion); otherwise `max` wins. When neither
// key is present, the default is 4.
func parseMaxDepth(options any) int {
	const defaultMax = 4
	if options == nil {
		return defaultMax
	}
	// Number form: `3` or `[3]`.
	if arr, ok := options.([]interface{}); ok {
		if len(arr) == 0 {
			return defaultMax
		}
		if n, ok := utils.CoerceInt(arr[0]); ok {
			return n
		}
	} else if n, ok := utils.CoerceInt(options); ok {
		return n
	}
	// Object form: `{ max: 3 }` or `[{ max: 3 }]`. Use the shared extractor so
	// both the array-wrapped (rule_tester / multi-element CLI) and bare-object
	// (single-option CLI) shapes are handled uniformly.
	m := utils.GetOptionsMap(options)
	if m == nil {
		return defaultMax
	}
	_, hasMaximum := m["maximum"]
	_, hasMax := m["max"]
	if !hasMaximum && !hasMax {
		// Matches ESLint: when neither key is present, the option object is
		// ignored and the default is used.
		return defaultMax
	}
	if hasMaximum {
		if v, ok := utils.CoerceInt(m["maximum"]); ok && v != 0 {
			return v
		}
	}
	if hasMax {
		if v, ok := utils.CoerceInt(m["max"]); ok {
			return v
		}
	}
	// `option.maximum` is present but coerces to 0 / non-numeric, and
	// `option.max` is absent or non-numeric: ESLint sets `maxDepth = undefined`
	// here, which makes every `len > maxDepth` comparison false and effectively
	// disables the check. MaxInt produces the same observable behavior.
	return math.MaxInt
}

