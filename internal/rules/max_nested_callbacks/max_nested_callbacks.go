package max_nested_callbacks

import (
	"fmt"
	"math"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// MaxNestedCallbacksRule enforces a maximum depth of nested callbacks.
// https://eslint.org/docs/latest/rules/max-nested-callbacks
var MaxNestedCallbacksRule = rule.Rule{
	Name: "max-nested-callbacks",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		threshold := parseThreshold(options)

		// `callbackStack` mirrors ESLint's stack of FunctionExpression /
		// ArrowFunctionExpression nodes that sit *directly* under a
		// CallExpression — i.e. function-likes used as call arguments or as
		// the callee of an immediately-invoked call. Push happens only when
		// the (paren-flattened) parent is a CallExpression; pop fires on
		// every function-like exit, even when the entry didn't push. The
		// asymmetry is preserved verbatim from ESLint to keep diagnostic
		// counts byte-for-byte identical with upstream.
		var callbackStack []*ast.Node

		check := func(node *ast.Node) {
			// tsgo preserves ParenthesizedExpression nodes that ESTree
			// flattens. Walk them so `(function(){})()` and `foo((fn))` are
			// recognized as the function's "real" parent being a CallExpression.
			parent := ast.WalkUpParenthesizedExpressions(node.Parent)
			if parent != nil && ast.IsCallExpression(parent) {
				callbackStack = append(callbackStack, node)
			}
			if len(callbackStack) > threshold {
				ctx.ReportNode(node, buildExceedMessage(len(callbackStack), threshold))
			}
		}
		pop := func(node *ast.Node) {
			// JS Array#pop on an empty array is a no-op; bound the slice
			// likewise so the asymmetric exit-pop can't underflow.
			if len(callbackStack) > 0 {
				callbackStack = callbackStack[:len(callbackStack)-1]
			}
		}

		return rule.RuleListeners{
			// FunctionExpression / ArrowFunction map directly to ESLint's
			// FunctionExpression / ArrowFunctionExpression — push when the
			// (paren-flattened) parent is a CallExpression, pop on exit.
			ast.KindFunctionExpression:                      check,
			rule.ListenerOnExit(ast.KindFunctionExpression): pop,
			ast.KindArrowFunction:                           check,
			rule.ListenerOnExit(ast.KindArrowFunction):      pop,

			// In ESTree, class methods / getters / setters / constructors and
			// object-shorthand methods are all `FunctionExpression` nodes
			// nested inside `MethodDefinition` / `Property`. ESLint's listener
			// therefore fires on each of them — performing both the threshold
			// check on entry (without pushing, since parent is not a
			// CallExpression) and the unconditional pop on exit. tsgo
			// represents these as distinct AST kinds with no inner FE node, so
			// we wire entry+exit on each kind to reproduce the diagnostic
			// count exactly. The reported start position is the tsgo node
			// itself rather than the wrapped FE — message text, messageId,
			// num / max values, and line all match upstream; column may shift
			// by the length of the method name and modifiers.
			ast.KindMethodDeclaration:                      check,
			rule.ListenerOnExit(ast.KindMethodDeclaration): pop,
			ast.KindGetAccessor:                            check,
			rule.ListenerOnExit(ast.KindGetAccessor):       pop,
			ast.KindSetAccessor:                            check,
			rule.ListenerOnExit(ast.KindSetAccessor):       pop,
			ast.KindConstructor:                            check,
			rule.ListenerOnExit(ast.KindConstructor):       pop,
		}
	},
}

func buildExceedMessage(num, threshold int) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "exceed",
		Description: fmt.Sprintf("Too many nested callbacks (%d). Maximum allowed is %d.", num, threshold),
	}
}

// parseThreshold resolves the configured maximum nesting depth, mirroring
// ESLint's `option.maximum || option.max` coercion. The legacy `maximum` key
// is honored only when truthy (matching JS coercion); otherwise `max` wins.
// When neither key is present, the default is 10.
func parseThreshold(options any) int {
	const defaultMax = 10
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
	// `option.max` is absent or non-numeric: ESLint sets `THRESHOLD = undefined`
	// here, which makes every `length > THRESHOLD` comparison false and
	// effectively disables the check. MaxInt produces the same observable
	// behavior.
	return math.MaxInt
}
