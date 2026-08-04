package utils

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	internalUtils "github.com/web-infra-dev/rslint/internal/utils"
)

type RstestTestCallbacks struct {
	Functions          map[*ast.Node]bool
	ContextReceivers   map[*ast.Symbol]bool
	ContextExpectNames map[*ast.Symbol]bool
	fnCalls            *rstestFnCallCache
}

// rstestFnCallCache memoizes call parsing so that the collection walk and the
// rule traversal that follows it do not each parse every call expression.
type rstestFnCallCache struct {
	ctx    rule.RuleContext
	parsed map[*ast.Node]*ParsedRstestFnCall
}

func (cache *rstestFnCallCache) parse(node *ast.Node) *ParsedRstestFnCall {
	if cache == nil {
		return nil
	}
	if parsed, ok := cache.parsed[node]; ok {
		return parsed
	}
	parsed := ParseRstestFnCallWithOfficialExtensions(node, cache.ctx)
	cache.parsed[node] = parsed
	return parsed
}

// ParseFnCall parses node the way CollectRstestTestCallbacks did, reusing the
// cached result when the collection walk already visited node.
func (callbacks RstestTestCallbacks) ParseFnCall(node *ast.Node) *ParsedRstestFnCall {
	if callbacks.fnCalls == nil {
		return nil
	}
	return callbacks.fnCalls.parse(node)
}

type rstestCallbackInfo struct {
	functionNode *ast.Node
	name         string
	parsed       *ParsedRstestFnCall
}

func CollectRstestTestCallbacks(ctx rule.RuleContext) RstestTestCallbacks {
	result := RstestTestCallbacks{
		Functions:          map[*ast.Node]bool{},
		ContextReceivers:   map[*ast.Symbol]bool{},
		ContextExpectNames: map[*ast.Symbol]bool{},
		fnCalls: &rstestFnCallCache{
			ctx:    ctx,
			parsed: map[*ast.Node]*ParsedRstestFnCall{},
		},
	}
	pending := map[string][]*ParsedRstestFnCall{}

	var visit func(*ast.Node)
	visit = func(node *ast.Node) {
		if node == nil {
			return
		}
		if node.Kind == ast.KindCallExpression {
			parsed := result.fnCalls.parse(node)
			if parsed != nil && parsed.Kind == RstestFnTypeTest {
				info := resolveRstestTestCallback(ctx, parsed, node.AsCallExpression())
				if info.functionNode != nil {
					recordRstestCallback(ctx, &result, info.functionNode, parsed)
				} else if info.name != "" {
					pending[info.name] = append(pending[info.name], parsed)
				}
			}
		}
		node.ForEachChild(func(child *ast.Node) bool {
			visit(child)
			return false
		})
	}

	if ctx.SourceFile != nil {
		visit(ctx.SourceFile.Node.AsNode())
	}
	if len(pending) > 0 {
		resolvePendingRstestCallbacks(ctx, &result, pending)
	}
	return result
}

func resolveRstestTestCallback(
	ctx rule.RuleContext,
	parsed *ParsedRstestFnCall,
	call *ast.CallExpression,
) rstestCallbackInfo {
	if call == nil || call.Arguments == nil || len(call.Arguments.Nodes) < 2 {
		return rstestCallbackInfo{}
	}

	arguments := call.Arguments.Nodes
	if len(arguments) >= 3 {
		if info := resolveRstestCallbackArgument(ctx, arguments[2]); info.functionNode != nil || info.name != "" {
			info.parsed = parsed
			return info
		}
	}

	info := resolveRstestCallbackArgument(ctx, arguments[1])
	info.parsed = parsed
	return info
}

func resolveRstestCallbackArgument(ctx rule.RuleContext, argument *ast.Node) rstestCallbackInfo {
	argument = ast.SkipParentheses(argument)
	if argument == nil {
		return rstestCallbackInfo{}
	}
	if ast.IsFunctionExpressionOrArrowFunction(argument) {
		return rstestCallbackInfo{functionNode: argument}
	}
	if argument.Kind != ast.KindIdentifier {
		return rstestCallbackInfo{}
	}

	name := argument.AsIdentifier().Text
	declaration := internalUtils.GetDeclaration(ctx.TypeChecker, argument)
	if declaration == nil {
		return rstestCallbackInfo{name: name}
	}
	switch declaration.Kind {
	case ast.KindFunctionDeclaration:
		return rstestCallbackInfo{functionNode: declaration, name: name}
	case ast.KindVariableDeclaration:
		initializer := ast.SkipParentheses(declaration.AsVariableDeclaration().Initializer)
		if ast.IsFunctionExpressionOrArrowFunction(initializer) {
			return rstestCallbackInfo{functionNode: initializer, name: name}
		}
	}
	return rstestCallbackInfo{}
}

func resolvePendingRstestCallbacks(
	ctx rule.RuleContext,
	result *RstestTestCallbacks,
	pending map[string][]*ParsedRstestFnCall,
) {
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
			if declaration != nil && declaration.Name() != nil && declaration.Name().Kind == ast.KindIdentifier {
				name = declaration.Name().AsIdentifier().Text
				initializer := ast.SkipParentheses(declaration.Initializer)
				if ast.IsFunctionExpressionOrArrowFunction(initializer) {
					function = initializer
				}
			}
		}

		if function != nil {
			for _, parsed := range pending[name] {
				recordRstestCallback(ctx, result, function, parsed)
			}
			delete(pending, name)
		}

		node.ForEachChild(func(child *ast.Node) bool {
			visit(child)
			return false
		})
	}

	if ctx.SourceFile != nil {
		visit(ctx.SourceFile.Node.AsNode())
	}
}

func recordRstestCallback(
	ctx rule.RuleContext,
	result *RstestTestCallbacks,
	function *ast.Node,
	parsed *ParsedRstestFnCall,
) {
	if function == nil {
		return
	}
	result.Functions[function] = true

	contextIndex := 0
	switch parsed.ParameterizedKind {
	case RstestParameterizedEach:
		return
	case RstestParameterizedFor:
		contextIndex = 1
	}

	parameters := function.Parameters()
	if contextIndex >= len(parameters) {
		return
	}
	parameter := parameters[contextIndex].AsParameterDeclaration()
	if parameter == nil || parameter.Name() == nil || ctx.TypeChecker == nil {
		return
	}
	name := parameter.Name()
	switch name.Kind {
	case ast.KindIdentifier:
		if symbol := ctx.TypeChecker.GetSymbolAtLocation(name); symbol != nil {
			result.ContextReceivers[symbol] = true
		}
	case ast.KindObjectBindingPattern:
		pattern := name.AsBindingPattern()
		if pattern == nil || pattern.Elements == nil {
			return
		}
		for _, element := range pattern.Elements.Nodes {
			if element == nil || element.Kind != ast.KindBindingElement {
				continue
			}
			binding := element.AsBindingElement()
			if binding == nil || binding.Name() == nil || binding.Name().Kind != ast.KindIdentifier {
				continue
			}
			propertyName := binding.Name().AsIdentifier().Text
			if binding.PropertyName != nil {
				if value, ok := internalUtils.GetStaticStringLiteralValue(binding.PropertyName); ok {
					propertyName = value
				} else if binding.PropertyName.Kind == ast.KindIdentifier {
					propertyName = binding.PropertyName.AsIdentifier().Text
				}
			}
			if propertyName != "expect" {
				continue
			}
			if symbol := ctx.TypeChecker.GetSymbolAtLocation(binding.Name()); symbol != nil {
				result.ContextExpectNames[symbol] = true
			}
		}
	}
}
