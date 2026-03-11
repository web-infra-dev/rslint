package array_callback_return

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// Options for array-callback-return rule
type Options struct {
	AllowImplicit bool `json:"allowImplicit"`
	CheckForEach  bool `json:"checkForEach"`
	AllowVoid     bool `json:"allowVoid"`
}

func parseOptions(options any) Options {
	opts := Options{
		AllowImplicit: false,
		CheckForEach:  false,
		AllowVoid:     false,
	}

	if options == nil {
		return opts
	}

	// Parse options with dual-format support (handles both array and object formats)
	var optsMap map[string]interface{}
	var ok bool

	// Handle array format: [{ option: value }]
	if optArray, isArray := options.([]interface{}); isArray && len(optArray) > 0 {
		optsMap, ok = optArray[0].(map[string]interface{})
	} else {
		// Handle direct object format: { option: value }
		optsMap, ok = options.(map[string]interface{})
	}

	if ok {
		if v, ok := optsMap["allowImplicit"].(bool); ok {
			opts.AllowImplicit = v
		}
		if v, ok := optsMap["checkForEach"].(bool); ok {
			opts.CheckForEach = v
		}
		if v, ok := optsMap["allowVoid"].(bool); ok {
			opts.AllowVoid = v
		}
	}
	return opts
}

// Message builders
func buildExpectedReturnValue(methodName string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "expectedReturnValue",
		Description: "Array.prototype." + methodName + "() expects a return value from arrow function.",
	}
}

func buildExpectedInside(methodName string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "expectedInside",
		Description: "Array.prototype." + methodName + "() expects a value to be returned at the end of arrow function.",
	}
}

func buildExpectedNoReturnValue(methodName string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "expectedNoReturnValue",
		Description: "Array.prototype." + methodName + "() expects no useless return value from arrow function.",
	}
}

func buildExpectedAtEnd(methodName string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "expectedAtEnd",
		Description: "Array.prototype." + methodName + "() expects a return value at the end of function.",
	}
}

// Target array methods that require return values
var targetMethods = map[string]bool{
	"every":         true,
	"filter":        true,
	"find":          true,
	"findIndex":     true,
	"findLast":      true,
	"findLastIndex": true,
	"flatMap":       true,
	"map":           true,
	"reduce":        true,
	"reduceRight":   true,
	"some":          true,
	"sort":          true,
	"toSorted":      true,
	// from is handled separately
}

// Methods that should not have return values
var forEachMethod = "forEach"

// isTargetMethod checks if the method is a target array method
func isTargetMethod(name string) bool {
	return targetMethods[name]
}

// isForEachMethod checks if the method is forEach
func isForEachMethod(name string) bool {
	return name == forEachMethod
}

// getMethodName extracts the method name from a CallExpression
func getMethodName(node *ast.Node) string {
	if node == nil || node.Kind != ast.KindCallExpression {
		return ""
	}

	expr := node.Expression()
	if expr == nil {
		return ""
	}

	// Handle PropertyAccessExpression (e.g., arr.map)
	if expr.Kind == ast.KindPropertyAccessExpression {
		name := expr.Name()
		if name != nil && name.Kind == ast.KindIdentifier {
			return name.Text()
		}
	}

	return ""
}

// isArrayFromCall checks if this is Array.from()
func isArrayFromCall(node *ast.Node) bool {
	if node == nil || node.Kind != ast.KindCallExpression {
		return false
	}

	expr := node.Expression()
	if expr == nil || expr.Kind != ast.KindPropertyAccessExpression {
		return false
	}

	obj := expr.Expression()
	if obj == nil || obj.Kind != ast.KindIdentifier || obj.Text() != "Array" {
		return false
	}

	name := expr.Name()
	if name == nil || name.Kind != ast.KindIdentifier {
		return false
	}

	return name.Text() == "from"
}

// isFunctionNode checks if a node is a function
func isFunctionNode(node *ast.Node) bool {
	if node == nil {
		return false
	}
	return node.Kind == ast.KindFunctionDeclaration ||
		node.Kind == ast.KindFunctionExpression ||
		node.Kind == ast.KindArrowFunction ||
		node.Kind == ast.KindMethodDeclaration
}

// isGeneratorOrAsyncFunction checks if a node is a generator or async function
func isGeneratorOrAsyncFunction(node *ast.Node) bool {
	if node == nil {
		return false
	}

	// Check if it has async modifier
	modifiers := node.Modifiers()
	if modifiers != nil {
		for _, mod := range modifiers.Nodes {
			if mod != nil && mod.Kind == ast.KindAsyncKeyword {
				return true
			}
		}
	}

	// Check for generator function (function*)
	if containsYield(node.Body()) {
		return true
	}

	return false
}

// containsYield recursively checks if a node contains a yield expression
func containsYield(node *ast.Node) bool {
	if node == nil {
		return false
	}
	if node.Kind == ast.KindYieldExpression {
		return true
	}
	// Recursively check children
	found := false
	node.ForEachChild(func(child *ast.Node) bool {
		if containsYield(child) {
			found = true
			return true // Stop iteration
		}
		return false // Continue iteration
	})
	return found
}

// isVoidExpression checks if an expression is a void expression
func isVoidExpression(node *ast.Node) bool {
	if node == nil {
		return false
	}
	return node.Kind == ast.KindVoidExpression
}

// checkCallbackReturn validates that a callback has proper return statements
func checkCallbackReturn(ctx rule.RuleContext, funcNode *ast.Node, methodName string, opts Options, checkForEach bool) {
	if funcNode == nil {
		return
	}

	// Skip generator and async functions - they have different control flow semantics
	if isGeneratorOrAsyncFunction(funcNode) {
		return
	}

	body := funcNode.Body()
	if body == nil {
		return
	}

	// For arrow functions with expression bodies, check if it's an implicit return
	if funcNode.Kind == ast.KindArrowFunction {
		// Expression body (no braces) - always returns the expression value
		if body.Kind != ast.KindBlock {
			if checkForEach {
				// forEach shouldn't return values, but expression bodies always do
				// unless it's a void expression
				if opts.AllowVoid && isVoidExpression(body) {
					return
				}
				ctx.ReportNode(funcNode, buildExpectedNoReturnValue(methodName))
			}
			// For other methods, expression bodies always return, which is good
			return
		}
	}

	analysis := utils.AnalyzeFunctionReturns(funcNode)

	if checkForEach {
		// forEach callbacks should not return values
		if analysis.HasReturnWithValue {
			ctx.ReportNode(funcNode, buildExpectedNoReturnValue(methodName))
		}
		return
	}

	// When allowImplicit is false, empty returns (return;) don't count as valid returns.
	// When allowImplicit is true, empty returns are acceptable.
	hasNoReturns := !analysis.HasReturnWithValue && (!opts.AllowImplicit || !analysis.HasEmptyReturn)
	var allPathsReturn bool
	if opts.AllowImplicit {
		allPathsReturn = !analysis.EndReachable && (analysis.HasReturnWithValue || analysis.HasEmptyReturn)
	} else {
		// Without allowImplicit, all return statements must have values.
		// Mixed return-with-value + empty-return is invalid (e.g., if (a) return 1; else return;)
		allPathsReturn = !analysis.EndReachable && analysis.HasReturnWithValue && !analysis.HasEmptyReturn
	}

	if hasNoReturns {
		if funcNode.Kind == ast.KindArrowFunction {
			ctx.ReportNode(funcNode, buildExpectedReturnValue(methodName))
		} else {
			ctx.ReportNode(funcNode, buildExpectedAtEnd(methodName))
		}
	} else if !allPathsReturn {
		if funcNode.Kind == ast.KindArrowFunction {
			ctx.ReportNode(funcNode, buildExpectedInside(methodName))
		} else {
			ctx.ReportNode(funcNode, buildExpectedAtEnd(methodName))
		}
	}
}

// ArrayCallbackReturnRule enforces return statements in callbacks of array methods
var ArrayCallbackReturnRule = rule.CreateRule(rule.Rule{
	Name: "array-callback-return",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := parseOptions(options)

		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				methodName := getMethodName(node)
				isArrayFrom := false

				// Check if it's Array.from (needs special handling)
				if isArrayFromCall(node) {
					methodName = "from"
					isArrayFrom = true
				}

				// Check if we have a method name
				if methodName == "" {
					return
				}

				// Check if it's a target method
				isTarget := isTargetMethod(methodName)
				isForEach := opts.CheckForEach && isForEachMethod(methodName)

				if !isTarget && !isForEach && !isArrayFrom {
					return
				}

				// Get the callback argument
				args := node.Arguments()
				if len(args) == 0 {
					return
				}

				var callbackArg *ast.Node
				if isArrayFrom {
					// Array.from(arr, callback) - callback is second argument
					if len(args) >= 2 {
						callbackArg = args[1]
					}
				} else {
					// For other methods, callback is first argument
					callbackArg = args[0]
				}

				if callbackArg == nil {
					return
				}

				// Check if the argument is a function
				if !isFunctionNode(callbackArg) {
					return
				}

				// Check the callback
				checkCallbackReturn(ctx, callbackArg, methodName, opts, isForEach)
			},
		}
	},
})
