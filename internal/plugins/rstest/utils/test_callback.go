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

// rstestCallbackRegistration ties a callback function back to one of the
// registrations that runs it. A single function can be registered more than
// once, so ownership is a slice rather than a single registration.
type rstestCallbackRegistration struct {
	call   *ast.Node
	parsed *ParsedRstestFnCall
}

type rstestCallbackInfo struct {
	functionNode *ast.Node
	name         string
}

// rstestFunctionEntry is one name in the file-wide function index. ambiguous
// marks a name declared more than once, which makes the entry unusable for
// attributing a callback.
type rstestFunctionEntry struct {
	node      *ast.Node
	ambiguous bool
}

func newRstestTestCallbacks() RstestTestCallbacks {
	return RstestTestCallbacks{
		Functions:          map[*ast.Node]bool{},
		ContextReceivers:   map[*ast.Symbol]bool{},
		ContextExpectNames: map[*ast.Symbol]bool{},
	}
}

// walkRstestCallbackRegistrations visits every registration accepted by parse
// together with the function that runs its callback. Callbacks passed by name
// are held back until the whole file has been seen, so a registration can
// reference a function declared after it.
func walkRstestCallbackRegistrations(
	analysis *RstestCallAnalysis,
	parse func(*ast.Node) *ParsedRstestFnCall,
	visit func(function *ast.Node, registration rstestCallbackRegistration),
) {
	pending := map[string][]rstestCallbackRegistration{}

	for _, node := range analysis.calls {
		parsed := parse(node)
		if parsed == nil {
			continue
		}
		registration := rstestCallbackRegistration{call: node, parsed: parsed}
		info := analysis.callbackInfo(node)
		if info.functionNode != nil {
			visit(info.functionNode, registration)
		} else if info.name != "" {
			pending[info.name] = append(pending[info.name], registration)
		}
	}

	for name, registrations := range pending {
		entry := analysis.functions[name]
		// The index is file-wide and scope-blind, so a name that is declared
		// twice, or declared inside some other callback, cannot be attributed
		// to this registration: `describe('s', () => { function cb() {} });
		// test.concurrent('x', cb)` would otherwise be answered with the nested
		// `cb`, which `test.concurrent` never runs. Reaching here at all means
		// the checker could not resolve the identifier, so declining to guess
		// costs coverage on code that already fails to compile.
		if entry.node == nil || entry.ambiguous || !isModuleTopLevelFunction(entry.node) {
			continue
		}
		for _, registration := range registrations {
			visit(entry.node, registration)
		}
	}
}

// collectRstestTestCallbacks records the TestContext bindings each test
// callback receives. Describe callbacks are deliberately excluded: their
// parameters are not a TestContext.
func collectRstestTestCallbacks(analysis *RstestCallAnalysis) RstestTestCallbacks {
	result := newRstestTestCallbacks()
	walkRstestCallbackRegistrations(
		analysis,
		analysis.ParseTestCall,
		func(function *ast.Node, registration rstestCallbackRegistration) {
			recordRstestTestCallback(analysis, &result, function, registration.parsed)
		},
	)
	return result
}

// collectRstestCallbackOwnership indexes both test and describe callbacks by
// the registrations that run them, which is what an execution mode is
// inherited through.
func collectRstestCallbackOwnership(
	analysis *RstestCallAnalysis,
) map[*ast.Node][]rstestCallbackRegistration {
	ownership := map[*ast.Node][]rstestCallbackRegistration{}
	walkRstestCallbackRegistrations(
		analysis,
		analysis.parseRegistrationCall,
		func(function *ast.Node, registration rstestCallbackRegistration) {
			ownership[function] = append(ownership[function], registration)
		},
	)
	return ownership
}

// isModuleTopLevelFunction reports whether function is declared directly at
// module scope, the one place a file-wide name lookup is visible from every
// call site in the file. The walk runs only for callbacks the checker failed
// to resolve, so it is off the common path.
func isModuleTopLevelFunction(function *ast.Node) bool {
	if function == nil {
		return false
	}
	for node := function.Parent; node != nil; node = node.Parent {
		switch node.Kind {
		case ast.KindSourceFile:
			return true
		case ast.KindVariableDeclaration,
			ast.KindVariableDeclarationList,
			ast.KindVariableStatement,
			ast.KindParenthesizedExpression:
			// The declaration chain a `const cb = () => {}` hangs from, and the
			// parentheses SkipParentheses already looked through.
		default:
			return false
		}
	}
	return false
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
	argument = internalUtils.SkipAssertionsAndParens(argument)
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
		initializer = internalUtils.SkipAssertionsAndParens(initializer)
		if ast.IsFunctionExpressionOrArrowFunction(initializer) {
			return rstestCallbackInfo{functionNode: initializer, name: name}
		}
	}
	return rstestCallbackInfo{}
}

func recordRstestTestCallback(
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
