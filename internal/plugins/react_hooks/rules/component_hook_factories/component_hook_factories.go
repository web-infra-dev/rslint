package component_hook_factories

import (
	"fmt"
	"strings"

	"github.com/dlclark/regexp2"
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/react_hooks/react_hooksutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

type componentHookFactoriesOptions struct {
	hookPattern *regexp2.Regexp
}

// ComponentHookFactoriesRule is the rslint port of upstream
// `react-hooks/component-hook-factories`.
//
// The original rule is produced by React Compiler diagnostics. This port keeps
// the Factories category local to rslint: flag functions that dynamically
// create a nested React component or Hook instead of defining it at module
// scope.
var ComponentHookFactoriesRule = rule.Rule{
	Name: "react-hooks/component-hook-factories",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := parseOptions(options)
		check := func(node *ast.Node) {
			validateNoDynamicallyCreatedComponentsOrHooks(ctx, node, opts)
		}
		return rule.RuleListeners{
			ast.KindFunctionDeclaration: check,
			ast.KindFunctionExpression:  check,
			ast.KindArrowFunction:       check,
		}
	},
}

func parseOptions(raw any) componentHookFactoriesOptions {
	optsMap := utils.GetOptionsMap(raw)
	if optsMap == nil {
		return componentHookFactoriesOptions{}
	}
	environment, ok := optsMap["environment"].(map[string]interface{})
	if !ok {
		return componentHookFactoriesOptions{}
	}
	pattern, ok := environment["hookPattern"].(string)
	if !ok || pattern == "" {
		return componentHookFactoriesOptions{}
	}
	compiled, err := utils.CompileRegexp2(pattern, utils.JSRegexOptions)
	if err != nil {
		return componentHookFactoriesOptions{}
	}
	return componentHookFactoriesOptions{hookPattern: compiled}
}

func validateNoDynamicallyCreatedComponentsOrHooks(ctx rule.RuleContext, fn *ast.Node, opts componentHookFactoriesOptions) {
	if isInsideClass(fn) {
		return
	}
	parentNameNode := react_hooksutil.GetFunctionName(fn)
	parentName := functionDisplayName(parentNameNode)
	if parentName == "" {
		parentName = "<anonymous>"
	}
	reportNode := parentNameNode
	if reportNode == nil {
		reportNode = fn
	}

	walkDirectChildren(fn, func(nestedFn *ast.Node) {
		nestedType := react_hooksutil.GetCompilerReactFunctionType(nestedFn, react_hooksutil.CompilerFunctionOptions{
			HookPattern: opts.hookPattern,
		})
		if nestedType == "" {
			return
		}
		nestedName := functionDisplayName(react_hooksutil.GetFunctionName(nestedFn))
		if nestedName == "" {
			nestedName = "<anonymous>"
		}
		ctx.ReportNode(reportNode, buildDynamicFactoryMessage(parentName, nestedName, string(nestedType)))
	})
}

func isInsideClass(node *ast.Node) bool {
	for parent := node.Parent; parent != nil; parent = parent.Parent {
		if ast.IsClassLike(parent) {
			return true
		}
		if ast.IsSourceFile(parent) {
			return false
		}
	}
	return false
}

func walkDirectChildren(root *ast.Node, visitFunction func(*ast.Node)) {
	var walk func(*ast.Node)
	walk = func(node *ast.Node) {
		if node == nil {
			return
		}
		if node != root && react_hooksutil.IsCompilerFunctionKind(node) {
			visitFunction(node)
			return
		}
		node.ForEachChild(func(child *ast.Node) bool {
			walk(child)
			return false
		})
	}
	walk(root)
}

func functionDisplayName(name *ast.Node) string {
	if name == nil {
		return ""
	}
	name = ast.SkipParentheses(name)
	if ast.IsIdentifier(name) {
		return name.AsIdentifier().Text
	}
	return ""
}

func buildDynamicFactoryMessage(parentName, nestedName, nestedType string) rule.RuleMessage {
	return rule.RuleMessage{
		Id: "componentHookFactory",
		Description: fmt.Sprintf(
			"Components and hooks cannot be created dynamically. The function `%s` appears to be a React %s, but it's defined inside `%s`. Components and Hooks should always be declared at module scope.",
			nestedName,
			strings.ToLower(nestedType),
			parentName,
		),
		Data: map[string]string{
			"nestedName": nestedName,
			"nestedType": nestedType,
			"parentName": parentName,
		},
	}
}
