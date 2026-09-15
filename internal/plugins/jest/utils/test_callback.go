package utils

import (
	"slices"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

type JestTestCallbacks struct {
	Functions   map[*ast.Node]bool
	IgnoredDone map[*ast.Node]bool
	fnCalls     *jestFnCallCache
}

type jestFnCallCache struct {
	ctx    rule.RuleContext
	parsed map[*ast.Node]*ParsedJestFnCall
}

func (cache *jestFnCallCache) parse(node *ast.Node) *ParsedJestFnCall {
	if cache == nil {
		return nil
	}
	if parsed, ok := cache.parsed[node]; ok {
		return parsed
	}
	parsed := ParseJestFnCall(node, cache.ctx)
	cache.parsed[node] = parsed
	return parsed
}

func (callbacks JestTestCallbacks) ParseFnCall(node *ast.Node) *ParsedJestFnCall {
	if callbacks.fnCalls == nil {
		return nil
	}
	return callbacks.fnCalls.parse(node)
}

type jestCallbackUsage struct {
	tracked bool
	ignored bool
}

type pendingJestCallback struct {
	name   string
	call   *ast.Node
	parsed *ParsedJestFnCall
}

func CollectJestTestCallbacks(ctx rule.RuleContext) JestTestCallbacks {
	result := JestTestCallbacks{
		Functions:   map[*ast.Node]bool{},
		IgnoredDone: map[*ast.Node]bool{},
		fnCalls: &jestFnCallCache{
			ctx:    ctx,
			parsed: map[*ast.Node]*ParsedJestFnCall{},
		},
	}
	usages := map[*ast.Node]jestCallbackUsage{}
	var pending []pendingJestCallback

	var visit func(*ast.Node)
	visit = func(node *ast.Node) {
		if node == nil {
			return
		}
		if node.Kind == ast.KindCallExpression {
			parsed := result.fnCalls.parse(node)
			if parsed != nil && parsed.Kind == JestFnTypeTest {
				function, name := resolveJestTestCallback(ctx, node.AsCallExpression())
				ignored := isDoneAmbiguousJestCallback(node, parsed, function)
				if function != nil {
					recordJestCallbackUsage(usages, function, ignored)
				} else if name != "" {
					pending = append(pending, pendingJestCallback{
						name:   name,
						call:   node,
						parsed: parsed,
					})
				}
			}
		}
		node.ForEachChild(func(child *ast.Node) bool {
			visit(child)
			return false
		})
	}

	if ctx.SourceFile != nil {
		visit(ctx.SourceFile.AsNode())
	}
	if len(pending) > 0 {
		resolvePendingJestCallbacks(ctx, pending, usages)
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
) (*ast.Node, string) {
	if call == nil || call.Arguments == nil || len(call.Arguments.Nodes) < 2 {
		return nil, ""
	}
	callback := ast.SkipParentheses(call.Arguments.Nodes[1])
	if callback == nil {
		return nil, ""
	}
	if ast.IsFunctionExpressionOrArrowFunction(callback) {
		return callback, ""
	}
	info := ResolveTestCallbackFunction(ctx, call)
	return info.FunctionNode, info.Name
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

func resolvePendingJestCallbacks(
	ctx rule.RuleContext,
	pending []pendingJestCallback,
	usages map[*ast.Node]jestCallbackUsage,
) {
	byName := map[string][]pendingJestCallback{}
	for _, item := range pending {
		byName[item.name] = append(byName[item.name], item)
	}

	var visit func(*ast.Node)
	visit = func(node *ast.Node) {
		if node == nil {
			return
		}
		name := ""
		var function *ast.Node
		switch node.Kind {
		case ast.KindFunctionDeclaration:
			declaration := node.AsFunctionDeclaration()
			if declaration != nil && declaration.Name() != nil {
				name = declaration.Name().Text()
				function = node
			}
		case ast.KindVariableDeclaration:
			declaration := node.AsVariableDeclaration()
			if declaration != nil && declaration.Name() != nil &&
				declaration.Name().Kind == ast.KindIdentifier &&
				declaration.Initializer != nil {
				name = declaration.Name().Text()
				initializer := ast.SkipParentheses(declaration.Initializer)
				if ast.IsFunctionExpressionOrArrowFunction(initializer) {
					function = initializer
				}
			}
		}
		if function != nil {
			for _, item := range byName[name] {
				ignored := isDoneAmbiguousJestCallback(item.call, item.parsed, function)
				recordJestCallbackUsage(usages, function, ignored)
			}
			delete(byName, name)
		}
		node.ForEachChild(func(child *ast.Node) bool {
			visit(child)
			return false
		})
	}
	if ctx.SourceFile != nil {
		visit(ctx.SourceFile.AsNode())
	}
}
