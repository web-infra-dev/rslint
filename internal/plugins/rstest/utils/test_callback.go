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
}

type rstestCallbackInfo struct {
	functionNode *ast.Node
	name         string
}

func newRstestTestCallbacks() RstestTestCallbacks {
	return RstestTestCallbacks{
		Functions:          map[*ast.Node]bool{},
		ContextReceivers:   map[*ast.Symbol]bool{},
		ContextExpectNames: map[*ast.Symbol]bool{},
	}
}

func collectRstestTestCallbacks(analysis *RstestCallAnalysis) RstestTestCallbacks {
	ctx := analysis.ctx
	result := newRstestTestCallbacks()
	pending := map[string][]*ParsedRstestFnCall{}

	for _, node := range analysis.calls {
		parsed := analysis.ParseTestCall(node)
		if parsed != nil {
			info := resolveRstestTestCallback(ctx, node.AsCallExpression())
			if info.functionNode != nil {
				recordRstestCallback(analysis, &result, info.functionNode, parsed)
			} else if info.name != "" {
				pending[info.name] = append(pending[info.name], parsed)
			}
		}
	}

	for name, parsedCalls := range pending {
		function := analysis.functions[name]
		if function == nil {
			continue
		}
		for _, parsed := range parsedCalls {
			recordRstestCallback(analysis, &result, function, parsed)
		}
	}
	return result
}

func resolveRstestTestCallback(
	ctx rule.RuleContext,
	call *ast.CallExpression,
) rstestCallbackInfo {
	if call == nil || call.Arguments == nil || len(call.Arguments.Nodes) < 2 {
		return rstestCallbackInfo{}
	}

	arguments := call.Arguments.Nodes
	// The third argument is the callback only in the `(name, options, fn)` shape.
	// In `(name, fn, timeout)` it is a timeout, and an unresolvable identifier
	// there must not shadow the real callback in the second position.
	if len(arguments) >= 3 {
		if info := resolveRstestCallbackArgument(ctx, arguments[2]); info.functionNode != nil {
			return info
		}
	}

	info := resolveRstestCallbackArgument(ctx, arguments[1])
	if info.functionNode == nil && info.name == "" && len(arguments) >= 3 {
		// The second argument is not a callback at all, so an unresolved name in
		// the third position is still worth deferring to the pending walk.
		if third := resolveRstestCallbackArgument(ctx, arguments[2]); third.name != "" {
			return third
		}
	}
	return info
}

func resolveRstestCallbackArgument(ctx rule.RuleContext, argument *ast.Node) rstestCallbackInfo {
	if argument == nil {
		return rstestCallbackInfo{}
	}
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
		initializer := declaration.AsVariableDeclaration().Initializer
		if initializer == nil {
			return rstestCallbackInfo{}
		}
		initializer = ast.SkipParentheses(initializer)
		if ast.IsFunctionExpressionOrArrowFunction(initializer) {
			return rstestCallbackInfo{functionNode: initializer, name: name}
		}
	}
	return rstestCallbackInfo{}
}

func recordRstestCallback(
	analysis *RstestCallAnalysis,
	result *RstestTestCallbacks,
	function *ast.Node,
	parsed *ParsedRstestFnCall,
) {
	ctx := analysis.ctx
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
	if parameter == nil || parameter.Name() == nil {
		return
	}
	name := parameter.Name()
	switch name.Kind {
	case ast.KindIdentifier:
		analysis.addExpectRootName(name.AsIdentifier().Text)
		if ctx.TypeChecker == nil {
			return
		}
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
			analysis.addExpectRootName(binding.Name().AsIdentifier().Text)
			if ctx.TypeChecker == nil {
				continue
			}
			if symbol := ctx.TypeChecker.GetSymbolAtLocation(binding.Name()); symbol != nil {
				result.ContextExpectNames[symbol] = true
			}
		}
	}
}
