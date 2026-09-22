package utils

import (
	"slices"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

type JestTestCallbacks struct {
	Functions   map[*ast.Node]bool
	IgnoredDone map[*ast.Node]bool
}

type jestCallbackInfo struct {
	functionNode *ast.Node
	name         string
}

type jestCallbackUsage struct {
	tracked bool
	ignored bool
}

func collectJestTestCallbacks(analysis *JestCallAnalysis) JestTestCallbacks {
	result := JestTestCallbacks{
		Functions:   map[*ast.Node]bool{},
		IgnoredDone: map[*ast.Node]bool{},
	}
	usages := map[*ast.Node]jestCallbackUsage{}
	analysis.indexSourceFile()
	for _, node := range analysis.calls {
		parsed := analysis.ParseTestCall(node)
		if parsed == nil {
			continue
		}
		info := analysis.testCallbackInfo(node)
		if info.functionNode == nil {
			info.functionNode = analysis.fallbackCallbackFunction(info.name)
		}
		if info.functionNode == nil {
			continue
		}
		ignored := isDoneAmbiguousJestCallback(node, parsed, info.functionNode)
		recordJestCallbackUsage(usages, info.functionNode, ignored)
	}
	for function, usage := range usages {
		result.Functions[function] = true
		if usage.ignored && !usage.tracked {
			result.IgnoredDone[function] = true
		}
	}
	return result
}

func resolveJestTestCallback(
	ctx rule.RuleContext,
	call *ast.CallExpression,
) jestCallbackInfo {
	if call == nil || call.Arguments == nil || len(call.Arguments.Nodes) < 2 {
		return jestCallbackInfo{}
	}
	callback := ast.SkipParentheses(call.Arguments.Nodes[1])
	if callback == nil {
		return jestCallbackInfo{}
	}
	if ast.IsFunctionExpressionOrArrowFunction(callback) {
		return jestCallbackInfo{functionNode: callback}
	}
	return resolveNamedTestCallback(ctx, call)
}

func isDoneAmbiguousJestCallback(
	callNode *ast.Node,
	parsed *ParsedJestFnCall,
	function *ast.Node,
) bool {
	if function == nil {
		return false
	}
	if slices.Contains(parsed.Members, "each") {
		callee := ast.SkipParentheses(callNode.AsCallExpression().Expression)
		if callee == nil || callee.Kind != ast.KindTaggedTemplateExpression {
			return true
		}
		return len(function.Parameters()) == 2
	}
	return len(function.Parameters()) == 1
}

func recordJestCallbackUsage(
	usages map[*ast.Node]jestCallbackUsage,
	function *ast.Node,
	ignored bool,
) {
	if function == nil {
		return
	}
	usage := usages[function]
	if ignored {
		usage.ignored = true
	} else {
		usage.tracked = true
	}
	usages[function] = usage
}
