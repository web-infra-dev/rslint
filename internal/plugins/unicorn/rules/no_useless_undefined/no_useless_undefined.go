// Ported from eslint-plugin-unicorn v79.0.0. See LICENSE.
package no_useless_undefined

import (
	_ "embed"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/unicornutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

//go:embed no_useless_undefined.schema.json
var schema []byte

var uselessUndefined = rule.RuleMessage{
	Id:          "no-useless-undefined",
	Description: "Do not use useless `undefined`.",
}

var NoUselessUndefinedRule = rule.Rule{
	Name:   "unicorn/no-useless-undefined",
	Schema: rule.NewSchema(schema),
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		checkArguments, checkArrowBody := true, true
		if len(options) > 0 {
			option, _ := options[0].(map[string]any)
			if value, ok := option["checkArguments"].(bool); ok {
				checkArguments = value
			}
			if value, ok := option["checkArrowFunctionBody"].(bool); ok {
				checkArrowBody = value
			}
		}
		kind := core.GetScriptKindFromFileName(ctx.SourceFile.FileName())
		isTypeScript := kind == core.ScriptKindTS || kind == core.ScriptKindTSX
		listeners := rule.RuleListeners{
			ast.KindIdentifier: func(node *ast.Node) {
				if node.Text() == "undefined" {
					checkIdentifier(ctx, node, checkArrowBody, isTypeScript)
				}
			},
			ast.KindConditionalExpression: indexedAccessListener(ctx),
		}
		if checkArguments && !isTypeScript {
			listeners[ast.KindCallExpression] = func(node *ast.Node) {
				checkLastArgument(ctx, node)
			}
		}
		return listeners
	},
}

func checkIdentifier(ctx rule.RuleContext, node *ast.Node, checkArrowBody, isTypeScript bool) {
	parent := utils.ESTreeParent(node)
	if parent == nil {
		return
	}
	switch parent.Kind {
	case ast.KindReturnStatement:
		expression := parent.Expression()
		if utils.ESTreeRuntimeExpression(expression) != node {
			return
		}
		if function := ast.GetContainingFunction(parent); function != nil {
			annotation := function.Type()
			for annotation != nil && annotation.Kind == ast.KindParenthesizedType {
				annotation = annotation.Type()
			}
			// An explicit union or other value type retains the explicit return.
			if annotation != nil &&
				annotation.Kind != ast.KindUndefinedKeyword && annotation.Kind != ast.KindVoidKeyword {
				return
			}
		}
		reportExpression(ctx, node, expression, "")
	case ast.KindYieldExpression:
		yield := parent.AsYieldExpression()
		if yield.AsteriskToken == nil && utils.ESTreeRuntimeExpression(yield.Expression) == node {
			reportExpression(ctx, node, yield.Expression, "")
		}
	case ast.KindArrowFunction:
		if checkArrowBody && parent.Type() == nil && utils.ESTreeRuntimeExpression(parent.Body()) == node {
			reportExpression(ctx, node, parent.Body(), " {}")
		}
	case ast.KindVariableDeclaration:
		if ast.IsVarLet(parent) && ast.IsIdentifier(parent.Name()) {
			reportInitializer(ctx, node, parent.Name(), parent.Initializer(), isTypeScript)
		}
	case ast.KindParameter, ast.KindBindingElement:
		reportInitializer(ctx, node, parent.Name(), parent.Initializer(), isTypeScript)
	case ast.KindShorthandPropertyAssignment:
		if utils.IsInDestructuringAssignment(parent) {
			reportInitializer(ctx, node, parent.Name(), parent.AsShorthandPropertyAssignment().ObjectAssignmentInitializer, isTypeScript)
		}
	case ast.KindBinaryExpression:
		binary := parent.AsBinaryExpression()
		if utils.ESTreeRuntimeExpression(binary.Right) == node && utils.IsDefaultValueInDestructuringAssignment(parent) {
			reportInitializer(ctx, node, binary.Left, binary.Right, isTypeScript)
		}
	}
}

func checkLastArgument(ctx rule.RuleContext, node *ast.Node) {
	call := node.AsCallExpression()
	args := node.Arguments()
	if len(args) == 0 {
		return
	}
	argument := args[len(args)-1]
	undefined := utils.ESTreeRuntimeExpression(argument)
	if !utils.IsUndefinedIdentifier(undefined) {
		return
	}
	callee := utils.ESTreeCallCallee(call.Expression)
	if callee != nil && callee.Kind == ast.KindImportKeyword {
		return
	}
	var name string
	if callee != nil {
		switch callee.Kind {
		case ast.KindIdentifier:
			name = callee.Text()
		case ast.KindPropertyAccessExpression:
			if ast.IsIdentifier(callee.Name()) {
				name = callee.Name().Text()
			}
			// Only the first argument of a non-optional .bind() is the thisArg.
			if name == "bind" && call.QuestionDotToken == nil && len(args) > 1 {
				return
			}
		}
	}
	if ignoresTrailingUndefined(name) || !unicornutil.IsGlobalReference(ctx, undefined) {
		return
	}
	ctx.ReportNodeWithDeferredFixes(undefined, uselessUndefined, func() []rule.RuleFix {
		return lastArgumentFix(ctx, argument, len(args) == 1)
	})
}

func ignoresTrailingUndefined(name string) bool {
	switch name {
	// Assertions and APIs for which omitting the argument can change behavior.
	case "is", "equal", "notEqual", "strictEqual", "notStrictEqual", "propertyVal",
		"notPropertyVal", "not", "include", "property", "toBe", "toHaveBeenCalledWith",
		"toContain", "toContainEqual", "toEqual", "same", "notSame", "strictSame", "strictNotSame",
		"push", "unshift", "includes", "concat", "splice", "toSpliced", "lastIndexOf", "reduce", "reduceRight",
		"String", "fromCharCode", "fromCodePoint", "Number", "max", "min", "hypot",
		"Array", "of", "from", "structuredClone", "add", "has", "delete", "set", "createContext", "useRef", "ref":
		return true
	}
	return strings.HasPrefix(name, "set") && len(name) > 3 && name[3] >= 'A' && name[3] <= 'Z'
}
