package require_hook

import (
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/jest/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
)

func buildUseHookMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "useHook",
		Description: "This should be done within a hook",
	}
}

type Options struct {
	AllowedFunctionCalls []string
}

func parseAllowedFunctionCalls(raw any) []string {
	items, ok := raw.([]interface{})
	if !ok {
		return nil
	}
	out := make([]string, 0, len(items))
	for _, item := range items {
		if s, ok := item.(string); ok {
			out = append(out, s)
		}
	}
	return out
}

func parseOptions(options any) Options {
	opts := Options{AllowedFunctionCalls: nil}
	if options == nil {
		return opts
	}

	optArray := rule.NormalizeOptions(options)
	if len(optArray) == 0 {
		return opts
	}
	optsMap, ok := optArray[0].(map[string]interface{})
	if !ok {
		return opts
	}
	if raw, ok := optsMap["allowedFunctionCalls"]; ok {
		opts.AllowedFunctionCalls = parseAllowedFunctionCalls(raw)
	}
	return opts
}

func isNullOrUndefined(node *ast.Node) bool {
	node = ast.SkipParentheses(node)
	if node == nil {
		return false
	}
	switch node.Kind {
	case ast.KindNullKeyword:
		return true
	case ast.KindIdentifier:
		return node.AsIdentifier().Text == "undefined"
	default:
		return false
	}
}

func isJestFnCall(node *ast.Node, ctx rule.RuleContext) bool {
	if utils.ParseJestFnCall(node, ctx) != nil {
		return true
	}
	name := utils.CalleeChainName(node)
	return strings.HasPrefix(name, "jest.")
}

func containsString(list []string, target string) bool {
	for _, item := range list {
		if item == target {
			return true
		}
	}
	return false
}

func shouldBeInHook(node *ast.Node, ctx rule.RuleContext, allowedFunctionCalls []string) bool {
	if node == nil {
		return false
	}

	switch node.Kind {
	case ast.KindExpressionStatement:
		return shouldBeInHook(node.AsExpressionStatement().Expression, ctx, allowedFunctionCalls)
	case ast.KindCallExpression:
		if isJestFnCall(node, ctx) {
			return false
		}
		name := utils.CalleeChainName(node)
		return !containsString(allowedFunctionCalls, name)
	case ast.KindVariableStatement:
		declList := node.AsVariableStatement().DeclarationList
		if declList == nil || declList.Flags&ast.NodeFlagsConst != 0 {
			return false
		}
		decls := declList.AsVariableDeclarationList().Declarations
		if decls == nil {
			return false
		}
		for _, decl := range decls.Nodes {
			vd := decl.AsVariableDeclaration()
			if vd.Initializer != nil && !isNullOrUndefined(vd.Initializer) {
				return true
			}
		}
		return false
	default:
		return false
	}
}

func getFunctionBodyBlock(fn *ast.Node) *ast.Node {
	if fn == nil || !utils.IsFunction(fn) {
		return nil
	}
	switch fn.Kind {
	case ast.KindArrowFunction:
		body := fn.AsArrowFunction().Body
		if body != nil && body.Kind == ast.KindBlock {
			return body
		}
	case ast.KindFunctionExpression:
		return fn.AsFunctionExpression().Body
	case ast.KindFunctionDeclaration:
		return fn.AsFunctionDeclaration().Body
	}
	return nil
}

func describeCallbackBody(call *ast.Node, ctx rule.RuleContext) *ast.Node {
	if call == nil ||
		call.Kind != ast.KindCallExpression ||
		!utils.IsTypeOfJestFnCall(call, ctx, utils.JestFnTypeDescribe) {
		return nil
	}

	args := call.AsCallExpression().Arguments
	if args == nil || len(args.Nodes) < 2 {
		return nil
	}
	return getFunctionBodyBlock(args.Nodes[1])
}

func checkBlockBody(ctx rule.RuleContext, body []*ast.Node, allowedFunctionCalls []string) {
	for _, statement := range body {
		if shouldBeInHook(statement, ctx, allowedFunctionCalls) {
			ctx.ReportNode(statement, buildUseHookMessage())
		}
	}
}

var RequireHookRule = rule.Rule{
	Name: "jest/require-hook",
	Run: func(ctx rule.RuleContext, _options []any) rule.RuleListeners {
		options := rule.LegacyUnwrapOptions(_options)
		opts := parseOptions(options)

		if ctx.SourceFile != nil && ctx.SourceFile.Statements != nil {
			checkBlockBody(ctx, ctx.SourceFile.Statements.Nodes, opts.AllowedFunctionCalls)
		}

		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				block := describeCallbackBody(node, ctx)
				if block == nil || block.AsBlock().Statements == nil {
					return
				}
				checkBlockBody(ctx, block.AsBlock().Statements.Nodes, opts.AllowedFunctionCalls)
			},
		}
	},
}
